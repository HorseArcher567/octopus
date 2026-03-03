package app

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

const defaultShutdownTimeout = 30 * time.Second

// Run starts all configured components and blocks until context cancellation or component error.
// Run can only be called once for an App instance.
func (a *App) Run(ctx context.Context) (retErr error) {
	a.runMu.Lock()
	if a.hasRun {
		a.runMu.Unlock()
		return errors.New("app: Run can only be called once")
	}
	a.hasRun = true
	a.runMu.Unlock()

	if ctx == nil {
		ctx = context.Background()
	}

	a.log.Info("starting application")
	defer func() {
		retErr = errors.Join(retErr, a.shutdown())
	}()

	if err := a.prepareModules(); err != nil {
		return err
	}
	if err := a.initModules(ctx); err != nil {
		return err
	}
	if err := a.execStartupHooks(ctx); err != nil {
		return err
	}

	g, runCtx := errgroup.WithContext(ctx)
	if a.rpcServer != nil {
		g.Go(func() error { return a.rpcServer.Run(runCtx) })
	}
	if a.apiServer != nil {
		g.Go(func() error { return a.apiServer.Run(runCtx) })
	}
	if a.jobScheduler != nil {
		g.Go(func() error { return a.jobScheduler.Run(runCtx) })
	}

	retErr = g.Wait()
	if errors.Is(retErr, context.Canceled) || errors.Is(retErr, context.DeadlineExceeded) {
		return nil
	}
	return retErr
}

func (a *App) prepareModules() error {
	order, err := resolveModuleOrder(a.modules)
	if err != nil {
		return err
	}
	a.orderedModules = order
	return nil
}

func (a *App) initModules(ctx context.Context) error {
	for _, mod := range a.orderedModules {
		if err := mod.Init(ctx, a); err != nil {
			rollbackErr := a.rollbackModules(ctx)
			return errors.Join(fmt.Errorf("module %q init: %w", mod.ID(), err), rollbackErr)
		}
		a.initializedModules = append(a.initializedModules, mod)
	}
	return nil
}

func (a *App) rollbackModules(ctx context.Context) error {
	if len(a.initializedModules) == 0 {
		return nil
	}
	err := a.closeModules(ctx, a.initializedModules)
	a.initializedModules = nil
	return err
}

func (a *App) closeInitializedModules(ctx context.Context) error {
	err := a.closeModules(ctx, a.initializedModules)
	a.initializedModules = nil
	return err
}

func (a *App) closeModules(ctx context.Context, mods []Module) error {
	var errs []error
	for i := len(mods) - 1; i >= 0; i-- {
		mod := mods[i]
		if err := mod.Close(ctx); err != nil {
			a.log.Error("module close error", "module_id", mod.ID(), "error", err)
			errs = append(errs, fmt.Errorf("module %q close: %w", mod.ID(), err))
		}
	}
	return errors.Join(errs...)
}

func (a *App) execStartupHooks(ctx context.Context) error {
	for _, h := range a.startupHooks {
		if err := h(ctx, a); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) execShutdownHooks(ctx context.Context) error {
	var errs []error
	for _, h := range a.shutdownHooks {
		if err := h(ctx, a); err != nil {
			errs = append(errs, err)
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
		if a.apiServer != nil {
			if err := a.apiServer.Stop(shutdownCtx); err != nil {
				errs = append(errs, fmt.Errorf("stop api server: %w", err))
			}
		}
		if a.rpcServer != nil {
			if err := a.rpcServer.Stop(shutdownCtx); err != nil {
				errs = append(errs, fmt.Errorf("stop rpc server: %w", err))
			}
		}
		if a.jobScheduler != nil {
			if err := a.jobScheduler.Stop(shutdownCtx); err != nil {
				errs = append(errs, fmt.Errorf("stop job scheduler: %w", err))
			}
		}

		if err := a.execShutdownHooks(shutdownCtx); err != nil {
			errs = append(errs, fmt.Errorf("shutdown hooks: %w", err))
		}
		if err := a.closeInitializedModules(shutdownCtx); err != nil {
			errs = append(errs, err)
		}
		if a.resources != nil {
			if err := a.resources.Close(); err != nil {
				errs = append(errs, fmt.Errorf("close resources: %w", err))
			}
		}
		a.CloseRpcClients()
		if err := a.log.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close logger: %w", err))
		}
		shutdownErr = errors.Join(errs...)
	})
	return shutdownErr
}

func resolveModuleOrder(mods []Module) ([]Module, error) {
	if len(mods) == 0 {
		return nil, nil
	}

	byID := make(map[string]Module, len(mods))
	orderIdx := make(map[string]int, len(mods))
	for idx, mod := range mods {
		id := strings.TrimSpace(mod.ID())
		if id == "" {
			return nil, errors.New("app: module id cannot be empty")
		}
		if _, exists := byID[id]; exists {
			return nil, fmt.Errorf("app: duplicate module id %q", id)
		}
		byID[id] = mod
		orderIdx[id] = idx
	}

	indegree := make(map[string]int, len(mods))
	outs := make(map[string][]string, len(mods))
	for id := range byID {
		indegree[id] = 0
	}

	for _, mod := range mods {
		dm, ok := mod.(DependedModule)
		if !ok {
			continue
		}
		from := mod.ID()
		for _, depID := range dm.DependsOn() {
			depID = strings.TrimSpace(depID)
			if depID == "" {
				continue
			}
			if _, exists := byID[depID]; !exists {
				return nil, fmt.Errorf("app: module %q depends on unknown module %q", from, depID)
			}
			outs[depID] = append(outs[depID], from)
			indegree[from]++
		}
	}

	ready := make([]string, 0)
	for id, d := range indegree {
		if d == 0 {
			ready = append(ready, id)
		}
	}
	sort.Slice(ready, func(i, j int) bool { return orderIdx[ready[i]] < orderIdx[ready[j]] })

	result := make([]Module, 0, len(mods))
	for len(ready) > 0 {
		id := ready[0]
		ready = ready[1:]
		result = append(result, byID[id])

		for _, to := range outs[id] {
			indegree[to]--
			if indegree[to] == 0 {
				ready = append(ready, to)
			}
		}
		sort.Slice(ready, func(i, j int) bool { return orderIdx[ready[i]] < orderIdx[ready[j]] })
	}

	if len(result) != len(mods) {
		remaining := make([]string, 0)
		for id, d := range indegree {
			if d > 0 {
				remaining = append(remaining, id)
			}
		}
		sort.Strings(remaining)
		return nil, fmt.Errorf("app: module dependency cycle detected: %s", strings.Join(remaining, ", "))
	}

	return result, nil
}
