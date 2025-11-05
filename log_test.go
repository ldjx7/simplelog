package simplelog

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestInitSimpleLog(t *testing.T) {
	tmpDir := t.TempDir()

	config := &Config{
		LogDir:   tmpDir,
		LogLevel: LevelDebug,
		LogFile:  "test",
		MaxSize:  10,
		MaxAge:   3,
	}

	InitSimpleLog(config)

	if defaultCfg.LogDir != tmpDir {
		t.Errorf("Expected LogDir %s, got %s", tmpDir, defaultCfg.LogDir)
	}

	if defaultCfg.LogLevel != LevelDebug {
		t.Errorf("Expected LogLevel %s, got %s", LevelDebug, defaultCfg.LogLevel)
	}

	files, err := filepath.Glob(filepath.Join(tmpDir, "test_*.log"))
	if err != nil {
		t.Fatalf("Failed to list log files: %v", err)
	}

	if len(files) == 0 {
		t.Error("Expected at least one log file to be created")
	}
}

func TestLogLevels(t *testing.T) {
	tmpDir := t.TempDir()

	config := &Config{
		LogDir:   tmpDir,
		LogLevel: LevelInfo,
		LogFile:  "level_test",
	}

	InitSimpleLog(config)

	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message")

	Close()

	files, _ := filepath.Glob(filepath.Join(tmpDir, "level_test_*.log"))
	if len(files) == 0 {
		t.Fatal("No log file created")
	}

	content, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	if strings.Contains(logContent, "DEBUG") {
		t.Error("DEBUG message should not be logged when level is INFO")
	}

	if !strings.Contains(logContent, "INFO") {
		t.Error("INFO message should be logged")
	}

	if !strings.Contains(logContent, "WARNING") {
		t.Error("WARN message should be logged")
	}

	if !strings.Contains(logContent, "ERROR") {
		t.Error("ERROR message should be logged")
	}
}

func TestConcurrentWrites(t *testing.T) {
	tmpDir := t.TempDir()

	config := &Config{
		LogDir:   tmpDir,
		LogLevel: LevelInfo,
		LogFile:  "concurrent_test",
	}

	InitSimpleLog(config)

	var wg sync.WaitGroup
	numGoroutines := 100
	messagesPerGoroutine := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				Info("goroutine %d message %d", id, j)
			}
		}(i)
	}

	wg.Wait()
	Close()

	files, _ := filepath.Glob(filepath.Join(tmpDir, "concurrent_test_*.log"))
	if len(files) == 0 {
		t.Fatal("No log file created")
	}

	content, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	nonEmptyLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines++
		}
	}

	expectedLines := numGoroutines * messagesPerGoroutine
	if nonEmptyLines < expectedLines {
		t.Errorf("Expected at least %d log lines, got %d", expectedLines, nonEmptyLines)
	}
}

func TestLogFileRotation(t *testing.T) {
	tmpDir := t.TempDir()

	config := &Config{
		LogDir:   tmpDir,
		LogLevel: LevelInfo,
		LogFile:  "rotation_test",
		MaxSize:  1,
	}

	InitSimpleLog(config)

	longMessage := strings.Repeat("A", 100000)
	for i := 0; i < 20; i++ {
		Info(longMessage)
	}

	Close()

	files, _ := filepath.Glob(filepath.Join(tmpDir, "rotation_test_*.log"))
	if len(files) < 2 {
		t.Error("Expected multiple log files due to size rotation")
	}
}

func TestTimeFormat(t *testing.T) {
	timeStr := getTimeString()

	if len(timeStr) != 23 {
		t.Errorf("Expected time string length 23, got %d: %s", len(timeStr), timeStr)
	}

	parts := strings.Split(timeStr, " ")
	if len(parts) != 2 {
		t.Errorf("Expected 2 parts in time string, got %d", len(parts))
	}

	datePart := parts[0]
	timePart := parts[1]

	if len(strings.Split(datePart, "-")) != 3 {
		t.Errorf("Invalid date format: %s", datePart)
	}

	if len(strings.Split(timePart, ":")) != 3 {
		t.Errorf("Invalid time format: %s", timePart)
	}

	if !strings.Contains(timePart, ".") {
		t.Error("Time should contain milliseconds")
	}
}

func TestLogBuild(t *testing.T) {
	builder := NewLogBuilder(levelInfo, "test message: %s %d", "", "hello", 123)
	result := builder.build()

	if !strings.Contains(result, "INFO") {
		t.Error("Log should contain INFO level")
	}

	if !strings.Contains(result, "test message: hello 123") {
		t.Error("Log should contain formatted message")
	}

	if !strings.Contains(result, ":") {
		t.Error("Log should contain caller info with line number")
	}
}

func TestFormatMessage(t *testing.T) {
	result := formatMessage("hello %s %d", "world", 123)
	expected := "hello world 123"

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestCleanOldLogs(t *testing.T) {
	tmpDir := t.TempDir()

	oldFile := filepath.Join(tmpDir, "test_2020-01-01.log")
	if err := os.WriteFile(oldFile, []byte("old log"), 0644); err != nil {
		t.Fatalf("Failed to create old log file: %v", err)
	}

	oldTime := time.Now().AddDate(0, 0, -10)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to change file time: %v", err)
	}

	config := &Config{
		LogDir:   tmpDir,
		LogLevel: LevelInfo,
		LogFile:  "test",
		MaxAge:   7,
	}

	InitSimpleLog(config)
	Info("test message")
	Close()

	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("Old log file should have been deleted")
	}
}

func BenchmarkInfoLogging(b *testing.B) {
	tmpDir := b.TempDir()

	config := &Config{
		LogDir:   tmpDir,
		LogLevel: LevelInfo,
		LogFile:  "bench",
	}

	InitSimpleLog(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Info("benchmark message %d", i)
	}
	b.StopTimer()

	Close()
}

func BenchmarkConcurrentLogging(b *testing.B) {
	tmpDir := b.TempDir()

	config := &Config{
		LogDir:   tmpDir,
		LogLevel: LevelInfo,
		LogFile:  "bench_concurrent",
	}

	InitSimpleLog(config)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			Info("concurrent message %d", i)
			i++
		}
	})
	b.StopTimer()

	Close()
}
