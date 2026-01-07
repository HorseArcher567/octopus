package rpc

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/HorseArcher567/octopus/pkg/logger"
	"github.com/HorseArcher567/octopus/pkg/rpc/internal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/resolver"
)

var (
	registerOnce  sync.Once
	globalBuilder *internal.EtcdResolverBuilder
)

// ClientConfig 客户端配置
// 支持两种模式：
// 1. etcd 服务发现模式：配置 appName + etcdAddr
// 2. 直连模式：配置 endpoints（支持多个，自动负载均衡）
// 自动检测逻辑：如果 endpoints 存在且非空，使用直连模式；否则使用 etcd 服务发现模式
type ClientConfig struct {
	// AppName 目标应用名称（在服务发现系统中注册的名称，etcd 模式必需）
	AppName string `yaml:"appName" json:"appName" toml:"appName"`

	// EtcdAddr etcd 地址列表（etcd 模式必需）
	EtcdAddr []string `yaml:"etcdAddr" json:"etcdAddr" toml:"etcdAddr"`

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

// NewClient 创建 RPC 客户端
// 自动检测使用 etcd 服务发现模式或直连模式
// 如果配置了 endpoints，使用直连模式；否则使用 etcd 服务发现模式
func NewClient(ctx context.Context, config *ClientConfig, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	// 验证配置
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	// 自动检测模式：如果 endpoints 存在且非空，使用直连模式
	useDirectMode := len(config.Endpoints) > 0

	if useDirectMode {
		return newDirectClient(ctx, config, opts...)
	}

	return newEtcdClient(ctx, config, opts...)
}

// validateConfig 验证配置的有效性
func validateConfig(config *ClientConfig) error {
	hasEndpoints := len(config.Endpoints) > 0
	hasEtcdConfig := config.AppName != "" && len(config.EtcdAddr) > 0

	// 两种模式不能同时配置
	if hasEndpoints && hasEtcdConfig {
		return fmt.Errorf("cannot configure both endpoints (direct mode) and appName+etcdAddr (etcd mode) at the same time")
	}

	// 必须配置其中一种模式
	if !hasEndpoints && !hasEtcdConfig {
		return fmt.Errorf("must configure either endpoints (direct mode) or appName+etcdAddr (etcd mode)")
	}

	// 直连模式验证
	if hasEndpoints {
		for i, ep := range config.Endpoints {
			if ep == "" {
				return fmt.Errorf("endpoints[%d] cannot be empty", i)
			}
		}
	}

	// etcd 模式验证
	if hasEtcdConfig {
		if config.AppName == "" {
			return fmt.Errorf("appName is required for etcd mode")
		}
		if len(config.EtcdAddr) == 0 {
			return fmt.Errorf("etcdAddr is required for etcd mode")
		}
		for i, addr := range config.EtcdAddr {
			if addr == "" {
				return fmt.Errorf("etcdAddr[%d] cannot be empty", i)
			}
		}
	}

	return nil
}

// newDirectClient 创建直连模式的客户端
func newDirectClient(ctx context.Context, config *ClientConfig, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	log := logger.FromContext(ctx)

	// 构建默认选项
	defaultOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	}

	// Keepalive 配置
	if config.EnableKeepalive {
		kaParams := keepalive.ClientParameters{
			Time:                config.KeepaliveTime,
			Timeout:             config.KeepaliveTimeout,
			PermitWithoutStream: config.PermitWithoutStream,
		}
		if kaParams.Time == 0 {
			kaParams.Time = 10 * time.Second
		}
		if kaParams.Timeout == 0 {
			kaParams.Timeout = 3 * time.Second
		}
		defaultOpts = append(defaultOpts, grpc.WithKeepaliveParams(kaParams))
	}

	// 合并用户自定义选项
	opts = append(defaultOpts, opts...)

	// 构建目标地址：多个 endpoints 用逗号分隔，gRPC 会自动处理负载均衡
	target := strings.Join(config.Endpoints, ",")

	log.Info("connecting to service via direct mode",
		"endpoints", config.Endpoints,
	)

	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		log.Error("failed to create connection",
			"error", err,
			"endpoints", config.Endpoints,
		)
		return nil, fmt.Errorf("failed to connect to endpoints %v: %w", config.Endpoints, err)
	}

	return conn, nil
}

// newEtcdClient 创建 etcd 服务发现模式的客户端
func newEtcdClient(ctx context.Context, config *ClientConfig, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	log := logger.FromContext(ctx)

	// 注册 etcd resolver（全局只注册一次）
	registerOnce.Do(func() {
		globalBuilder = internal.NewBuilder(ctx, config.EtcdAddr)
		resolver.Register(globalBuilder)

		log.Info("etcd resolver registered",
			"etcd_endpoints", config.EtcdAddr,
		)
	})

	// 构建默认选项
	defaultOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	}

	// Keepalive 配置
	if config.EnableKeepalive {
		kaParams := keepalive.ClientParameters{
			Time:                config.KeepaliveTime,
			Timeout:             config.KeepaliveTimeout,
			PermitWithoutStream: config.PermitWithoutStream,
		}
		if kaParams.Time == 0 {
			kaParams.Time = 10 * time.Second
		}
		if kaParams.Timeout == 0 {
			kaParams.Timeout = 3 * time.Second
		}
		defaultOpts = append(defaultOpts, grpc.WithKeepaliveParams(kaParams))
	}

	// 合并用户自定义选项
	opts = append(defaultOpts, opts...)

	// 创建连接
	target := fmt.Sprintf("etcd:///%s", config.AppName)
	log.Info("connecting to service via etcd discovery",
		"target_app", config.AppName,
		"etcd_endpoints", config.EtcdAddr,
	)

	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		log.Error("failed to create connection",
			"error", err,
			"target_app", config.AppName,
		)
		return nil, fmt.Errorf("failed to connect to %s: %w", config.AppName, err)
	}

	return conn, nil
}

// NewClientWithEndpoints 创建 RPC 客户端（直连模式，用于开发测试）
// 这是一个便捷函数，内部调用 NewClient
func NewClientWithEndpoints(ctx context.Context, endpoints []string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	config := &ClientConfig{
		Endpoints: endpoints,
	}
	return NewClient(ctx, config, opts...)
}

// MustNewClient 创建客户端，失败时 panic
func MustNewClient(ctx context.Context, config *ClientConfig, opts ...grpc.DialOption) *grpc.ClientConn {
	conn, err := NewClient(ctx, config, opts...)
	if err != nil {
		panic(err)
	}
	return conn
}

// NewClientFromConfig 从配置文件中加载客户端配置并创建客户端
// cfg 是配置管理器实例，key 是配置的键名（如 "client"）
func NewClientFromConfig(ctx context.Context, cfg interface {
	UnmarshalKey(string, interface{}) error
}, key string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	var config ClientConfig
	if err := cfg.UnmarshalKey(key, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal client config from key %s: %w", key, err)
	}
	return NewClient(ctx, &config, opts...)
}

// MustNewClientFromConfig 从配置文件中加载客户端配置并创建客户端，失败时 panic
func MustNewClientFromConfig(ctx context.Context, cfg interface {
	UnmarshalKey(string, interface{}) error
}, key string, opts ...grpc.DialOption) *grpc.ClientConn {
	conn, err := NewClientFromConfig(ctx, cfg, key, opts...)
	if err != nil {
		panic(err)
	}
	return conn
}
