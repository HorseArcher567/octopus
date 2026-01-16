package rpc

import (
	"github.com/HorseArcher567/octopus/pkg/etcd"
	"google.golang.org/grpc"
)

// Option defines a functional option for configuring the RPC Server.
type Option func(s *Server)

// WithEtcdConfig sets the etcd configuration for the server.
func WithEtcdConfig(cfg *etcd.Config) Option {
	return func(s *Server) {
		s.etcdConfig = cfg
	}
}

// WithGRPCOptions configures the underlying grpc.Server with the provided options.
func WithGRPCOptions(opts ...grpc.ServerOption) Option {
	return func(s *Server) {
		// Collect options; they will be applied when creating grpc.Server in NewServer.
		s.grpcOptions = append(s.grpcOptions, opts...)
	}
}
