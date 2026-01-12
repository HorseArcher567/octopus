package etcd

import (
	"errors"
	"sync"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// 定义错误
var (
	ErrEmptyConfig = errors.New("etcd: config is empty or endpoints not configured")
)

var (
	defaultConfigMu sync.RWMutex
	defaultConfig   *Config
)

// Default 返回默认的 etcd 配置
// 如果未设置，返回 nil
func Default() *Config {
	defaultConfigMu.RLock()
	defer defaultConfigMu.RUnlock()
	return defaultConfig
}

// SetDefault 设置默认的 etcd 配置
func SetDefault(cfg *Config) {
	defaultConfigMu.Lock()
	defer defaultConfigMu.Unlock()
	defaultConfig = cfg
}

// NewClient 使用配置创建 etcd client
// 如果 cfg 为 nil，则使用默认配置
// 如果默认配置也为 nil 或为空，返回错误
func NewClient(cfg *Config) (*clientv3.Client, error) {
	if cfg == nil {
		cfg = Default()
	}

	if cfg.IsEmpty() {
		return nil, ErrEmptyConfig
	}

	clientCfg := cfg.ClientV3Config()
	return clientv3.New(clientCfg)
}

// MustNewClient 使用配置创建 etcd client，失败时 panic
func MustNewClient(cfg *Config) *clientv3.Client {
	client, err := NewClient(cfg)
	if err != nil {
		panic("etcd: failed to create client: " + err.Error())
	}
	return client
}
