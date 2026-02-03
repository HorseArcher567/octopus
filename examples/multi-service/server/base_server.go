package main

import (
	"context"
	"log/slog"

	"github.com/HorseArcher567/octopus/pkg/xlog"
	"google.golang.org/grpc/metadata"
)

// BaseServer 提供所有 Server 的公共功能，包括 logger 管理
type BaseServer struct {
	logger *xlog.Logger
}

// Log 获取带请求上下文的 logger
// 优先使用 context 中已有的 logger（中间件已注入 method 和 request_id）
// 如果没有，则使用 app logger 并手动添加请求字段
func (b *BaseServer) Log(ctx context.Context, attrs ...any) *xlog.Logger {
	// 尝试从 context 获取 logger（中间件已注入，包含 method 和 request_id）
	ctxLog := xlog.FromContext(ctx)

	// 如果 context 中的 logger 不是默认的 slog.Default()，说明中间件已经注入了
	// 直接使用它并添加额外字段
	if ctxLog != nil && ctxLog.Logger != slog.Default() {
		if len(attrs) > 0 {
			// With 返回 *slog.Logger，需要包装回 *xlog.Logger
			return &xlog.Logger{Logger: ctxLog.With(attrs...)}
		}
		return ctxLog
	}

	// Fallback：使用 app logger 并手动添加请求字段
	logger := b.logger

	// 从 metadata 提取 request_id
	if requestID := extractRequestID(ctx); requestID != "" {
		logger = &xlog.Logger{Logger: logger.With("request_id", requestID)}
	}

	// 添加额外的字段
	if len(attrs) > 0 {
		logger = &xlog.Logger{Logger: logger.With(attrs...)}
	}

	return logger
}

// extractRequestID 从 gRPC metadata 中提取 request_id
func extractRequestID(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if values := md.Get("x-request-id"); len(values) > 0 {
			return values[0]
		}
	}
	return ""
}
