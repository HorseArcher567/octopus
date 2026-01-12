package xlog

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/HorseArcher567/octopus/pkg/rotate"
)



// Logger 封装了 slog.Logger 和资源清理逻辑
// 通过嵌入 *slog.Logger，可以直接调用所有 slog 的方法
type Logger struct {
	*slog.Logger
	closer io.Closer
}

func (l *Logger) Close() error {
	if l == nil || l.closer == nil {
		return nil
	}
	return l.closer.Close()
}

// New 根据配置创建一个新的 Logger
// 返回的 Logger 实现了 io.Closer，使用完毕后应调用 Close() 关闭资源
func New(cfg Config) (*Logger, error) {
	cfg = normalize(cfg)

	writer, closer, err := resolveWriter(cfg)
	if err != nil {
		return nil, err
	}

	level, err := resolveLevel(cfg.Level)
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("unsupported log format: %s", cfg.Format)
	}

	return &Logger{
		Logger: slog.New(handler),
		closer: closer,
	}, nil
}

// MustNew 根据配置创建一个新的 Logger（失败时 panic）
func MustNew(cfg Config) *Logger {
	logger, err := New(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}
	return logger
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
