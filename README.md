# Octopus

Octopus 是一个面向 Go 服务的轻量框架，聚焦三件事：

- 统一应用生命周期（RPC / HTTP / Job）
- etcd 服务发现与 gRPC 客户端接入
- 可组合的模块化启动方式（`Module + Runtime`）

## 安装

```bash
go get github.com/HorseArcher567/octopus
```

## 快速开始

### 1. 运行示例服务

```bash
cd examples/multi-service/server
cp .env.example .env
export $(grep -v '^#' .env | xargs)
go run ./cmd/server -config configs/config.dev.yaml
```

### 2. 运行示例客户端

```bash
go run ./examples/multi-service/client \
  -config examples/multi-service/client/config.yaml \
  -target etcd:///multi-service-demo \
  -api http://127.0.0.1:8090/hello
```

## 核心抽象

```go
type Module interface {
    ID() string
    Init(ctx context.Context, rt Runtime) error
    Close(ctx context.Context) error
}

type DependedModule interface {
    DependsOn() []string
}
```

- `ID`: 模块唯一标识
- `Init`: 启动阶段初始化（按依赖拓扑顺序）
- `Close`: 关闭阶段回收（逆序执行）
- `DependsOn`: 可选依赖声明

## Runtime 能力

模块通过 `Runtime` 获取受限能力，不直接依赖 `*App`：

- `Logger()`
- `MySQL(name)` / `Redis(name)`
- `RegisterRPC(...)` / `RegisterHTTP(...)`
- `AddJob(...)`
- `NewRPCClient(target)`

### RPC 客户端复用

- `NewRPCClient(target)` 会按 `target` 复用连接。
- 相同 `target` 多次调用返回同一个 `*grpc.ClientConn`。
- 框架在 `App` 关闭阶段自动调用 `CloseRpcClients()` 统一释放连接。
- 如果业务在运行期主动调用 `CloseRpcClients()`，后续再调用 `NewRPCClient(target)` 会创建新连接。

## 应用入口示例

```go
infra := bootstrap.NewInfraModule()
app.MustRun(configFile, []app.Module{
    infra,
    bootstrap.NewRPCModule(infra),
    bootstrap.NewAPIModule(),
})
```

其中：

- `infra` 模块初始化数据库和仓储
- `rpc` 模块依赖 `infra`，注册 gRPC 服务
- `api` 模块注册 HTTP 路由

## 项目结构

```text
octopus/
├── pkg/
│   ├── app/        # 生命周期与模块运行时
│   ├── rpc/        # gRPC server/client + etcd 注册发现
│   ├── api/        # HTTP API server (gin)
│   ├── resource/   # MySQL/Redis 资源管理
│   ├── config/     # 配置加载
│   └── xlog/       # 日志
├── examples/
│   └── multi-service/
│       ├── server/
│       └── client/
└── cmd/
    └── octopus-cli/
```

## 测试

```bash
GOCACHE=/tmp/go-build go test ./...
```

## 许可证

MIT
