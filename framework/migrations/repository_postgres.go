//go:build postgres

package migrations

import (
	"database/sql"

	"github.com/golibry/go-migrations/execution"
	"github.com/golibry/go-migrations/execution/repository"
)

func newRepository(options Options, db *sql.DB) (execution.Repository, error) {
	if canonicalDriverName(options.Driver) != DriverPostgres {
		return nil, unsupportedDriverError(options.Driver)
	}

	return repository.NewPostgresHandler(options.DSN, options.ExecutionsTable, options.Context, db)
}
