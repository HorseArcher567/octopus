package database

import (
	"errors"
	"time"
)

// PoolConfig configures the SQL connection pool and startup ping behavior.
type PoolConfig struct {
	MaxOpenConns    int           `yaml:"maxOpenConns" json:"maxOpenConns" toml:"maxOpenConns"`
	MaxIdleConns    int           `yaml:"maxIdleConns" json:"maxIdleConns" toml:"maxIdleConns"`
	ConnMaxLifetime time.Duration `yaml:"connMaxLifetime" json:"connMaxLifetime" toml:"connMaxLifetime"`
	ConnMaxIdleTime time.Duration `yaml:"connMaxIdleTime" json:"connMaxIdleTime" toml:"connMaxIdleTime"`
	PingTimeout     time.Duration `yaml:"pingTimeout" json:"pingTimeout" toml:"pingTimeout"`
}

func (c *PoolConfig) Normalize() {
	if c == nil {
		return
	}
	if c.PingTimeout == 0 {
		c.PingTimeout = 5 * time.Second
	}
}

func (c *PoolConfig) Validate() error {
	if c == nil {
		return nil
	}
	if c.MaxOpenConns < 0 {
		return errors.New("database: MaxOpenConns cannot be negative")
	}
	if c.MaxIdleConns < 0 {
		return errors.New("database: MaxIdleConns cannot be negative")
	}
	if c.ConnMaxLifetime < 0 {
		return errors.New("database: ConnMaxLifetime cannot be negative")
	}
	if c.ConnMaxIdleTime < 0 {
		return errors.New("database: ConnMaxIdleTime cannot be negative")
	}
	if c.PingTimeout < 0 {
		return errors.New("database: PingTimeout cannot be negative")
	}
	return nil
}
