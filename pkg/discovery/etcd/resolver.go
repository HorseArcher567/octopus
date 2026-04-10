package etcd

// This file contains the etcd-backed implementation of discovery.Resolver.

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/HorseArcher567/octopus/pkg/discovery"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Resolver is a minimal etcd-backed discovery resolver.
type Resolver struct {
	log    *xlog.Logger
	client *clientv3.Client
}

// prefix returns the etcd key prefix used for the given service.
func prefix(service string) string {
	return fmt.Sprintf("/octopus/rpc/apps/%s/", service)
}

// Resolve returns the currently registered instances for service.
func (r *Resolver) Resolve(ctx context.Context, service string) ([]discovery.Instance, error) {
	resp, err := r.client.Get(ctx, prefix(service), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	instances := make([]discovery.Instance, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var ins discovery.Instance
		if err := json.Unmarshal(kv.Value, &ins); err != nil {
			continue
		}
		instances = append(instances, ins)
	}
	return instances, nil
}

// Watch emits refreshed instance snapshots for service until ctx is canceled
// or the underlying watch fails.
func (r *Resolver) Watch(ctx context.Context, service string) (<-chan []discovery.Instance, error) {
	ch := make(chan []discovery.Instance, 1)
	go func() {
		defer close(ch)
		for {
			instances, err := r.Resolve(ctx, service)
			if err != nil {
				return
			}
			select {
			case ch <- instances:
			case <-ctx.Done():
				return
			}
			watchCh := r.client.Watch(ctx, prefix(service), clientv3.WithPrefix())
			select {
			case <-ctx.Done():
				return
			case _, ok := <-watchCh:
				if !ok {
					return
				}
			}
		}
	}()
	return ch, nil
}
