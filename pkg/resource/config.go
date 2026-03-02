package resource

import (
	"errors"
	"time"

	"github.com/HorseArcher567/octopus/pkg/database"
	redisclient "github.com/HorseArcher567/octopus/pkg/redis"
)

// Config describes all shared infrastructure resources.
type Config struct {
	// MySQL defines named MySQL connections.
	MySQL map[string]database.Config `yaml:"mysql" json:"mysql" toml:"mysql"`

	// Redis defines named Redis connections.
	Redis map[string]redisclient.Config `yaml:"redis" json:"redis" toml:"redis"`

	// Init controls initialization behavior.
	Init InitConfig `yaml:"init" json:"init" toml:"init"`
}

// InitConfig controls resource bootstrap behavior.
type InitConfig struct {
	// PingTimeout is used for startup health checks.
	PingTimeout time.Duration `yaml:"pingTimeout" json:"pingTimeout" toml:"pingTimeout"`
}

// Validate validates all configured resources.
func (c *Config) Validate() error {
	if c == nil {
		return errors.New("resource: config is nil")
	}

	for name, mysqlCfg := range c.MySQL {
		cfg := mysqlCfg
		if err := cfg.Validate(); err != nil {
			return errors.New("resource: mysql[" + name + "]: " + err.Error())
		}
	}

	for name, redisCfg := range c.Redis {
		cfg := redisCfg
		if err := cfg.Validate(); err != nil {
			return errors.New("resource: redis[" + name + "]: " + err.Error())
		}
	}

	if c.Init.PingTimeout < 0 {
		return errors.New("resource: init.pingTimeout cannot be negative")
	}

	return nil
}
