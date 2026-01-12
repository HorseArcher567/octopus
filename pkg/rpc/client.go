package rpc

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/HorseArcher567/octopus/pkg/etcd"
	"github.com/HorseArcher567/octopus/pkg/rpc/resolver"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	grpcresolver "google.golang.org/grpc/resolver"
)

var (
	registerOnce  sync.Once
	globalBuilder *resolver.EtcdResolverBuilder
)

// NewClient 创建 RPC 客户端
// etcdConfig 是 etcd 配置，如果为 nil 则使用默认配置
// 自动检测使用 etcd 服务发现模式或直连模式
// 如果配置了 endpoints，使用直连模式；否则使用 etcd 服务发现模式
func NewClient(ctx context.Context, etcdConfig *etcd.Config, config *ClientConfig, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	// 验证配置
	if err := validateConfig(config, etcdConfig); err != nil {
		return nil, err
	}

	// 自动检测模式：如果 endpoints 存在且非空，使用直连模式
	useDirectMode := len(config.Endpoints) > 0

	if useDirectMode {
		return newDirectClient(ctx, config, opts...)
	}

	return newEtcdClient(ctx, etcdConfig, config, opts...)
}

// validateConfig 验证配置的有效性
func validateConfig(config *ClientConfig, etcdConfig *etcd.Config) error {
	hasEndpoints := len(config.Endpoints) > 0
	cfg := etcdConfig
	if cfg == nil {
		cfg = etcd.Default()
	}
	hasEtcdConfig := !cfg.IsEmpty() && config.AppName != ""

	// 两种模式不能同时配置
	if hasEndpoints && hasEtcdConfig {
		return fmt.Errorf("cannot configure both endpoints (direct mode) and etcd discovery mode at the same time")
	}

	// 必须配置其中一种模式
	if !hasEndpoints && !hasEtcdConfig {
		return fmt.Errorf("must configure either endpoints (direct mode) or etcd config + appName (etcd mode)")
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
	}

	return nil
}

// newDirectClient 创建直连模式的客户端
func newDirectClient(ctx context.Context, config *ClientConfig, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	log := xlog.FromContext(ctx)

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

	// 使用 direct resolver，将多个 endpoints 通过 target（direct:///ip1:port1,ip2:port2）暴露给 gRPC
	directBuilder := resolver.NewDirectBuilder(log)
	opts = append(opts, grpc.WithResolvers(directBuilder))

	// 使用 direct scheme 构建目标地址：direct:///ip1:port1,ip2:port2
	rawEndpoints := strings.Join(config.Endpoints, ",")
	target := fmt.Sprintf("direct:///%s", rawEndpoints)

	log.Info("connecting to service via direct mode",
		"endpoints", config.Endpoints,
		"target", target,
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
func newEtcdClient(ctx context.Context, etcdConfig *etcd.Config, config *ClientConfig, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	log := xlog.FromContext(ctx)

	cfg := etcdConfig
	if cfg == nil {
		cfg = etcd.Default()
	}

	if cfg.IsEmpty() {
		return nil, fmt.Errorf("etcd config is required for etcd discovery mode")
	}

	// 注册 etcd resolver（全局只注册一次）
	registerOnce.Do(func() {
		globalBuilder = resolver.NewBuilder(ctx, cfg)
		grpcresolver.Register(globalBuilder)

		log.Info("etcd resolver registered",
			"etcd_endpoints", cfg.Endpoints,
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
		"etcd_endpoints", cfg.Endpoints,
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
	return NewClient(ctx, nil, config, opts...)
}

// MustNewClient 创建客户端，失败时 panic
func MustNewClient(ctx context.Context, etcdConfig *etcd.Config, config *ClientConfig, opts ...grpc.DialOption) *grpc.ClientConn {
	conn, err := NewClient(ctx, etcdConfig, config, opts...)
	if err != nil {
		panic(err)
	}
	return conn
}

// NewClientFromConfig 从配置文件中加载客户端配置并创建客户端
// cfg 是配置管理器实例，key 是配置的键名（如 "client"）
// etcdConfig 是 etcd 配置，如果为 nil 则使用默认配置
func NewClientFromConfig(ctx context.Context, etcdConfig *etcd.Config, cfg interface {
	UnmarshalKey(string, interface{}) error
}, key string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	var config ClientConfig
	if err := cfg.UnmarshalKey(key, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal client config from key %s: %w", key, err)
	}
	return NewClient(ctx, etcdConfig, &config, opts...)
}

// MustNewClientFromConfig 从配置文件中加载客户端配置并创建客户端，失败时 panic
func MustNewClientFromConfig(ctx context.Context, etcdConfig *etcd.Config, cfg interface {
	UnmarshalKey(string, interface{}) error
}, key string, opts ...grpc.DialOption) *grpc.ClientConn {
	conn, err := NewClientFromConfig(ctx, etcdConfig, cfg, key, opts...)
	if err != nil {
		panic(err)
	}
	return conn
}
