# Config - 灵活的配置管理包

一个功能强大、易于使用的 Go 语言配置管理库，支持多种配置格式（JSON、YAML、TOML），提供灵活的配置加载和访问方式。

## 特性

- 🎯 **多格式支持**: JSON、YAML、TOML 三种主流配置格式，自动识别
- 📁 **多种加载方式**: 从文件、字节流、Map 对象加载
- 🔑 **路径访问**: 支持点号分隔的嵌套路径访问（如 `database.host`）
- 🌍 **环境变量**: 支持环境变量替换 `${ENV_VAR}` 和默认值 `${ENV_VAR:default}`
- 🔄 **结构体转换**: 集成 mapstruct 包，轻松转换为结构体
- ✨ **类型安全**: 提供带默认值的类型安全方法，无需类型断言
- 📋 **切片支持**: 便捷获取字符串和整数数组
- 🛡️ **线程安全**: 内置读写锁，支持并发访问
- 📦 **零配置**: 开箱即用，无需复杂配置

## 包结构

```text
pkg/config/
├── config.go       # 核心 - Config 结构体及其方法 + 环境变量处理
├── sugar.go        # 语法糖 - 包级加载函数和 Must* 便捷方法
├── format.go       # 格式处理 - JSON/YAML/TOML 解析和序列化
└── config_test.go  # 测试
```

**职责划分**：

- **config.go** (531行): Config 结构体定义和所有实例方法（Get/Set/Unmarshal/WriteToFile 等）+ 环境变量替换
- **sugar.go** (138行): 包级便捷函数（Load*/Must* 系列）
- **format.go** (135行): 配置格式解析和序列化

## 快速开始

### 安装

```bash
go get octopus/pkg/config
```

### 基本使用

#### 1. 最简洁的用法

适用于应用启动时加载配置，加载失败时直接 panic。这是最简洁的使用方式：

```go
package main

import (
    "flag"
    "fmt"
    "octopus/pkg/config"
)

type Config struct {
    App struct {
        Name string
        Port int
    }
    Database struct {
        Host string
        Port int
    }
}

func main() {
    // 解析命令行参数
    configFile := flag.String("f", "config.yaml", "配置文件路径")
    flag.Parse()

    // 一行代码加载并解析配置（类似 go-zero 的 conf.MustLoad）
    var c Config
    config.MustUnmarshal(*configFile, &c)

    // 直接使用配置
    fmt.Printf("Starting %s at port %d\n", c.App.Name, c.App.Port)
    fmt.Printf("Database: %s:%d\n", c.Database.Host, c.Database.Port)
}
```

**支持环境变量的版本**:

```go
func main() {
    var c Config
    // 支持 ${ENV_VAR} 和 ${ENV_VAR:default} 格式的环境变量替换
    config.MustUnmarshal("config.yaml", &c)
    
    fmt.Printf("Database: %s:%d\n", c.Database.Host, c.Database.Port)
}
```

**其他 Must* 便捷方法**:

```go
// 只加载配置，不解析
cfg := config.MustLoad("config.yaml")
port := cfg.GetInt("app.port")


// 加载并支持环境变量
cfg := config.MustLoad("config.yaml")
```

#### 2. 需要错误处理的场景

如果需要自定义错误处理而不是直接 panic，使用普通的加载方法：

```go
// 加载并解析
cfg, err := config.Load("config.json")
if err != nil {
    log.Printf("failed to load config: %v", err)
    // 自定义错误处理逻辑
    return
}

var app AppConfig
if err := cfg.Unmarshal(&app); err != nil {
    log.Printf("failed to unmarshal config: %v", err)
    return
}
```

#### 3. 从文件加载配置

```go
package main

import (
    "fmt"
    "octopus/pkg/config"
)

func main() {
    // 加载配置文件（自动识别 JSON/YAML/TOML 格式）
    cfg, err := config.Load("config.json")
    if err != nil {
        panic(err)
    }

    // 读取配置值
    name := cfg.GetString("app.name")
    port := cfg.GetInt("app.port")
    enabled := cfg.GetBool("app.enabled")

    fmt.Printf("App: %s, Port: %d, Enabled: %v\n", name, port, enabled)
}
```

**config.json**:

