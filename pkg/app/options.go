package app

import (
	"io"
	"log/slog"

	"google.golang.org/grpc"
)

// Option 用于自定义 App 的初始化行为。
type Option func(a *App)

// WithConfigFile 指定配置文件路径（默认 config.yaml）。
func WithConfigFile(path string) Option {
	return func(a *App) {
		if path != "" {
			a.cfgPath = path
		}
	}
}

// WithConfig 直接提供配置对象，跳过文件加载。
func WithConfig(cfg *Config) Option {
	return func(a *App) {
		if cfg != nil {
			a.cfg = cfg
		}
	}
}

// WithLogger 使用已有的 logger 实例。
// closer 可为 nil；如不为 nil，会在 Run 结束后自动关闭。
func WithLogger(log *slog.Logger, closer io.Closer) Option {
	return func(a *App) {
		if log != nil {
			a.log = log
			a.logCloser = closer
		}
	}
}

// WithServerOptions 传入 gRPC Server 选项。
func WithServerOptions(opts ...grpc.ServerOption) Option {
	return func(a *App) {
		if len(opts) == 0 {
			return
		}
		a.grpcOpt = append(a.grpcOpt, opts...)
	}
}
