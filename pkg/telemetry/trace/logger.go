package trace

import (
	"context"

	"github.com/HorseArcher567/octopus/pkg/xlog"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// EnrichLogger attaches trace metadata from context to the logger when present.
func EnrichLogger(ctx context.Context, l *xlog.Logger) *xlog.Logger {
	if l == nil {
		return nil
	}
	spanCtx := oteltrace.SpanContextFromContext(ctx)
	if !spanCtx.IsValid() {
		return l
	}
	return l.With("trace_id", spanCtx.TraceID().String(), "span_id", spanCtx.SpanID().String())
}
