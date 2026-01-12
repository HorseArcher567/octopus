package etcd

import (
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// Config etcd 连接配置
type Config struct {
	// Endpoints etcd 节点地址列表
	Endpoints []string `yaml:"endpoints" json:"endpoints" toml:"endpoints"`

	// DialTimeout 连接超时时间（默认 5s）
	DialTimeout time.Duration `yaml:"dialTimeout" json:"dialTimeout" toml:"dialTimeout"`

	// AutoSyncInterval 自动同步集群成员间隔（默认 60s）
	AutoSyncInterval time.Duration `yaml:"autoSyncInterval" json:"autoSyncInterval" toml:"autoSyncInterval"`

	// Username etcd 用户名（可选）
	Username string `yaml:"username" json:"username" toml:"username"`

	// Password etcd 密码（可选）
	Password string `yaml:"password" json:"password" toml:"password"`
}

// ClientV3Config 生成用于创建 etcd 客户端的 clientv3.Config
func (c *Config) ClientV3Config() clientv3.Config {
	cfg := clientv3.Config{
		Endpoints: c.Endpoints,
	}

	if c.DialTimeout > 0 {
		cfg.DialTimeout = c.DialTimeout
	} else {
		cfg.DialTimeout = 5 * time.Second
	}

	if c.AutoSyncInterval > 0 {
		cfg.AutoSyncInterval = c.AutoSyncInterval
	} else {
		cfg.AutoSyncInterval = 60 * time.Second
	}

	if c.Username != "" {
		cfg.Username = c.Username
		cfg.Password = c.Password
	}

	return cfg
}

// IsEmpty 检查配置是否为空（未配置 endpoints）
func (c *Config) IsEmpty() bool {
	return c == nil || len(c.Endpoints) == 0
}
