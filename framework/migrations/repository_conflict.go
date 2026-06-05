//go:build (mysql && postgres) || (mysql && sqlite) || (postgres && sqlite)

package migrations

import (
	"database/sql"

	"github.com/golibry/go-migrations/execution"
)

func newRepository(options Options, _ *sql.DB) (execution.Repository, error) {
	return nil, conflictingRepositoryBuildTagsError(options.Driver)
}
