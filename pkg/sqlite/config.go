package sqlite

import (
	"fmt"
	"strings"
	"time"

	"github.com/HorseArcher567/octopus/pkg/database"
)

type Config struct {
	Name string              `yaml:"name" json:"name" toml:"name"`
	DSN  string              `yaml:"dsn" json:"dsn" toml:"dsn"`
	Pool database.PoolConfig `yaml:"pool" json:"pool" toml:"pool"`
}

func (c *Config) Normalize() {
	if c == nil {
		return
	}
	c.Pool.Normalize()
	if c.Pool.MaxOpenConns == 0 {
		c.Pool.MaxOpenConns = 1
	}
	if c.Pool.MaxIdleConns == 0 {
		c.Pool.MaxIdleConns = 1
	}
	if c.Pool.PingTimeout == 0 {
		c.Pool.PingTimeout = 5 * time.Second
	}
}

func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("sqlite: config cannot be nil")
	}
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("sqlite: name is required")
	}
	if strings.TrimSpace(c.DSN) == "" {
		return fmt.Errorf("sqlite: DSN is required")
	}
	if err := c.Pool.Validate(); err != nil {
		return err
	}
	return nil
}
