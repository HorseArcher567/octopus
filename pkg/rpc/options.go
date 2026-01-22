package rpc

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

// Option defines a functional option for configuring the RPC Server.
type Option func(s *Server)

// WithEtcdClient sets the etcd client for the server.
func WithEtcdClient(etcdClient *clientv3.Client) Option {
	return func(s *Server) {
		s.etcdClient = etcdClient
	}
}

// WithGRPCOptions configures the underlying grpc.Server with the provided options.
func WithGRPCOptions(opts ...grpc.ServerOption) Option {
	return func(s *Server) {
		// Collect options; they will be applied when creating grpc.Server in NewServer.
		s.grpcOptions = append(s.grpcOptions, opts...)
	}
}
