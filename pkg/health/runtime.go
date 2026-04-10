package health

// Runtime holds health capabilities assembled by the framework.
type Runtime struct {
	Registry *Registry
	Path     string
}

// NewRuntime creates a health runtime from config.
func NewRuntime(cfg *Config) *Runtime {
	if cfg == nil {
		cfg = &Config{}
	}
	cfg.FillDefaults()
	if !cfg.Enabled {
		return &Runtime{}
	}
	return &Runtime{
		Registry: New(),
		Path:     cfg.Path,
	}
}
