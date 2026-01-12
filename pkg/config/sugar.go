package config

import (
	"fmt"
)

// ============================================================================
// 包级加载函数 - 便捷接口
// ============================================================================

// Load 加载单个配置文件（自动识别格式），默认支持环境变量替换
// 环境变量格式: ${ENV_VAR} 或 ${ENV_VAR:default_value}
// 这是最常用的加载方式，适合大多数场景
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

// LoadWithoutEnv 加载配置文件但不替换环境变量
// 适用于不需要环境变量替换的场景
func LoadWithoutEnv(path string) (*Config, error) {
	cfg := New()
	if err := cfg.Load(path); err != nil {
		return nil, err
	}
	return cfg, nil
}

// LoadFromBytes 从字节流加载配置
func LoadFromBytes(data []byte, format Format) (*Config, error) {
	cfg := New()
	if err := cfg.LoadBytes(data, format); err != nil {
		return nil, err
	}
	return cfg, nil
}

// ============================================================================
// Must* 系列方法 - 失败时 panic，适用于启动阶段
// ============================================================================

// MustLoad 加载配置文件（默认支持环境变量替换），失败时 panic
// 适用于程序启动阶段，配置加载失败时程序无法继续运行
func MustLoad(path string) *Config {
	cfg, err := Load(path)
	if err != nil {
		panic(fmt.Errorf("config: failed to load config from %s: %w", path, err))
	}
	return cfg
}

// MustLoadWithoutEnv 加载配置文件但不替换环境变量，失败时 panic
// 适用于程序启动阶段，配置加载失败时程序无法继续运行
func MustLoadWithoutEnv(path string) *Config {
	cfg, err := LoadWithoutEnv(path)
	if err != nil {
		panic(fmt.Errorf("config: failed to load config from %s: %w", path, err))
	}
	return cfg
}

// MustUnmarshal 加载配置（默认支持环境变量替换）并直接解析到结构体，失败时 panic
// 这是最便捷的方式，适合在 main 函数中使用
func MustUnmarshal(path string, target interface{}) {
	cfg := MustLoad(path)
	if err := cfg.Unmarshal(target); err != nil {
		panic(fmt.Errorf("config: failed to unmarshal config from %s: %w", path, err))
	}
}

// MustUnmarshalWithoutEnv 加载配置（不支持环境变量替换）并解析到结构体，失败时 panic
// 适用于不需要环境变量替换的场景
func MustUnmarshalWithoutEnv(path string, target interface{}) {
	cfg := MustLoadWithoutEnv(path)
	if err := cfg.Unmarshal(target); err != nil {
		panic(fmt.Errorf("config: failed to unmarshal config from %s: %w", path, err))
	}
}
