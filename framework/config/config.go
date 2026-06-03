package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/go-playground/validator/v10"
	goconfig "github.com/golibry/go-config/config"
)

type AppProvider interface {
	AppRef() *App
}

func BuildCustom[T AppProvider](composite T) error {
	if any(composite) == nil {
		return fmt.Errorf("composite config cannot be nil")
	}

	app := composite.AppRef()
	if app == nil {
		return fmt.Errorf("composite must provide a non-nil App via AppRef()")
	}

	if err := app.Populate(); err != nil {
		return fmt.Errorf("failed to populate app config: %w", err)
	}

	setDefaultLocalDevAppBaseDir(app)
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := RegisterLogValidator(validate); err != nil {
		return fmt.Errorf("failed to register custom validators: %w", err)
	}

	loader := goconfig.NewLoaderWithValidator(validate)
	if err := loader.LoadInto(composite, goconfig.Options{
		Env:     app.AppEnv,
		BaseDir: app.AppBaseDir,
	}); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	return nil
}

// DebugAny returns a masked debug string for any config struct.
func DebugAny(v interface{}) string {
	sensitiveKeys := []string{
		"dsn", "password", "pass", "secret", "token", "apikey",
		"api_key", "key", "auth", "authorization",
	}
	return goconfig.Debug(v, sensitiveKeys)
}

func setDefaultLocalDevAppBaseDir(app *App) {
	if app.AppEnv == "loc" && app.AppBaseDir == "" {
		_, filePath, _, _ := runtime.Caller(2)
		start := filepath.Dir(filePath)
		app.AppBaseDir = walkUpDirForAny(start, ".env", ".env.local")
	}
}

func walkUpDirForAny(start string, names ...string) string {
	dir := start
	for {
		for _, n := range names {
			p := filepath.Join(dir, n)
			if pathExists(p) { // accept both file or dir markers (file or directory)
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir { // reached filesystem root
			return ""
		}
		dir = parent
	}
}

// pathExists returns true if the path exists (file or directory).
func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
