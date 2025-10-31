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

	"{{.Module}}/internal/config"
	"{{.Module}}/internal/logic"
	"{{.Module}}/internal/server"
	"{{.Module}}/proto/pb"

	"google.golang.org/grpc"
	"octopus/pkg/config"
	"octopus/pkg/rpc"
)

func main() {
	// 1. 加载配置
	var cfg config.Config
	config.MustLoadWithEnvAndUnmarshal("etc/config.yaml", &cfg)

	// 2. 创建 Logic
	logic := logic.NewLogic()

	// 3. 创建 Server
	srv := server.NewServer(logic)

	// 4. 创建 RPC Server（直接使用配置）
	cfg.Server.EnableReflection = cfg.Mode == "dev"
	cfg.Server.EnableHealth = true
	rpcServer := rpc.NewServer(&cfg.Server)

	// 5. 注册服务
	rpcServer.RegisterService(func(s *grpc.Server) {
		pb.Register{{.ServiceNameCamel}}Server(s, srv)
	})

	// 6. 启动服务
	if err := rpcServer.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
`
	return writeFromTemplate(filepath.Join(projectDir, "cmd/main.go"), tmpl, data)
}

// generateConfig 生成 config.go
func generateConfig(projectDir string, data TemplateData) error {
	tmpl := `package config

import "octopus/pkg/rpc"

// Config 服务配置
type Config struct {
	Server rpc.ServerConfig  // 服务配置（直接使用 rpc.ServerConfig）
	Mode   string            // 运行模式: dev/prod
}
`
	return writeFile(filepath.Join(projectDir, "internal/config/config.go"), tmpl)
}

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

	"{{.Module}}/internal/logic"
	"{{.Module}}/proto/pb"
)

// Server gRPC 服务实现
type Server struct {
	pb.Unimplemented{{.ServiceNameCamel}}Server
	logic *logic.Logic
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

option go_package = "./pb";

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
  Name: {{.ServiceName}}
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
	@mkdir -p proto/pb
	@protoc --go_out=proto/pb --go_opt=paths=source_relative \
		--go-grpc_out=proto/pb --go-grpc_opt=paths=source_relative \
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
	cmd := exec.Command("go", "mod", "init", module)
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
`
