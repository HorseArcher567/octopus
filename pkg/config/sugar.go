package config

import (
	"fmt"
)

// ============================================================================
// Package-level helpers
// ============================================================================

// Load loads a single config file with automatic format detection.
// Environment variables are expanded by default using ${ENV_VAR} or ${ENV_VAR:default_value}.
// This is the most common entry point.
func Load(path string) (*Config, error) {
	format := detectFormat(path)
	if format == FormatUnknown {
		return nil, fmt.Errorf("cannot detect format from file extension: %s", path)
	}

	cfg := New()
	if err := cfg.Load(path); err != nil {
		return nil, err
	}

	replaceEnvVars(cfg.data)
	return cfg, nil
}

// LoadWithoutEnv loads a config file without environment variable expansion.
func LoadWithoutEnv(path string) (*Config, error) {
	cfg := New()
	if err := cfg.Load(path); err != nil {
		return nil, err
	}
	return cfg, nil
}

// LoadFromBytes loads config from raw bytes.
func LoadFromBytes(data []byte, format Format) (*Config, error) {
	cfg := New()
	if err := cfg.LoadBytes(data, format); err != nil {
		return nil, err
	}
	return cfg, nil
}

// ============================================================================
// Must* helpers
// ============================================================================

// MustLoad loads a config file and panics on failure.
// Environment variable expansion is enabled by default.
func MustLoad(path string) *Config {
	cfg, err := Load(path)
	if err != nil {
		panic(fmt.Errorf("config: failed to load config from %s: %w", path, err))
	}
	return cfg
}

// MustLoadWithoutEnv loads a config file without environment variable expansion and panics on failure.
func MustLoadWithoutEnv(path string) *Config {
	cfg, err := LoadWithoutEnv(path)
	if err != nil {
		panic(fmt.Errorf("config: failed to load config from %s: %w", path, err))
	}
	return cfg
}

// MustUnmarshal loads config, unmarshals it into target, and panics on failure.
func MustUnmarshal(path string, target interface{}) {
	cfg := MustLoad(path)
	if err := cfg.Unmarshal(target); err != nil {
		panic(fmt.Errorf("config: failed to unmarshal config from %s: %w", path, err))
	}
}

// MustUnmarshalWithoutEnv loads config without environment variable expansion,
// unmarshals it into target, and panics on failure.
func MustUnmarshalWithoutEnv(path string, target interface{}) {
	cfg := MustLoadWithoutEnv(path)
	if err := cfg.Unmarshal(target); err != nil {
		panic(fmt.Errorf("config: failed to unmarshal config from %s: %w", path, err))
	}
}
