package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/HorseArcher567/octopus/pkg/config"
)

func resetDefaultAppForTest() {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	defaultApp = nil
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

func TestInitPanicWhenCalledTwice(t *testing.T) {
	resetDefaultAppForTest()
	t.Cleanup(resetDefaultAppForTest)

	cfg := mustLoadTestConfig(t)
	Init(cfg)
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when calling Init twice")
		}
	}()
	Init(cfg)
}

func TestInit(t *testing.T) {
	resetDefaultAppForTest()
	t.Cleanup(resetDefaultAppForTest)

	path := writeTestConfig(t)
	Init(mustLoadTestConfigPath(path))
	if Default() == nil {
		t.Fatalf("default app should be initialized")
	}
}

func TestMustRun(t *testing.T) {
	resetDefaultAppForTest()
	t.Cleanup(resetDefaultAppForTest)

	path := writeTestConfig(t)
	wired := false
	oldRunFn := runWithSignalsFn
	runWithSignalsFn = func() error { return nil }
	t.Cleanup(func() { runWithSignalsFn = oldRunFn })

	MustRun(path, func() error {
		wired = true
		return nil
	})

	if !wired {
		t.Fatalf("wire function should be called")
	}
}

func mustLoadTestConfig(t *testing.T) *config.Config {
	t.Helper()
	path := writeTestConfig(t)
	return mustLoadTestConfigPath(path)
}

func mustLoadTestConfigPath(path string) *config.Config {
	return config.MustLoad(path)
}
