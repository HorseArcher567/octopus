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

// Module is the minimal unit managed by App.
type Module interface {
	ID() string
}

// DependentModule optionally declares module dependencies by module ID.
type DependentModule interface {
	DependsOn() []string
}

// BuildModule participates in dependency construction.
type BuildModule interface {
	Build(ctx context.Context, b BuildContext) error
}

// CloseModule owns explicit cleanup.
type CloseModule interface {
	Close(ctx context.Context) error
}

// RegisterRPCModule participates in RPC registration.
type RegisterRPCModule interface {
	RegisterRPC(ctx context.Context, r RPCRegistrar) error
}

// RegisterHTTPModule participates in HTTP route registration.
type RegisterHTTPModule interface {
	RegisterHTTP(ctx context.Context, r HTTPRegistrar) error
}

// RegisterJobsModule participates in background job registration.
type RegisterJobsModule interface {
	RegisterJobs(ctx context.Context, r JobRegistrar) error
}

// RunModule owns a long-running loop managed by App.
type RunModule interface {
	Run(ctx context.Context) error
}

// Resolver exposes read-only dependency lookup.
type Resolver interface {
	Resolve(target any) error
	MustResolve(target any)
}

// Container exposes dependency publication during build.
type Container interface {
	Resolver
	Provide(value any) error
}

// BuildContext exposes host capabilities needed during dependency construction.
type BuildContext interface {
	Logger() *xlog.Logger
	MySQL(name string) (*database.DB, error)
	Redis(name string) (*redisclient.Client, error)
	RPCClient(target string) (*grpc.ClientConn, error)
	Container
}

// RPCRegistrar exposes RPC service registration.
type RPCRegistrar interface {
	Logger() *xlog.Logger
	Resolver
	RegisterRPC(register func(s *grpc.Server)) error
}

// HTTPRegistrar exposes HTTP route registration.
type HTTPRegistrar interface {
	Logger() *xlog.Logger
	Resolver
	RegisterHTTP(register func(engine *api.Engine)) error
}

// JobRegistrar exposes background job registration.
type JobRegistrar interface {
	Logger() *xlog.Logger
	Resolver
	AddJob(name string, fn job.Func) error
}
