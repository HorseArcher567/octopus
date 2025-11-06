package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HorseArcher567/octopus/pkg/registry"
)

func main() {
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
	reg, err := registry.NewRegistry(cfg, instance)
	if err != nil {
		log.Fatalf("Failed to create registry: %v", err)
	}
	defer reg.Close()

	// 4. 注册服务
	if err := reg.Register(context.Background()); err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}
	log.Println("Application registered successfully")

	// 5. 这里可以启动你的实际业务服务
	// 例如：启动gRPC服务器、HTTP服务器等
	log.Println("Application is running...")

	// 6. 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan

	log.Println("Shutting down gracefully...")

	// 7. 注销服务
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := reg.Unregister(ctx); err != nil {
		log.Printf("Failed to unregister: %v", err)
	}

	log.Println("Shutdown complete")
}
