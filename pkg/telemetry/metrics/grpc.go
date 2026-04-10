package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	grpcRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "octopus_grpc_requests_total",
			Help: "Total number of gRPC requests.",
		},
		[]string{"method", "code"},
	)
	grpcRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "octopus_grpc_request_duration_seconds",
			Help:    "gRPC request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "code"},
	)
)

func init() {
	prometheus.MustRegister(grpcRequestsTotal, grpcRequestDuration)
}

// UnaryServerInterceptor records basic unary gRPC server metrics.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		code := status.Code(err).String()
		grpcRequestsTotal.WithLabelValues(info.FullMethod, code).Inc()
		grpcRequestDuration.WithLabelValues(info.FullMethod, code).Observe(time.Since(start).Seconds())
		return resp, err
	}
}

// StreamServerInterceptor records basic stream gRPC server metrics.
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		err := handler(srv, ss)
		code := status.Code(err).String()
		grpcRequestsTotal.WithLabelValues(info.FullMethod, code).Inc()
		grpcRequestDuration.WithLabelValues(info.FullMethod, code).Observe(time.Since(start).Seconds())
		return err
	}
}
