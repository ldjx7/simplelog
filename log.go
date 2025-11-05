// Package simplelog 提供简单易用的日志功能，支持日志分级、自动轮转和旧日志清理
package simplelog

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"
)

var once sync.Once

// Config 日志配置结构体
type Config struct {
	LogDir     string
	LogLevel   string
	LogFile    string
	MaxSize    int64 // 单个日志文件最大大小（MB），0 表示不限制
	MaxAge     int   // 日志文件保留天数，0 表示不清理
	MaxBackups int   // 保留的旧日志文件数量，0 表示不限制
}

var (
	logFile     *os.File
	logFileName string = "root"
	mu          sync.Mutex
)

var defaultCfg = &Config{
	LogDir:   defaultLogDir,
	LogLevel: LevelInfo,
	MaxSize:  100,
	MaxAge:   7,
}

// 默认初始化函数，确保只执行一次
func defaultInit() {
	once.Do(func() {
		InitSimpleLog(defaultCfg) // 使用默认配置初始化
	})
}

// InitSimpleLog 日志初始化函数
func InitSimpleLog(config *Config) {
	if config != nil {
		if config.LogDir != "" {
			defaultCfg.LogDir = config.LogDir
		}
		if config.LogLevel != "" {
			if _, ok := LevelMap[config.LogLevel]; !ok {
				log.Fatalf("[error] Invalid log level: %s", config.LogLevel)
			}
			defaultCfg.LogLevel = config.LogLevel
		}
		if config.LogFile != "" {
			logFileName = config.LogFile
		}
		if config.MaxSize > 0 {
			defaultCfg.MaxSize = config.MaxSize
		}
		if config.MaxAge > 0 {
			defaultCfg.MaxAge = config.MaxAge
		}
		if config.MaxBackups > 0 {
			defaultCfg.MaxBackups = config.MaxBackups
		}
	}

	level = LevelMap[defaultCfg.LogLevel]

	if err := os.MkdirAll(defaultCfg.LogDir, 0755); err != nil {
		log.Fatalf("[error] Make log dir error: %v", err)
	}

	writer := &dailyLogWriter{
		logDir:     defaultCfg.LogDir,
		maxSize:    defaultCfg.MaxSize * 1024 * 1024,
		maxAge:     defaultCfg.MaxAge,
		maxBackups: defaultCfg.MaxBackups,
		currentDate: time.Now().Format("2006-01-02"),
	}

	if err := writer.switchLogFile(); err != nil {
		log.Fatalf("[error] Switch log file error: %v", err)
	}

	multiWriters := io.MultiWriter(writer, os.Stdout)
	log.SetOutput(multiWriters)
	log.SetFlags(0)
}

type dailyLogWriter struct {
	logDir      string
	maxSize     int64
	maxAge      int
	maxBackups  int
	currentDate string
}

func (w *dailyLogWriter) Write(p []byte) (n int, err error) {
	mu.Lock()
	defer mu.Unlock()

	newDate := time.Now().Format("2006-01-02")
	if newDate != w.currentDate {
		w.currentDate = newDate
		if err := w.switchLogFile(); err != nil {
			return 0, err
		}
	}

	if w.maxSize > 0 && logFile != nil {
		if stat, err := logFile.Stat(); err == nil {
			if stat.Size()+int64(len(p)) > w.maxSize {
				if err := w.rotateLogFile(); err != nil {
					return 0, err
				}
			}
		}
	}

	if logFile == nil {
		return 0, fmt.Errorf("log file is not initialized")
	}

	return logFile.Write(p)
}

func (w *dailyLogWriter) switchLogFile() error {
	if logFile != nil {
		if err := logFile.Close(); err != nil {
			return fmt.Errorf("close log file: %w", err)
		}
	}

	logfileName := fmt.Sprintf("%s_%s.log", logFileName, w.currentDate)
	fullPath := filepath.Join(w.logDir, logfileName)

	file, err := os.OpenFile(fullPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}

	logFile = file
	w.cleanOldLogs()
	return nil
}

