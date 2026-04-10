// Package app provides Octopus application orchestration, bootstrap assembly,
// phased module execution, and lifecycle management.
package app

import (
	"context"
	"sync"
	"time"

	"github.com/HorseArcher567/octopus/pkg/telemetry"
	"github.com/HorseArcher567/octopus/pkg/xlog"
)

// StartupHook is executed before components are started.
// If it returns an error, startup is aborted.
type StartupHook func(ctx context.Context, a *App) error

// ShutdownHook is executed during shutdown.
// Even if it returns an error, subsequent shutdown hooks continue to run.
type ShutdownHook func(ctx context.Context, a *App) error

// Option customizes an App instance.
type Option func(*App)

// WithRPCRuntime injects the RPC runtime.
func WithRPCRuntime(rt RPCRuntime) Option {
	return func(a *App) {
		a.rpc = rt
	}
}

// WithAPIRuntime injects the API runtime.
func WithAPIRuntime(rt APIRuntime) Option {
	return func(a *App) {
		a.api = rt
	}
}

// WithJobRuntime injects the job runtime.
func WithJobRuntime(rt JobRuntime) Option {
	return func(a *App) {
		a.jobs = rt
	}
}

// WithResourceRuntime injects the resource runtime.
func WithResourceRuntime(rt ResourceRuntime) Option {
	return func(a *App) {
		a.resources = rt
	}
}

// WithTelemetry injects the telemetry runtime.
func WithTelemetry(rt *telemetry.Runtime) Option {
	return func(a *App) {
		a.telemetry = rt
	}
}

// WithShutdownTimeout overrides the default shutdown timeout.
func WithShutdownTimeout(timeout time.Duration) Option {
	return func(a *App) {
		a.shutdownTimeout = timeout
	}
}

// App encapsulates module orchestration and application lifecycle.
type App struct {
	log *xlog.Logger

	rpc       RPCRuntime
	api       APIRuntime
	jobs      JobRuntime
	resources ResourceRuntime
	telemetry *telemetry.Runtime
	container *container

	shutdownTimeout time.Duration
	startupHooks    []StartupHook
	shutdownHooks   []ShutdownHook

	modules         []Module
	orderedModules  []Module
	activeClosers   []moduleCloser
	activeCloserIDs map[string]struct{}

	shutdownOnce sync.Once
	runMu        sync.Mutex
	hasRun       bool
}

type moduleCloser struct {
	id string
	fn CloseModule
}

// New creates a new App from explicitly injected runtimes.
func New(log *xlog.Logger, opts ...Option) *App {
	if log == nil {
		log = xlog.MustNew(nil)
	}
	a := &App{
		log:             log,
		container:       newContainer(),
		activeCloserIDs: make(map[string]struct{}),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(a)
		}
	}
	return a
}

// OnStartup registers a startup hook.
func (a *App) OnStartup(h StartupHook) *App {
	if h != nil {
		a.startupHooks = append(a.startupHooks, h)
	}
	return a
}

// OnShutdown registers a shutdown hook.
func (a *App) OnShutdown(h ShutdownHook) *App {
	if h != nil {
		a.shutdownHooks = append(a.shutdownHooks, h)
	}
	return a
}

// Use registers one or more modules on the app instance.
func (a *App) Use(mods ...Module) *App {
	for _, m := range mods {
		if m != nil {
			a.modules = append(a.modules, m)
		}
	}
	return a
}
