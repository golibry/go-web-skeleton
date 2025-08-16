package test

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/golibry/go-web-skeleton/infrastructure/registry"
	"github.com/golibry/go-web-skeleton/migrations/runner"
)

var (
	globalBootstrap     *Bootstrap
	globalBootstrapOnce sync.Once
	globalBootstrapErr  error
)

// Bootstrap provides common test setup functionality for the entire application.
// It ensures the test environment is properly configured before running tests.
type Bootstrap struct {
	ConfigService *registry.ConfigService
	DbService     *registry.DbService
}

// SetupTestEnvironment initializes the test environment with the proper configuration.
// It should be called in TestMain or SetupSuite methods to ensure a consistent test setup.
// This function sets the APP_ENV to "test" if not already set and initializes core services.
func SetupTestEnvironment() (*Bootstrap, error) {
	baseErr := "Could not bootstrap tests."

	// Ensure we're running in the test environment
	if os.Getenv("APP_ENV") == "" {
		err := os.Setenv("APP_ENV", "test")
		if err != nil {
			panic(fmt.Sprintf("%s Error setting APP_ENV: %s", baseErr, err))
		}
	}

	configService, err := registry.NewConfigService()
	if err != nil {
		panic(fmt.Sprintf("%s Error building config service: %s", baseErr, err))
	}

	dbService, err := registry.NewDbService(configService)
	if err != nil {
		panic(fmt.Sprintf("%s Failed to open database connection: %s", baseErr, err))
	}

	testBootstrap := &Bootstrap{
		ConfigService: configService,
		DbService:     dbService,
	}

	err = testBootstrap.RemoveTestDb()
	if err != nil {
		panic(fmt.Sprintf("%s Error cleaning up test db: %s", baseErr, err))
	}

	err = testBootstrap.SetupTestDb()
	if err != nil {
		panic(fmt.Sprintf("%s Error setting up test db: %s", baseErr, err))
	}

	return testBootstrap, nil
}

func getRootDb(configService *registry.ConfigService) (*sql.DB, error) {
	return sql.Open(
		"mysql",
		strings.Replace(
			configService.Config().Db.Dsn,
			configService.Config().Db.DbName,
			"",
			1,
		),
	)
}

// TeardownTestEnvironment cleans up resources after tests complete.
// It should be called in TestMain or TearDownSuite methods.
func (tb *Bootstrap) TeardownTestEnvironment() {
	_ = tb.RemoveTestDb()
	if tb.DbService != nil {
		_ = tb.DbService.Close()
	}
}

// RemoveTestDb removes test data from all tables.
// This is useful for cleaning up between test runs to ensure test isolation.
func (tb *Bootstrap) RemoveTestDb() error {
	db, err := getRootDb(tb.ConfigService)
	defer func(db *sql.DB) {
		if db != nil {
			_ = db.Close()
		}
	}(db)

	if err != nil {
		return err
	}

	_, err = db.Exec(
		fmt.Sprintf("DROP DATABASE IF EXISTS %s", tb.ConfigService.Config().Db.DbName),
	)

	return err
}

func (tb *Bootstrap) SetupTestDb() error {
	db, err := getRootDb(tb.ConfigService)
	defer func(db *sql.DB) {
		if db != nil {
			_ = db.Close()
		}
	}(db)

	if err != nil {
		return err
	}

	_, err = db.Exec(
		fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", tb.ConfigService.Config().Db.DbName),
	)

	migrationsOutput := &bytes.Buffer{}
	migrationsOk := false
	runner.RunMigrations(
		[]string{"up"},
		func(code int) {
			if code == 0 {
				migrationsOk = true
			}
		},
		migrationsOutput,
	)

	if !migrationsOk {
		return fmt.Errorf("failed to run migrations: %s", migrationsOutput.String())
	}

	return err
}

// initializeGlobalBootstrap sets up the global test bootstrap if we're running tests
func initializeGlobalBootstrap() {
	globalBootstrapOnce.Do(
		func() {
			globalBootstrap, globalBootstrapErr = SetupTestEnvironment()
		},
	)
}

// GetGlobalTestBootstrap returns the global test bootstrap instance.
// It automatically initializes the bootstrap if not already done.
// This provides easy access to test infrastructure for any test package.
func GetGlobalTestBootstrap() (*Bootstrap, error) {
	initializeGlobalBootstrap()
	return globalBootstrap, globalBootstrapErr
}

// MainBootstrap provides a TestMain function implementation that handles
// environment setup and teardown for the entire test package.
// This now uses the global bootstrap for consistency.
// Usage in your test package:
//
//	func TestMain(m *testing.M) {
//	    test.MainBootstrap(m)
//	}
// func MainBootstrap(m *testing.M) {
// 	bootstrap, err := GetGlobalTestBootstrap()
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	// Run tests
// 	code := m.Run()
//
// 	// Cleanup only if we have a bootstrap instance
// 	if bootstrap != nil {
// 		bootstrap.TeardownTestEnvironment()
// 	}
//
// 	os.Exit(code)
// }
