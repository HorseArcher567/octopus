package rpc

import (
	"context"
	"fmt"
	"net"

	"github.com/HorseArcher567/octopus/pkg/rpc/registry"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server wraps a gRPC server with service configuration, logging, and registry support.
type Server struct {
	log    *xlog.Logger
	config *ServerConfig

	grpcServer  *grpc.Server
	grpcOptions []grpc.ServerOption

	registry   *registry.Registry
	etcdClient *clientv3.Client
}

// MustNewServer creates a new Server and panics if initialization fails.
func MustNewServer(log *xlog.Logger, config *ServerConfig, opts ...Option) *Server {
	server, err := NewServer(log, config, opts...)
	if err != nil {
		panic(err)
	}
	return server
}

// NewServer creates a new RPC Server.
// Functional options can be used to configure the logger, etcd integration, and
// underlying grpc.Server behavior.
func NewServer(log *xlog.Logger, config *ServerConfig, opts ...Option) (*Server, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	s := &Server{
		log:    log,
		config: config,
	}

	for _, opt := range opts {
		opt(s)
	}

	s.grpcServer = grpc.NewServer(s.grpcOptions...)
	return s, nil
}

// RegisterServices registers one or more gRPC services on the underlying grpc.Server.
// It can be called multiple times to register different services.
func (s *Server) RegisterServices(registerFunc func(*grpc.Server)) {
	registerFunc(s.grpcServer)
}

// Start starts the gRPC server in a background goroutine and returns immediately.
// It enables reflection if configured and registers the service to etcd if configured.
// If the server fails to start, it will panic (to fail-fast during startup).
// Use Stop to gracefully shut down the server.
func (s *Server) Start() error {
	// Enable reflection if configured.
	if s.config.EnableReflection {
		reflection.Register(s.grpcServer)
		s.log.Info("grpc reflection enabled")
	}

	// Create listener.
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		s.log.Error("failed to listen", "error", err)
		return err
	}

	s.log.Info("starting rpc server", "addr", addr)

	// Start gRPC server in background.
	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			s.log.Error("rpc server stopped unexpectedly", "error", err)
			panic(err) // Fail fast if server crashes unexpectedly
		}
	}()

	// Register service instance to etcd, if configured.
	if s.config.ShouldRegisterInstance() {
		if err := s.registerInstance(); err != nil {
			s.log.Error("failed to register service", "error", err)
			return err
		}
	}

	return nil
}

// Stop gracefully stops the server and unregisters it from the registry if present.
// It blocks until the server has finished shutting down.
func (s *Server) Stop(ctx context.Context) error {
	s.log.Info("shutting down rpc server gracefully")

	if s.registry != nil {
		s.registry.Unregister()
	}

	// GracefulStop will block until all connections are closed or ctx is cancelled
	done := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		s.log.Info("rpc server shutdown complete")
		return nil
	case <-ctx.Done():
		s.log.Warn("rpc server shutdown timeout, forcing stop")
		s.grpcServer.Stop() // Force stop
		return ctx.Err()
	}
}

// registerInstance registers the service instance to the service registry (etcd) using the provided configuration.
func (s *Server) registerInstance() error {
	instance := &registry.Instance{
		Name: s.config.Name,
		Addr: s.config.AdvertiseAddr,
		Port: s.config.Port,
	}

	// Create a context that carries the logger.
	r, err := registry.NewRegistry(s.log, s.etcdClient, instance)
	if err != nil {
		return err
	}

	r.Register()
	s.registry = r
	s.log.Info("instance registered to etcd", "instance", instance)

	return nil
}
