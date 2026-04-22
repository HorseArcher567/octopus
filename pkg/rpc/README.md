# pkg/rpc

`pkg/rpc` provides the gRPC server and thin client helpers used by Octopus.

Key pieces:

- `Server`: inbound gRPC server lifecycle
- `NewClient(...)`: thin outbound dial helper

Server extension points:

- built-in unary interceptors for logger injection and request logging
- built-in stream interceptors for logger injection and stream logging
- `WithUnaryInterceptors(...)`
- `WithStreamInterceptors(...)`
- `WithServerOptions(...)`
- `WithStatsHandlers(...)`
- `WithRegistrar(...)`
- `ServerConfig.Advertise` for config-driven registration intent

Discovery usage:

- RPC server registration uses `pkg/discovery.Registrar`
- RPC client dialing is explicit
- resolver builders may be registered globally by scheme before dialing

Example:

```go
conn, err := rpc.NewClient(
    "direct:///127.0.0.1:9001,127.0.0.1:9002",
    grpc.WithTransportCredentials(insecure.NewCredentials()),
)
```
