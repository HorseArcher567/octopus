package assemble

import (
	"fmt"

	"github.com/HorseArcher567/octopus/pkg/app"
)

func build(s *state, opts ...Option) (*app.App, error) {
	o := buildOptions(opts...)

	ctx, err := newContext(s)
	if err != nil {
		return nil, err
	}
	if err := applyActions(ctx, o.actions); err != nil {
		return nil, err
	}

	return assembleApp(s, ctx), nil
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

func applyActions(ctx *Context, actions []Action) error {
	for i, action := range actions {
		if action == nil {
			continue
		}
		if err := action(ctx); err != nil {
			return fmt.Errorf("assemble: action %d: %w", i, err)
		}
	}
	return nil
}

func assembleApp(s *state, ctx *Context) *app.App {
	appOpts := make([]app.Option, 0, 2)
	appOpts = append(appOpts, app.WithStore(s.store))
	if s.cfg.App.ShutdownTimeout > 0 {
		appOpts = append(appOpts, app.WithShutdownTimeout(s.cfg.App.ShutdownTimeout))
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
