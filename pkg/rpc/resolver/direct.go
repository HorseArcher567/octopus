package resolver

import (
	"strings"

	"github.com/HorseArcher567/octopus/pkg/xlog"
	grpcresolver "google.golang.org/grpc/resolver"
)

// DirectResolverBuilder implements a simple direct connection resolver.Builder.
// It parses addresses from the target endpoint in the format "ip1:port1,ip2:port2"
// and provides them to gRPC without service discovery or dynamic updates.
// The resolver only pushes the address list once during initialization.
type DirectResolverBuilder struct {
	log *xlog.Logger
}

// NewDirectBuilder creates a new DirectResolverBuilder.
// The log parameter is used for logging output.
func NewDirectBuilder(log *xlog.Logger) *DirectResolverBuilder {
	return &DirectResolverBuilder{
		log: log,
	}
}

// Scheme returns the resolver scheme used in gRPC target URLs.
// It should be used with grpc.WithResolvers(builder) and does not require global registration.
func (b *DirectResolverBuilder) Scheme() string {
	return SchemeDirect
}

// Build creates a new resolver instance for the given target.
// It parses comma-separated addresses from target.Endpoint and updates
// the gRPC client connection with the resolved addresses.
func (b *DirectResolverBuilder) Build(target grpcresolver.Target, cc grpcresolver.ClientConn, opts grpcresolver.BuildOptions) (grpcresolver.Resolver, error) {
	r := &directResolver{
		cc: cc,
	}

	// Parse comma-separated addresses from endpoint (e.g., "ip1:port1,ip2:port2")
	raw := target.Endpoint()
	parts := strings.Split(raw, ",")

	addrs := make([]grpcresolver.Address, 0, len(parts))
	for _, ep := range parts {
		ep = strings.TrimSpace(ep)
		if ep == "" {
			continue
		}
		addrs = append(addrs, grpcresolver.Address{Addr: ep})
	}

	if len(addrs) == 0 {
		b.log.Warn("direct resolver initialized with empty endpoints",
			"raw_endpoint", raw,
		)
	} else {
		b.log.Info("direct resolver initialized",
			"endpoints", raw,
		)
	}

	cc.UpdateState(grpcresolver.State{Addresses: addrs})
	return r, nil
}

// directResolver implements the gRPC resolver.Resolver interface
// for direct connection without service discovery.
// Since addresses are fixed in direct mode, no additional resolution
// or monitoring is performed.
type directResolver struct {
	cc grpcresolver.ClientConn
}

// ResolveNow triggers an immediate resolution of the target.
// For direct resolver with fixed addresses, no additional logic is needed.
func (r *directResolver) ResolveNow(opts grpcresolver.ResolveNowOptions) {
	// no-op
}

// Close closes the resolver and releases associated resources.
// For direct resolver, no additional cleanup is needed.
func (r *directResolver) Close() {
	// no-op
}
