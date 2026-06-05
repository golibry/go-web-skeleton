//go:build mysql && !postgres && !sqlite

package migrations

import (
	"database/sql"

	"github.com/golibry/go-migrations/execution"
	"github.com/golibry/go-migrations/execution/repository"
)

func newRepository(options Options, db *sql.DB) (execution.Repository, error) {
	if canonicalDriverName(options.Driver) != DriverMySQL {
		return nil, repositoryBuildMismatchError(options.Driver, DriverMySQL)
	}

	return repository.NewMysqlHandler(options.DSN, options.ExecutionsTable, options.Context, db)
}
