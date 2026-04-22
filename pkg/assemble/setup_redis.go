package assemble

import (
	"fmt"
	"strings"

	"github.com/HorseArcher567/octopus/pkg/config"
	redisclient "github.com/HorseArcher567/octopus/pkg/redis"
	"github.com/HorseArcher567/octopus/pkg/store"
)

func setupRedis(c *setupContext) error {
	value, ok := c.get("redis")
	if !ok {
		return nil
	}
	rawItems, ok := value.([]any)
	if !ok {
		return fmt.Errorf("decode config %q: invalid type %T", "redis", value)
	}
	items := make([]redisclient.Config, 0, len(rawItems))
	for i, raw := range rawItems {
		m, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("assemble: redis[%d]: invalid config type %T", i, raw)
		}
		tmp := config.New()
		for k, v := range m {
			tmp.Set(k, v)
		}
		var item redisclient.Config
		if err := tmp.UnmarshalStrict(&item); err != nil {
			return fmt.Errorf("assemble: redis[%d]: %w", i, err)
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		return nil
	}
	if err := validateRedisConfigs(items); err != nil {
		return err
	}
	for _, item := range items {
		client, err := redisclient.New(&item)
		if err != nil {
			return fmt.Errorf("assemble: redis[%s]: %w", item.Name, err)
		}
		if err := client.PingTimeout(item.PingTimeout); err != nil {
			_ = client.Close()
			return fmt.Errorf("assemble: redis[%s]: ping failed: %w", item.Name, err)
		}
		if err := c.provide(item.Name, client, store.WithClose(client.Close)); err != nil {
			_ = client.Close()
			return fmt.Errorf("assemble: redis[%s]: %w", item.Name, err)
		}
	}
	return nil
}

func validateRedisConfigs(items []redisclient.Config) error {
	seen := make(map[string]struct{}, len(items))
	for i, item := range items {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			return fmt.Errorf("assemble: redis[%d]: name is required", i)
		}
		if _, ok := seen[name]; ok {
			return fmt.Errorf("assemble: redis[%s]: duplicate name", name)
		}
		seen[name] = struct{}{}
		if err := item.Validate(); err != nil {
			return fmt.Errorf("assemble: redis[%s]: %w", name, err)
		}
	}
	return nil
}
