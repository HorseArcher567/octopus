package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Generate 生成服务代码
func Generate(serviceName, module, outputDir string) error {
	// 1. 设置默认值
	if module == "" {
		module = serviceName
	}

	// 服务名称转换（user-service -> UserService）
	serviceNameCamel := toCamelCase(serviceName)

	// 项目根目录
	projectDir := filepath.Join(outputDir, serviceName)

	// 2. 创建目录结构
	if err := createDirs(projectDir); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// 3. 生成文件
	data := TemplateData{
		ServiceName:      serviceName,
		ServiceNameCamel: serviceNameCamel,
		Module:           module,
	}

	// 生成 main.go
	if err := generateMain(projectDir, data); err != nil {
		return fmt.Errorf("failed to generate main.go: %w", err)
	}

	// 不再生成 internal/config/config.go（改到 server 包定义）

	// 生成 server.go
	if err := generateServer(projectDir, data); err != nil {
		return fmt.Errorf("failed to generate server.go: %w", err)
	}

	// 生成 logic.go
	if err := generateLogic(projectDir, data); err != nil {
		return fmt.Errorf("failed to generate logic.go: %w", err)
	}

	// 生成 proto 文件
	if err := generateProto(projectDir, data); err != nil {
		return fmt.Errorf("failed to generate proto: %w", err)
	}

	// 生成 config.yaml
	if err := generateConfigYaml(projectDir, data); err != nil {
		return fmt.Errorf("failed to generate config.yaml: %w", err)
	}

	// 生成 Makefile
	if err := generateMakefile(projectDir, data); err != nil {
		return fmt.Errorf("failed to generate Makefile: %w", err)
	}

	// 生成 .gitignore
	if err := generateGitignore(projectDir); err != nil {
		return fmt.Errorf("failed to generate .gitignore: %w", err)
	}

	// 4. 初始化 go.mod
	if err := initGoMod(projectDir, module); err != nil {
		return fmt.Errorf("failed to init go.mod: %w", err)
	}

	fmt.Printf("✨ Generated files:\n")
	fmt.Printf("  📁 %s/\n", serviceName)
	fmt.Printf("    ├── cmd/main.go\n")
	fmt.Printf("    ├── internal/\n")
	// 移除 internal/config 展示
	fmt.Printf("    │   ├── logic/logic.go\n")
	fmt.Printf("    │   └── server/server.go\n")
	fmt.Printf("    ├── proto/%s.proto\n", serviceName)
	fmt.Printf("    ├── etc/config.yaml\n")
	fmt.Printf("    ├── go.mod\n")
	fmt.Printf("    └── Makefile\n")

	return nil
}

// createDirs 创建目录结构
func createDirs(projectDir string) error {
	dirs := []string{
		"cmd",
		// 移除 internal/config 目录
		"internal/logic",
		"internal/server",
		"proto",
		"etc",
	}

	for _, dir := range dirs {
		path := filepath.Join(projectDir, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}

	return nil
}

// toCamelCase 转换为驼峰命名
func toCamelCase(s string) string {
	// user-service -> UserService
	parts := strings.Split(s, "-")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}
