package assemble

import (
	"fmt"
	"strings"

	"github.com/HorseArcher567/octopus/pkg/discovery"
	"github.com/HorseArcher567/octopus/pkg/rpc"
	"github.com/HorseArcher567/octopus/pkg/store"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func setupRPCResolver(c *setupContext) error {
	if _, ok := c.get("rpcResolver"); !ok {
		return nil
	}
	var cfg rpc.ResolverConfig
	if err := c.decodeStruct("rpcResolver", &cfg); err != nil {
		return err
	}
	if cfg.Direct {
		rpc.RegisterResolver(discovery.NewDirectResolver(c.state.log).Builder())
	}
	if strings.TrimSpace(cfg.Etcd) != "" {
		client, err := store.GetNamed[*clientv3.Client](c.state.store, cfg.Etcd)
		if err != nil {
			return fmt.Errorf("assemble: rpcResolver.etcd: %w", err)
		}
		rpc.RegisterResolver(discovery.NewEtcdResolver(c.state.log, client).Builder())
	}
	return nil
}
