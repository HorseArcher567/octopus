.PHONY: help build test run-server run-client clean install

help: ## 显示帮助信息
	@echo "可用命令:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

install: ## 安装依赖
	go mod tidy
	go mod download

install-cli: ## 安装代码生成工具
	@echo "Installing octopus-cli..."
	@go install ./cmd/octopus-cli
	@echo "✅ octopus-cli installed successfully"
	@echo ""
	@echo "Usage:"
	@echo "  octopus-cli new <service-name>    Create a new service"
	@echo "  octopus-cli version                Show version"

build: ## 编译项目
	@echo "Building examples..."
	@cd examples/simple/server && go build -o ../../../bin/server main.go
	@cd examples/simple/client && go build -o ../../../bin/client main.go
	@echo "Build complete! Binaries in bin/"

test: ## 运行测试
	go test -v -race -coverprofile=coverage.out ./pkg/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-short: ## 运行快速测试
	go test -short ./pkg/...

run-server: ## 运行服务端示例
	@echo "Starting server..."
	@cd examples/simple/server && go run main.go

run-client: ## 运行客户端示例
	@echo "Starting client..."
	@cd examples/simple/client && go run main.go

lint: ## 运行代码检查
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

fmt: ## 格式化代码
	go fmt ./...
	goimports -w .

clean: ## 清理构建产物
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -testcache

check-etcd: ## 检查etcd是否运行
	@etcdctl endpoint health || (echo "❌ etcd not running. Please start etcd first." && exit 1)
	@echo "✅ etcd is running"

demo: check-etcd ## 运行完整演示
	@echo "🚀 Starting demo..."
	@echo "Starting server in background..."
	@cd examples/simple/server && go run main.go &
	@sleep 2
	@echo "\n📡 Running client..."
	@cd examples/simple/client && timeout 10 go run main.go || true
	@echo "\n✅ Demo complete"

.DEFAULT_GOAL := help

