//go:build sqlite && !mysql && !postgres

package migrations

import (
	"database/sql"

	"github.com/golibry/go-migrations/execution"
	"github.com/golibry/go-migrations/execution/repository"
)

func newRepository(options Options, db *sql.DB) (execution.Repository, error) {
	if canonicalDriverName(options.Driver) != DriverSQLite {
		return nil, repositoryBuildMismatchError(options.Driver, DriverSQLite)
	}

	return repository.NewSqliteHandler(options.DSN, options.ExecutionsTable, options.Context, db)
}
