package registry

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

// Config 注册器配置（简化版，etcd 配置由 etcd 包统一管理）
type Config struct {
	AppName string // 应用名称
	TTL     int64  // 租约TTL（秒）
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		TTL: 60,
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.AppName == "" {
		return ErrEmptyAppName
	}
	if c.TTL < 10 {
		return ErrInvalidTTL
	}
	return nil
}
