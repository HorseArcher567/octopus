package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/etcd"
	"github.com/HorseArcher567/octopus/pkg/job"
	"github.com/HorseArcher567/octopus/pkg/rpc"
	"github.com/HorseArcher567/octopus/pkg/rpc/resolver"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

// BeforeRunHook is executed before the service starts. If it returns an error, the startup process will be aborted.
type BeforeRunHook func(ctx context.Context, a *App) error

// ShutdownHook is executed during service shutdown. Even if it returns an error, subsequent hooks will continue to execute.
type ShutdownHook func(ctx context.Context, a *App) error

// App encapsulates the application lifecycle that hosts both gRPC and HTTP API services, and job execution.
type App struct {
	// log is the application-wide logger instance.
	log *xlog.Logger

	// rpcServer is the gRPC server managed by the application.
	rpcServer *rpc.Server
	// apiServer is the HTTP API server managed by the application.
	apiServer *api.Server
	// jobScheduler is the job scheduler managed by the application.
	jobScheduler *job.Scheduler

	// rpcCliOptions are the options applied when creating RPC clients.
	rpcCliOptions []grpc.DialOption

	// etcdClient is the shared etcd client used for service discovery and configuration.
	etcdClient *clientv3.Client

	// shutdownTimeout is the timeout for graceful shutdown.
	shutdownTimeout time.Duration

	// beforeRunHooks are hooks executed before the application starts running.
	beforeRunHooks []BeforeRunHook
	// shutdownHooks are hooks executed during application shutdown.
	shutdownHooks []ShutdownHook
}

// New creates a new App instance and immediately initializes it based on the framework configuration.
// framework is the framework configuration. Users should load their own configuration externally (embedding app.Framework), then extract and pass the Framework part.
//
// Example:
//
//	type AppConfig struct {
//	    app.Framework
//	    Database struct { ... } `yaml:"database"`
//	}
//
//	var cfg AppConfig
//	config.MustUnmarshal("config.yaml", &cfg)
//	application := app.New(&cfg.Framework)
func New(framework *Framework) *App {
	if framework == nil {
		panic("app: framework config cannot be nil")
	}

	a := &App{
		shutdownTimeout: framework.ShutdownTimeout,
	}

	// Initialize logger
	a.initLogger(&framework.LoggerCfg)

	// Init RPC client options if configured
	a.initRpcCliOptions(&framework.RpcCliOptions)

	// Init job schedule if configured
	a.initJobSchedule()

	// Init etcd client if configured
	if framework.EtcdCfg != nil {
		a.initEtcdClient(framework.EtcdCfg)
	}

	// Init RPC server if configured
	if framework.RpcSvrCfg != nil {
		a.initRpcServer(framework.RpcSvrCfg)
	}

	// Init HTTP API server if configured
	if framework.ApiSvrCfg != nil {
		a.initApiServer(framework.ApiSvrCfg)
	}

	return a
}

// initLogger initializes the logger if configured.
func (a *App) initLogger(cfg *xlog.Config) {
	a.log = xlog.MustNew(cfg)
}

func (a *App) initEtcdClient(cfg *etcd.Config) {
	a.etcdClient = etcd.MustNewClient(cfg)
}

// initRpcServer initializes the RPC server if configured.
// If the server should register to etcd, it also sets the etcd configuration.
func (a *App) initRpcServer(cfg *rpc.ServerConfig) {
	a.rpcServer = rpc.MustNewServer(a.log, cfg, rpc.WithEtcdClient(a.etcdClient))
}

func (a *App) initRpcCliOptions(cfg *rpc.ClientOptions) {
	a.rpcCliOptions = cfg.BuildDialOptions()
}

func (a *App) initJobSchedule() {
	a.jobScheduler = job.NewScheduler(a.log)
}

// initApiServer initializes the HTTP API server if configured.
func (a *App) initApiServer(cfg *api.ServerConfig) {
	a.apiServer = api.MustNewServer(a.log, cfg)
}

// OnBeforeRun registers a hook to be executed before Run.
// Hooks are executed in registration order. The first error encountered will abort the startup process.
func (a *App) OnBeforeRun(h BeforeRunHook) *App {
	if h != nil {
		a.beforeRunHooks = append(a.beforeRunHooks, h)
	}
	return a
}

// OnShutdown registers a hook to be executed during service shutdown.
// Even if a hook returns an error, subsequent hooks will continue to execute.
func (a *App) OnShutdown(h ShutdownHook) *App {
	if h != nil {
		a.shutdownHooks = append(a.shutdownHooks, h)
	}
	return a
}

// RegisterRpcServices registers gRPC services.
func (a *App) RegisterRpcServices(register func(s *grpc.Server)) *App {
	if a.rpcServer == nil {
		panic("app: rpc server is not initialized (check RpcServer config)")
	}
	a.rpcServer.RegisterServices(register)
	return a
}

// RegisterApiRoutes registers HTTP API routes.
// Usually called in main, registering routes on gin.Engine through the provided function.
func (a *App) RegisterApiRoutes(register func(engine *api.Engine)) *App {
	if a.apiServer == nil {
		panic("app: api server is not initialized (check ApiServer config)")
	}
	if register != nil {
		register(a.apiServer.Engine())
	}
	return a
}

