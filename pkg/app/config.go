package app

import (
	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/etcd"
	"github.com/HorseArcher567/octopus/pkg/rpc"
	"github.com/HorseArcher567/octopus/pkg/xlog"
)

// Framework holds framework-level configuration such as logging, RPC, API server,
// job runner and RPC clients.
// It is intended to be embedded into the user's own application config struct.
// After loading the full config externally, the embedded Framework part should be
// passed to app.Init.
//
// Example:
//
//	type AppConfig struct {
//	    app.Framework
//	    Database struct {
//	        Host string `yaml:"host"`
//	        Port int    `yaml:"port"`
//	    } `yaml:"database"`
//	}
//
//	func main() {
//	    var cfg AppConfig
//	    config.MustUnmarshal("config.yaml", &cfg)
//	    app.Init(&cfg.Framework)
//	}
type Framework struct {
	// LoggerCfg configures the application logger.
	LoggerCfg xlog.Config `yaml:"logger" json:"logger" toml:"logger"`

	// EtcdCfg configures the shared etcd client used for service discovery and other integrations.
	EtcdCfg *etcd.Config `yaml:"etcd" json:"etcd" toml:"etcd"`

	// RpcSvrCfg configures the gRPC server.
	RpcSvrCfg *rpc.ServerConfig `yaml:"rpcServer" json:"rpcServer" toml:"rpcServer"`

	// RpcCliOptions configures default options applied when creating RPC clients.
	RpcCliOptions rpc.ClientOptions `yaml:"rpcClientOptions" json:"rpcClientOptions" toml:"rpcClientOptions"`

	// ApiSvrCfg configures the HTTP API server.
	ApiSvrCfg *api.ServerConfig `yaml:"apiServer" json:"apiServer" toml:"apiServer"`
}
