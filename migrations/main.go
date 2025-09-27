package main

import (
	"context"
	"database/sql"
	"os"

	"github.com/golibry/go-migrations/migration"
	"github.com/golibry/go-web-skeleton/migrations/runner"
	"github.com/golibry/go-web-skeleton/migrations/versions"
)

func main() {
	runner.RunMigrations(os.Args[1:], runner.MigrationsRunnerOptions{
		ExitFunc:     os.Exit,
		OutputWriter: os.Stdout,
		DbDsn:        os.Getenv("DB_DSN"),
		MigrationsFactory: func(db *sql.DB, ctx context.Context) []migration.Migration {
			return []migration.Migration{
				&versions.Migration1755033922{Db: db, Ctx: ctx},
			}
		},
	})
}
