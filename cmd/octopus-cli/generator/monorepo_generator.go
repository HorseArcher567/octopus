package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// InitMonorepo 初始化 monorepo 项目
func InitMonorepo(projectName, module, outputDir string) error {
	// 1. 设置默认值
	if module == "" {
		module = projectName
	}

	// 项目根目录
	projectDir := filepath.Join(outputDir, projectName)

	// 检查目录是否存在
	if _, err := os.Stat(projectDir); !os.IsNotExist(err) {
		return fmt.Errorf("directory %s already exists", projectDir)
	}

	// 2. 创建 monorepo 目录结构
	if err := createMonorepoDirs(projectDir); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// 3. 生成 monorepo 基础文件
	data := MonorepoData{
		ProjectName: projectName,
		Module:      module,
	}

	// 生成 README.md
	if err := generateMonorepoReadme(projectDir, data); err != nil {
		return fmt.Errorf("failed to generate README.md: %w", err)
	}

	// 生成根 Makefile
	if err := generateMonorepoMakefile(projectDir, data); err != nil {
		return fmt.Errorf("failed to generate Makefile: %w", err)
	}

	// 生成 .gitignore
	if err := generateMonorepoGitignore(projectDir); err != nil {
		return fmt.Errorf("failed to generate .gitignore: %w", err)
	}

	// 生成 go.work（可选）
	if err := generateGoWork(projectDir, data); err != nil {
		return fmt.Errorf("failed to generate go.work: %w", err)
	}

	// 4. 初始化 go.mod
	if err := initMonorepoGoMod(projectDir, module); err != nil {
		return fmt.Errorf("failed to init go.mod: %w", err)
	}

	fmt.Printf("✨ Monorepo project '%s' initialized!\n", projectName)
	fmt.Printf("\n📁 Project structure:\n")
	fmt.Printf("  %s/\n", projectName)
	fmt.Printf("    ├── apps/          (your applications)\n")
	fmt.Printf("    ├── proto/         (proto definitions)\n")
	fmt.Printf("    ├── pkg/           (shared packages)\n")
	fmt.Printf("    ├── go.mod\n")
	fmt.Printf("    ├── go.work\n")
	fmt.Printf("    ├── Makefile\n")
	fmt.Printf("    └── README.md\n")

	fmt.Printf("\n📝 Next steps:\n")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Printf("  octopus-cli add user --port 9001\n")
	fmt.Printf("  octopus-cli add order --port 9002\n")

	return nil
}

// AddApp 向 monorepo 添加新应用
func AddApp(appName string, port int, monorepoRoot string) error {
	// 1. 检查是否在 monorepo 根目录
	if !isMonorepoRoot(monorepoRoot) {
		return fmt.Errorf("not in a monorepo root directory (missing apps/ directory)")
	}

	// 2. 读取 go.mod 获取 module 名称
	module, err := getModuleName(monorepoRoot)
	if err != nil {
		return fmt.Errorf("failed to read module name: %w", err)
	}

	// 3. 转换服务名称
	appNameCamel := toCamelCase(appName)

	// 4. 应用目录
	appDir := filepath.Join(monorepoRoot, "apps", appName)

	// 检查应用是否已存在
	if _, err := os.Stat(appDir); !os.IsNotExist(err) {
		return fmt.Errorf("application %s already exists", appName)
	}

	// 5. 创建应用目录结构
	if err := createAppDirs(appDir); err != nil {
		return fmt.Errorf("failed to create app directories: %w", err)
	}

	// 6. 生成应用文件
	data := AppData{
		AppName:      appName,
		AppNameCamel: appNameCamel,
		Module:       module,
		Port:         port,
		ServiceName:  appName + "-service", // user -> user-service
	}

	// 生成 main.go
	if err := generateAppMain(appDir, data); err != nil {
		return fmt.Errorf("failed to generate main.go: %w", err)
	}

	// 生成 logic.go
	if err := generateAppLogic(appDir, data); err != nil {
		return fmt.Errorf("failed to generate logic.go: %w", err)
	}

	// 生成 server.go
	if err := generateAppServer(appDir, data); err != nil {
		return fmt.Errorf("failed to generate server.go: %w", err)
	}

	// 生成 config.yaml
	if err := generateAppConfig(appDir, data); err != nil {
		return fmt.Errorf("failed to generate config.yaml: %w", err)
	}

	// 7. 生成 proto 文件（proto/<app>.proto，生成的 *.pb.go 输出到 proto/<app>/）
	protoDir := filepath.Join(monorepoRoot, "proto")
	if err := os.MkdirAll(protoDir, 0755); err != nil {
		return fmt.Errorf("failed to create proto directory: %w", err)
	}

	if err := generateAppProto(protoDir, data); err != nil {
		return fmt.Errorf("failed to generate proto: %w", err)
	}

	// 8. 更新根 Makefile，添加新应用的构建目标
	if err := updateMonorepoMakefile(monorepoRoot, appName); err != nil {
		fmt.Printf("⚠️  Warning: failed to update Makefile: %v\n", err)
		fmt.Printf("   Please manually add build targets for %s\n", appName)
	}

	fmt.Printf("✨ Application '%s' added to monorepo!\n", appName)
	fmt.Printf("\n📁 Generated files:\n")
	fmt.Printf("  apps/%s/\n", appName)
	fmt.Printf("    ├── cmd/main.go\n")
	fmt.Printf("    ├── internal/\n")
	fmt.Printf("    │   ├── logic/%s.go\n", appName)
	fmt.Printf("    │   └── server/%s.go\n", appName)
	fmt.Printf("    └── etc/config.yaml\n")
	fmt.Printf("  proto/%s.proto\n", appName)

	fmt.Printf("\n📝 Next steps:\n")
	fmt.Printf("  make proto\n")
	fmt.Printf("  make build-%s\n", appName)
	fmt.Printf("  make run-%s\n", appName)

	return nil
}

