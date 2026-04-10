# pkg/rpc/registry

This package contains legacy etcd-backed RPC registration code.

Status:

- transitional compatibility path only
- not recommended for new application code
- may change or be removed once discovery migration is complete

It is retained while registration logic is being moved to the top-level discovery abstraction under `pkg/discovery`.

New discovery-facing development should prefer:

- `pkg/discovery`
- `pkg/discovery/etcd`
