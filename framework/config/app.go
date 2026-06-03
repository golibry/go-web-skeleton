package config

import (
	"os"

	"github.com/golibry/go-params/params"
)

const AppBaseDirEnvName = "APP_BASE_DIR"

type App struct {
	// AppBaseDir is the base directory path of the application.
	// This is used as the root directory for relative paths and must be a valid directory.
	// Defaults to the current working directory.
	AppBaseDir string `env:"APP_BASE_DIR" validate:"required,dir"`

	// AppEnv specifies the application environment.
	// Valid values are: "prod", "dev", "test".
	AppEnv string `env:"APP_ENV" default:"loc" validate:"required,oneof=prod dev test loc stg"`
}

// Populate implements the go-config Config interface for App.
// It reads values from environment variables providing sensible defaults.
func (c *App) Populate() error {
	// Check if config already set
	if c.AppBaseDir == "" {
		appDir, _ := params.GetEnvAsString(AppBaseDirEnvName, "")
		c.AppBaseDir = appDir
	} else {
		_ = os.Setenv(AppBaseDirEnvName, c.AppBaseDir)
	}

	if c.AppEnv == "" {
		appEnv, _ := params.GetEnvAsString("APP_ENV", "loc")
		c.AppEnv = appEnv
	} else {
		_ = os.Setenv("APP_ENV", c.AppEnv)
	}

	return nil
}
