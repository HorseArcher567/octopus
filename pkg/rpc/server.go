package rpc

import (
	"context"
	"fmt"
	"net"

	"github.com/HorseArcher567/octopus/pkg/discovery"
	"github.com/HorseArcher567/octopus/pkg/rpc/middleware"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/stats"
)

// Server wraps a gRPC server with service configuration, logging, and
// discovery-backed registration support.
type Server struct {
	log    *xlog.Logger
	config *ServerConfig

	grpcServer         *grpc.Server
	serverOptions      []grpc.ServerOption
	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
	statsHandlers      []stats.Handler

	registrar discovery.Registrar
	instance  *discovery.Instance
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
//
// Functional options can be used to configure discovery integration,
// interceptors, stats handlers, and underlying grpc.Server behavior.
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

	// Configure keepalive if configured
	keepaliveOpts := s.config.Keepalive.BuildServerOptions()
	if len(keepaliveOpts) > 0 {
		// Log keepalive configuration
		// Note: keepaliveOpts is non-empty only if s.config.Keepalive is not nil
		if s.config.Keepalive.ServerParameters != nil {
			sp := s.config.Keepalive.ServerParameters
			logFields := []any{
				"maxConnectionIdle", formatDuration(sp.MaxConnectionIdle),
				"maxConnectionAge", formatDuration(sp.MaxConnectionAge),
				"maxConnectionAgeGrace", formatDuration(sp.MaxConnectionAgeGrace),
				"time", formatDuration(sp.Time),
				"timeout", formatDuration(sp.Timeout),
			}
			s.log.Info("configuring keepalive server parameters", logFields...)
		}
		if s.config.Keepalive.EnforcementPolicy != nil {
			ep := s.config.Keepalive.EnforcementPolicy
			s.log.Info("configuring keepalive enforcement policy",
				"minTime", formatDuration(ep.MinTime),
				"permitWithoutStream", ep.PermitWithoutStream)
		}
		s.serverOptions = append(s.serverOptions, keepaliveOpts...)
	}

	defaultUnary := []grpc.UnaryServerInterceptor{
		middleware.UnaryInjectLogger(s.log),
		middleware.UnaryServerLogging(),
	}
	defaultStream := []grpc.StreamServerInterceptor{
		middleware.StreamInjectLogger(s.log),
		middleware.StreamServerLogging(),
	}

	allUnary := append(defaultUnary, s.unaryInterceptors...)
	allStream := append(defaultStream, s.streamInterceptors...)
	if len(allUnary) > 0 {
		s.serverOptions = append(s.serverOptions, grpc.ChainUnaryInterceptor(allUnary...))
	}
	if len(allStream) > 0 {
		s.serverOptions = append(s.serverOptions, grpc.ChainStreamInterceptor(allStream...))
	}
	for _, h := range s.statsHandlers {
		if h != nil {
			s.serverOptions = append(s.serverOptions, grpc.StatsHandler(h))
		}
	}

	s.grpcServer = grpc.NewServer(s.serverOptions...)
	return s, nil
}

// UnaryInterceptorCount returns the effective unary interceptor count, including defaults.
func (s *Server) UnaryInterceptorCount() int {
	return 2 + len(s.unaryInterceptors)
}

// StreamInterceptorCount returns the effective stream interceptor count, including defaults.
func (s *Server) StreamInterceptorCount() int {
	return 2 + len(s.streamInterceptors)
}

// RegisterServices registers one or more gRPC services on the underlying grpc.Server.
// It can be called multiple times to register different services.
func (s *Server) RegisterServices(registerFunc func(*grpc.Server)) {
	registerFunc(s.grpcServer)
}

// Register applies service registration to the underlying grpc.Server.
func (s *Server) Register(register func(*grpc.Server)) error {
	if register != nil {
		s.RegisterServices(register)
	}
	return nil
}

// Run starts the gRPC server and blocks until ctx is cancelled or server errors.
// Note: Run does NOT call Stop; Stop is called by App uniformly.
func (s *Server) Run(ctx context.Context) error {
	// Register to etcd if configured (do this before starting server)
	if s.config.ShouldRegisterInstance() {
		if err := s.registerInstance(ctx); err != nil {
			s.log.Error("failed to register instance", "error", err)
			return err
		}
	}

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

	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		if err := s.grpcServer.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	// Wait for ctx cancelled or server error
	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return nil
	}
}

// Stop gracefully stops the server and deregisters its discovery instance when present.
// It blocks until the server has finished shutting down.
func (s *Server) Stop(ctx context.Context) error {
	s.log.Info("shutting down rpc server gracefully")

	if s.registrar != nil && s.instance != nil {
		if err := s.registrar.Deregister(ctx, *s.instance); err != nil {
			s.log.Warn("failed to deregister instance", "error", err)
		}
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

// registerInstance registers the service instance through the configured discovery registrar.
func (s *Server) registerInstance(ctx context.Context) error {
	if s.registrar == nil {
		return fmt.Errorf("rpc: discovery registrar is not configured")
	}
	instance := discovery.Instance{
		Name: s.config.Name,
		Host: s.config.AdvertiseAddr,
		Port: s.config.Port,
	}
	if err := s.registrar.Register(ctx, instance); err != nil {
		return err
	}
	s.instance = &instance
	s.log.Info("instance registered", "instance", instance)
	return nil
}
