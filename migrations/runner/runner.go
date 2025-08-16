package runner

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golibry/go-migrations/cli"
	"github.com/golibry/go-migrations/execution/repository"
	"github.com/golibry/go-migrations/migration"
	"github.com/golibry/go-web-skeleton/config"
	"github.com/golibry/go-web-skeleton/infrastructure/registry"
	"github.com/golibry/go-web-skeleton/migrations/versions"
)

type MigrationsContainer struct {
	*registry.LoggerService
	*registry.ConfigService
	*registry.DbService
}

func RunMigrations(args []string, exitFunc func(int), outputWriter io.Writer) {
	baseErr := "Could not bootstrap migrations."

	configService, err := registry.NewConfigService()
	if err != nil {
		panic(fmt.Sprintf("%s Error building config service: %s", baseErr, err))
	}

	loggerService, err := registry.NewLoggerService(configService)
	if err != nil {
		panic(fmt.Sprintf("%s Error building logger service: %s", baseErr, err))
	}

	db, err := createDatabaseConnection(configService.Config().Db)
	if err != nil {
		panic(fmt.Sprintf("%s Error building db service: %s", baseErr, err))
	}

	dbService := registry.NewDbServiceFromDb(db)

	container := &MigrationsContainer{
		LoggerService: loggerService,
		ConfigService: configService,
		DbService:     dbService,
	}

	defer func() {
		if recoverErr := recover(); recoverErr != nil {
			switch v := recoverErr.(type) {
			case string:
				recoverErr = errors.New(v)
			case error:
				recoverErr = v
			default:
				recoverErr = errors.New(fmt.Sprint(v))
			}
			cmdErr := recoverErr.(error)
			container.Logger().Error(cmdErr.Error())
		}

		_ = loggerService.Close()
		_ = dbService.Close()
	}()

	ctx := context.Background()
	dirPath := createMigrationsDirPath(container.ConfigService)
	cli.Bootstrap(
		args,
		buildRegistry(dirPath, ctx, container),
		createMysqlRepository(container, ctx),
		dirPath,
		nil,
		outputWriter,
		exitFunc,
		&cli.BootstrapSettings{
			RunMigrationsExclusively: true,
			RunLockFilesDirPath:      os.TempDir(),
			MigrationsCmdLockName:    "my-app-migrations-lock",
		},
	)
}

func createMigrationsDirPath(container *registry.ConfigService) migration.MigrationsDirPath {
	dirPath, err := migration.NewMigrationsDirPath(
		filepath.Join(container.Config().AppBaseDir, "migrations", "versions"),
	)

	if err != nil {
		panic(fmt.Errorf("invalid migrations path: %w", err))
	}

	return dirPath
}

func createMysqlRepository(
	container *MigrationsContainer,
	ctx context.Context,
) *repository.MysqlHandler {
	repo, err := repository.NewMysqlHandler(
		"", "migration_executions", ctx, container.Db(),
	)

	if err != nil {
		panic(fmt.Errorf("failed to build migration executions repository: %w", err))
	}

	return repo
}

// buildRegistry This will create a new registry and register all migrations
func buildRegistry(
	dirPath migration.MigrationsDirPath,
	ctx context.Context,
	container *MigrationsContainer,
) *migration.DirMigrationsRegistry {
	// Manually instantiate all migrations with their dependencies
	allMigrations := []migration.Migration{
		&versions.Migration1755033922{
			Db:  container.Db(),
			Ctx: ctx,
		},
	}

	// Create a registry with manually instantiated migrations
	return migration.NewDirMigrationsRegistry(dirPath, allMigrations)
}

// createDatabaseConnection creates and configures a database connection
func createDatabaseConnection(config config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("mysql", config.Dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool settings
	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(1)
	db.SetConnMaxIdleTime(time.Minute * 10)
	db.SetConnMaxLifetime(time.Minute * 10)

	return db, nil
}
