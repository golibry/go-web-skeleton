package config

import (
	"time"

	"github.com/golibry/go-params/params"
)

// DatabaseConfig contains database connection and pool configuration settings.
// It defines how the application connects to and manages database connections.
type DatabaseConfig struct {
	// Dsn is the Data Source Name for the database connection.
	// It is a required field and is validated to ensure it is not empty.
	// The Dsn should contain all the necessary information to connect to the database,
	// including the username, password, host, port, and database name.
	// Example: "user:password@tcp(localhost:3306)/dbname"
	Dsn string `validate:"required"`
	
	// MaxIdleConnections sets the maximum number of connections in the idle connection pool.
	// Must be between 0 and 99. A value of 0 means no idle connections are retained.
	MaxIdleConnections int `validate:"number,gte=0,lte=99"`
	
	// MaxOpenConnections sets the maximum number of open connections to the database.
	// Must be between 0 and 99. A value of 0 means unlimited connections.
	MaxOpenConnections int `validate:"number,gte=0,lte=99"`
	
	// ConnectionMaxIdleTime sets the maximum amount of time a connection may be idle.
	// Expired connections may be closed lazily before reuse.
	ConnectionMaxIdleTime time.Duration
	
	// ConnectionMaxLifetime sets the maximum amount of time a connection may be reused.
	// Expired connections may be closed lazily before reuse.
	ConnectionMaxLifetime time.Duration
	
	// MigrationsDirPath specifies the directory path containing database migration files.
	// This must be a valid directory path and is required for database migrations.
	MigrationsDirPath string `validate:"required,dir"`
}

func newDatabaseConfig() DatabaseConfig {
	dsn, _ := params.GetEnvAsString("DB_DSN", "")
	maxIdleConnections, _ := params.GetEnvAsInt("DB_MAX_IDLE_CONNECTIONS", 2)
	maxOpenConnections, _ := params.GetEnvAsInt("DB_MAX_OPEN_CONNECTIONS", 10)
	connectionMaxIdleTime, _ := params.GetEnvAsDuration(
		"DB_CONNECTION_MAX_IDLE_TIME",
		time.Minute*3,
	)
	connectionMaxLifetime, _ := params.GetEnvAsDuration(
		"DB_CONNECTION_MAX_LIFETIME",
		time.Minute*3,
	)
	migrationsDirPath, _ := params.GetEnvAsString("DB_MIGRATIONS_DIR_PATH", "")

	return DatabaseConfig{
		Dsn:                   dsn,
		MaxIdleConnections:    maxIdleConnections,
		MaxOpenConnections:    maxOpenConnections,
		ConnectionMaxIdleTime: connectionMaxIdleTime,
		ConnectionMaxLifetime: connectionMaxLifetime,
		MigrationsDirPath:     migrationsDirPath,
	}
}
