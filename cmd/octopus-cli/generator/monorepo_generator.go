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

	// 生成 go.work
	if err := generateGoWork(projectDir, data); err != nil {
		return fmt.Errorf("failed to generate go.work: %w", err)
	}

	// 4. 初始化 proto module（proto/ 目录有自己的 go.mod）
	protoDir := filepath.Join(projectDir, "proto")
	if err := initProtoGoMod(protoDir, projectName); err != nil {
		return fmt.Errorf("failed to init proto go.mod: %w", err)
	}

	fmt.Printf("✨ Monorepo project '%s' initialized!\n", projectName)
	fmt.Printf("\n📁 Project structure:\n")
	fmt.Printf("  %s/\n", projectName)
	fmt.Printf("    ├── apps/          (your applications)\n")
	fmt.Printf("    ├── proto/         (proto module with go.mod)\n")
	fmt.Printf("    ├── pkg/           (shared packages)\n")
	fmt.Printf("    ├── go.work        (manages all modules)\n")
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

	// 2. 读取 proto go.mod 获取 proto module 名称
	protoGoModPath := filepath.Join(monorepoRoot, "proto", "go.mod")
	protoModule, err := getModuleNameFromFile(protoGoModPath)
	if err != nil {
		return fmt.Errorf("failed to read proto module name: %w", err)
	}

	// 3. 从 proto module 名称提取项目名称（去掉 /proto 后缀）
	// protoModule 格式：my-project/proto
	// projectName 应该是：my-project
	projectName := strings.TrimSuffix(protoModule, "/proto")

	// 4. 转换服务名称
	appNameCamel := toCamelCase(appName)

	// 5. 应用目录
	appDir := filepath.Join(monorepoRoot, "apps", appName)

	// 检查应用是否已存在
	if _, err := os.Stat(appDir); !os.IsNotExist(err) {
		return fmt.Errorf("application %s already exists", appName)
	}

	// 6. 创建应用目录结构
	if err := createAppDirs(appDir); err != nil {
		return fmt.Errorf("failed to create app directories: %w", err)
	}

	// 7. 为 app 创建独立的 go.mod（使用简化的路径）
	appModule := fmt.Sprintf("%s/apps/%s", projectName, appName)
	if err := initAppGoMod(appDir, appModule, protoModule); err != nil {
		return fmt.Errorf("failed to init app go.mod: %w", err)
	}

	// 8. 生成应用文件
	data := AppData{
		AppName:      appName,
		AppNameCamel: appNameCamel,
		Module:       appModule,   // 使用 app 自己的 module
		RootModule:   protoModule, // proto module，用于导入 proto 代码
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

	// 8. 生成 proto 文件（proto/<app>.proto，生成的 *.pb.go 输出到 proto/<app>/）
	protoDir := filepath.Join(monorepoRoot, "proto")
	if err := os.MkdirAll(protoDir, 0755); err != nil {
		return fmt.Errorf("failed to create proto directory: %w", err)
	}

	if err := generateAppProto(protoDir, data); err != nil {
		return fmt.Errorf("failed to generate proto: %w", err)
	}

	// 9. 更新 go.work，添加新的 app module（proto 已经在 go.work 中）
	if err := updateGoWork(monorepoRoot, appName); err != nil {
		fmt.Printf("⚠️  Warning: failed to update go.work: %v\n", err)
		fmt.Printf("   Please manually add apps/%s to go.work\n", appName)
	}

	// 10. 更新根 Makefile，添加新应用的构建目标
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
	fmt.Printf("    ├── etc/config.yaml\n")
	fmt.Printf("    └── go.mod          (independent module)\n")
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

// getModuleNameFromFile 从指定的 go.mod 文件读取 module 名称
func getModuleNameFromFile(goModPath string) (string, error) {
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

// initProtoGoMod 初始化 proto module 的 go.mod
func initProtoGoMod(protoDir, projectName string) error {
	// proto module 名称：{projectName}/proto（使用简化的路径，不包含 GitHub 路径）
	protoModule := fmt.Sprintf("%s/proto", projectName)

	// 初始化 go.mod
	cmd := exec.Command("go", "mod", "init", protoModule)
	cmd.Dir = protoDir
	if err := cmd.Run(); err != nil {
		return err
	}

	// 添加 grpc 依赖（proto 生成的代码需要）
	cmd = exec.Command("go", "get", "google.golang.org/grpc@latest")
	cmd.Dir = protoDir
	if err := cmd.Run(); err != nil {
		return err
	}

	// 添加 protobuf 依赖（proto 生成的代码需要）
	cmd = exec.Command("go", "get", "google.golang.org/protobuf@latest")
	cmd.Dir = protoDir
	if err := cmd.Run(); err != nil {
		return err
	}

	// 整理依赖
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = protoDir
	return cmd.Run()
}

// initAppGoMod 为 app 初始化独立的 go.mod
func initAppGoMod(appDir, appModule, protoModule string) error {
	// 初始化 go.mod
	cmd := exec.Command("go", "mod", "init", appModule)
	cmd.Dir = appDir
	if err := cmd.Run(); err != nil {
		return err
	}

	// 添加 octopus 依赖
	cmd = exec.Command("go", "get", "github.com/HorseArcher567/octopus@latest")
	cmd.Dir = appDir
	if err := cmd.Run(); err != nil {
		return err
	}

	// 添加 grpc 相关依赖
	cmd = exec.Command("go", "get", "google.golang.org/grpc@latest")
	cmd.Dir = appDir
	if err := cmd.Run(); err != nil {
		return err
	}

	// 添加 protobuf 依赖
	cmd = exec.Command("go", "get", "google.golang.org/protobuf@latest")
	cmd.Dir = appDir
	if err := cmd.Run(); err != nil {
		return err
	}

	// 手动添加 proto module 依赖（用于访问 proto 生成的代码）
	// 在 workspace 模式下，直接添加 require，go.work 会自动处理本地路径
	goModPath := filepath.Join(appDir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return err
	}

	contentStr := string(content)

	// 检查是否已经包含 proto module
	if strings.Contains(contentStr, protoModule) {
		// 已经存在，直接运行 tidy
		cmd = exec.Command("go", "mod", "tidy")
		cmd.Dir = appDir
		return cmd.Run()
	}

	// 添加 proto module require
	// 简单方式：在文件末尾添加 require（go mod tidy 会自动整理）
	protoRequire := fmt.Sprintf("\nrequire %s v0.0.0\n", protoModule)
	newContent := contentStr + protoRequire

	if err := os.WriteFile(goModPath, []byte(newContent), 0644); err != nil {
		return err
	}

	// 整理依赖（go.work 会自动处理 proto module 的本地路径）
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = appDir
	return cmd.Run()
}

// updateGoWork 更新 go.work，添加新的 app module
func updateGoWork(monorepoRoot, appName string) error {
	goWorkPath := filepath.Join(monorepoRoot, "go.work")

	// 读取现有的 go.work
	content, err := os.ReadFile(goWorkPath)
	if err != nil {
		return err
	}

	goWork := string(content)

	// 检查是否已经包含该 app
	appPath := fmt.Sprintf("./apps/%s", appName)
	if strings.Contains(goWork, appPath) {
		return nil // 已存在，不需要更新
	}

	// 解析 go.work 内容
	lines := strings.Split(goWork, "\n")
	var newLines []string
	inUseBlock := false
	added := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 检测 use ( 块开始
		if trimmed == "use (" {
			inUseBlock = true
			newLines = append(newLines, line)
			continue
		}

		// 在 use 块内，跳过 "." 引用（根目录没有 go.mod）
		if inUseBlock && (trimmed == "." || trimmed == "./") {
			continue
		}

		// 在 use 块内，找到 ) 之前插入新 app
		if inUseBlock && trimmed == ")" {
			// 在 ) 之前添加新 app
			newLines = append(newLines, fmt.Sprintf("\t%s", appPath))
			newLines = append(newLines, line)
			added = true
			inUseBlock = false
			continue
		}

		newLines = append(newLines, line)
	}

	// 如果没找到 use 块或没添加成功，重新创建
	if !added {
		newLines = []string{
			"go 1.21",
			"",
			"use (",
			"\t./proto",
			fmt.Sprintf("\t%s", appPath),
			")",
			"",
			"// Applications will be added automatically when you run: octopus-cli add <app-name>",
		}
	}

	// 写回文件
	return os.WriteFile(goWorkPath, []byte(strings.Join(newLines, "\n")), 0644)
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
build-%s: proto sync-deps
	@echo "Building %s..."
	@go build -o bin/%s-service apps/%s/cmd/main.go
	@echo "✅ %s built"

run-%s: proto sync-deps
	@echo "Starting %s..."
	@go run apps/%s/cmd/main.go
`, appName, appName, appName, appName, appName, appName, appName, appName, appName)

	// 追加到文件末尾
	makefile += newTargets

	// 写回文件
	return os.WriteFile(makefilePath, []byte(makefile), 0644)
}
