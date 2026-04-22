package assemble

import (
	"fmt"
	"strings"

	"github.com/HorseArcher567/octopus/pkg/discovery"
	"github.com/HorseArcher567/octopus/pkg/rpc"
	"github.com/HorseArcher567/octopus/pkg/store"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func setupRPC(c *setupContext) error {
	if _, ok := c.get("rpcServer"); !ok {
		return nil
	}
	var cfg rpc.ServerConfig
	if err := c.decodeStruct("rpcServer", &cfg); err != nil {
		return err
	}
	log, err := selectComponentLogger(cfg.Logger, c.state.log, c.state.store)
	if err != nil {
		return fmt.Errorf("assemble: rpcServer.logger: %w", err)
	}
	opts := []rpc.Option{}
	if cfg.Advertise != nil {
		if strings.TrimSpace(cfg.Advertise.Etcd) == "" {
			return fmt.Errorf("assemble: rpcServer.advertise.etcd is required")
		}
		client, err := store.GetNamed[*clientv3.Client](c.state.store, cfg.Advertise.Etcd)
		if err != nil {
			return fmt.Errorf("assemble: rpcServer.advertise.etcd: %w", err)
		}
		opts = append(opts, rpc.WithRegistrar(discovery.NewEtcdRegistrar(log, client)))
	}
	server, err := rpc.NewServer(log, &cfg, opts...)
	if err != nil {
		return fmt.Errorf("assemble: rpc server: %w", err)
	}
	c.state.rpc = server
	return nil
}
