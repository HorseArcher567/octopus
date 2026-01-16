# Rotate - 简易日志轮转包

`rotate` 包提供了一个简单易用的按天日志轮转功能，实现了 `io.WriteCloser` 接口。

## 特性

- ✅ 每天自动轮转日志文件
- ✅ 自动清理过期日志
- ✅ 线程安全的并发写入
- ✅ 人性化的备份文件命名

## 快速开始

```go
package main

import (
    "log"
    "github.com/HorseArcher567/octopus/pkg/rotate"
)

func main() {
    // 创建轮转写入器，保留7天
    writer, closer := rotate.MustNew(rotate.Config{
        Filename: "logs/app.log",
        MaxAge:   7,
    })
    defer closer.Close()

    log.SetOutput(writer)
    log.Println("Hello, World!")
}
```

## 配置

```go
type Config struct {
    // Filename 日志文件路径（必填）
    Filename string

    // MaxAge 保留旧日志文件的最大天数，0 表示不删除
    MaxAge int
}
```

## 文件命名

- **当前日志**：`app.log`
- **备份文件**：`app-2024-01-15.log`

备份文件格式为 `{basename}-{date}{ext}`，日期格式固定为 `YYYY-MM-DD`。

## API

### New

```go
func New(config Config) (io.Writer, io.Closer, error)
```

创建轮转写入器，返回错误供调用方处理。

### MustNew

```go
func MustNew(config Config) (io.Writer, io.Closer)
```

创建轮转写入器，失败时 panic。推荐用于应用启动阶段。

## 与 slog 集成

```go
writer, closer := rotate.MustNew(rotate.Config{
    Filename: "logs/app.log",
    MaxAge:   30,
})
defer closer.Close()

handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
    Level: slog.LevelInfo,
})
logger := slog.New(handler)
logger.Info("application started")
```

## 轮转机制

- 每天第一次写入时，检测到日期变更后自动轮转
- 程序重启时，如果现有日志文件是旧日期的，会立即轮转
- 过期文件清理在后台异步执行
