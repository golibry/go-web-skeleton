package registry

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golibry/go-web-skeleton/config"
)

type DbService struct {
	db *sql.DB
}

func NewDbService(cfg *ConfigService) (*DbService, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config service cannot be nil")
	}

	dbConfig := cfg.Config().Db
	db, err := createDatabaseConnection(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	service := &DbService{
		db: db,
	}

	// Test the connection with context and timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := service.Ping(ctx); err != nil {
		_ = service.Close() // Clean up on failure
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return service, nil
}

func NewDbServiceFromDb(db *sql.DB) *DbService {
	return &DbService{
		db: db,
	}
}

func (d *DbService) Db() *sql.DB {
	return d.db
}

func (d *DbService) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

func (d *DbService) Ping(ctx context.Context) error {
	if d.db == nil {
		return fmt.Errorf("database connection is nil")
	}
	return d.db.PingContext(ctx)
}

// createDatabaseConnection creates and configures a database connection
func createDatabaseConnection(config config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("mysql", config.Dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool settings
	db.SetMaxIdleConns(config.MaxIdleConnections)
	db.SetMaxOpenConns(config.MaxOpenConnections)
	db.SetConnMaxIdleTime(config.ConnectionMaxIdleTime)
	db.SetConnMaxLifetime(config.ConnectionMaxLifetime)

	return db, nil
}
