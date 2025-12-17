package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/HorseArcher567/octopus/pkg/logger"
	"github.com/HorseArcher567/octopus/pkg/registry"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
)

// EtcdResolverBuilder 实现gRPC的resolver.Builder接口
type EtcdResolverBuilder struct {
	etcdEndpoints []string
}

// NewBuilder 创建resolver builder
func NewBuilder(endpoints []string) *EtcdResolverBuilder {
	return &EtcdResolverBuilder{
		etcdEndpoints: endpoints,
	}
}

// Scheme 返回resolver的scheme（用于gRPC URL）
func (b *EtcdResolverBuilder) Scheme() string {
	return "etcd"
}

// Build 创建一个新的resolver实例
func (b *EtcdResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	logger.Info("building resolver for service",
		"service", target.Endpoint(),
		"etcd_endpoints", b.etcdEndpoints,
	)

	r := &etcdResolver{
		target:    target,
		cc:        cc,
		endpoints: b.etcdEndpoints,
		ctx:       context.Background(),
		addrs:     make(map[string]resolver.Address),
		closed:    false,
	}

	r.ctx, r.cancel = context.WithCancel(r.ctx)

	// 连接etcd
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   b.etcdEndpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		logger.Error("failed to connect to etcd",
			"error", err,
			"etcd_endpoints", b.etcdEndpoints,
			"service", target.Endpoint(),
		)
		return nil, fmt.Errorf("failed to connect etcd: %w", err)
	}
	r.client = client

	// 启动服务发现
	go r.watch()

	return r, nil
}

// etcdResolver 实现gRPC的resolver.Resolver接口
type etcdResolver struct {
	target    resolver.Target
	cc        resolver.ClientConn
	client    *clientv3.Client
	endpoints []string
	ctx       context.Context
	cancel    context.CancelFunc
	addrs     map[string]resolver.Address
	mu        sync.RWMutex
	closed    bool // 标记是否已关闭
}

// ResolveNow gRPC调用此方法请求立即解析
func (r *etcdResolver) ResolveNow(resolver.ResolveNowOptions) {
	// 触发一次立即查询
	if err := r.loadServices(); err != nil {
		logger.Error("resolve now failed",
			"error", err,
			"service", r.target.Endpoint(),
		)
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
	logger.Info("resolver closed",
		"service", r.target.Endpoint(),
	)
}

// watch 监听服务变化（带自动重连）
func (r *etcdResolver) watch() {
	// 首先加载现有服务
	if err := r.loadServices(); err != nil {
		logger.Error("failed to load services",
			"error", err,
			"service", r.target.Endpoint(),
		)
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
		logger.Warn("watch error, retrying",
			"error", err,
			"retry_delay", backoff,
			"service", r.target.Endpoint(),
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
		logger.Error("failed to get services from etcd",
			"error", err,
			"service", r.target.Endpoint(),
		)
		return err
	}

	r.mu.Lock()
	r.addrs = make(map[string]resolver.Address)
	for _, kv := range resp.Kvs {
		var instance registry.ServiceInstance
		if err := json.Unmarshal(kv.Value, &instance); err != nil {
			logger.Error("failed to unmarshal instance",
				"error", err,
				"key", string(kv.Key),
				"service", r.target.Endpoint(),
			)
			continue
		}

		addr := fmt.Sprintf("%s:%d", instance.Addr, instance.Port)
		r.addrs[string(kv.Key)] = resolver.Address{
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
			logger.Error("failed to unmarshal instance",
				"error", err,
				"key", key,
				"service", r.target.Endpoint(),
			)
			return
		}
		addr := fmt.Sprintf("%s:%d", instance.Addr, instance.Port)
		r.addrs[key] = resolver.Address{
			Addr:     addr,
			Metadata: &instance,
		}
		logger.Info("service instance added",
			"addr", addr,
			"key", key,
			"service", r.target.Endpoint(),
		)

	case mvccpb.DELETE:
		if addr, ok := r.addrs[key]; ok {
			logger.Info("service instance removed",
				"addr", addr.Addr,
				"key", key,
				"service", r.target.Endpoint(),
			)
			delete(r.addrs, key)
		}
	}
}

// updateState 更新gRPC连接状态
func (r *etcdResolver) updateState() {
	r.mu.RLock()
	addrs := make([]resolver.Address, 0, len(r.addrs))
	for _, addr := range r.addrs {
		addrs = append(addrs, addr)
	}
	r.mu.RUnlock()

	if len(addrs) == 0 {
		logger.Warn("no service instances found",
			"service", r.target.Endpoint(),
		)
	} else {
		addrList := make([]string, len(addrs))
		for i, addr := range addrs {
			addrList[i] = addr.Addr
		}
		logger.Info("discovered service instances",
			"service", r.target.Endpoint(),
			"count", len(addrs),
			"instances", addrList,
		)
	}

	r.cc.UpdateState(resolver.State{Addresses: addrs})
}
