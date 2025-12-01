package logger

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

var levelNames = map[Level]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

var levelColors = map[Level]string{
	DEBUG: "\033[36m", // Cyan
	INFO:  "\033[32m", // Green
	WARN:  "\033[33m", // Yellow
	ERROR: "\033[31m", // Red
	FATAL: "\033[35m", // Magenta
}

const resetColor = "\033[0m"

type Logger struct {
	mu          sync.Mutex
	out         io.Writer
	level       Level
	serviceName string
	useColor    bool
}

var (
	defaultLogger *Logger
	once          sync.Once
)

type Config struct {
	ServiceName string
	Level       string
	Output      io.Writer
	UseColor    bool
}

func New(cfg Config) *Logger {
	level := parseLevel(cfg.Level)
	out := cfg.Output
	if out == nil {
		out = os.Stdout
	}

	return &Logger{
		out:         out,
		level:       level,
		serviceName: cfg.ServiceName,
		useColor:    cfg.UseColor,
	}
}

func Init(cfg Config) *Logger {
	once.Do(func() {
		defaultLogger = New(cfg)
	})
	return defaultLogger
}

// NewNop returns a no-op logger for testing purposes
func NewNop() *Logger {
	return &Logger{
		out:         io.Discard,
		level:       FATAL + 1, // Higher than any level, so nothing gets logged
		serviceName: "test",
		useColor:    false,
	}
}

func Default() *Logger {
	if defaultLogger == nil {
		defaultLogger = New(Config{
			ServiceName: "app",
			Level:       "INFO",
			Output:      os.Stdout,
			UseColor:    true,
		})
	}
	return defaultLogger
}

func parseLevel(level string) Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return INFO
	}
}

func (l *Logger) log(level Level, msg string, fields ...Field) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	levelStr := levelNames[level]

	var sb strings.Builder

	// Format: timestamp | level | service | message | key=value key=value
	if l.useColor {
		sb.WriteString(fmt.Sprintf("%s | %s%-5s%s | %s | %s",
			timestamp,
			levelColors[level],
			levelStr,
			resetColor,
			l.serviceName,
			msg,
		))
	} else {
		sb.WriteString(fmt.Sprintf("%s | %-5s | %s | %s",
			timestamp,
			levelStr,
			l.serviceName,
			msg,
		))
	}

	if len(fields) > 0 {
		sb.WriteString(" |")
		for _, f := range fields {
			sb.WriteString(fmt.Sprintf(" %s=%v", f.Key, f.Value))
		}
	}

	sb.WriteString("\n")
	fmt.Fprint(l.out, sb.String())

	if level == FATAL {
		os.Exit(1)
	}
}

func (l *Logger) Debug(msg string, fields ...Field) {
	l.log(DEBUG, msg, fields...)
}

func (l *Logger) Info(msg string, fields ...Field) {
	l.log(INFO, msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...Field) {
	l.log(WARN, msg, fields...)
}

func (l *Logger) Error(msg string, fields ...Field) {
	l.log(ERROR, msg, fields...)
}

func (l *Logger) Fatal(msg string, fields ...Field) {
	l.log(FATAL, msg, fields...)
}

func (l *Logger) With(fields ...Field) *Logger {
	return l
}

func (l *Logger) WithCaller() *Logger {
	return l
}

func (l *Logger) Sync() error {
	return nil
}

// Field represents a log field
type Field struct {
	Key   string
	Value interface{}
}

// Field constructors
func String(key, val string) Field {
	return Field{Key: key, Value: val}
}

func Int(key string, val int) Field {
	return Field{Key: key, Value: val}
}

func Int64(key string, val int64) Field {
	return Field{Key: key, Value: val}
}

func Uint(key string, val uint) Field {
	return Field{Key: key, Value: val}
}

func Float64(key string, val float64) Field {
	return Field{Key: key, Value: val}
}

func Bool(key string, val bool) Field {
	return Field{Key: key, Value: val}
}

func Err(err error) Field {
	if err == nil {
		return Field{Key: "error", Value: "nil"}
	}
	return Field{Key: "error", Value: err.Error()}
}

func Duration(key string, val time.Duration) Field {
	return Field{Key: key, Value: val.String()}
}

func Any(key string, val interface{}) Field {
	return Field{Key: key, Value: val}
}

func Caller() Field {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return Field{Key: "caller", Value: "unknown"}
	}
	// Get only the file name, not full path
	parts := strings.Split(file, "/")
	shortFile := parts[len(parts)-1]
	return Field{Key: "caller", Value: fmt.Sprintf("%s:%d", shortFile, line)}
}

// Package-level functions that use the default logger
func Debug(msg string, fields ...Field) {
	Default().Debug(msg, fields...)
}

func Info(msg string, fields ...Field) {
	Default().Info(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	Default().Warn(msg, fields...)
}

func Error(msg string, fields ...Field) {
	Default().Error(msg, fields...)
}

func Fatal(msg string, fields ...Field) {
	Default().Fatal(msg, fields...)
}
