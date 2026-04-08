# Runtime Refactor Plan

## Goal

Refactor `pkg/app` from the current large `Runtime` model to a phased module model with simpler boundaries, clearer lifecycle stages, and more explicit dependency wiring.

## Design Philosophy

**框架负责编排，模块负责业务，让边界清晰、依赖显式、生命周期自然。**

When evolving this framework, prefer these rules:

- let `App` orchestrate order, startup, shutdown, and shared host components
- let modules express business construction, registration, and runtime behavior only
- expose only the capability needed in each phase instead of passing a broad runtime object
- make dependency flow explicit through module contracts and container resolution, not peer module references

## Why Change

The current design has three main problems:

- `Runtime` mixes unrelated concerns: resources, outbound clients, inbound registration, and jobs.
- `Module.Init(ctx, rt)` is forced to do too much in one method.
- Modules can see more framework capability than they actually need.

The refactor should make the framework easier to read, test, and extend without turning it into a heavy DI system.

## Design Principles

- Keep names short and literal.
- Prefer one obvious path over multiple optional patterns.
- Separate build-time wiring from runtime registration.
- Let modules depend on capabilities, not on peer module instances.
- Keep the container minimal and type-based.

## Final Naming

This plan uses one stable naming set throughout:

- `Module`
- `DependentModule`
- `BuildModule`
- `CloseModule`
- `RegisterRPCModule`
- `RegisterHTTPModule`
- `RegisterJobsModule`
- `RunModule`
- `Resolver`
- `Container`
- `BuildContext`
- `RPCRegistrar`
- `HTTPRegistrar`
- `JobRegistrar`

These names are intentionally plain:

- `BuildModule` means "participates in build phase"
- `CloseModule` means "owns explicit cleanup"
- `RegisterRPCModule` means "participates in RPC registration phase"
- `RunModule` means "owns a long-running loop"
- `Container` is writable
- `Resolver` is read-only

## Target Model

### Module Interfaces

```go
package app

import (
	"context"

	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/database"
	"github.com/HorseArcher567/octopus/pkg/job"
	redisclient "github.com/HorseArcher567/octopus/pkg/redis"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"google.golang.org/grpc"
)

type Module interface {
	ID() string
}

type DependentModule interface {
	DependsOn() []string
}

type BuildModule interface {
	Build(context.Context, BuildContext) error
}

type CloseModule interface {
	Close(context.Context) error
}

type RegisterRPCModule interface {
	RegisterRPC(context.Context, RPCRegistrar) error
}

type RegisterHTTPModule interface {
	RegisterHTTP(context.Context, HTTPRegistrar) error
}

type RegisterJobsModule interface {
	RegisterJobs(context.Context, JobRegistrar) error
}

type RunModule interface {
	Run(context.Context) error
}

type Resolver interface {
	Resolve(target any) error
	MustResolve(target any)
}

type Container interface {
	Resolver
	Provide(value any) error
}

type BuildContext interface {
	Logger() *xlog.Logger
	MySQL(name string) (*database.DB, error)
	Redis(name string) (*redisclient.Client, error)
	RPCClient(target string) (*grpc.ClientConn, error)
	Container
}

type RPCRegistrar interface {
	Logger() *xlog.Logger
	Resolver
	RegisterRPC(func(*grpc.Server)) error
}

type HTTPRegistrar interface {
	Logger() *xlog.Logger
	Resolver
	RegisterHTTP(func(*api.Engine)) error
}

type JobRegistrar interface {
	Logger() *xlog.Logger
	Resolver
	AddJob(name string, fn job.Func) error
}
```

### Why This Shape

- The build phase is the only phase allowed to create and publish dependencies.
- Registration phases are read-only and only expose the host they need.
- `RunModule` matches the existing `Run(ctx)` style already used by servers and schedulers.
- `CloseModule` gives build-only modules an explicit cleanup path.
- Registration methods return `error` instead of panicking on missing host config.

## Lifecycle

The new lifecycle is fixed and explicit:

1. Resolve module dependency order.
2. Run all `BuildModule`s.
3. Run all registration phases:
   - `RegisterRPC`
   - `RegisterHTTP`
   - `RegisterJobs`
4. Run startup hooks.
5. Start built-in runners:
   - RPC server
   - HTTP server
   - job scheduler