```json
{
  "app": {
    "name": "MyApp",
    "port": 8080,
    "enabled": true
  }
}
```

#### 2. 从字节流加载

```go
// JSON 格式
jsonData := []byte(`{"name": "test", "port": 8080}`)
cfg, err := config.LoadFromBytes(jsonData, config.FormatJSON)

// YAML 格式
yamlData := []byte(`
name: test
port: 8080
`)
cfg, err := config.LoadFromBytes(yamlData, config.FormatYAML)

// TOML 格式
tomlData := []byte(`
name = "test"
port = 8080
`)
cfg, err := config.LoadFromBytes(tomlData, config.FormatTOML)
```

#### 3. 转换为结构体

```go
type AppConfig struct {
    Name string
    Port int
    Database struct {
        Host string
        Port int
    }
}

// 加载配置
cfg, err := config.Load("config.json")
if err != nil {
    panic(err)
}

// 转换为结构体
var app AppConfig
if err := cfg.Unmarshal(&app); err != nil {
    panic(err)
}

fmt.Printf("App: %+v\n", app)
```

#### 4. 使用类型安全的默认值

```go
// 类型安全的默认值方法，无需类型断言
host := cfg.GetStringWithDefault("server.host", "localhost")
port := cfg.GetIntWithDefault("server.port", 8080)
debug := cfg.GetBoolWithDefault("app.debug", false)
timeout := cfg.GetFloatWithDefault("server.timeout", 30.0)

fmt.Printf("Server: %s:%d, Debug: %v, Timeout: %.1f\n", host, port, debug, timeout)
```

#### 5. 获取数组/切片配置

```go
// 获取字符串数组
hosts := cfg.GetStringSlice("database.hosts")
for _, host := range hosts {
    fmt.Println("Host:", host)
}

// 获取整数数组
ports := cfg.GetIntSlice("server.ports")
for _, port := range ports {
    fmt.Println("Port:", port)
}

// 获取任意类型数组
items := cfg.GetSlice("items")
```

## 高级功能

### 1. 导出配置到文件

`WriteToFile` 方法可以将配置导出为文件，自动根据文件扩展名选择格式。适用于以下场景：

```go
// 场景1: 环境变量替换后检查实际值
cfg, _ := config.Load("config.yaml")
cfg.WriteToFile("resolved-config.yaml") // 查看环境变量替换后的值

// 场景3: 动态修改后保存
cfg.Set("debug", false)
cfg.Set("cache.enabled", true)
cfg.WriteToFile("updated-config.toml") // 保存修改后的配置

// 场景4: 格式转换
cfg, _ := config.Load("config.yaml")
cfg.WriteToFile("config.json") // YAML 转 JSON
cfg.WriteToFile("config.toml") // YAML 转 TOML
```

### 3. 嵌套路径访问

```go
// 设置嵌套配置
cfg.Set("database.host", "localhost")
cfg.Set("database.port", 3306)
cfg.Set("database.credentials.username", "admin")
cfg.Set("database.credentials.password", "secret")

// 读取嵌套配置
host := cfg.GetString("database.host")
username := cfg.GetString("database.credentials.username")
```

### 4. 环境变量替换

配置文件中可以使用环境变量：

**config.json**:

```json
{
  "database": {
    "host": "${DB_HOST}",
    "port": "${DB_PORT:3306}",
    "username": "${DB_USER:admin}"
  }
}
```

加载时自动替换：

```go
// 设置环境变量
os.Setenv("DB_HOST", "localhost")
os.Setenv("DB_PORT", "5432")

// 加载配置并替换环境变量
cfg, err := config.Load("config.json")

// DB_HOST = "localhost" (从环境变量获取)
// DB_PORT = "5432" (从环境变量获取)
// DB_USER = "admin" (使用默认值)
```

### 5. 获取配置段落

```go
// 获取某个配置段落
dbConfig := cfg.GetSection("database")
// dbConfig 是一个 map[string]any

// 或者解码到结构体
type DatabaseConfig struct {
    Host string
    Port int
    Username string
}

var db DatabaseConfig
err := cfg.UnmarshalKey("database", &db)
```

### 6. 默认值和检查

