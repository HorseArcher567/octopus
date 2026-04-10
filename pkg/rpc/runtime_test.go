package rpc

import (
	"testing"

	"github.com/HorseArcher567/octopus/pkg/xlog"
)

func TestClientFactoryReuseAndClose(t *testing.T) {
	log := xlog.MustNew(nil)
	defer log.Close()

	factory := NewClientFactory(log, nil, nil, nil)
	target := "127.0.0.1:65535"

	c1, err := factory.Client(target)
	if err != nil {
		t.Fatalf("client #1: %v", err)
	}
	c2, err := factory.Client(target)
	if err != nil {
		t.Fatalf("client #2: %v", err)
	}
	if c1 != c2 {
		t.Fatal("expected cached client connection to be reused")
	}

	if err := factory.Close(); err != nil {
		t.Fatalf("close clients: %v", err)
	}

	c3, err := factory.Client(target)
	if err != nil {
		t.Fatalf("client #3: %v", err)
	}
	if c3 == c1 {
		t.Fatal("expected a new connection after Close")
	}
	if err := factory.Close(); err != nil {
		t.Fatalf("close clients #2: %v", err)
	}
}

func TestClientFactoryRequiresEtcdClientForEtcdTargets(t *testing.T) {
	log := xlog.MustNew(nil)
	defer log.Close()

	factory := NewClientFactory(log, nil, nil, nil)
	if _, err := factory.Client("etcd:///demo"); err == nil {
		t.Fatal("expected etcd target to fail without etcd client")
	}
}

func TestRuntimeRegisterRequiresServer(t *testing.T) {
	log := xlog.MustNew(nil)
	defer log.Close()

	rt, err := NewRuntime(log, nil)
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	defer rt.Close()

	if err := rt.Register(nil); err == nil {
		t.Fatal("expected register to fail without configured server")
	}
}
