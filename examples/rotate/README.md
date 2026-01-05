# Rotate Package Example

这个示例展示了如何使用 `rotate` 包进行日志轮转。

## 功能演示

- 基于文件大小的自动轮转（1MB）
- 保留3个备份文件
- 保留7天的日志
- 自动压缩旧日志文件
- 使用本地时间命名备份文件
- 通过 SIGHUP 信号手动触发轮转

## 运行示例

```bash
# 从项目根目录运行
cd examples/rotate
go run main.go
```

## 手动触发轮转

在程序运行时，可以发送 SIGHUP 信号手动触发日志轮转：

```bash
# 查找进程ID
ps aux | grep "go run main.go"

# 发送 SIGHUP 信号
kill -HUP <PID>
```

## 查看日志文件

程序运行后，日志文件会保存在 `logs/` 目录：

```bash
# 查看所有日志文件
ls -lh logs/

# 查看当前日志
tail -f logs/app.log

# 查看压缩的旧日志
zcat logs/app-*.log.gz | less
```

## 期望输出

运行程序后，你应该会看到：

1. `logs/app.log` - 当前活动的日志文件
2. `logs/app-2024-11-21-*.log.gz` - 轮转并压缩的旧日志文件
3. 最多保留3个备份文件（根据配置）

## 配置说明

示例中使用的配置：

```go
rotate.Config{
    Filename:   "logs/app.log",   // 日志文件路径
    MaxSize:    1,                 // 1MB (演示用)
    MaxBackups: 3,                 // 保留3个备份
    MaxAge:     7,                 // 保留7天
    Compress:   true,              // 压缩旧日志
    LocalTime:  true,              // 使用本地时间
}
```

生产环境建议配置：

```go
rotate.Config{
    Filename:   "/var/log/myapp/app.log",
    MaxSize:    100,               // 100MB
    MaxBackups: 30,                // 保留30个备份
    MaxAge:     90,                // 保留90天
    Compress:   true,              // 压缩旧日志
    LocalTime:  false,             // 使用UTC时间
}
```

## 集成方式

### 方式1：与 slog 直接集成

```go
rotator, _ := rotate.New(config)
handler := slog.NewJSONHandler(rotator, opts)
logger := slog.New(handler)
slog.SetDefault(logger)
```

### 方式2：与 octopus/logger 集成

需要修改 logger 包以支持自定义 writer：

```go
rotator, _ := rotate.New(config)
// 传递 rotator 给 logger
```

## 监控日志轮转

可以通过日志输出监控轮转事件：

```bash
# 筛选轮转相关日志
tail -f logs/app.log | grep "rotat"
```

## 磁盘使用

根据配置，磁盘使用量为：

- 当前日志：最大 1MB
- 备份文件：3个 × 1MB = 3MB（压缩后约 0.2-0.5MB）
- 总计：约 1.5-2MB

## 性能测试

示例程序会生成100条日志，每条间隔10ms，用于测试轮转性能：

- 总运行时间：约1秒
- 日志大小：取决于日志内容
- 轮转次数：取决于生成的日志量是否超过1MB

## 故障排查

### 日志文件未创建

检查目录权限：

```bash
mkdir -p logs
chmod 755 logs
```

### 轮转未触发

检查当前日志文件大小：

```bash
ls -lh logs/app.log
```

如果小于 MaxSize，轮转不会触发。

### 压缩失败

检查磁盘空间：

```bash
df -h
```

确保有足够空间完成压缩操作。

