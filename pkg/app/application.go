package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/config"
	"github.com/HorseArcher567/octopus/pkg/database"
	"github.com/HorseArcher567/octopus/pkg/etcd"
	"github.com/HorseArcher567/octopus/pkg/job"
	redisclient "github.com/HorseArcher567/octopus/pkg/redis"
	"github.com/HorseArcher567/octopus/pkg/resource"
	"github.com/HorseArcher567/octopus/pkg/rpc"
	"github.com/HorseArcher567/octopus/pkg/rpc/resolver"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/sync/errgroup"
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

	// resources holds shared infrastructure resources initialized from configuration.
	resources *resource.Manager

	// shutdownTimeout is the timeout for graceful shutdown.
	shutdownTimeout time.Duration

	// startupHooks are startup hooks executed before components start.
	startupHooks []StartupHook
	// shutdownHooks are hooks executed during shutdown.
	shutdownHooks []ShutdownHook

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

func (a *App) initResources(cfg *resource.Config) {
	a.resources = resource.MustNew(cfg)
}

// OnStartup registers a startup hook.
// Hooks run in registration order. The first error aborts startup.
func (a *App) OnStartup(h StartupHook) *App {
	if h != nil {
		a.startupHooks = append(a.startupHooks, h)
	}
	return a
}

// OnShutdown registers a shutdown hook.
// Even if a hook returns an error, subsequent hooks continue to run.
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

// Logger returns the application logger.
func (a *App) Logger() *xlog.Logger {
	return a.log
}

// Resources returns the shared infrastructure resources manager.
func (a *App) Resources() *resource.Manager {
	return a.resources
}

// MySQL returns a named MySQL connection from the shared resource manager.
func (a *App) MySQL(name string) (*database.DB, error) {
	if a.resources == nil {
		return nil, errors.New("app: resources are not initialized (check resources config)")
	}
	return a.resources.MySQL(name)
}

// Redis returns a named Redis client from the shared resource manager.
func (a *App) Redis(name string) (*redisclient.Client, error) {
	if a.resources == nil {
		return nil, errors.New("app: resources are not initialized (check resources config)")
	}
	return a.resources.Redis(name)
}

const defaultShutdownTimeout = 30 * time.Second

// Run starts all configured components (such as the RPC server, API server, and job scheduler)
// and blocks until ctx is canceled or any component exits with an error.
// Run guarantees graceful shutdown and resource cleanup before returning.
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

	a.log.Info("starting application ...")
	defer func() {
		shutdownErr := a.shutdown()
		retErr = errors.Join(retErr, shutdownErr)
	}()

	// 1. Before hooks
	if err := a.execStartupHooks(ctx); err != nil {
		a.log.Error("failed to run application startup hooks", "error", err)
		a.log.Info("application exited with error", "error", err)
		retErr = err
		return retErr
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
	retErr = g.Wait()
	a.log.Info("all components exited", "error", retErr)
	if errors.Is(retErr, context.Canceled) || errors.Is(retErr, context.DeadlineExceeded) {
		return nil
	}
	return retErr
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
			a.log.Error("shutdown hook error", "error", err)
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
				a.log.Error("error stopping api server", "error", err)
				errs = append(errs, fmt.Errorf("stop api server: %w", err))
			}
		}
		if a.rpcServer != nil {
			if err := a.rpcServer.Stop(shutdownCtx); err != nil {
				a.log.Error("error stopping rpc server", "error", err)
				errs = append(errs, fmt.Errorf("stop rpc server: %w", err))
			}
		}
		if a.jobScheduler != nil {
			if err := a.jobScheduler.Stop(shutdownCtx); err != nil {
				a.log.Error("error stopping job scheduler", "error", err)
				errs = append(errs, fmt.Errorf("stop job scheduler: %w", err))
			}
		}

		if err := a.execShutdownHooks(shutdownCtx); err != nil {
			errs = append(errs, fmt.Errorf("shutdown hooks: %w", err))
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
