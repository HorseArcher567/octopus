package rpc

import (
	"context"
	"fmt"
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
type ClientConfig struct {
	AppName  string   // 目标应用名称（在服务发现系统中注册的名称）
	EtcdAddr []string // etcd 地址

	// 可选配置
	EnableKeepalive bool // 是否启用 keepalive
}

// NewClient 创建 RPC 客户端（自动服务发现）
func NewClient(ctx context.Context, config *ClientConfig, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	// 1. 注册 etcd resolver（全局只注册一次）
	registerOnce.Do(func() {
		globalBuilder = internal.NewBuilder(ctx, config.EtcdAddr)
		resolver.Register(globalBuilder)

		// 从 builder 中获取 logger 用于后续日志
		log := logger.FromContext(ctx)
		log.Info("etcd resolver registered",
			"etcd_endpoints", config.EtcdAddr,
		)
	})

	// 2. 构建默认选项
	defaultOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	}

	// Keepalive 配置
	if config.EnableKeepalive {
		kaParams := keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}
		defaultOpts = append(defaultOpts, grpc.WithKeepaliveParams(kaParams))
	}

	// 合并用户自定义选项
	opts = append(defaultOpts, opts...)

	// 3. 创建连接
	target := fmt.Sprintf("etcd:///%s", config.AppName)
	log := logger.FromContext(ctx)
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
func NewClientWithEndpoints(endpoints []string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("endpoints cannot be empty")
	}

	// 构建默认选项
	defaultOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// 合并用户自定义选项
	opts = append(defaultOpts, opts...)

	// 创建连接（直连第一个地址）
	conn, err := grpc.NewClient(endpoints[0], opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", endpoints[0], err)
	}

	return conn, nil
}

// MustNewClient 创建客户端，失败时 panic
func MustNewClient(ctx context.Context, config *ClientConfig, opts ...grpc.DialOption) *grpc.ClientConn {
	conn, err := NewClient(ctx, config, opts...)
	if err != nil {
		panic(err)
	}
	return conn
}