6. Start all `RunModule`s.
7. Wait until context cancellation or error.
8. Shut down in reverse order.

The shutdown model should be orchestrated by `App`, not inferred from runner behavior. `App` should explicitly trigger stop, then wait for runners to exit, then close remaining resources.

### Pseudocode

```go
func (a *App) Run(ctx context.Context) (retErr error) {
	if err := a.markRunOnce(); err != nil {
		return err
	}
	if ctx == nil {
		ctx = context.Background()
	}

	if err := a.prepareModules(); err != nil {
		return err
	}
	if err := a.buildModules(ctx); err != nil {
		return err
	}
	if err := a.registerRPCModules(ctx); err != nil {
		return err
	}
	if err := a.registerHTTPModules(ctx); err != nil {
		return err
	}
	if err := a.registerJobsModules(ctx); err != nil {
		return err
	}
	if err := a.execStartupHooks(ctx); err != nil {
		return err
	}

	runCtx, cancel := context.WithCancel(ctx)
	g, groupCtx := errgroup.WithContext(runCtx)
	a.runBuiltins(g, groupCtx)
	a.runModules(g, groupCtx)

	retErr = g.Wait()
	cancel()
	retErr = errors.Join(retErr, a.shutdown())
	if errors.Is(retErr, context.Canceled) || errors.Is(retErr, context.DeadlineExceeded) {
		return nil
	}
	return retErr
}
```

### Shutdown Order

Recommended shutdown order:

1. stop HTTP server
2. stop RPC server
3. stop job scheduler
4. wait for `RunModule`s to exit
5. run `CloseModule`s in reverse order
6. run shutdown hooks
7. close resources
8. close RPC clients
9. close logger

This keeps inbound traffic from entering while internals are already closing.

## Container

The container should stay deliberately small.

### API

```go
type Resolver interface {
	Resolve(target any) error
	MustResolve(target any)
}

type Container interface {
	Resolver
	Provide(value any) error
}
```

### Rules

- registration is by the static type passed to `Provide`
- only one value per type in v1
- duplicate `Provide` returns error
- `Resolve` requires a pointer target
- `Resolve(&iface)` resolves by interface type
- `Resolve(&ptr)` resolves by pointer type
- no named keys in v1
- no auto construction
- no recursive injection
- no field tags

This keeps the API generic without becoming clever.

### Usage Rule

If a caller wants to register a value as an interface, it must first assign it to a variable of that interface type.

```go
var userRepo repository.UserRepository = repository.NewUserRepository(db)
if err := b.Provide(userRepo); err != nil {
	return err
}
```

This rule avoids reflection guessing and keeps registration predictable.

## App Responsibilities

After refactor, `App` should remain the owner of framework-level components:

- logger
- resource manager
- RPC server
- HTTP server
- job scheduler
- RPC client cache
- container

`App` should not be a general service locator. It should expose capabilities only through phase-specific contexts and registrars.

## Error Handling

- build failure: stop immediately
- registration failure: stop immediately
- startup hook failure: stop immediately
- builtin runner error: trigger shutdown
- `RunModule` error: trigger shutdown
- `CloseModule` error: aggregate during shutdown
- shutdown errors: aggregate with `errors.Join`
- missing server during registration: return error, never panic

This makes failure behavior predictable and testable.

## `examples/multi-service` Migration

### Target Structure

```text
examples/multi-service/server/internal/bootstrap/
  infra_module.go
  service_module.go
  rpc_module.go
  api_module.go
```

### `infra_module.go`

Responsibilities:

- get MySQL from `BuildContext`
- create repositories
- provide repositories into the container

```go
type InfraModule struct{}

func (m *InfraModule) ID() string { return "infra" }

func (m *InfraModule) Build(ctx context.Context, b app.BuildContext) error {
	db, err := b.MySQL("primary")
	if err != nil {
		return err
	}

	var userRepo repository.UserRepository = repository.NewUserRepository(db)
	var orderRepo repository.OrderRepository = repository.NewOrderRepository(db)
	var productRepo repository.ProductRepository = repository.NewProductRepository(db)

	if err := b.Provide(userRepo); err != nil {
		return err
	}
	if err := b.Provide(orderRepo); err != nil {
		return err
	}
	return b.Provide(productRepo)
}
```

