package etcd

import "context"

type etcdKey struct{}

func FromContext(ctx context.Context) *Config {
	if c, ok := ctx.Value(etcdKey{}).(*Config); ok {
		return c
	}
	return nil
}

func WithContext(ctx context.Context, c *Config) context.Context {
	return context.WithValue(ctx, etcdKey{}, c)
}
