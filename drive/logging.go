package drive

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// LogLevel defines the verbosity of logging
type LogLevel int

const (
	LogLevelSilent LogLevel = iota
	LogLevelError
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
	LogLevelTrace
)

var logLevelNames = map[LogLevel]string{
	LogLevelSilent: "SILENT",
	LogLevelError:  "ERROR",
	LogLevelWarn:   "WARN",
	LogLevelInfo:   "INFO",
	LogLevelDebug:  "DEBUG",
	LogLevelTrace:  "TRACE",
}

// Logger handles the client's logging functionality
type Logger struct {
	level  LogLevel
	logger *log.Logger
	output io.Writer
}

// NewLogger creates a new logger with the specified level
func NewLogger(level LogLevel, output io.Writer) *Logger {
	if output == nil {
		output = os.Stderr
	}
	return &Logger{
		level:  level,
		logger: log.New(output, "", log.LstdFlags),
		output: output,
	}
}

// SetLevel changes the current log level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// GetLevel returns the current log level
func (l *Logger) GetLevel() LogLevel {
	return l.level
}

// SetOutput changes the output destination
func (l *Logger) SetOutput(output io.Writer) {
	l.output = output
	l.logger.SetOutput(output)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level >= LogLevelError {
		l.logger.Printf("[ERROR] "+format, args...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level >= LogLevelWarn {
		l.logger.Printf("[WARN] "+format, args...)
	}
}

// Info logs an informational message
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level >= LogLevelInfo {
		l.logger.Printf("[INFO] "+format, args...)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level >= LogLevelDebug {
		l.logger.Printf("[DEBUG] "+format, args...)
	}
}

// Trace logs a trace message (very verbose)
func (l *Logger) Trace(format string, args ...interface{}) {
	if l.level >= LogLevelTrace {
		l.logger.Printf("[TRACE] "+format, args...)
	}
}

// ParseLogLevel converts a string to LogLevel
func ParseLogLevel(level string) (LogLevel, error) {
	switch strings.ToUpper(level) {
	case "SILENT":
		return LogLevelSilent, nil
	case "ERROR":
		return LogLevelError, nil
	case "WARN", "WARNING":
		return LogLevelWarn, nil
	case "INFO":
		return LogLevelInfo, nil
	case "DEBUG":
		return LogLevelDebug, nil
	case "TRACE":
		return LogLevelTrace, nil
	default:
		return LogLevelInfo, fmt.Errorf("unknown log level: %s", level)
	}
}
