package etcd

// This file contains the etcd-backed implementation of discovery.Registrar.

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/HorseArcher567/octopus/pkg/discovery"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	leaseTTL = 60
	timeout  = 3 * time.Second
)

// Registrar is a minimal etcd-backed registrar.
type Registrar struct {
	log    *xlog.Logger
	client *clientv3.Client
}

// key returns the etcd key used for ins.
func key(ins discovery.Instance) string {
	return fmt.Sprintf("/octopus/rpc/apps/%s/%s:%d", ins.Service, ins.Address, ins.Port)
}

// Register publishes ins into etcd with a leased key.
func (r *Registrar) Register(ctx context.Context, ins discovery.Instance) error {
	if ins.Service == "" || ins.Address == "" || ins.Port <= 0 {
		return fmt.Errorf("discovery/etcd: invalid instance")
	}
	leaseCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	lease, err := r.client.Grant(leaseCtx, leaseTTL)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(ins)
	if err != nil {
		return err
	}
	putCtx, putCancel := context.WithTimeout(ctx, timeout)
	defer putCancel()
	_, err = r.client.Put(putCtx, key(ins), string(payload), clientv3.WithLease(lease.ID))
	return err
}

// Deregister removes ins from etcd.
func (r *Registrar) Deregister(ctx context.Context, ins discovery.Instance) error {
	delCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	_, err := r.client.Delete(delCtx, key(ins))
	return err
}
