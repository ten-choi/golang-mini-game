package logger

import (
	"context"
	"log"
	"os"
)

// Logger provides structured logging capabilities
type Logger struct {
	*log.Logger
}

var defaultLogger *Logger

func init() {
	defaultLogger = &Logger{
		Logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// GetLogger returns the default logger
func GetLogger() *Logger {
	return defaultLogger
}

// Info logs informational messages
func (l *Logger) Info(msg string, args ...interface{}) {
	l.Printf("[INFO] "+msg, args...)
}

// Error logs error messages
func (l *Logger) Error(msg string, args ...interface{}) {
	l.Printf("[ERROR] "+msg, args...)
}

// Debug logs debug messages
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.Printf("[DEBUG] "+msg, args...)
}

// Warn logs warning messages
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.Printf("[WARN] "+msg, args...)
}

// WithContext returns a logger with context values
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// Extract request ID or other context values if needed
	return l
}

// Global convenience functions
func Info(msg string, args ...interface{}) {
	defaultLogger.Info(msg, args...)
}

func Error(msg string, args ...interface{}) {
	defaultLogger.Error(msg, args...)
}

func Debug(msg string, args ...interface{}) {
	defaultLogger.Debug(msg, args...)
}

func Warn(msg string, args ...interface{}) {
	defaultLogger.Warn(msg, args...)
}
