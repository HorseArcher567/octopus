package telemetry

// Option customizes telemetry config during assembly.
type Option func(*Config)

// WithServiceName sets the telemetry service name.
func WithServiceName(name string) Option {
	return func(c *Config) { c.ServiceName = name }
}

// WithMetrics customizes metrics telemetry.
func WithMetrics(enabled bool, path string) Option {
	return func(c *Config) {
		c.Metrics.Enabled = enabled
		if path != "" {
			c.Metrics.Path = path
		}
	}
}

// WithTrace customizes trace telemetry.
func WithTrace(enabled bool, exporter, endpoint string) Option {
	return func(c *Config) {
		c.Trace.Enabled = enabled
		if exporter != "" {
			c.Trace.Exporter = exporter
		}
		if endpoint != "" {
			c.Trace.Endpoint = endpoint
		}
	}
}
