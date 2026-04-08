package app

import (
	"errors"

	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/database"
	"github.com/HorseArcher567/octopus/pkg/job"
	redisclient "github.com/HorseArcher567/octopus/pkg/redis"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"google.golang.org/grpc"
)

var (
	_ BuildContext  = (*buildContext)(nil)
	_ RPCRegistrar  = (*rpcRegistrar)(nil)
	_ HTTPRegistrar = (*httpRegistrar)(nil)
	_ JobRegistrar  = (*jobRegistrar)(nil)
)

func (a *App) Logger() *xlog.Logger {
	return a.log
}

func (a *App) MySQL(name string) (*database.DB, error) {
	if a.resources == nil {
		return nil, errors.New("app: resource runtime is not initialized")
	}
	return a.resources.MySQL(name)
}

func (a *App) Redis(name string) (*redisclient.Client, error) {
	if a.resources == nil {
		return nil, errors.New("app: resource runtime is not initialized")
	}
	return a.resources.Redis(name)
}

func (a *App) RPCClient(target string) (*grpc.ClientConn, error) {
	if a.rpc == nil {
		return nil, errors.New("app: rpc runtime is not initialized")
	}
	return a.rpc.Client(target)
}

// NewRPCClient keeps the old public name for direct callers.
func (a *App) NewRPCClient(target string) (*grpc.ClientConn, error) {
	return a.RPCClient(target)
}

func (a *App) CloseRpcClients() {
	if a.rpc == nil {
		return
	}
	if err := a.rpc.CloseClients(); err != nil {
		a.log.Error("failed to close rpc clients", "error", err)
	}
}

type buildContext struct {
	a *App
}

func newBuildContext(a *App) BuildContext {
	return &buildContext{a: a}
}

func (c *buildContext) Logger() *xlog.Logger {
	return c.a.Logger()
}

func (c *buildContext) MySQL(name string) (*database.DB, error) {
	return c.a.MySQL(name)
}

func (c *buildContext) Redis(name string) (*redisclient.Client, error) {
	return c.a.Redis(name)
}

func (c *buildContext) RPCClient(target string) (*grpc.ClientConn, error) {
	return c.a.RPCClient(target)
}

func (c *buildContext) Provide(value any) error {
	return c.a.container.Provide(value)
}

func (c *buildContext) Resolve(target any) error {
	return c.a.container.Resolve(target)
}

func (c *buildContext) MustResolve(target any) {
	c.a.container.MustResolve(target)
}

type rpcRegistrar struct {
	a *App
}

func newRPCRegistrar(a *App) RPCRegistrar {
	return &rpcRegistrar{a: a}
}

func (r *rpcRegistrar) Logger() *xlog.Logger {
	return r.a.Logger()
}

func (r *rpcRegistrar) Resolve(target any) error {
	return r.a.container.Resolve(target)
}

func (r *rpcRegistrar) MustResolve(target any) {
	r.a.container.MustResolve(target)
}

func (r *rpcRegistrar) RegisterRPC(register func(s *grpc.Server)) error {
	if r.a.rpc == nil {
		return errors.New("app: rpc runtime is not initialized")
	}
	return r.a.rpc.Register(register)
}

type httpRegistrar struct {
	a *App
}

func newHTTPRegistrar(a *App) HTTPRegistrar {
	return &httpRegistrar{a: a}
}

func (r *httpRegistrar) Logger() *xlog.Logger {
	return r.a.Logger()
}

func (r *httpRegistrar) Resolve(target any) error {
	return r.a.container.Resolve(target)
}

func (r *httpRegistrar) MustResolve(target any) {
	r.a.container.MustResolve(target)
}

func (r *httpRegistrar) RegisterHTTP(register func(engine *api.Engine)) error {
	if r.a.http == nil {
		return errors.New("app: http runtime is not initialized")
	}
	return r.a.http.Register(register)
}

type jobRegistrar struct {
	a *App
}

func newJobRegistrar(a *App) JobRegistrar {
	return &jobRegistrar{a: a}
}

func (r *jobRegistrar) Logger() *xlog.Logger {
	return r.a.Logger()
}

func (r *jobRegistrar) Resolve(target any) error {
	return r.a.container.Resolve(target)
}

func (r *jobRegistrar) MustResolve(target any) {
	r.a.container.MustResolve(target)
}

func (r *jobRegistrar) AddJob(name string, fn job.Func) error {
	if r.a.jobs == nil {
		return errors.New("app: job runtime is not initialized")
	}
	return r.a.jobs.Add(name, fn)
}
