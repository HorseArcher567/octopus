package rpc

import (
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	"octopus/pkg/resolver"
)

// ClientConfig 客户端配置
type ClientConfig struct {
	ServiceName string   // 服务名称
	EtcdAddr    []string // etcd 地址
	Timeout     int      // 超时时间（毫秒）

	// 可选配置
	EnableKeepalive bool // 是否启用 keepalive
}

// NewClient 创建 RPC 客户端（自动服务发现）
func NewClient(config *ClientConfig, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	// 1. 注册 etcd resolver
	builder := resolver.NewBuilder(config.EtcdAddr)
	resolver.Register(builder)

	// 2. 构建默认选项
	defaultOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	}

	// 超时配置
	if config.Timeout > 0 {
		timeout := time.Duration(config.Timeout) * time.Millisecond
		defaultOpts = append(defaultOpts, grpc.WithTimeout(timeout))
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
	target := fmt.Sprintf("etcd:///%s", config.ServiceName)
	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", config.ServiceName, err)
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
	conn, err := grpc.Dial(endpoints[0], opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", endpoints[0], err)
	}

	return conn, nil
}

// MustNewClient 创建客户端，失败时 panic
func MustNewClient(config *ClientConfig, opts ...grpc.DialOption) *grpc.ClientConn {
	conn, err := NewClient(config, opts...)
	if err != nil {
		panic(err)
	}
	return conn
}
