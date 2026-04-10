# Octopus

Octopus is a lightweight Go service runtime focused on a few strong primitives:

- unified application lifecycle orchestration
- modular phased startup
- built-in API / gRPC runtime integration
- shared resource management
- baseline health and telemetry (metrics / tracing)

## Design philosophy

**The framework orchestrates. Modules own business logic. Dependencies stay explicit.**

That translates into four practical rules:

- `App` owns lifecycle ordering, shared runtimes, and shutdown behavior
- modules express business construction, registration, and run loops
- each phase only exposes the capabilities it should expose
- dependencies flow through explicit assembly and DI, not hidden module coupling

## Stability and package boundaries

Octopus is currently usable, but the public API surface is still being refined.

Stability expectations:

- `pkg/app`, `pkg/di`, `pkg/api`, `pkg/rpc`, `pkg/resource`, `pkg/config`, `pkg/health`, `pkg/telemetry`, and `pkg/xlog` are the primary packages intended for framework consumers.
- `pkg/discovery` and `pkg/discovery/etcd` are the preferred discovery-facing abstractions.
- `pkg/rpc/registry` and etcd-specific compatibility code under `pkg/rpc/resolver` are transitional compatibility paths and should not be used for new application code.
- `cmd/octopus-cli` is a scaffolding tool; generated templates may evolve along with the framework.

If you are adopting Octopus in production, pin a version and review release notes or commit history before upgrading.

Production adoption note:

- prefer the public packages listed below
- avoid direct dependencies on transitional compatibility paths
- validate startup, shutdown, discovery, and telemetry behavior in your own integration environment before rolling out broadly

## Installation

```bash
go get github.com/HorseArcher567/octopus
```

## Quick start

### 1. Run the example service

```bash
cd examples/multi-service/server
cp .env.example .env
export $(grep -v '^#' .env | xargs)
go run ./cmd/server -config configs/config.dev.yaml
```

### 2. Run the example client

```bash
go run ./examples/multi-service/client \
  -config examples/multi-service/client/config.yaml \
  -target etcd:///multi-service-demo \
  -api http://127.0.0.1:8090/hello
```

### 3. Run the minimal example

```bash
cd examples/hello-module
go run . -config config.yaml
```

## Core abstractions

```go
type Module interface {
    ID() string
}

type DependentModule interface {
    DependsOn() []string
}

type BuildModule interface {
    Build(ctx context.Context, b BuildContext) error
}

type RegisterRPCModule interface {
    RegisterRPC(ctx context.Context, r RPCRegistrar) error
}

type RegisterAPIModule interface {
    RegisterAPI(ctx context.Context, r APIRegistrar) error
}

type RegisterJobsModule interface {
    RegisterJobs(ctx context.Context, r JobRegistrar) error
}

type CloseModule interface {
    Close(ctx context.Context) error
}
```

## Build phase capabilities

`BuildContext` now exposes grouped capabilities instead of infrastructure-specific shortcuts:

- `Logger()`
- `Container()`
- `Resources()`
- `RPC()`
- `Telemetry()`

Example:

```go
db, err := resource.As[*database.DB](b.Resources(), resource.KindMySQL, "primary")
if err != nil {
    return err
}

if err := b.Container().Provide(NewRepo(db)); err != nil {
    return err
}
```

## Runtime capabilities

### API

The API runtime supports:

- default middleware stack
- custom middleware extension via `api.WithMiddleware(...)`
- disabling built-in middleware via `api.WithoutDefaultMiddleware()`
- `Register(...)`
- `Run(ctx)` / `Stop(ctx)`

### gRPC

The gRPC runtime supports:

- default logging and logger-injection interceptors
- custom unary interceptors via `rpc.WithUnaryInterceptors(...)`
- custom stream interceptors via `rpc.WithStreamInterceptors(...)`
- additional server options via `rpc.WithServerOptions(...)`
- cached outbound clients

### Resources

Resources are now accessed through a generic runtime model:

- `Register(kind, name, value, closeFn)`
- `Get(kind, name)`
- `MustGet(kind, name)`
- `List(kind)`
- `Close()`

Built-in kinds currently include:

- `resource.KindMySQL`
- `resource.KindRedis`

### Dependency injection

The built-in dependency injection container in `pkg/di` supports:

- `Provide(...)`
- `ProvideNamed(...)`
- `Resolve(...)`
- `ResolveNamed(...)`
- `ResolveAll(...)`
- `ResolveAllNamed(...)`
- `Invoke(...)`

## Health and telemetry

Octopus now includes built-in health and telemetry support.

### Health

- API `/health`
- aggregate health status
- checker registry

### Telemetry metrics

- API `/metrics`
- API request counters and duration histograms
- gRPC request counters and duration histograms

### Telemetry tracing

- Gin tracing middleware
- gRPC tracing via stats handler
- OpenTelemetry tracer provider runtime

> Current tracing bootstrap uses a minimal default setup suitable for development. It should be further refined into explicit config in a later pass.

## RPC client reuse

- `NewRPCClient(target)` reuses connections by target
- repeated calls with the same target return the same `*grpc.ClientConn`
- `CloseRpcClients()` releases cached clients

## Application entry example

```go
app.MustRun(configFile, []app.Module{
    bootstrap.NewInfraModule(),
    bootstrap.NewServiceModule(),
    bootstrap.NewRPCModule(),
    bootstrap.NewAPIModule(),
})
```

## Project structure

```text
octopus/
├── pkg/
│   ├── app/            # lifecycle orchestration and assembly
│   ├── di/             # dependency injection container
│   ├── rpc/            # gRPC runtime and client factory
│   ├── api/            # API runtime
│   ├── resource/       # generic shared resources
│   ├── health/         # health runtime and checks
│   ├── telemetry/      # metrics / tracing runtime
│   ├── config/         # configuration loading
│   └── xlog/           # logging
├── examples/
│   ├── hello-module/
│   └── multi-service/
└── cmd/
    └── octopus-cli/
```

## Package guidance

Use these packages directly in application code:

- `pkg/app`
- `pkg/di`
- `pkg/api`
- `pkg/rpc`
- `pkg/resource`
- `pkg/config`
- `pkg/health`
- `pkg/telemetry`
- `pkg/xlog`
- `pkg/discovery` / `pkg/discovery/etcd` for discovery integration

Avoid depending directly on these transitional internal paths in new code:

- `pkg/rpc/registry`
- etcd-specific compatibility pieces under `pkg/rpc/resolver`

## Testing

```bash
go test ./...
```

## License

MIT
