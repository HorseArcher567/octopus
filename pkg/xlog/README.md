# Logger - 简易 slog 工厂

`logger` 包提供了一个简单的工厂函数，用于创建配置好的 `slog.Logger`。

## 特性

- ✅ 支持 JSON 和 Text 格式
- ✅ 支持多种日志级别
- ✅ 可选源码位置
- ✅ 支持文件输出和日志轮转

## 快速开始

```go
package main

import (
    "log/slog"
    "github.com/HorseArcher567/octopus/pkg/logger"
)

func main() {
    // 创建 logger
    log, closer := logger.MustNew(logger.Config{
        Level:  "info",
        Format: "json",
    })
    if closer != nil {
        defer closer.Close()
    }

    // 设为默认 logger
    slog.SetDefault(log)

    // 使用
    slog.Info("application started", "version", "1.0.0")
}
```

## 配置

```go
type Config struct {
    // Level 日志级别：debug/info/warn/error（默认 info）
    Level string

    // Format 日志格式：json/text（默认 text）
    Format string

    // AddSource 是否添加源码位置（文件名、行号）
    AddSource bool

    // Output 输出目标：stdout/stderr/文件路径（默认 stdout）
    // 文件输出自动启用按天轮转
    Output string

    // MaxAge 日志保留天数（仅文件输出有效），0 表示不删除
    MaxAge int
}
```

## API

### New

```go
func New(cfg Config) (*slog.Logger, io.Closer, error)
```

创建 `slog.Logger`，返回错误供调用方处理。

### MustNew

```go
func MustNew(cfg Config) (*slog.Logger, io.Closer)
```

创建 `slog.Logger`，失败时 panic。推荐用于应用启动阶段。

## 使用示例

### 输出到控制台

```go
log, _ := logger.MustNew(logger.Config{
    Level:     "debug",
    Format:    "text",
    AddSource: true,
})
slog.SetDefault(log)
```

### 输出到文件（自动按天轮转）

```go
log, closer := logger.MustNew(logger.Config{
    Level:  "info",
    Format: "json",
    Output: "/var/log/app.log",
    MaxAge: 7, // 保留7天
})
defer closer.Close()
slog.SetDefault(log)
```

### 创建子 logger

```go
log, _ := logger.MustNew(logger.Config{Level: "info"})

// 添加固定字段
serviceLogger := log.With("service", "user-api")
serviceLogger.Info("request received", "path", "/api/users")
```

## 设计理念

- **简单**：只做一件事——创建配置好的 `slog.Logger`
- **标准**：返回标准库的 `slog.Logger`，无额外封装
- **灵活**：用户自行管理 logger 实例和生命周期
