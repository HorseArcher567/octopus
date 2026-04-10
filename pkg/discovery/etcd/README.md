# pkg/discovery/etcd

`pkg/discovery/etcd` is the etcd-backed implementation of the Octopus discovery abstraction.

It currently provides:

- `Provider`
- `Registrar`
- `Resolver`
- gRPC resolver builder adapter

Notes:

- registration is implemented through etcd key/value records
- resolution reads service instances from the `/octopus/rpc/apps/<service>/` prefix
- watch support is implemented in a minimal refresh-oriented form
- a gRPC resolver builder adapter is provided so RPC clients can consume discovery through the provider abstraction

This package is part of the migration away from RPC-internal etcd coupling.
