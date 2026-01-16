package rpc

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

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

// Start starts the gRPC server and blocks until a shutdown signal is received
// and the server has been gracefully stopped.
func (s *Server) Start() {
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
		panic(err)
	}

	// Start gRPC server.
	s.log.Info("starting rpc server", "addr", addr)
	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			s.log.Error("server stopped", "error", err)
			panic(err)
		}
	}()

	// Register service instance to etcd, if configured.
	if s.config.ShouldRegisterInstance() {
		if err := s.registerInstance(); err != nil {
			s.log.Error("failed to register service", "error", err)
			panic(err)
		}
	}

	// Wait for shutdown signal.
	s.waitForShutdown()
}

// registerInstance registers the service instance to the service registry (etcd) using the provided configuration.
func (s *Server) registerInstance() error {
	instance := &registry.Instance{
		Name: s.config.Name,
		Addr: s.config.AdvertiseAddr,
		Port: s.config.Port,
	}

	// Create a context that carries the logger.
	reg, err := registry.NewRegistry(s.log, s.etcdClient, instance)
	if err != nil {
		return err
	}

	reg.Register()
	s.registry = reg
	s.log.Info("instance registered to etcd", "instance", instance)

	return nil
}

// waitForShutdown blocks until a termination signal is received and then
// performs a graceful shutdown sequence.
func (s *Server) waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan

	s.log.Info("shutting down gracefully")

	s.Stop()

	s.log.Info("shutdown complete")
}

// Stop gracefully stops the server and unregisters it from the registry if present.
func (s *Server) Stop() {
	if s.registry != nil {
		s.registry.Unregister()
	}

	s.grpcServer.GracefulStop()
}
