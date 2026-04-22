# Octopus

Octopus is a lightweight Go service framework built around a small set of clear primitives:

- minimal application runtime lifecycle
- simple application assembly facade
- built-in API / gRPC / jobs integration
- shared dependency store
- explicit configuration loading

The framework aims to stay small, direct, and easy to integrate.

---

## Design philosophy

Octopus is organized around a simple boundary:

- `pkg/assemble` builds the application
- `pkg/app` runs the application

That translates into a few practical rules:

- `pkg/app` stays a minimal runtime kernel
- `pkg/assemble` is the only facade application code normally needs
- business code contributes assembly through small actions
- shared dependencies live in `pkg/store`
- setup and assembly details should not leak into application entry code

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

    "github.com/HorseArcher567/octopus/pkg/assemble"
)

func main() {
    a, err := assemble.Load(
        "config.yaml",
        assemble.With(
            user.Assemble,
            order.Assemble,
        ),
    )
    if err != nil {
        panic(err)
    }

    if err := a.Run(context.Background()); err != nil {
        panic(err)
    }
}
```

### Run the example service

The multi-service server example is organized by business capability (`user`, `order`, `product`), assembled through business actions, and also shows a minimal custom setup step via `WithSetupSteps(...)`.

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
  -api http://127.0.0.1:8090/hello
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

type StartupHook func(ctx context.Context) error
type ShutdownHook func(ctx context.Context) error
```

A `Service` is any long-running runtime unit managed by `pkg/app`.
Typical examples include:

- API server
- RPC server
- job scheduler
- consumer
- watcher
- worker loop

### Application assembly

```go
type Action func(*assemble.Context) error
```

Business code contributes to application construction through `Action`.

An action may:

- register API routes
- register gRPC handlers
- register jobs
- add startup hooks
- add shutdown hooks
- add custom runtime services
- read dependencies from the shared store

---

## Shared store

Etcd, MySQL, Redis, and other shared dependencies are loaded into `pkg/store`.

The store supports typed lookup and named registrations.

Typical operations:

- `Set(...)`
- `SetNamed(...)`
- `Get[T](...)`
- `GetNamed[T](...)`
- `MustGet[T](...)`
- `MustGetNamed[T](...)`
- `Close()`

---

## Built-in components

### API

The API server supports:

- default middleware stack
- custom middleware via `api.WithMiddleware(...)`
- disabling built-in middleware via `api.WithoutDefaultMiddleware()`
- route registration through assembly
- `Run(ctx)` / `Stop(ctx)`

### gRPC

The gRPC helpers support:

- server creation
- service registration through assembly
- custom unary interceptors via `rpc.WithUnaryInterceptors(...)`
- custom stream interceptors via `rpc.WithStreamInterceptors(...)`
- additional server options via `rpc.WithServerOptions(...)`
- explicit outbound dialing via `rpc.NewClient(...)`

### Jobs

Jobs are registered during assembly and run as an application-managed runtime service.

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
