package generator

import (
	"os"
	"path/filepath"
	"text/template"
)

// MonorepoData monorepo 项目数据
type MonorepoData struct {
	ProjectName string // my-project
	Module      string // github.com/xxx/my-project
}

// AppData 应用数据
type AppData struct {
	AppName      string // user
	AppNameCamel string // User
	Module       string // github.com/xxx/my-project/apps/user (app 自己的 module)
	RootModule   string // github.com/xxx/my-project (根 module，用于 proto)
	Port         int    // 9001
	ServiceName  string // user-service
}

// generateMonorepoReadme 生成 README.md
func generateMonorepoReadme(projectDir string, data MonorepoData) error {
	tmpl := `# {{.ProjectName}}

A monorepo project built with Octopus RPC Framework.

## Project Structure

` + "```" + `
{{.ProjectName}}/
├── apps/              # Applications
│   ├── user/          # User service
│   ├── order/         # Order service
│   └── ...
├── proto/             # Protocol buffer definitions
│   ├── user.proto     # proto files at root
│   ├── order.proto
│   ├── product.proto
│   ├── user/          # generated pb go files live here
│   ├── order/
│   └── product/
├── pkg/              # Shared packages
│   ├── middleware/
│   ├── utils/
│   └── errors/
├── scripts/          # Build and deploy scripts
├── go.work           # Go workspace (manages all modules)
└── Makefile          # Build tasks
` + "```" + `

## Getting Started

### Prerequisites

- Go 1.21+ (with workspace support)
- Protocol Buffers compiler (protoc)
- etcd (for service discovery)

### Project Structure

This is a **multi-module monorepo**:
- **proto/ has its own go.mod**: Proto definitions and generated code
- **pkg/ has its own go.mod**: Shared packages that can be used by all apps
- **Each app has its own go.mod**: Located in apps/<app-name>/go.mod
- **go.work**: Manages all modules in the workspace, automatically updated when adding apps
- **No root go.mod**: Root directory is just for organization

### Add a new application

` + "```bash" + `
octopus-cli add <app-name> --port <port>
` + "```" + `

This will:
- Create app structure in apps/<app-name>/
- Initialize independent go.mod for the app
- Add the app to go.work automatically
- Generate proto file in proto/<app-name>.proto

### Build

` + "```bash" + `
# Build all applications
make build-all

# Build specific application
make build-user
make build-order
` + "```" + `

### Run

` + "```bash" + `
# Run specific application
make run-user
make run-order
` + "```" + `

## Development

### Generate proto files

` + "```bash" + `
make proto
` + "```" + `

### Run tests

` + "```bash" + `
make test
` + "```" + `

## License

MIT
`
	return writeFromTemplate(filepath.Join(projectDir, "README.md"), tmpl, data)
}

// generateMonorepoMakefile 生成根 Makefile
func generateMonorepoMakefile(projectDir string, data MonorepoData) error {
	tmpl := `.PHONY: proto build-all clean deps sync-deps test help

# Help target
help:
	@echo "Available targets:"
	@echo "  make proto       - Generate all proto files"
	@echo "  make build-all   - Build all applications"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make deps        - Install dependencies"
	@echo "  make test        - Run tests"
	@echo ""
	@echo "Application-specific targets will be added when you add apps:"
	@echo "  make build-<app> - Build specific application"
	@echo "  make run-<app>   - Run specific application"

# Generate all proto files
proto:
	@echo "Generating proto files..."
	@if [ -d "proto" ]; then \
		PROTO_MODULE=$$(grep "^module " proto/go.mod | cut -d' ' -f2); \
		if [ -z "$$PROTO_MODULE" ]; then \
			echo "⚠️  Failed to read proto module from proto/go.mod"; \
			exit 1; \
		fi; \
		find proto -name "*.proto" -type f | while read proto_file; do \
			echo "  Processing $$proto_file..."; \
			protoc --go_out=proto --go_opt=module=$$PROTO_MODULE --go-grpc_out=proto --go-grpc_opt=module=$$PROTO_MODULE $$proto_file; \
		done; \
		echo "✅ Proto files generated"; \
	else \
		echo "⚠️  No proto directory found"; \
	fi

# Build all applications
build-all: proto sync-deps
	@echo "Building all applications..."
	@if [ -d "apps" ]; then \
		for app in apps/*/; do \
			app_name=$$(basename $$app); \
			if [ -f "$$app/cmd/main.go" ]; then \
				echo "  Building $$app_name..."; \
				go build -o bin/$${app_name}-service $$app/cmd/main.go; \
			fi; \
		done; \
		echo "✅ All applications built"; \
	else \
		echo "⚠️  No apps directory found"; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@find proto -name "*.pb.go" -type f -delete 2>/dev/null || true
	@echo "✅ Cleaned"

# Install dependencies (tidy all modules and sync workspace)
deps: sync-deps

# Sync dependencies across all modules in workspace
sync-deps:
	@echo "Syncing dependencies..."
	@if [ -f "proto/go.mod" ]; then \
		echo "  Tidy proto module..."; \
		cd proto && go mod tidy; \
	fi
	@if [ -f "pkg/go.mod" ]; then \
		echo "  Tidy pkg module..."; \
		cd pkg && go mod tidy; \
	fi
	@if [ -d "apps" ]; then \
		for app in apps/*/; do \
			if [ -f "$$app/go.mod" ]; then \
				app_name=$$(basename $$app); \
				echo "  Tidy $$app_name module..."; \
				cd $$app && go mod tidy; \
			fi; \
		done; \
	fi
	@echo "  Syncing workspace..."
	@go work sync
	@echo "✅ Dependencies synced"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...
	@echo "✅ Tests completed"

# Individual app targets will be added here automatically
`
	return writeFromTemplate(filepath.Join(projectDir, "Makefile"), tmpl, data)
}

