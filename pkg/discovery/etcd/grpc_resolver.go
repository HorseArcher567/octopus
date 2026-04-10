package etcd

// This file contains the gRPC resolver adapter built on top of the discovery
// resolver abstraction.

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/HorseArcher567/octopus/pkg/discovery"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	grpcresolver "google.golang.org/grpc/resolver"
)

// GRPCResolverBuilder adapts the etcd discovery resolver to gRPC resolver.Builder.
type GRPCResolverBuilder struct {
	log      *xlog.Logger
	resolver discovery.Resolver
}

// NewGRPCResolverBuilder creates a gRPC resolver builder backed by resolver.
func NewGRPCResolverBuilder(log *xlog.Logger, resolver discovery.Resolver) *GRPCResolverBuilder {
	return &GRPCResolverBuilder{log: log, resolver: resolver}
}

// Scheme reports the gRPC resolver scheme handled by the builder.
func (b *GRPCResolverBuilder) Scheme() string { return "etcd" }

// Build creates a gRPC resolver for target.
func (b *GRPCResolverBuilder) Build(target grpcresolver.Target, cc grpcresolver.ClientConn, _ grpcresolver.BuildOptions) (grpcresolver.Resolver, error) {
	r := &grpcResolver{
		log:      b.log,
		resolver: b.resolver,
		cc:       cc,
		service:  target.Endpoint(),
		done:     make(chan struct{}),
	}
	r.ctx, r.cancel = context.WithCancel(context.Background())
	go r.watch()
	return r, nil
}

// grpcResolver adapts discovery updates into gRPC resolver state updates.
type grpcResolver struct {
	log      *xlog.Logger
	resolver discovery.Resolver
	cc       grpcresolver.ClientConn
	service  string
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.RWMutex
	addrs    []grpcresolver.Address
	done     chan struct{}
}

// ResolveNow triggers an immediate refresh.
func (r *grpcResolver) ResolveNow(grpcresolver.ResolveNowOptions) {
	_ = r.reload()
}

// Close stops background resolution.
func (r *grpcResolver) Close() {
	if r.cancel != nil {
		r.cancel()
	}
	select {
	case <-r.done:
	case <-time.After(100 * time.Millisecond):
	}
}

// watch forwards discovery watch updates into gRPC state updates.
func (r *grpcResolver) watch() {
	defer close(r.done)
	_ = r.reload()
	ch, err := r.resolver.Watch(r.ctx, r.service)
	if err != nil {
		return
	}
	for {
		select {
		case <-r.ctx.Done():
			return
		case instances, ok := <-ch:
			if !ok {
				return
			}
			r.update(instances)
		}
	}
}

// reload resolves the latest instance snapshot.
func (r *grpcResolver) reload() error {
	instances, err := r.resolver.Resolve(r.ctx, r.service)
	if err != nil {
		return err
	}
	r.update(instances)
	return nil
}

// update converts discovery instances into gRPC resolver addresses.
func (r *grpcResolver) update(instances []discovery.Instance) {
	addrs := make([]grpcresolver.Address, 0, len(instances))
	for _, ins := range instances {
		addrs = append(addrs, grpcresolver.Address{
			Addr:     fmt.Sprintf("%s:%d", ins.Address, ins.Port),
			Metadata: ins,
		})
	}
	r.mu.Lock()
	r.addrs = addrs
	r.mu.Unlock()
	r.cc.UpdateState(grpcresolver.State{Addresses: addrs})
}
