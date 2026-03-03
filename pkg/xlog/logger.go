package xlog

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/HorseArcher567/octopus/pkg/xlog/rotate"
)

// Logger is an owned logger handle.
//
// Only root loggers created by New carry a closer.
// Child loggers created via With/WithGroup intentionally do not inherit closer,
// which prevents accidentally closing shared sinks through derived loggers.
type Logger struct {
	*slog.Logger
	closer io.Closer
}

// With returns a derived logger with attrs attached.
func (l *Logger) With(attrs ...any) *Logger {
	if l == nil {
		return nil
	}
	return &Logger{Logger: l.Logger.With(attrs...)}
}

// WithGroup returns a derived logger scoped under name.
func (l *Logger) WithGroup(name string) *Logger {
	if l == nil {
		return nil
	}
	return &Logger{Logger: l.Logger.WithGroup(name)}
}

// Close releases resources owned by this logger.
// Calling Close on a non-root logger is a no-op.
func (l *Logger) Close() error {
	if l == nil || l.closer == nil {
		return nil
	}
	return l.closer.Close()
}

// MustNew is like New but panics on error.
func MustNew(cfg *Config) *Logger {
	logger, err := New(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}
	return logger
}

// New constructs a root logger from cfg.
// The returned logger may own file resources; callers should close it before exit.
func New(cfg *Config) (*Logger, error) {
	if cfg == nil {
		cfg = &Config{}
	}
	normalize(cfg)

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

func normalize(cfg *Config) {
	if cfg.Level == "" {
		cfg.Level = "info"
	}
	if cfg.Format == "" {
		cfg.Format = "text"
	}
	if cfg.Output == "" {
		cfg.Output = "stdout"
	}
}

func resolveWriter(cfg *Config) (io.Writer, io.Closer, error) {
	switch strings.ToLower(cfg.Output) {
	case "stdout":
		return os.Stdout, nil, nil
	case "stderr":
		return os.Stderr, nil, nil
	default:
		// File output uses daily rotation.
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
