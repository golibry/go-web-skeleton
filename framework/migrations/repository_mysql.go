//go:build mysql

package migrations

import (
	"database/sql"

	"github.com/golibry/go-migrations/execution"
	"github.com/golibry/go-migrations/execution/repository"
)

func newRepository(options Options, db *sql.DB) (execution.Repository, error) {
	if canonicalDriverName(options.Driver) != DriverMySQL {
		return nil, unsupportedDriverError(options.Driver)
	}

	return repository.NewMysqlHandler(options.DSN, options.ExecutionsTable, options.Context, db)
}
