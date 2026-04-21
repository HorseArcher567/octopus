package discovery

import (
	"context"
	"fmt"

	grpcresolver "google.golang.org/grpc/resolver"
)

// Instance describes one reachable gRPC service instance.
type Instance struct {
	ID       string
	Name     string
	Host     string
	Port     int
	Metadata map[string]string
}

// Addr returns host:port.
func (i Instance) Addr() string {
	return fmt.Sprintf("%s:%d", i.Host, i.Port)
}

// Registrar publishes and removes service instances.
type Registrar interface {
	Register(ctx context.Context, instance Instance) error
	Deregister(ctx context.Context, instance Instance) error
}

// Resolver exposes a gRPC resolver builder.
type Resolver interface {
	Builder() grpcresolver.Builder
}

// Discovery combines server-side registration and client-side gRPC target resolution.
type Discovery interface {
	Registrar
	Resolver
}
