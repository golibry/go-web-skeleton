package config

import (
	"fmt"
	"log/slog"
	"os"
	"path"

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
	AppEnv string `validate:"oneof=prod dev test"`
	
	// LogLevel defines the minimum log level for the application logger.
	// Uses slog.Level which supports debug, info, warn, and error levels.
	LogLevel slog.Level
	
	// LogPath specifies where log output should be directed.
	// Valid values are: "stdout", "stderr", or a file path.
	LogPath string `validate:"required,oneof=stdout stderr|file"`
	
	// HttpServer contains HTTP server configuration settings.
	HttpServer HttpServerConfig `validate:"required"`
	
	// Db contains database connection and configuration settings.
	Db DatabaseConfig `validate:"required"`
}

// BuildConfig Builds the config struct
func BuildConfig() (Config, error) {
	appDir, _ := params.GetEnvAsString("APP_BASE_DIR", "")
	appEnv, _ := params.GetEnvAsString("APP_ENV", "")
	legLevel, _ := params.GetEnvAsString("APP_LOG_LEVEL", "warn")
	logPath, _ := params.GetEnvAsString("APP_LOG_PATH", "stdout")
	err := loadEnvVars(appEnv, appDir)
	if err != nil {
		return Config{}, err
	}

	config := Config{
		AppBaseDir: appDir,
		AppEnv:     appEnv,
		LogLevel:   parseLogLevel(legLevel),
		LogPath:    logPath,
		HttpServer: newHttpServerConfig(),
		Db:         newDatabaseConfig(),
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	return config, validate.Struct(config)
}

// loadEnvVars Loads the entries from env files and sets them as env variables for this process.
// Loads each file in order: .env.{dev|prod|test}.local, .env.local, .env.{dev|prod|test}, .env.
// The first loaded file has priority. Files will not overwrite the values of the already loaded
// env vars (already loaded from env files or via other means).
func loadEnvVars(env string, appBaseDir string) error {
	localEnvFileName := path.Join(appBaseDir, ".env."+env+".local")
	if _, err := os.Stat(localEnvFileName); err == nil {
		err := godotenv.Load(localEnvFileName)

		if err != nil {
			return formatEnvLoadErr(localEnvFileName, err)
		}
	}

	genericLocalFileName := path.Join(appBaseDir, ".env.local")
	if env != "test" {
		if _, err := os.Stat(genericLocalFileName); err == nil {
			err := godotenv.Load(genericLocalFileName)

			if err != nil {
				return formatEnvLoadErr(genericLocalFileName, err)
			}
		}
	}

	genericEnvFileName := path.Join(appBaseDir, ".env."+env)
	if _, err := os.Stat(genericEnvFileName); err == nil {
		err := godotenv.Load(genericEnvFileName)

		if err != nil {
			return formatEnvLoadErr(genericEnvFileName, err)
		}
	}

	baseEnvFileName := path.Join(appBaseDir, ".env")
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
