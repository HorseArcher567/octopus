# Multi-Service Server Example

这个示例演示了如何在一个 gRPC 服务器上注册多个服务，并从配置文件读取配置。

## 功能特性

- ✅ 在单个服务器上注册多个 gRPC 服务（User、Order、Product）
- ✅ 从 YAML 配置文件读取配置
- ✅ 支持命令行参数指定配置文件路径
- ✅ 支持环境变量替换
- ✅ 集成日志系统，自动注入 request_id 和 method
- ✅ 服务注册到 etcd
- ✅ 优雅关闭

## 快速开始

### 1. 使用默认配置文件启动

```bash
go run main.go
```

这将使用当前目录下的 `config.yaml` 文件。

### 2. 指定配置文件路径

```bash
go run main.go -config /path/to/your/config.yaml
```

或使用长参数：

```bash
go run main.go --config /path/to/your/config.yaml
```

### 3. 查看帮助信息

```bash
go run main.go -h
```

## 配置文件说明

配置文件使用 YAML 格式，包含两个主要部分：

### 日志配置 (logger)

```yaml
logger:
  level: debug          # 日志级别: debug/info/warn/error
  format: text          # 日志格式: json/text
  add_source: true      # 是否添加源码位置
  output: stdout        # 输出目标: stdout/stderr/文件路径
  max_age: 7            # 日志保留天数 (仅文件输出有效)
```

### 服务器配置 (server)

```yaml
server:
  app_name: multi-service-demo  # 应用名称
  host: 0.0.0.0                 # 监听地址
  port: 9000                    # 监听端口
  advertise_addr: ""            # 注册地址（留空则自动获取）
  etcd_addr:                    # etcd 地址列表
    - localhost:2379
  ttl: 10                       # 租约时间 (秒)
  enable_reflection: true       # 是否启用 gRPC 反射
```

#### 监听地址 vs 注册地址（重要概念）

这是分布式服务配置的核心概念：

- **`host` (监听地址)**：服务实际监听的网络接口
  - `0.0.0.0` - 监听所有网络接口（生产环境推荐）
  - `127.0.0.1` - 仅监听本地回环接口（本地开发）
  - 具体 IP - 监听指定的网络接口

- **`advertise_addr` (注册地址)**：注册到 etcd 的地址，供其他服务访问
  - **当 `host` 为 `0.0.0.0`、`127.0.0.1` 或 `localhost` 时必须配置**
  - 必须是其他机器可以访问的 IP 地址
  - 支持环境变量 - `${ADVERTISE_ADDR}`

**为什么需要分开配置？**

当监听地址为 `0.0.0.0`、`127.0.0.1` 或 `localhost` 时，这些地址无法被其他机器访问：
- `0.0.0.0` - 仅表示"监听所有接口"，不是有效的连接地址
- `127.0.0.1`/`localhost` - 回环地址，只能本机访问

因此需要通过 `advertise_addr` 明确指定一个可达的 IP 地址。

**Fail-Fast 原则：**

如果配置了 `etcd_addr`（需要服务注册），但 `host` 是不可访问的地址且 `advertise_addr` 未配置，服务启动时会**立即失败**并给出清晰的错误提示和本机 IP 列表，强制用户正确配置。

## 环境变量支持

配置文件支持环境变量替换，格式为：

- `${ENV_VAR}` - 使用环境变量值
- `${ENV_VAR:default}` - 如果环境变量不存在，使用默认值

示例：

```yaml
server:
  host: ${SERVER_HOST:127.0.0.1}
  port: ${SERVER_PORT:9000}
  etcd_addr:
    - ${ETCD_ADDR:localhost:2379}
```

使用示例：

```bash
# 使用环境变量覆盖配置
export SERVER_PORT=8080
export ETCD_ADDR=192.168.1.100:2379
go run main.go

# 指定注册地址（多网卡场景）
export ADVERTISE_ADDR=192.168.1.100
go run main.go

# 完整的生产环境示例
export APP_NAME=my-service
export SERVER_HOST=0.0.0.0
export SERVER_PORT=9000
export ADVERTISE_ADDR=10.0.1.50
export ETCD_ADDR_1=10.0.1.10:2379
export ETCD_ADDR_2=10.0.1.11:2379
export ETCD_ADDR_3=10.0.1.12:2379
go run main.go -config config.prod.yaml
```

## 测试服务

### 使用 grpcurl 测试

