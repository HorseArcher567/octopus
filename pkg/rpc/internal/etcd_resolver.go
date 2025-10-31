package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"

	"github.com/HorseArcher567/octopus/pkg/registry"
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
	r := &etcdResolver{
		target:    target,
		cc:        cc,
		endpoints: b.etcdEndpoints,
		ctx:       context.Background(),
		addrs:     make(map[string]resolver.Address),
	}

	r.ctx, r.cancel = context.WithCancel(r.ctx)

	// 连接etcd
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   b.etcdEndpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
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
}

// ResolveNow gRPC调用此方法请求立即解析
func (r *etcdResolver) ResolveNow(resolver.ResolveNowOptions) {
	// 触发一次立即查询
	if err := r.loadServices(); err != nil {
		log.Printf("ResolveNow failed: %v", err)
	}
}

// Close 关闭resolver
func (r *etcdResolver) Close() {
	r.cancel()
	if r.client != nil {
		r.client.Close()
	}
}

// watch 监听服务变化（带自动重连）
func (r *etcdResolver) watch() {
	// 首先加载现有服务
	if err := r.loadServices(); err != nil {
		log.Printf("Failed to load services: %v", err)
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

		// 发生错误，等待后重试
		log.Printf("Watch error: %v, retrying in %v", err, backoff)
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
	prefix := fmt.Sprintf("/services/%s/", r.target.Endpoint())
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
	prefix := fmt.Sprintf("/services/%s/", r.target.Endpoint())
	resp, err := r.client.Get(r.ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.addrs = make(map[string]resolver.Address)
	for _, kv := range resp.Kvs {
		var instance registry.ServiceInstance
		if err := json.Unmarshal(kv.Value, &instance); err != nil {
			log.Printf("Failed to unmarshal: %v", err)
			continue
		}

		addr := fmt.Sprintf("%s:%d", instance.Addr, instance.Port)
		r.addrs[string(kv.Key)] = resolver.Address{
			Addr:     addr,
			Metadata: &instance,
		}
	}

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
			log.Printf("Failed to unmarshal: %v", err)
			return
		}
		addr := fmt.Sprintf("%s:%d", instance.Addr, instance.Port)
		r.addrs[key] = resolver.Address{
			Addr:     addr,
			Metadata: &instance,
		}
		log.Printf("Service added: %s", addr)

	case mvccpb.DELETE:
		if addr, ok := r.addrs[key]; ok {
			log.Printf("Service removed: %s", addr.Addr)
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

	r.cc.UpdateState(resolver.State{Addresses: addrs})
}
