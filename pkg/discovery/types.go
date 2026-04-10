// Package discovery defines Octopus service discovery abstractions used by
// runtimes and infrastructure providers.
package discovery

import "context"

// Instance describes a discoverable service instance.
//
// Service identifies the logical service name, while Address and Port identify
// a reachable endpoint. Metadata may contain provider-specific attributes.
type Instance struct {
	Service  string
	Address  string
	Port     int
	Metadata map[string]string
}

// Registrar registers and deregisters service instances.
//
// Implementations are responsible for publishing instance lifecycle changes to
// the underlying discovery backend.
type Registrar interface {
	Register(ctx context.Context, ins Instance) error
	Deregister(ctx context.Context, ins Instance) error
}

// Resolver resolves and watches service instances.
//
// Resolve returns the current instance set for a service. Watch returns a
// channel that emits updated instance sets until the context is canceled or the
// watch fails.
type Resolver interface {
	Resolve(ctx context.Context, service string) ([]Instance, error)
	Watch(ctx context.Context, service string) (<-chan []Instance, error)
}

// Provider groups registration and resolution capabilities.
//
// Provider may additionally expose transport-specific adapters, such as a gRPC
// resolver builder, for runtimes that integrate directly with discovery.
type Provider interface {
	Registrar() Registrar
	Resolver() Resolver
	GRPCResolverBuilder() any
	Close() error
}
