# pkg/rpc

`pkg/rpc` provides the gRPC runtime used by Octopus.

Key pieces:

- `Server`: inbound gRPC server lifecycle
- `ClientFactory`: cached outbound clients
- `Runtime`: combines server, client factory, and discovery integration
- `resolver/`: built-in resolver implementations such as `direct:///`

Server extension points:

- built-in unary interceptors for logger injection and request logging
- built-in stream interceptors for logger injection and stream logging
- `WithUnaryInterceptors(...)`
- `WithStreamInterceptors(...)`
- `WithServerOptions(...)`
- `WithStatsHandlers(...)`
- `WithRegistrar(...)`

Discovery status:

- RPC server registration now uses the top-level discovery registrar abstraction
- RPC client discovery prefers provider-backed gRPC resolver builders when available
- `pkg/discovery` is now the primary abstraction layer
- some legacy etcd-specific code still remains under `pkg/rpc/registry` and `pkg/rpc/resolver` as transitional compatibility paths
