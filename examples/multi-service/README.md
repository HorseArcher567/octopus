# Multi-Service Example

这个示例展示了如何在一个 gRPC 服务器中注册和使用多个服务。

## 架构说明

本示例在一个 gRPC 服务器进程中注册了三个服务：
- **UserService**: 用户管理服务
- **OrderService**: 订单管理服务
- **ProductService**: 产品管理服务

所有服务共享同一个端口（9000），通过 gRPC 的 service 路由机制进行区分。

## 目录结构

```
multi-service/
├── proto/
│   ├── services.proto      # 定义三个服务的 proto 文件
│   └── pb/                 # 生成的代码（自动生成）
├── server/
│   └── main.go             # 服务器实现，注册多个服务
├── client/
│   └── main.go             # 客户端测试代码
├── Makefile                # 构建和测试脚本
└── README.md
```

## 快速开始

### 前置条件

确保 etcd 已经运行：

```bash
# macOS
brew install etcd
etcd

# 或使用 Docker
docker run -d --name etcd \
  -p 2379:2379 \
  quay.io/coreos/etcd:v3.5.0 \
  etcd --listen-client-urls http://0.0.0.0:2379 \
       --advertise-client-urls http://0.0.0.0:2379
```

### 1. 生成 Proto 代码

```bash
make proto
```

### 2. 运行服务器

```bash
make run-server
```

服务器会：
- 监听在 `127.0.0.1:9000`
- 注册三个服务
- 将服务信息注册到 etcd

### 3. 运行客户端测试

在另一个终端运行：

```bash
make run-client
```

客户端会：
- 通过 etcd 自动发现服务
- 依次调用三个服务的方法
- 验证多服务功能

## 使用 grpcurl 测试

如果安装了 `grpcurl`，可以使用以下命令测试：

### 列出所有服务

```bash
make grpcurl-list
```

### 测试 UserService

```bash
make grpcurl-user
```

### 测试 OrderService

```bash
make grpcurl-order
```

### 测试 ProductService

```bash
make grpcurl-product
```

或手动调用：

```bash
# 获取用户信息
grpcurl -plaintext -d '{"user_id": 1001}' localhost:9000 multi.UserService/GetUser

# 创建用户
grpcurl -plaintext -d '{"username": "testuser", "email": "test@example.com"}' \
  localhost:9000 multi.UserService/CreateUser

# 获取订单
grpcurl -plaintext -d '{"order_id": 2001}' localhost:9000 multi.OrderService/GetOrder

# 创建订单
grpcurl -plaintext -d '{"user_id": 1001, "product_name": "Test Product", "amount": 99.99}' \
  localhost:9000 multi.OrderService/CreateOrder

# 获取产品
grpcurl -plaintext -d '{"product_id": 1}' localhost:9000 multi.ProductService/GetProduct

# 列出产品
grpcurl -plaintext -d '{"page": 1, "page_size": 10}' \
  localhost:9000 multi.ProductService/ListProducts
```

## 健康检查

服务器启用了健康检查，可以查询各个服务的健康状态：

```bash
# 检查整体健康状态
grpcurl -plaintext localhost:9000 grpc.health.v1.Health/Check

# 检查特定服务的健康状态
grpcurl -plaintext -d '{"service": "UserService"}' \
  localhost:9000 grpc.health.v1.Health/Check

grpcurl -plaintext -d '{"service": "OrderService"}' \
  localhost:9000 grpc.health.v1.Health/Check

grpcurl -plaintext -d '{"service": "ProductService"}' \
  localhost:9000 grpc.health.v1.Health/Check
```

## 核心代码说明

### 服务器端注册多个服务

```go
// 创建 RPC 服务器
server := rpc.MustNewServer(config)

// 注册多个服务
server.RegisterService(func(s *grpc.Server) {
    pb.RegisterUserServiceServer(s, userService)
})

server.RegisterService(func(s *grpc.Server) {
    pb.RegisterOrderServiceServer(s, orderService)
})

server.RegisterService(func(s *grpc.Server) {
    pb.RegisterProductServiceServer(s, productService)
})

// 启动服务器
if err := server.Start(); err != nil {
    // handle error
}
```

### 客户端连接使用

```go
// 通过 etcd 服务发现自动连接
conn, err := rpc.NewClient(ctx, &rpc.ClientConfig{
    Target:   "multi-service-demo",
    EtcdAddr: []string{"127.0.0.1:2379"},
})

// 创建不同服务的客户端（共享同一个连接）
userClient := pb.NewUserServiceClient(conn)
orderClient := pb.NewOrderServiceClient(conn)
productClient := pb.NewProductServiceClient(conn)

// 调用不同服务的方法
userClient.GetUser(ctx, &pb.GetUserRequest{...})
orderClient.GetOrder(ctx, &pb.GetOrderRequest{...})
productClient.GetProduct(ctx, &pb.GetProductRequest{...})
```

## 优势

1. **简化部署**: 多个相关服务可以部署在一个进程中
2. **共享资源**: 共享连接池、配置、中间件等
3. **统一管理**: 统一的健康检查、监控、日志
4. **灵活组合**: 可以根据需要灵活组合不同的服务
5. **服务发现**: 通过 etcd 实现自动服务发现和负载均衡
6. **高可用**: 支持多实例部署，客户端自动故障转移

## 注意事项

- 所有服务共享同一个进程，一个服务崩溃可能影响其他服务
- 服务之间不能有循环依赖
- 适合相关性强、耦合度高的服务聚合场景
- 如果服务需要独立扩展，建议拆分为独立进程

