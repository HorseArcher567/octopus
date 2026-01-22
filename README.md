# Octopus - etcd服务注册发现框架

[![Go Version](https://img.shields.io/badge/go-1.19+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Octopus是一个基于etcd的Go语言服务注册发现框架，提供了简单易用的API和完善的生产级特性。

## ✨ 特性

- 🚀 **简单易用** - 简洁的API设计，开箱即用
- 🔄 **自动重连** - 租约失效自动重注册，Watch断开自动重连
- 🎯 **gRPC集成** - 内置gRPC Resolver实现
- ⚙️ **Job/Task支持** - 一次性任务、定时任务、长期Worker，灵活执行
- 📊 **可观测性** - Prometheus指标和健康检查端点
- 🛡️ **生产就绪** - 经过充分测试，可直接用于生产环境
- 🔧 **灵活配置** - 支持多种配置选项和扩展

## 📦 安装

```bash
go get github.com/your-username/octopus
```

**依赖**：
- Go 1.19+
- etcd 3.5+

## 🚀 快速开始

### 方式 1: 使用代码生成器创建新服务（推荐）⭐

```bash
# 1. 安装代码生成工具
make install-cli

# 2. 创建新服务
octopus-cli new user-service --module=github.com/yourname/user-service

# 3. 进入项目
cd user-service

# 4. 生成 Proto 代码
make proto

# 5. 启动 etcd（如果还未启动）
# Docker 方式
docker run -d -p 2379:2379 --name etcd \
  quay.io/coreos/etcd:v3.5.0 \
  etcd --listen-client-urls http://0.0.0.0:2379 \
       --advertise-client-urls http://0.0.0.0:2379

# 6. 运行服务
make run
# ✅ 服务启动，自动注册到 etcd！
```

**生成的项目结构**：
```
user-service/
├── cmd/main.go              # 服务入口（已集成框架）
├── internal/
│   ├── config/config.go     # 配置定义
│   ├── logic/logic.go       # 业务逻辑（你只需编辑这里！）
│   └── server/server.go     # gRPC 服务
├── proto/user-service.proto # Proto 定义
├── etc/config.yaml          # 配置文件
└── Makefile                 # 构建脚本
```

**开发流程**：
1. 编辑 `proto/user-service.proto` 定义接口
2. 运行 `make proto` 生成代码  
3. 在 `internal/logic/logic.go` 实现业务逻辑
4. 运行 `make run` 启动服务

查看完整指南：📖 **[RPC 实现方案](docs/RPC_IMPLEMENTATION.md)**

---

### 方式 2: 手动使用框架

### 前置条件

确保 etcd 已启动：

```bash
# 使用 Homebrew (macOS)
brew install etcd
etcd

# 或使用 Docker
docker run -d -p 2379:2379 --name etcd \
  quay.io/coreos/etcd:v3.5.0 \
  etcd --listen-client-urls http://0.0.0.0:2379 \
       --advertise-client-urls http://0.0.0.0:2379
```

### 服务注册

```go
package main

import (
    "context"
    "octopus/pkg/registry"
)

func main() {
    // 创建服务实例信息
    instance := &registry.ServiceInstance{
        Addr:    "127.0.0.1",
        Port:    50051,
        Version: "v1.0.0",
    }

    // 创建配置
    cfg := registry.DefaultConfig()
    cfg.EtcdEndpoints = []string{"localhost:2379"}
    cfg.ServiceName = "user-service"
    cfg.InstanceID = "instance-001"

    // 创建注册器
    reg, _ := registry.NewRegistry(ctx, cfg, instance)
    
    // 注册服务
    reg.Register(context.Background())
    defer reg.Unregister(context.Background())
    
    // 运行你的服务...
}
```

### 服务发现

```go
package main

import (
    "context"
    "octopus/pkg/registry"
)

func main() {
    // 创建发现器
    discovery, _ := registry.NewDiscovery([]string{"localhost:2379"})
    
    // 监听服务变化
    discovery.Watch(context.Background(), "user-service")
    
    // 获取服务实例
    instances := discovery.GetInstances()
    for _, inst := range instances {
        // 使用实例地址进行调用
        println(inst.Addr, inst.Port)
    }
}
```

### gRPC集成

```go
import (
    "octopus/pkg/resolver"
    "google.golang.org/grpc"
    "google.golang.org/grpc/resolver"
)

func main() {
    // 注册etcd resolver
    builder := resolver.NewBuilder([]string{"localhost:2379"})
    resolver.Register(builder)
    
    // 使用etcd://前缀连接
    conn, _ := grpc.Dial(
        "etcd:///user-service",
        grpc.WithInsecure(),
        grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
    )
    defer conn.Close()
    
    // 创建客户端并调用...
}
```

### Job/Task 任务执行

Octopus 支持三种类型的任务执行：

#### 1. 一次性任务（Once）- 启动前执行

```go
package main

import (
    "context"
    "github.com/HorseArcher567/octopus/pkg/app"
    "github.com/HorseArcher567/octopus/pkg/config"
    "github.com/HorseArcher567/octopus/pkg/job"
)

func main() {
    var cfg AppConfig
    config.MustUnmarshal("config.yaml", &cfg)
    app.Init(&cfg.Framework)

    // 注册一次性任务（数据迁移、初始化等）
    app.RegisterJob("init-db", job.JobTypeOnce, func(ctx context.Context) error {
        // 初始化数据库
        return initDatabase()
    })

    app.Run() // 执行完任务后自动退出
}
```

#### 2. 定时任务（Cron）- 按计划执行

```go
// 注册定时任务（定时清理、报表生成等）
app.RegisterCronJob("cleanup", "0 * * * *", func(ctx context.Context) error {
    // 每小时执行一次清理
    return cleanupExpiredData()
})

app.RegisterCronJob("report", "0 0 * * *", func(ctx context.Context) error {
    // 每天 0 点生成报表
    return generateDailyReport()
})
```

#### 3. 长期 Worker - 持续运行

```go
// 注册 Worker 任务（消息队列消费者、监控等）
app.RegisterJob("mq-consumer", job.JobTypeWorker, func(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return nil // 优雅退出
        default:
            msg := consumeMessage()
            processMessage(msg)
        }
    }
})
```

#### 4. 混合模式 - Job + RPC + API 同时运行

```go
func main() {
    app.Init(&cfg.Framework)

    // 启动前执行初始化
    app.RegisterJob("init", job.JobTypeOnce, initResources)

    // 注册 RPC 服务
    app.RegisterRpcServices(func(s *grpc.Server) {
        pb.RegisterUserServer(s, &UserServer{})
    })

    // 注册 HTTP API
    app.RegisterApiRoutes(func(engine *api.Engine) {
        engine.GET("/health", healthHandler)
    })

    // 注册后台任务（与服务并行运行）
    app.RegisterCronJob("cleanup", "*/5 * * * *", cleanupTask)
    app.RegisterJob("worker", job.JobTypeWorker, backgroundWorker)

    app.Run() // 所有组件并发运行
}
```

**配置示例**：

```yaml
logger:
  level: info

job:
  enabled: true           # 启用 Job 功能
  concurrentLimit: 10     # 并发任务数限制

rpcServer:                # 可选：RPC 服务配置
  host: 0.0.0.0
  port: 50051

apiServer:                # 可选：HTTP API 配置
  host: 0.0.0.0
  port: 8080
```

查看完整示例：[examples/job-demo/](examples/job-demo/)

## 🔧 高级特性

### 自动重连和重注册

```go
// 配置自动重连参数
cfg := registry.DefaultConfig()
cfg.TTL = 10                    // 租约TTL
cfg.RetryInterval = time.Second // 重试间隔
cfg.MaxRetries = 3              // 最大重试次数

// 注册器会自动：
// 1. 租约过期时自动重注册
// 2. etcd连接断开时自动重连（指数退避）
// 3. Watch断开时自动恢复
```

### 优雅关闭

```go
// 创建可取消的context
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// 监听系统信号
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

go func() {
    <-sigChan
    // 注销服务
    reg.Unregister(ctx)
    cancel()
}()
```

### 负载均衡策略

```go
// Round Robin（轮询）
conn, _ := grpc.Dial(
    "etcd:///user-service",
    grpc.WithInsecure(),
    grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
)

// 更多策略: "pick_first", "round_robin"
```

## 📦 核心包说明

### pkg/app
**应用框架核心** - 统一管理 RPC Server、API Server 和 Job Runner 的生命周期。

### pkg/rpc
**gRPC 服务管理** - 提供 gRPC Server 和 Client 的封装，支持服务注册和发现。

### pkg/api
**HTTP API 服务** - 基于 Gin 的 HTTP 服务器，提供 RESTful API 支持。

### pkg/job
**任务执行框架** - 支持一次性任务、定时任务和长期 Worker 的统一管理。
详见：[examples/job-demo/README.md](examples/job-demo/README.md)

### pkg/registry
服务注册发现的核心实现，提供 Registry 和 Discovery 两个主要组件。

### pkg/resolver  
gRPC Resolver 实现，让 gRPC 客户端可以直接使用 `etcd:///service-name` 格式的地址。

### pkg/config
**灵活的配置管理包** - 支持 JSON/YAML/TOML，支持环境变量替换，类型安全的访问方法。
详见：[pkg/config/README.md](pkg/config/README.md)

### pkg/xlog
**结构化日志** - 提供上下文感知的日志记录，支持日志轮转和多种输出格式。

### pkg/mapstruct
**Map 到结构体的转换工具** - 提供灵活的类型转换，支持标签映射和嵌套结构。
详见：[pkg/mapstruct/README.md](pkg/mapstruct/README.md)

## 📚 完整文档

### 📖 [文档中心](docs/README.md)

访问完整的文档导航和学习路径

### 核心文档

- 🎯 **[RPC 实现方案](docs/RPC_IMPLEMENTATION.md)** - 快速创建新服务（推荐）
  - 一键生成完整项目（类似 go-zero）
  - 代码自动生成（Server/Logic/Model）
  - 标准项目结构和分层架构
  - 内置服务治理（注册、日志、监控）
  - 数据库集成和 CRUD 生成

- 🏗️ **[技术架构设计](docs/TECHNICAL_DESIGN.md)** - 系统架构和技术方案
  - 整体架构设计
  - 核心模块详细设计（Registry/Resolver/Config）
  - 关键技术实现（心跳、负载均衡、容错）
  - 数据模型和接口设计
  - 性能指标和方案对比

- 📘 **[gRPC 框架使用指南](docs/GRPC_FRAMEWORK.md)** - 完整的使用教程
  - 框架概述和快速开始
  - 完整示例和高级功能
  - 配置参考和最佳实践
  - 常见问题和故障排查

- 🚀 **[部署和运维指南](docs/DEPLOYMENT.md)** - 生产环境部署方案
  - Docker 和 Kubernetes 部署
  - etcd 集群配置和管理
  - 监控、日志和性能调优
  - 安全加固和灾难恢复

- ⚡ **[快速参考手册](docs/QUICK_REFERENCE.md)** - 常用 API 和配置速查
  - 服务注册/发现 API
  - gRPC 客户端配置
  - 配置管理速查
  - 命令行工具和错误码

### 组件文档

- [Config 配置管理](pkg/config/README.md) - 灵活的配置加载和管理
- [MapStruct 数据转换](pkg/mapstruct/README.md) - Map 到结构体转换

### 示例代码

查看 [examples/](examples/) 目录获取完整示例：
- **multi-service/** - 多服务 RPC + API 完整示例
- **job-demo/** - Job/Task 任务执行示例
  - **once-job/** - 一次性任务（数据迁移、初始化）
  - **cron-job/** - 定时任务（定时清理、报表生成）
  - **worker-job/** - 长期 Worker（消息队列消费者）
  - **mixed-mode/** - 混合模式（Job + RPC + API）
- **simple/** - 基础的注册发现示例
- **grpc/** - gRPC 集成示例
- **config/** - 配置管理示例

## 🏗️ 项目结构

```
octopus/
├── pkg/                    # 核心包
│   ├── app/               # 应用框架核心
│   ├── rpc/               # gRPC 服务管理
│   ├── api/               # HTTP API 服务
│   ├── job/               # 任务执行框架
│   ├── etcd/              # etcd 客户端封装
│   ├── xlog/              # 结构化日志
│   ├── config/            # 配置管理
│   └── mapstruct/         # Map结构体转换
├── examples/              # 示例代码
│   ├── multi-service/    # 多服务示例
│   ├── job-demo/         # Job 任务示例
│   ├── simple/           # 简单示例
│   ├── grpc/             # gRPC集成示例
│   └── config/           # 配置示例
├── cmd/                   # 命令行工具
│   └── octopus-cli/      # 代码生成器
├── docs/                  # 文档
│   ├── GRPC_FRAMEWORK.md # gRPC框架完整指南
│   └── DEPLOYMENT.md     # 部署运维指南
└── Makefile              # 构建脚本
```

## 🤝 贡献

欢迎贡献代码、报告问题或提出建议！

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 🙏 致谢

- [etcd](https://github.com/etcd-io/etcd) - 分布式键值存储
- [gRPC](https://grpc.io/) - 高性能RPC框架
- [Prometheus](https://prometheus.io/) - 监控和告警工具

