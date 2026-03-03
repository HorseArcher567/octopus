package xlog

import (
	"context"
	"log/slog"
)

type loggerKey struct{}

// Get returns the logger stored in ctx.
// If none is stored, it returns a wrapper around slog.Default().
// Get never mutates ctx.
func Get(ctx context.Context) *Logger {
	if l, ok := Lookup(ctx); ok {
		return l
	}
	return &Logger{Logger: slog.Default()}
}

// Lookup reports whether ctx contains a logger and returns it when present.
func Lookup(ctx context.Context) (*Logger, bool) {
	l, ok := ctx.Value(loggerKey{}).(*Logger)
	return l, ok
}

// Put returns a derived context that carries l.
func Put(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

// GetOr returns the logger in ctx, or fallback when missing.
// Unlike Get, this function never falls back to slog.Default().
func GetOr(ctx context.Context, fallback *Logger) *Logger {
	if l, ok := Lookup(ctx); ok {
		return l
	}
	return fallback
}
