package app

// This file defines the module contracts and phase-specific capability
// interfaces used by App during orchestration.

import (
	"context"

	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/job"
	"github.com/HorseArcher567/octopus/pkg/telemetry"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"google.golang.org/grpc"
)

// Module is the minimal unit managed by App.
//
// Every module must provide a stable non-empty ID. Additional lifecycle
// participation is expressed by implementing optional phase interfaces such as
// BuildModule, RegisterRPCModule, RegisterAPIModule, RegisterJobsModule,
// RunModule, and CloseModule.
type Module interface {
	ID() string
}

// DependentModule optionally declares module dependencies by module ID.
//
// Dependencies are used to compute module execution order across all phases.
type DependentModule interface {
	DependsOn() []string
}

// BuildModule participates in dependency construction.
//
// Build is intended for wiring dependencies, publishing container bindings, and
// acquiring shared resources needed by later phases.
type BuildModule interface {
	Build(ctx context.Context, b BuildContext) error
}

// CloseModule owns explicit cleanup.
//
// Close is called during application shutdown for modules that became active in
// at least one phase.
type CloseModule interface {
	Close(ctx context.Context) error
}

// RegisterRPCModule participates in RPC registration.
//
// RegisterRPC should register inbound gRPC services against the provided
// registrar.
type RegisterRPCModule interface {
	RegisterRPC(ctx context.Context, r RPCRegistrar) error
}

// RegisterAPIModule participates in API route registration.
//
// RegisterAPI should register API routes or middleware against the provided
// registrar.
type RegisterAPIModule interface {
	RegisterAPI(ctx context.Context, r APIRegistrar) error
}

// RegisterJobsModule participates in background job registration.
//
// RegisterJobs should register scheduled or background work against the
// provided registrar.
type RegisterJobsModule interface {
	RegisterJobs(ctx context.Context, r JobRegistrar) error
}

// RunModule owns a long-running loop managed by App.
//
// Run is started after build and registration phases complete successfully.
type RunModule interface {
	Run(ctx context.Context) error
}

// Resolver exposes read-only dependency lookup.
//
// Resolver is primarily used during registration phases where modules need
// access to values produced earlier during Build.
type Resolver interface {
	Resolve(target any) error
	MustResolve(target any)
}

// Container exposes dependency publication during build.
//
// In addition to read-only resolution, Container supports publishing unnamed
// and named bindings and invoking functions with auto-resolved arguments.
type Container interface {
	Resolver
	Provide(value any) error
	ProvideNamed(name string, value any) error
	ResolveNamed(name string, target any) error
	Invoke(fn any) error
}

// ResourceResolver exposes generic resource lookup.
//
// ResourceResolver provides access to shared infrastructure resources by kind
// and name.
type ResourceResolver interface {
	Get(kind, name string) (any, error)
	MustGet(kind, name string) any
}

// RPCClientResolver exposes outbound RPC client creation.
//
// Targets are interpreted by the configured RPC runtime and may use direct or
// discovery-backed resolution schemes.
type RPCClientResolver interface {
	Client(target string) (*grpc.ClientConn, error)
}

// BuildContext exposes host capabilities needed during dependency construction.
//
// BuildContext intentionally groups capabilities instead of exposing concrete
// infrastructure shortcuts so modules depend on stable framework contracts.
type BuildContext interface {
	Logger() *xlog.Logger
	Container() Container
	Resources() ResourceResolver
	RPC() RPCClientResolver
	Telemetry() *telemetry.Runtime
}

// RPCRegistrar exposes RPC service registration.
//
// It also provides read-only dependency resolution for services assembled
// during the Build phase.
type RPCRegistrar interface {
	Logger() *xlog.Logger
	Resolver
	RegisterRPC(register func(s *grpc.Server)) error
}

// APIRegistrar exposes API route registration.
//
// It also provides read-only dependency resolution for handlers assembled
// during the Build phase.
type APIRegistrar interface {
	Logger() *xlog.Logger
	Resolver
	RegisterAPI(register func(engine *api.Engine)) error
}

// JobRegistrar exposes background job registration.
//
// It also provides read-only dependency resolution for job dependencies
// assembled during the Build phase.
type JobRegistrar interface {
	Logger() *xlog.Logger
	Resolver
	AddJob(name string, fn job.Func) error
}
