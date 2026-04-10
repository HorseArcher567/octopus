package app

// This file defines the runtime abstractions consumed by App and the adapters
// that expose those capabilities during module build and registration phases.

import (
	"context"
	"errors"

	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/job"
	"github.com/HorseArcher567/octopus/pkg/telemetry"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"google.golang.org/grpc"
)

// RPCRuntime owns inbound RPC registration and outbound client creation.
type RPCRuntime interface {
	Register(func(*grpc.Server)) error
	Client(target string) (*grpc.ClientConn, error)
	CloseClients() error
	Run(context.Context) error
	Stop(context.Context) error
	Close() error
}

// APIRuntime owns API route registration and server lifecycle.
type APIRuntime interface {
	Register(func(*api.Engine)) error
	Run(context.Context) error
	Stop(context.Context) error
}

// JobRuntime owns background job registration and scheduler lifecycle.
type JobRuntime interface {
	Add(name string, fn job.Func) error
	Run(context.Context) error
	Stop(context.Context) error
}

// ResourceRuntime owns shared infrastructure resources.
type ResourceRuntime interface {
	Register(kind, name string, value any, closeFn func() error) error
	Get(kind, name string) (any, error)
	MustGet(kind, name string) any
	List(kind string) []string
	Close() error
}

var (
	_ BuildContext = (*buildContext)(nil)
	_ RPCRegistrar = (*rpcRegistrar)(nil)
	_ APIRegistrar = (*apiRegistrar)(nil)
	_ JobRegistrar = (*jobRegistrar)(nil)
)

// Logger returns the application logger.
func (a *App) Logger() *xlog.Logger {
	return a.log
}

// Get returns the named resource of the given kind.
func (a *App) Get(kind, name string) (any, error) {
	if a.resources == nil {
		return nil, errors.New("app: resource runtime is not initialized")
	}
	return a.resources.Get(kind, name)
}

// MustGet returns the named resource of the given kind and panics on error.
func (a *App) MustGet(kind, name string) any {
	if a.resources == nil {
		panic("app: resource runtime is not initialized")
	}
	return a.resources.MustGet(kind, name)
}

// RPCClient returns an outbound RPC client connection for target.
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

// CloseRpcClients closes cached outbound RPC clients.
func (a *App) CloseRpcClients() {
	if a.rpc == nil {
		return
	}
	if err := a.rpc.CloseClients(); err != nil {
		a.log.Error("failed to close rpc clients", "error", err)
	}
}

// buildContext adapts App capabilities to the BuildContext interface.
type buildContext struct {
	a *App
}

// newBuildContext creates a BuildContext backed by a.
func newBuildContext(a *App) BuildContext {
	return &buildContext{a: a}
}

func (c *buildContext) Logger() *xlog.Logger {
	return c.a.Logger()
}

func (c *buildContext) Container() Container {
	return c.a.container
}

func (c *buildContext) Resources() ResourceResolver {
	return c.a
}

func (c *buildContext) RPC() RPCClientResolver {
	return rpcClientResolverFunc(c.a.RPCClient)
}

// Telemetry returns the assembled telemetry runtime.
func (c *buildContext) Telemetry() *telemetry.Runtime {
	return c.a.telemetry
}

// rpcClientResolverFunc adapts a function to RPCClientResolver.
type rpcClientResolverFunc func(target string) (*grpc.ClientConn, error)

func (f rpcClientResolverFunc) Client(target string) (*grpc.ClientConn, error) {
	return f(target)
}

// rpcRegistrar adapts App capabilities to the RPCRegistrar interface.
type rpcRegistrar struct {
	a *App
}

// newRPCRegistrar creates an RPCRegistrar backed by a.
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

// apiRegistrar adapts App capabilities to the APIRegistrar interface.
type apiRegistrar struct {
	a *App
}

// newAPIRegistrar creates an APIRegistrar backed by a.
func newAPIRegistrar(a *App) APIRegistrar {
	return &apiRegistrar{a: a}
}

func (r *apiRegistrar) Logger() *xlog.Logger {
	return r.a.Logger()
}

func (r *apiRegistrar) Resolve(target any) error {
	return r.a.container.Resolve(target)
}

func (r *apiRegistrar) MustResolve(target any) {
	r.a.container.MustResolve(target)
}

func (r *apiRegistrar) RegisterAPI(register func(engine *api.Engine)) error {
	if r.a.api == nil {
		return errors.New("app: api runtime is not initialized")
	}
	return r.a.api.Register(register)
}

// jobRegistrar adapts App capabilities to the JobRegistrar interface.
type jobRegistrar struct {
	a *App
}

// newJobRegistrar creates a JobRegistrar backed by a.
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
