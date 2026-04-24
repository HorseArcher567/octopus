package rpc

import (
	"sync"

	grpcresolver "google.golang.org/grpc/resolver"
)

var resolverRegistry struct {
	mu         sync.Mutex
	registered map[string]struct{}
}

func init() {
	resolverRegistry.registered = make(map[string]struct{})
}

// RegisterResolver registers a gRPC resolver builder by scheme once per process.
// It returns true when the scheme is newly registered and false when it was already registered.
func RegisterResolver(builder grpcresolver.Builder) bool {
	if builder == nil {
		return false
	}
	scheme := builder.Scheme()
	resolverRegistry.mu.Lock()
	defer resolverRegistry.mu.Unlock()
	if _, ok := resolverRegistry.registered[scheme]; ok {
		return false
	}
	grpcresolver.Register(builder)
	resolverRegistry.registered[scheme] = struct{}{}
	return true
}
