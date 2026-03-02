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

const defaultPingTimeout = 5 * time.Second

// Manager holds all initialized infrastructure resources.
type Manager struct {
	mysql       map[string]*database.DB
	redis       map[string]*redisclient.Client
	pingTimeout time.Duration
}

// New initializes all configured resources and validates connectivity with ping checks.
func New(cfg *Config) (*Manager, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	m := &Manager{
		mysql:       make(map[string]*database.DB, len(cfg.MySQL)),
		redis:       make(map[string]*redisclient.Client, len(cfg.Redis)),
		pingTimeout: cfg.Init.PingTimeout,
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
func MustNew(cfg *Config) *Manager {
	m, err := New(cfg)
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
		m.mysql[name] = db
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
		m.redis[name] = client
	}
	return nil
}

// MySQL returns a named MySQL connection.
func (m *Manager) MySQL(name string) (*database.DB, error) {
	db, ok := m.mysql[name]
	if !ok {
		return nil, fmt.Errorf("resource: mysql[%s] not found", name)
	}
	return db, nil
}

// MustMySQL returns a named MySQL connection or panics if it does not exist.
func (m *Manager) MustMySQL(name string) *database.DB {
	db, err := m.MySQL(name)
	if err != nil {
		panic(err)
	}
	return db
}

// Redis returns a named Redis client.
func (m *Manager) Redis(name string) (*redisclient.Client, error) {
	client, ok := m.redis[name]
	if !ok {
		return nil, fmt.Errorf("resource: redis[%s] not found", name)
	}
	return client, nil
}

// MustRedis returns a named Redis client or panics if it does not exist.
func (m *Manager) MustRedis(name string) *redisclient.Client {
	client, err := m.Redis(name)
	if err != nil {
		panic(err)
	}
	return client
}

// Close closes all managed resources.
func (m *Manager) Close() error {
	if m == nil {
		return nil
	}

	var errs []error
	for _, name := range sortedKeys(m.redis) {
		if err := m.redis[name].Close(); err != nil {
			errs = append(errs, fmt.Errorf("close redis[%s]: %w", name, err))
		}
	}

	for _, name := range sortedKeys(m.mysql) {
		if err := m.mysql[name].Close(); err != nil {
			errs = append(errs, fmt.Errorf("close mysql[%s]: %w", name, err))
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
