package resource

import "testing"

func TestManagerCloseNil(t *testing.T) {
	var manager *Manager
	if err := manager.Close(); err != nil {
		t.Fatalf("expected nil close error, got %v", err)
	}
}

func TestManagerMissingNamedResources(t *testing.T) {
	manager := &Manager{
		resources: make(map[string]map[string]entry),
	}

	if _, err := manager.Get(KindMySQL, "primary"); err == nil {
		t.Fatal("expected missing mysql resource error")
	}
	if _, err := manager.Get(KindRedis, "cache"); err == nil {
		t.Fatal("expected missing redis resource error")
	}
}

func TestManagerRegisterAndList(t *testing.T) {
	manager := &Manager{resources: make(map[string]map[string]entry)}
	if err := manager.Register("custom", "main", struct{}{}, nil); err != nil {
		t.Fatalf("register: %v", err)
	}
	if got := manager.List("custom"); len(got) != 1 || got[0] != "main" {
		t.Fatalf("unexpected list: %v", got)
	}
}
