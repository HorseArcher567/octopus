package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// ============================================================================
// 包级加载函数 - 便捷接口
// ============================================================================

// Load 加载单个配置文件（自动识别格式）
func Load(path string) (*Config, error) {
	cfg := New()
	if err := cfg.LoadAndReplace(path); err != nil {
		return nil, err
	}
	return cfg, nil
}

// LoadFiles 加载多个配置文件并合并
func LoadFiles(paths ...string) (*Config, error) {
	cfg := New()
	for _, path := range paths {
		if err := cfg.Load(path); err != nil {
			return nil, fmt.Errorf("failed to load file %s: %w", path, err)
		}
	}
	return cfg, nil
}

// LoadDir 加载目录下的所有配置文件
func LoadDir(dir string) (*Config, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	cfg := New()
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// 只加载支持的配置文件
		if isSupportedFile(entry.Name()) {
			path := filepath.Join(dir, entry.Name())
			if err := cfg.Load(path); err != nil {
				return nil, fmt.Errorf("failed to load file %s: %w", path, err)
			}
		}
	}

	return cfg, nil
}

// LoadFromBytes 从字节流加载配置
func LoadFromBytes(data []byte, format Format) (*Config, error) {
	cfg := New()
	if err := cfg.LoadBytesAndReplace(data, format); err != nil {
		return nil, err
	}
	return cfg, nil
}

// LoadFromMap 从map加载配置
func LoadFromMap(data map[string]any) *Config {
	cfg := New()
	cfg.MergeMap(data)
	return cfg
}

// LoadWithEnv 加载配置并支持环境变量替换
// 环境变量格式: ${ENV_VAR} 或 ${ENV_VAR:default_value}
func LoadWithEnv(path string) (*Config, error) {
	cfg := New()
	if err := cfg.LoadAndReplace(path); err != nil {
		return nil, err
	}

	replaceEnvVars(cfg.data)
	return cfg, nil
}

// ============================================================================
// Must* 系列方法 - 失败时 panic，适用于启动阶段
// ============================================================================

// MustLoad 加载配置文件，失败时 panic
func MustLoad(path string) *Config {
	cfg, err := Load(path)
	if err != nil {
		panic(fmt.Sprintf("failed to load config from %s: %v", path, err))
	}
	return cfg
}

// MustLoadFiles 加载并合并多个配置文件，失败时 panic
func MustLoadFiles(paths ...string) *Config {
	cfg, err := LoadFiles(paths...)
	if err != nil {
		panic(fmt.Sprintf("failed to load config files: %v", err))
	}
	return cfg
}

// MustLoadWithEnv 加载配置并替换环境变量，失败时 panic
func MustLoadWithEnv(path string) *Config {
	cfg, err := LoadWithEnv(path)
	if err != nil {
		panic(fmt.Sprintf("failed to load config with env from %s: %v", path, err))
	}
	return cfg
}

// MustLoadAndUnmarshal 加载配置并直接解析到结构体，失败时 panic
// 这是最便捷的方式，适合在 main 函数中使用
func MustLoadAndUnmarshal(path string, target interface{}) {
	cfg, err := Load(path)
	if err != nil {
		panic(fmt.Sprintf("failed to load config from %s: %v", path, err))
	}
	if err := cfg.Unmarshal(target); err != nil {
		panic(fmt.Sprintf("failed to unmarshal config: %v", err))
	}
}

// MustLoadWithEnvAndUnmarshal 加载配置（支持环境变量）并解析到结构体，失败时 panic
func MustLoadWithEnvAndUnmarshal(path string, target interface{}) {
	cfg, err := LoadWithEnv(path)
	if err != nil {
		panic(fmt.Sprintf("failed to load config with env from %s: %v", path, err))
	}
	if err := cfg.Unmarshal(target); err != nil {
		panic(fmt.Sprintf("failed to unmarshal config: %v", err))
	}
}
