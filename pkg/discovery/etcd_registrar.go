package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/HorseArcher567/octopus/pkg/xlog"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	etcdLeaseTTL = 60
	etcdTimeout  = 3 * time.Second
)

// EtcdRegistrar publishes one active service instance into etcd and keeps its
// lease alive until deregistration.
type EtcdRegistrar struct {
	log    *xlog.Logger
	client *clientv3.Client

	mu      sync.Mutex
	key     string
	leaseID clientv3.LeaseID
	cancel  context.CancelFunc
	done    chan struct{}
}

// NewEtcdRegistrar creates an etcd-backed registrar.
func NewEtcdRegistrar(log *xlog.Logger, client *clientv3.Client) *EtcdRegistrar {
	return &EtcdRegistrar{log: log, client: client}
}

func etcdKey(instance Instance) string {
	return fmt.Sprintf("/octopus/rpc/apps/%s/%s:%d", instance.Name, instance.Host, instance.Port)
}

// Register publishes instance into etcd with a leased key and starts lease keepalive.
func (r *EtcdRegistrar) Register(ctx context.Context, instance Instance) error {
	if instance.Name == "" || instance.Host == "" || instance.Port <= 0 {
		return fmt.Errorf("discovery: invalid instance")
	}

	key := etcdKey(instance)

	r.mu.Lock()
	if r.key != "" {
		r.mu.Unlock()
		return fmt.Errorf("discovery: registrar already has an active instance")
	}
	r.mu.Unlock()

	leaseCtx, cancel := context.WithTimeout(ctx, etcdTimeout)
	lease, err := r.client.Grant(leaseCtx, etcdLeaseTTL)
	cancel()
	if err != nil {
		return err
	}

	payload, err := json.Marshal(instance)
	if err != nil {
		return err
	}

	putCtx, putCancel := context.WithTimeout(ctx, etcdTimeout)
	_, err = r.client.Put(putCtx, key, string(payload), clientv3.WithLease(lease.ID))
	putCancel()
	if err != nil {
		_, _ = r.client.Revoke(context.Background(), lease.ID)
		return err
	}

	keepaliveCtx, keepaliveCancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	r.mu.Lock()
	r.key = key
	r.leaseID = lease.ID
	r.cancel = keepaliveCancel
	r.done = done
	r.mu.Unlock()

	go r.keepalive(keepaliveCtx, key, lease.ID, done)
	return nil
}

// Deregister removes the active instance from etcd and stops its lease keepalive.
func (r *EtcdRegistrar) Deregister(ctx context.Context, instance Instance) error {
	key := etcdKey(instance)

	r.mu.Lock()
	activeKey := r.key
	leaseID := r.leaseID
	cancel := r.cancel
	done := r.done
	if activeKey != "" && activeKey == key {
		r.key = ""
		r.leaseID = 0
		r.cancel = nil
		r.done = nil
	}
	r.mu.Unlock()

	if activeKey == "" {
		return nil
	}
	if activeKey != key {
		return fmt.Errorf("discovery: active instance does not match deregistration target")
	}

	if cancel != nil {
		cancel()
	}
	if done != nil {
		select {
		case <-done:
		case <-time.After(etcdTimeout):
		}
	}

	delCtx, delCancel := context.WithTimeout(ctx, etcdTimeout)
	defer delCancel()
	if _, err := r.client.Delete(delCtx, key); err != nil {
		return err
	}

	revokeCtx, revokeCancel := context.WithTimeout(context.Background(), etcdTimeout)
	_, _ = r.client.Revoke(revokeCtx, leaseID)
	revokeCancel()
	return nil
}

func (r *EtcdRegistrar) keepalive(ctx context.Context, key string, leaseID clientv3.LeaseID, done chan struct{}) {
	defer close(done)

	for {
		ch, err := r.client.KeepAlive(ctx, leaseID)
		if err != nil {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
				continue
			}
		}

		for {
			select {
			case <-ctx.Done():
				return
			case resp, ok := <-ch:
				if !ok || resp == nil {
					select {
					case <-ctx.Done():
						return
					case <-time.After(time.Second):
					}
					goto retry
				}
			}
		}

	retry:
		r.log.Warn("discovery: etcd keepalive restarted", "key", key)
	}
}
