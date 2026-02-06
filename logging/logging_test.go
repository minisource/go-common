package logging

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLoggerZap(t *testing.T) {
	cfg := &LoggerConfig{
		Level:       "debug",
		Logger:      "zap",
		Encoding:    "json",
		ConsoleOnly: true,
	}

	logger := NewLogger(cfg)
	assert.NotNil(t, logger)
}

func TestNewLoggerZerolog(t *testing.T) {
	cfg := &LoggerConfig{
		Level:       "info",
		Logger:      "zerolog",
		ConsoleOnly: true,
	}

	logger := NewLogger(cfg)
	assert.NotNil(t, logger)
}

func TestNewLoggerFiber(t *testing.T) {
	cfg := &LoggerConfig{
		Level:       "info",
		Logger:      "fiber",
		ConsoleOnly: true,
	}

	logger := NewLogger(cfg)
	assert.NotNil(t, logger)
}

func TestLoggerWithFilePath(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	cfg := &LoggerConfig{
		Level:       "debug",
		Logger:      "zap",
		Encoding:    "json",
		FilePath:    logFile,
		ConsoleOnly: false,
	}

	logger := NewLogger(cfg)
	logger.Init()

	extra := map[ExtraKey]interface{}{
		"userId": "123",
	}
	logger.Info(General, Api, "Test log message", extra)

	// Verify file was created
	_, err := os.Stat(logFile)
	assert.NoError(t, err)
}

func TestLoggerLevels(t *testing.T) {
	cfg := &LoggerConfig{
		Level:       "debug",
		Logger:      "zap",
		Encoding:    "console",
		ConsoleOnly: true,
	}

	logger := NewLogger(cfg)
	logger.Init()

	extra := map[ExtraKey]interface{}{
		"key": "value",
	}

	// These should not panic
	logger.Debug(General, Api, "Debug message", extra)
	logger.Info(General, Api, "Info message", extra)
	logger.Warn(General, Api, "Warn message", extra)
	logger.Error(General, Api, "Error message", extra)
}

func TestLoggerFormattedOutput(t *testing.T) {
	cfg := &LoggerConfig{
		Level:       "debug",
		Logger:      "zap",
		Encoding:    "console",
		ConsoleOnly: true,
	}

	logger := NewLogger(cfg)
	logger.Init()

	// These should not panic
	logger.Debugf("Debug: %s", "test")
	logger.Infof("Info: %s", "test")
	logger.Warnf("Warn: %s", "test")
	logger.Errorf("Error: %s", "test")
}

func TestCategories(t *testing.T) {
	// Test that categories are defined
	assert.NotEqual(t, "", string(General))
	assert.NotEqual(t, "", string(Internal))
	assert.NotEqual(t, "", string(Postgres))
}

func TestSubCategories(t *testing.T) {
	// Test that subcategories are defined
	assert.NotEqual(t, "", string(Api))
	assert.NotEqual(t, "", string(ExternalService))
}
