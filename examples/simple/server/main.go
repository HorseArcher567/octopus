package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HorseArcher567/octopus/pkg/logger"
	"github.com/HorseArcher567/octopus/pkg/registry"
)

func main() {
	// 0. 初始化日志
	log, _, _ := logger.New(logger.Config{
		Level:     "debug",
		Format:    "text",
		AddSource: true,
	})
	slog.SetDefault(log)

	// 1. 配置服务实例信息
	instance := &registry.ServiceInstance{
		Addr:    "127.0.0.1",
		Port:    50052,
		Version: "v1.0.0",
		Zone:    "zone-1",
		Weight:  100,
		Tags: map[string]string{
			"env": "production",
		},
	}

	// 2. 创建注册器配置
	cfg := registry.DefaultConfig()
	cfg.EtcdEndpoints = []string{"localhost:2379",
		"localhost:2381",
		"localhost:2383"}
	cfg.AppName = "user-service"

	// 3. 创建注册器
	ctx := context.Background()
	reg, err := registry.NewRegistry(ctx, cfg, instance)
	if err != nil {
		slog.Error("failed to create registry", "error", err)
		os.Exit(1)
	}
	defer reg.Close()

	// 4. 注册服务
	if err := reg.Register(context.Background()); err != nil {
		slog.Error("failed to register service", "error", err)
		os.Exit(1)
	}
	slog.Info("application registered successfully")

	// 5. 这里可以启动你的实际业务服务
	// 例如：启动gRPC服务器、HTTP服务器等
	slog.Info("application is running",
		"addr", instance.Addr,
		"port", instance.Port,
	)

	// 6. 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan

	slog.Info("shutting down gracefully")

	// 7. 注销服务
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := reg.Unregister(shutdownCtx); err != nil {
		slog.Error("failed to unregister", "error", err)
	}

	slog.Info("shutdown complete")
}
