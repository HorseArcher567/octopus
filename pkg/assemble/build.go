package assemble

import (
	"fmt"
	"strings"

	"github.com/HorseArcher567/octopus/pkg/app"
	"github.com/HorseArcher567/octopus/pkg/config"
)

func build(raw *config.Config, s *state, opts ...Option) (*app.App, error) {
	o := buildOptions(opts...)

	setupCtx, err := newSetupContext(raw, s)
	if err != nil {
		return nil, err
	}
	if err := applySetupSteps(setupCtx, o.setupSteps); err != nil {
		return nil, err
	}

	ctx, err := newContext(s)
	if err != nil {
		return nil, err
	}
	ctx.startupHooks = append(ctx.startupHooks, o.startupHooks...)
	ctx.shutdownHooks = append(ctx.shutdownHooks, o.shutdownHooks...)
	if err := applyDomains(ctx, o.domains); err != nil {
		return nil, err
	}

	appCfg, err := loadAppConfig(raw)
	if err != nil {
		return nil, err
	}
	return assembleApp(s, ctx, appCfg), nil
}

func loadAppConfig(raw *config.Config) (*app.Config, error) {
	if raw == nil {
		return nil, fmt.Errorf("assemble: config cannot be nil")
	}
	var cfg app.Config
	if _, ok := raw.Get("app"); !ok {
		return &cfg, nil
	}
	if err := raw.UnmarshalKey("app", &cfg); err != nil {
		return nil, fmt.Errorf("assemble: decode config \"app\": %w", err)
	}
	return &cfg, nil
}

func buildOptions(opts ...Option) options {
	var o options
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	return o
}

func applySetupSteps(ctx *SetupContext, steps []SetupStep) error {
	seen := make(map[string]struct{}, len(steps))
	for _, step := range steps {
		name := strings.TrimSpace(step.Name)
		if name == "" {
			return fmt.Errorf("assemble: setup step name is required")
		}
		if _, ok := seen[name]; ok {
			return fmt.Errorf("assemble: duplicate setup step %q", name)
		}
		seen[name] = struct{}{}
		if step.Run == nil {
			return fmt.Errorf("assemble: setup step %q run is required", name)
		}
		if err := step.Run(ctx); err != nil {
			return fmt.Errorf("assemble: setup step %s: %w", name, err)
		}
	}
	return nil
}

func applyDomains(ctx *DomainContext, domains []Domain) error {
	for i, domain := range domains {
		if domain == nil {
			continue
		}
		if err := domain(ctx); err != nil {
			return fmt.Errorf("assemble: domain %d: %w", i, err)
		}
	}
	return nil
}

func assembleApp(s *state, ctx *DomainContext, appCfg *app.Config) *app.App {
	appOpts := make([]app.Option, 0, 2)
	appOpts = append(appOpts, app.WithStore(s.store))
	if appCfg != nil && appCfg.ShutdownTimeout > 0 {
		appOpts = append(appOpts, app.WithShutdownTimeout(appCfg.ShutdownTimeout))
	}

	a := app.New(s.log, appOpts...)
	a.AddServices(ctx.services...)
	for _, h := range ctx.startupHooks {
		a.OnStartup(h)
	}
	for _, h := range ctx.shutdownHooks {
		a.OnShutdown(h)
	}
	// Register builtin services after custom services so builtin services stop
	// first during pkg/app's reverse-order shutdown.
	for _, svc := range builtinServices(s) {
		a.AddServices(svc)
	}
	return a
}
