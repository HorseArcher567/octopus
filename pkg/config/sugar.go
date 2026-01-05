package config

import (
	"fmt"
)

// ============================================================================
// 包级加载函数 - 便捷接口
// ============================================================================

// Load 加载单个配置文件（自动识别格式）
func Load(path string) (*Config, error) {
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

// LoadWithEnv 加载配置并支持环境变量替换
// 环境变量格式: ${ENV_VAR} 或 ${ENV_VAR:default_value}
func LoadWithEnv(path string) (*Config, error) {
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

// MustLoadWithEnv 加载配置并替换环境变量，失败时 panic
func MustLoadWithEnv(path string) *Config {
	cfg, err := LoadWithEnv(path)
	if err != nil {
		panic(fmt.Sprintf("failed to load config with env from %s: %v", path, err))
	}
	return cfg
}

// MustUnmarshal 加载配置并直接解析到结构体，失败时 panic
// 这是最便捷的方式，适合在 main 函数中使用
func MustUnmarshal(path string, target interface{}) {
	cfg := MustLoad(path)
	if err := cfg.Unmarshal(target); err != nil {
		panic(fmt.Sprintf("failed to unmarshal config: %v", err))
	}
}

// MustUnmarshalWithEnv 加载配置（支持环境变量）并解析到结构体，失败时 panic
func MustUnmarshalWithEnv(path string, target interface{}) {
	cfg := MustLoadWithEnv(path)
	if err := cfg.Unmarshal(target); err != nil {
		panic(fmt.Sprintf("failed to unmarshal config: %v", err))
	}
}
