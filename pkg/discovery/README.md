# pkg/discovery

`pkg/discovery` defines the service discovery abstraction used by Octopus.

Core abstractions:

- `Instance`
- `Registrar`
- `Resolver`
- `Provider`

Responsibilities:

- service registration
- service deregistration
- service resolution
- service watch/refresh integration
- optional gRPC resolver builder exposure for client-side discovery

This package exists so RPC runtime code can depend on discovery capabilities instead of directly depending on etcd-specific registration and resolution details.

Current status:

- discovery abstraction is in place
- etcd provider exists under `pkg/discovery/etcd`
- RPC server registration already uses discovery registrar
- RPC client discovery prefers provider-backed gRPC resolver builder when available

Some legacy etcd-related code still remains under `pkg/rpc/registry` and `pkg/rpc/resolver` as transitional compatibility paths.
