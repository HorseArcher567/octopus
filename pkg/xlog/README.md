# pkg/xlog

`pkg/xlog` provides the logger used by Octopus.

It offers:

- structured logging based on `log/slog`
- JSON and text output
- configurable log levels
- optional file output and rotation support
- context-aware logger propagation

Basic usage:

```go
log := xlog.MustNew(nil)
defer log.Close()

log.Info("application started", "version", "1.0.0")
```

Context helpers:

- `xlog.Put(ctx, log)`
- `xlog.Get(ctx)`
- `xlog.Lookup(ctx)`
- `xlog.GetOr(ctx, fallback)`

Tracing integration helpers live in `pkg/observability/trace`, e.g. `trace.EnrichLogger(ctx, log)`.
