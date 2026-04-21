package config

import (
	"fmt"
	"maps"
	"os"
	"strings"
	"sync"

	"github.com/HorseArcher567/octopus/pkg/mapstruct"
)

// Config stores application configuration data.
type Config struct {
	data   map[string]any
	format Format // Preserved for config source format tracking. Decoding currently relies on the decoder's default tag behavior.
	mu     sync.RWMutex
}

// New creates an empty config container.
// The format defaults to FormatUnknown and is usually set during Load or LoadBytes.
func New() *Config {
	return &Config{
		data:   make(map[string]any),
		format: FormatUnknown,
	}
}

// Load reads config from a file and replaces the current contents.
func (c *Config) Load(filepath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	format := detectFormat(filepath)
	if format == FormatUnknown {
		return fmt.Errorf("cannot detect format from file extension: %s", filepath)
	}

	data, err := parseFile(filepath, format)
	if err != nil {
		return fmt.Errorf("failed to load config from file %s: %w", filepath, err)
	}

	c.format = format
	c.data = data
	return nil
}

// LoadBytes reads config from raw bytes and replaces the current contents.
func (c *Config) LoadBytes(data []byte, format Format) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	parsed, err := parse(data, format)
	if err != nil {
		return fmt.Errorf("failed to load config from bytes: %w", err)
	}

	c.format = format
	c.data = parsed
	return nil
}

// Set writes a config value. Dotted keys such as "database.host" are supported.
func (c *Config) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if strings.Contains(key, ".") {
		c.setNested(key, value)
	} else {
		c.data[key] = value
	}
}

// Get returns a config value.
func (c *Config) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if strings.Contains(key, ".") {
		return c.getNested(key)
	}

	val, exists := c.data[key]
	return val, exists
}

// GetString returns a string config value.
func (c *Config) GetString(key string) string {
	if val, ok := c.Get(key); ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// GetInt returns an integer config value.
func (c *Config) GetInt(key string) int {
	if val, ok := c.Get(key); ok {
		if i, ok := toInt(val); ok {
			return i
		}
	}
	return 0
}

// GetBool returns a boolean config value.
func (c *Config) GetBool(key string) bool {
	if val, ok := c.Get(key); ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// GetFloat returns a floating-point config value.
func (c *Config) GetFloat(key string) float64 {
	if val, ok := c.Get(key); ok {
		if f, ok := toFloat64(val); ok {
			return f
		}
	}
	return 0.0
}

// toInt converts a value to int when possible.
func toInt(val any) (int, bool) {
	switch v := val.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

// toFloat64 converts a value to float64 when possible.
func toFloat64(val any) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0.0, false
	}
}

// GetWithDefault returns a config value or a default.
func (c *Config) GetWithDefault(key string, defaultValue any) any {
	if val, ok := c.Get(key); ok {
		return val
	}
	return defaultValue
}

// GetStringWithDefault returns a string config value or a default.
func (c *Config) GetStringWithDefault(key string, defaultValue string) string {
	if val, ok := c.Get(key); ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

// GetIntWithDefault returns an integer config value or a default.
func (c *Config) GetIntWithDefault(key string, defaultValue int) int {
	if val, ok := c.Get(key); ok {
		if i, ok := toInt(val); ok {
			return i
		}
	}
	return defaultValue
}

// GetBoolWithDefault returns a boolean config value or a default.
func (c *Config) GetBoolWithDefault(key string, defaultValue bool) bool {
	if val, ok := c.Get(key); ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// GetFloatWithDefault returns a floating-point config value or a default.
func (c *Config) GetFloatWithDefault(key string, defaultValue float64) float64 {
	if val, ok := c.Get(key); ok {
		if f, ok := toFloat64(val); ok {
			return f
		}
	}
	return defaultValue
}

// GetStringSlice returns a string slice config value.
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
		// Handle []string directly as a less common fast path.
		if slice, ok := val.([]string); ok {
			return slice
		}
	}
	return []string{}
}

// GetIntSlice returns an int slice config value.
func (c *Config) GetIntSlice(key string) []int {
	if val, ok := c.Get(key); ok {
		if slice, ok := val.([]any); ok {
			result := make([]int, 0, len(slice))
			for _, item := range slice {
				if i, ok := toInt(item); ok {
					result = append(result, i)
				}
			}
			return result
		}
		// Handle []int directly as a less common fast path.
		if slice, ok := val.([]int); ok {
			return slice
		}
	}
	return []int{}
}

// GetSlice returns a slice config value.
func (c *Config) GetSlice(key string) []any {
	if val, ok := c.Get(key); ok {
		if slice, ok := val.([]any); ok {
			return slice
		}
	}
	return []any{}
}

// Has reports whether a config key exists.
func (c *Config) Has(key string) bool {
	_, exists := c.Get(key)
	return exists
}

// GetAll returns all top-level config data.
func (c *Config) GetAll() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// A shallow copy protects the top-level map from direct external mutation.
	// Nested maps still share references, which is acceptable for this package's usage.
	return maps.Clone(c.data)
}

