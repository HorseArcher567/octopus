package main

import (
	"context"
	"log"
	"time"

	"github.com/HorseArcher567/octopus/pkg/registry"
)

func main() {
	ctx := context.Background()

	// 1. 创建服务发现器
	discovery, err := registry.NewDiscovery(ctx, []string{"localhost:2379",
		"localhost:2381",
		"localhost:2383"})
	if err != nil {
		log.Fatalf("Failed to create discovery: %v", err)
	}
	defer discovery.Close()

	// 2. 监听服务变化
	if err := discovery.Watch(ctx, "user-service"); err != nil {
		log.Fatalf("Failed to watch service: %v", err)
	}

	log.Println("Watching service: user-service")

	// 3. 定期获取服务实例列表
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		instances := discovery.GetInstances()
		log.Printf("Found %d instance(s):", len(instances))
		for i, inst := range instances {
			log.Printf("  [%d] %s:%d (version: %s, zone: %s, weight: %d)",
				i+1, inst.Addr, inst.Port, inst.Version, inst.Zone, inst.Weight)
		}

		// 这里可以使用实例地址进行实际的服务调用
		// 例如：建立gRPC连接、发起HTTP请求等
	}
}
