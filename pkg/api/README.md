# pkg/api

`pkg/api` provides the HTTP server used by Octopus.

It wraps a Gin engine with:

- built-in middleware stack
- configurable custom middleware via `WithMiddleware(...)`
- the ability to disable built-in middleware via `WithoutDefaultMiddleware()`
- optional `pprof`
- `Register(...)` for route assembly
- `Run(ctx)` / `Stop(ctx)` lifecycle methods

Default middleware stack:

- logger injection
- recovery
- request logging

Health and telemetry routes such as `/health` and `/metrics` are mounted by the app setup/bootstrap layer, not hard-coded in the HTTP server itself.
