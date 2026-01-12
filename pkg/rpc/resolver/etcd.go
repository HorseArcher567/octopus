package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/HorseArcher567/octopus/pkg/etcd"
	"github.com/HorseArcher567/octopus/pkg/rpc/registry"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	grpcresolver "google.golang.org/grpc/resolver"
)

// EtcdResolverBuilder 实现gRPC的resolver.Builder接口
type EtcdResolverBuilder struct {
	etcdConfig *etcd.Config
	log        *slog.Logger
}

// NewBuilder 创建resolver builder
// etcdConfig 是 etcd 配置，如果为 nil 则使用默认配置
// 从 context 中获取 logger，如果没有则使用 slog.Default()
func NewBuilder(ctx context.Context, etcdConfig *etcd.Config) *EtcdResolverBuilder {
	return &EtcdResolverBuilder{
		etcdConfig: etcdConfig,
		log:        xlog.FromContext(ctx),
	}
}

// Scheme 返回resolver的scheme（用于gRPC URL）
func (b *EtcdResolverBuilder) Scheme() string {
	return "etcd"
}

// Build 创建一个新的resolver实例
func (b *EtcdResolverBuilder) Build(target grpcresolver.Target, cc grpcresolver.ClientConn, opts grpcresolver.BuildOptions) (grpcresolver.Resolver, error) {
	log := b.log.With("component", "resolver", "service", target.Endpoint())

	// 使用 etcd.NewClient 创建 client
	cfg := b.etcdConfig
	if cfg == nil {
		cfg = etcd.Default()
	}

	client, err := etcd.NewClient(cfg)
	if err != nil {
		log.Error("failed to connect to etcd", "error", err)
		return nil, fmt.Errorf("failed to connect etcd: %w", err)
	}

	log.Info("building resolver", "etcd_endpoints", cfg.Endpoints)

	r := &etcdResolver{
		target: target,
		cc:     cc,
		ctx:    context.Background(),

		log:    log,
		addrs:  make(map[string]grpcresolver.Address),
		closed: false,
	}

	r.ctx, r.cancel = context.WithCancel(r.ctx)
	r.client = client

	// 启动服务发现
	go r.watch()

	return r, nil
}

// etcdResolver 实现gRPC的resolver.Resolver接口
type etcdResolver struct {
	target grpcresolver.Target
	cc     grpcresolver.ClientConn
	client *clientv3.Client
	ctx    context.Context
	cancel context.CancelFunc
	log    *slog.Logger
	addrs  map[string]grpcresolver.Address
	mu     sync.RWMutex
	closed bool // 标记是否已关闭
}

// ResolveNow gRPC调用此方法请求立即解析
func (r *etcdResolver) ResolveNow(grpcresolver.ResolveNowOptions) {
	// 触发一次立即查询
	if err := r.loadServices(); err != nil {
		r.log.Error("resolve now failed", "error", err)
	}
}

// Close 关闭resolver
func (r *etcdResolver) Close() {
	r.mu.Lock()
	r.closed = true
	r.mu.Unlock()

	r.cancel()
	if r.client != nil {
		r.client.Close()
	}
	r.log.Info("resolver closed")
}

// watch 监听服务变化（带自动重连）
func (r *etcdResolver) watch() {
	// 首先加载现有服务
	if err := r.loadServices(); err != nil {
		r.log.Error("failed to load services", "error", err)
	}

	backoff := time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-r.ctx.Done():
			return
		default:
		}

		// 尝试监听
		err := r.watchOnce()
		if err == nil {
			backoff = time.Second // 正常退出，重置退避时间
			continue
		}

		// 检查是否已关闭
		r.mu.RLock()
		closed := r.closed
		r.mu.RUnlock()

		if closed {
			// 正常关闭，不打印错误
			return
		}

		// 发生错误，等待后重试
		r.log.Warn("watch error, retrying",
			"error", err,
			"retry_delay", backoff,
		)
		select {
		case <-time.After(backoff):
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		case <-r.ctx.Done():
			return
		}
	}
}

// watchOnce 执行单次监听
func (r *etcdResolver) watchOnce() error {
	prefix := fmt.Sprintf("/octopus/applications/%s/", r.target.Endpoint())
	watchChan := r.client.Watch(r.ctx, prefix, clientv3.WithPrefix())

	for watchResp := range watchChan {
		// 检查是否被取消
		if watchResp.Canceled {
			return fmt.Errorf("watch was canceled")
		}

		// 检查错误
		if err := watchResp.Err(); err != nil {
			return fmt.Errorf("watch error: %w", err)
		}

		// 处理事件
		for _, event := range watchResp.Events {
			r.handleEvent(event)
		}

		// 更新gRPC连接
		r.updateState()
	}

	// watchChan关闭，可能需要重连
	return fmt.Errorf("watch channel closed")
}

// loadServices 加载现有服务实例
func (r *etcdResolver) loadServices() error {
	prefix := fmt.Sprintf("/octopus/applications/%s/", r.target.Endpoint())

	resp, err := r.client.Get(r.ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		r.log.Error("failed to get services from etcd", "error", err)
		return err
	}

	r.mu.Lock()
	r.addrs = make(map[string]grpcresolver.Address)
	for _, kv := range resp.Kvs {
		var instance registry.ServiceInstance
		if err := json.Unmarshal(kv.Value, &instance); err != nil {
			r.log.Error("failed to unmarshal instance",
				"error", err,
				"key", string(kv.Key),
			)
			continue
		}

		addr := fmt.Sprintf("%s:%d", instance.Addr, instance.Port)
		r.addrs[string(kv.Key)] = grpcresolver.Address{
			Addr:     addr,
			Metadata: &instance,
		}
	}
	r.mu.Unlock()

	// 在锁外调用 updateState，避免死锁
	r.updateState()
	return nil
}

// handleEvent 处理etcd事件
func (r *etcdResolver) handleEvent(event *clientv3.Event) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := string(event.Kv.Key)

	switch event.Type {
	case mvccpb.PUT:
		var instance registry.ServiceInstance
		if err := json.Unmarshal(event.Kv.Value, &instance); err != nil {
			r.log.Error("failed to unmarshal instance",
				"error", err,
				"key", key,
			)
			return
		}
		addr := fmt.Sprintf("%s:%d", instance.Addr, instance.Port)
		r.addrs[key] = grpcresolver.Address{
			Addr:     addr,
			Metadata: &instance,
		}
		r.log.Info("service instance added",
			"addr", addr,
			"key", key,
		)

	case mvccpb.DELETE:
		if addr, ok := r.addrs[key]; ok {
			r.log.Info("service instance removed",
				"addr", addr.Addr,
				"key", key,
			)
			delete(r.addrs, key)
		}
	}
}

// updateState 更新gRPC连接状态
func (r *etcdResolver) updateState() {
	r.mu.RLock()
	addrs := make([]grpcresolver.Address, 0, len(r.addrs))
	for _, addr := range r.addrs {
		addrs = append(addrs, addr)
	}
	r.mu.RUnlock()

	if len(addrs) == 0 {
		r.log.Warn("no service instances found")
	} else {
		addrList := make([]string, len(addrs))
		for i, addr := range addrs {
			addrList[i] = addr.Addr
		}
		r.log.Info("discovered service instances",
			"count", len(addrs),
			"instances", addrList,
		)
	}

	r.cc.UpdateState(grpcresolver.State{Addresses: addrs})
}
