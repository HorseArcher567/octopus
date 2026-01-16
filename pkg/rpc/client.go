package rpc

import (
	"fmt"
	"strings"

	"github.com/HorseArcher567/octopus/pkg/etcd"
	"github.com/HorseArcher567/octopus/pkg/rpc/resolver"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"google.golang.org/grpc"
	grpcresolver "google.golang.org/grpc/resolver"
)

// RegisterResolver registers the etcd and direct resolvers globally.
// If etcdConfig is valid, it registers the etcd resolver for service discovery.
// The direct resolver is always registered for direct connections.
func RegisterResolver(log *xlog.Logger, etcdConfig *etcd.Config) {
	if etcdConfig.Validate() == nil {
		etcdBuilder := resolver.NewEtcdBuilder(etcdConfig, log)
		grpcresolver.Register(etcdBuilder)
		log.Info("etcd resolver registered", "etcdConfig", etcdConfig)
	}

	directBuilder := resolver.NewDirectBuilder(log)
	grpcresolver.Register(directBuilder)
	log.Info("direct resolver registered")
}

// NewClient creates a new gRPC client connection.
//
// The config.Target must be specified in one of the following formats:
//   - "192.168.1.100:50051" for standard gRPC connection (no resolver registration needed)
//   - "etcd:///serviceName" for etcd service discovery (requires RegisterResolver)
//   - "direct:///ip1:port1,ip2:port2" for direct connection with multiple endpoints (requires RegisterResolver)
//
// The client will use the configured transport credentials and load balancing policy.
// For etcd:/// and direct:/// targets, the resolvers must be registered globally
// by calling RegisterResolver before creating clients.
//
// It returns a gRPC client connection and an error if the connection cannot be established.
func NewClient(log *xlog.Logger, config *ClientConfig) (*grpc.ClientConn, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	opts := config.BuildDialOptions()

	// Create the connection.
	// The resolver is already registered globally in RegisterResolver, so no need to register again.
	conn, err := grpc.NewClient(config.Target, opts...)
	if err != nil {
		log.Error("failed to create connection",
			"error", err,
			"target", config.Target,
		)
		return nil, fmt.Errorf("failed to create connection %s: %w", config.Target, err)
	}

	return conn, nil
}

// NewClientWithEndpoints creates a new gRPC client connection using direct connection mode.
//
// It is a convenience function for development and testing that constructs a direct:///
// target from the provided endpoints and calls NewClient internally.
//
// The endpoints parameter is a list of service addresses in the format "host:port".
func NewClientWithEndpoints(log *xlog.Logger, endpoints []string) (*grpc.ClientConn, error) {
	rawEndpoints := strings.Join(endpoints, ",")
	target := fmt.Sprintf("direct:///%s", rawEndpoints)
	config := &ClientConfig{
		Target: target,
	}
	return NewClient(log, config)
}

// MustNewClient creates a new gRPC client connection, panicking on error.
//
// It is a convenience wrapper around NewClient that panics if the connection
// cannot be established. Use this only when you are certain the connection will succeed.
func MustNewClient(log *xlog.Logger, config *ClientConfig) *grpc.ClientConn {
	conn, err := NewClient(log, config)
	if err != nil {
		panic(err)
	}
	return conn
}
