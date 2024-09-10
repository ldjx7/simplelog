package simplelog

import (
	"fmt"
	"runtime"
	"time"
)

const logPattern = "%s %s %s:%d-%s"

const (
	levelDebug = iota
	levelInfo
	levelWarning
	levelError
	levelFatal
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
	LevelFatal = "fatal"
	// 定义默认日志路径
	defaultLogDir = "applog/logs"
)

var (
	levelNameMap = map[int]string{
		levelDebug:   "DEBUG",
		levelInfo:    "INFO",
		levelWarning: "WARNING",
		levelError:   "ERROR",
		levelFatal:   "FATAL",
	}
	level int = levelInfo // 默认日志级别
)

var LevelMap = map[string]int{
	LevelDebug: levelDebug,
	LevelInfo:  levelInfo,
	LevelWarn:  levelWarning,
	LevelError: levelError,
	LevelFatal: levelFatal,
}

type LogBuild struct {
	level    int
	format   string
	errStack string
	v        []any
}

func NewLogBuilder(level int, format, errStack string, v ...any) *LogBuild {
	return &LogBuild{level, format, errStack, v}
}

func (l *LogBuild) build() string {
	name, line := callerInfoSplice()
	msg := fmt.Sprintf(l.format, l.v...)
	if l.errStack != "" {
		msg = fmt.Sprintf("%s\n%s", msg, l.errStack)
	}
	return fmt.Sprintf(logPattern, getTimeString(), levelNameMap[l.level], name, line, msg)
}

func callerInfoSplice() (string, int) {
	pc, _, line, _ := runtime.Caller(3)
	return runtime.FuncForPC(pc).Name(), line
}

func getTimeString() string {
	return time.Now().Format("2006-01-02 15:04:05.000")
}
