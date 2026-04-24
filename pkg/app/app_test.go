package app

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/HorseArcher567/octopus/pkg/hook"
)

type testService struct {
	name    string
	runFn   func(context.Context) error
	stopFn  func(context.Context) error
	runCnt  int
	stopCnt int
	mu      sync.Mutex
}

func (s *testService) Name() string { return s.name }

func (s *testService) Run(ctx context.Context) error {
	s.mu.Lock()
	s.runCnt++
	s.mu.Unlock()
	if s.runFn != nil {
		return s.runFn(ctx)
	}
	<-ctx.Done()
	return ctx.Err()
}

func (s *testService) Stop(ctx context.Context) error {
	s.mu.Lock()
	s.stopCnt++
	s.mu.Unlock()
	if s.stopFn != nil {
		return s.stopFn(ctx)
	}
	return nil
}

func TestAppRun_StartsServicesAndStopsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	var events []string
	record := func(v string) {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, v)
	}

	svc := &testService{
		name: "svc",
		runFn: func(ctx context.Context) error {
			record("service-run")
			<-ctx.Done()
			return ctx.Err()
		},
		stopFn: func(ctx context.Context) error {
			record("service-stop")
			return nil
		},
	}

	a := New(nil).
		OnStartup(func(*hook.Context) error {
			record("startup-1")
			return nil
		}).
		OnStartup(func(*hook.Context) error {
			record("startup-2")
			return nil
		}).
		OnShutdown(func(*hook.Context) error {
			record("shutdown-1")
			return nil
		}).
		OnShutdown(func(*hook.Context) error {
			record("shutdown-2")
			return nil
		}).
		AddServices(svc)

	done := make(chan error, 1)
	go func() { done <- a.Run(ctx) }()

	time.Sleep(50 * time.Millisecond)
	cancel()

	if err := <-done; err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	want := []string{"startup-1", "startup-2", "service-run", "service-stop", "shutdown-2", "shutdown-1"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestAppRun_StartupHookErrorAbortsRun(t *testing.T) {
	svc := &testService{name: "svc"}
	a := New(nil).
		OnStartup(func(*hook.Context) error { return errors.New("boom") }).
		AddServices(svc)

	err := a.Run(context.Background())
	if err == nil || err.Error() != "startup hook 0: boom" {
		t.Fatalf("Run() error = %v", err)
	}
	if svc.runCnt != 0 {
		t.Fatalf("service should not run, runCnt=%d", svc.runCnt)
	}
}

func TestAppRun_ServiceErrorTriggersShutdown(t *testing.T) {
	svc := &testService{
		name: "svc",
		runFn: func(context.Context) error { return errors.New("svc boom") },
	}
	a := New(nil).AddServices(svc)

	err := a.Run(context.Background())
	if err == nil || err.Error() != "service \"svc\": svc boom" {
		t.Fatalf("Run() error = %v", err)
	}
	if svc.stopCnt != 1 {
		t.Fatalf("stopCnt = %d, want 1", svc.stopCnt)
	}
}

func TestAppRun_ShutdownHookErrorsAreJoined(t *testing.T) {
	a := New(nil).
		OnShutdown(func(*hook.Context) error { return errors.New("err1") }).
		OnShutdown(func(*hook.Context) error { return errors.New("err2") })

	err := a.Run(context.Background())
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, errors.New("err1")) && !errors.Is(err, errors.New("err2")) {
		// fall through to string checks below
	}
	msg := err.Error()
	if msg != "shutdown hook 1: err2\nshutdown hook 0: err1" && msg != "shutdown hook 0: err1\nshutdown hook 1: err2" {
		t.Fatalf("unexpected error = %v", err)
	}
}

func TestAppRun_CanOnlyRunOnce(t *testing.T) {
	a := New(nil)
	if err := a.Run(context.Background()); err != nil {
		t.Fatalf("first Run() error = %v", err)
	}
	if err := a.Run(context.Background()); err == nil || err.Error() != "app: Run can only be called once" {
		t.Fatalf("second Run() error = %v", err)
	}
}

func TestAppRun_NilContextUsesBackground(t *testing.T) {
	svc := &testService{
		name: "svc",
		runFn: func(ctx context.Context) error {
			select {
			case <-time.After(10 * time.Millisecond):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	}
	if err := New(nil).AddServices(svc).Run(nil); err != nil {
		t.Fatalf("Run(nil) error = %v", err)
	}
}

func TestAppRun_ServiceStopRespectsShutdownContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := &testService{
		name: "svc",
		runFn: func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		},
		stopFn: func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}

	a := New(nil, WithShutdownTimeout(20*time.Millisecond)).AddServices(svc)
	done := make(chan error, 1)
	go func() { done <- a.Run(ctx) }()
	time.Sleep(20 * time.Millisecond)
	cancel()

	err := <-done
	if err == nil || err.Error() != "stop service \"svc\": context deadline exceeded" {
		t.Fatalf("Run() error = %v", err)
	}
}
