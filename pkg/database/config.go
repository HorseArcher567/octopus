package database

import (
	"errors"
	"time"
)

// Config is the configuration for database connection.
type Config struct {
	// DSN is the data source name (e.g., "user:pass@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local").
	DSN string `yaml:"dsn" json:"dsn" toml:"dsn"`

	// DriverName is the database driver name (default: "mysql").
	DriverName string `yaml:"driverName" json:"driverName" toml:"driverName"`

	// MaxOpenConns is the maximum number of open connections to the database (default: 0, unlimited).
	MaxOpenConns int `yaml:"maxOpenConns" json:"maxOpenConns" toml:"maxOpenConns"`

	// MaxIdleConns is the maximum number of idle connections.
	// If not set or set to 0, it defaults to 2.
	MaxIdleConns int `yaml:"maxIdleConns" json:"maxIdleConns" toml:"maxIdleConns"`

	// ConnMaxLifetime is the maximum amount of time a connection may be reused (default: 0, unlimited).
	// Unit: seconds
	ConnMaxLifetime int `yaml:"connMaxLifetime" json:"connMaxLifetime" toml:"connMaxLifetime"`

	// ConnMaxIdleTime is the maximum amount of time a connection may be idle (default: 0, unlimited).
	// Unit: seconds
	ConnMaxIdleTime int `yaml:"connMaxIdleTime" json:"connMaxIdleTime" toml:"connMaxIdleTime"`
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.DSN == "" {
		return errors.New("database: DSN is required")
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

	return nil
}

// GetDriverName returns the driver name, default to "mysql".
func (c *Config) GetDriverName() string {
	if c.DriverName == "" {
		return "mysql"
	}
	return c.DriverName
}

// GetConnMaxLifetime returns the connection max lifetime duration.
func (c *Config) GetConnMaxLifetime() time.Duration {
	if c.ConnMaxLifetime <= 0 {
		return 0
	}
	return time.Duration(c.ConnMaxLifetime) * time.Second
}

// GetConnMaxIdleTime returns the connection max idle time duration.
func (c *Config) GetConnMaxIdleTime() time.Duration {
	if c.ConnMaxIdleTime <= 0 {
		return 0
	}
	return time.Duration(c.ConnMaxIdleTime) * time.Second
}
