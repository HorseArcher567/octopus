package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Discovery 服务发现器
type Discovery struct {
	client      *clientv3.Client
	serviceName string
	instances   map[string]*ServiceInstance
	mu          sync.RWMutex

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewDiscovery 创建发现器
func NewDiscovery(etcdEndpoints []string) (*Discovery, error) {
	if len(etcdEndpoints) == 0 {
		return nil, ErrEmptyEndpoints
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:        etcdEndpoints,
		DialTimeout:      5 * time.Second,
		AutoSyncInterval: 60 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	return &Discovery{
		client:    client,
		instances: make(map[string]*ServiceInstance),
	}, nil
}

// Watch 监听服务变化
func (d *Discovery) Watch(ctx context.Context, serviceName string) error {
	if serviceName == "" {
		return ErrEmptyServiceName
	}

	d.serviceName = serviceName
	prefix := fmt.Sprintf("/services/%s/", serviceName)

	// 1. 首先获取已有的服务实例
	if err := d.loadInstances(ctx, prefix); err != nil {
		return fmt.Errorf("failed to load instances: %w", err)
	}

	// 2. 启动监听
	watchCtx, cancel := context.WithCancel(ctx)
	d.cancel = cancel

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.watchChanges(watchCtx, prefix)
	}()

	return nil
}

// loadInstances 加载已有的服务实例
func (d *Discovery) loadInstances(ctx context.Context, prefix string) error {
	resp, err := d.client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return fmt.Errorf("failed to get instances: %w", err)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	for _, kv := range resp.Kvs {
		var instance ServiceInstance
		if err := json.Unmarshal(kv.Value, &instance); err != nil {
			log.Printf("Failed to unmarshal instance: %v", err)
			continue
		}
		d.instances[string(kv.Key)] = &instance
		log.Printf("Loaded instance: %s -> %s:%d", string(kv.Key), instance.Addr, instance.Port)
	}

	return nil
}

// watchChanges 监听服务变化（带自动重连）
func (d *Discovery) watchChanges(ctx context.Context, prefix string) {
	backoff := time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-ctx.Done():
			log.Printf("Watch stopped: %v", ctx.Err())
			return
		default:
		}

		err := d.watchSingle(ctx, prefix)
		if err == nil {
			backoff = time.Second
			continue
		}

		log.Printf("Watch error: %v, retrying in %v", err, backoff)
		select {
		case <-time.After(backoff):
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		case <-ctx.Done():
			return
		}
	}
}

// watchSingle 执行单次监听
func (d *Discovery) watchSingle(ctx context.Context, prefix string) error {
	watchChan := d.client.Watch(ctx, prefix, clientv3.WithPrefix())

	for watchResp := range watchChan {
		if watchResp.Canceled {
			return fmt.Errorf("watch was canceled")
		}

		if err := watchResp.Err(); err != nil {
			return fmt.Errorf("watch error: %w", err)
		}

		d.mu.Lock()
		for _, event := range watchResp.Events {
			switch event.Type {
			case mvccpb.PUT:
				var instance ServiceInstance
				if err := json.Unmarshal(event.Kv.Value, &instance); err != nil {
					log.Printf("Failed to unmarshal instance: %v", err)
					continue
				}
				d.instances[string(event.Kv.Key)] = &instance
				log.Printf("Instance added/updated: %s -> %s:%d",
					string(event.Kv.Key), instance.Addr, instance.Port)

			case mvccpb.DELETE:
				if inst, ok := d.instances[string(event.Kv.Key)]; ok {
					log.Printf("Instance removed: %s -> %s:%d",
						string(event.Kv.Key), inst.Addr, inst.Port)
					delete(d.instances, string(event.Kv.Key))
				}
			}
		}
		d.mu.Unlock()
	}

	return fmt.Errorf("watch channel closed")
}

// GetInstances 获取所有可用实例（返回副本，避免并发修改）
func (d *Discovery) GetInstances() []*ServiceInstance {
	d.mu.RLock()
	defer d.mu.RUnlock()

	instances := make([]*ServiceInstance, 0, len(d.instances))
	for _, instance := range d.instances {
		// 创建副本，避免外部修改影响缓存
		instanceCopy := *instance
		instances = append(instances, &instanceCopy)
	}
	return instances
}

// GetInstanceCount 获取实例数量
func (d *Discovery) GetInstanceCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.instances)
}

// Stop 停止监听
func (d *Discovery) Stop() {
	if d.cancel != nil {
		d.cancel()
	}

	// 等待goroutine退出
	done := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Printf("Discovery stopped cleanly")
	case <-time.After(5 * time.Second):
		log.Printf("Warning: Discovery did not stop in time")
	}
}

// Close 关闭发现器
func (d *Discovery) Close() error {
	d.Stop()
	if d.client != nil {
		return d.client.Close()
	}
	return nil
}

// IsHealthy 检查是否健康
func (d *Discovery) IsHealthy() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.client != nil && len(d.instances) > 0
}

// GetStatus 获取状态详情
func (d *Discovery) GetStatus() map[string]interface{} {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return map[string]interface{}{
		"service_name":   d.serviceName,
		"instance_count": len(d.instances),
		"healthy":        d.IsHealthy(),
	}
}
