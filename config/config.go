package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/golibry/go-params/params"
	"github.com/joho/godotenv"
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

// BuildConfig builds the config struct with comprehensive validation and error handling.
func BuildConfig() (Config, error) {
	defaultAppDir, _ := determineAppBaseDir()
	appDir, _ := params.GetEnvAsString("APP_BASE_DIR", defaultAppDir)
	appEnv, _ := params.GetEnvAsString("APP_ENV", "dev")
	logLevel, _ := params.GetEnvAsString("APP_LOG_LEVEL", "warn")
	logPath, _ := params.GetEnvAsString("APP_LOG_PATH", "stdout")

	// Load environment variables first
	if err := loadEnvVars(appEnv, appDir); err != nil {
		return Config{}, fmt.Errorf("failed to load environment variables: %w", err)
	}

	// Build configuration struct
	config := Config{
		AppBaseDir: appDir,
		AppEnv:     appEnv,
		LogLevel:   parseLogLevel(logLevel),
		LogPath:    logPath,
		HttpServer: newHttpServerConfig(),
		Db:         newDatabaseConfig(),
	}

	// Validate configuration with custom validators
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := registerCustomValidators(validate); err != nil {
		return Config{}, fmt.Errorf("failed to register custom validators: %w", err)
	}

	if err := validate.Struct(config); err != nil {
		return Config{}, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// String returns a safe string representation of the configuration with sensitive data masked.
// This method is useful for logging configuration without exposing sensitive information.
func (c Config) String() string {
	return fmt.Sprintf(
		"Config{AppBaseDir: %s, AppEnv: %s, LogLevel: %s, LogPath: %s, HttpServer: %s, Db: %s}",
		c.AppBaseDir,
		c.AppEnv,
		c.LogLevel.String(),
		c.LogPath,
		c.HttpServer.String(),
		c.Db.String(),
	)
}

// maskSensitiveData masks sensitive information in a string for safe logging.
func maskSensitiveData(data string) string {
	if data == "" {
		return ""
	}

	// For DSN strings, mask everything after the first character and before the last character
	if strings.Contains(data, "@") && strings.Contains(data, ":") {
		parts := strings.Split(data, "@")
		if len(parts) >= 2 {
			// Mask the user:password part
			userPass := parts[0]
			if len(userPass) > 2 {
				masked := string(userPass[0]) + strings.Repeat(
					"*",
					len(userPass)-2,
				) + string(userPass[len(userPass)-1])
				return masked + "@" + parts[1]
			}
		}
	}

	// For other sensitive data, show the first and last character with asterisks in between
	if len(data) <= 2 {
		return strings.Repeat("*", len(data))
	}

	return string(data[0]) + strings.Repeat("*", len(data)-2) + string(data[len(data)-1])
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

// loadEnvVars Loads the entries from env files and sets them as env variables for this process.
// Loads each file in order: .env.{dev|prod|test}.local, .env.local, .env.{dev|prod|test}, .env.
// The first loaded file has priority. Files will not overwrite the values of the already loaded
// env vars (already loaded from env files or via other means).
func loadEnvVars(env string, appBaseDir string) error {
	localEnvFileName := filepath.Join(appBaseDir, ".env."+env+".local")
	if _, err := os.Stat(localEnvFileName); err == nil {
		err := godotenv.Load(localEnvFileName)

		if err != nil {
			return formatEnvLoadErr(localEnvFileName, err)
		}
	}

	genericLocalFileName := filepath.Join(appBaseDir, ".env.local")

	if env != "test" {
		if _, err := os.Stat(genericLocalFileName); err == nil {
			err := godotenv.Load(genericLocalFileName)

			if err != nil {
				return formatEnvLoadErr(genericLocalFileName, err)
			}
		}
	}

	genericEnvFileName := filepath.Join(appBaseDir, ".env."+env)
	if _, err := os.Stat(genericEnvFileName); err == nil {
		err := godotenv.Load(genericEnvFileName)

		if err != nil {
			return formatEnvLoadErr(genericEnvFileName, err)
		}
	}

	baseEnvFileName := filepath.Join(appBaseDir, ".env")
	if _, err := os.Stat(baseEnvFileName); err == nil {
		err := godotenv.Load(baseEnvFileName)

		if err != nil {
			return formatEnvLoadErr(baseEnvFileName, err)
		}
	}

	return nil
}

func formatEnvLoadErr(fileName string, err error) error {
	return fmt.Errorf(
		"error occurred while trying to load env file: %s. Error message: %s",
		fileName,
		err.Error(),
	)
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

func determineAppBaseDir() (string, error) {
	// Option 1: Use the current working directory
	if wd, err := os.Getwd(); err == nil {
		return wd, nil
	}

	// Option 2: Use executable directory as fallback
	if ex, err := os.Executable(); err == nil {
		return filepath.Dir(ex), nil
	}

	return "", fmt.Errorf("unable to determine application base directory")
}
