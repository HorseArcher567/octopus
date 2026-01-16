package app

import (
	"context"
	"sync"
	"time"

	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/etcd"
	"github.com/HorseArcher567/octopus/pkg/rpc"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

// BeforeRunHook is executed before the service starts. If it returns an error, the startup process will be aborted.
type BeforeRunHook func(ctx context.Context, a *App)

// ShutdownHook is executed during service shutdown. Even if it returns an error, subsequent hooks will continue to execute.
type ShutdownHook func(ctx context.Context, a *App)

// App encapsulates the application lifecycle that hosts both gRPC and HTTP API services.
type App struct {
	log *xlog.Logger

	rpcServer *rpc.Server
	apiServer *api.Server

	etcdClient *clientv3.Client

	beforeRunHooks []BeforeRunHook
	shutdownHooks  []ShutdownHook
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

	a := &App{}

	// Initialize logger
	a.initLogger(&framework.LoggerCfg)

	// Init etcd client if configured
	if framework.EtcdCfg != nil {
		a.initEtcdClient(framework.EtcdCfg)
	}

	// Init RPC server if configured
	if framework.RpcSvrCfg != nil {
		a.initRpcServer(framework.RpcSvrCfg)
		if framework.RpcSvrCfg.ShouldRegisterInstance() {
		}
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
	a.rpcServer = rpc.MustNewServer(a.log, cfg)
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

// Run starts the application and blocks until all enabled services stop.
// Execution order:
// 1) Run OnBeforeRun hooks (any error aborts startup);
// 2) Start RPC Server and HTTP API Server concurrently (enabled based on configuration);
// 3) Run OnShutdown hooks after all services stop.
func (a *App) Run() {
	// 1. BeforeRun hooks
	a.runBeforeRunHooks()

	// 2. Start services concurrently
	var wg sync.WaitGroup

	if a.rpcServer != nil {
		wg.Go(func() {
			a.rpcServer.Start()
		})
	}

	if a.apiServer != nil {
		wg.Go(func() {
			a.apiServer.Start()
		})
	}

	wg.Wait()

	// 3. Shutdown hooks (execute even if services reported errors)
	a.runShutdownHooks()

	// Close logger resources
	a.log.Close()
}

func (a *App) runBeforeRunHooks() {
	ctx := context.Background()
	for _, h := range a.beforeRunHooks {
		h(ctx, a)
	}
}

func (a *App) runShutdownHooks() {
	// Fixed timeout duration, can be extended to read from configuration later.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, h := range a.shutdownHooks {
		h(ctx, a)
	}
}
