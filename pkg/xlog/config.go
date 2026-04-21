// Package xlog wraps slog with:
//   - context propagation helpers
//   - explicit logger ownership and Close lifecycle
//   - optional daily file rotation
package xlog

// Config controls logger construction behavior.
type Config struct {
	// Name is the logical logger name used in configuration and dependency wiring.
	Name string `yaml:"name" json:"name" toml:"name"`

	// Level is the minimum enabled level. Supported values: debug/info/warn/error.
	// Empty means info.
	Level string `yaml:"level" json:"level" toml:"level"`

	// Format is the output encoder. Supported values: text/json.
	// Empty means text.
	Format string `yaml:"format" json:"format" toml:"format"`

	// AddSource includes source location fields in each record.
	AddSource bool `yaml:"addSource" json:"addSource" toml:"addSource"`

	// Output selects the sink target:
	//   - "stdout"
	//   - "stderr"
	//   - file path (enables daily rotation)
	// Empty means stdout.
	Output string `yaml:"output" json:"output" toml:"output"`

	// MaxAge is the retention window in days for rotated files.
	// It is ignored for stdout/stderr sinks. Zero disables deletion.
	MaxAge int `yaml:"maxAge" json:"maxAge" toml:"maxAge"`
}
