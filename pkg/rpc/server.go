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

	"github.com/HorseArcher567/octopus/pkg/logger"
	"github.com/HorseArcher567/octopus/pkg/netutil"
	"github.com/HorseArcher567/octopus/pkg/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// ServerConfig 服务端配置
type ServerConfig struct {
	AppName          string   `yaml:"app_name" json:"app_name" toml:"app_name"`                            // 应用名称
	Host             string   `yaml:"host" json:"host" toml:"host"`                                        // 监听地址（如 0.0.0.0, 127.0.0.1）
	Port             int      `yaml:"port" json:"port" toml:"port"`                                        // 监听端口
	AdvertiseAddr    string   `yaml:"advertise_addr" json:"advertise_addr" toml:"advertise_addr"`          // 注册到 etcd 的地址（留空则自动获取本机 IP）
	EtcdAddr         []string `yaml:"etcd_addr" json:"etcd_addr" toml:"etcd_addr"`                         // etcd 地址（可选，留空则不注册到服务发现）
	TTL              int64    `yaml:"ttl" json:"ttl" toml:"ttl"`                                           // 租约时间（秒，默认 60）
	EnableReflection bool     `yaml:"enable_reflection" json:"enable_reflection" toml:"enable_reflection"` // 是否启用反射（推荐开发/测试环境启用，便于 grpcurl/grpcui 调试）
}

// Server RPC 服务器封装
type Server struct {
	config     *ServerConfig
	grpcServer *grpc.Server
	registry   *registry.Registry
	listener   net.Listener
	log        *slog.Logger
}

// NewServer 创建 RPC 服务器
// 从 context 中获取 logger，如果没有则使用 slog.Default()
func NewServer(ctx context.Context, config *ServerConfig, opts ...grpc.ServerOption) *Server {
	log := logger.FromContext(ctx).With("component", "rpc.server", "app_name", config.AppName)
	return &Server{
		config:     config,
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
	if len(s.config.EtcdAddr) > 0 {
		if err := s.registerToEtcd(); err != nil {
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
func (s *Server) registerToEtcd() error {
	// 确定注册地址
	advertiseAddr := s.config.AdvertiseAddr

	// 如果 advertise_addr 为空，检查 host 是否可用作注册地址
	if advertiseAddr == "" {
		// 如果 Host 不是回环地址且不是 0.0.0.0，则使用 Host
		if s.config.Host != "" && s.config.Host != "0.0.0.0" &&
			s.config.Host != "127.0.0.1" && s.config.Host != "localhost" {
			advertiseAddr = s.config.Host
		} else {
			// Host 无法被其他机器访问，必须配置 advertise_addr
			// 获取本机 IP 列表用于错误提示
			var ipHint string
			if ips, err := netutil.GetAllLocalIPs(); err == nil && len(ips) > 0 {
				ipHint = "\n\nDetected available IPs on this machine:\n"
				for i, ip := range ips {
					ipHint += fmt.Sprintf("  %d. %s\n", i+1, ip)
				}
				ipHint += "\nRecommended configuration:\n"
				ipHint += "  server:\n"
				ipHint += fmt.Sprintf("    host: %s\n", s.config.Host)
				ipHint += fmt.Sprintf("    advertise_addr: %s  # or use ${ADVERTISE_ADDR} for env variable\n", ips[0])
			} else {
				ipHint = "\n\nPlease manually configure advertise_addr in your config file."
			}

			return fmt.Errorf(
				"cannot register to etcd: host '%s' is not accessible from other machines\n"+
					"You must explicitly set 'advertise_addr' to an IP address that other services can reach.%s",
				s.config.Host, ipHint,
			)
		}
	}

	instance := &registry.ServiceInstance{
		Addr: advertiseAddr,
		Port: s.config.Port,
	}

	cfg := registry.DefaultConfig()
	cfg.EtcdEndpoints = s.config.EtcdAddr
	cfg.AppName = s.config.AppName
	if s.config.TTL > 0 {
		cfg.TTL = s.config.TTL
	}

	// 创建带 logger 的 context
	ctx := logger.WithContext(context.Background(), s.log)
	reg, err := registry.NewRegistry(ctx, cfg, instance)
	if err != nil {
		return err
	}

	if err := reg.Register(context.Background()); err != nil {
		return err
	}

	s.registry = reg
	s.log.Info("application registered to etcd",
		"advertise_addr", advertiseAddr,
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
