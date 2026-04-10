package rpc

import (
	"testing"

	rpcmiddleware "github.com/HorseArcher567/octopus/pkg/rpc/middleware"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"google.golang.org/grpc"
)

func TestServerDefaultInterceptorsInstalled(t *testing.T) {
	log := xlog.MustNew(nil)
	defer log.Close()

	s, err := NewServer(log, &ServerConfig{
		Name: "rpc-test",
		Host: "127.0.0.1",
		Port: 50051,
	})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	if got := s.UnaryInterceptorCount(); got != 2 {
		t.Fatalf("expected 2 default unary interceptors, got %d", got)
	}
	if got := s.StreamInterceptorCount(); got != 2 {
		t.Fatalf("expected 2 default stream interceptors, got %d", got)
	}
}

func TestServerCustomInterceptorsAppended(t *testing.T) {
	log := xlog.MustNew(nil)
	defer log.Close()

	unary := grpc.UnaryServerInterceptor(rpcmiddleware.UnaryServerLogging())
	stream := grpc.StreamServerInterceptor(rpcmiddleware.StreamServerLogging())

	s, err := NewServer(log, &ServerConfig{
		Name: "rpc-test",
		Host: "127.0.0.1",
		Port: 50052,
	}, WithUnaryInterceptors(unary), WithStreamInterceptors(stream))
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	if got := s.UnaryInterceptorCount(); got != 3 {
		t.Fatalf("expected 3 unary interceptors, got %d", got)
	}
	if got := s.StreamInterceptorCount(); got != 3 {
		t.Fatalf("expected 3 stream interceptors, got %d", got)
	}
}
