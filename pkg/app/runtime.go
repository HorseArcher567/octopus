package app

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/database"
	"github.com/HorseArcher567/octopus/pkg/job"
	redisclient "github.com/HorseArcher567/octopus/pkg/redis"
	"github.com/HorseArcher567/octopus/pkg/rpc"
	"github.com/HorseArcher567/octopus/pkg/rpc/resolver"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"google.golang.org/grpc"
)

var _ Runtime = (*App)(nil)

func (a *App) Logger() *xlog.Logger {
	return a.log
}

func (a *App) MySQL(name string) (*database.DB, error) {
	if a.resources == nil {
		return nil, errors.New("app: resources are not initialized (check resources config)")
	}
	return a.resources.MySQL(name)
}

func (a *App) Redis(name string) (*redisclient.Client, error) {
	if a.resources == nil {
		return nil, errors.New("app: resources are not initialized (check resources config)")
	}
	return a.resources.Redis(name)
}

func (a *App) RegisterRPC(register func(s *grpc.Server)) {
	if a.rpcServer == nil {
		panic("app: rpc server is not initialized (check rpcServer config)")
	}
	if register != nil {
		a.rpcServer.RegisterServices(register)
	}
}

func (a *App) RegisterHTTP(register func(engine *api.Engine)) {
	if a.apiServer == nil {
		panic("app: api server is not initialized (check apiServer config)")
	}
	if register != nil {
		register(a.apiServer.Engine())
	}
}

func (a *App) AddJob(name string, fn job.Func) {
	a.jobScheduler.AddJob(&job.Job{Name: name, Func: fn})
}

// NewRPCClient returns a new gRPC client connection for the given target.
func (a *App) NewRPCClient(target string) (*grpc.ClientConn, error) {
	a.rpcCliMu.Lock()
	if conn, ok := a.rpcClients[target]; ok {
		a.rpcCliMu.Unlock()
		a.log.Debug("reuse rpc client connection", "target", target)
		return conn, nil
	}
	a.rpcCliMu.Unlock()

	dialOpts := append([]grpc.DialOption{}, a.rpcCliOptions...)

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

	conn, err := rpc.NewClient(target, dialOpts...)
	if err != nil {
		return nil, err
	}

	a.rpcCliMu.Lock()
	if existing, ok := a.rpcClients[target]; ok {
		a.rpcCliMu.Unlock()
		_ = conn.Close()
		a.log.Debug("reuse rpc client connection after race", "target", target)
		return existing, nil
	}
	a.rpcClients[target] = conn
	a.rpcCliMu.Unlock()
	a.log.Info("created rpc client connection", "target", target)

	return conn, nil
}

func (a *App) CloseRpcClients() {
	start := time.Now()
	a.rpcCliMu.Lock()
	clients := a.rpcClients
	a.rpcClients = make(map[string]*grpc.ClientConn)
	a.rpcCliMu.Unlock()

	if len(clients) == 0 {
		a.log.Debug("no rpc clients to close")
		return
	}

	var errs []error
	for target, conn := range clients {
		if err := conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close rpc client %q: %w", target, err))
		}
	}
	if len(errs) > 0 {
		a.log.Error("failed to close some rpc clients", "error", errors.Join(errs...))
		return
	}
	a.log.Info("closed rpc clients", "count", len(clients), "duration", time.Since(start))
}
