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
	if err := a.markRunOnce(); err != nil {
		return err
	}
	if ctx == nil {
		ctx = context.Background()
	}

	a.log.Info("starting application")
	cancelRun := func() {}
	defer func() {
		cancelRun()
		retErr = errors.Join(retErr, a.shutdown())
	}()

	if err := a.prepareModules(); err != nil {
		return err
	}
	if err := a.buildModules(ctx); err != nil {
		return err
	}
	if err := a.registerRPCModules(ctx); err != nil {
		return err
	}
	if err := a.registerHTTPModules(ctx); err != nil {
		return err
	}
	if err := a.registerJobsModules(ctx); err != nil {
		return err
	}
	if err := a.execStartupHooks(ctx); err != nil {
		return err
	}

	runCtx, cancel := context.WithCancel(ctx)
	cancelRun = cancel
	g, groupCtx := errgroup.WithContext(runCtx)
	a.runBuiltins(g, groupCtx)
	a.runModules(g, groupCtx)

	waitErr := g.Wait()
	if errors.Is(waitErr, context.Canceled) || errors.Is(waitErr, context.DeadlineExceeded) {
		return nil
	}
	return waitErr
}

func (a *App) prepareModules() error {
	order, err := resolveModuleOrder(a.modules)
	if err != nil {
		return err
	}
	a.orderedModules = order
	return nil
}

func (a *App) markRunOnce() error {
	a.runMu.Lock()
	defer a.runMu.Unlock()
	if a.hasRun {
		return errors.New("app: Run can only be called once")
	}
	a.hasRun = true
	return nil
}

func (a *App) buildModules(ctx context.Context) error {
	bctx := newBuildContext(a)
	for _, mod := range a.orderedModules {
		buildMod, ok := mod.(BuildModule)
		if !ok {
			continue
		}
		if err := buildMod.Build(ctx, bctx); err != nil {
			return fmt.Errorf("build module %q: %w", mod.ID(), err)
		}
		a.activateCloser(mod)
	}
	return nil
}

func (a *App) registerRPCModules(ctx context.Context) error {
	registrar := newRPCRegistrar(a)
	for _, mod := range a.orderedModules {
		rpcMod, ok := mod.(RegisterRPCModule)
		if !ok {
			continue
		}
		if err := rpcMod.RegisterRPC(ctx, registrar); err != nil {
			return fmt.Errorf("register rpc module %q: %w", mod.ID(), err)
		}
		a.activateCloser(mod)
	}
	return nil
}

func (a *App) registerHTTPModules(ctx context.Context) error {
	registrar := newHTTPRegistrar(a)
	for _, mod := range a.orderedModules {
		httpMod, ok := mod.(RegisterHTTPModule)
		if !ok {
			continue
		}
		if err := httpMod.RegisterHTTP(ctx, registrar); err != nil {
			return fmt.Errorf("register http module %q: %w", mod.ID(), err)
		}
		a.activateCloser(mod)
	}
	return nil
}

func (a *App) registerJobsModules(ctx context.Context) error {
	registrar := newJobRegistrar(a)
	for _, mod := range a.orderedModules {
		jobMod, ok := mod.(RegisterJobsModule)
		if !ok {
			continue
		}
		if err := jobMod.RegisterJobs(ctx, registrar); err != nil {
			return fmt.Errorf("register jobs module %q: %w", mod.ID(), err)
		}
		a.activateCloser(mod)
	}
	return nil
}

func (a *App) runBuiltins(g *errgroup.Group, ctx context.Context) {
	if a.rpc != nil {
		g.Go(func() error { return a.rpc.Run(ctx) })
	}
	if a.http != nil {
		g.Go(func() error { return a.http.Run(ctx) })
	}
	if a.jobs != nil {
		g.Go(func() error { return a.jobs.Run(ctx) })
	}
}

func (a *App) runModules(g *errgroup.Group, ctx context.Context) {
	for _, mod := range a.orderedModules {
		runMod, ok := mod.(RunModule)
		if !ok {
			continue
		}
		a.activateCloser(mod)
		runFn := runMod
		id := mod.ID()
		g.Go(func() error {
			if err := runFn.Run(ctx); err != nil {
				return fmt.Errorf("run module %q: %w", id, err)
			}
			return nil
		})
	}
}

func (a *App) activateCloser(mod Module) {
	closeMod, ok := mod.(CloseModule)
	if !ok {
		return
	}
	id := mod.ID()
	if _, exists := a.activeCloserIDs[id]; exists {
		return
	}
	a.activeCloserIDs[id] = struct{}{}
	a.activeClosers = append(a.activeClosers, moduleCloser{id: id, fn: closeMod})
}

func (a *App) closeActiveClosers(ctx context.Context) error {
	var errs []error
	for i := len(a.activeClosers) - 1; i >= 0; i-- {
		mod := a.activeClosers[i]
		if err := mod.fn.Close(ctx); err != nil {
			a.log.Error("module close error", "module_id", mod.id, "error", err)
			errs = append(errs, fmt.Errorf("module %q close: %w", mod.id, err))
		}
	}
	a.activeClosers = nil
	a.activeCloserIDs = make(map[string]struct{})
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
		errs = append(errs, a.stopBuiltins(shutdownCtx)...)
		if err := a.closeActiveClosers(shutdownCtx); err != nil {
			errs = append(errs, err)
		}
		if err := a.execShutdownHooks(shutdownCtx); err != nil {
			errs = append(errs, fmt.Errorf("shutdown hooks: %w", err))
		}
		if a.resources != nil {
			if err := a.resources.Close(); err != nil {
				errs = append(errs, fmt.Errorf("close resources: %w", err))
			}
		}
		if a.rpc != nil {
			if err := a.rpc.Close(); err != nil {
				errs = append(errs, fmt.Errorf("close rpc runtime: %w", err))
			}
		}
		if err := a.log.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close logger: %w", err))
		}
		shutdownErr = errors.Join(errs...)
	})
	return shutdownErr
}

func (a *App) stopBuiltins(ctx context.Context) []error {
	var errs []error
	if a.http != nil {
		if err := a.http.Stop(ctx); err != nil {
			errs = append(errs, fmt.Errorf("stop http runtime: %w", err))
		}
	}
	if a.rpc != nil {
		if err := a.rpc.Stop(ctx); err != nil {
			errs = append(errs, fmt.Errorf("stop rpc runtime: %w", err))
		}
	}
	if a.jobs != nil {
		if err := a.jobs.Stop(ctx); err != nil {
			errs = append(errs, fmt.Errorf("stop job runtime: %w", err))
		}
	}
	return errs
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
		dm, ok := mod.(DependentModule)
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
