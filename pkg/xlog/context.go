package xlog

import (
	"context"
	"log/slog"
)

type loggerKey struct{}

// FromContext returns the *Logger from context, or a Logger wrapping slog.Default() if not found.
func FromContext(ctx context.Context) *Logger {
	if l, ok := ctx.Value(loggerKey{}).(*Logger); ok {
		return l
	}
	return &Logger{Logger: slog.Default()}
}

// WithContext stores the *Logger in context.
func WithContext(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

// WithAttrs adds attributes to the logger in context and returns a new context.
func WithAttrs(ctx context.Context, args ...any) context.Context {
	logger := FromContext(ctx)
	return WithContext(ctx, &Logger{Logger: logger.With(args...)})
}
