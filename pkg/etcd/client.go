package etcd

import (
	clientv3 "go.etcd.io/etcd/client/v3"
)

// NewClient 使用配置创建 etcd client
// 如果 cfg 为 nil，则使用默认配置
// 如果默认配置也为 nil 或为空，返回错误
func NewClient(cfg *Config) (*clientv3.Client, error) {
	clientV3Config, err := cfg.ClientV3Config()
	if err != nil {
		return nil, err
	}
	return clientv3.New(*clientV3Config)
}

// MustNewClient 使用配置创建 etcd client，失败时 panic
func MustNewClient(cfg *Config) *clientv3.Client {
	client, err := NewClient(cfg)
	if err != nil {
		panic("etcd: failed to create client: " + err.Error())
	}
	return client
}
