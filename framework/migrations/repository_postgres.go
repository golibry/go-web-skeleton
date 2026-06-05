//go:build postgres && !mysql && !sqlite

package migrations

import (
	"database/sql"

	"github.com/golibry/go-migrations/execution"
	"github.com/golibry/go-migrations/execution/repository"
)

func newRepository(options Options, db *sql.DB) (execution.Repository, error) {
	if canonicalDriverName(options.Driver) != DriverPostgres {
		return nil, repositoryBuildMismatchError(options.Driver, DriverPostgres)
	}

	return repository.NewPostgresHandler(options.DSN, options.ExecutionsTable, options.Context, db)
}
