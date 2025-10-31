package registry

import "time"

// ServiceInstance 服务实例信息
type ServiceInstance struct {
	Addr     string            `json:"addr"`
	Port     int               `json:"port"`
	Version  string            `json:"version"`
	Zone     string            `json:"zone,omitempty"`
	Weight   int               `json:"weight,omitempty"`
	Tags     map[string]string `json:"tags,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Config 注册器配置
type Config struct {
	// etcd配置
	EtcdEndpoints    []string      // etcd节点列表
	DialTimeout      time.Duration // 连接超时
	AutoSyncInterval time.Duration // 自动同步集群成员间隔

	// 服务配置
	ServiceName string // 服务名称
	TTL         int64  // 租约TTL（秒）

	// 认证配置（可选）
	Username string // etcd用户名
	Password string // etcd密码
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		DialTimeout:      5 * time.Second,
		AutoSyncInterval: 60 * time.Second,
		TTL:              60,
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	if len(c.EtcdEndpoints) == 0 {
		return ErrEmptyEndpoints
	}
	if c.ServiceName == "" {
		return ErrEmptyServiceName
	}
	if c.TTL < 10 {
		return ErrInvalidTTL
	}
	return nil
}
