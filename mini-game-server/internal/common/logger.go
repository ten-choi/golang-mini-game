package common

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// Logger provides structured logging capabilities
type Logger struct {
	zapLogger *zap.SugaredLogger
}

var defaultLogger *Logger

func init() {
	// Create production config
	config := zap.NewProductionConfig()
	config.Encoding = "console"
	config.DisableStacktrace = true
	config.DisableCaller = false

	// Build logger
	zapLogger, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize zap logger: %v", err))
	}

	defaultLogger = &Logger{
		zapLogger: zapLogger.Sugar(),
	}
}

// GetLogger returns the default logger
func GetLogger() *Logger {
	return defaultLogger
}

// Info logs informational messages
func (l *Logger) Info(msg string, args ...interface{}) {
	l.zapLogger.Infof(msg, args...)
}

// Error logs error messages
func (l *Logger) Error(msg string, args ...interface{}) {
	l.zapLogger.Errorf(msg, args...)
}

// Debug logs debug messages
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.zapLogger.Debugf(msg, args...)
}

// Warn logs warning messages
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.zapLogger.Warnf(msg, args...)
}

// WithContext returns a logger with context values
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// Extract request ID or other context values if needed
	return l
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() {
	_ = l.zapLogger.Sync()
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

// Sync flushes any buffered log entries (global)
func Sync() {
	defaultLogger.Sync()
}
