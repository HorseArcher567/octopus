package app

import (
	"context"

	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/database"
	"github.com/HorseArcher567/octopus/pkg/job"
	redisclient "github.com/HorseArcher567/octopus/pkg/redis"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"google.golang.org/grpc"
)

// Module defines a business module lifecycle.
// Init runs during startup. Close runs during shutdown in reverse order.
type Module interface {
	ID() string
	Init(ctx context.Context, rt Runtime) error
	Close(ctx context.Context) error
}

// DependedModule optionally declares module dependencies by module ID.
type DependedModule interface {
	DependsOn() []string
}

// Runtime exposes framework capabilities to modules.
// It is the single interaction boundary between modules and the app runtime.
type Runtime interface {
	Logger() *xlog.Logger
	MySQL(name string) (*database.DB, error)
	Redis(name string) (*redisclient.Client, error)
	NewRPCClient(target string) (*grpc.ClientConn, error)

	RegisterRPC(register func(s *grpc.Server))
	RegisterHTTP(register func(engine *api.Engine))
	AddJob(name string, fn job.Func)
}
