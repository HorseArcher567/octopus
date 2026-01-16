// Package resolver provides gRPC resolver implementations for service discovery.
// It includes an etcd-based resolver that watches etcd for service instance changes.
package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/HorseArcher567/octopus/pkg/rpc/registry"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	grpcresolver "google.golang.org/grpc/resolver"
)

// EtcdResolverBuilder implements the gRPC resolver.Builder interface
// for etcd-based service discovery.
type EtcdResolverBuilder struct {
	// log is the logger for logging operations.
	log *xlog.Logger
	// etcdClient is the etcd client for service discovery operations.
	etcdClient *clientv3.Client
}

// NewEtcdBuilder creates a new EtcdResolverBuilder.
// The etcdClient must be a valid etcd client instance.
func NewEtcdBuilder(log *xlog.Logger, etcdClient *clientv3.Client) *EtcdResolverBuilder {
	return &EtcdResolverBuilder{
		log:        log,
		etcdClient: etcdClient,
	}
}

// Scheme returns the resolver scheme used in gRPC target URLs.
// It returns SchemeEtcd.
func (b *EtcdResolverBuilder) Scheme() string {
	return SchemeEtcd
}

// Build creates a new resolver instance for the given target.
// It starts a background goroutine to watch for service instance changes.
// Build returns the resolver and nil error on success.
func (b *EtcdResolverBuilder) Build(target grpcresolver.Target, cc grpcresolver.ClientConn,
	opts grpcresolver.BuildOptions) (grpcresolver.Resolver, error) {
	r := &etcdResolver{
		log:        b.log,
		etcdClient: b.etcdClient,
		cc:         cc,
		target:     target,
		addresses:  make(map[string]grpcresolver.Address),
		done:       make(chan struct{}),
	}

	r.ctx, r.cancel = context.WithCancel(context.Background())
	go r.startWatch()

	return r, nil
}

// etcdResolver implements the gRPC resolver.Resolver interface
// for etcd-based service discovery.
type etcdResolver struct {
	// log is the logger for logging operations.
	log *xlog.Logger
	// etcdClient is the etcd client for service discovery operations.
	etcdClient *clientv3.Client

	// ctx is the context for cancellation.
	ctx context.Context
	// cancel is the cancel function to stop operations.
	cancel context.CancelFunc

	// cc is the gRPC client connection to update with discovered addresses.
	cc grpcresolver.ClientConn
	// target is the gRPC target to resolve.
	target grpcresolver.Target

	// mu protects the addresses map.
	mu sync.RWMutex
	// addresses maps etcd keys to gRPC addresses.
	addresses map[string]grpcresolver.Address

	// done is closed when the watch goroutine has stopped.
	done chan struct{}
}

// ResolveNow triggers an immediate resolution of the target.
// It reloads service instances from etcd and updates the client connection state.
func (r *etcdResolver) ResolveNow(grpcresolver.ResolveNowOptions) {
	if err := r.loadInstances(); err != nil {
		r.log.Error("resolve now failed", "error", err)
	}
}

// Close stops the resolver and releases associated resources.
// It cancels the watch goroutine and stops monitoring etcd for changes.
// Close returns immediately after canceling the context, without waiting
// for the watch goroutine to exit. The goroutine will exit when it detects
// the context cancellation.
func (r *etcdResolver) Close() {
	if r.cancel != nil {
		r.cancel()
	}
	// Wait briefly for the goroutine to exit, but don't block indefinitely.
	// This helps ensure clean shutdown, but Close() should return quickly.
	select {
	case <-r.done:
		r.log.Info("resolver closed, watch goroutine stopped")
	case <-time.After(100 * time.Millisecond):
		r.log.Info("resolver closed, watch goroutine may still be running")
	}
}

