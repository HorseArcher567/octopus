package app

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"
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

func TestAppRun_StartupHooksThenServicesThenShutdown(t *testing.T) {
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
		OnStartup(func(ctx context.Context) error {
			record("startup-1")
			return nil
		}).
		OnStartup(func(ctx context.Context) error {
			record("startup-2")
			return nil
		}).
		OnShutdown(func(ctx context.Context) error {
			record("shutdown-1")
			return nil
		}).
		OnShutdown(func(ctx context.Context) error {
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
		OnStartup(func(ctx context.Context) error { return errors.New("boom") }).
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
	var stopped bool
	svc := &testService{
		name:  "svc",
		runFn: func(ctx context.Context) error { return errors.New("run failed") },
		stopFn: func(ctx context.Context) error {
			stopped = true
			return nil
		},
	}

	err := New(nil).AddServices(svc).Run(context.Background())
	if err == nil || err.Error() != "service \"svc\": run failed" {
		t.Fatalf("Run() error = %v", err)
	}
	if !stopped {
		t.Fatalf("expected service stop to be called")
	}
}

func TestAppRun_ShutdownStopsServicesInReverseOrder(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	var stops []string
	recordStop := func(name string) {
		mu.Lock()
		defer mu.Unlock()
		stops = append(stops, name)
	}

	svc1 := &testService{
		name: "svc1",
		runFn: func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		},
		stopFn: func(ctx context.Context) error {
			recordStop("svc1")
			return nil
		},
	}
	svc2 := &testService{
		name: "svc2",
		runFn: func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		},
		stopFn: func(ctx context.Context) error {
			recordStop("svc2")
			return nil
		},
	}

	a := New(nil).AddServices(svc1, svc2)
	done := make(chan error, 1)
	go func() { done <- a.Run(ctx) }()

	time.Sleep(50 * time.Millisecond)
	cancel()

	if err := <-done; err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	want := []string{"svc2", "svc1"}
	if !reflect.DeepEqual(stops, want) {
		t.Fatalf("stops = %v, want %v", stops, want)
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

func TestAppRun_InternalCanceledErrorIsNotTreatedAsGracefulShutdown(t *testing.T) {
	svc := &testService{
		name:  "svc",
		runFn: func(ctx context.Context) error { return context.Canceled },
	}

	err := New(nil).AddServices(svc).Run(context.Background())
	if err == nil || err.Error() != "service \"svc\": context canceled" {
		t.Fatalf("Run() error = %v", err)
	}
}
