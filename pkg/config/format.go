package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// Format 配置文件格式
type Format string

const (
	FormatJSON Format = "json"
	FormatYAML Format = "yaml"
	FormatTOML Format = "toml"
	// FormatUnknown 表示未知或无法从上下文推断的格式
	// 在需要自动检测的场景下，通常会先返回 FormatUnknown，然后由调用方决定如何处理
	FormatUnknown Format = "unknown"
)

// parseFile 从文件解析配置
// 注意：调用方需要在调用前确定并传入正确的 format
func parseFile(filepath string, format Format) (map[string]any, error) {
	// 读取文件内容
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return parse(data, format)
}

// parse 解析字节流
func parse(data []byte, format Format) (map[string]any, error) {
	switch format {
	case FormatJSON:
		return parseJSON(data)
	case FormatYAML:
		return parseYAML(data)
	case FormatTOML:
		return parseTOML(data)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// parseJSON 解析JSON格式
func parseJSON(data []byte) (map[string]any, error) {
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return result, nil
}

// parseYAML 解析YAML格式
func parseYAML(data []byte) (map[string]any, error) {
	var result map[string]any
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	return result, nil
}

// parseTOML 解析TOML格式
func parseTOML(data []byte) (map[string]any, error) {
	var result map[string]any
	if err := toml.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
	}
	return result, nil
}

// detectFormat 根据文件扩展名检测格式
func detectFormat(filename string) Format {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".json":
		return FormatJSON
	case ".yaml", ".yml":
		return FormatYAML
	case ".toml":
		return FormatTOML
	default:
		return FormatUnknown
	}
}

// marshal 将map序列化为指定格式的字节流
func marshal(data map[string]any, format Format) ([]byte, error) {
	switch format {
	case FormatJSON:
		return json.MarshalIndent(data, "", "  ")
	case FormatYAML:
		return yaml.Marshal(data)
	case FormatTOML:
		var buf strings.Builder
		encoder := toml.NewEncoder(&buf)
		if err := encoder.Encode(data); err != nil {
			return nil, fmt.Errorf("failed to marshal TOML: %w", err)
		}
		return []byte(buf.String()), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// writeFile 将配置写入文件
func writeFile(filepath string, data map[string]any) error {
	format := detectFormat(filepath)
	if format == FormatUnknown {
		return fmt.Errorf("cannot detect format from file extension: %s", filepath)
	}

	bytes, err := marshal(data, format)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, bytes, 0644)
}
