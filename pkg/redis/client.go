package redis

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
)

const defaultPingTimeout = 5 * time.Second

// Client wraps a go-redis client with unified initialization helpers.
type Client struct {
	*goredis.Client
}

// New creates a new Redis client with the given configuration.
func New(cfg *Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	c := goredis.NewClient(&goredis.Options{
		Addr:         cfg.Addr,
		Username:     cfg.Username,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		// Keep startup logs quiet and avoid capability probing on servers
		// that do not implement CLIENT MAINT_NOTIFICATIONS.
		MaintNotificationsConfig: &maintnotifications.Config{
			Mode: maintnotifications.ModeDisabled,
		},
	})

	client := &Client{Client: c}
	if err := client.PingTimeout(defaultPingTimeout); err != nil {
		_ = client.Close()
		return nil, err
	}

	return client, nil
}

// MustNew creates a new Redis client and panics if an error occurs.
func MustNew(cfg *Config) *Client {
	c, err := New(cfg)
	if err != nil {
		panic("redis: failed to create client: " + err.Error())
	}
	return c
}

// PingTimeout verifies the connection to Redis with a timeout.
func (c *Client) PingTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return c.Ping(ctx).Err()
}
