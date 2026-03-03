package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/config"
	"github.com/HorseArcher567/octopus/pkg/etcd"
	"github.com/HorseArcher567/octopus/pkg/job"
	"github.com/HorseArcher567/octopus/pkg/resource"
	"github.com/HorseArcher567/octopus/pkg/rpc"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

// StartupHook is executed before components are started.
// If it returns an error, startup is aborted.
type StartupHook func(ctx context.Context, a *App) error

// ShutdownHook is executed during shutdown.
// Even if it returns an error, subsequent shutdown hooks continue to run.
type ShutdownHook func(ctx context.Context, a *App) error

// App encapsulates the application lifecycle that hosts both gRPC and HTTP API services, and job execution.
type App struct {
	log *xlog.Logger

	rpcServer    *rpc.Server
	apiServer    *api.Server
	jobScheduler *job.Scheduler

	rpcCliOptions []grpc.DialOption
	etcdClient    *clientv3.Client
	resources     *resource.Manager
	rpcClients    map[string]*grpc.ClientConn
	rpcCliMu      sync.Mutex

	shutdownTimeout time.Duration
	startupHooks    []StartupHook
	shutdownHooks   []ShutdownHook

	modules            []Module
	orderedModules     []Module
	initializedModules []Module

	shutdownOnce sync.Once
	runMu        sync.Mutex
	hasRun       bool
}

// New creates a new App from a raw config object.
// It reads framework keys from cfg (logger/etcd/rpcServer/rpcClientOptions/apiServer/resources/shutdownTimeout).
func New(cfg *config.Config) *App {
	if cfg == nil {
		panic("app: config cannot be nil")
	}

	var (
		loggerCfg       xlog.Config
		etcdCfg         *etcd.Config
		rpcSvrCfg       *rpc.ServerConfig
		rpcCliOptions   rpc.ClientOptions
		apiSvrCfg       *api.ServerConfig
		resourcesCfg    *resource.Config
		shutdownTimeout time.Duration
	)

	unmarshalKeyIfExists(cfg, "logger", &loggerCfg)
	unmarshalKeyIfExists(cfg, "etcd", &etcdCfg)
	unmarshalKeyIfExists(cfg, "rpcServer", &rpcSvrCfg)
	unmarshalKeyIfExists(cfg, "rpcClientOptions", &rpcCliOptions)
	unmarshalKeyIfExists(cfg, "apiServer", &apiSvrCfg)
	unmarshalKeyIfExists(cfg, "resources", &resourcesCfg)
	shutdownTimeout = mustLoadDurationIfExists(cfg, "shutdownTimeout")

	a := &App{
		shutdownTimeout: shutdownTimeout,
		rpcClients:      make(map[string]*grpc.ClientConn),
	}
	a.initLogger(&loggerCfg)
	a.initRpcCliOptions(&rpcCliOptions)
	a.initJobSchedule()

	if etcdCfg != nil {
		a.initEtcdClient(etcdCfg)
	}
	if rpcSvrCfg != nil {
		a.initRpcServer(rpcSvrCfg)
	}
	if apiSvrCfg != nil {
		a.initApiServer(apiSvrCfg)
	}
	if resourcesCfg != nil {
		a.initResources(resourcesCfg)
	}

	return a
}

func unmarshalKeyIfExists(cfg *config.Config, key string, target any) {
	if !cfg.Has(key) {
		return
	}
	if err := cfg.UnmarshalKey(key, target); err != nil {
		panic(fmt.Errorf("app: invalid %s config: %w", key, err))
	}
}

func mustLoadDurationIfExists(cfg *config.Config, key string) time.Duration {
	value, ok := cfg.Get(key)
	if !ok {
		return 0
	}

	switch v := value.(type) {
	case time.Duration:
		return v
	case string:
		d, err := time.ParseDuration(v)
		if err != nil {
			panic(fmt.Errorf("app: invalid %s duration %q: %w", key, v, err))
		}
		return d
	case int:
		return time.Duration(v) * time.Second
	case int64:
		return time.Duration(v) * time.Second
	case float64:
		return time.Duration(v * float64(time.Second))
	default:
		panic(fmt.Errorf("app: invalid %s type %T, expected duration string or numeric seconds", key, value))
	}
}

func (a *App) initLogger(cfg *xlog.Config) {
	a.log = xlog.MustNew(cfg)
}

func (a *App) initEtcdClient(cfg *etcd.Config) {
	a.etcdClient = etcd.MustNewClient(cfg)
}

func (a *App) initRpcServer(cfg *rpc.ServerConfig) {
	a.rpcServer = rpc.MustNewServer(a.log, cfg, rpc.WithEtcdClient(a.etcdClient))
}

func (a *App) initRpcCliOptions(cfg *rpc.ClientOptions) {
	a.rpcCliOptions = cfg.BuildDialOptions()
}

func (a *App) initJobSchedule() {
	a.jobScheduler = job.NewScheduler(a.log)
}

func (a *App) initApiServer(cfg *api.ServerConfig) {
	a.apiServer = api.MustNewServer(a.log, cfg)
}

func (a *App) initResources(cfg *resource.Config) {
	a.resources = resource.MustNew(cfg)
}

// OnStartup registers a startup hook.
func (a *App) OnStartup(h StartupHook) *App {
	if h != nil {
		a.startupHooks = append(a.startupHooks, h)
	}
	return a
}

// OnShutdown registers a shutdown hook.
func (a *App) OnShutdown(h ShutdownHook) *App {
	if h != nil {
		a.shutdownHooks = append(a.shutdownHooks, h)
	}
	return a
}

// Use registers one or more modules on the app instance.
func (a *App) Use(mods ...Module) *App {
	for _, m := range mods {
		if m != nil {
			a.modules = append(a.modules, m)
		}
	}
	return a
}
