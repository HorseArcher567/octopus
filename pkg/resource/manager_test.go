package resource

import (
	"testing"

	"github.com/HorseArcher567/octopus/pkg/database"
	redisclient "github.com/HorseArcher567/octopus/pkg/redis"
)

func TestManagerCloseNil(t *testing.T) {
	var manager *Manager
	if err := manager.Close(); err != nil {
		t.Fatalf("expected nil close error, got %v", err)
	}
}

func TestManagerMissingNamedResources(t *testing.T) {
	manager := &Manager{
		mysql: make(map[string]*database.DB),
		redis: make(map[string]*redisclient.Client),
	}

	if _, err := manager.MySQL("primary"); err == nil {
		t.Fatal("expected missing mysql resource error")
	}
	if _, err := manager.Redis("cache"); err == nil {
		t.Fatal("expected missing redis resource error")
	}
}
