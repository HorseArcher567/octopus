package app

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/HorseArcher567/octopus/pkg/hook"
	"github.com/HorseArcher567/octopus/pkg/store"
	"github.com/HorseArcher567/octopus/pkg/xlog"
)

const defaultShutdownTimeout = 30 * time.Second

// Service is a long-running runtime unit managed by App.
type Service interface {
	Name() string
	Run(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Config defines application runtime policy loaded from the app config section.
type Config struct {
	Logger          string        `yaml:"logger" json:"logger" toml:"logger"`
	ShutdownTimeout time.Duration `yaml:"shutdownTimeout" json:"shutdownTimeout" toml:"shutdownTimeout"`
}

// Option customizes an App instance.
type Option func(*App)

// WithStore injects the shared dependency store owned by the app.
func WithStore(s store.Store) Option {
	return func(a *App) {
		a.store = s
	}
}

// WithShutdownTimeout overrides the default shutdown timeout.
func WithShutdownTimeout(timeout time.Duration) Option {
	return func(a *App) {
		a.shutdownTimeout = timeout
	}
}

// App is the minimal lifecycle kernel of Octopus.
type App struct {
	log   *xlog.Logger
	store store.Store

	services      []Service
	startupHooks  []hook.Func
	shutdownHooks []hook.Func

	runMu  sync.Mutex
	hasRun bool

	shutdownOnce    sync.Once
	shutdownTimeout time.Duration
}

// New creates a new App from already assembled runtime inputs.
func New(log *xlog.Logger, opts ...Option) *App {
	if log == nil {
		log = xlog.MustNew(nil)
	}
	a := &App{log: log}
	for _, opt := range opts {
		if opt != nil {
			opt(a)
		}
	}
	return a
}

// Logger returns the application logger.
func (a *App) Logger() *xlog.Logger {
	return a.log
}

// AddServices appends runtime services to the app.
func (a *App) AddServices(services ...Service) *App {
	for _, svc := range services {
		if svc != nil {
			a.services = append(a.services, svc)
		}
	}
	return a
}

// OnStartup registers a startup hook.
func (a *App) OnStartup(h hook.Func) *App {
	if h != nil {
		a.startupHooks = append(a.startupHooks, h)
	}
	return a
}

// OnShutdown registers a shutdown hook.
func (a *App) OnShutdown(h hook.Func) *App {
	if h != nil {
		a.shutdownHooks = append(a.shutdownHooks, h)
	}
	return a
}

// markRunOnce ensures Run is only executed once per App instance.
func (a *App) markRunOnce() error {
	a.runMu.Lock()
	defer a.runMu.Unlock()
	if a.hasRun {
		return errors.New("app: Run can only be called once")
	}
	a.hasRun = true
	return nil
}
