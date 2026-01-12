package rpc

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HorseArcher567/octopus/pkg/etcd"
	"github.com/HorseArcher567/octopus/pkg/rpc/registry"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server RPC 服务器封装
type Server struct {
	config     *ServerConfig
	etcdConfig *etcd.Config
	grpcServer *grpc.Server
	registry   *registry.Registry
	listener   net.Listener
	log        *slog.Logger
}

// NewServer 创建 RPC 服务器
// etcdConfig 是 etcd 配置，如果为 nil 则使用默认配置
// 从 context 中获取 logger，如果没有则使用 slog.Default()
func NewServer(ctx context.Context, config *ServerConfig, etcdConfig *etcd.Config, opts ...grpc.ServerOption) *Server {
	log := xlog.FromContext(ctx).With("component", "rpc.server", "appName", config.AppName)
	return &Server{
		config:     config,
		etcdConfig: etcdConfig,
		grpcServer: grpc.NewServer(opts...),
		log:        log,
	}
}

// RegisterService 注册 gRPC 服务（支持多次调用以注册多个服务）
func (s *Server) RegisterService(registerFunc func(*grpc.Server)) {
	registerFunc(s.grpcServer)
}

// Start 启动服务器
func (s *Server) Start() error {
	// 1. 启用反射（如果配置了）
	if s.config.EnableReflection {
		reflection.Register(s.grpcServer)
		s.log.Info("grpc reflection enabled")
	}

	// 2. 创建监听器
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	s.listener = lis

	// 3. 注册服务到 etcd（如果配置了）
	cfg := s.etcdConfig
	if cfg == nil {
		cfg = etcd.Default()
	}
	if !cfg.IsEmpty() {
		if err := s.registerToEtcd(cfg); err != nil {
			return fmt.Errorf("failed to register service: %w", err)
		}
	}

	// 4. 启动 gRPC 服务器
	s.log.Info("starting rpc server", "addr", addr)
	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			s.log.Error("server stopped", "error", err)
		}
	}()

	// 5. 等待退出信号
	s.waitForShutdown()

	return nil
}

// registerToEtcd 注册服务到 etcd
func (s *Server) registerToEtcd(etcdConfig *etcd.Config) error {
	// 确定注册地址
	advertiseAddr := s.config.AdvertiseAddr

	instance := &registry.ServiceInstance{
		Addr: advertiseAddr,
		Port: s.config.Port,
	}

	regCfg := registry.DefaultConfig()
	regCfg.AppName = s.config.AppName
	if s.config.TTL > 0 {
		regCfg.TTL = s.config.TTL
	}

	// 创建带 logger 的 context
	ctx := xlog.WithContext(context.Background(), s.log)
	reg, err := registry.NewRegistry(ctx, etcdConfig, regCfg, instance)
	if err != nil {
		return err
	}

	if err := reg.Register(context.Background()); err != nil {
		return err
	}

	s.registry = reg
	s.log.Info("application registered to etcd",
		"advertiseAddr", advertiseAddr,
		"port", s.config.Port,
	)
	return nil
}

// waitForShutdown 等待关闭信号
func (s *Server) waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan

	s.log.Info("shutting down gracefully")

	// 1. 注销服务
	if s.registry != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.registry.Unregister(ctx); err != nil {
			s.log.Error("failed to unregister", "error", err)
		}
	}

	// 2. 停止接受新请求
	s.grpcServer.GracefulStop()

	s.log.Info("shutdown complete")
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
