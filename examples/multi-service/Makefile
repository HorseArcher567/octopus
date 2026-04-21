.PHONY: proto build-server build-client run-server run-client test test-e2e clean

# 生成 Proto 代码
proto:
	@echo "Generating proto..."
	@protoc --go_out=. \
		--go-grpc_out=. \
		proto/services.proto
	@echo "✅ Proto generated"

# 构建服务器
build-server: proto
	@echo "Building server..."
	@go build -o bin/server ./server
	@echo "✅ Server built: bin/server"

# 构建客户端
build-client: proto
	@echo "Building client..."
	@go build -o bin/client client/main.go
	@echo "✅ Client built: bin/client"

# 构建所有
build: build-server build-client

# 运行服务器
run-server: proto
	@echo "Starting multi-service server..."
	@go run ./server -config server/config.yaml

# 运行客户端
run-client: proto
	@echo "Running client tests..."
	@go run ./client -config client/config.yaml

# 测试
test:
	@echo "Running unit and integration tests..."
	@GOCACHE=/tmp/go-build go test ./...

test-e2e:
	@echo "Running e2e test..."
	@echo "Requires OCTOPUS_TEST_MYSQL_DSN to point to an existing MySQL database."
	@GOCACHE=/tmp/go-build go test ./server/tests/e2e -v

# 清理
clean:
	@rm -rf bin/ proto/pb/
	@echo "✅ Cleaned"

# 使用 grpcurl 测试服务
grpcurl-list:
	@echo "Listing services..."
	@grpcurl -plaintext localhost:9000 list

grpcurl-user:
	@echo "Testing UserService.GetUser..."
	@grpcurl -plaintext -d '{"user_id": 1001}' localhost:9001 multi.User/GetUser

grpcurl-order:
	@echo "Testing OrderService.GetOrder..."
	@grpcurl -plaintext -d '{"order_id": 2001}' localhost:9001 multi.Order/GetOrder

grpcurl-product:
	@echo "Testing ProductService.GetProduct..."
	@grpcurl -plaintext -d '{"product_id": 1}' localhost:9001 multi.Product/GetProduct