// createMonorepoDirs 创建 monorepo 目录结构
func createMonorepoDirs(projectDir string) error {
	dirs := []string{
		"apps",
		"proto",
		"pkg/middleware",
		"pkg/utils",
		"pkg/errors",
		"scripts",
	}

	for _, dir := range dirs {
		path := filepath.Join(projectDir, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}

	return nil
}

// createAppDirs 创建应用目录结构
func createAppDirs(appDir string) error {
	dirs := []string{
		"cmd",
		"internal/logic",
		"internal/server",
		"etc",
	}

	for _, dir := range dirs {
		path := filepath.Join(appDir, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}

	return nil
}

// isMonorepoRoot 检查是否在 monorepo 根目录
func isMonorepoRoot(dir string) bool {
	appsDir := filepath.Join(dir, "apps")
	_, err := os.Stat(appsDir)
	return err == nil
}

// getModuleName 从 go.mod 读取 module 名称
func getModuleName(dir string) (string, error) {
	goModPath := filepath.Join(dir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimPrefix(line, "module "), nil
		}
	}

	return "", fmt.Errorf("module name not found in go.mod")
}

// initMonorepoGoMod 初始化 monorepo 的 go.mod
func initMonorepoGoMod(projectDir, module string) error {
	// 初始化 go.mod
	cmd := exec.Command("go", "mod", "init", module)
	cmd.Dir = projectDir
	if err := cmd.Run(); err != nil {
		return err
	}

	// 添加 octopus 依赖
	cmd = exec.Command("go", "get", "github.com/HorseArcher567/octopus@latest")
	cmd.Dir = projectDir
	if err := cmd.Run(); err != nil {
		return err
	}

	// 整理依赖
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = projectDir
	return cmd.Run()
}

// updateMonorepoMakefile 更新 Makefile，添加新应用的构建目标
func updateMonorepoMakefile(monorepoRoot, appName string) error {
	makefilePath := filepath.Join(monorepoRoot, "Makefile")

	// 读取现有的 Makefile
	content, err := os.ReadFile(makefilePath)
	if err != nil {
		return err
	}

	makefile := string(content)

	// 检查是否已经包含该应用
	if strings.Contains(makefile, "build-"+appName) {
		return nil // 已存在，不需要更新
	}

	// 添加构建和运行目标
	newTargets := fmt.Sprintf(`
# %s targets
build-%s: proto
	@echo "Building %s..."
	@go build -o bin/%s-service apps/%s/cmd/main.go
	@echo "✅ %s built"

run-%s: proto
	@echo "Starting %s..."
	@go run apps/%s/cmd/main.go
`, appName, appName, appName, appName, appName, appName, appName, appName, appName)

	// 追加到文件末尾
	makefile += newTargets

	// 写回文件
	return os.WriteFile(makefilePath, []byte(makefile), 0644)
}
