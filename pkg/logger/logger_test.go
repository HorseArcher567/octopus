package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func newTestLogger(t *testing.T, cfg Config, buf *bytes.Buffer) *slog.Logger {
	t.Helper()

	level, err := resolveLevel(cfg.Level)
	if err != nil {
		t.Fatalf("invalid level: %v", err)
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
	}

	var handler slog.Handler
	switch cfg.Format {
	case "json":
		handler = slog.NewJSONHandler(buf, opts)
	default:
		handler = slog.NewTextHandler(buf, opts)
	}

	return slog.New(handler)
}

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "text format",
			config: Config{
				Level:     "info",
				Format:    "text",
				AddSource: false,
				Output:    "stdout",
			},
		},
		{
			name: "json format",
			config: Config{
				Level:     "debug",
				Format:    "json",
				AddSource: true,
				Output:    "stderr",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, closer, err := New(tt.config)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			if closer != nil {
				closer.Close()
			}
			if logger == nil {
				t.Error("New() returned nil logger")
			}
		})
	}
}

func TestLogLevels(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(t, Config{
		Level:     "debug",
		Format:    "text",
		AddSource: false,
	}, buf)

	logger.Debug("debug message", "key", "value")
	logger.Info("info message", "key", "value")
	logger.Warn("warn message", "key", "value")
	logger.Error("error message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, `msg="debug message"`) {
		t.Errorf("debug message not found in output:\n%s", output)
	}
	if !strings.Contains(output, `msg="info message"`) {
		t.Errorf("info message not found in output:\n%s", output)
	}
	if !strings.Contains(output, `msg="warn message"`) {
		t.Errorf("warn message not found in output:\n%s", output)
	}
	if !strings.Contains(output, `msg="error message"`) {
		t.Errorf("error message not found in output:\n%s", output)
	}
}

func TestWith(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(t, Config{
		Level:     "info",
		Format:    "text",
		AddSource: false,
	}, buf)

	childLogger := logger.With("component", "test", "version", "1.0")
	childLogger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "component=test") {
		t.Error("component field not found")
	}
	if !strings.Contains(output, "version=1.0") {
		t.Error("version field not found")
	}
}

func TestAddSource(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(t, Config{
		Level:     "info",
		Format:    "text",
		AddSource: true,
	}, buf)

	logger.Info("test source location")

	output := buf.String()
	if !strings.Contains(output, "source=") {
		t.Error("source location not found")
	}
}

func TestJSONFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(t, Config{
		Level:     "info",
		Format:    "json",
		AddSource: false,
	}, buf)

	logger.Info("json test", "key", "value")

	output := buf.String()
	if !strings.Contains(output, `"msg":"json test"`) {
		t.Error("msg field not found in JSON")
	}
	if !strings.Contains(output, `"key":"value"`) {
		t.Error("key field not found in JSON")
	}
}

func TestMustNew(t *testing.T) {
	logger, closer := MustNew(Config{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	})
	if closer != nil {
		closer.Close()
	}
	if logger == nil {
		t.Error("MustNew() returned nil logger")
	}
}

func TestMustNewPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustNew() should panic with invalid config")
		}
	}()

	MustNew(Config{
		Level:  "invalid",
		Format: "text",
		Output: "stdout",
	})
}
