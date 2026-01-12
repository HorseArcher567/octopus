package xlog

import (
	"context"
	"log/slog"
)

type loggerKey struct{}

// FromContext 从 context 获取 *slog.Logger，如果没有则返回 slog.Default()
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

// WithContext 将 *slog.Logger 存入 context
func WithContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

// WithAttrs 为 context 中的 logger 添加属性，返回新的 context
func WithAttrs(ctx context.Context, args ...any) context.Context {
	return WithContext(ctx, FromContext(ctx).With(args...))
}
