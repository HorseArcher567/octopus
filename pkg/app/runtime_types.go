package app

import (
	"context"

	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/database"
	"github.com/HorseArcher567/octopus/pkg/job"
	redisclient "github.com/HorseArcher567/octopus/pkg/redis"
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

// HTTPRuntime owns HTTP route registration and server lifecycle.
type HTTPRuntime interface {
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
	MySQL(name string) (*database.DB, error)
	Redis(name string) (*redisclient.Client, error)
	Close() error
}
