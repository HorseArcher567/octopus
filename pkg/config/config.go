package config

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/HorseArcher567/octopus/pkg/mapstruct"
)

// Config 配置管理器
type Config struct {
	data map[string]any
	mu   sync.RWMutex
}

// New 创建一个新的配置管理器
func New() *Config {
	return &Config{
		data: make(map[string]any),
	}
}

// Load 从文件加载配置并合并到现有配置
// 如果需要完全替换配置，使用 LoadAndReplace
func (c *Config) Load(filepath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := parseFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to load config from file %s: %w", filepath, err)
	}

	c.data = mergeMaps(c.data, data)
	return nil
}

// LoadAndReplace 从文件加载配置并完全替换现有配置
func (c *Config) LoadAndReplace(filepath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := parseFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to load config from file %s: %w", filepath, err)
	}

	c.data = data
	return nil
}

// LoadBytes 从字节流加载配置并合并到现有配置
func (c *Config) LoadBytes(data []byte, format Format) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	parsed, err := parse(data, format)
	if err != nil {
		return fmt.Errorf("failed to load config from bytes: %w", err)
	}

	c.data = mergeMaps(c.data, parsed)
	return nil
}

// LoadBytesAndReplace 从字节流加载配置并完全替换现有配置
func (c *Config) LoadBytesAndReplace(data []byte, format Format) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	parsed, err := parse(data, format)
	if err != nil {
		return fmt.Errorf("failed to load config from bytes: %w", err)
	}

	c.data = parsed
	return nil
}

// Merge 合并配置（新配置会覆盖旧配置）
func (c *Config) Merge(other *Config) {
	c.mu.Lock()
	defer c.mu.Unlock()

	other.mu.RLock()
	defer other.mu.RUnlock()

	c.data = mergeMaps(c.data, other.data)
}

// MergeMap 合并map配置
func (c *Config) MergeMap(data map[string]any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = mergeMaps(c.data, data)
}

// Set 设置配置值，支持路径访问（如 "database.host"）
func (c *Config) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if strings.Contains(key, ".") {
		c.setNested(key, value)
	} else {
		c.data[key] = value
	}
}

// Get 获取配置值
func (c *Config) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if strings.Contains(key, ".") {
		return c.getNested(key)
	}

	val, exists := c.data[key]
	return val, exists
}

