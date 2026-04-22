# pkg/assemble

`pkg/assemble` is the application construction facade of Octopus.

It is the primary package application developers should use to build an `*app.App`.

Its job is to hide internal construction details and expose one compact integration surface.
The public entrypoints live in `assemble.go`, while internal setup and final app construction stay in `setup.go` and `build.go`.

---

## Responsibilities

`pkg/assemble` is responsible for:

- loading or accepting config
- performing internal setup
- creating the shared store
- creating named infrastructure instances and placing them into the store
- initializing logger infrastructure from `logger` config and selecting component loggers (`app`, `apiServer`, `rpcServer`, `jobScheduler`) from named loggers
- initializing framework runtime objects through an internal ordered setup pipeline
- exposing a setup-facing extension context and a business-facing assembly context
- collecting startup hooks, shutdown hooks, and runtime services
- constructing and returning `*app.App`

It is the outer application-construction layer.

---

## Public API

```go
type Action func(*Context) error

type SetupStep struct {
    Name string
    Run  func(*SetupContext) error
}

type Option func(*options)

func With(actions ...Action) Option
func WithSetupSteps(steps ...SetupStep) Option
func Load(path string, opts ...Option) (*app.App, error)
func New(cfg *config.Config, opts ...Option) (*app.App, error)
```

For most applications, `Load(...)` is the primary entrypoint.

---

## Configuration model

A typical framework config shape is:

```yaml
logger:
  - name: default
    level: info
    format: text
    output: stdout

  - name: http
    level: info
    format: json
    output: ./logs/http.log

  - name: rpc
    level: debug
    format: text
    output: ./logs/rpc.log

  - name: jobs
    level: info
    format: text
    output: ./logs/jobs.log

app:
  logger: default
  shutdownTimeout: 30s

apiServer:
  logger: http
  name: demo
  host: 0.0.0.0
  port: 8090

rpcServer:
  logger: rpc
  name: demo
  host: 0.0.0.0
  port: 9001
  advertise:
    address: 127.0.0.1
    etcd: default

jobScheduler:
  logger: jobs

rpcResolver:
  direct: true
  etcd: default
```

Semantics:

- `logger`: defines named logger instances and is handled like other infrastructure config sections
- `app.logger`: selects the default logger used by the assembled app
- `apiServer.logger`, `rpcServer.logger`, `jobScheduler.logger`: optionally override the app logger for those builtin components
- `rpcServer.advertise.address`: publishes the service instance address to service discovery
- `rpcServer.advertise.etcd`: selects the named etcd client used for service registration
- `rpcResolver.direct`: registers the `direct:///` resolver scheme for RPC clients
- `rpcResolver.etcd`: selects the named etcd client used to register the `etcd:///` resolver scheme
- `app.shutdownTimeout`: configures graceful shutdown timeout

All configured loggers are created during builtin setup and placed into the shared store.
The app logger is selected from the configured named loggers via `app.logger`.
Builtin components then either:

- use their own configured logger override, or
- fall back to the app logger when no component logger is configured

---

## Custom setup extension model

In addition to builtin setup, applications may contribute custom setup steps.
A custom setup step runs after builtin framework setup and before business `Action`s.
It is the intended extension point for preparing shared infrastructure that business assembly will later consume from the store.
Builtin setup steps and custom setup steps both follow the same step-driven model: each step reads the config section it needs, performs setup, and writes setup results into shared runtime state or the store.

Typical uses include:

- initializing sqlite or other custom database clients
- creating third-party SDK clients
- setting up feature-flag, metrics, or messaging clients
- registering shared infrastructure resources into the store for later use by actions

The execution order is:

```text
builtin setup -> custom setup steps -> business actions -> app assembly
```

`SetupContext` intentionally exposes a narrow capability surface:

```go
func DecodeSetupConfig[T any](ctx *SetupContext, key string) (*T, error)
func (c *SetupContext) Logger() *xlog.Logger
func (c *SetupContext) NamedLogger(name string) (*xlog.Logger, error)
func (c *SetupContext) Provide(name string, value any, opts ...store.SetOption) error
```

Semantics:

- `DecodeSetupConfig(...)`: decodes a config subtree for a custom setup step
- `Logger()`: returns the app logger for ordinary setup logging
- `NamedLogger(name)`: selects a specific configured logger by name
- `Provide(...)`: registers a shared infrastructure resource into the store for later setup steps or actions

