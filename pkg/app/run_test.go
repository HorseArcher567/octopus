package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/HorseArcher567/octopus/pkg/config"
)

type lifecycleModule struct {
	id string

	initErr  error
	closeErr error

	inited bool
	closed bool
}

func (m *lifecycleModule) ID() string { return m.id }

func (m *lifecycleModule) Init(_ context.Context, _ Runtime) error {
	m.inited = true
	return m.initErr
}

func (m *lifecycleModule) Close(_ context.Context) error {
	m.closed = true
	return m.closeErr
}

func TestRunInitError(t *testing.T) {
	cfgPath := writeTestConfig(t)
	mod := &lifecycleModule{id: "broken-init", initErr: errors.New("boom")}

	err := Run(cfgPath, []Module{mod}, WithContext(context.Background()))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), `module "broken-init" init`) {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mod.inited {
		t.Fatalf("expected module initialized state true: %+v", mod)
	}
}

func TestRunRequiresConfigPath(t *testing.T) {
	err := Run("", nil, WithContext(context.Background()))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "config path is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMustRun(t *testing.T) {
	path := writeTestConfig(t)
	mod := &lifecycleModule{id: "ok"}

	MustRun(path, []Module{mod}, WithContext(context.Background()))

	if !mod.inited || !mod.closed {
		t.Fatalf("module lifecycle should run: %+v", mod)
	}
}

func TestRunModuleDependencyMissing(t *testing.T) {
	path := writeTestConfig(t)
	mod := &depModule{lifecycleModule: lifecycleModule{id: "rpc"}, deps: []string{"infra"}}

	err := Run(path, []Module{mod}, WithContext(context.Background()))
	if err == nil || !strings.Contains(err.Error(), `depends on unknown module "infra"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

type depModule struct {
	lifecycleModule
	deps []string
}

func (m *depModule) DependsOn() []string { return m.deps }

func TestResolveModuleOrderDuplicateID(t *testing.T) {
	m1 := &lifecycleModule{id: "dup"}
	m2 := &lifecycleModule{id: "dup"}
	_, err := resolveModuleOrder([]Module{m1, m2})
	if err == nil || !strings.Contains(err.Error(), `duplicate module id "dup"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveModuleOrderCycle(t *testing.T) {
	m1 := &depModule{lifecycleModule: lifecycleModule{id: "a"}, deps: []string{"b"}}
	m2 := &depModule{lifecycleModule: lifecycleModule{id: "b"}, deps: []string{"a"}}
	_, err := resolveModuleOrder([]Module{m1, m2})
	if err == nil || !strings.Contains(err.Error(), "module dependency cycle detected") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunInitRollbackReverseOrder(t *testing.T) {
	path := writeTestConfig(t)
	closed := make([]string, 0, 2)

	ok1 := &trackingModule{id: "m1", onClose: func(id string) { closed = append(closed, id) }}
	ok2 := &trackingModule{id: "m2", onClose: func(id string) { closed = append(closed, id) }}
	fail := &trackingModule{id: "m3", initErr: errors.New("init fail")}

	err := Run(path, []Module{ok1, ok2, fail}, WithContext(context.Background()))
	if err == nil || !strings.Contains(err.Error(), `module "m3" init`) {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(closed) != 2 || closed[0] != "m2" || closed[1] != "m1" {
		t.Fatalf("unexpected close order: %v", closed)
	}
	if ok1.closeCalls != 1 || ok2.closeCalls != 1 {
		t.Fatalf("expected one close call per initialized module, got m1=%d m2=%d", ok1.closeCalls, ok2.closeCalls)
	}
}

func TestNewRPCClientReuseAndClose(t *testing.T) {
	path := writeTestConfig(t)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	a := New(cfg)
	defer a.Logger().Close()

	target := "127.0.0.1:65535"
	c1, err := a.NewRPCClient(target)
	if err != nil {
		t.Fatalf("new rpc client #1: %v", err)
	}
	c2, err := a.NewRPCClient(target)
	if err != nil {
		t.Fatalf("new rpc client #2: %v", err)
	}
	if c1 != c2 {
		t.Fatal("expected cached client connection to be reused")
	}

	a.CloseRpcClients()

	c3, err := a.NewRPCClient(target)
	if err != nil {
		t.Fatalf("new rpc client #3: %v", err)
	}
	if c3 == c1 {
		t.Fatal("expected a new connection after closing cached clients")
	}
	a.CloseRpcClients()
}

type trackingModule struct {
	id         string
	initErr    error
	closeCalls int
	onClose    func(id string)
}

func (m *trackingModule) ID() string { return m.id }

func (m *trackingModule) Init(_ context.Context, _ Runtime) error { return m.initErr }

func (m *trackingModule) Close(_ context.Context) error {
	m.closeCalls++
	if m.onClose != nil {
		m.onClose(m.id)
	}
	return nil
}

func writeTestConfig(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("logger:\n  level: info\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write config failed: %v", err)
	}
	return path
}
