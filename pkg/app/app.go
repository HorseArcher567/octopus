package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/HorseArcher567/octopus/pkg/config"
	"github.com/HorseArcher567/octopus/pkg/logger"
	"github.com/HorseArcher567/octopus/pkg/rpc"
	"google.golang.org/grpc"
)

// Config 是应用级配置，聚合日志与 RPC 服务配置。
type Config struct {
	Logger logger.Config    `yaml:"logger" json:"logger" toml:"logger"`
	Server rpc.ServerConfig `yaml:"server" json:"server" toml:"server"`
}

// BeforeRunHook 在服务启动前执行，如果返回错误将中止启动流程。
type BeforeRunHook func(ctx context.Context, a *App) error

// ShutdownHook 在服务关闭阶段执行，即使返回错误也会继续执行后续 Hook。
type ShutdownHook func(ctx context.Context, a *App) error

// App 封装一个 gRPC 服务应用的完整生命周期。
type App struct {
	cfgPath string  // 配置文件路径
	cfg     *Config // 解析后的配置
	grpcOpt []grpc.ServerOption

	log       *slog.Logger
	logCloser io.Closer

	server *rpc.Server
	ctx    context.Context

	beforeRunHooks []BeforeRunHook
	shutdownHooks  []ShutdownHook
}

// New 创建一个新的 App 实例，会立即根据 Option 完成初始化。
func New(opts ...Option) *App {
	a := &App{
		cfgPath: "config.yaml",
		ctx:     context.Background(),
	}

	for _, opt := range opts {
		opt(a)
	}

	a.init()
	return a
}

// init 完成配置加载、日志初始化、RPC 服务器创建。
func (a *App) init() {
	// 1. 加载配置
	if a.cfg == nil {
		var cfg Config
		config.MustUnmarshalWithEnv(a.cfgPath, &cfg)
		a.cfg = &cfg
	}

	// 2. 初始化日志
	if a.log == nil {
		log, closer := logger.MustNew(a.cfg.Logger)
		a.log = log
		a.logCloser = closer
	}
	if a.log != nil {
		slog.SetDefault(a.log)
	}

	// 3. 创建带 logger 的根 context
	a.ctx = logger.WithContext(a.ctx, a.log)

	// 4. 创建 RPC 服务器
	a.server = rpc.NewServer(a.ctx, &a.cfg.Server, a.grpcOpt...)
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

// RegisterService 注册 gRPC 服务。
func (a *App) RegisterService(register func(s *grpc.Server)) *App {
	if a.server == nil {
		panic("app: server is not initialized")
	}
	a.server.RegisterService(register)
	return a
}

// Run 启动应用，阻塞直到服务停止。
// 执行顺序：
// 1) 运行 OnBeforeRun Hooks（任一出错则中止启动）；
// 2) 启动 RPC Server 并阻塞；
// 3) Server 停止后运行 OnShutdown Hooks。
func (a *App) Run() error {
	if a.server == nil {
		return fmt.Errorf("app: server is not initialized")
	}

	// 1. BeforeRun hooks
	if err := a.runBeforeRunHooks(); err != nil {
		return err
	}

	if a.log != nil {
		a.log.Info("starting app",
			"app_name", a.cfg.Server.AppName,
			"host", a.cfg.Server.Host,
			"port", a.cfg.Server.Port,
		)
	}

	// 2. 启动 RPC 服务器（内部会阻塞直到优雅关闭完成）
	err := a.server.Start()

	// 3. Shutdown hooks（无论 server.Start 是否报错，都尝试执行）
	shutdownErr := a.runShutdownHooks()

	// 关闭日志 writer（如果有）
	if a.logCloser != nil {
		_ = a.logCloser.Close()
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
