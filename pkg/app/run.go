package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/HorseArcher567/octopus/pkg/hook"
	"golang.org/x/sync/errgroup"
)

// Run starts all configured services and blocks until context cancellation
// or service error. Run may be called at most once per App instance.
func (a *App) Run(ctx context.Context) (retErr error) {
	if err := a.markRunOnce(); err != nil {
		return err
	}
	if ctx == nil {
		ctx = context.Background()
	}

	a.log.Info("starting application")
	defer func() {
		a.log.Info("shutting down application")
		retErr = errors.Join(retErr, a.shutdown())
	}()

	if err := a.runStartupHooks(ctx); err != nil {
		return err
	}

	g, groupCtx := errgroup.WithContext(ctx)
	for _, svc := range a.services {
		svc := svc
		a.log.Info("starting service", "service", svc.Name())
		g.Go(func() error {
			if err := svc.Run(groupCtx); err != nil {
				a.log.Error("service exited with error", "service", svc.Name(), "error", err)
				return fmt.Errorf("service %q: %w", svc.Name(), err)
			}
			a.log.Info("service exited", "service", svc.Name())
			return nil
		})
	}

	waitErr := g.Wait()
	if waitErr == nil {
		return nil
	}
	if ctx.Err() != nil && (errors.Is(waitErr, context.Canceled) || errors.Is(waitErr, context.DeadlineExceeded)) {
		return nil
	}
	return waitErr
}

func (a *App) runStartupHooks(ctx context.Context) error {
	hookCtx := hook.NewContext(ctx, a.log, a.store)
	for i, h := range a.startupHooks {
		a.log.Info("running startup hook", "hook", i)
		if err := h(hookCtx); err != nil {
			a.log.Error("startup hook failed", "hook", i, "error", err)
			return fmt.Errorf("startup hook %d: %w", i, err)
		}
	}
	return nil
}

func (a *App) runShutdownHooks(ctx context.Context) error {
	hookCtx := hook.NewContext(ctx, a.log, a.store)
	var errs []error
	for i := len(a.shutdownHooks) - 1; i >= 0; i-- {
		a.log.Info("running shutdown hook", "hook", i)
		if err := a.shutdownHooks[i](hookCtx); err != nil {
			a.log.Error("shutdown hook failed", "hook", i, "error", err)
			errs = append(errs, fmt.Errorf("shutdown hook %d: %w", i, err))
		}
	}
	return errors.Join(errs...)
}

func (a *App) shutdown() error {
	var shutdownErr error
	a.shutdownOnce.Do(func() {
		timeout := a.shutdownTimeout
		if timeout == 0 {
			timeout = defaultShutdownTimeout
		}
		shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		var errs []error
		err := a.stopServices(shutdownCtx)
		if err == nil {
			a.log.Info("all services stopped")
		}
		if err != nil {
			errs = append(errs, err)
		}
		err = a.runShutdownHooks(shutdownCtx)
		if err != nil {
			errs = append(errs, err)
		}
		if a.store != nil {
			a.log.Info("closing store")
			if err := a.store.Close(); err != nil {
				a.log.Error("close store failed", "error", err)
				errs = append(errs, fmt.Errorf("close store: %w", err))
			}
		}
		shutdownErr = errors.Join(errs...)
	})
	return shutdownErr
}

func (a *App) stopServices(ctx context.Context) error {
	var errs []error
	for i := len(a.services) - 1; i >= 0; i-- {
		svc := a.services[i]
		a.log.Info("stopping service", "service", svc.Name())
		if err := svc.Stop(ctx); err != nil {
			a.log.Error("stop service failed", "service", svc.Name(), "error", err)
			errs = append(errs, fmt.Errorf("stop service %q: %w", svc.Name(), err))
		}
	}
	return errors.Join(errs...)
}
