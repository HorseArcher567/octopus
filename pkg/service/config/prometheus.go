package config

type Prometheus struct {
	Server struct {
		Enabled            bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
		Namespace          string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
		Subsystem          string `json:"subsystem,omitempty" yaml:"subsystem,omitempty"`
		CountsHandlingTime bool   `json:"countsHandlingTime,omitempty" yaml:"countsHandlingTime,omitempty"`
	} `json:"server,omitempty" yaml:"server,omitempty"`
	Client struct {
		Enabled   bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
		Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
		Subsystem string `json:"subsystem,omitempty" yaml:"subsystem,omitempty"`
	} `json:"client,omitempty" yaml:"client,omitempty"`
}
