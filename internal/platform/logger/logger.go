package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/yoosuf/hopper/internal/platform/config"
)

// LogLevel represents the severity of a log message
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// Logger is the interface for logging
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	With(fields ...Field) Logger
	WithContext(ctx context.Context) Logger
}

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value interface{}
}

// F creates a new Field
func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// loggerImpl is the concrete implementation of Logger
type loggerImpl struct {
	writer  io.Writer
	level   LogLevel
	fields  []Field
	context context.Context
	format  string
}

// New creates a new logger
func New(cfg *config.Config) Logger {
	var writer io.Writer = os.Stdout
	format := cfg.Logging.Format
	if format == "" {
		format = "json"
	}

	return &loggerImpl{
		writer:  writer,
		level:   LogLevel(cfg.Logging.Level),
		format:  format,
		fields:  make([]Field, 0),
		context: context.Background(),
	}
}

// Debug logs a debug message
func (l *loggerImpl) Debug(msg string, fields ...Field) {
	l.log(LevelDebug, msg, fields...)
}

// Info logs an info message
func (l *loggerImpl) Info(msg string, fields ...Field) {
	l.log(LevelInfo, msg, fields...)
}

// Warn logs a warning message
func (l *loggerImpl) Warn(msg string, fields ...Field) {
	l.log(LevelWarn, msg, fields...)
}

// Error logs an error message
func (l *loggerImpl) Error(msg string, fields ...Field) {
	l.log(LevelError, msg, fields...)
}

// With returns a new logger with additional fields
func (l *loggerImpl) With(fields ...Field) Logger {
	newFields := make([]Field, len(l.fields)+len(fields))
	copy(newFields, l.fields)
	copy(newFields[len(l.fields):], fields)
	return &loggerImpl{
		writer:  l.writer,
		level:   l.level,
		fields:  newFields,
		context: l.context,
		format:  l.format,
	}
}

// WithContext returns a new logger with context
func (l *loggerImpl) WithContext(ctx context.Context) Logger {
	return &loggerImpl{
		writer:  l.writer,
		level:   l.level,
		fields:  l.fields,
		context: ctx,
		format:  l.format,
	}
}

// log is the internal logging method
func (l *loggerImpl) log(level LogLevel, msg string, fields ...Field) {
	if !l.shouldLog(level) {
		return
	}

	allFields := make([]Field, len(l.fields)+len(fields))
	copy(allFields, l.fields)
	copy(allFields[len(l.fields):], fields)

	timestamp := time.Now().UTC()

	if l.format == "json" {
		l.logJSON(level, timestamp, msg, allFields)
	} else {
		l.logText(level, timestamp, msg, allFields)
	}
}

// shouldLog determines if a message should be logged based on level
func (l *loggerImpl) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LevelDebug: 0,
		LevelInfo:  1,
		LevelWarn:  2,
		LevelError: 3,
	}
	return levels[level] >= levels[l.level]
}

// logJSON logs in JSON format
func (l *loggerImpl) logJSON(level LogLevel, timestamp time.Time, msg string, fields []Field) {
	fmt.Fprintf(l.writer, `{"level":"%s","time":"%s","message":"%s"`, level, timestamp.Format(time.RFC3339), msg)
	for _, f := range fields {
		fmt.Fprintf(l.writer, `,"%s":%v`, f.Key, formatValue(f.Value))
	}
	fmt.Fprintln(l.writer, "}")
}

// logText logs in text format
func (l *loggerImpl) logText(level LogLevel, timestamp time.Time, msg string, fields []Field) {
	fmt.Fprintf(l.writer, "[%s] %s %s", level, timestamp.Format(time.RFC3339), msg)
	for _, f := range fields {
		fmt.Fprintf(l.writer, " %s=%v", f.Key, formatValue(f.Value))
	}
	fmt.Fprintln(l.writer)
}

// formatValue formats a value for logging, with sensitive data redaction
func formatValue(v interface{}) interface{} {
	// Redact sensitive values
	if str, ok := v.(string); ok {
		if isSensitive(str) {
			return "[REDACTED]"
		}
	}
	return v
}

// isSensitive checks if a string value appears to be sensitive
func isSensitive(s string) bool {
	sensitiveKeywords := []string{
		"password", "secret", "token", "api_key", "private_key",
		"authorization", "credit_card", "ssn", "social_security",
	}
	for _, kw := range sensitiveKeywords {
		if contains(s, kw) {
			return true
		}
	}
	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	sLower := strings.ToLower(s)
	substrLower := strings.ToLower(substr)
	return strings.Contains(sLower, substrLower)
}
