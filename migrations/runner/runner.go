package runner

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golibry/go-migrations/cli"
	"github.com/golibry/go-migrations/execution/repository"
	"github.com/golibry/go-migrations/migration"
)

type MigrationsRunnerOptions struct {
	Logger           *slog.Logger
	OutputWriter     io.Writer
	ExitFunc         func(int)
	DbDsn            string
	MigrationsDir    string
	ExecutionsTable  string
	LockName         string
	LockFilesDirPath string
	// MigrationsFactory must return the list of migrations, instantiated with the given db and ctx.
	MigrationsFactory func(db *sql.DB, ctx context.Context) []migration.Migration
}

// RunMigrations boots the migrations CLI using only stdlib and go-migrations deps.
// It does not depend on any packages from this project.
func RunMigrations(args []string, options MigrationsRunnerOptions) {
	if options.DbDsn == "" {
		panic("Could not bootstrap migrations. DbDsn is required")
	}

	if options.Logger == nil {
		panic("Could not bootstrap migrations. Logger is required")
	}

	if options.MigrationsDir == "" {
		panic("Could not bootstrap migrations. MigrationsDir is required")
	}

	// Defaults
	if options.OutputWriter == nil {
		options.OutputWriter = os.Stdout
	}
	if options.ExitFunc == nil {
		options.ExitFunc = os.Exit
	}
	if options.ExecutionsTable == "" {
		options.ExecutionsTable = "migration_executions"
	}
	if options.LockName == "" {
		options.LockName = "app-migrations-lock"
	}
	if options.LockFilesDirPath == "" {
		options.LockFilesDirPath = os.TempDir()
	}

	db, err := createDatabaseConnection(options.DbDsn)
	if err != nil {
		panic(
			fmt.Errorf(
				"could not bootstrap migrations. Error building db connection: %w",
				err,
			),
		)
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	var dirPath migration.MigrationsDirPath
	dirPath, err = migration.NewMigrationsDirPath(options.MigrationsDir)
	if err != nil {
		panic(fmt.Errorf("invalid migrations directory path: %w", err))
	}

	repo, err := repository.NewMysqlHandler("", options.ExecutionsTable, ctx, db)
	if err != nil {
		panic(fmt.Errorf("failed to build migration executions repository: %w", err))
	}

	var migrations []migration.Migration
	if options.MigrationsFactory != nil {
		migrations = options.MigrationsFactory(db, ctx)
	}
	reg := migration.NewDirMigrationsRegistry(dirPath, migrations)

	cli.Bootstrap(
		args,
		reg,
		repo,
		dirPath,
		nil,
		options.OutputWriter,
		options.ExitFunc,
		&cli.BootstrapSettings{
			RunMigrationsExclusively: true,
			RunLockFilesDirPath:      options.LockFilesDirPath,
			MigrationsCmdLockName:    options.LockName,
		},
	)
}

// createDatabaseConnection creates and configures a database connection
func createDatabaseConnection(dbDsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dbDsn)
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
