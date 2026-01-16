// Package registry provides service registration functionality for etcd.
// It handles service instance registration, lease management, and keepalive.
package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/HorseArcher567/octopus/pkg/xlog"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// leaseTTL is the default lease time-to-live in seconds.
const leaseTTL = 60

// timeout is the default timeout for etcd operations.
const timeout = 3 * time.Second

// Registry handles service instance registration and keepalive with etcd.
type Registry struct {
	// log is the logger for logging operations.
	log *xlog.Logger
	// etcdClient is the etcd client for registration operations.
	etcdClient *clientv3.Client
	// instance is the service instance to register.
	instance *Instance

	// ctx is the context for cancellation.
	ctx context.Context
	// cancel is the cancel function to stop operations.
	cancel context.CancelFunc

	// done is closed when the register goroutine has stopped.
	done chan struct{}
}

// NewRegistry creates a new Registry instance.
// The instance must be valid (pass Validate()).
func NewRegistry(log *xlog.Logger, etcdClient *clientv3.Client, instance *Instance) (*Registry, error) {
	if err := instance.Validate(); err != nil {
		return nil, fmt.Errorf("invalid service instance: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Registry{
		log:        log,
		instance:   instance,
		etcdClient: etcdClient,
		ctx:        ctx,
		cancel:     cancel,
		done:       make(chan struct{}),
	}, nil
}

// Register registers the service instance with etcd and starts keepalive.
// It creates a lease, stores the instance information, and starts a background
// goroutine to maintain the lease.
//
// Register returns immediately after starting the registration goroutine.
func (r *Registry) Register() {
	// Start keepalive goroutine.
	go r.register()
}

// Unregister unregisters the service instance and stops keepalive.
// It waits for the keepalive goroutine to stop (with timeout) and revokes the lease.
// Unregister does not wait for the operation to complete if it times out.
func (r *Registry) Unregister() {
	r.log.Info("start unregistering service")
	if r.cancel != nil {
		r.cancel()
	}

	select {
	case <-r.done:
		return
	case <-time.After(timeout):
		r.log.Error("failed to unregister service", "error", "timeout")
	}
}

// register creates a lease and registers the service instance to etcd.
// If registration fails, it cleans up the created lease.
// It retries with exponential backoff until the context is canceled.
func (r *Registry) register() {
	defer close(r.done)

	backoff := time.Second
	maxBackoff := timeout

	for {
		leaseID, err := r.grantLease()
		if err == nil {
			if err = r.putInstance(leaseID); err == nil {
				// Keepalive blocks until it fails or context is canceled.
				r.keepalive(leaseID)
				backoff = time.Second
			}
			// Always revoke lease after keepalive or failed registration.
			r.revokeLease(leaseID)
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

// grantLease grants a new lease from etcd with the default TTL.
// It returns the lease ID on success, or an error if the grant fails.
func (r *Registry) grantLease() (clientv3.LeaseID, error) {
	ctx, cancel := context.WithTimeout(r.ctx, timeout)
	defer cancel()

	if resp, err := r.etcdClient.Grant(ctx, leaseTTL); err != nil {
		r.log.Error("failed to grant lease", "error", err)
		return 0, err
	} else {
		return resp.ID, nil
	}
}

// putInstance stores the service instance to etcd with the given lease ID.
// It serializes the instance information and stores it at the instance's key.
func (r *Registry) putInstance(leaseID clientv3.LeaseID) error {
	data, err := json.Marshal(r.instance)
	if err != nil {
		r.log.Error("failed to marshal instance", "instance", r.instance, "error", err)
		return err
	}

	ctx, cancel := context.WithTimeout(r.ctx, timeout)
	defer cancel()

	_, err = r.etcdClient.Put(ctx, r.instance.Key(), string(data), clientv3.WithLease(leaseID))
	if err != nil {
		r.log.Error("failed to put service", "error", err)
		return err
	}

	return nil
}

// keepalive maintains the lease by sending keepalive requests to etcd.
// It blocks until the keepalive channel is closed, an error occurs, or the context is canceled.
// The leaseID is passed as a parameter to avoid frequent locking to read r.leaseID.
func (r *Registry) keepalive(leaseID clientv3.LeaseID) {
	kaChannel, err := r.etcdClient.KeepAlive(r.ctx, leaseID)
	if err != nil {
		r.log.Error("failed to start keepalive", "error", err)
		return
	}

	for {
		select {
		case resp, ok := <-kaChannel:
			if !ok {
				r.log.Info("keepalive channel closed")
				return
			} else if resp == nil {
				r.log.Error("keepalive response is nil")
				return
			}
		case <-r.ctx.Done():
			r.log.Info("keepalive stopped", "error", r.ctx.Err())
			return
		}
	}
}

// revokeLease revokes the specified lease.
func (r *Registry) revokeLease(leaseID clientv3.LeaseID) {
	revokeCtx, cancel := context.WithTimeout(r.ctx, timeout)
	defer cancel()

	_, err := r.etcdClient.Revoke(revokeCtx, leaseID)
	if err != nil {
		r.log.Error("failed to revoke lease", "error", err)
	}
}
