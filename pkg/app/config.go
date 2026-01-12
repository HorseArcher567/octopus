package app

import (
	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/etcd"
	"github.com/HorseArcher567/octopus/pkg/rpc"
	"github.com/HorseArcher567/octopus/pkg/xlog"
)

// Framework 是框架级配置，聚合日志、RPC 与 API 服务配置。
// 用户应该在自己的配置结构体中嵌入此类型，在外部加载配置后，将 Framework 部分传给 app.Init。
//
// 示例：
//
//	type AppConfig struct {
//	    app.Framework  // 嵌入框架配置
//	    Database struct {
//	        Host string `yaml:"host"`
//	        Port int    `yaml:"port"`
//	    } `yaml:"database"`
//	}
//
//	func main() {
//	    var cfg AppConfig
//	    config.MustUnmarshal("config.yaml", &cfg)
//	    app.Init(&cfg.Framework)  // 只传入框架配置部分
//	}
type Framework struct {
	LoggerCfg *xlog.Config      `yaml:"logger" json:"logger" toml:"logger"`
	EtcdCfg   *etcd.Config      `yaml:"etcd" json:"etcd" toml:"etcd"`
	RpcSvrCfg *rpc.ServerConfig `yaml:"rpcServer" json:"rpcServer" toml:"rpcServer"`
	ApiSvrCfg *api.ServerConfig `yaml:"apiServer" json:"apiServer" toml:"apiServer"`
}
