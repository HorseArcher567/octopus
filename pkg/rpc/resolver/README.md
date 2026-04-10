# pkg/rpc/resolver

This package contains gRPC resolver implementations used by the RPC runtime.

Current status:

- `direct:///` remains a normal built-in resolver here
- etcd-related resolver code remains as a transitional compatibility path
- provider-backed discovery resolver integration is being moved under `pkg/discovery/etcd`

Guidance:

- new discovery-oriented application code should prefer `pkg/discovery` and `pkg/discovery/etcd`
- do not depend on etcd-specific compatibility pieces here unless you are maintaining transitional internal wiring
