# pkg/rpc/resolver

This package contains gRPC resolver implementations used by the RPC runtime.

Current status:

- `direct:///` remains a normal built-in resolver here
- etcd-related resolver code remains as a transitional compatibility path
- provider-backed discovery resolver integration is being moved under `pkg/discovery/etcd`

New discovery-oriented work should prefer the discovery abstraction under `pkg/discovery`.
