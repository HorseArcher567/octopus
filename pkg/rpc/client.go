package rpc

import (
	"fmt"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

// ClientOption is a function that configures the RPC client.
type ClientOption func(*clientOptions)

// clientOptions holds the optional configurations for creating a gRPC client.
type clientOptions struct {
	etcdClient *clientv3.Client
}

// WithClientEtcdClient sets the etcd client for service discovery.
// This is required when using etcd:/// target schemes.
func WithClientEtcdClient(client *clientv3.Client) ClientOption {
	return func(o *clientOptions) {
		o.etcdClient = client
	}
}

// NewClient creates a new gRPC client connection with automatic resolver setup.
//
// The config.Target must be specified in one of the following formats:
//   - "192.168.1.100:50051" for standard gRPC connection (no resolver needed)
//   - "etcd:///serviceName" for etcd service discovery (requires WithEtcdClient option)
//   - "direct:///ip1:port1,ip2:port2" for direct connection with multiple endpoints
//
// The client will automatically configure the appropriate resolver based on the target scheme.
// For etcd:/// targets, you must provide the WithEtcdClient option.
// For direct:/// targets, the resolver is automatically configured.
//
// Example usage:
//
//	// For direct connection
//	conn, err := rpc.NewClient(log, &rpc.ClientConfig{Target: "direct:///localhost:8080"})
//
//	// For etcd service discovery
//	conn, err := rpc.NewClient(log, &rpc.ClientConfig{Target: "etcd:///myservice"},
//		rpc.WithClientEtcdClient(etcdClient))
//
// It returns a gRPC client connection and an error if the connection cannot be established.
func NewClient(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	dialOpts := opts

	// Create the connection
	conn, err := grpc.NewClient(target, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection %s: %w", target, err)
	}

	return conn, nil
}

// MustNewClient creates a new gRPC client connection, panicking on error.
//
// It is a convenience wrapper around NewClient that panics if the connection
// cannot be established. Use this only when you are certain the connection will succeed.
func MustNewClient(target string, opts ...grpc.DialOption) *grpc.ClientConn {
	conn, err := NewClient(target, opts...)
	if err != nil {
		panic(err)
	}
	return conn
}
