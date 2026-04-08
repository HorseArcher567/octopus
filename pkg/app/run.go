package app

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
)

type runOptions struct {
	ctx           context.Context
	signals       []os.Signal
	startupHooks  []StartupHook
	shutdownHooks []ShutdownHook
}

// RunOption customizes app.Run behavior.
type RunOption func(*runOptions)

// WithSignals overrides default process signals (SIGTERM/SIGINT).
func WithSignals(sig ...os.Signal) RunOption {
	return func(o *runOptions) {
		o.signals = append([]os.Signal(nil), sig...)
	}
}

// WithContext runs the app with a specific context instead of signal context.
func WithContext(ctx context.Context) RunOption {
	return func(o *runOptions) {
		o.ctx = ctx
	}
}

// WithStartupHook appends a startup hook.
func WithStartupHook(h StartupHook) RunOption {
	return func(o *runOptions) {
		if h != nil {
			o.startupHooks = append(o.startupHooks, h)
		}
	}
}

// WithShutdownHook appends a shutdown hook.
func WithShutdownHook(h ShutdownHook) RunOption {
	return func(o *runOptions) {
		if h != nil {
			o.shutdownHooks = append(o.shutdownHooks, h)
		}
	}
}

// Run creates an app instance and runs it with module lifecycle support.
func Run(configPath string, mods []Module, opts ...RunOption) error {
	o := runOptions{
		signals: []os.Signal{syscall.SIGTERM, syscall.SIGINT},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}

	if configPath == "" {
		return errors.New("app: config path is required")
	}
	a, err := Load(configPath)
	if err != nil {
		return err
	}
	a.Use(mods...)
	for _, h := range o.startupHooks {
		a.OnStartup(h)
	}
	for _, h := range o.shutdownHooks {
		a.OnShutdown(h)
	}

	if o.ctx != nil {
		return a.Run(o.ctx)
	}
	return runWithSignals(a, o.signals)
}

// MustRun panics when Run returns error.
func MustRun(configPath string, mods []Module, opts ...RunOption) {
	if err := Run(configPath, mods, opts...); err != nil {
		panic(err)
	}
}

func runWithSignals(a *App, sig []os.Signal) error {
	ctx, stop := signal.NotifyContext(context.Background(), sig...)
	defer stop()
	return a.Run(ctx)
}