// GetString 获取字符串配置值
func (c *Config) GetString(key string) string {
	if val, ok := c.Get(key); ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// GetInt 获取整数配置值
func (c *Config) GetInt(key string) int {
	if val, ok := c.Get(key); ok {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return 0
}

// GetBool 获取布尔配置值
func (c *Config) GetBool(key string) bool {
	if val, ok := c.Get(key); ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// GetFloat 获取浮点数配置值
func (c *Config) GetFloat(key string) float64 {
	if val, ok := c.Get(key); ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return 0.0
}

// GetWithDefault 获取配置值，如果不存在则返回默认值
func (c *Config) GetWithDefault(key string, defaultValue any) any {
	if val, ok := c.Get(key); ok {
		return val
	}
	return defaultValue
}

// GetStringWithDefault 获取字符串配置值，如果不存在则返回默认值
func (c *Config) GetStringWithDefault(key string, defaultValue string) string {
	if val, ok := c.Get(key); ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

// GetIntWithDefault 获取整数配置值，如果不存在则返回默认值
func (c *Config) GetIntWithDefault(key string, defaultValue int) int {
	if val, ok := c.Get(key); ok {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return defaultValue
}

// GetBoolWithDefault 获取布尔配置值，如果不存在则返回默认值
func (c *Config) GetBoolWithDefault(key string, defaultValue bool) bool {
	if val, ok := c.Get(key); ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// GetFloatWithDefault 获取浮点数配置值，如果不存在则返回默认值
func (c *Config) GetFloatWithDefault(key string, defaultValue float64) float64 {
	if val, ok := c.Get(key); ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return defaultValue
}

// GetStringSlice 获取字符串切片配置值
func (c *Config) GetStringSlice(key string) []string {
	if val, ok := c.Get(key); ok {
		if slice, ok := val.([]any); ok {
			result := make([]string, 0, len(slice))
			for _, item := range slice {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
		// 如果是字符串切片类型（虽然少见）
		if slice, ok := val.([]string); ok {
			return slice
		}
	}
	return []string{}
}

// GetIntSlice 获取整数切片配置值
func (c *Config) GetIntSlice(key string) []int {
	if val, ok := c.Get(key); ok {
		if slice, ok := val.([]any); ok {
			result := make([]int, 0, len(slice))
			for _, item := range slice {
				switch v := item.(type) {
				case int:
					result = append(result, v)
				case int64:
					result = append(result, int(v))
				case float64:
					result = append(result, int(v))
				}
			}
			return result
		}
		// 如果是整数切片类型（虽然少见）
		if slice, ok := val.([]int); ok {
			return slice
		}
	}
	return []int{}
}

// GetSlice 获取任意类型的切片配置值
func (c *Config) GetSlice(key string) []any {
	if val, ok := c.Get(key); ok {
		if slice, ok := val.([]any); ok {
			return slice
		}
	}
	return []any{}
}

// Has 检查配置项是否存在
func (c *Config) Has(key string) bool {
	_, exists := c.Get(key)
	return exists
}

// GetAll 获取所有配置数据
func (c *Config) GetAll() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 深拷贝以防止外部修改
	return copyMap(c.data)
}

// GetSection 获取配置的某个段落
func (c *Config) GetSection(key string) map[string]any {
	if val, ok := c.Get(key); ok {
		if m, ok := val.(map[string]any); ok {
			return copyMap(m)
		}
	}
	return make(map[string]any)
}

// Unmarshal 将配置解码到指定的结构体
func (c *Config) Unmarshal(target interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	decoder := mapstruct.New()
	return decoder.Decode(c.data, target)
}

// UnmarshalKey 将指定key的配置解码到结构体
func (c *Config) UnmarshalKey(key string, target interface{}) error {
	val, ok := c.Get(key)
	if !ok {
		return fmt.Errorf("config key '%s' not found", key)
	}

	dataMap, ok := val.(map[string]any)
	if !ok {
		return fmt.Errorf("config key '%s' cannot be unmarshaled to struct (type: %T, expected: map/object)", key, val)
	}

	decoder := mapstruct.New()
	if err := decoder.Decode(dataMap, target); err != nil {
		return fmt.Errorf("failed to unmarshal config key '%s': %w", key, err)
	}
	return nil
}

// UnmarshalWithDecoder 使用自定义decoder解码
func (c *Config) UnmarshalWithDecoder(decoder *mapstruct.Decoder, target interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return decoder.Decode(c.data, target)
}

// Clear 清空所有配置
func (c *Config) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]any)
}

// WriteToFile 将配置导出到文件
// 根据文件扩展名自动选择格式（.json/.yaml/.yml/.toml）
// 适用于调试、检查配置合并结果、保存修改后的配置等场景
func (c *Config) WriteToFile(filepath string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return writeFile(filepath, c.data)
}

// getNested 获取嵌套的配置值
func (c *Config) getNested(key string) (any, bool) {
	keys := strings.Split(key, ".")
	current := c.data

	// 遍历除最后一个key外的所有key
	for _, k := range keys[:len(keys)-1] {
		val, exists := current[k]
		if !exists {
			return nil, false
		}

		// 尝试转换为map以继续向下查找
		m, ok := val.(map[string]any)
		if !ok {
			return nil, false
		}
		current = m
	}

	// 获取最后一个key的值
	val, exists := current[keys[len(keys)-1]]
	return val, exists
}

// setNested 设置嵌套的配置值
func (c *Config) setNested(key string, value any) {
	keys := strings.Split(key, ".")
	current := c.data

	for i, k := range keys {
		// 如果是最后一个key，设置值
		if i == len(keys)-1 {
			current[k] = value
			return
		}

		// 否则继续向下创建或获取map
		if val, exists := current[k]; exists {
			if m, ok := val.(map[string]any); ok {
				current = m
			} else {
				// 如果存在但不是map，覆盖为map
				newMap := make(map[string]any)
				current[k] = newMap
				current = newMap
			}
		} else {
			// 不存在则创建新map
			newMap := make(map[string]any)
			current[k] = newMap
			current = newMap
		}
	}
}

// mergeMaps 合并两个map（递归合并）
func mergeMaps(dst, src map[string]any) map[string]any {
	result := copyMap(dst)

	for key, srcVal := range src {
		if dstVal, exists := result[key]; exists {
			// 如果两边都是map，递归合并
			if dstMap, dstOk := dstVal.(map[string]any); dstOk {
				if srcMap, srcOk := srcVal.(map[string]any); srcOk {
					result[key] = mergeMaps(dstMap, srcMap)
					continue
				}
			}
		}
		// 否则直接覆盖
		result[key] = srcVal
	}

	return result
}

// copyMap 深拷贝map
func copyMap(src map[string]any) map[string]any {
	dst := make(map[string]any, len(src))

	for key, val := range src {
		if m, ok := val.(map[string]any); ok {
			dst[key] = copyMap(m)
		} else {
			dst[key] = val
		}
	}

	return dst
}

// ============================================================================
// 环境变量处理
// ============================================================================

// replaceEnvVars 递归替换环境变量
func replaceEnvVars(data map[string]any) {
	for key, val := range data {
		switch v := val.(type) {
		case string:
			data[key] = expandEnvVar(v)
		case map[string]any:
			replaceEnvVars(v)
		case []any:
			for i, item := range v {
				if str, ok := item.(string); ok {
					v[i] = expandEnvVar(str)
				} else if m, ok := item.(map[string]any); ok {
					replaceEnvVars(m)
				}
			}
		}
	}
}

// expandEnvVar 展开环境变量
// 支持格式: ${ENV_VAR} 或 ${ENV_VAR:default_value}
func expandEnvVar(value string) string {
	if !strings.Contains(value, "${") {
		return value
	}

	result := value
	start := 0
	for {
		startIdx := strings.Index(result[start:], "${")
		if startIdx == -1 {
			break
		}
		startIdx += start

		endIdx := strings.Index(result[startIdx:], "}")
		if endIdx == -1 {
			break
		}
		endIdx += startIdx

		// 提取环境变量名和默认值
		envExpr := result[startIdx+2 : endIdx]
		envName := envExpr
		defaultValue := ""

		if colonIdx := strings.Index(envExpr, ":"); colonIdx != -1 {
			envName = envExpr[:colonIdx]
			defaultValue = envExpr[colonIdx+1:]
		}

		// 获取环境变量值
		envValue := os.Getenv(envName)
		if envValue == "" {
			envValue = defaultValue
		}

		// 替换
		result = result[:startIdx] + envValue + result[endIdx+1:]
		start = startIdx + len(envValue)
	}

	return result
}
