package main

import (
	"context"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golibry/go-migrations/cli"
	"github.com/golibry/go-migrations/execution/repository"
	"github.com/golibry/go-migrations/migration"
	"github.com/golibry/go-web-skeleton/infrastructure/registry"
	"github.com/golibry/go-web-skeleton/migrations/versions"
	"os"
	"path/filepath"
	"reflect"
)

func main() {
	baseErr := "Could not bootstrap migrations."
	container, err := registry.NewContainer()
	if err != nil {
		panic(fmt.Sprintf("%s Error building container registry: %s", baseErr, err))
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
	}()

	ctx := context.Background()
	dirPath := createMigrationsDirPath(container)
	cli.Bootstrap(
		os.Args[1:],
		buildRegistry(dirPath, ctx, container),
		createMysqlRepository(container, ctx),
		dirPath,
		nil,
		os.Stdout,
		os.Exit,
		&cli.BootstrapSettings{
			RunMigrationsExclusively: true,
			RunLockFilesDirPath:      os.TempDir(),
			MigrationsCmdLockName:    "my-app-migrations-lock",
		},
	)
}

func createMigrationsDirPath(container *registry.Container) migration.MigrationsDirPath {
	dirPath, err := migration.NewMigrationsDirPath(
		filepath.Join(container.Config().AppBaseDir, "migrations", "versions"),
	)

	if err != nil {
		panic(fmt.Errorf("invalid migrations path: %w", err))
	}

	return dirPath
}

func createMysqlRepository(
	container *registry.Container,
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
	container *registry.Container,
) *migration.DirMigrationsRegistry {
	return migration.NewAutoDiscoveryDirMigrationsRegistry(
		dirPath,
		func(migrationType reflect.Type) []reflect.Value {
			// Provide dependencies based on what each migration type needs
			dependencies := []reflect.Value{reflect.ValueOf(container.Db())}

			// Some migrations might need additional dependencies like context
			// Check if the migration has a Ctx field
			for i := 0; i < migrationType.NumField(); i++ {
				field := migrationType.Field(i)
				if field.Name == "Ctx" && field.Type.String() == "context.Context" {
					dependencies = append(dependencies, reflect.ValueOf(ctx))
					break
				}
			}

			return dependencies
		},
		&versions.Dummy{}, // Example to tell the system which package to scan
	)
}
