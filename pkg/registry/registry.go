package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// Registry 服务注册器
type Registry struct {
	client   *clientv3.Client
	config   *Config
	instance *ServiceInstance

	leaseID clientv3.LeaseID
	ttl     int64
	key     string

	// 用于控制keepAlive goroutine
	keepAliveCancel context.CancelFunc
	closeChan       chan struct{}
	wg              sync.WaitGroup

	mu         sync.RWMutex
	registered bool
}

// NewRegistry 创建注册器
func NewRegistry(cfg *Config, instance *ServiceInstance) (*Registry, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if instance == nil {
		return nil, fmt.Errorf("instance cannot be nil")
	}

	// 创建etcd客户端
	clientCfg := clientv3.Config{
		Endpoints:        cfg.EtcdEndpoints,
		DialTimeout:      cfg.DialTimeout,
		AutoSyncInterval: cfg.AutoSyncInterval,
	}

	if cfg.Username != "" {
		clientCfg.Username = cfg.Username
		clientCfg.Password = cfg.Password
	}

	client, err := clientv3.New(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	return &Registry{
		client:    client,
		config:    cfg,
		instance:  instance,
		key:       fmt.Sprintf("/services/%s/%s", cfg.ServiceName, cfg.InstanceID),
		closeChan: make(chan struct{}),
	}, nil
}

// Register 注册服务
func (r *Registry) Register(ctx context.Context) error {
	r.mu.Lock()
	if r.registered {
		r.mu.Unlock()
		return ErrAlreadyRegistered
	}
	r.mu.Unlock()

	r.ttl = r.config.TTL

	// 1. 创建租约
	grantResp, err := r.client.Grant(ctx, r.ttl)
	if err != nil {
		return fmt.Errorf("failed to grant lease: %w", err)
	}
	r.leaseID = grantResp.ID

	// 2. 注册服务信息（绑定租约）
	data, err := json.Marshal(r.instance)
	if err != nil {
		return fmt.Errorf("failed to marshal instance: %w", err)
	}

	_, err = r.client.Put(ctx, r.key, string(data), clientv3.WithLease(r.leaseID))
	if err != nil {
		return fmt.Errorf("failed to put service: %w", err)
	}

	// 3. 启动心跳保活
	keepAliveCtx, cancel := context.WithCancel(context.Background())
	r.keepAliveCancel = cancel

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		r.keepAlive(keepAliveCtx)
	}()

	r.mu.Lock()
	r.registered = true
	r.mu.Unlock()

	log.Printf("Service registered: %s (LeaseID: %d, TTL: %ds)", r.key, r.leaseID, r.ttl)
	return nil
}

// keepAlive 保持心跳（带自动重注册）
func (r *Registry) keepAlive(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		kaChannel, err := r.client.KeepAlive(ctx, r.leaseID)
		if err != nil {
			log.Printf("Failed to start keep alive: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		lastRenew := time.Now()
		shouldReregister := false

		for {
			select {
			case resp := <-kaChannel:
				if resp == nil {
					log.Printf("KeepAlive channel closed, attempting to re-register...")
					shouldReregister = true
					break
				}
				lastRenew = time.Now()

			case <-ticker.C:
				log.Printf("Lease active (ID: %d, last renewed: %v ago)",
					r.leaseID, time.Since(lastRenew))

			case <-ctx.Done():
				log.Printf("KeepAlive stopped by context cancellation")
				return

			case <-r.closeChan:
				log.Printf("KeepAlive stopped by close signal")
				return
			}

			if shouldReregister {
				break
			}
		}

		if shouldReregister && ctx.Err() == nil {
			if err := r.reRegister(ctx); err != nil {
				log.Printf("Failed to re-register: %v, retrying in 5s", err)
				time.Sleep(5 * time.Second)
			} else {
				log.Printf("Successfully re-registered with new lease: %d", r.leaseID)
			}
		} else {
			return
		}
	}
}

// reRegister 重新注册服务
func (r *Registry) reRegister(ctx context.Context) error {
	// 1. 创建新租约（使用原始TTL）
	grantResp, err := r.client.Grant(ctx, r.ttl)
	if err != nil {
		return fmt.Errorf("failed to grant lease: %w", err)
	}
	r.leaseID = grantResp.ID

	// 2. 重新注册服务信息
	data, err := json.Marshal(r.instance)
	if err != nil {
		return fmt.Errorf("failed to marshal instance: %w", err)
	}

	_, err = r.client.Put(ctx, r.key, string(data), clientv3.WithLease(r.leaseID))
	if err != nil {
		return fmt.Errorf("failed to put service: %w", err)
	}

	return nil
}

// Unregister 注销服务
func (r *Registry) Unregister(ctx context.Context) error {
	r.mu.Lock()
	if !r.registered {
		r.mu.Unlock()
		return ErrNotRegistered
	}
	r.mu.Unlock()

	// 1. 发送停止信号
	if r.keepAliveCancel != nil {
		r.keepAliveCancel()
	}
	close(r.closeChan)

	// 2. 等待keepAlive goroutine退出（带超时）
	done := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Printf("KeepAlive goroutine exited cleanly")
	case <-time.After(5 * time.Second):
		log.Printf("Warning: KeepAlive goroutine did not exit in time")
	}

	// 3. 撤销租约，自动删除所有关联键值对
	if r.leaseID != 0 {
		revokeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		_, err := r.client.Revoke(revokeCtx, r.leaseID)
		if err != nil {
			log.Printf("Failed to revoke lease: %v", err)
		}
	}

	r.mu.Lock()
	r.registered = false
	r.mu.Unlock()

	log.Printf("Service unregistered: %s", r.key)
	return nil
}

// Close 关闭Registry并释放资源
func (r *Registry) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// IsHealthy 检查服务是否健康
func (r *Registry) IsHealthy() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.registered && r.leaseID != 0 && r.client != nil
}

// GetStatus 获取状态详情
func (r *Registry) GetStatus() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return map[string]interface{}{
		"registered":   r.registered,
		"lease_id":     r.leaseID,
		"ttl":          r.ttl,
		"key":          r.key,
		"service_name": r.config.ServiceName,
		"instance_id":  r.config.InstanceID,
		"healthy":      r.IsHealthy(),
	}
}
