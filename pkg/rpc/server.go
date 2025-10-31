package rpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"octopus/pkg/registry"
)

// ServerConfig 服务端配置
type ServerConfig struct {
	Name     string   // 服务名称
	Host     string   // 监听地址
	Port     int      // 监听端口
	EtcdAddr []string // etcd 地址
	TTL      int64    // 租约时间（秒）

	// 可选配置
	EnableReflection bool // 是否启用反射（开发环境使用）
	EnableHealth     bool // 是否启用健康检查
}

// Server RPC 服务器封装
type Server struct {
	config     *ServerConfig
	grpcServer *grpc.Server
	registry   *registry.Registry
	listener   net.Listener
}

// NewServer 创建 RPC 服务器
func NewServer(config *ServerConfig, opts ...grpc.ServerOption) *Server {
	return &Server{
		config:     config,
		grpcServer: grpc.NewServer(opts...),
	}
}

// RegisterService 注册 gRPC 服务
func (s *Server) RegisterService(registerFunc func(*grpc.Server)) {
	registerFunc(s.grpcServer)

	// 启用健康检查（默认开启）
	if s.config.EnableHealth {
		healthServer := health.NewServer()
		healthServer.SetServingStatus(s.config.Name, grpc_health_v1.HealthCheckResponse_SERVING)
		grpc_health_v1.RegisterHealthServer(s.grpcServer, healthServer)
	}

	// 启用反射（方便使用 grpcurl 等工具调试）
	if s.config.EnableReflection {
		reflection.Register(s.grpcServer)
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	// 1. 创建监听器
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	s.listener = lis

	// 2. 注册服务到 etcd
	if len(s.config.EtcdAddr) > 0 {
		if err := s.registerToEtcd(); err != nil {
			return fmt.Errorf("failed to register service: %w", err)
		}
	}

	// 3. 启动 gRPC 服务器
	log.Printf("Starting RPC server at %s", addr)
	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			log.Printf("Server stopped: %v", err)
		}
	}()

	// 4. 等待退出信号
	s.waitForShutdown()

	return nil
}

// registerToEtcd 注册服务到 etcd
func (s *Server) registerToEtcd() error {
	instance := &registry.ServiceInstance{
		Addr: s.config.Host,
		Port: s.config.Port,
	}

	cfg := registry.DefaultConfig()
	cfg.EtcdEndpoints = s.config.EtcdAddr
	cfg.ServiceName = s.config.Name
	cfg.InstanceID = s.getInstanceID()
	if s.config.TTL > 0 {
		cfg.TTL = s.config.TTL
	}

	reg, err := registry.NewRegistry(cfg, instance)
	if err != nil {
		return err
	}

	if err := reg.Register(context.Background()); err != nil {
		return err
	}

	s.registry = reg
	log.Printf("Service registered: %s (instance: %s)", s.config.Name, cfg.InstanceID)
	return nil
}

// waitForShutdown 等待关闭信号
func (s *Server) waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan

	log.Println("Shutting down gracefully...")

	// 1. 注销服务
	if s.registry != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.registry.Unregister(ctx); err != nil {
			log.Printf("Failed to unregister: %v", err)
		}
	}

	// 2. 停止接受新请求
	s.grpcServer.GracefulStop()

	log.Println("Shutdown complete")
}

// Stop 停止服务器
func (s *Server) Stop() {
	if s.registry != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		s.registry.Unregister(ctx)
	}
	s.grpcServer.GracefulStop()
}

// getInstanceID 获取实例 ID
func (s *Server) getInstanceID() string {
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Sprintf("%s-%d", s.config.Name, time.Now().Unix())
	}
	return hostname
}
