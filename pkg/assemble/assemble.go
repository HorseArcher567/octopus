package assemble

import (
	"github.com/HorseArcher567/octopus/pkg/app"
	"github.com/HorseArcher567/octopus/pkg/config"
)

// Action contributes business assembly to the application build process.
type Action func(*Context) error

// SetupStep contributes custom setup work that runs after builtin setup and
// before business actions. Setup steps are intended for preparing shared
// infrastructure resources that later actions will consume.
type SetupStep struct {
	Name string
	Run  func(*SetupContext) error
}

type options struct {
	actions    []Action
	setupSteps []SetupStep
}

// Option customizes facade assembly behavior.
type Option func(*options)

// With registers one or more assembly actions.
func With(actions ...Action) Option {
	return func(o *options) {
		o.actions = append(o.actions, actions...)
	}
}

// WithSetupSteps registers one or more custom setup steps.
func WithSetupSteps(steps ...SetupStep) Option {
	return func(o *options) {
		o.setupSteps = append(o.setupSteps, steps...)
	}
}

// Load loads config from path, performs internal setup and assembly,
// and returns a ready-to-run app.
func Load(path string, opts ...Option) (*app.App, error) {
	cfg, err := config.Load(path)
	if err != nil {
		return nil, err
	}
	return New(cfg, opts...)
}

// New builds an app from an already-loaded config, performs internal
// setup and assembly, and returns a ready-to-run app.
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
