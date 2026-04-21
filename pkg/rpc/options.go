package rpc

import (
	"github.com/HorseArcher567/octopus/pkg/discovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
)

// Option defines a functional option for configuring a Server.
type Option func(s *Server)

// WithRegistrar sets the discovery registrar used for instance registration.
func WithRegistrar(registrar discovery.Registrar) Option {
	return func(s *Server) {
		s.registrar = registrar
	}
}

// WithUnaryInterceptors appends unary server interceptors after the built-in defaults.
func WithUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) Option {
	return func(s *Server) {
		s.unaryInterceptors = append(s.unaryInterceptors, interceptors...)
	}
}

// WithStreamInterceptors appends stream server interceptors after the built-in defaults.
func WithStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) Option {
	return func(s *Server) {
		s.streamInterceptors = append(s.streamInterceptors, interceptors...)
	}
}

// WithServerOptions configures the underlying grpc.Server with non-interceptor server options.
func WithServerOptions(opts ...grpc.ServerOption) Option {
	return func(s *Server) {
		s.serverOptions = append(s.serverOptions, opts...)
	}
}

// WithStatsHandlers appends gRPC stats handlers.
func WithStatsHandlers(handlers ...stats.Handler) Option {
	return func(s *Server) {
		s.statsHandlers = append(s.statsHandlers, handlers...)
	}
}
