package assemble

import (
	"fmt"
	"strings"

	"github.com/HorseArcher567/octopus/pkg/config"
	sqlitepkg "github.com/HorseArcher567/octopus/pkg/sqlite"
	"github.com/HorseArcher567/octopus/pkg/store"
)

func setupSQLite(c *setupContext) error {
	value, ok := c.get("sqlite")
	if !ok {
		return nil
	}
	rawItems, ok := value.([]any)
	if !ok {
		if m, ok := value.(map[string]any); ok {
			rawItems = []any{m}
		} else {
			return fmt.Errorf("decode config %q: invalid type %T", "sqlite", value)
		}
	}
	items := make([]sqlitepkg.Config, 0, len(rawItems))
	for i, raw := range rawItems {
		m, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("assemble: sqlite[%d]: invalid config type %T", i, raw)
		}
		tmp := config.New()
		for k, v := range m {
			tmp.Set(k, v)
		}
		var item sqlitepkg.Config
		if err := tmp.UnmarshalStrict(&item); err != nil {
			return fmt.Errorf("assemble: sqlite[%d]: %w", i, err)
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		return nil
	}
	if err := validateSQLiteConfigs(items); err != nil {
		return err
	}
	for _, item := range items {
		db, err := sqlitepkg.New(&item)
		if err != nil {
			return fmt.Errorf("assemble: sqlite[%s]: %w", item.Name, err)
		}
		if err := c.provide(item.Name, db, store.WithClose(db.Close)); err != nil {
			_ = db.Close()
			return fmt.Errorf("assemble: sqlite[%s]: %w", item.Name, err)
		}
	}
	return nil
}

func validateSQLiteConfigs(items []sqlitepkg.Config) error {
	seen := make(map[string]struct{}, len(items))
	for i, item := range items {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			return fmt.Errorf("assemble: sqlite[%d]: name is required", i)
		}
		if _, ok := seen[name]; ok {
			return fmt.Errorf("assemble: sqlite[%s]: duplicate name", name)
		}
		seen[name] = struct{}{}
		if err := item.Validate(); err != nil {
			return fmt.Errorf("assemble: sqlite[%s]: %w", name, err)
		}
	}
	return nil
}
