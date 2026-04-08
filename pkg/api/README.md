# pkg/api

`pkg/api` provides the HTTP runtime used by Octopus.

It wraps a Gin engine with:

- standard middleware
- optional `pprof`
- `Register(...)` for route assembly
- `Run(ctx)` / `Stop(ctx)` lifecycle methods