// startWatch starts the watch loop that monitors etcd for service changes.
// It handles retries with exponential backoff on errors.
// The watch loop continues until the context is canceled.
func (r *etcdResolver) startWatch() {
	defer close(r.done)

	if err := r.loadInstances(); err != nil {
		r.log.Error("failed to load instances", "error", err)
	}

	backoff := time.Second
	maxBackoff := 8 * time.Second
	// healthyThreshold is the minimum duration a watch must run to be considered healthy.
	// If a watch runs longer than this, we reset the backoff on the next retry.
	healthyThreshold := 15 * time.Second

	for {
		start := time.Now()
		r.doWatch()
		duration := time.Since(start)

		if duration >= healthyThreshold {
			// Reset backoff to initial value after a healthy watch duration.
			// This allows quick recovery after temporary network issues.
			backoff = time.Second
		}

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

// doWatch performs a single watch operation on etcd.
// It returns when the watch channel is closed or an error occurs.
// Before starting a new watch, it reloads all instances to ensure state consistency
// after a watch interruption.
//
// If loadInstances fails, doWatch returns immediately to avoid establishing
// a watch with incomplete state. Watch only provides incremental updates,
// so without the initial full state, we may miss existing instances and
// cause load balancing to fail or services to be unreachable.
func (r *etcdResolver) doWatch() {
	if err := r.loadInstances(); err != nil {
		r.log.Error("failed to load instances before watch", "error", err)
		return
	}

	watchChan := r.etcdClient.Watch(r.ctx, r.prefix(), clientv3.WithPrefix())

	for watchResp := range watchChan {
		if err := watchResp.Err(); err != nil {
			r.log.Error("watch error", "error", err)
			break
		}

		r.handleEvents(watchResp.Events)

		r.updateState()
	}
}

// loadInstances loads all existing service instances from etcd.
// It replaces the internal address map with the current state from etcd.
func (r *etcdResolver) loadInstances() error {
	resp, err := r.etcdClient.Get(r.ctx, r.prefix(), clientv3.WithPrefix())
	if err != nil {
		r.log.Error("failed to get instances from etcd", "error", err)
		return err
	}

	addresses := make(map[string]grpcresolver.Address, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		key := string(kv.Key)
		var instance registry.Instance
		if err := json.Unmarshal(kv.Value, &instance); err != nil {
			r.log.Error("failed to unmarshal instance",
				"error", err, "key", key)
			continue
		}

		addresses[key] = grpcresolver.Address{
			Addr:     r.formatAddr(&instance),
			Metadata: &instance,
		}
	}

	r.mu.Lock()
	r.addresses = addresses
	r.mu.Unlock()

	r.updateState()
	return nil
}

// handleEvents processes multiple etcd watch events.
// It updates the internal address map based on PUT and DELETE events.
// Events are processed in batch to minimize lock contention.
func (r *etcdResolver) handleEvents(events []*clientv3.Event) {
	for _, event := range events {
		key := string(event.Kv.Key)

		switch event.Type {
		case mvccpb.PUT:
			var instance registry.Instance
			if err := json.Unmarshal(event.Kv.Value, &instance); err != nil {
				r.log.Error("failed to unmarshal instance",
					"error", err, "key", key)
				continue
			}
			addr := r.formatAddr(&instance)
			r.mu.Lock()
			r.addresses[key] = grpcresolver.Address{
				Addr:     addr,
				Metadata: &instance,
			}
			r.mu.Unlock()
			r.log.Info("service instance added", "addr", addr, "key", key)

		case mvccpb.DELETE:
			r.mu.Lock()
			if addr, ok := r.addresses[key]; ok {
				delete(r.addresses, key)
				r.mu.Unlock()
				r.log.Info("service instance removed", "addr", addr.Addr, "key", key)
			} else {
				r.mu.Unlock()
			}
		}
	}
}

// updateState updates the gRPC client connection state with the current
// set of discovered service instances.
// It reads the address map under a read lock and calls UpdateState on the client connection.
func (r *etcdResolver) updateState() {
	r.mu.RLock()
	addrs := make([]grpcresolver.Address, 0, len(r.addresses))
	for _, addr := range r.addresses {
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
			"count", len(addrs), "instances", addrList)
	}

	r.cc.UpdateState(grpcresolver.State{Addresses: addrs})
}

// prefix returns the etcd key prefix for the target service.
// The prefix is constructed from the target endpoint.
func (r *etcdResolver) prefix() string {
	return fmt.Sprintf("/octopus/rpc/apps/%s/", r.target.Endpoint())
}

// formatAddr formats an instance address as "host:port".
// It combines the instance's Addr and Port fields.
func (r *etcdResolver) formatAddr(instance *registry.Instance) string {
	return fmt.Sprintf("%s:%d", instance.Addr, instance.Port)
}
