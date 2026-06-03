package app

import (
	"database/sql"
	"fmt"

	"github.com/golibry/go-web-skeleton/framework/config"
)

type SqlDbService struct {
	db *sql.DB
}

func NewDbService(dbConfig config.Database) (*SqlDbService, error) {
	db, err := createDbConnectionPool(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create sql db service: %w", err)
	}

	service := &SqlDbService{
		db: db,
	}

	return service, nil
}

func (d *SqlDbService) Db() *sql.DB {
	return d.db
}

func (d *SqlDbService) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// createDbConnectionPool creates and configures a database connection
func createDbConnectionPool(config config.Database) (*sql.DB, error) {
	db, err := sql.Open(config.Driver, config.Dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to sql db connection pool: %w", err)
	}

	// Configure connection pool settings
	db.SetMaxIdleConns(config.MaxIdleConnections)
	db.SetMaxOpenConns(config.MaxOpenConnections)
	db.SetConnMaxIdleTime(config.ConnectionMaxIdleTime)
	db.SetConnMaxLifetime(config.ConnectionMaxLifetime)

	return db, nil
}
