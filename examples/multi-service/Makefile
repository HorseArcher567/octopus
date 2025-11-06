.PHONY: proto build-server build-client run-server run-client test clean

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
	@go build -o bin/server server/main.go
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
	@go run server/main.go

# 运行客户端
run-client: proto
	@echo "Running client tests..."
	@go run client/main.go

# 测试（启动服务器并运行客户端测试）
test:
	@echo "Running integration test..."
	@go run server/main.go & \
	SERVER_PID=$$!; \
	sleep 2; \
	go run client/main.go; \
	TEST_RESULT=$$?; \
	kill $$SERVER_PID; \
	exit $$TEST_RESULT

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
	@grpcurl -plaintext -d '{"user_id": 1001}' localhost:9000 multi.UserService/GetUser

grpcurl-order:
	@echo "Testing OrderService.GetOrder..."
	@grpcurl -plaintext -d '{"order_id": 2001}' localhost:9000 multi.OrderService/GetOrder

grpcurl-product:
	@echo "Testing ProductService.GetProduct..."
	@grpcurl -plaintext -d '{"product_id": 1}' localhost:9000 multi.ProductService/GetProduct

