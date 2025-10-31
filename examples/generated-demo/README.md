# Generated Demo Service

这是使用 `octopus-cli` 生成的示例服务。

## 快速开始

### 1. 生成项目

```bash
# 安装 octopus-cli
cd octopus
go install ./cmd/octopus-cli

# 生成服务
octopus-cli new user-service --module=github.com/yourname/user-service
```

### 2. 生成 Proto 代码

```bash
cd user-service
make proto
```

### 3. 运行服务

```bash
# 确保 etcd 正在运行
make run
```

### 4. 测试服务

```bash
# 使用 grpcurl 测试
grpcurl -plaintext -d '{"name":"World"}' \
  localhost:9000 \
  pb.UserService/SayHello
```

## 项目结构

```
user-service/
├── cmd/
│   └── main.go              # 服务入口
├── internal/
│   ├── config/
│   │   └── config.go        # 配置定义
│   ├── logic/
│   │   └── logic.go         # 业务逻辑（你编写这里）
│   └── server/
│       └── server.go        # gRPC 服务实现
├── proto/
│   ├── user-service.proto   # Proto 定义
│   └── pb/                  # 生成的代码
├── etc/
│   └── config.yaml          # 配置文件
├── go.mod
└── Makefile
```

## 开发流程

### 1. 定义接口

编辑 `proto/user-service.proto`:

```protobuf
service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
}
```

### 2. 生成代码

```bash
make proto
```

### 3. 实现业务逻辑

编辑 `internal/logic/logic.go`:

```go
func (l *Logic) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    // 实现你的业务逻辑
    return &pb.GetUserResponse{
        User: &pb.User{
            Id:   req.Id,
            Name: "John Doe",
        },
    }, nil
}
```

### 4. 更新 Server 层

编辑 `internal/server/server.go`:

```go
func (s *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    return s.logic.GetUser(ctx, req)
}
```

### 5. 运行测试

```bash
make run
```

## 特性

- ✅ 自动服务注册到 etcd
- ✅ 自动服务发现和负载均衡
- ✅ 健康检查
- ✅ gRPC 反射（开发模式）
- ✅ 优雅关闭
- ✅ 配置管理

## 配置说明

`etc/config.yaml`:

```yaml
Name: user-service    # 服务名称
Host: 0.0.0.0        # 监听地址
Port: 9000           # 监听端口
Mode: dev            # 运行模式

Etcd:
  Endpoints:
    - 127.0.0.1:2379  # etcd 地址
  TTL: 10             # 租约时间（秒）
```

## 客户端调用

```go
package main

import (
    "context"
    "log"

    "octopus/pkg/rpc"
    "your-module/proto/pb"
)

func main() {
    // 创建客户端（自动服务发现）
    conn, err := rpc.NewClient(&rpc.ClientConfig{
        ServiceName: "user-service",
        EtcdAddr:    []string{"127.0.0.1:2379"},
        Timeout:     5000,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    client := pb.NewUserServiceClient(conn)

    // 调用服务
    resp, err := client.SayHello(context.Background(), &pb.HelloRequest{
        Name: "World",
    })
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Response: %s", resp.Message)
}
```

## 生产部署

### Docker

创建 `Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /build
COPY . .
RUN go mod download
RUN go build -o app cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /build/app .
COPY --from=builder /build/etc ./etc
EXPOSE 9000
CMD ["./app"]
```

构建和运行:

```bash
docker build -t user-service:latest .
docker run -p 9000:9000 user-service:latest
```

### Kubernetes

参考 `octopus/docs/DEPLOYMENT.md`