```bash
# 列出所有服务
grpcurl -plaintext localhost:9000 list

# 获取用户信息
grpcurl -plaintext -d '{"user_id": 123}' localhost:9000 user.UserService/GetUser

# 创建用户
grpcurl -plaintext -d '{"username": "john", "email": "john@example.com"}' \
  localhost:9000 user.UserService/CreateUser

# 获取订单信息
grpcurl -plaintext -d '{"order_id": 456}' localhost:9000 order.OrderService/GetOrder

# 获取产品信息
grpcurl -plaintext -d '{"product_id": 789}' localhost:9000 product.ProductService/GetProduct

# 列出产品
grpcurl -plaintext -d '{"page": 1, "page_size": 10}' \
  localhost:9000 product.ProductService/ListProducts
```

## 配置文件示例

完整的配置示例请参考 `config.yaml` 文件。

你也可以创建多个配置文件用于不同环境：

- `config.yaml` - 默认配置
- `config.dev.yaml` - 开发环境配置
- `config.prod.yaml` - 生产环境配置

然后使用：

```bash
# 开发环境
go run main.go -config config.dev.yaml

# 生产环境
go run main.go -config config.prod.yaml
```

## 常见使用场景

### 场景 1: 本地开发（单机，不需要服务发现）

```yaml
server:
  host: 127.0.0.1        # 仅本地访问
  port: 9000
  advertise_addr: ""     # 不需要配置
  etcd_addr: []          # 不注册到 etcd
```

### 场景 2: 本地开发（多服务互调，需要服务发现）

```yaml
server:
  host: 0.0.0.0          # 监听所有接口
  port: 9000
  advertise_addr: 192.168.1.100  # 必须明确指定本机 IP
  etcd_addr:
    - localhost:2379
```

或者直接使用具体 IP 作为监听地址：

```yaml
server:
  host: 192.168.1.100    # 直接监听具体 IP
  port: 9000
  advertise_addr: ""     # 可以不配置，会使用 host
  etcd_addr:
    - localhost:2379
```

### 场景 3: 生产环境（通过环境变量）

**推荐方式：使用环境变量配置**

```yaml
server:
  host: ${SERVER_HOST:0.0.0.0}
  port: ${SERVER_PORT:9000}
  advertise_addr: ${ADVERTISE_ADDR}  # 运行时通过环境变量指定
  etcd_addr:
    - ${ETCD_ADDR_1:etcd1:2379}
    - ${ETCD_ADDR_2:etcd2:2379}
```

```bash
export ADVERTISE_ADDR=192.168.1.100
export ETCD_ADDR_1=10.0.1.10:2379
export ETCD_ADDR_2=10.0.1.11:2379
./server -config config.prod.yaml
```

### 场景 4: 生产环境（配置文件直接指定）

```yaml
server:
  host: 0.0.0.0                    # 监听所有接口
  port: 9000
  advertise_addr: 192.168.1.100    # 明确指定内网 IP
  etcd_addr:
    - etcd1:2379
    - etcd2:2379
```

### 场景 5: Docker 容器

```yaml
server:
  host: 0.0.0.0
  port: 9000
  advertise_addr: ${HOST_IP}       # 通过环境变量注入
  etcd_addr:
    - ${ETCD_ADDR:etcd:2379}
```

```bash
docker run -e HOST_IP=192.168.1.100 -e ETCD_ADDR=etcd:2379 myservice
```

### 场景 6: Kubernetes

```yaml
# ConfigMap
server:
  host: 0.0.0.0
  port: 9000
  advertise_addr: ${POD_IP}        # 使用 Pod IP
  etcd_addr:
    - etcd-service:2379
```

```yaml
# Deployment
spec:
  containers:
  - name: myservice
    env:
    - name: POD_IP
      valueFrom:
        fieldRef:
          fieldPath: status.podIP
```

## 注意事项

1. **配置文件路径**：如果不指定 `-config` 参数，程序会读取当前目录的 `config.yaml`

2. **Fail-Fast 原则**：当需要注册到 etcd 时，如果配置有误，服务会立即启动失败：
   - `host` 为 `0.0.0.0`/`127.0.0.1`/`localhost` 且 `advertise_addr` 未配置 → **启动失败**
   - 错误信息会显示本机检测到的 IP 列表和配置示例
   - 强制你在启动前就正确配置，避免生产环境问题

3. **不需要服务发现的场景**：如果 `etcd_addr` 为空，`advertise_addr` 可以不配置

4. **推荐配置方式**：
   - 开发环境：`host: 127.0.0.1`，不配置 `etcd_addr`
   - 生产环境：`host: 0.0.0.0`，通过环境变量 `${ADVERTISE_ADDR}` 指定注册地址

5. **多网卡场景**：
   - 方案 1：显式配置 `advertise_addr`（推荐）
   - 方案 2：直接使用具体 IP 作为 `host`（简单但灵活性低）

6. **容器环境**：必须通过环境变量注入 `advertise_addr`，如 Pod IP、Host IP 等