### `service_module.go`

Responsibilities:

- resolve repositories
- create services
- provide services

```go
type ServiceModule struct{}

func (m *ServiceModule) ID() string { return "service" }

func (m *ServiceModule) DependsOn() []string { return []string{"infra"} }

func (m *ServiceModule) Build(ctx context.Context, b app.BuildContext) error {
	var userRepo repository.UserRepository
	var orderRepo repository.OrderRepository
	var productRepo repository.ProductRepository

	b.MustResolve(&userRepo)
	b.MustResolve(&orderRepo)
	b.MustResolve(&productRepo)

	if err := b.Provide(service.NewUserService(userRepo)); err != nil {
		return err
	}
	if err := b.Provide(service.NewOrderService(orderRepo)); err != nil {
		return err
	}
	return b.Provide(service.NewProductService(productRepo))
}
```

### Optional `CloseModule`

If a module builds resources that are not already owned by `App`, it should also implement `CloseModule`.

```go
type CacheModule struct {
	client *cache.Client
}

func (m *CacheModule) ID() string { return "cache" }

func (m *CacheModule) Build(ctx context.Context, b app.BuildContext) error {
	client, err := cache.NewClient(...)
	if err != nil {
		return err
	}
	m.client = client
	return nil
}

func (m *CacheModule) Close(ctx context.Context) error {
	if m.client == nil {
		return nil
	}
	return m.client.Close()
}
```

### `rpc_module.go`

Responsibilities:

- resolve services
- create handlers
- register gRPC services

```go
type RPCModule struct{}

func (m *RPCModule) ID() string { return "rpc" }

func (m *RPCModule) DependsOn() []string { return []string{"service"} }

func (m *RPCModule) RegisterRPC(ctx context.Context, r app.RPCRegistrar) error {
	var userSvc *service.UserService
	var orderSvc *service.OrderService
	var productSvc *service.ProductService

	r.MustResolve(&userSvc)
	r.MustResolve(&orderSvc)
	r.MustResolve(&productSvc)

	log := r.Logger()
	return r.RegisterRPC(func(s *grpc.Server) {
		pb.RegisterUserServer(s, grpcx.NewUserHandler(userSvc, log))
		pb.RegisterOrderServer(s, grpcx.NewOrderHandler(orderSvc, log))
		pb.RegisterProductServer(s, grpcx.NewProductHandler(productSvc, log))
	})
}
```

### `api_module.go`

Responsibilities:

- register HTTP routes only

```go
type APIModule struct{}

func (m *APIModule) ID() string { return "api" }

func (m *APIModule) RegisterHTTP(ctx context.Context, r app.HTTPRegistrar) error {
	return r.RegisterHTTP(func(engine *api.Engine) {
		httpx.RegisterRoutes(engine)
	})
}
```

### New Wiring

`main.go` should become conceptually like this:

```go
app.MustRun(*configFile, []app.Module{
	bootstrap.NewInfraModule(),
	bootstrap.NewServiceModule(),
	bootstrap.NewRPCModule(),
	bootstrap.NewAPIModule(),
})
```

The key improvement is that transport modules no longer hold references to infra modules.

## Implementation Order

1. replace `Runtime`-based interfaces in `pkg/app/module.go`
2. add a small type-based `container` implementation
3. add `BuildContext`, `RPCRegistrar`, `HTTPRegistrar`, `JobRegistrar`
4. rewrite `pkg/app/lifecycle.go` around phased execution
5. adapt `pkg/app/app.go` to own the new runtime pieces
6. migrate `examples/multi-service`
7. add focused tests

## Test Plan

Add targeted tests for:

- module dependency ordering
- build phase failure
- RPC registration failure
- HTTP registration failure
- jobs registration failure
- `RunModule` error propagation
- `CloseModule` ordering and error aggregation
- shutdown ordering
- container `Provide` and `Resolve`
- `examples/multi-service` wiring

## Scope Guardrails

Keep v1 small:

- no named bindings
- no auto wiring
- no reflection-heavy magic
- no module-local mini containers
- no multiple competing lifecycle styles

The goal is not to build a DI framework. The goal is to make module boundaries obvious and boring.
