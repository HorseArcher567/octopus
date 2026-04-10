package telemetry

// Config controls built-in telemetry features.
type Config struct {
	ServiceName string        `yaml:"serviceName" json:"serviceName" toml:"serviceName"`
	Metrics     MetricsConfig `yaml:"metrics" json:"metrics" toml:"metrics"`
	Trace       TraceConfig   `yaml:"trace" json:"trace" toml:"trace"`
}

// MetricsConfig controls metrics endpoint behavior.
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled" toml:"enabled"`
	Path    string `yaml:"path" json:"path" toml:"path"`
}

// TraceConfig controls tracing runtime behavior.
type TraceConfig struct {
	Enabled  bool   `yaml:"enabled" json:"enabled" toml:"enabled"`
	Exporter string `yaml:"exporter" json:"exporter" toml:"exporter"`
	Endpoint string `yaml:"endpoint" json:"endpoint" toml:"endpoint"`
}

// FillDefaults applies conservative defaults.
func (c *Config) FillDefaults() {
	if c.ServiceName == "" {
		c.ServiceName = "octopus"
	}
	if c.Metrics.Path == "" {
		c.Metrics.Path = "/metrics"
	}
	if c.Trace.Exporter == "" {
		c.Trace.Exporter = "stdout"
	}
}