// generateMonorepoGitignore 生成 .gitignore
func generateMonorepoGitignore(projectDir string) error {
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
go.work.sum

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Generated proto files
*.pb.go
proto/pb/

# Vendor
vendor/

# Temporary files
*.tmp
*.log
`
	return writeFile(filepath.Join(projectDir, ".gitignore"), content)
}

// generateGoWork 生成 go.work
func generateGoWork(projectDir string, data MonorepoData) error {
	tmpl := `go 1.21

use (
	./proto
	./pkg
)

// Applications will be added automatically when you run: octopus-cli add <app-name>
// Each app has its own go.mod in apps/<app-name>/
`
	return writeFile(filepath.Join(projectDir, "go.work"), tmpl)
}

// generateAppMain 生成应用的 main.go
func generateAppMain(appDir string, data AppData) error {
	tmpl := `package main

import (
	"log"

	"{{.Module}}/internal/logic"
	"{{.Module}}/internal/server"
	"{{.RootModule}}/{{.AppName}}"

	"google.golang.org/grpc"
	"github.com/HorseArcher567/octopus/pkg/config"
	"github.com/HorseArcher567/octopus/pkg/rpc"
)

func main() {
	// 1. Load configuration
	var cfg server.Config
	config.MustLoadWithEnvAndUnmarshal("etc/config.yaml", &cfg)

	// 2. Create Logic
	logic := logic.NewLogic()

	// 3. Create Server
	srv := server.NewServer(logic)

	// 4. Create RPC Server
	cfg.Server.EnableReflection = cfg.Mode == "dev"
	rpcServer := rpc.NewServer(&cfg.Server)

	// 5. Register service
	rpcServer.RegisterService(func(s *grpc.Server) {
		{{.AppName}}.Register{{.AppNameCamel}}ServiceServer(s, srv)
	})

	// 6. Start server
	log.Printf("Starting {{.ServiceName}} on :%d", cfg.Server.Port)
	if err := rpcServer.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
`
	return writeFromTemplate(filepath.Join(appDir, "cmd/main.go"), tmpl, data)
}

// generateAppLogic 生成应用的 logic.go
func generateAppLogic(appDir string, data AppData) error {
	tmpl := `package logic

import (
	"context"

	"{{.RootModule}}/{{.AppName}}"
)

// Logic {{.AppNameCamel}} business logic layer
type Logic struct {
	// TODO: Add dependencies (database, cache, etc.)
}

// NewLogic creates a new Logic instance
func NewLogic() *Logic {
	return &Logic{}
}

// SayHello example method (implement your business logic)
func (l *Logic) SayHello(ctx context.Context, req *{{.AppName}}.HelloRequest) (*{{.AppName}}.HelloResponse, error) {
	// TODO: Implement your business logic
	
	return &{{.AppName}}.HelloResponse{
		Message: "Hello from {{.ServiceName}}: " + req.Name,
	}, nil
}
`
	return writeFromTemplate(filepath.Join(appDir, "internal/logic", data.AppName+".go"), tmpl, data)
}

// generateAppServer 生成应用的 server.go
func generateAppServer(appDir string, data AppData) error {
	tmpl := `package server

import (
	"context"

	"github.com/HorseArcher567/octopus/pkg/rpc"

	"{{.Module}}/internal/logic"
	"{{.RootModule}}/{{.AppName}}"
)

// Server gRPC service implementation
type Server struct {
	{{.AppName}}.Unimplemented{{.AppNameCamel}}ServiceServer
	logic *logic.Logic
}

// Config application configuration
type Config struct {
	Server rpc.ServerConfig // Server configuration
	Mode   string           // Running mode: dev/prod
}

// NewServer creates a new Server instance
func NewServer(logic *logic.Logic) *Server {
	return &Server{
		logic: logic,
	}
}

// SayHello implements gRPC method
func (s *Server) SayHello(ctx context.Context, req *{{.AppName}}.HelloRequest) (*{{.AppName}}.HelloResponse, error) {
	return s.logic.SayHello(ctx, req)
}
`
	return writeFromTemplate(filepath.Join(appDir, "internal/server", data.AppName+".go"), tmpl, data)
}

// generateAppConfig 生成应用的 config.yaml
func generateAppConfig(appDir string, data AppData) error {
	tmpl := `Server:
  AppName: {{.ServiceName}}
  Host: 0.0.0.0
  Port: {{.Port}}
  EtcdAddr:
    - 127.0.0.1:2379
  TTL: 10

Mode: dev
`
	return writeFromTemplate(filepath.Join(appDir, "etc/config.yaml"), tmpl, data)
}

// generateAppProto 生成应用的 proto 文件
func generateAppProto(protoDir string, data AppData) error {
	tmpl := `syntax = "proto3";

package {{.AppName}};

// Generated Go files reside in {{.AppName}}/ with package name {{.AppName}}
// Import path: {{.RootModule}}/{{.AppName}}
option go_package = "{{.RootModule}}/{{.AppName}};{{.AppName}}";

// {{.AppNameCamel}}Service service definition
service {{.AppNameCamel}}Service {
	rpc SayHello(HelloRequest) returns (HelloResponse);
}

message HelloRequest {
	string name = 1;
}

message HelloResponse {
	string message = 1;
}
`
	return writeFromTemplate(filepath.Join(protoDir, data.AppName+".proto"), tmpl, data)
}

// Helper function: write file from template
func writeFromTemplateMonorepo(path, tmplStr string, data interface{}) error {
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

// Helper function: write file directly
func writeFileMonorepo(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
