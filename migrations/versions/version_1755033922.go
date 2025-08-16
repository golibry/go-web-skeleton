package versions

import (
	"context"
	"database/sql"
)

type Migration1755033922 struct {
	Db  *sql.DB
	Ctx context.Context
}

func (migration *Migration1755033922) Version() uint64 {
	return 1755033922 // Do not edit this! If you do, migrations may run out of order
}

func (migration *Migration1755033922) Up() error {
	query := `
		CREATE TABLE dummy (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL
		)
	`

	_, err := migration.Db.ExecContext(migration.Ctx, query)
	return err
}

func (migration *Migration1755033922) Down() error {
	query := `DROP TABLE IF EXISTS dummy`

	_, err := migration.Db.ExecContext(migration.Ctx, query)
	return err
}
