package resolver

import (
	"log/slog"
	"strings"

	grpcresolver "google.golang.org/grpc/resolver"
)

// DirectResolverBuilder 实现一个简单的直连 resolver.Builder。
// 约定 target 形如：direct:///ip1:port1,ip2:port2，由 resolver 从 Target.Endpoint 中解析出多地址。
// 该 resolver 不做服务发现和动态更新，只在构建时推送一次地址列表。
type DirectResolverBuilder struct {
	log *slog.Logger
}

// NewDirectBuilder 创建一个直连模式的 resolver builder。
// log 为用于输出的 logger。
func NewDirectBuilder(log *slog.Logger) *DirectResolverBuilder {
	if log == nil {
		log = slog.Default()
	}
	return &DirectResolverBuilder{
		log: log,
	}
}

// Scheme 返回直连 resolver 使用的 scheme。
// 仅在通过 grpc.WithResolvers(builder) 传入时使用，无需全局注册。
func (b *DirectResolverBuilder) Scheme() string {
	return "direct"
}

// Build 创建一个新的 resolver 实例，并从 target.Endpoint 中解析 endpoints 推送给 gRPC。
func (b *DirectResolverBuilder) Build(target grpcresolver.Target, cc grpcresolver.ClientConn, opts grpcresolver.BuildOptions) (grpcresolver.Resolver, error) {
	r := &directResolver{
		cc:  cc,
		log: b.log.With("component", "resolver", "scheme", "direct"),
	}

	raw := target.Endpoint() // 形如 "ip1:port1,ip2:port2"
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
		r.log.Warn("direct resolver initialized with empty endpoints",
			"raw_endpoint", raw,
		)
	} else {
		r.log.Info("direct resolver initialized",
			"endpoints", raw,
		)
	}

	cc.UpdateState(grpcresolver.State{Addresses: addrs})
	return r, nil
}

// directResolver 实现 grpcresolver.Resolver 接口。
// 由于直连模式的地址是固定的，这里不做额外的解析或监听。
type directResolver struct {
	cc  grpcresolver.ClientConn
	log *slog.Logger
}

// ResolveNow 对于固定地址的直连 resolver，无需实现额外逻辑。
func (r *directResolver) ResolveNow(opts grpcresolver.ResolveNowOptions) {
	// no-op
}

// Close 关闭 resolver，这里也无需额外清理。
func (r *directResolver) Close() {
	// no-op
}
