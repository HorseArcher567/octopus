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

// Format identifies a configuration serialization format.
type Format string

const (
	FormatJSON Format = "json"
	FormatYAML Format = "yaml"
	FormatTOML Format = "toml"
	// FormatUnknown indicates that the format is unknown or could not be inferred.
	FormatUnknown Format = "unknown"
)

// parseFile parses config from a file using the provided format.
func parseFile(filepath string, format Format) (map[string]any, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return parse(data, format)
}

// parse parses config from raw bytes.
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

// parseJSON parses JSON data.
func parseJSON(data []byte) (map[string]any, error) {
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return result, nil
}

// parseYAML parses YAML data.
func parseYAML(data []byte) (map[string]any, error) {
	var result map[string]any
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	return result, nil
}

// parseTOML parses TOML data.
func parseTOML(data []byte) (map[string]any, error) {
	var result map[string]any
	if err := toml.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
	}
	return result, nil
}

// detectFormat infers the format from a file extension.
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

// marshal serializes config data into the requested format.
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

// writeFile writes config data to a file.
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
