package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/go-playground/validator/v10"
	goconfig "github.com/golibry/go-config/config"
	"github.com/golibry/go-params/params"
)

// Config includes all app configuration.
// The main Config struct can have other structs that handle a specific set of config attributes.
type Config struct {
	// AppBaseDir is the base directory path of the application.
	// This is used as the root directory for relative paths and must be a valid directory.
	AppBaseDir string `validate:"required,dir"`

	// AppEnv specifies the application environment.
	// Valid values are: "prod", "dev", "test".
	AppEnv string `validate:"required,oneof=prod dev test"`

	// LogLevel defines the minimum log level for the application logger.
	// Uses slog.Level which supports debug, info, warn, and error levels.
	LogLevel slog.Level

	// LogPath specifies where log output should be directed.
	// Valid values are: "stdout", "stderr", or a file path.
	LogPath string `validate:"required,logpath"`

	// HttpServer contains HTTP server configuration settings.
	HttpServer HttpServerConfig `validate:"required"`

	// Db contains database connection and configuration settings.
	Db DatabaseConfig `validate:"required"`
}

// Populate implements the go-config Config interface for the top-level Config.
// It reads values from environment variables and sets defaults.
func (c *Config) Populate() error {
	appDir, _ := params.GetEnvAsString("APP_BASE_DIR", determineAppBaseDir())
	appEnv, _ := params.GetEnvAsString("APP_ENV", "dev")
	logLevel, _ := params.GetEnvAsString("APP_LOG_LEVEL", "warn")
	logPath, _ := params.GetEnvAsString("APP_LOG_PATH", "stdout")

	c.AppBaseDir = appDir
	c.AppEnv = appEnv
	c.LogLevel = parseLogLevel(logLevel)
	c.LogPath = logPath
	// Nested configs are populated by CompositeConfig
	return nil
}

// BuildConfig builds the config struct using golibry/go-config and validates it.
func BuildConfig() (Config, error) {
	appDir, _ := params.GetEnvAsString("APP_BASE_DIR", determineAppBaseDir())
	appEnv, _ := params.GetEnvAsString("APP_ENV", "dev")

	// Load environment variables using the shared go-config loader (respects priority order)
	if err := goconfig.LoadEnvVars(appEnv, appDir); err != nil {
		return Config{}, fmt.Errorf("failed to load environment variables: %w", err)
	}

	cfg := Config{}
	// Populate top-level config fields
	if err := cfg.Populate(); err != nil {
		return Config{}, fmt.Errorf("failed to populate top-level config: %w", err)
	}

	// Prepare validator and register custom rules
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := registerCustomValidators(validate); err != nil {
		return Config{}, fmt.Errorf("failed to register custom validators: %w", err)
	}

	// Build composite config and populate nested configs + validate all
	composite := goconfig.NewCompositeConfig(validate)
	if err := composite.PopulateAndValidate(&cfg, appEnv, appDir); err != nil {
		return Config{}, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// Debug returns a safe, human-readable string representation of the configuration.
// It leverages the go-config Debug() helper to recursively print the config while masking
// sensitive values (e.g., DSNs, passwords, tokens, secrets).
func (c Config) Debug() string {
	// Keywords are checked against field names (case-insensitive) to decide masking.
	sensitiveKeys := []string{
		"dsn", "password", "pass", "secret", "token", "apikey",
		"api_key", "key", "auth", "authorization",
	}
	return goconfig.Debug(c, sensitiveKeys)
}

// registerCustomValidators registers custom validation functions with the validator instance.
func registerCustomValidators(validate *validator.Validate) error {
	// Register log path validation
	if err := validate.RegisterValidation("logpath", validateLogPath); err != nil {
		return fmt.Errorf("failed to register log path validator: %w", err)
	}

	return nil
}

// validateLogPath validates the log path which can be stdout, stderr, or a file path.
func validateLogPath(fl validator.FieldLevel) bool {
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

func determineAppBaseDir() string {
	_, filePath, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(filePath))
}
