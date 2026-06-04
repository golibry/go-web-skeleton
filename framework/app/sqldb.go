package app

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/golibry/go-web-skeleton/framework/config"
)

type SQLDBOptions struct {
	PingOnStartup bool
}

type SQLDBService struct {
	db *sql.DB
}

type SqlDbService = SQLDBService

func NewDBService(dbConfig config.Database, options ...SQLDBOptions) (*SQLDBService, error) {
	db, err := createDbConnectionPool(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create sql db service: %w", err)
	}

	dbOptions := SQLDBOptions{}
	if len(options) > 0 {
		dbOptions = options[0]
	}
	if dbOptions.PingOnStartup {
		if err := db.PingContext(context.Background()); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("failed to ping sql db: %w", err)
		}
	}

	service := &SQLDBService{
		db: db,
	}

	return service, nil
}

func NewDbService(dbConfig config.Database) (*SQLDBService, error) {
	return NewDBService(dbConfig)
}

func (d *SQLDBService) DB() *sql.DB {
	return d.db
}

func (d *SQLDBService) Db() *sql.DB {
	return d.DB()
}

func (d *SQLDBService) Close() error {
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