Custom setup steps should generally focus on infrastructure preparation, not business assembly.
Business module wiring, API/RPC registration, jobs, hooks, and custom runtime services should remain in `Action`.
In practice, setup steps are the recommended place to write shared resources into the store, while actions are the recommended place to read those resources and assemble business modules.

## Business integration model

Business code contributes assembly through `Action`:

```go
func AssembleUser(ctx *assemble.Context) error {
    return ctx.RegisterAPI(func(engine *api.Engine) {
        // register user routes
    })
}
```

An action may:

- register API routes
- register gRPC handlers
- register jobs
- add startup hooks
- add shutdown hooks
- add custom runtime services
- read shared dependencies from the store

Jobs are backed by the default in-process scheduler created during setup, so `RegisterJob(...)` is available by default.
The scheduler itself can also choose a dedicated logger through `jobScheduler.logger`; otherwise it uses the app logger.

---

## Context API

```go
func (c *Context) Logger() *xlog.Logger
func (c *Context) Store() store.Store
func (c *Context) RegisterAPI(fn func(*api.Engine)) error
func (c *Context) RegisterRPC(fn func(*grpc.Server)) error
func (c *Context) RegisterJob(name string, fn job.Func) error
func (c *Context) OnStartup(h app.StartupHook)
func (c *Context) OnShutdown(h app.ShutdownHook)
func (c *Context) AddService(s app.Service)
```

`Context` is intentionally small.
It is the only assembly surface business code should need.
Internally, it is the business-facing capability view over assemble's private setup state plus collectors for hooks and custom services.

`Store()` is primarily intended for reading shared dependencies assembled by builtin setup or custom setup steps. Writing new shared resources directly through `Store()` should be done sparingly and with clear project-level conventions; for setup-time registration, prefer `SetupContext.Provide(...)`.

---

## Resource ownership

`pkg/assemble` creates shared resources and places them into `store.Store`.

The store is then injected into `pkg/app`, and `pkg/app` closes the store during shutdown.
Builtin services are registered after custom services, so during shutdown the builtin services stop first due to reverse-order teardown.

This means store-managed resources such as:

- loggers
- databases
- redis clients
- etcd clients

are all released through one shutdown path.

---

## Example

```go
package main

import (
    "context"
    "database/sql"

    _ "modernc.org/sqlite"

    "github.com/HorseArcher567/octopus/pkg/assemble"
    "github.com/HorseArcher567/octopus/pkg/store"
)

type SQLiteConfig struct {
    Name string `yaml:"name"`
    DSN  string `yaml:"dsn"`
}

func sqliteStep() assemble.SetupStep {
    return assemble.SetupStep{
        Name: "sqlite",
        Run: func(ctx *assemble.SetupContext) error {
            cfg, err := assemble.DecodeSetupConfig[SQLiteConfig](ctx, "sqlite")
            if err != nil {
                return err
            }

            db, err := sql.Open("sqlite", cfg.DSN)
            if err != nil {
                return err
            }
            if err := db.Ping(); err != nil {
                _ = db.Close()
                return err
            }
            if err := ctx.Provide(cfg.Name, db, store.WithClose(db.Close)); err != nil {
                _ = db.Close()
                return err
            }
            return nil
        },
    }
}

func main() {
    a, err := assemble.Load(
        "config.yaml",
        assemble.WithSetupSteps(sqliteStep()),
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

In later business actions, the provided resource can be read from the store:

```go
db, err := store.GetNamed[*sql.DB](ctx.Store(), "default")
```

---

## Internal setup model

Internally, `pkg/assemble` runs an ordered builtin setup pipeline directly against raw config, then applies any custom setup steps, and finally applies business `Action`s to construct the app.
Builtin and custom setup both participate in the same overall step-driven setup model, but builtin setup remains framework-controlled while custom setup is user-contributed through `WithSetupSteps(...)`.

Conceptually the flow is:

```text
raw config -> builtin setup steps -> custom setup steps -> action context -> app assembly
```

The setup pipeline is intentionally internal and ordered. It exists to keep the setup mainline short while still allowing framework setup capabilities to grow as a small, explicit list.

## Design intent

`pkg/assemble` is a facade.

Application developers should not need to know or manage:

- internal setup stages
- app construction details
- framework module systems
- resource ownership wiring

The intended experience is simple:

```text
build my app from config and business actions
```

That is the core purpose of this package.
