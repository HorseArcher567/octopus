package rpc

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// ClientConfig is the configuration for the RPC client.
// Target is the gRPC target URL that specifies how to connect to the service.
// Examples:
//   - "192.168.1.100:50051" for standard gRPC connection
//   - "etcd:///serviceName" for etcd service discovery
//   - "direct:///127.0.0.1:8080,127.0.0.1:8081" for direct connection with multiple endpoints
type ClientConfig struct {
	// Target is the gRPC target URL.
	// Examples:
	//   - "192.168.1.100:50051" for standard gRPC connection
	//   - "etcd:///serviceName" for etcd service discovery
	//   - "direct:///127.0.0.1:8080,127.0.0.1:8081" for direct connection with multiple endpoints
	Target string `yaml:"target" json:"target" toml:"target"`

	// LoadBalancingPolicy is the load balancing policy for the gRPC client.
	// Common values: "round_robin", "pick_first", "grpclb" (default: "round_robin").
	LoadBalancingPolicy string `yaml:"loadBalancingPolicy" json:"loadBalancingPolicy" toml:"loadBalancingPolicy"`

	// EnableKeepalive enables keepalive pings.
	EnableKeepalive bool `yaml:"enableKeepalive" json:"enableKeepalive" toml:"enableKeepalive"`

	// KeepaliveTime is the keepalive time interval (default: 10 seconds).
	KeepaliveTime time.Duration `yaml:"keepaliveTime" json:"keepaliveTime" toml:"keepaliveTime"`

	// KeepaliveTimeout is the keepalive timeout (default: 3 seconds).
	KeepaliveTimeout time.Duration `yaml:"keepaliveTimeout" json:"keepaliveTimeout" toml:"keepaliveTimeout"`

	// PermitWithoutStream allows sending keepalive pings even when there are no active streams.
	PermitWithoutStream bool `yaml:"permitWithoutStream" json:"permitWithoutStream" toml:"permitWithoutStream"`
}

// Validate validates the client configuration.
func (c *ClientConfig) Validate() error {
	if c.Target == "" {
		return errors.New("target is required")
	}

	// Validate target format.
	// Support three formats:
	// 1. Standard gRPC format: "host:port" (e.g., "192.168.1.100:50051")
	// 2. etcd service discovery: "etcd:///serviceName"
	// 3. Direct connection: "direct:///ip1:port1,ip2:port2"
	if strings.HasPrefix(c.Target, "etcd:///") || strings.HasPrefix(c.Target, "direct:///") {
		// Custom scheme format, already validated by prefix
		return nil
	}

	// Standard gRPC format: validate host:port
	if _, _, err := net.SplitHostPort(c.Target); err != nil {
		return fmt.Errorf("invalid target format: %w (expected 'host:port', 'etcd:///serviceName', or 'direct:///ip1:port1,ip2:port2')", err)
	}

	return nil
}

// Normalize sets default values for the client configuration.
// It is called automatically by NewClient before creating the connection.
func (c *ClientConfig) Normalize() {
	if c.LoadBalancingPolicy == "" {
		c.LoadBalancingPolicy = "round_robin"
	}

	if c.KeepaliveTime == 0 {
		c.KeepaliveTime = 10 * time.Second
	}
	if c.KeepaliveTimeout == 0 {
		c.KeepaliveTimeout = 3 * time.Second
	}
}

// BuildDialOptions builds gRPC dial options from the client configuration.
//
// It normalizes the configuration first, then constructs dial options for:
//   - Transport credentials (always insecure)
//   - Load balancing policy
//   - Keepalive parameters (if enabled)
//
// The returned options can be used directly with grpc.NewClient.
func (c *ClientConfig) BuildDialOptions() []grpc.DialOption {
	c.Normalize()

	opts := []grpc.DialOption{}

	// Set transport credentials to insecure by default.
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	// Set load balancing policy.
	serviceConfig := fmt.Sprintf(`{"loadBalancingPolicy":"%s"}`, c.LoadBalancingPolicy)
	opts = append(opts, grpc.WithDefaultServiceConfig(serviceConfig))

	// Configure keepalive if enabled.
	if c.EnableKeepalive {
		kaParams := keepalive.ClientParameters{
			Time:                c.KeepaliveTime,
			Timeout:             c.KeepaliveTimeout,
			PermitWithoutStream: c.PermitWithoutStream,
		}
		opts = append(opts, grpc.WithKeepaliveParams(kaParams))
	}

	return opts
}

// ServerConfig is the configuration for the RPC server.
type ServerConfig struct {
	// Name is the service name used when registering to the service registry.
	Name string `yaml:"name" json:"name" toml:"name"`

	// Host is the listen address (e.g., 0.0.0.0, 127.0.0.1).
	Host string `yaml:"host" json:"host" toml:"host"`

	// Port is the listen port.
	Port int `yaml:"port" json:"port" toml:"port"`

	// EnableReflection enables gRPC reflection.
	// Recommended for development/test environments to enable grpcurl/grpcui debugging.
	EnableReflection bool `yaml:"enableReflection" json:"enableReflection" toml:"enableReflection"`

	// AdvertiseAddr is the address registered to etcd.
	// If empty, the service will not be registered to etcd.
	AdvertiseAddr string `yaml:"advertiseAddr" json:"advertiseAddr" toml:"advertiseAddr"`
}

// Validate validates the server configuration.
func (c *ServerConfig) Validate() error {
	if c.Name == "" {
		return errors.New("server name is required")
	}

	if c.Host == "" {
		return errors.New("server host is required")
	}

	if c.Port <= 0 {
		return errors.New("server port is required")
	}

	return nil
}

// ShouldRegisterInstance reports whether the service instance should be registered.
// If AdvertiseAddr is configured, the service instance will be registered to etcd.
func (c *ServerConfig) ShouldRegisterInstance() bool {
	return c != nil && c.AdvertiseAddr != ""
}
