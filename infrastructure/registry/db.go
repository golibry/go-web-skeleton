package registry

import (
	_ "github.com/go-sql-driver/mysql"

	"database/sql"

	"github.com/golibry/go-web-skeleton/config"
)

func newMysqlDbConnectionPool(config config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("mysql", config.Dsn)
	if db == nil {
		return nil, err
	}

	db.SetMaxIdleConns(config.MaxIdleConnections)
	db.SetMaxOpenConns(config.MaxOpenConnections)
	db.SetConnMaxIdleTime(config.ConnectionMaxIdleTime)
	db.SetConnMaxLifetime(config.ConnectionMaxLifetime)
	return db, err
}
