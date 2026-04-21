package rpc

import (
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// formatDuration formats a duration for human-readable logging.
// Zero duration is formatted as "infinity" (unlimited).
func formatDuration(d time.Duration) string {
	if d == 0 {
		return "infinity"
	}
	return d.String()
}

// ClientKeepalive is the keepalive configuration for gRPC client.
type ClientKeepalive struct {
	// Time is the keepalive time interval.
	// After a duration of this time if the client doesn't see any activity it
	// pings the server to see if the transport is still alive.
	// If set below 10s, a minimum value of 10s will be used instead.
	Time time.Duration `yaml:"time" json:"time" toml:"time"`

	// Timeout is the keepalive timeout.
	// After having pinged for keepalive check, the client waits for a duration
	// of Timeout and if no activity is seen even after that the connection is closed.
	// If keepalive is enabled, and this value is not explicitly set, the default is 20 seconds.
	Timeout time.Duration `yaml:"timeout" json:"timeout" toml:"timeout"`

	// PermitWithoutStream allows sending keepalive pings even when there are no active streams.
	// If false, when there are no active RPCs, Time and Timeout will be ignored and no
	// keepalive pings will be sent.
	PermitWithoutStream bool `yaml:"permitWithoutStream" json:"permitWithoutStream" toml:"permitWithoutStream"`
}

// Normalize normalizes the client keepalive configuration.
// It validates and adjusts values according to gRPC requirements.
func (ck *ClientKeepalive) Normalize() {
	if ck == nil {
		return
	}

	// Note: Zero values are preserved to let gRPC use its defaults.
	// Time: 0 means gRPC will use its default behavior
	// Timeout: 0 means gRPC will use default 20 seconds

	// Adjust Time if set below minimum (10s)
	// gRPC will also do this, but we do it here for consistency
	if ck.Time > 0 && ck.Time < 10*time.Second {
		ck.Time = 10 * time.Second
	}

	// Validate Timeout if set (should be positive)
	// Zero value is allowed (means gRPC default)
	if ck.Timeout < 0 {
		ck.Timeout = 0 // Reset to zero (use gRPC default) if negative
	}
}

type ClientOptions struct {
	// LoadBalancingPolicy is the load balancing policy for the gRPC client.
	// Common values: "round_robin", "pick_first", "grpclb" (default: "round_robin").
	LoadBalancingPolicy string `yaml:"loadBalancingPolicy" json:"loadBalancingPolicy" toml:"loadBalancingPolicy"`

	// Keepalive is the keepalive configuration for the client.
	// If nil, keepalive will not be enabled.
	Keepalive *ClientKeepalive `yaml:"keepalive" json:"keepalive" toml:"keepalive"`
}

// Normalize sets default values for the client configuration.
// It is called automatically by NewClient before creating the connection.
func (c *ClientOptions) Normalize() {
	if c.LoadBalancingPolicy == "" {
		c.LoadBalancingPolicy = "round_robin"
	}

	// Normalize keepalive configuration if configured
	if c.Keepalive != nil {
		c.Keepalive.Normalize()
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
func (c *ClientOptions) BuildDialOptions() []grpc.DialOption {
	c.Normalize()

	opts := []grpc.DialOption{}

	// Set transport credentials to insecure by default.
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	// Set load balancing policy.
	serviceConfig := fmt.Sprintf(`{"loadBalancingPolicy":"%s"}`, c.LoadBalancingPolicy)
	opts = append(opts, grpc.WithDefaultServiceConfig(serviceConfig))

	// Configure keepalive if configured.
	// If Keepalive pointer is not nil, enable keepalive.
	if c.Keepalive != nil {
		kaParams := keepalive.ClientParameters{
			Time:                c.Keepalive.Time,
			Timeout:             c.Keepalive.Timeout,
			PermitWithoutStream: c.Keepalive.PermitWithoutStream,
		}
		opts = append(opts, grpc.WithKeepaliveParams(kaParams))
	}

	return opts
}

// ServerParameters is the server keepalive parameters configuration.
type ServerParameters struct {
	// MaxConnectionIdle is a duration for the amount of time after which an
	// idle connection would be closed by sending a GoAway.
	// Zero value means infinity (no limit).
	MaxConnectionIdle time.Duration `yaml:"maxConnectionIdle" json:"maxConnectionIdle" toml:"maxConnectionIdle"`

	// MaxConnectionAge is a duration for the maximum amount of time a
	// connection may exist before it will be closed by sending a GoAway.
	// A random jitter of +/-10% will be added to MaxConnectionAge to spread out connection storms.
	// Zero value means infinity (no limit).
	MaxConnectionAge time.Duration `yaml:"maxConnectionAge" json:"maxConnectionAge" toml:"maxConnectionAge"`

	// MaxConnectionAgeGrace is an additive period after MaxConnectionAge after
	// which the connection will be forcibly closed.
	// Zero value means infinity (no limit).
	MaxConnectionAgeGrace time.Duration `yaml:"maxConnectionAgeGrace" json:"maxConnectionAgeGrace" toml:"maxConnectionAgeGrace"`

	// Time is the keepalive ping interval.
	// After a duration of this time if the server doesn't see any activity it
	// pings the client to see if the transport is still alive.
	// If set below 1s, a minimum value of 1s will be used instead.
	// Zero value means gRPC default (2 hours).
	Time time.Duration `yaml:"time" json:"time" toml:"time"`

	// Timeout is the keepalive ping timeout.
	// After having pinged for keepalive check, the server waits for a duration
	// of Timeout and if no activity is seen even after that the connection is closed.
	// Zero value means gRPC default (20 seconds).
	Timeout time.Duration `yaml:"timeout" json:"timeout" toml:"timeout"`
}

// Normalize normalizes the server keepalive parameters.
// It validates and adjusts values according to gRPC requirements.
func (sp *ServerParameters) Normalize() {
	if sp == nil {
		return
	}

	// Note: Zero values are preserved to let gRPC use its defaults.
	// Time: 0 means gRPC default (2 hours)
	// Timeout: 0 means gRPC default (20 seconds)
	// MaxConnectionIdle/Age/AgeGrace: 0 means infinity (no limit)

	// Adjust Time if set below minimum (1s)
	// gRPC will also do this, but we do it here for consistency
	if sp.Time > 0 && sp.Time < time.Second {
		sp.Time = time.Second
	}

	// Validate Timeout if set (should be positive)
	// Zero value is allowed (means gRPC default)
	if sp.Timeout < 0 {
		sp.Timeout = 0 // Reset to zero (use gRPC default) if negative
	}
}

// EnforcementPolicy is the keepalive enforcement policy configuration.
type EnforcementPolicy struct {
	// MinTime is the minimum amount of time a client should wait before sending
	// a keepalive ping.
	// Zero value means gRPC default (5 minutes).
	MinTime time.Duration `yaml:"minTime" json:"minTime" toml:"minTime"`

	// PermitWithoutStream allows keepalive pings even when there are no active streams(RPCs).
	// If false, and client sends ping when there are no active streams, server will send
	// GOAWAY and close the connection.
	PermitWithoutStream bool `yaml:"permitWithoutStream" json:"permitWithoutStream" toml:"permitWithoutStream"`
}

// Normalize normalizes the keepalive enforcement policy.
// It sets default values according to gRPC requirements.
func (ep *EnforcementPolicy) Normalize() {
	if ep == nil {
		return
	}

	// Set default MinTime if not configured (gRPC default is 5 minutes)
	if ep.MinTime == 0 {
		ep.MinTime = 5 * time.Minute
	}

	// Validate MinTime if set (should be positive)
	if ep.MinTime < 0 {
		ep.MinTime = 5 * time.Minute // Reset to default if negative
	}
}

// ServerKeepalive is the keepalive configuration for gRPC server.
type ServerKeepalive struct {
	// ServerParameters is the server keepalive parameters.
	// If nil, gRPC defaults will be used.
	ServerParameters *ServerParameters `yaml:"serverParameters" json:"serverParameters" toml:"serverParameters"`

	// EnforcementPolicy is the keepalive enforcement policy.
	// If nil, gRPC defaults will be used.
	EnforcementPolicy *EnforcementPolicy `yaml:"enforcementPolicy" json:"enforcementPolicy" toml:"enforcementPolicy"`
}

// BuildServerOptions builds gRPC server options from the keepalive configuration.
//
// It constructs server options for:
//   - Keepalive server parameters (if configured)
//   - Keepalive enforcement policy (if configured)
//
// The returned options can be used directly with grpc.NewServer.
func (k *ServerKeepalive) BuildServerOptions() []grpc.ServerOption {
	if k == nil {
		return nil
	}

	opts := []grpc.ServerOption{}

	// Configure keepalive server parameters if configured
	if k.ServerParameters != nil {
		k.ServerParameters.Normalize()
		sp := k.ServerParameters
		serverParams := keepalive.ServerParameters{
			MaxConnectionIdle:     sp.MaxConnectionIdle,
			MaxConnectionAge:      sp.MaxConnectionAge,
			MaxConnectionAgeGrace: sp.MaxConnectionAgeGrace,
			Time:                  sp.Time,
			Timeout:               sp.Timeout,
		}
		opts = append(opts, grpc.KeepaliveParams(serverParams))
	}

	// Configure keepalive enforcement policy if configured
	if k.EnforcementPolicy != nil {
		k.EnforcementPolicy.Normalize()
		ep := k.EnforcementPolicy
		enforcementPolicy := keepalive.EnforcementPolicy{
			MinTime:             ep.MinTime,
			PermitWithoutStream: ep.PermitWithoutStream,
		}
		opts = append(opts, grpc.KeepaliveEnforcementPolicy(enforcementPolicy))
	}

	return opts
}

// ServerConfig is the configuration for the RPC server.
type ServerConfig struct {
	// Logger is the name of the logger to use for the RPC server.
	// If empty, the app logger will be used.
	Logger string `yaml:"logger" json:"logger" toml:"logger"`

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

	// Keepalive is the keepalive configuration for the server.
	// If nil, gRPC defaults will be used.
	Keepalive *ServerKeepalive `yaml:"keepalive" json:"keepalive" toml:"keepalive"`
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
