package resource

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/HorseArcher567/octopus/pkg/database"
	redisclient "github.com/HorseArcher567/octopus/pkg/redis"
)

const (
	KindMySQL = "mysql"
	KindRedis = "redis"
)

type entry struct {
	value   any
	closeFn func() error
}

const defaultPingTimeout = 5 * time.Second

// Manager holds all initialized infrastructure resources.
type Manager struct {
	resources   map[string]map[string]entry
	pingTimeout time.Duration
}

// New initializes all configured resources and validates connectivity with ping checks.
func New(cfg *Config, opts ...Option) (*Manager, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	m := &Manager{
		resources:   make(map[string]map[string]entry, 2),
		pingTimeout: cfg.Init.PingTimeout,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(m)
		}
	}
	if m.pingTimeout <= 0 {
		m.pingTimeout = defaultPingTimeout
	}

	if err := m.initMySQL(cfg.MySQL); err != nil {
		_ = m.Close()
		return nil, err
	}

	if err := m.initRedis(cfg.Redis); err != nil {
		_ = m.Close()
		return nil, err
	}

	return m, nil
}

// MustNew initializes resources and panics if an error occurs.
func MustNew(cfg *Config, opts ...Option) *Manager {
	m, err := New(cfg, opts...)
	if err != nil {
		panic("resource: failed to initialize resources: " + err.Error())
	}
	return m
}

func (m *Manager) initMySQL(cfgs map[string]database.Config) error {
	for _, name := range sortedKeys(cfgs) {
		cfg := cfgs[name]
		db, err := database.New(&cfg)
		if err != nil {
			return fmt.Errorf("mysql[%s]: %w", name, err)
		}
		if err := db.Ping(); err != nil {
			_ = db.Close()
			return fmt.Errorf("mysql[%s]: ping failed: %w", name, err)
		}
		if err := m.Register(KindMySQL, name, db, db.Close); err != nil {
			_ = db.Close()
			return err
		}
	}
	return nil
}

func (m *Manager) initRedis(cfgs map[string]redisclient.Config) error {
	for _, name := range sortedKeys(cfgs) {
		cfg := cfgs[name]
		client, err := redisclient.New(&cfg)
		if err != nil {
			return fmt.Errorf("redis[%s]: %w", name, err)
		}
		if err := client.PingTimeout(m.pingTimeout); err != nil {
			_ = client.Close()
			return fmt.Errorf("redis[%s]: ping failed: %w", name, err)
		}
		if err := m.Register(KindRedis, name, client, client.Close); err != nil {
			_ = client.Close()
			return err
		}
	}
	return nil
}

// Register stores a resource instance under a kind/name pair.
func (m *Manager) Register(kind, name string, value any, closeFn func() error) error {
	if kind == "" {
		return errors.New("resource: kind is required")
	}
	if name == "" {
		return errors.New("resource: name is required")
	}
	if value == nil {
		return fmt.Errorf("resource: %s[%s] cannot be nil", kind, name)
	}
	if m.resources[kind] == nil {
		m.resources[kind] = make(map[string]entry)
	}
	if _, exists := m.resources[kind][name]; exists {
		return fmt.Errorf("resource: duplicate %s[%s]", kind, name)
	}
	m.resources[kind][name] = entry{value: value, closeFn: closeFn}
	return nil
}

// Get returns a resource by kind and name.
func (m *Manager) Get(kind, name string) (any, error) {
	entries, ok := m.resources[kind]
	if !ok {
		return nil, fmt.Errorf("resource: kind %q not found", kind)
	}
	e, ok := entries[name]
	if !ok {
		return nil, fmt.Errorf("resource: %s[%s] not found", kind, name)
	}
	return e.value, nil
}

// MustGet returns a resource by kind and name or panics.
func (m *Manager) MustGet(kind, name string) any {
	v, err := m.Get(kind, name)
	if err != nil {
		panic(err)
	}
	return v
}

// List returns all registered resource names for a kind.
func (m *Manager) List(kind string) []string {
	return sortedKeys(m.resources[kind])
}

// Close closes all managed resources.
func (m *Manager) Close() error {
	if m == nil {
		return nil
	}

	var errs []error
	for _, kind := range sortedKeys(m.resources) {
		for _, name := range sortedKeys(m.resources[kind]) {
			e := m.resources[kind][name]
			if e.closeFn == nil {
				continue
			}
			if err := e.closeFn(); err != nil {
				errs = append(errs, fmt.Errorf("close %s[%s]: %w", kind, name, err))
			}
		}
	}

	return errors.Join(errs...)
}

func sortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return slices.Clip(keys)
}
