package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/HorseArcher567/octopus/pkg/rotate"
)

// New 根据配置创建一个新的 slog.Logger
// 如果输出目标是文件，返回的 io.Closer 需要调用方关闭
func New(cfg Config) (*slog.Logger, io.Closer, error) {
	cfg = normalize(cfg)

	writer, closer, err := resolveWriter(cfg)
	if err != nil {
		return nil, nil, err
	}

	level, err := resolveLevel(cfg.Level)
	if err != nil {
		return nil, nil, err
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
	}

	var handler slog.Handler
	switch strings.ToLower(cfg.Format) {
	case "json":
		handler = slog.NewJSONHandler(writer, opts)
	case "text":
		handler = slog.NewTextHandler(writer, opts)
	default:
		return nil, nil, fmt.Errorf("unsupported log format: %s", cfg.Format)
	}

	return slog.New(handler), closer, nil
}

// MustNew 根据配置创建一个新的 slog.Logger（失败时 panic）
func MustNew(cfg Config) (*slog.Logger, io.Closer) {
	logger, closer, err := New(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}
	return logger, closer
}

func normalize(cfg Config) Config {
	if cfg.Level == "" {
		cfg.Level = "info"
	}
	if cfg.Format == "" {
		cfg.Format = "text"
	}
	if cfg.Output == "" {
		cfg.Output = "stdout"
	}
	return cfg
}

func resolveWriter(cfg Config) (io.Writer, io.Closer, error) {
	switch strings.ToLower(cfg.Output) {
	case "stdout":
		return os.Stdout, nil, nil
	case "stderr":
		return os.Stderr, nil, nil
	default:
		// 文件输出，自动启用按天轮转
		return rotate.New(rotate.Config{
			Filename: cfg.Output,
			MaxAge:   cfg.MaxAge,
		})
	}
}

func resolveLevel(level string) (slog.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("invalid log level: %s", level)
	}
}
