package telemetry

import (
	"context"

	oteltrace "github.com/HorseArcher567/octopus/pkg/telemetry/trace"
)

// Runtime holds telemetry capabilities assembled by the framework.
type Runtime struct {
	MetricsPath string
	Trace       *oteltrace.Runtime
	ServiceName string
}

// New creates a telemetry runtime from config.
func New(cfg *Config, opts ...Option) (*Runtime, error) {
	if cfg == nil {
		cfg = &Config{}
	}
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}
	cfg.FillDefaults()

	rt := &Runtime{ServiceName: cfg.ServiceName}
	if cfg.Metrics.Enabled {
		rt.MetricsPath = cfg.Metrics.Path
	}
	traceRuntime, err := oteltrace.New(oteltrace.Config{
		Enabled:  cfg.Trace.Enabled,
		Exporter: cfg.Trace.Exporter,
		Endpoint: cfg.Trace.Endpoint,
		Service:  cfg.ServiceName,
	})
	if err != nil {
		return nil, err
	}
	rt.Trace = traceRuntime
	return rt, nil
}

// Close flushes telemetry runtimes that require shutdown.
func (r *Runtime) Close() error {
	if r == nil || r.Trace == nil {
		return nil
	}
	return r.Trace.Shutdown(context.Background())
}
