package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/HorseArcher567/octopus/pkg/xlog"
	clientv3 "go.etcd.io/etcd/client/v3"
	grpcresolver "google.golang.org/grpc/resolver"
)

// EtcdResolver exposes an etcd-backed gRPC resolver builder.
type EtcdResolver struct {
	log    *xlog.Logger
	client *clientv3.Client
}

// NewEtcdResolver creates an etcd-backed resolver.
func NewEtcdResolver(log *xlog.Logger, client *clientv3.Client) *EtcdResolver {
	return &EtcdResolver{log: log, client: client}
}

// Builder returns a gRPC resolver builder for etcd:/// targets.
func (r *EtcdResolver) Builder() grpcresolver.Builder {
	return &etcdGRPCResolverBuilder{log: r.log, client: r.client}
}

type etcdGRPCResolverBuilder struct {
	log    *xlog.Logger
	client *clientv3.Client
}

func (b *etcdGRPCResolverBuilder) Scheme() string { return "etcd" }

func (b *etcdGRPCResolverBuilder) Build(target grpcresolver.Target, cc grpcresolver.ClientConn, _ grpcresolver.BuildOptions) (grpcresolver.Resolver, error) {
	r := &etcdGRPCResolver{
		log:       b.log,
		client:    b.client,
		cc:        cc,
		target:    target.Endpoint(),
		addresses: make(map[string]grpcresolver.Address),
		done:      make(chan struct{}),
	}
	r.ctx, r.cancel = context.WithCancel(context.Background())
	go r.watch()
	return r, nil
}

type etcdGRPCResolver struct {
	log    *xlog.Logger
	client *clientv3.Client
	cc     grpcresolver.ClientConn
	target string

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}

	mu        sync.RWMutex
	addresses map[string]grpcresolver.Address
}

func (r *etcdGRPCResolver) ResolveNow(grpcresolver.ResolveNowOptions) {
	_ = r.reload()
}

func (r *etcdGRPCResolver) Close() {
	if r.cancel != nil {
		r.cancel()
	}
	select {
	case <-r.done:
	case <-time.After(100 * time.Millisecond):
	}
}

func (r *etcdGRPCResolver) watch() {
	defer close(r.done)
	_ = r.reload()
	for {
		watchCh := r.client.Watch(r.ctx, etcdPrefix(r.target), clientv3.WithPrefix())
		for resp := range watchCh {
			if err := resp.Err(); err != nil {
				break
			}
			_ = r.reload()
		}
		select {
		case <-r.ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}

func (r *etcdGRPCResolver) reload() error {
	resp, err := r.client.Get(r.ctx, etcdPrefix(r.target), clientv3.WithPrefix())
	if err != nil {
		return err
	}
	addresses := make(map[string]grpcresolver.Address, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		key := string(kv.Key)
		var instance Instance
		if err := json.Unmarshal(kv.Value, &instance); err != nil {
			continue
		}
		addresses[key] = grpcresolver.Address{Addr: instance.Addr(), Metadata: instance}
	}
	r.mu.Lock()
	r.addresses = addresses
	r.mu.Unlock()
	r.updateState()
	return nil
}

func (r *etcdGRPCResolver) updateState() {
	r.mu.RLock()
	addrs := make([]grpcresolver.Address, 0, len(r.addresses))
	for _, addr := range r.addresses {
		addrs = append(addrs, addr)
	}
	r.mu.RUnlock()
	r.cc.UpdateState(grpcresolver.State{Addresses: addrs})
}

// EtcdDiscovery combines etcd-backed registration and resolution.
type EtcdDiscovery struct {
	registrar *EtcdRegistrar
	resolver  *EtcdResolver
}

// NewEtcdDiscovery creates an etcd-backed discovery implementation.
func NewEtcdDiscovery(log *xlog.Logger, client *clientv3.Client) *EtcdDiscovery {
	return &EtcdDiscovery{
		registrar: NewEtcdRegistrar(log, client),
		resolver:  NewEtcdResolver(log, client),
	}
}

func (d *EtcdDiscovery) Register(ctx context.Context, instance Instance) error {
	return d.registrar.Register(ctx, instance)
}

func (d *EtcdDiscovery) Deregister(ctx context.Context, instance Instance) error {
	return d.registrar.Deregister(ctx, instance)
}

func (d *EtcdDiscovery) Builder() grpcresolver.Builder {
	return d.resolver.Builder()
}

func etcdPrefix(name string) string {
	return fmt.Sprintf("/octopus/rpc/apps/%s/", name)
}
