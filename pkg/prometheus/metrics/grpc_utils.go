package metrics

import (
	"google.golang.org/grpc"
	"strings"
)

const (
	Unary        = "unary"
	ClientStream = "client_stream"
	ServerStream = "server_stream"
	BidiStream   = "bidi_stream"
)

func splitGrpcMethodName(fullMethodName string) (string, string) {
	fullMethodName = strings.TrimPrefix(fullMethodName, "/")
	if i := strings.Index(fullMethodName, "/"); i >= 0 {
		return fullMethodName[:i], fullMethodName[i+1:]
	}
	return "unknown", "unknown"
}

func grpcStreamType(info *grpc.StreamServerInfo) string {
	if info.IsClientStream && !info.IsServerStream {
		return ClientStream
	} else if !info.IsClientStream && info.IsServerStream {
		return ServerStream
	}
	return BidiStream
}
