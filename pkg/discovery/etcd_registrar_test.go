package discovery

import (
	"context"
	"testing"

	"github.com/HorseArcher567/octopus/pkg/xlog"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestEtcdRegistrarRejectsDuplicateActiveRegister(t *testing.T) {
	log := xlog.MustNew(nil)
	defer log.Close()

	r := NewEtcdRegistrar(log, &clientv3.Client{})
	r.key = "/octopus/rpc/apps/demo/127.0.0.1:9001"

	err := r.Register(context.Background(), Instance{Name: "demo", Host: "127.0.0.1", Port: 9001})
	if err == nil {
		t.Fatal("expected duplicate active register to fail")
	}
}

func TestEtcdRegistrarRejectsMismatchedDeregister(t *testing.T) {
	log := xlog.MustNew(nil)
	defer log.Close()

	r := NewEtcdRegistrar(log, &clientv3.Client{})
	r.key = "/octopus/rpc/apps/demo/127.0.0.1:9001"

	err := r.Deregister(context.Background(), Instance{Name: "demo", Host: "127.0.0.1", Port: 9002})
	if err == nil {
		t.Fatal("expected mismatched deregister to fail")
	}
}

func TestEtcdRegistrarDeregisterWithoutActiveInstance(t *testing.T) {
	log := xlog.MustNew(nil)
	defer log.Close()

	r := NewEtcdRegistrar(log, &clientv3.Client{})
	if err := r.Deregister(context.Background(), Instance{Name: "demo", Host: "127.0.0.1", Port: 9001}); err != nil {
		t.Fatalf("expected empty deregister to succeed: %v", err)
	}
}
