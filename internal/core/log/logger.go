package corelog

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"opskit/internal/core/fsx"
	"opskit/internal/core/timeutil"
)

type Level string

const (
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

type Logger struct {
	mu      sync.Mutex
	console io.Writer
	file    io.Writer
}

func New(console io.Writer, file io.Writer) *Logger {
	if console == nil {
		console = os.Stdout
	}
	return &Logger{console: console, file: file}
}

func NewWithFile(logPath string) (*Logger, io.Closer, error) {
	if err := fsx.EnsureDir(filepath.Dir(logPath), 0o755); err != nil {
		return nil, nil, err
	}
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, nil, err
	}
	return New(os.Stdout, f), f, nil
}

func (l *Logger) Info(msg string) {
	l.write(LevelInfo, msg)
}

func (l *Logger) Warn(msg string) {
	l.write(LevelWarn, msg)
}

func (l *Logger) Error(msg string) {
	l.write(LevelError, msg)
}

func (l *Logger) write(level Level, msg string) {
	if l == nil {
		return
	}
	line := fmt.Sprintf("%s [%s] %s\n", timeutil.NowISO8601(), level, msg)

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.console != nil {
		_, _ = io.WriteString(l.console, line)
	}
	if l.file != nil {
		_, _ = io.WriteString(l.file, line)
	}
}
