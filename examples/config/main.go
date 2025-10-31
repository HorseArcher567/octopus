package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/HorseArcher567/octopus/pkg/config"
)

// Config 应用配置结构
type Config struct {
	App struct {
		Name string
		Port int
	}
	Database struct {
		Host string
		Port int
		User string
	}
}

func main() {
	// 解析命令行参数
	configFile := flag.String("f", "config.yaml", "配置文件路径")
	flag.Parse()

	// 一行代码加载并解析配置（类似 go-zero）
	// 支持环境变量替换: ${ENV_VAR} 或 ${ENV_VAR:default}
	os.Setenv("DB_HOST", "127.0.0.1")
	var c Config
	config.MustLoadWithEnvAndUnmarshal(*configFile, &c)

	// 直接使用配置
	fmt.Printf("Starting %s at port %d\n", c.App.Name, c.App.Port)
	fmt.Printf("Database: %s@%s:%d\n", c.Database.User, c.Database.Host, c.Database.Port)
}
