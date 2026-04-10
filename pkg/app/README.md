# pkg/app

`pkg/app` is the orchestration kernel of Octopus.

It is responsible for:

- module graph ordering
- phased module execution (`Build`, `RegisterRPC`, `RegisterAPI`, `RegisterJobs`, `Run`)
- startup and shutdown hooks
- bootstrap-based runtime assembly
- wiring phase-specific capabilities over injected runtimes

Key concepts:

- `Bootstrap`: assembled framework runtimes before `App` construction
- `Assemble(...)`: builds logger / rpc / api / resources / health / telemetry from config
- `NewFromBootstrap(...)`: creates an `App` from assembled runtimes
- `BuildContext`: exposes grouped capabilities instead of concrete infrastructure shortcuts
- `Container`: a `pkg/di` capability that supports unnamed and named bindings, single and multi-resolution, and invoke

`pkg/app` intentionally does not own low-level API/gRPC/resource/DI implementation details; it coordinates them.

## Architecture

`pkg/app` is organized in three layers:

```text
pkg/app
├── Layer 1: Public orchestration surface
│   ├── app.go
│   ├── module.go
│   ├── run.go
│   └── lifecycle.go
│
├── Layer 2: Assembly / bootstrap
│   └── bootstrap.go
│
└── Layer 3: Runtime capability adapters
    └── runtime.go
```

### Layer 1: Public orchestration surface

This layer defines the main application object, module contracts, execution phases, and lifecycle management.

- `app.go`: `App` construction and core state
- `module.go`: module contracts and phase capability interfaces
- `run.go`: phased module execution and runtime startup
- `lifecycle.go`: startup/shutdown hook handling

### Layer 2: Assembly / bootstrap

This layer assembles framework runtimes from configuration before the `App` is created.

- `bootstrap.go`: `Bootstrap`, bootstrap options, `Load`, `FromConfig`, `Assemble`, `NewFromBootstrap`

### Layer 3: Runtime capability adapters

This layer defines the runtime abstraction interfaces consumed by `App` and adapts them into build/register phase capabilities.

- `runtime.go`: runtime interfaces, `BuildContext` adapters, RPC/API/job registrars, and `pkg/di` integration

## Runtime flow

```text
config
  -> bootstrap.go / Load / FromConfig / Assemble
  -> app.go / NewFromBootstrap / New
  -> run.go / phased module execution
  -> lifecycle.go / shutdown
```

## Module flow

```text
run.go
  -> module.Build(...)
  -> runtime.go / BuildContext
  -> container/resources/rpc/telemetry capabilities
```

```text
run.go
  -> module.RegisterRPC(...)  -> runtime.go / rpcRegistrar  -> RPCRuntime
  -> module.RegisterAPI(...) -> runtime.go / apiRegistrar -> APIRuntime
  -> module.RegisterJobs(...) -> runtime.go / jobRegistrar  -> JobRuntime
```
