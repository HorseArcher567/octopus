package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/HorseArcher567/octopus/pkg/etcd"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Registry 服务注册器
type Registry struct {
	client   *clientv3.Client
	config   *Config
	instance *ServiceInstance
	log      *slog.Logger

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
// etcdConfig 是 etcd 配置，如果为 nil 则使用默认配置
// 从 context 中获取 logger，如果没有则使用 slog.Default()
func NewRegistry(ctx context.Context, etcdConfig *etcd.Config, cfg *Config, instance *ServiceInstance) (*Registry, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if instance == nil {
		return nil, fmt.Errorf("instance cannot be nil")
	}

	// 使用 etcd.NewClient 创建 client
	client, err := etcd.NewClient(etcdConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	// 从 context 获取 logger 并添加组件信息
	log := xlog.FromContext(ctx).With("component", "registry", "app_name", cfg.AppName)

	return &Registry{
		client:    client,
		config:    cfg,
		instance:  instance,
		log:       log,
		key:       fmt.Sprintf("/octopus/applications/%s/%s:%d", cfg.AppName, instance.Addr, instance.Port),
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

	r.log.Info("application registered",
		"key", r.key,
		"lease_id", r.leaseID,
		"ttl", r.ttl,
		"instance", fmt.Sprintf("%s:%d", r.instance.Addr, r.instance.Port),
	)
	return nil
}

// keepAlive 保持心跳（带自动重注册）
func (r *Registry) keepAlive(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		kaChannel, err := r.client.KeepAlive(ctx, r.leaseID)
		if err != nil {
			r.log.Error("failed to start keep alive",
				"error", err,
				"lease_id", r.leaseID,
			)
			time.Sleep(5 * time.Second)
			continue
		}

		lastRenew := time.Now()
		shouldReregister := false

		for {
			select {
			case resp := <-kaChannel:
				if resp == nil {
					r.log.Warn("keepalive channel closed, attempting to re-register",
						"lease_id", r.leaseID,
					)
					shouldReregister = true
					break
				}
				lastRenew = time.Now()

			case <-ticker.C:
				r.log.Debug("lease active",
					"lease_id", r.leaseID,
					"last_renewed", time.Since(lastRenew),
				)

			case <-ctx.Done():
				r.log.Info("keepalive stopped by context cancellation")
				return

			case <-r.closeChan:
				r.log.Info("keepalive stopped by close signal")
				return
			}

			if shouldReregister {
				break
			}
		}

		if shouldReregister && ctx.Err() == nil {
			if err := r.reRegister(ctx); err != nil {
				r.log.Error("failed to re-register, retrying",
					"error", err,
					"retry_delay", "5s",
				)
				time.Sleep(5 * time.Second)
			} else {
				r.log.Info("successfully re-registered",
					"lease_id", r.leaseID,
				)
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
		r.log.Info("keepalive goroutine exited cleanly")
	case <-time.After(5 * time.Second):
		r.log.Warn("keepalive goroutine did not exit in time", "timeout", "5s")
	}

	// 3. 撤销租约，自动删除所有关联键值对
	if r.leaseID != 0 {
		revokeCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		_, err := r.client.Revoke(revokeCtx, r.leaseID)
		if err != nil {
			r.log.Error("failed to revoke lease",
				"error", err,
				"lease_id", r.leaseID,
			)
		}
	}

	r.mu.Lock()
	r.registered = false
	r.mu.Unlock()

	r.log.Info("service unregistered", "key", r.key)
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
		"registered": r.registered,
		"lease_id":   r.leaseID,
		"ttl":        r.ttl,
		"key":        r.key,
		"app_name":   r.config.AppName,
		"instance":   fmt.Sprintf("%s:%d", r.instance.Addr, r.instance.Port),
		"healthy":    r.IsHealthy(),
	}
}
