package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/HorseArcher567/octopus/pkg/api"
	"github.com/HorseArcher567/octopus/pkg/etcd"
	"github.com/HorseArcher567/octopus/pkg/rpc"
	"github.com/HorseArcher567/octopus/pkg/xlog"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)



// BeforeRunHook 在服务启动前执行，如果返回错误将中止启动流程。
type BeforeRunHook func(ctx context.Context, a *App) error

// ShutdownHook 在服务关闭阶段执行，即使返回错误也会继续执行后续 Hook。
type ShutdownHook func(ctx context.Context, a *App) error

// App 封装一个同时托管 gRPC 与 HTTP API 的应用生命周期。
type App struct {
	ctx context.Context

	log *xlog.Logger // 封装的 Logger，管理资源生命周期

	rpcServer *rpc.Server
	apiServer *api.Server

	beforeRunHooks []BeforeRunHook
	shutdownHooks  []ShutdownHook
}

// New 创建一个新的 App 实例，会立即根据框架配置完成初始化。
// cfg 是框架配置，用户应该在外部加载自己的配置（嵌入 app.Framework），然后提取 Framework 部分传入。
//
// 示例：
//
//	type AppConfig struct {
//	    app.Framework
//	    Database struct { ... } `yaml:"database"`
//	}
//
//	var cfg AppConfig
//	config.MustUnmarshal("config.yaml", &cfg)
//	application := app.New(&cfg.Framework)
func New(cfg *Framework) *App {
	if cfg == nil {
		panic("app: framework config cannot be nil")
	}

	a := &App{
		ctx: context.Background(),
	}

	// 1. 初始化日志
	if cfg.LoggerCfg != nil {
		a.log = xlog.MustNew(*cfg.LoggerCfg)
		if a.log != nil {
			slog.SetDefault(a.log.Logger)
		}
	} else {
		// 使用默认的 slog，创建一个包装器（stdout 输出，无需关闭）
		a.log = xlog.MustNew(xlog.Config{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		})
	}

	// 2. 创建带 logger 的根 context
	a.ctx = xlog.WithContext(a.ctx, a.log.Logger)

	// 3. 设置 etcd 默认配置
	if cfg.EtcdCfg != nil {
		etcd.SetDefault(cfg.EtcdCfg)
	}

	// 4. 创建 RPC 服务器（如果配置了）
	if cfg.RpcSvrCfg != nil && cfg.RpcSvrCfg.Port > 0 {
		a.rpcServer = rpc.NewServer(a.ctx, cfg.RpcSvrCfg, cfg.EtcdCfg)
	}

	// 5. 创建 HTTP API 服务器（如果配置了）
	if cfg.ApiSvrCfg != nil && cfg.ApiSvrCfg.Port > 0 {
		a.apiServer = api.NewServer(a.ctx, cfg.ApiSvrCfg)
	}

	return a
}

// OnBeforeRun 注册在 Run 之前执行的 Hook。
// 按注册顺序执行，遇到第一个错误将中止启动流程。
func (a *App) OnBeforeRun(h BeforeRunHook) *App {
	if h != nil {
		a.beforeRunHooks = append(a.beforeRunHooks, h)
	}
	return a
}

// OnShutdown 注册在服务关闭阶段执行的 Hook。
// 即使某个 Hook 返回错误，也会继续执行后续 Hook。
func (a *App) OnShutdown(h ShutdownHook) *App {
	if h != nil {
		a.shutdownHooks = append(a.shutdownHooks, h)
	}
	return a
}

// RegisterRpcService 注册 gRPC 服务。
func (a *App) RegisterRpcService(register func(s *grpc.Server)) *App {
	if a.rpcServer == nil {
		panic("app: rpc server is not initialized (check RpcServer config)")
	}
	a.rpcServer.RegisterService(register)
	return a
}

// RegisterApiRoutes 注册 HTTP API 路由。
// 通常在 main 中调用，通过传入函数在 gin.Engine 上注册路由。
func (a *App) RegisterApiRoutes(register func(engine *api.Engine)) *App {
	if a.apiServer == nil {
		panic("app: api server is not initialized (check ApiServer config)")
	}
	if register != nil {
		register(a.apiServer.Engine())
	}
	return a
}

// Run 启动应用，阻塞直到所有已启用的服务停止。
// 执行顺序：
// 1) 运行 OnBeforeRun Hooks（任一出错则中止启动）；
// 2) 并发启动 RPC Server 和 HTTP API Server（根据配置决定是否启用）；
// 3) 所有服务停止后运行 OnShutdown Hooks。
func (a *App) Run() error {
	if a.rpcServer == nil && a.apiServer == nil {
		return fmt.Errorf("app: no server is initialized, check RpcServer/ApiServer config")
	}

	// 1. BeforeRun hooks
	if err := a.runBeforeRunHooks(); err != nil {
		return err
	}

	// 2. 并发启动服务
	var g errgroup.Group

	if a.rpcServer != nil {
		srv := a.rpcServer
		g.Go(func() error {
			return srv.Start()
		})
	}

	if a.apiServer != nil {
		httpSrv := a.apiServer
		g.Go(func() error {
			return httpSrv.Start()
		})
	}

	err := g.Wait()

	// 3. Shutdown hooks（无论服务是否报错，都尝试执行）
	shutdownErr := a.runShutdownHooks()

	// 关闭日志资源（如果有）
	if a.log != nil {
		_ = a.log.Close()
	}

	// 将 server.Start 的错误作为主错误返回，如果没有则返回 shutdown 错误信息。
	if err != nil {
		return err
	}
	return shutdownErr
}

func (a *App) runBeforeRunHooks() error {
	ctx := a.ctx
	for i, h := range a.beforeRunHooks {
		if h == nil {
			continue
		}
		if err := h(ctx, a); err != nil {
			if a.log != nil {
				a.log.Error("before run hook failed",
					"index", i,
					"error", err,
				)
			}
			return err
		}
	}
	return nil
}

func (a *App) runShutdownHooks() error {
	if len(a.shutdownHooks) == 0 {
		return nil
	}

	// 固定超时时间，后续可扩展到从配置读取。
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var firstErr error
	for i, h := range a.shutdownHooks {
		if h == nil {
			continue
		}
		if err := h(ctx, a); err != nil {
			if a.log != nil {
				a.log.Error("shutdown hook failed",
					"index", i,
					"error", err,
				)
			}
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}
