package api

import "time"

// ServerConfig 是 HTTP API 服务器配置。
//
// 示例配置:
// apiServer:
//
//	appName: my-service
//	host: 0.0.0.0
//	port: 8080
//	mode: release
//	enablePProf: true
//	readTimeout: 5s
//	writeTimeout: 10s
//	idleTimeout: 60s
type ServerConfig struct {
	// AppName 应用名称，用于日志等标识。
	AppName string `yaml:"appName" json:"appName" toml:"appName"`

	// Host 监听地址（如 0.0.0.0, 127.0.0.1）。
	Host string `yaml:"host" json:"host" toml:"host"`

	// Port 监听端口，大于 0 时才会启动 HTTP Server。
	Port int `yaml:"port" json:"port" toml:"port"`

	// Mode Gin 运行模式: debug / release。
	Mode string `yaml:"mode" json:"mode" toml:"mode"`

	// EnablePProf 是否启用 pprof 路由。
	EnablePProf bool `yaml:"enablePProf" json:"enablePProf" toml:"enablePProf"`

	// ReadTimeout 读超时时间。
	ReadTimeout time.Duration `yaml:"readTimeout" json:"readTimeout" toml:"readTimeout"`

	// WriteTimeout 写超时时间。
	WriteTimeout time.Duration `yaml:"writeTimeout" json:"writeTimeout" toml:"writeTimeout"`

	// IdleTimeout 空闲连接超时时间。
	IdleTimeout time.Duration `yaml:"idleTimeout" json:"idleTimeout" toml:"idleTimeout"`
}
