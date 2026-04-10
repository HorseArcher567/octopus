package health

// Config controls the built-in health endpoint.
type Config struct {
	Enabled bool   `yaml:"enabled" json:"enabled" toml:"enabled"`
	Path    string `yaml:"path" json:"path" toml:"path"`
}

// FillDefaults applies conservative defaults.
func (c *Config) FillDefaults() {
	if c.Path == "" {
		c.Path = "/health"
	}
}