// GetSection returns a top-level config section.
func (c *Config) GetSection(key string) map[string]any {
	if val, ok := c.Get(key); ok {
		if m, ok := val.(map[string]any); ok {
			return maps.Clone(m)
		}
	}
	return make(map[string]any)
}

// Unmarshal decodes config into target.
func (c *Config) Unmarshal(target interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	decoder := mapstruct.New()
	return decoder.Decode(c.data, target)
}

// UnmarshalStrict decodes the entire config into target using strict field decoding.
func (c *Config) UnmarshalStrict(target interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	decoder := mapstruct.New().WithStrictMode(true)
	return decoder.Decode(c.data, target)
}

// UnmarshalKey decodes a config subtree into target.
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

// UnmarshalKeyStrict decodes a config subtree with strict field decoding.
func (c *Config) UnmarshalKeyStrict(key string, target interface{}) error {
	val, ok := c.Get(key)
	if !ok {
		return fmt.Errorf("config key '%s' not found", key)
	}

	dataMap, ok := val.(map[string]any)
	if !ok {
		return fmt.Errorf("config key '%s' cannot be unmarshaled to struct (type: %T, expected: map/object)", key, val)
	}

	decoder := mapstruct.New().WithStrictMode(true)
	if err := decoder.Decode(dataMap, target); err != nil {
		return fmt.Errorf("failed to unmarshal config key '%s': %w", key, err)
	}
	return nil
}

// Clear removes all config data.
func (c *Config) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]any)
}

// WriteToFile writes config data to a file.
// The output format is inferred from the file extension.
func (c *Config) WriteToFile(filepath string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return writeFile(filepath, c.data)
}

// getNested returns a nested config value.
func (c *Config) getNested(key string) (any, bool) {
	keys := strings.Split(key, ".")
	current := c.data

	// Walk all keys except the last one.
	for _, k := range keys[:len(keys)-1] {
		val, exists := current[k]
		if !exists {
			return nil, false
		}

		// Convert to a map before continuing downward.
		m, ok := val.(map[string]any)
		if !ok {
			return nil, false
		}
		current = m
	}

	// Read the final key.
	val, exists := current[keys[len(keys)-1]]
	return val, exists
}

// setNested writes a nested config value.
func (c *Config) setNested(key string, value any) {
	keys := strings.Split(key, ".")
	current := c.data

	for i, k := range keys {
		// If this is the last key, assign the value.
		if i == len(keys)-1 {
			current[k] = value
			return
		}

		// Otherwise create or reuse the next nested map.
		if val, exists := current[k]; exists {
			if m, ok := val.(map[string]any); ok {
				current = m
			} else {
				// If the existing value is not a map, replace it with one.
				newMap := make(map[string]any)
				current[k] = newMap
				current = newMap
			}
		} else {
			// Create a new map when the key does not exist.
			newMap := make(map[string]any)
			current[k] = newMap
			current = newMap
		}
	}
}

// ============================================================================
// Environment variable expansion
// ============================================================================

// replaceEnvVars recursively expands environment variables.
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

// expandEnvVar expands environment variables.
// Supported forms: ${ENV_VAR} and ${ENV_VAR:default_value}.
func expandEnvVar(value string) string {
	if !strings.Contains(value, "${") {
		return value
	}

	var builder strings.Builder
	builder.Grow(len(value) * 2) // Pre-grow to reduce reallocations.

	start := 0
	for {
		startIdx := strings.Index(value[start:], "${")
		if startIdx == -1 {
			builder.WriteString(value[start:])
			break
		}
		startIdx += start

		// Write the content before ${.
		builder.WriteString(value[start:startIdx])

		endIdx := strings.Index(value[startIdx:], "}")
		if endIdx == -1 {
			builder.WriteString(value[startIdx:])
			break
		}
		endIdx += startIdx

		// Parse the environment variable name and default value.
		envExpr := value[startIdx+2 : endIdx]
		envName := envExpr
		defaultValue := ""

		if colonIdx := strings.Index(envExpr, ":"); colonIdx != -1 {
			envName = envExpr[:colonIdx]
			defaultValue = envExpr[colonIdx+1:]
		}

		// Resolve the environment variable value.
		envValue := os.Getenv(envName)
		if envValue == "" {
			envValue = defaultValue
		}

		// Write the expanded value.
		builder.WriteString(envValue)
		start = endIdx + 1
	}

	return builder.String()
}
