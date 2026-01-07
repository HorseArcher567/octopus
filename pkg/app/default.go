package app

import (
	"fmt"

	"google.golang.org/grpc"
)

// defaultApp 是类似 slog.Default() 的全局默认应用实例。
var defaultApp *App

// Init 初始化默认应用实例。
// 应在 main 函数启动阶段调用一次。
func Init(opts ...Option) {
	defaultApp = New(opts...)
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

// RegisterService 在默认应用上注册 gRPC 服务。
func RegisterService(register func(s *grpc.Server)) {
	Default().RegisterService(register)
}

// Run 启动默认应用。
func Run() error {
	if defaultApp == nil {
		return fmt.Errorf("app: defaultApp is not initialized, call app.Init() first")
	}
	return defaultApp.Run()
}
