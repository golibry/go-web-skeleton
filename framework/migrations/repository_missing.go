//go:build !mysql && !postgres && !sqlite

package migrations

import (
	"database/sql"

	"github.com/golibry/go-migrations/execution"
)

func newRepository(options Options, _ *sql.DB) (execution.Repository, error) {
	return nil, missingRepositoryBuildTagError(options.Driver)
}
