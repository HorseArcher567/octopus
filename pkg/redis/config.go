package redis

import (
	"errors"
	"time"
)

// Config is the configuration for Redis connection.
type Config struct {
	// Addr is the Redis server address in host:port format.
	Addr string `yaml:"addr" json:"addr" toml:"addr"`

	// Username is the ACL username.
	Username string `yaml:"username" json:"username" toml:"username"`

	// Password is the ACL password.
	Password string `yaml:"password" json:"password" toml:"password"`

	// DB is the Redis database index.
	DB int `yaml:"db" json:"db" toml:"db"`

	// PoolSize is the base number of socket connections.
	PoolSize int `yaml:"poolSize" json:"poolSize" toml:"poolSize"`

	// MinIdleConns is the minimum number of idle connections.
	MinIdleConns int `yaml:"minIdleConns" json:"minIdleConns" toml:"minIdleConns"`

	// MaxRetries is the maximum number of retries before giving up.
	MaxRetries int `yaml:"maxRetries" json:"maxRetries" toml:"maxRetries"`

	// DialTimeout is the timeout for establishing new connections.
	DialTimeout time.Duration `yaml:"dialTimeout" json:"dialTimeout" toml:"dialTimeout"`

	// ReadTimeout is the timeout for socket reads.
	ReadTimeout time.Duration `yaml:"readTimeout" json:"readTimeout" toml:"readTimeout"`

	// WriteTimeout is the timeout for socket writes.
	WriteTimeout time.Duration `yaml:"writeTimeout" json:"writeTimeout" toml:"writeTimeout"`
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c == nil {
		return errors.New("redis: config is nil")
	}

	if c.Addr == "" {
		return errors.New("redis: addr is required")
	}

	if c.DB < 0 {
		return errors.New("redis: db cannot be negative")
	}

	if c.PoolSize < 0 {
		return errors.New("redis: poolSize cannot be negative")
	}

	if c.MinIdleConns < 0 {
		return errors.New("redis: minIdleConns cannot be negative")
	}

	if c.MaxRetries < 0 {
		return errors.New("redis: maxRetries cannot be negative")
	}

	if c.DialTimeout < 0 {
		return errors.New("redis: dialTimeout cannot be negative")
	}

	if c.ReadTimeout < 0 {
		return errors.New("redis: readTimeout cannot be negative")
	}

	if c.WriteTimeout < 0 {
		return errors.New("redis: writeTimeout cannot be negative")
	}

	return nil
}
