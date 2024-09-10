package simplelog

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"sync"
	"time"
)

var once sync.Once

// Config 日志配置结构体
type Config struct {
	LogDir   string
	LogLevel string
	LogFile  string
}

var logFile *os.File
var currentDate string
var logFileName string = "root"

// 默认初始化函数，确保只执行一次
func defaultInit() {
	once.Do(func() {
		InitSimpleLog(nil) // 使用默认配置初始化
	})
}

// InitSimpleLog 日志初始化函数
func InitSimpleLog(config *Config) {
	defaultCfg := &Config{
		LogDir:   defaultLogDir,
		LogLevel: LevelInfo,
	}

	if config != nil {
		defaultCfg.LogDir = getOrDefault(config.LogDir, defaultCfg.LogDir)
		defaultCfg.LogLevel = getOrDefault(config.LogLevel, defaultCfg.LogLevel)
		logFileName = getOrDefault(config.LogFile, logFileName)
	}

	// 设置日志级别
	level = LevelMap[defaultCfg.LogLevel]

	// 创建日志目录
	if err := os.MkdirAll(defaultCfg.LogDir, 0766); err != nil {
		log.Fatalf("[error] Make log dir error: %v", err)
	}

	// 初始化当前日期
	currentDate = time.Now().Format("2006-01-02")

	// 设置日志输出
	writer := &dailyLogWriter{logDir: defaultCfg.LogDir}
	writer.switchLogFile()
	multiWriters := io.MultiWriter(writer, os.Stdout)
	log.SetOutput(multiWriters)
	// 关闭默认时间戳和短文件信息
	log.SetFlags(0)
}

// Custom log writer that switches log file based on date
type dailyLogWriter struct {
	logDir string
}

func (w *dailyLogWriter) Write(p []byte) (n int, err error) {
	newDate := time.Now().Format("2006-01-02")
	if newDate != currentDate {
		currentDate = newDate
		w.switchLogFile()
	}
	return logFile.Write(p)
}

func (w *dailyLogWriter) switchLogFile() {
	if logFile != nil {
		logFile.Close()
	}

	logfileName := fmt.Sprintf("%s_%s.log", logFileName, currentDate)
	fullPath := filepath.Join(w.logDir, logfileName)

	var err error
	logFile, err = os.OpenFile(fullPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("[error] Open log file error: %v", err)
		return
	}
}

// 返回非空值。如果目标值为空，则返回传入的值，否则返回目标值。
func getOrDefault[T any](target T, value T) T {
	// 判断目标值是否为零值
	if reflect.DeepEqual(target, reflect.Zero(reflect.TypeOf(target)).Interface()) {
		return value
	}
	return target
}

func Debug(format string, v ...any) {
	defaultInit()
	if level > levelDebug {
		return
	}
	log.Printf(NewLogBuilder(levelDebug, format, "", v...).build())
}

func Info(format string, v ...any) {
	defaultInit()
	if level > levelInfo {
		return
	}
	log.Printf(NewLogBuilder(levelInfo, format, "", v...).build())
}

func Warn(format string, v ...any) {
	defaultInit()
	if level > levelWarning {
		return
	}
	log.Printf(NewLogBuilder(levelWarning, format, "", v...).build())
}

func Error(format string, v ...any) {
	defaultInit()
	if level > levelError {
		return
	}
	log.Printf(NewLogBuilder(levelError, format, string(debug.Stack()), v...).build())
}

func Fatal(format string, v ...any) {
	defaultInit()
	if level > levelFatal {
		return
	}
	log.Printf(NewLogBuilder(levelFatal, format, string(debug.Stack()), v...).build())
	os.Exit(1)
}
