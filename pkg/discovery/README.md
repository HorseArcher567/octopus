# pkg/discovery

`pkg/discovery` provides the unified gRPC discovery model used by Octopus.

Core abstractions:

- `Instance`
- `Registrar`
- `Resolver`
- `Discovery`

Built-in implementations:

- `EtcdRegistrar`
- `EtcdResolver`
- `EtcdDiscovery`
- `DirectResolver`

Responsibilities:

- service registration for gRPC servers
- gRPC target resolution for clients
- support for `etcd:///service-name`
- support for `direct:///host1:port1,host2:port2`

Usage:

### Direct target dialing

```go
resolver := discovery.NewDirectResolver(log)
conn, err := rpc.NewClient(
    "direct:///127.0.0.1:9001,127.0.0.1:9002",
    grpc.WithResolvers(resolver.Builder()),
)
if err != nil {
    return err
}
defer conn.Close()
```

### Etcd-based discovery

```go
resolver := discovery.NewEtcdResolver(log, etcdClient)
conn, err := rpc.NewClient(
    "etcd:///user-service",
    grpc.WithResolvers(resolver.Builder()),
)
if err != nil {
    return err
}
defer conn.Close()
```
