## `pkg/app` 使用说明

`app` 包为基于 gRPC 的服务提供统一的初始化与运行模式，封装了：

- 配置加载（支持环境变量替换）
- 日志初始化（基于 `pkg/logger` + `slog`）
- RPC 服务器创建（基于 `pkg/rpc`）
- 生命周期管理（BeforeRun / Shutdown Hook）

并采用类似 `slog` 的设计：既有**默认实例**（包级函数），也支持显式创建多个 `App` 实例。

---

### 核心概念

- `type Config struct { Logger logger.Config; Server rpc.ServerConfig }`
- `type App struct { ... }`：聚合配置、日志、RPC server、根 `context.Context` 与 Hook。
- Hook：
  - `type BeforeRunHook func(ctx context.Context, a *App) error`
  - `type ShutdownHook func(ctx context.Context, a *App) error`

执行顺序：

1. 初始化阶段（`Init` / `New`）：
   - 解析配置（默认 `config.yaml`）
   - 初始化 logger 并 `slog.SetDefault`
   - 创建带 logger 的根 context
   - 创建 `rpc.Server`
2. 运行阶段（`Run`）：
   - 按顺序执行 `BeforeRun` Hooks（任一失败则中止启动）
   - 启动 RPC Server（内部阻塞直到优雅关闭）
   - Server 停止后按顺序执行 `Shutdown` Hooks（即使失败也继续执行后续）

---

### 默认实例：推荐主路径

大多数应用只需要使用包级默认实例：

```go
func main() {
    app.Init(app.WithConfigFile("config.yaml"))

    app.OnBeforeRun(func(ctx context.Context, a *app.App) error {
        // 初始化数据库 / Redis / Kafka 等依赖
        return nil
    })

    app.OnShutdown(func(ctx context.Context, a *app.App) error {
        // 关闭资源 / flush 日志 等
        return nil
    })

    app.RegisterService(func(s *grpc.Server) {
        pb.RegisterUserServer(s, &UserServer{})
        pb.RegisterOrderServer(s, &OrderServer{})
        pb.RegisterProductServer(s, &ProductServer{})
    })

    if err := app.Run(); err != nil {
        slog.Error("failed to run app", "error", err)
    }
}
```

#### 错误处理策略

- `OnBeforeRun`：
  - 按注册顺序执行
  - **任一 Hook 返回 error 即中止后续 Hook 和启动流程**
  - 错误会记录到日志，并由 `Run` 返回
- `OnShutdown`：
  - 按注册顺序执行
  - **即使某个 Hook 返回 error，也会继续执行后续 Hook**
  - 所有错误都会记录到日志，`Run` 最终可能返回第一个错误

---

### 多实例 / 高级用法

如果需要在一个进程中管理多个 App，或在测试中手动控制实例，可以使用 `New`：

```go
func main() {
    a := app.New(app.WithConfigFile("config-a.yaml"))
    b := app.New(app.WithConfigFile("config-b.yaml"))

    a.OnBeforeRun(initADeps)
    b.OnBeforeRun(initBDeps)

    go func() {
        if err := a.Run(); err != nil {
            slog.Error("app A exit", "error", err)
        }
    }()

    if err := b.Run(); err != nil {
        slog.Error("app B exit", "error", err)
    }
}
```

---

### Option 一览

```go
type Option func(a *App)
```

- `WithConfigFile(path string)`：
  - 指定配置文件路径（默认 `config.yaml`）。
- `WithConfig(cfg *app.Config)`：
  - 直接提供配置对象，跳过文件加载。
- `WithLogger(log *slog.Logger, closer io.Closer)`：
  - 使用已有 logger 实例，可选指定 `closer` 在 `Run` 结束时自动关闭。
- `WithServerOptions(opts ...grpc.ServerOption)`：
  - 透传 gRPC Server 选项（拦截器、限流等）。

---

### 与现有示例的结合

`examples/multi-service/server/main.go` 已经演示了使用默认实例的模式：

- `app.Init(app.WithConfigFile("config.yaml"))` 负责：
  - 加载 `logger` 与 `server` 配置
  - 初始化日志并设置为 `slog` 默认 logger
  - 创建 RPC server
- 业务代码只需要：
  - 实现各 gRPC 服务（`UserServer` / `OrderServer` / `ProductServer`）
  - 在 `app.RegisterService` 中完成注册
  - 调用 `app.Run()` 启动服务

这样可以将各类应用的启动过程约束在统一的固定模式下，减少重复和出错点。 