```go
// 方式1: 使用类型安全的默认值方法（推荐）
host := cfg.GetStringWithDefault("server.host", "localhost")
port := cfg.GetIntWithDefault("server.port", 8080)
debug := cfg.GetBoolWithDefault("app.debug", false)
timeout := cfg.GetFloatWithDefault("server.timeout", 30.0)

// 方式2: 通用的默认值方法（需要类型断言）
port := cfg.GetWithDefault("server.port", 8080).(int)

// 检查配置是否存在
if cfg.Has("feature.enabled") {
    enabled := cfg.GetBool("feature.enabled")
    // ...
}
```

### 7. 动态修改配置

```go
// 动态设置配置值
cfg.Set("server.port", 9090)
cfg.Set("database.pool.size", 100)
cfg.Set("cache.enabled", true)
cfg.Set("cache.ttl", 300)
```

## API 文档

### Config 结构

#### 加载方法

- `Load(filepath string) error` - 从文件加载配置，完全替换现有配置
- `LoadBytes(data []byte, format Format) error` - 从字节流加载配置，完全替换现有配置
- `Clear()` - 清空所有配置

#### 读取方法

**基本类型获取：**

- `Get(key string) (any, bool)` - 获取任意类型的值
- `GetString(key string) string` - 获取字符串
- `GetInt(key string) int` - 获取整数
- `GetBool(key string) bool` - 获取布尔值
- `GetFloat(key string) float64` - 获取浮点数

**带默认值的类型安全获取（推荐）：**

- `GetStringWithDefault(key string, defaultValue string) string` - 获取字符串或默认值
- `GetIntWithDefault(key string, defaultValue int) int` - 获取整数或默认值
- `GetBoolWithDefault(key string, defaultValue bool) bool` - 获取布尔值或默认值
- `GetFloatWithDefault(key string, defaultValue float64) float64` - 获取浮点数或默认值
- `GetWithDefault(key string, defaultValue any) any` - 获取值或默认值（需要类型断言）

**切片/数组获取：**

- `GetStringSlice(key string) []string` - 获取字符串切片
- `GetIntSlice(key string) []int` - 获取整数切片
- `GetSlice(key string) []any` - 获取任意类型切片

**其他：**

- `GetSection(key string) map[string]any` - 获取配置段落
- `GetAll() map[string]any` - 获取所有配置

#### 设置方法

- `Set(key string, value any)` - 设置配置值（支持路径）

#### 转换方法

- `Unmarshal(target interface{}) error` - 将配置转换为结构体
- `UnmarshalKey(key string, target interface{}) error` - 将指定key转换为结构体

#### 辅助方法

- `Has(key string) bool` - 检查配置是否存在
- `Clear()` - 清空所有配置
- `WriteToFile(filepath string) error` - 导出配置到文件（自动识别格式）

### 包级加载函数

**普通加载函数（返回错误）：**

- `Load(path string) (*Config, error)` - 加载单个配置文件（自动识别格式，默认支持环境变量替换）
- `LoadWithoutEnv(path string) (*Config, error)` - 加载配置文件但不替换环境变量
- `LoadFromBytes(data []byte, format Format) (*Config, error)` - 从字节流加载

**Must* 便捷方法（失败时 panic，适合启动阶段）：**

- `MustLoad(path string) *Config` - 加载配置文件（默认支持环境变量替换），失败时 panic
- `MustLoadWithoutEnv(path string) *Config` - 加载配置文件但不替换环境变量，失败时 panic
- `MustUnmarshal(path string, target interface{})` - 加载并直接解析到结构体（默认支持环境变量替换），失败时 panic（最便捷）
- `MustUnmarshalWithoutEnv(path string, target interface{})` - 加载并解析到结构体（不支持环境变量替换），失败时 panic

## 完整示例

### 示例 1: Web 应用配置

**config.yaml**:

```yaml
app:
  name: MyWebApp
  version: 1.0.0
  debug: true

server:
  host: ${SERVER_HOST:0.0.0.0}
  port: ${SERVER_PORT:8080}
  timeout: 30

database:
  driver: mysql
  host: ${DB_HOST:localhost}
  port: ${DB_PORT:3306}
  name: ${DB_NAME:myapp}
  username: ${DB_USER:root}
  password: ${DB_PASS}
  pool:
    max_open: 100
    max_idle: 10

redis:
  host: ${REDIS_HOST:localhost}
  port: ${REDIS_PORT:6379}
  db: 0

logging:
  level: info
  format: json
  output: stdout
```

