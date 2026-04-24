# Octopus

Octopus is a lightweight Go service framework built around a small set of clear primitives:

- minimal application runtime lifecycle
- simple application construction facade
- built-in API / gRPC / jobs integration
- shared dependency store
- explicit configuration loading

The framework aims to stay small, direct, and easy to integrate.

The primary extension surface is explicit:
- `assemble.WithSetup(...)` for custom setup work
- `assemble.WithDomains(...)` for business domain registration

---

## Design philosophy

Octopus is organized around a simple boundary:

- `pkg/assemble` creates the application
- `pkg/app` runs the application

That translates into a few practical rules:

- `pkg/app` stays a minimal runtime kernel
- `pkg/assemble` is the main facade application code normally needs
- business code contributes domain registration through small domains
- shared dependencies live in `pkg/store`
- setup and domain registration details should not leak into application entry code

---

## Stability and package boundaries

Octopus is usable, but the public API is still evolving.

Primary packages for application code:

- `pkg/assemble`
- `pkg/app`
- `pkg/store`
- `pkg/api`
- `pkg/rpc`
- `pkg/config`
- `pkg/xlog`
- `pkg/hook`
- `pkg/discovery` for discovery integration

If you are adopting Octopus in production, pin a version and review changes before upgrading.

---

## Installation

```bash
go get github.com/HorseArcher567/octopus
```

---

## Quick start

### Minimal entry

```go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"

    "github.com/HorseArcher567/octopus/pkg/assemble"
)

func main() {
    a, err := assemble.Load(
        "config.yaml",
        assemble.WithDomains(
            user.Register,
            order.Register,
        ),
    )
    if err != nil {
        panic(err)
    }

    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    if err := a.Run(ctx); err != nil {
        panic(err)
    }
}
```

### Run the example service

The multi-service server example is organized by business domain (`user`, `order`, `product`), assembled through business domains, and also shows both a minimal custom setup step via `WithSetup(...)` and an app-level startup hook via `WithStartupHooks(...)` used to initialize the demo SQLite schema.

```bash
cd examples/multi-service/server
go run . -config config.yaml
```

### Run the example client

The multi-service client example runs as a short-lived app whose primary runtime work is a set of registered jobs. Each job is one self-contained demo scenario.

```bash
go run ./examples/multi-service/client \
  -config examples/multi-service/client/config.yaml \
  -target etcd:///multi-service-demo \
  -api http://127.0.0.1:8090
```

---

## Core abstractions

### Application runtime

```go
type Service interface {
    Name() string
    Run(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

A `Service` is any long-running runtime unit managed by `pkg/app`.
Typical examples include:

- API server
- RPC server
- job scheduler
- consumer
- watcher
- worker loop

### Domain registration

```go
type Domain func(*assemble.DomainContext) error
```

Business code contributes to application creation through `Domain`.

A domain may:

- register API routes
- register gRPC handlers
- register jobs
- add startup hooks
- add shutdown hooks
- add custom runtime services
- read dependencies from the shared store through `store.Reader`

### Lifecycle hooks

```go
type Func func(*hook.Context) error
```

Lifecycle hooks are registered from `assemble.DomainContext`:

- `ctx.OnStartup(...)`
- `ctx.OnShutdown(...)`

A hook receives:

- `Context() context.Context`
- `Logger() *xlog.Logger`
- `store.Reader` for resolving shared dependencies

### Jobs

```go
type Func func(*job.Context) error
```

A job receives:

- `Context() context.Context`
- `Logger() *xlog.Logger`
- `Name() string`
- `store.Reader` for resolving shared dependencies

---

## Shared store

Etcd, MySQL, SQLite, Redis, and other shared dependencies are loaded into `pkg/store`.

The store now separates read and write capabilities:

- `store.Reader`
- `store.Writer`
- `store.Store`

Typical operations:

- `Set(...)`
- `SetNamed(...)`
- `Get[T](...)`
- `GetNamed[T](...)`
- `MustGet[T](...)`
- `MustGetNamed[T](...)`
- `Close()`

Read-only application contexts such as domain, hook, and job contexts expose `store.Reader`, not the full store.

---

## Built-in components

### API

The API server supports:

- default middleware stack
- custom middleware via `api.WithMiddleware(...)`
- disabling built-in middleware via `api.WithoutDefaultMiddleware()`
- route registration through domain registration
- `Run(ctx)` / `Stop(ctx)`

### gRPC

The gRPC helpers support:

- server creation
- service registration through domain registration
- custom unary interceptors via `rpc.WithUnaryInterceptors(...)`
- custom stream interceptors via `rpc.WithStreamInterceptors(...)`
- additional server options via `rpc.WithServerOptions(...)`
- explicit outbound dialing via `rpc.NewClient(...)`

### Jobs

Jobs are registered during domain registration and run as an application-managed runtime service.

---

## Recommended application flow

```text
config -> assemble.Load/New -> *app.App -> Run(ctx)
```

Application code should usually interact with only:

- `pkg/assemble`
- `app.Run(...)`

---

## Project structure

```text
octopus/
├── pkg/
│   ├── assemble/      # application construction facade
│   ├── app/           # minimal runtime lifecycle kernel
│   ├── store/         # shared object store
│   ├── hook/          # lifecycle hook context and hook func model
│   ├── job/           # job execution context and job func model
│   ├── rpc/           # gRPC server and client helpers
│   ├── api/           # API server
│   ├── config/        # configuration loading
│   └── xlog/          # logging
├── examples/
│   ├── hello-module/
│   └── multi-service/
└── cmd/
    └── octopus-cli/
```

---

## Testing

```bash
go test ./...
```

---

## License

MIT
