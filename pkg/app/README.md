# pkg/app

`pkg/app` is the minimal runtime kernel of Octopus.

It is intentionally small.

Its responsibility is only to:

- hold the application logger
- hold the shared dependency store owned by the app
- hold runtime services
- hold startup and shutdown hooks
- run startup, service execution, and shutdown cleanly

`pkg/app` does **not** own:

- config loading
- resource setup logic
- application assembly
- API/RPC/Jobs registration
- business integration

Those concerns belong outside `pkg/app`.

---

## Core concepts

### `App`

`App` is the lifecycle orchestrator.

It runs:

1. startup hooks
2. services
3. shutdown sequence
4. store close

### `Service`

A `Service` is any long-running runtime unit managed by `App`.

Typical examples:

- API server
- RPC server
- job scheduler
- consumer
- watcher
- worker loop

```go
type Service interface {
    Name() string
    Run(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

### Hooks

Hooks are one-shot lifecycle callbacks.

```go
type StartupHook func(ctx context.Context) error
type ShutdownHook func(ctx context.Context) error
```

Startup hooks run before services start.
Shutdown hooks run after services have been stopped and before the store is closed.

### Store ownership

`App` owns the shared `store.Store` when one is provided.

The store is closed during shutdown, after services stop and shutdown hooks run.
Shutdown hooks therefore should be treated as post-service cleanup hooks.
This allows shared resources such as loggers, databases, redis clients, and etcd clients to be released in one place.

---

## Runtime flow

A typical runtime flow is:

```text
startup hooks
  -> run services
  -> wait for context cancellation or service error
  -> stop services
  -> shutdown hooks
  -> close store
```

Recommended ordering semantics:

- startup hooks: forward order
- services run: concurrent
- services stop: reverse order
- shutdown hooks: reverse order, after services have stopped
- store close: final step

---

## Public API

```go
type Config struct {
    Logger          string
    ShutdownTimeout time.Duration
}

func New(log *xlog.Logger, opts ...Option) *App // if log is nil, a default logger is created
func WithStore(s store.Store) Option
func WithShutdownTimeout(timeout time.Duration) Option
func (a *App) AddServices(services ...Service) *App
func (a *App) OnStartup(h StartupHook) *App
func (a *App) OnShutdown(h ShutdownHook) *App
func (a *App) Run(ctx context.Context) error
```

This package is meant to stay boring, small, and dependable.

---

## Example

```go
a := app.New(log, app.WithStore(st))
a.AddServices(apiService, rpcService, jobsService)
a.OnStartup(func(ctx context.Context) error {
    return nil
})
a.OnShutdown(func(ctx context.Context) error {
    return nil
})

if err := a.Run(context.Background()); err != nil {
    panic(err)
}
```

In real applications, `App` is usually constructed by `pkg/assemble`, not directly by business code.

---

## Relationship with `pkg/assemble`

`pkg/app` is the runtime kernel.

`pkg/assemble` is the application construction facade.

The intended relationship is:

```text
assemble builds the app
app runs the app
```

That boundary should remain stable.
