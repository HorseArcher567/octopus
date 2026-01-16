package etcd

import (
	"errors"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// Config is the configuration for etcd client connection.
type Config struct {
	// Endpoints is the list of etcd node addresses.
	Endpoints []string `yaml:"endpoints" json:"endpoints" toml:"endpoints"`

	// DialTimeout is the connection timeout (default: 5s).
	DialTimeout time.Duration `yaml:"dialTimeout" json:"dialTimeout" toml:"dialTimeout"`

	// Username is the etcd username (optional).
	Username string `yaml:"username" json:"username" toml:"username"`

	// Password is the etcd password (optional).
	Password string `yaml:"password" json:"password" toml:"password"`
}

// ClientV3Config returns a clientv3.Config for creating an etcd client.
func (c *Config) ClientV3Config() (*clientv3.Config, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	cfg := &clientv3.Config{
		Endpoints: c.Endpoints,
	}

	if c.DialTimeout > 0 {
		cfg.DialTimeout = c.DialTimeout
	} else {
		cfg.DialTimeout = 5 * time.Second
	}

	if c.Username != "" {
		cfg.Username = c.Username
		cfg.Password = c.Password
	}

	return cfg, nil
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if len(c.Endpoints) == 0 {
		return errors.New("endpoints not configured")
	}

	return nil
}
