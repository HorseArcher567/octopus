package trace

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

// Config controls tracing provider creation.
type Config struct {
	Enabled  bool
	Exporter string
	Endpoint string
	Service  string
}

// Runtime owns the tracer provider lifecycle.
type Runtime struct {
	enabled  bool
	provider *sdktrace.TracerProvider
}

// New creates a tracing runtime.
func New(cfg Config) (*Runtime, error) {
	if !cfg.Enabled {
		return &Runtime{}, nil
	}
	if cfg.Service == "" {
		cfg.Service = "octopus"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := resource.New(ctx, resource.WithAttributes(semconv.ServiceName(cfg.Service)))
	if err != nil {
		return nil, fmt.Errorf("trace: resource: %w", err)
	}

	var exporter sdktrace.SpanExporter
	switch cfg.Exporter {
	case "", "stdout":
		exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
	case "otlp":
		opts := []otlptracehttp.Option{}
		if cfg.Endpoint != "" {
			opts = append(opts, otlptracehttp.WithEndpoint(cfg.Endpoint), otlptracehttp.WithInsecure())
		}
		exporter, err = otlptracehttp.New(ctx, opts...)
	default:
		return nil, fmt.Errorf("trace: unsupported exporter %q", cfg.Exporter)
	}
	if err != nil {
		return nil, fmt.Errorf("trace: exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	return &Runtime{enabled: true, provider: tp}, nil
}

// Enabled reports whether tracing is active.
func (r *Runtime) Enabled() bool { return r != nil && r.enabled }

// Shutdown flushes and stops the tracer provider.
func (r *Runtime) Shutdown(ctx context.Context) error {
	if r == nil || r.provider == nil {
		return nil
	}
	return r.provider.Shutdown(ctx)
}
