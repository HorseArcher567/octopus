package assemble

import (
	"fmt"
	"strings"

	"github.com/HorseArcher567/octopus/pkg/config"
	mysqlpkg "github.com/HorseArcher567/octopus/pkg/mysql"
	"github.com/HorseArcher567/octopus/pkg/store"
)

func setupMySQL(c *setupContext) error {
	value, ok := c.get("mysql")
	if !ok {
		return nil
	}
	rawItems, ok := value.([]any)
	if !ok {
		return fmt.Errorf("decode config %q: invalid type %T", "mysql", value)
	}
	items := make([]mysqlpkg.Config, 0, len(rawItems))
	for i, raw := range rawItems {
		m, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("assemble: mysql[%d]: invalid config type %T", i, raw)
		}
		tmp := config.New()
		for k, v := range m {
			tmp.Set(k, v)
		}
		var item mysqlpkg.Config
		if err := tmp.UnmarshalStrict(&item); err != nil {
			return fmt.Errorf("assemble: mysql[%d]: %w", i, err)
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		return nil
	}
	if err := validateMySQLConfigs(items); err != nil {
		return err
	}
	for _, item := range items {
		db, err := mysqlpkg.New(&item)
		if err != nil {
			return fmt.Errorf("assemble: mysql[%s]: %w", item.Name, err)
		}
		if err := c.provide(item.Name, db, store.WithClose(db.Close)); err != nil {
			_ = db.Close()
			return fmt.Errorf("assemble: mysql[%s]: %w", item.Name, err)
		}
	}
	return nil
}

func validateMySQLConfigs(items []mysqlpkg.Config) error {
	seen := make(map[string]struct{}, len(items))
	for i, item := range items {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			return fmt.Errorf("assemble: mysql[%d]: name is required", i)
		}
		if _, ok := seen[name]; ok {
			return fmt.Errorf("assemble: mysql[%s]: duplicate name", name)
		}
		seen[name] = struct{}{}
		if err := item.Validate(); err != nil {
			return fmt.Errorf("assemble: mysql[%s]: %w", name, err)
		}
	}
	return nil
}
