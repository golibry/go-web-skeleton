package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/golibry/go-params/params"
)

type Log struct {
	// LogLevel defines the minimum log level for the application logger.
	// Uses slog.Level which supports debug, info, warn, and error levels.
	LogLevel slog.Level

	// LogPath specifies where log output should be directed.
	// Valid values are: "stdout", "stderr", or a file path.
	LogPath string `env:"APP_LOG_PATH" default:"stdout" validate:"required,logpath"`
}

// Populate implements the go-config Config interface for the top-level Log.
// It reads values from environment variables and sets defaults.
func (c *Log) Populate() error {
	logLevel, _ := params.GetEnvAsString("APP_LOG_LEVEL", "warn")
	logPath, _ := params.GetEnvAsString("APP_LOG_PATH", "stdout")

	c.LogLevel = parseLogLevel(logLevel)
	c.LogPath = logPath
	return nil
}

// RegisterLogValidator registers custom log validation functions with the validator instance.
func RegisterLogValidator(validate *validator.Validate) error {
	// Register log path validation
	if err := validate.RegisterValidation("logpath", ValidateLogPath); err != nil {
		return fmt.Errorf("failed to register log path validator: %w", err)
	}

	return nil
}

// ValidateLogPath validates the log path which can be stdout, stderr, or a file path.
func ValidateLogPath(fl validator.FieldLevel) bool {
	path := fl.Field().String()

	// Allow stdout and stderr
	if path == "stdout" || path == "stderr" {
		return true
	}

	// For file paths, check if the directory exists (the file doesn't need to exist yet)
	if path != "" {
		dir := filepath.Dir(path)
		if dir != "." {
			info, err := os.Stat(dir)
			if err != nil || !info.IsDir() {
				return false
			}
		}
		return true
	}

	return false
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelWarn
	}
}
