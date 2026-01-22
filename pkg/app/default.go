package app

import (
	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/job"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"google.golang.org/grpc"
)

// defaultApp is the global default application instance, similar to slog.Default.
var defaultApp *App

// Init initializes the global default application instance.
// It should be called once during program startup, typically in main.
// The framework parameter is the framework configuration, usually embedded in the user's own config struct.
func Init(framework *Framework) {
	defaultApp = New(framework)
}

// Default returns the current global default application instance.
// It panics if Init has not been called.
func Default() *App {
	if defaultApp == nil {
		panic("app: defaultApp is not initialized, call app.Init() first")
	}
	return defaultApp
}

// Logger returns the logger of the default application instance.
func Logger() *xlog.Logger {
	return Default().log
}

// OnBeforeRun registers a hook to be executed before Run on the default application instance.
func OnBeforeRun(h BeforeRunHook) {
	Default().OnBeforeRun(h)
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

// Run starts the default application instance.
func Run() {
	if defaultApp == nil {
		panic("app: defaultApp is not initialized, call app.Init() first")
	}
	defaultApp.Run()
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