**main.go**:

```go
package main

import (
    "fmt"
    "octopus/pkg/config"
)

type Config struct {
    App struct {
        Name    string
        Version string
        Debug   bool
    }
    Server struct {
        Host    string
        Port    int
        Timeout int
    }
    Database struct {
        Driver   string
        Host     string
        Port     int
        Name     string
        Username string
        Password string
        Pool     struct {
            MaxOpen int `mapstruct:"max_open"`
            MaxIdle int `mapstruct:"max_idle"`
        }
    }
    Redis struct {
        Host string
        Port int
        DB   int
    }
    Logging struct {
        Level  string
        Format string
        Output string
    }
}

func main() {
    // 方式1（推荐）: 最简洁 - 一行代码加载并解析（支持环境变量）
    var appConfig Config
    config.MustUnmarshal("config.yaml", &appConfig)
    
    fmt.Printf("Starting %s v%s\n", appConfig.App.Name, appConfig.App.Version)
    fmt.Printf("Server: %s:%d\n", appConfig.Server.Host, appConfig.Server.Port)
    fmt.Printf("Database: %s@%s:%d/%s\n", 
        appConfig.Database.Username, 
        appConfig.Database.Host, 
        appConfig.Database.Port, 
        appConfig.Database.Name)

    // 方式2: 需要错误处理时
    cfg, err := config.Load("config.yaml")
    if err != nil {
        panic(err)
    }
    
    // 直接读取（使用类型安全的默认值）
    appName := cfg.GetStringWithDefault("app.name", "MyApp")
    serverPort := cfg.GetIntWithDefault("server.port", 8080)
    debug := cfg.GetBoolWithDefault("app.debug", false)
    fmt.Printf("App: %s, Port: %d, Debug: %v\n", appName, serverPort, debug)

    // 方式3: 只转换部分配置
    type DatabaseConfig struct {
        Host     string
        Port     int
        Username string
        Password string
    }

    var dbConfig DatabaseConfig
    if err := cfg.UnmarshalKey("database", &dbConfig); err != nil {
        panic(err)
    }

    fmt.Printf("Database: %+v\n", dbConfig)
}
```

### 示例 2: 多环境配置

```go
package main

import (
    "os"
    "octopus/pkg/config"
)

func loadConfig() (*config.Config, error) {
    // 获取当前环境
    env := os.Getenv("APP_ENV")
    if env == "" {
        env = "development"
    }

    // 根据环境加载对应的配置文件
    configFile := fmt.Sprintf("config/%s.yaml", env)
    cfg, err := config.Load(configFile)
    if err != nil {
        return nil, err
    }

    return cfg, nil
}

func main() {
    cfg, err := loadConfig()
    if err != nil {
        panic(err)
    }

    // 使用配置（类型安全的默认值）
    port := cfg.GetIntWithDefault("server.port", 8080)
    host := cfg.GetStringWithDefault("server.host", "0.0.0.0")
    fmt.Printf("Server starting on %s:%d\n", host, port)
}
```

### 示例 3: 配置热更新（监听文件变化）

```go
package main

import (
    "log"
    "time"
    "octopus/pkg/config"
)

type ConfigManager struct {
    cfg      *config.Config
    filepath string
    onReload func(*config.Config)
}

func NewConfigManager(filepath string, onReload func(*config.Config)) *ConfigManager {
    cm := &ConfigManager{
        filepath: filepath,
        onReload: onReload,
    }
    
    // 初始加载
    if err := cm.Reload(); err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    return cm
}

func (cm *ConfigManager) Reload() error {
    cfg, err := config.Load(cm.filepath)
    if err != nil {
        return err
    }
    
    cm.cfg = cfg
    
    if cm.onReload != nil {
        cm.onReload(cfg)
    }
    
    return nil
}

func (cm *ConfigManager) Get() *config.Config {
    return cm.cfg
}

func (cm *ConfigManager) Watch(interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    var lastModTime time.Time
    
    for range ticker.C {
        info, err := os.Stat(cm.filepath)
        if err != nil {
            log.Printf("Failed to stat config file: %v", err)
            continue
        }
        
        modTime := info.ModTime()
        if modTime.After(lastModTime) {
            log.Println("Config file changed, reloading...")
            if err := cm.Reload(); err != nil {
                log.Printf("Failed to reload config: %v", err)
            } else {
                lastModTime = modTime
                log.Println("Config reloaded successfully")
            }
        }
    }
}

func main() {
    cm := NewConfigManager("config.yaml", func(cfg *config.Config) {
        log.Println("Config updated!")
        // 在这里处理配置更新逻辑
    })
    
    // 启动配置监听（每10秒检查一次）
    go cm.Watch(10 * time.Second)
    
    // 使用配置
    cfg := cm.Get()
    port := cfg.GetInt("server.port")
    log.Printf("Server starting on port %d\n", port)
    
    // ... 应用逻辑
}
```

