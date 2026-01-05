package generator

import (
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

// TemplateData 模板数据
type TemplateData struct {
	ServiceName      string // user-service
	ServiceNameCamel string // UserService
	Module           string // github.com/xxx/user-service
}

// generateMain 生成 main.go
func generateMain(projectDir string, data TemplateData) error {
	tmpl := `package main

import (
    "log"

    "{{.Module}}/internal/logic"
    "{{.Module}}/internal/server"
    "{{.Module}}/proto/pb"

    "google.golang.org/grpc"
    "github.com/HorseArcher567/octopus/pkg/config"
    "github.com/HorseArcher567/octopus/pkg/rpc"
)

func main() {
    // 1. 加载配置
    var cfg server.Config
    config.MustUnmarshalWithEnv("etc/config.yaml", &cfg)

	// 2. 创建 Logic
	logic := logic.NewLogic()

	// 3. 创建 Server
	srv := server.NewServer(logic)

	// 4. 创建 RPC Server（直接使用配置）
	cfg.Server.EnableReflection = cfg.Mode == "dev"
	rpcServer := rpc.NewServer(ctx, &cfg.Server)

	// 5. 注册服务（支持注册多个服务）
	rpcServer.RegisterService(func(s *grpc.Server) {
		pb.Register{{.ServiceNameCamel}}Server(s, srv)
	})
	
	// 如果有多个服务，可以继续注册：
	// rpcServer.RegisterService(func(s *grpc.Server) {
	//     pb.RegisterAnotherServiceServer(s, anotherSrv)
	// }, "AnotherService") // 可选：指定服务名用于健康检查

	// 6. 启动服务
	if err := rpcServer.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
`
	return writeFromTemplate(filepath.Join(projectDir, "cmd/main.go"), tmpl, data)
}

// generateConfig 生成 config.go
// （移除）不再生成 internal/config，配置定义改到 server 包中

// generateLogic 生成 logic.go
func generateLogic(projectDir string, data TemplateData) error {
	tmpl := `package logic

import (
	"context"

	"{{.Module}}/proto/pb"
)

// Logic 业务逻辑层
type Logic struct {
	// TODO: 添加依赖（数据库、缓存等）
}

// NewLogic 创建 Logic 实例
func NewLogic() *Logic {
	return &Logic{}
}

// SayHello 示例方法（实现业务逻辑）
func (l *Logic) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	// TODO: 实现你的业务逻辑
	
	return &pb.HelloResponse{
		Message: "Hello " + req.Name,
	}, nil
}
`
	return writeFromTemplate(filepath.Join(projectDir, "internal/logic/logic.go"), tmpl, data)
}

// generateServer 生成 server.go
func generateServer(projectDir string, data TemplateData) error {
	tmpl := `package server

import (
	"context"

    "github.com/HorseArcher567/octopus/pkg/rpc"

	"{{.Module}}/internal/logic"
	"{{.Module}}/proto/pb"
)

// Server gRPC 服务实现
type Server struct {
	pb.Unimplemented{{.ServiceNameCamel}}Server
	logic *logic.Logic
}

// Config 应用配置
type Config struct {
    Server rpc.ServerConfig // 服务配置（直接使用 rpc.ServerConfig）
    Mode   string           // 运行模式: dev/prod
}

// NewServer 创建 Server 实例
func NewServer(logic *logic.Logic) *Server {
	return &Server{
		logic: logic,
	}
}

// SayHello 实现 gRPC 方法
func (s *Server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	return s.logic.SayHello(ctx, req)
}
`
	return writeFromTemplate(filepath.Join(projectDir, "internal/server/server.go"), tmpl, data)
}

// generateProto 生成 proto 文件
func generateProto(projectDir string, data TemplateData) error {
	tmpl := `syntax = "proto3";

package pb;

// 使用模块相对路径，避免在本地生成多一层模块目录
option go_package = "proto/pb;pb";

// {{.ServiceNameCamel}} 服务定义
service {{.ServiceNameCamel}} {
  rpc SayHello(HelloRequest) returns (HelloResponse);
}

message HelloRequest {
  string name = 1;
}

message HelloResponse {
  string message = 1;
}
`
	return writeFromTemplate(filepath.Join(projectDir, "proto", data.ServiceName+".proto"), tmpl, data)
}

// generateConfigYaml 生成 config.yaml
func generateConfigYaml(projectDir string, data TemplateData) error {
	tmpl := `Server:
  AppName: {{.ServiceName}}
  Host: 0.0.0.0
  Port: 9000
  EtcdAddr:
    - 127.0.0.1:2379
  TTL: 10

Mode: dev
`
	return writeFromTemplate(filepath.Join(projectDir, "etc/config.yaml"), tmpl, data)
}

// generateMakefile 生成 Makefile
func generateMakefile(projectDir string, data TemplateData) error {
	tmpl := `.PHONY: proto build run clean

# 生成 Proto 代码
proto:
	@echo "Generating proto..."
	@protoc --go_out=. \
		--go-grpc_out=. \
		proto/{{.ServiceName}}.proto
	@echo "✅ Proto generated"

# 构建
build: proto
	@echo "Building..."
	@go build -o bin/{{.ServiceName}} cmd/main.go
	@echo "✅ Build complete: bin/{{.ServiceName}}"

# 运行
run: proto
	@echo "Starting {{.ServiceName}}..."
	@go run cmd/main.go

# 清理
clean:
	@rm -rf bin/ proto/pb/
	@echo "✅ Cleaned"

# 安装依赖
deps:
	@go mod tidy
	@echo "✅ Dependencies installed"

# 测试
test:
	@go test -v ./...
`
	return writeFromTemplate(filepath.Join(projectDir, "Makefile"), tmpl, data)
}

// generateGitignore 生成 .gitignore
func generateGitignore(projectDir string) error {
	content := `# Binaries
bin/
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary
*.test

# Output
*.out

# Go workspace file
go.work

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db
`
	return writeFile(filepath.Join(projectDir, ".gitignore"), content)
}

// initGoMod 初始化 go.mod
func initGoMod(projectDir, module string) error {
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

// writeFromTemplate 从模板写入文件
func writeFromTemplate(path, tmplStr string, data interface{}) error {
	tmpl, err := template.New("").Parse(tmplStr)
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, data)
}

// writeFile 直接写入文件
func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
