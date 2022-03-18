package config

type Prometheus struct {
	Enabled   bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Subsystem string `json:"subsystem,omitempty" yaml:"subsystem,omitempty"`
}
