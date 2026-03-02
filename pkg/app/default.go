package app

import (
	"context"
	"os/signal"
	"sync"
	"syscall"

	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/config"
	"github.com/HorseArcher567/octopus/pkg/database"
	"github.com/HorseArcher567/octopus/pkg/job"
	redisclient "github.com/HorseArcher567/octopus/pkg/redis"
	"github.com/HorseArcher567/octopus/pkg/resource"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"google.golang.org/grpc"
)

// defaultApp is the global default application instance, similar to slog.Default.
var defaultApp *App
var defaultMu sync.Mutex
var runWithSignalsFn = RunWithSignals

// Init initializes the global default application instance from cfg.
// It can only be called once.
func Init(cfg *config.Config) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	if defaultApp != nil {
		panic("app: Init can only be called once")
	}
	defaultApp = New(cfg)
}

// MustRun loads config, initializes the default application, optionally wires dependencies,
// then runs with SIGTERM/SIGINT handling. It panics on any error.
func MustRun(configPath string, wire func() error) {
	cfg := config.MustLoad(configPath)
	Init(cfg)
	if wire != nil {
		if err := wire(); err != nil {
			panic(err)
		}
	}
	if err := runWithSignalsFn(); err != nil {
		panic(err)
	}
}

// Default returns the current global default application instance.
// It panics if Init has not been called.
func Default() *App {
	if defaultApp == nil {
		panic("app: defaultApp is not initialized, call app.Init() or app.MustRun() first")
	}
	return defaultApp
}

// Logger returns the logger of the default application instance.
func Logger() *xlog.Logger {
	return Default().Logger()
}

// OnStartup registers a startup hook on the default application instance.
func OnStartup(h StartupHook) {
	Default().OnStartup(h)
}

// OnShutdown registers a hook to be executed during shutdown on the default application instance.
func OnShutdown(h ShutdownHook) {
	Default().OnShutdown(h)
}

// RegisterRpcServices registers gRPC services on the default application instance.
func RegisterRpcServices(register func(s *grpc.Server)) {
	Default().RegisterRpcServices(register)
}

// RegisterApiRoutes registers HTTP API routes on the default application instance.
func RegisterApiRoutes(register func(engine *api.Engine)) {
	Default().RegisterApiRoutes(register)
}

// Resources returns the shared resource manager on the default application instance.
func Resources() *resource.Manager {
	return Default().Resources()
}

// MySQL returns a named MySQL connection from the default application instance.
func MySQL(name string) (*database.DB, error) {
	return Default().MySQL(name)
}

// Redis returns a named Redis client from the default application instance.
func Redis(name string) (*redisclient.Client, error) {
	return Default().Redis(name)
}

// Run starts the default application instance and blocks until ctx is cancelled.
func Run(ctx context.Context) error {
	return Default().Run(ctx)
}

// RunWithSignals starts the default application instance and blocks until ctx is cancelled or a signal is received.
func RunWithSignals() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	return Default().Run(ctx)
}

// MustNewRpcClient returns a gRPC client connection from the default application instance.
// It panics if the client cannot be created or retrieved.
func MustNewRpcClient(target string) *grpc.ClientConn {
	conn, err := Default().NewRpcClient(target)
	if err != nil {
		panic(err)
	}
	return conn
}

// NewRpcClient returns a gRPC client connection from the default application instance.
// If the connection does not exist, it is created lazily according to the configured options.
func NewRpcClient(target string) (*grpc.ClientConn, error) {
	return Default().NewRpcClient(target)
}

// AddJob adds a job to the default application instance.
func AddJob(name string, fn job.Func) {
	Default().AddJob(name, fn)
}
