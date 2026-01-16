package app

import (
	"github.com/HorseArcher567/octopus/pkg/api"
	"google.golang.org/grpc"
)

// defaultApp 是类似 slog.Default() 的全局默认应用实例。
var defaultApp *App

// Init 初始化默认应用实例。
// 应在 main 函数启动阶段调用一次。
// framework 是框架配置，用户应该在外部加载自己的配置（嵌入 app.Framework），然后提取 Framework 部分传入。
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
//	app.Init(&cfg.Framework)
func Init(framework *Framework) {
	defaultApp = New(framework)
}

// Default 返回当前默认应用实例。
// 如果尚未调用 Init，则 panic。
func Default() *App {
	if defaultApp == nil {
		panic("app: defaultApp is not initialized, call app.Init() first")
	}
	return defaultApp
}

// OnBeforeRun 注册在 Run 之前执行的 Hook（默认实例）。
func OnBeforeRun(h BeforeRunHook) {
	Default().OnBeforeRun(h)
}

// OnShutdown 注册在关闭阶段执行的 Hook（默认实例）。
func OnShutdown(h ShutdownHook) {
	Default().OnShutdown(h)
}

// RegisterRpcServices 在默认应用上注册 gRPC 服务。
func RegisterRpcServices(register func(s *grpc.Server)) {
	Default().RegisterRpcServices(register)
}

// RegisterApiRoutes 在默认应用上注册 HTTP API 路由。
func RegisterApiRoutes(register func(engine *api.Engine)) {
	Default().RegisterApiRoutes(register)
}

// Run 启动默认应用
func Run() {
	if defaultApp == nil {
		panic("app: defaultApp is not initialized, call app.Init() first")
	}
	defaultApp.Run()
}