## 最佳实践

### 1. 配置文件组织

```text
config/
  ├── default.yaml      # 默认配置
  ├── development.yaml  # 开发环境
  ├── testing.yaml      # 测试环境
  ├── staging.yaml      # 预发布环境
  ├── production.yaml   # 生产环境
  └── local.yaml        # 本地配置（不提交到版本控制）
```

### 2. 敏感信息处理

使用环境变量而不是硬编码：

```yaml
database:
  password: ${DB_PASSWORD}  # 从环境变量读取
  api_key: ${API_KEY}       # 敏感信息不写在配置文件中
```

### 3. 配置验证

```go
func validateConfig(cfg *config.Config) error {
    // 必需的配置项
    required := []string{
        "server.port",
        "database.host",
        "database.name",
    }
    
    for _, key := range required {
        if !cfg.Has(key) {
            return fmt.Errorf("required config %s is missing", key)
        }
    }
    
    // 值范围验证
    port := cfg.GetInt("server.port")
    if port < 1 || port > 65535 {
        return fmt.Errorf("invalid port: %d", port)
    }
    
    return nil
}
```

### 4. 使用 mapstruct 标签

```go
type Config struct {
    ServerPort int    `mapstruct:"server_port"`
    DBHost     string `mapstruct:"db_host"`
    Debug      bool   `mapstruct:"debug_mode"`
}

// 如果需要使用自定义解码器，可以直接使用 mapstruct
decoder := mapstruct.New().WithTagName("mapstruct")
var cfg Config
if err := decoder.Decode(configData.GetAll(), &cfg); err != nil {
    // 处理错误
}
```

## 与 mapstruct 的集成

本配置包完美集成了项目中的 `mapstruct` 包，提供强大的类型转换功能：

- 支持基本类型转换（string、int、bool、float等）
- 支持嵌套结构体
- 支持切片和数组
- 支持指针类型
- 支持时间类型（time.Time）
- 支持自定义标签（mapstruct、json等）

详见 [mapstruct 文档](../mapstruct/README.md)

## 性能建议

1. **配置缓存**: 配置加载后会缓存在内存中，避免重复解析
2. **并发安全**: Config 使用读写锁，支持并发读取
3. **浅拷贝**: GetAll() 和 GetSection() 返回浅拷贝，防止外部修改第一层。在config包的使用场景中，配置通常通过Unmarshal转换为结构体，很少直接修改返回的map，因此浅拷贝已足够且性能更好
4. **按需加载**: 只加载需要的配置文件，避免加载整个目录

## 常见问题

### Q: 如何处理配置文件不存在的情况？

A: `Load` 函数在文件不存在时会返回错误。可以先检查文件是否存在：

```go
configFile := "config.yaml"
if _, err := os.Stat(configFile); err != nil {
    // 使用默认配置文件
    configFile = "config/default.yaml"
}
cfg, err := config.Load(configFile)
```

### Q: 如何支持其他配置格式？

A: 实现自己的 Parser：

```go
type CustomParser struct{}

func (p CustomParser) Parse(data []byte) (map[string]any, error) {
    // 实现自定义解析逻辑
    return result, nil
}
```

### Q: 配置是否支持热更新？

A: 包本身不提供文件监听功能，但可以配合 fsnotify 等库实现，参考"示例 3"。

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 相关资源

- [mapstruct 文档](../mapstruct/README.md) - Map 转 Struct 工具
- [测试用例](./config_test.go) - 完整的测试示例
