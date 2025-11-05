package simplelog

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"
)

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

	var builder strings.Builder
	builder.Grow(256)

	builder.WriteString(getTimeString())
	builder.WriteByte(' ')
	builder.WriteString(levelNameMap[l.level])
	builder.WriteByte(' ')
	builder.WriteString(name)
	builder.WriteByte(':')
	builder.WriteString(strconv.Itoa(line))
	builder.WriteByte('-')

	if len(l.v) > 0 {
		builder.WriteString(formatMessage(l.format, l.v...))
	} else {
		builder.WriteString(l.format)
	}

	if l.errStack != "" {
		builder.WriteByte('\n')
		builder.WriteString(l.errStack)
	}

	return builder.String()
}

func formatMessage(format string, v ...any) string {
	return fmt.Sprintf(format, v...)
}

func callerInfoSplice() (string, int) {
	pc, _, line, _ := runtime.Caller(3)
	return runtime.FuncForPC(pc).Name(), line
}

func getTimeString() string {
	now := time.Now()
	year, month, day := now.Date()
	hour, min, sec := now.Clock()
	ms := now.Nanosecond() / 1000000

	var builder strings.Builder
	builder.Grow(23)

	appendInt(&builder, year, 4)
	builder.WriteByte('-')
	appendInt(&builder, int(month), 2)
	builder.WriteByte('-')
	appendInt(&builder, day, 2)
	builder.WriteByte(' ')
	appendInt(&builder, hour, 2)
	builder.WriteByte(':')
	appendInt(&builder, min, 2)
	builder.WriteByte(':')
	appendInt(&builder, sec, 2)
	builder.WriteByte('.')
	appendInt(&builder, ms, 3)

	return builder.String()
}

func appendInt(builder *strings.Builder, value, width int) {
	s := strconv.Itoa(value)
	for i := len(s); i < width; i++ {
		builder.WriteByte('0')
	}
	builder.WriteString(s)
}