const defaultShutdownTimeout = 30 * time.Second

// Run starts all configured components (such as the RPC server, API server, and job scheduler)
// and blocks until ctx is canceled or all components exit.
// It runs registered BeforeRun hooks, starts components concurrently, and waits for them to terminate.
// The caller is responsible for invoking Stop after Run returns to perform graceful shutdown and cleanup.
func (a *App) Run(ctx context.Context) error {
	a.log.Info("starting application ...")

	// 1. Before hooks
	if err := a.runBeforeRunHooks(ctx); err != nil {
		a.log.Error("failed to run application startup hooks", "error", err)
		a.log.Info("application exited with error", "error", err)
		return err
	}

	// 2. Start all components concurrently
	g, runCtx := errgroup.WithContext(ctx)

	if a.rpcServer != nil {
		a.log.Info("starting rpc server ...")
		g.Go(func() error { return a.rpcServer.Run(runCtx) })
	}

	if a.apiServer != nil {
		a.log.Info("starting api server ...")
		g.Go(func() error { return a.apiServer.Run(runCtx) })
	}

	if a.jobScheduler != nil {
		a.log.Info("starting job scheduler ...")
		g.Go(func() error { return a.jobScheduler.Run(runCtx) })
	}

	// 3. Wait for all components to exit
	// This blocks until: ctx is cancelled OR all components return
	runErr := g.Wait()
	a.log.Info("all components exited", "error", runErr)

	return runErr
}

// Stop gracefully shuts down all components and releases shared resources.
// It creates a bounded shutdown context, stops each component, runs registered Shutdown hooks,
// and then closes RPC clients and the logger.
// It is safe to call Stop multiple times; component shutdown should be idempotent.
func (a *App) Stop() {
	// 1. Create shutdown context for graceful stop
	timeout := a.shutdownTimeout
	if timeout == 0 {
		timeout = defaultShutdownTimeout
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 2. Stop all components gracefully (idempotent - safe even if already stopped)
	if a.jobScheduler != nil {
		if err := a.jobScheduler.Stop(shutdownCtx); err != nil {
			a.log.Error("error stopping job scheduler", "error", err)
		}
	}
	if a.apiServer != nil {
		if err := a.apiServer.Stop(shutdownCtx); err != nil {
			a.log.Error("error stopping api server", "error", err)
		}
	}
	if a.rpcServer != nil {
		if err := a.rpcServer.Stop(shutdownCtx); err != nil {
			a.log.Error("error stopping rpc server", "error", err)
		}
	}

	// 7. Shutdown hooks
	a.runShutdownHooks(shutdownCtx)

	// 8. Cleanup
	a.CloseRpcClients()
	a.log.Close()
}

func (a *App) runBeforeRunHooks(ctx context.Context) error {
	for _, h := range a.beforeRunHooks {
		if err := h(ctx, a); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) runShutdownHooks(ctx context.Context) {
	for _, h := range a.shutdownHooks {
		if err := h(ctx, a); err != nil {
			a.log.Error("shutdown hook error", "error", err)
		}
	}
}

// NewRpcClient returns a new gRPC client connection for the given target.
//
// The connection is created on each call.
// The resolver is automatically configured based on the target scheme:
//   - etcd:/// targets use etcd service discovery (requires etcd to be configured)
//   - direct:/// targets use direct connection resolver
//   - host:port targets use standard gRPC connection
//
// Example:
//
//	conn, err := app.NewRpcClient("etcd:///user-service")
//	if err != nil {
//	    log.Error("failed to get rpc client", "error", err)
//	    return err
//	}
//	userClient := pb.NewUserServiceClient(conn)
//
// Returns an error if:
//   - The connection cannot be established
//   - etcd:/// target is used but etcd is not configured
func (a *App) NewRpcClient(target string) (*grpc.ClientConn, error) {
	// Start with default dial options from configuration.
	dialOpts := append([]grpc.DialOption{}, a.rpcCliOptions...)

	// Add resolver based on target scheme.
	switch {
	case strings.HasPrefix(target, "etcd:///"):
		if a.etcdClient == nil {
			return nil, fmt.Errorf("etcd client is not configured for target %q", target)
		}
		etcdBuilder := resolver.NewEtcdBuilder(a.log, a.etcdClient)
		dialOpts = append(dialOpts, grpc.WithResolvers(etcdBuilder))
	case strings.HasPrefix(target, "direct:///"):
		directBuilder := resolver.NewDirectBuilder(a.log)
		dialOpts = append(dialOpts, grpc.WithResolvers(directBuilder))
	}

	// Create connection with automatic resolver configuration.
	return rpc.NewClient(target, dialOpts...)
}

// CloseRpcClients closes all cached RPC client connections.
// This is called automatically during app shutdown.
func (a *App) CloseRpcClients() {
	// TODO: implement
}

func (a *App) AddJob(name string, fn job.Func) {
	a.jobScheduler.AddJob(
		&job.Job{
			Name: name,
			Func: fn,
		})
}
