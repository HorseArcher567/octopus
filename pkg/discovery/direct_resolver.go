package discovery

import (
	"strings"

	"github.com/HorseArcher567/octopus/pkg/xlog"
	grpcresolver "google.golang.org/grpc/resolver"
)

// DirectResolver exposes a gRPC resolver builder for direct:/// targets.
type DirectResolver struct {
	log *xlog.Logger
}

// NewDirectResolver creates a direct target resolver.
func NewDirectResolver(log *xlog.Logger) *DirectResolver {
	return &DirectResolver{log: log}
}

// Builder returns a gRPC resolver builder for direct:/// targets.
func (r *DirectResolver) Builder() grpcresolver.Builder {
	return &directGRPCResolverBuilder{log: r.log}
}

type directGRPCResolverBuilder struct {
	log *xlog.Logger
}

func (b *directGRPCResolverBuilder) Scheme() string { return "direct" }

func (b *directGRPCResolverBuilder) Build(target grpcresolver.Target, cc grpcresolver.ClientConn, _ grpcresolver.BuildOptions) (grpcresolver.Resolver, error) {
	r := &directGRPCResolver{}
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
		b.log.Warn("direct resolver initialized with empty endpoints", "raw_endpoint", raw)
	} else {
		b.log.Info("direct resolver initialized", "endpoints", raw)
	}
	cc.UpdateState(grpcresolver.State{Addresses: addrs})
	return r, nil
}

type directGRPCResolver struct{}

func (r *directGRPCResolver) ResolveNow(grpcresolver.ResolveNowOptions) {}
func (r *directGRPCResolver) Close()                                    {}
