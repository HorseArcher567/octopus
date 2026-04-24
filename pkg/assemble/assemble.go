package assemble

import (
	"github.com/HorseArcher567/octopus/pkg/app"
	"github.com/HorseArcher567/octopus/pkg/config"
	"github.com/HorseArcher567/octopus/pkg/hook"
)

// Domain registers one business domain into the application construction process.
type Domain func(*DomainContext) error

// SetupStep contributes custom setup work that runs after builtin setup and
// before business domains. Setup steps are intended for preparing shared
// infrastructure resources that later domains will consume.
type SetupStep struct {
	Name string
	Run  func(*SetupContext) error
}

type options struct {
	domains       []Domain
	setupSteps    []SetupStep
	startupHooks  []hook.Func
	shutdownHooks []hook.Func
}

// Option customizes facade assembly behavior.
type Option func(*options)

// WithDomains registers one or more business domains.
func WithDomains(domains ...Domain) Option {
	return func(o *options) {
		o.domains = append(o.domains, domains...)
	}
}

// WithSetup registers one or more custom setup steps.
func WithSetup(steps ...SetupStep) Option {
	return func(o *options) {
		o.setupSteps = append(o.setupSteps, steps...)
	}
}

// WithStartupHooks registers one or more app-level startup hooks.
func WithStartupHooks(hooks ...hook.Func) Option {
	return func(o *options) {
		o.startupHooks = append(o.startupHooks, hooks...)
	}
}

// WithShutdownHooks registers one or more app-level shutdown hooks.
func WithShutdownHooks(hooks ...hook.Func) Option {
	return func(o *options) {
		o.shutdownHooks = append(o.shutdownHooks, hooks...)
	}
}

// Load loads config from path, performs internal setup and domain registration,
// and returns a ready-to-run app.
func Load(path string, opts ...Option) (*app.App, error) {
	cfg, err := config.Load(path)
	if err != nil {
		return nil, err
	}
	return New(cfg, opts...)
}

// New creates an app from an already-loaded config, performs internal
// setup and domain registration, and returns a ready-to-run app.
func New(cfg *config.Config, opts ...Option) (_ *app.App, retErr error) {
	state, err := setup(cfg)
	if err != nil {
		return nil, err
	}
	defer func() {
		if retErr != nil && state != nil && state.store != nil {
			_ = state.store.Close()
		}
	}()
	return build(cfg, state, opts...)
}
