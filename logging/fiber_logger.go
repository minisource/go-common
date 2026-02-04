package logging

import (
	"fmt"

	"github.com/gofiber/fiber/v2/log"
)

// fiberLogger implements the Logger interface for Fiber.
type fiberLogger struct {
	cfg *LoggerConfig
}

// newFiberLogger creates a new Fiber logger instance.
func newFiberLogger(cfg *LoggerConfig) *fiberLogger {
	return &fiberLogger{cfg: cfg}
}

// Init initializes the Fiber logger.
func (l *fiberLogger) Init() {
	// Fiber's logging is handled by the middleware, so no explicit initialization is needed here.
	// You can configure Fiber's logger in the Fiber app setup.
}

// Debug logs a debug message.
func (l *fiberLogger) Debug(cat Category, sub SubCategory, msg string, extra map[ExtraKey]interface{}) {
	l.log(cat, sub, msg, extra, "DEBUG")
}

// Debugf logs a formatted debug message.
func (l *fiberLogger) Debugf(template string, args ...interface{}) {
	l.logf(template, args, "DEBUG")
}

// Info logs an info message.
func (l *fiberLogger) Info(cat Category, sub SubCategory, msg string, extra map[ExtraKey]interface{}) {
	l.log(cat, sub, msg, extra, "INFO")
}

// Infof logs a formatted info message.
func (l *fiberLogger) Infof(template string, args ...interface{}) {
	l.logf(template, args, "INFO")
}

// Warn logs a warning message.
func (l *fiberLogger) Warn(cat Category, sub SubCategory, msg string, extra map[ExtraKey]interface{}) {
	l.log(cat, sub, msg, extra, "WARN")
}

// Warnf logs a formatted warning message.
func (l *fiberLogger) Warnf(template string, args ...interface{}) {
	l.logf(template, args, "WARN")
}

// Error logs an error message.
func (l *fiberLogger) Error(cat Category, sub SubCategory, msg string, extra map[ExtraKey]interface{}) {
	l.log(cat, sub, msg, extra, "ERROR")
}

// Errorf logs a formatted error message.
func (l *fiberLogger) Errorf(template string, args ...interface{}) {
	l.logf(template, args, "ERROR")
}

// Fatal logs a fatal message.
func (l *fiberLogger) Fatal(cat Category, sub SubCategory, msg string, extra map[ExtraKey]interface{}) {
	l.log(cat, sub, msg, extra, "FATAL")
}

// Fatalf logs a formatted fatal message.
func (l *fiberLogger) Fatalf(template string, args ...interface{}) {
	l.logf(template, args, "FATAL")
}

// log is a helper function to log messages with a specific level.
func (l *fiberLogger) log(cat Category, sub SubCategory, msg string, extra map[ExtraKey]interface{}, level string) {
	if extra == nil {
		extra = make(map[ExtraKey]interface{})
	}
	extra["Category"] = cat
	extra["SubCategory"] = sub

	// Format the log message
	logMessage := fmt.Sprintf("[%s] %s: %s %v", level, cat, msg, extra)

	// Log the message using Fiber's logger
	switch level {
	case "DEBUG":
		log.Debug(logMessage)
	case "INFO":
		log.Info(logMessage)
	case "WARN":
		log.Warn(logMessage)
	case "ERROR":
		log.Error(logMessage)
	case "FATAL":
		log.Fatal(logMessage)
	}
}

// logf is a helper function to log formatted messages with a specific level.
func (l *fiberLogger) logf(template string, args []interface{}, level string) {
	msg := fmt.Sprintf(template, args...)
	l.log("", "", msg, nil, level)
}