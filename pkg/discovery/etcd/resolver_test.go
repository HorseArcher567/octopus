package etcd

import (
	"context"
	"testing"

	"github.com/HorseArcher567/octopus/pkg/discovery"
	"github.com/HorseArcher567/octopus/pkg/xlog"
)

func TestProviderExposesComponents(t *testing.T) {
	p := NewProvider(xlog.MustNew(nil), nil)
	if p.Registrar() == nil {
		t.Fatal("expected registrar")
	}
	if p.Resolver() == nil {
		t.Fatal("expected resolver")
	}
	if p.GRPCResolverBuilder() == nil {
		t.Fatal("expected grpc resolver builder")
	}
}

func TestPrefix(t *testing.T) {
	got := prefix("svc")
	want := "/octopus/rpc/apps/svc/"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

type fakeResolver struct{}

func (fakeResolver) Resolve(context.Context, string) ([]discovery.Instance, error) {
	return []discovery.Instance{{Service: "svc", Address: "127.0.0.1", Port: 8080}}, nil
}

func (fakeResolver) Watch(ctx context.Context, service string) (<-chan []discovery.Instance, error) {
	ch := make(chan []discovery.Instance, 1)
	ch <- []discovery.Instance{{Service: service, Address: "127.0.0.1", Port: 8080}}
	close(ch)
	return ch, nil
}
