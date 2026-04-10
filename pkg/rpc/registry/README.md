# pkg/rpc/registry

This package contains legacy etcd-backed RPC registration code.

It is retained as a transitional compatibility path while registration logic is being moved to the top-level discovery abstraction under `pkg/discovery`.

New discovery-facing development should prefer:

- `pkg/discovery`
- `pkg/discovery/etcd`
