# simplelog 
> 对GO日志的简易封装

# Installation

```shell
> go get github.com/ldjx7/simplelog
```

## demo
```go
package main

import (
	"github.com/ldjx7/simplelog"
)

# 显示初始化log配置，指定日志文件路径，文件名称以及日志级别
# 默认路径为applog/logs，日志级别为：info，默认文件名称为root_日期.log
func init() {
	simplelog.InitSimpleLog(&simplelog.Config{
		LogLevel: simplelog.LevelError,
		LogDir:   "logs",
		LogFile:  "log",
	})
}
# 也可以不显示初始化，直接导包使用，所有的配置均为默认
func main() {
	simplelog.Debug("hello world")
	simplelog.Info("hello world")
	simplelog.Warn("hello world")
	simplelog.Error("hello world")
}
```
