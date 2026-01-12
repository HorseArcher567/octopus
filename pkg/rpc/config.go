package rpc

import "time"

// ClientConfig 客户端配置
// 支持两种模式：
// 1. etcd 服务发现模式：配置 appName + etcdConfig（从 App 获取）
// 2. 直连模式：配置 endpoints（支持多个，自动负载均衡）
// 自动检测逻辑：如果 endpoints 存在且非空，使用直连模式；否则使用 etcd 服务发现模式
type ClientConfig struct {
	// AppName 目标应用名称（在服务发现系统中注册的名称，etcd 模式必需）
	AppName string `yaml:"appName" json:"appName" toml:"appName"`

	// Endpoints 直连模式的服务地址列表（直连模式必需，支持多个地址实现负载均衡）
	Endpoints []string `yaml:"endpoints" json:"endpoints" toml:"endpoints"`

	// EnableKeepalive 是否启用 keepalive
	EnableKeepalive bool `yaml:"enableKeepalive" json:"enableKeepalive" toml:"enableKeepalive"`

	// KeepaliveTime keepalive 时间间隔（秒，默认 10）
	KeepaliveTime time.Duration `yaml:"keepaliveTime" json:"keepaliveTime" toml:"keepaliveTime"`

	// KeepaliveTimeout keepalive 超时时间（秒，默认 3）
	KeepaliveTimeout time.Duration `yaml:"keepaliveTimeout" json:"keepaliveTimeout" toml:"keepaliveTimeout"`

	// PermitWithoutStream 是否允许在没有活跃流时发送 keepalive ping
	PermitWithoutStream bool `yaml:"permitWithoutStream" json:"permitWithoutStream" toml:"permitWithoutStream"`
}

// ServerConfig 服务端配置
type ServerConfig struct {
	// AppName 应用名称
	AppName string `yaml:"appName" json:"appName" toml:"appName"`

	// Host 监听地址（如 0.0.0.0, 127.0.0.1）
	Host string `yaml:"host" json:"host" toml:"host"`

	// Port 监听端口
	Port int `yaml:"port" json:"port" toml:"port"`

	// AdvertiseAddr 注册到 etcd 的地址（留空则自动获取本机 IP）
	AdvertiseAddr string `yaml:"advertiseAddr" json:"advertiseAddr" toml:"advertiseAddr"`

	// TTL 租约时间（秒，默认 60）
	TTL int64 `yaml:"ttl" json:"ttl" toml:"ttl"`

	// EnableReflection 是否启用反射（推荐开发/测试环境启用，便于 grpcurl/grpcui 调试）
	EnableReflection bool `yaml:"enableReflection" json:"enableReflection" toml:"enableReflection"`
}
