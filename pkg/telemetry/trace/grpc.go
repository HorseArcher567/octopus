package trace

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc/stats"
)

// ServerHandler returns a gRPC stats handler for tracing instrumentation.
func ServerHandler() stats.Handler {
	return otelgrpc.NewServerHandler()
}
