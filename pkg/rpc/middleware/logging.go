package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/HorseArcher567/octopus/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryServerLogging 为 Unary RPC 提供日志中间件
func UnaryServerLogging() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// 创建带有请求信息的 logger（继承上游可能已有的字段）
		log := logger.FromContext(ctx).With("method", info.FullMethod)
		if requestID := extractRequestID(ctx); requestID != "" {
			log = log.With("request_id", requestID)
		}

		log.Info("grpc request started")

		// 将 logger 注入 context
		ctx = logger.WithContext(ctx, log)

		resp, err := handler(ctx, req)

		duration := time.Since(start)

		if err != nil {
			st := status.Convert(err)
			log.Error("grpc request failed",
				"duration", duration,
				"code", st.Code().String(),
				"error", st.Message(),
			)
		} else {
			log.Info("grpc request completed",
				"duration", duration,
			)
		}

		return resp, err
	}
}

// StreamServerLogging 为 Stream RPC 提供日志中间件
func StreamServerLogging() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		ctx := ss.Context()

		log := logger.FromContext(ctx).With("method", info.FullMethod)
		if requestID := extractRequestID(ctx); requestID != "" {
			log = log.With("request_id", requestID)
		}

		log.Info("grpc stream started",
			"is_client_stream", info.IsClientStream,
			"is_server_stream", info.IsServerStream,
		)

		// 包装 ServerStream 以注入 logger
		wrappedStream := &loggingServerStream{
			ServerStream: ss,
			ctx:          logger.WithContext(ctx, log),
		}

		err := handler(srv, wrappedStream)

		duration := time.Since(start)

		if err != nil {
			st := status.Convert(err)
			log.Error("grpc stream failed",
				"duration", duration,
				"code", st.Code().String(),
				"error", st.Message(),
			)
		} else {
			log.Info("grpc stream completed",
				"duration", duration,
			)
		}

		return err
	}
}

// UnaryClientLogging 为客户端 Unary RPC 提供日志中间件
func UnaryClientLogging() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()
		log := logger.FromContext(ctx)

		log.Debug("grpc client request started",
			"method", method,
			"target", cc.Target(),
		)

		err := invoker(ctx, method, req, reply, cc, opts...)

		duration := time.Since(start)

		if err != nil {
			st := status.Convert(err)
			if st.Code() != codes.Canceled {
				log.Error("grpc client request failed",
					"method", method,
					"target", cc.Target(),
					"duration", duration,
					"code", st.Code().String(),
					"error", st.Message(),
				)
			}
		} else {
			log.Debug("grpc client request completed",
				"method", method,
				"target", cc.Target(),
				"duration", duration,
			)
		}

		return err
	}
}

// StreamClientLogging 为客户端 Stream RPC 提供日志中间件
func StreamClientLogging() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		start := time.Now()
		log := logger.FromContext(ctx)

		log.Debug("grpc client stream started",
			"method", method,
			"target", cc.Target(),
			"is_client_stream", desc.ClientStreams,
			"is_server_stream", desc.ServerStreams,
		)

		clientStream, err := streamer(ctx, desc, cc, method, opts...)

		if err != nil {
			st := status.Convert(err)
			log.Error("grpc client stream creation failed",
				"method", method,
				"target", cc.Target(),
				"duration", time.Since(start),
				"code", st.Code().String(),
				"error", st.Message(),
			)
			return nil, err
		}

		return &loggingClientStream{
			ClientStream: clientStream,
			log:          log,
			method:       method,
			target:       cc.Target(),
			start:        start,
		}, nil
	}
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

type loggingServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *loggingServerStream) Context() context.Context {
	return s.ctx
}

type loggingClientStream struct {
	grpc.ClientStream
	log    *slog.Logger
	method string
	target string
	start  time.Time
}

func (s *loggingClientStream) RecvMsg(m interface{}) error {
	err := s.ClientStream.RecvMsg(m)
	if err != nil {
		duration := time.Since(s.start)
		st := status.Convert(err)
		if st.Code() != codes.Canceled {
			s.log.Debug("grpc client stream completed",
				"method", s.method,
				"target", s.target,
				"duration", duration,
				"code", st.Code().String(),
			)
		}
	}
	return err
}
