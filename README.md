# simplelog
> 对 GO 日志的简易封装

高性能、线程安全的 Go 日志库，支持日志分级、按日期和大小自动轮转、旧日志清理等功能。

## 特性

- 🚀 高性能：使用 strings.Builder 优化字符串拼接
- 🔒 并发安全：使用互斥锁保护文件操作
- 📅 日期轮转：每日自动创建新日志文件
- 📦 大小轮转：支持按文件大小自动轮转
- 🗑️ 自动清理：可配置保留天数和文件数
- 📝 多级日志：支持 Debug/Info/Warn/Error/Fatal 五个级别
- 🎯 零依赖：仅使用 Go 标准库

## 安装

```bash
go get github.com/ldjx7/simplelog
```

## 快速开始

```go
package main

import (
	"github.com/ldjx7/simplelog"
)

// 显式初始化日志配置，指定日志文件路径、文件名称以及日志级别
// 默认路径为 applog/logs，日志级别为 info，默认文件名称为 root_日期.log
func init() {
	simplelog.InitSimpleLog(&simplelog.Config{
		LogLevel:   simplelog.LevelError, // 日志级别
		LogDir:     "logs",                // 日志目录
		LogFile:    "app",                 // 日志文件前缀
		MaxSize:    100,                   // 单个文件最大大小（MB）
		MaxAge:     7,                     // 保留天数
		MaxBackups: 10,                    // 保留文件数
	})
}

// 也可以不显式初始化，直接使用，所有配置均为默认值
func main() {
	simplelog.Debug("debug message: %s", "test")
	simplelog.Info("info message: %d", 123)
	simplelog.Warn("warning message")
	simplelog.Error("error occurred: %v", err)

	// 程序退出前建议调用 Close 确保日志刷新
	defer simplelog.Close()
}
```

## 配置说明

| 参数 | 类型 | 默认值 | 说明 |
|-----|------|--------|------|
| LogLevel | string | "info" | 日志级别：debug/info/warn/error/fatal |
| LogDir | string | "applog/logs" | 日志文件目录 |
| LogFile | string | "root" | 日志文件名前缀 |
| MaxSize | int64 | 100 | 单个日志文件最大大小（MB），0 表示不限制 |
| MaxAge | int | 7 | 日志文件保留天数，0 表示不清理 |
| MaxBackups | int | 0 | 保留的日志文件数量，0 表示不限制 |

## 日志级别

```go
simplelog.LevelDebug  // "debug"
simplelog.LevelInfo   // "info"
simplelog.LevelWarn   // "warn"
simplelog.LevelError  // "error"
simplelog.LevelFatal  // "fatal"
```

## 日志格式

```
2025-11-05 14:30:45.123 INFO main.main:42-info message: 123
```

格式：`{日期时间} {级别} {函数名}:{行号}-{消息}`

## License

Apache License 2.0
