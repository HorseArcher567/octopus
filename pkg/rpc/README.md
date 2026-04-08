# pkg/rpc

`pkg/rpc` provides the RPC runtime used by Octopus.

Key pieces:

- `Server`: inbound gRPC server lifecycle
- `ClientFactory`: cached outbound clients
- `Runtime`: combines server, client factory, and optional etcd discovery
- `resolver/`: discovery backends such as `etcd:///` and `direct:///`
