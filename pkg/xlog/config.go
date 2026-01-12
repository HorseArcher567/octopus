package xlog

// Config 日志配置
type Config struct {
	// Level 日志级别：debug/info/warn/error（默认 info）
	Level string `yaml:"level" json:"level" toml:"level"`

	// Format 日志格式：json/text（默认 text）
	Format string `yaml:"format" json:"format" toml:"format"`

	// AddSource 是否添加源码位置（文件名、行号）
	AddSource bool `yaml:"add_source" json:"add_source" toml:"add_source"`

	// Output 输出目标：stdout/stderr/文件路径（默认 stdout）
	// 文件输出自动启用按天轮转
	Output string `yaml:"output" json:"output" toml:"output"`

	// MaxAge 日志保留天数（仅文件输出有效），0 表示不删除
	MaxAge int `yaml:"max_age" json:"max_age" toml:"max_age"`
}
