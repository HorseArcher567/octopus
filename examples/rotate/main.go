package main

import (
	"log/slog"
	"time"

	"github.com/HorseArcher567/octopus/pkg/rotate"
)

func main() {
	// 创建日志轮转器（使用 MustNew，日志初始化失败时 panic）
	// 日志是基础功能，如果初始化失败，应用不应该继续运行
	writer, closer := rotate.MustNew(rotate.Config{
		Filename: "logs/app.log",
		MaxAge:   7, // 保留7天
	})
	defer closer.Close()

	// 方式1: 使用 rotate 和 slog 创建自定义 logger
	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})
	customLogger := slog.New(handler)

	// 设置为默认 logger
	slog.SetDefault(customLogger)

	// 记录日志
	slog.Info("application started",
		"version", "1.0.0",
		"env", "production")

	// 模拟应用运行，生成大量日志
	for i := 0; i < 100; i++ {
		slog.Info("processing request",
			"request_id", generateRequestID(),
			"method", "GET",
			"path", "/api/users",
			"status", 200,
			"duration_ms", 42,
		)

		slog.Debug("detailed debug information",
			"iteration", i,
			"memory_mb", 128.5)

		if i%10 == 0 {
			slog.Warn("warning message",
				"iteration", i,
				"warning", "high memory usage")
		}

		if i%50 == 0 {
			slog.Error("error occurred",
				"iteration", i,
				"error", "database connection timeout")
		}

		time.Sleep(10 * time.Millisecond)
	}

	slog.Info("application stopped")
}

// generateRequestID 生成模拟的请求ID
func generateRequestID() string {
	return time.Now().Format("20060102150405.000000")
}
