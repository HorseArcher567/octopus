package assemble

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/config"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"google.golang.org/grpc"
)

type testService struct {
	name string
	run  func(context.Context) error
	stop func(context.Context) error
}

func (s *testService) Name() string { return s.name }
func (s *testService) Run(ctx context.Context) error {
	if s.run != nil {
		return s.run(ctx)
	}
	<-ctx.Done()
	return ctx.Err()
}
func (s *testService) Stop(ctx context.Context) error {
	if s.stop != nil {
		return s.stop(ctx)
	}
	return nil
}

func minimalConfig() *config.Config {
	cfg := config.New()
	cfg.Set("logger", []any{
		map[string]any{
			"name":  "default",
			"level": "debug",
		},
	})
	cfg.Set("app.logger", "default")
	return cfg
}

func TestNew_NilConfig(t *testing.T) {
	_, err := New(nil)
	if err == nil || err.Error() != "assemble: config cannot be nil" {
		t.Fatalf("New(nil) error = %v", err)
	}
}

func TestNew_ActionError(t *testing.T) {
	cfg := minimalConfig()
	_, err := New(cfg, With(func(ctx *Context) error {
		return errors.New("boom")
	}))
	if err == nil || err.Error() != "assemble: action 0: boom" {
		t.Fatalf("New() error = %v", err)
	}
}

func TestNew_ActionsCanRegisterHooksAndServices(t *testing.T) {
	cfg := minimalConfig()

	var mu sync.Mutex
	var events []string
	record := func(v string) {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, v)
	}

	a, err := New(cfg, With(func(ctx *Context) error {
		ctx.OnStartup(func(context.Context) error {
			record("startup")
			return nil
		})
		ctx.OnShutdown(func(context.Context) error {
			record("shutdown")
			return nil
		})
		ctx.AddService(&testService{
			name: "custom",
			run: func(ctx context.Context) error {
				record("service-run")
				return nil
			},
			stop: func(ctx context.Context) error {
				record("service-stop")
				return nil
			},
		})
		return nil
	}))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := a.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	want := []string{"startup", "service-run", "service-stop", "shutdown"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("logger:\n  - name: default\n    level: debug\napp:\n  logger: default\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	a, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if a == nil {
		t.Fatalf("Load() returned nil app")
	}
}

func TestContext_RegisterWithoutConfiguredRuntime(t *testing.T) {
	cfg := minimalConfig()
	_, err := New(cfg, With(func(ctx *Context) error {
		if err := ctx.RegisterAPI(func(*api.Engine) {}); err == nil || !errors.Is(err, ErrAPINotConfigured) {
			return errors.New("expected api not configured")
		}
		if err := ctx.RegisterRPC(func(*grpc.Server) {}); err == nil || !errors.Is(err, ErrRPCNotConfigured) {
			return errors.New("expected rpc not configured")
		}
		return nil
	}))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
}

func TestContext_RegisterJob_DefaultSchedulerAvailable(t *testing.T) {
	cfg := minimalConfig()
	_, err := New(cfg, With(func(ctx *Context) error {
		return ctx.RegisterJob("job", func(context.Context, *xlog.Logger) error { return nil })
	}))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
}

func TestNew_AppLoggerMustExistInConfiguredLoggers(t *testing.T) {
	cfg := config.New()
	cfg.Set("logger", []any{
		map[string]any{
			"name":  "default",
			"level": "debug",
		},
	})
	cfg.Set("app.logger", "missing")

	_, err := New(cfg)
	if err == nil || err.Error() != "assemble: setup app-logger: assemble: app.logger: logger \"missing\" not found" {
		t.Fatalf("New() error = %v", err)
	}
}

func TestNew_APIServerLoggerFallsBackToAppLogger(t *testing.T) {
	cfg := minimalConfig()
	cfg.Set("apiServer", map[string]any{
		"name": "demo",
		"host": "127.0.0.1",
		"port": 8080,
		"mode": "release",
	})

	_, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
}

func TestNew_ComponentLoggerMustExistInConfiguredLoggers(t *testing.T) {
	cfg := minimalConfig()
	cfg.Set("apiServer", map[string]any{
		"name":   "demo",
		"host":   "127.0.0.1",
		"port":   8080,
		"mode":   "release",
		"logger": "missing",
	})

	_, err := New(cfg)
	if err == nil || err.Error() != "assemble: setup api: assemble: apiServer.logger: logger \"missing\" not found" {
		t.Fatalf("New() error = %v", err)
	}
}

func TestNew_JobSchedulerLoggerMustExistInConfiguredLoggers(t *testing.T) {
	cfg := minimalConfig()
	cfg.Set("jobScheduler.logger", "missing")

	_, err := New(cfg)
	if err == nil || err.Error() != "assemble: setup jobs: assemble: jobScheduler.logger: logger \"missing\" not found" {
		t.Fatalf("New() error = %v", err)
	}
}
