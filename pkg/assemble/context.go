package assemble

import (
	"context"
	"fmt"
	"strings"

	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/app"
	"github.com/HorseArcher567/octopus/pkg/config"
	"github.com/HorseArcher567/octopus/pkg/job"
	"github.com/HorseArcher567/octopus/pkg/store"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"google.golang.org/grpc"
)

type apiServer interface {
	Register(func(*api.Engine)) error
	Run(ctx context.Context) error
	Stop(ctx context.Context) error
}

type rpcServer interface {
	Register(func(*grpc.Server)) error
	Run(ctx context.Context) error
	Stop(ctx context.Context) error
}

type jobScheduler interface {
	Add(name string, fn job.Func) error
	Run(ctx context.Context) error
	Stop(ctx context.Context) error
}

// SetupContext exposes the narrow setup-time capability surface available to
// custom setup steps.
type SetupContext struct {
	inner *setupContext
}

type Context struct {
	state *state

	startupHooks  []app.StartupHook
	shutdownHooks []app.ShutdownHook
	services      []app.Service
}

func newSetupContext(cfg *config.Config, s *state) (*SetupContext, error) {
	if cfg == nil {
		return nil, fmt.Errorf("assemble: config cannot be nil")
	}
	if s == nil {
		return nil, fmt.Errorf("assemble: state cannot be nil")
	}
	if s.store == nil {
		return nil, fmt.Errorf("assemble: state.store cannot be nil")
	}
	if s.log == nil {
		return nil, fmt.Errorf("assemble: state.log cannot be nil")
	}
	return &SetupContext{inner: &setupContext{cfg: cfg, state: s}}, nil
}

func newContext(s *state) (*Context, error) {
	if s == nil {
		return nil, fmt.Errorf("assemble: state cannot be nil")
	}
	if s.store == nil {
		return nil, fmt.Errorf("assemble: state.store cannot be nil")
	}
	if s.log == nil {
		return nil, fmt.Errorf("assemble: state.log cannot be nil")
	}
	if s.job == nil {
		return nil, fmt.Errorf("assemble: state.job cannot be nil")
	}
	return &Context{state: s}, nil
}

// DecodeSetupConfig decodes a config subtree for a custom setup step.
func DecodeSetupConfig[T any](c *SetupContext, key string) (*T, error) {
	var v T
	if c == nil || c.inner == nil {
		return nil, fmt.Errorf("setup context cannot be nil")
	}
	if err := c.inner.decodeStruct(key, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// Logger returns the app logger selected during builtin setup.
func (c *SetupContext) Logger() *xlog.Logger { return c.inner.state.log }

// NamedLogger returns a configured named logger from the shared store.
func (c *SetupContext) NamedLogger(name string) (*xlog.Logger, error) {
	selected := strings.TrimSpace(name)
	if selected == "" {
		return nil, fmt.Errorf("logger name is required")
	}
	return lookupLogger(selected, c.inner.state.store)
}

// Provide registers a shared infrastructure resource into the store.
func (c *SetupContext) Provide(name string, value any, opts ...store.SetOption) error {
	return c.inner.provide(name, value, opts...)
}

func (c *Context) Logger() *xlog.Logger { return c.state.log }

func (c *Context) Store() store.Store { return c.state.store }

func (c *Context) RegisterAPI(fn func(*api.Engine)) error {
	if c.state.api == nil {
		return ErrAPINotConfigured
	}
	return c.state.api.Register(fn)
}

func (c *Context) RegisterRPC(fn func(*grpc.Server)) error {
	if c.state.rpc == nil {
		return ErrRPCNotConfigured
	}
	return c.state.rpc.Register(fn)
}

func (c *Context) RegisterJob(name string, fn job.Func) error {
	return c.state.job.Add(name, fn)
}

func (c *Context) OnStartup(h app.StartupHook) {
	if h != nil {
		c.startupHooks = append(c.startupHooks, h)
	}
}

func (c *Context) OnShutdown(h app.ShutdownHook) {
	if h != nil {
		c.shutdownHooks = append(c.shutdownHooks, h)
	}
}

func (c *Context) AddService(s app.Service) {
	if s != nil {
		c.services = append(c.services, s)
	}
}
