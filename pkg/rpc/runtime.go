// Package rpc provides Octopus gRPC server and client runtime integration.
package rpc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/HorseArcher567/octopus/pkg/discovery"
	discoveryetcd "github.com/HorseArcher567/octopus/pkg/discovery/etcd"
	"github.com/HorseArcher567/octopus/pkg/etcd"
	"github.com/HorseArcher567/octopus/pkg/rpc/resolver"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	grpcresolver "google.golang.org/grpc/resolver"
)

// RuntimeConfig describes the full RPC runtime shape used by app assembly.
type RuntimeConfig struct {
	Server        *ServerConfig
	ClientOptions *ClientOptions
	Etcd          *etcd.Config
}

// ClientFactory owns reusable gRPC client connections.
//
// Connections are cached by target and may integrate with discovery-backed
// resolvers depending on runtime configuration.
type ClientFactory struct {
	log               *xlog.Logger
	dialOpts          []grpc.DialOption
	etcdClient        *clientv3.Client
	discoveryProvider discovery.Provider

	mu      sync.Mutex
	clients map[string]*grpc.ClientConn
}

// NewClientFactory builds a client factory with optional discovery support.
func NewClientFactory(log *xlog.Logger, cfg *ClientOptions, etcdClient *clientv3.Client, discoveryProvider discovery.Provider) *ClientFactory {
	if log == nil {
		log = xlog.MustNew(nil)
	}
	if cfg == nil {
		cfg = &ClientOptions{}
	}
	return &ClientFactory{
		log:               log,
		dialOpts:          cfg.BuildDialOptions(),
		etcdClient:        etcdClient,
		discoveryProvider: discoveryProvider,
		clients:           make(map[string]*grpc.ClientConn),
	}
}

// Client returns a cached connection for target, creating it on first use.
func (f *ClientFactory) Client(target string) (*grpc.ClientConn, error) {
	f.mu.Lock()
	if conn, ok := f.clients[target]; ok {
		f.mu.Unlock()
		f.log.Debug("reuse rpc client connection", "target", target)
		return conn, nil
	}
	f.mu.Unlock()

	dialOpts := append([]grpc.DialOption{}, f.dialOpts...)
	switch {
	case strings.HasPrefix(target, "etcd:///"):
		if f.discoveryProvider == nil {
			if f.etcdClient == nil {
				return nil, fmt.Errorf("rpc: etcd client is not configured for target %q", target)
			}
			dialOpts = append(dialOpts, grpc.WithResolvers(resolver.NewEtcdBuilder(f.log, f.etcdClient)))
			break
		}
		if builder, ok := f.discoveryProvider.GRPCResolverBuilder().(grpcresolver.Builder); ok {
			dialOpts = append(dialOpts, grpc.WithResolvers(builder))
		} else {
			return nil, fmt.Errorf("rpc: discovery provider does not expose a gRPC resolver builder")
		}
	case strings.HasPrefix(target, "direct:///"):
		dialOpts = append(dialOpts, grpc.WithResolvers(resolver.NewDirectBuilder(f.log)))
	}

	conn, err := NewClient(target, dialOpts...)
	if err != nil {
		return nil, err
	}

	f.mu.Lock()
	if existing, ok := f.clients[target]; ok {
		f.mu.Unlock()
		_ = conn.Close()
		f.log.Debug("reuse rpc client connection after race", "target", target)
		return existing, nil
	}
	f.clients[target] = conn
	f.mu.Unlock()
	f.log.Info("created rpc client connection", "target", target)
	return conn, nil
}

// Close closes all cached client connections.
func (f *ClientFactory) Close() error {
	start := time.Now()
	f.mu.Lock()
	clients := f.clients
	f.clients = make(map[string]*grpc.ClientConn)
	f.mu.Unlock()

	if len(clients) == 0 {
		f.log.Debug("no rpc clients to close")
		return nil
	}

	var errs []error
	for target, conn := range clients {
		if err := conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close rpc client %q: %w", target, err))
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	f.log.Info("closed rpc clients", "count", len(clients), "duration", time.Since(start))
	return nil
}

// Runtime owns both the inbound RPC server and outbound client factory.
//
// Runtime may also own discovery and etcd resources needed by server
// registration and client-side resolution.
type Runtime struct {
	log               *xlog.Logger
	server            *Server
	clients           *ClientFactory
	etcdClient        *clientv3.Client
	discoveryProvider discovery.Provider
	ownsEtcdConn      bool
}

// NewRuntime builds a full RPC runtime from declarative config.
func NewRuntime(log *xlog.Logger, cfg *RuntimeConfig, serverOpts ...Option) (*Runtime, error) {
	if log == nil {
		log = xlog.MustNew(nil)
	}
	if cfg == nil {
		cfg = &RuntimeConfig{}
	}

	rt := &Runtime{log: log}
	if cfg.Etcd != nil {
		etcdClient, err := etcd.NewClient(cfg.Etcd)
		if err != nil {
			return nil, err
		}
		rt.etcdClient = etcdClient
		rt.discoveryProvider = discoveryetcd.NewProvider(log, etcdClient)
		rt.ownsEtcdConn = true
	}

	rt.clients = NewClientFactory(log, cfg.ClientOptions, rt.etcdClient, rt.discoveryProvider)

	if cfg.Server != nil {
		opts := make([]Option, 0, len(serverOpts)+2)
		opts = append(opts, WithEtcdClient(rt.etcdClient))
		if rt.discoveryProvider != nil {
			opts = append(opts, WithRegistrar(rt.discoveryProvider.Registrar()))
		}
		opts = append(opts, serverOpts...)
		server, err := NewServer(log, cfg.Server, opts...)
		if err != nil {
			_ = rt.Close()
			return nil, err
		}
		rt.server = server
	}

	return rt, nil
}

// Register applies registration to the inbound server when it exists.
func (r *Runtime) Register(register func(*grpc.Server)) error {
	if r.server == nil {
		return errors.New("rpc runtime: server is not initialized")
	}
	return r.server.Register(register)
}

// Client resolves an outbound RPC client connection.
func (r *Runtime) Client(target string) (*grpc.ClientConn, error) {
	if r.clients == nil {
		return nil, errors.New("rpc runtime: client factory is not initialized")
	}
	return r.clients.Client(target)
}

// CloseClients closes all cached outbound connections.
func (r *Runtime) CloseClients() error {
	if r.clients == nil {
		return nil
	}
	return r.clients.Close()
}

// Run starts the inbound RPC server when configured.
func (r *Runtime) Run(ctx context.Context) error {
	if r.server == nil {
		return nil
	}
	return r.server.Run(ctx)
}

// Stop stops the inbound RPC server when configured.
func (r *Runtime) Stop(ctx context.Context) error {
	if r.server == nil {
		return nil
	}
	return r.server.Stop(ctx)
}

// Close releases outbound clients and runtime-owned etcd resources.
func (r *Runtime) Close() error {
	var errs []error
	if err := r.CloseClients(); err != nil {
		errs = append(errs, err)
	}
	if r.discoveryProvider != nil {
		if err := r.discoveryProvider.Close(); err != nil {
			errs = append(errs, err)
		}
		r.discoveryProvider = nil
	}
	if r.ownsEtcdConn && r.etcdClient != nil {
		if err := r.etcdClient.Close(); err != nil {
			errs = append(errs, err)
		}
		r.etcdClient = nil
	}
	return errors.Join(errs...)
}
