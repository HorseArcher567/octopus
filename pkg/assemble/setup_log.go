package assemble

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/HorseArcher567/octopus/pkg/app"
	"github.com/HorseArcher567/octopus/pkg/config"
	"github.com/HorseArcher567/octopus/pkg/store"
	"github.com/HorseArcher567/octopus/pkg/xlog"
)

func setupLoggers(c *setupContext) error {
	value, ok := c.get("logger")
	if !ok {
		return fmt.Errorf("assemble: logger is required")
	}
	rawItems, ok := value.([]any)
	if !ok {
		return fmt.Errorf("assemble: logger is required")
	}
	items := make([]xlog.Config, 0, len(rawItems))
	for i, raw := range rawItems {
		m, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("assemble: logger[%d]: invalid config type %T", i, raw)
		}
		tmp := config.New()
		tmp.Set("name", m["name"])
		for k, v := range m {
			tmp.Set(k, v)
		}
		var item xlog.Config
		if err := tmp.UnmarshalStrict(&item); err != nil {
			return fmt.Errorf("assemble: logger[%d]: %w", i, err)
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		return fmt.Errorf("assemble: logger is required")
	}

	seen := make(map[string]struct{}, len(items))
	for i, item := range items {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			return fmt.Errorf("assemble: logger[%d]: name is required", i)
		}
		if _, ok := seen[name]; ok {
			return fmt.Errorf("assemble: logger[%s]: duplicate name", name)
		}
		seen[name] = struct{}{}

		log, err := xlog.New(&item)
		if err != nil {
			return fmt.Errorf("assemble: logger[%s]: %w", name, err)
		}
		if err := c.provide(name, log, store.WithClose(log.Close)); err != nil {
			_ = log.Close()
			return fmt.Errorf("assemble: logger[%s]: %w", name, err)
		}
	}
	return nil
}

func selectAppLogger(c *setupContext) error {
	if _, ok := c.get("app"); !ok {
		return fmt.Errorf("assemble: app.logger is required")
	}
	var appCfg app.Config
	if err := c.decodeStruct("app", &appCfg); err != nil {
		return err
	}
	selected := strings.TrimSpace(appCfg.Logger)
	if selected == "" {
		return fmt.Errorf("assemble: app.logger is required")
	}

	log, err := lookupLogger(selected, c.state.store)
	if err != nil {
		return fmt.Errorf("assemble: app.logger: %w", err)
	}
	c.state.log = log
	return nil
}

func selectLogger(name string, fallback *xlog.Logger, st store.Store) (*xlog.Logger, error) {
	selected := strings.TrimSpace(name)
	if selected == "" {
		if fallback == nil {
			return nil, fmt.Errorf("logger fallback cannot be nil")
		}
		return fallback, nil
	}
	return lookupLogger(selected, st)
}

func lookupLogger(name string, st store.Store) (*xlog.Logger, error) {
	value, err := st.GetNamed(name, reflect.TypeFor[*xlog.Logger]())
	if err != nil {
		return nil, fmt.Errorf("logger %q not found", name)
	}
	log, ok := value.(*xlog.Logger)
	if !ok || log == nil {
		return nil, fmt.Errorf("logger %q has invalid type %T", name, value)
	}
	return log, nil
}