func (w *dailyLogWriter) rotateLogFile() error {
	if logFile != nil {
		if err := logFile.Close(); err != nil {
			return fmt.Errorf("close log file: %w", err)
		}
	}

	timestamp := time.Now().Format("20060102-150405")
	oldName := filepath.Join(w.logDir, fmt.Sprintf("%s_%s.log", logFileName, w.currentDate))
	newName := filepath.Join(w.logDir, fmt.Sprintf("%s_%s-%s.log", logFileName, w.currentDate, timestamp))

	if err := os.Rename(oldName, newName); err != nil {
		return fmt.Errorf("rename log file: %w", err)
	}

	file, err := os.OpenFile(oldName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("create new log file: %w", err)
	}

	logFile = file
	w.cleanOldLogs()
	return nil
}

func (w *dailyLogWriter) cleanOldLogs() {
	if w.maxAge <= 0 && w.maxBackups <= 0 {
		return
	}

	files, err := filepath.Glob(filepath.Join(w.logDir, fmt.Sprintf("%s_*.log", logFileName)))
	if err != nil {
		return
	}

	type fileInfo struct {
		path    string
		modTime time.Time
	}

	var fileInfos []fileInfo
	for _, file := range files {
		stat, err := os.Stat(file)
		if err != nil {
			continue
		}
		fileInfos = append(fileInfos, fileInfo{path: file, modTime: stat.ModTime()})
	}

	if w.maxAge > 0 {
		cutoff := time.Now().AddDate(0, 0, -w.maxAge)
		for _, fi := range fileInfos {
			if fi.modTime.Before(cutoff) {
				os.Remove(fi.path)
			}
		}
	}

	if w.maxBackups > 0 && len(fileInfos) > w.maxBackups {
		for i := 0; i < len(fileInfos)-1; i++ {
			for j := i + 1; j < len(fileInfos); j++ {
				if fileInfos[i].modTime.Before(fileInfos[j].modTime) {
					fileInfos[i], fileInfos[j] = fileInfos[j], fileInfos[i]
				}
			}
		}
		for i := w.maxBackups; i < len(fileInfos); i++ {
			os.Remove(fileInfos[i].path)
		}
	}
}

// Close 优雅关闭日志文件
func Close() error {
	mu.Lock()
	defer mu.Unlock()

	if logFile != nil {
		if err := logFile.Sync(); err != nil {
			return err
		}
		if err := logFile.Close(); err != nil {
			return err
		}
		logFile = nil
	}
	return nil
}

// Debug 输出 DEBUG 级别日志
func Debug(format string, v ...any) {
	defaultInit()
	if level > levelDebug {
		return
	}
	log.Printf(NewLogBuilder(levelDebug, format, "", v...).build())
}

// Info 输出 INFO 级别日志
func Info(format string, v ...any) {
	defaultInit()
	if level > levelInfo {
		return
	}
	log.Printf(NewLogBuilder(levelInfo, format, "", v...).build())
}

// Warn 输出 WARN 级别日志
func Warn(format string, v ...any) {
	defaultInit()
	if level > levelWarning {
		return
	}
	log.Printf(NewLogBuilder(levelWarning, format, "", v...).build())
}

// Error 输出 ERROR 级别日志，包含堆栈信息
func Error(format string, v ...any) {
	defaultInit()
	if level > levelError {
		return
	}
	log.Printf(NewLogBuilder(levelError, format, string(debug.Stack()), v...).build())
}

// Fatal 输出 FATAL 级别日志，包含堆栈信息，并终止程序
func Fatal(format string, v ...any) {
	defaultInit()
	if level > levelFatal {
		return
	}
	log.Printf(NewLogBuilder(levelFatal, format, string(debug.Stack()), v...).build())
	os.Exit(1)
}
