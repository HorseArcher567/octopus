# pkg/di

`pkg/di` provides Octopus dependency injection primitives.

It supports:

- unnamed bindings via `Provide(...)`
- named bindings via `ProvideNamed(...)`
- single-value resolution via `Resolve(...)` and `ResolveNamed(...)`
- multi-value resolution via `ResolveAll(...)` and `ResolveAllNamed(...)`
- reflective invocation via `Invoke(...)`

The container uses a name + type lookup model, preserves registration order for multi-bindings, and returns typed lookup errors for not-found and ambiguous resolutions.

`pkg/app` uses `pkg/di` as its dependency wiring subsystem, but `pkg/di` can also be used independently in other package-level assembly code.
