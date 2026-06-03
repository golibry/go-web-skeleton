package config

import (
	"path/filepath"
	"time"

	"github.com/golibry/go-params/params"
)

type Migrations struct {
	// MigrationsDirPath specifies the directory path containing database migration files.
	// This must be a valid directory path and is required for database migrations.
	MigrationsDirPath string `validate:"required,dir"`

	// ExecutionsTable specifies the name of the table that stores migration execution information.
	ExecutionsTable string `validate:"required"`
}

// Populate implements the go-config Config interface for Database.
// It reads values from environment variables providing sensible defaults.
func (d *Migrations) Populate() error {
	appBaseDir, _ := params.GetEnvAsString(AppBaseDirEnvName, "")
	defaultMigrationsDirPath := filepath.Join(appBaseDir, "migrations")
	migrationsDirPath, _ := params.GetEnvAsString(
		"DB_MIGRATIONS_DIR_PATH",
		defaultMigrationsDirPath,
	)
	executionsTable, _ := params.GetEnvAsString(
		"DB_MIGRATIONS_TABLE", "migrations_executions",
	)

	d.MigrationsDirPath = migrationsDirPath
	d.ExecutionsTable = executionsTable
	return nil
}

// Database contains database connection and pool configuration settings.
// It defines how the application connects to and manages database connections.
type Database struct {
	// Name is the configured application database name.
	Name string `validate:"required"`

	// Dsn is the Data Source Name for the database connection.
	// It is a required field and is validated to ensure it is not empty.
	// The Dsn should contain all the necessary information to connect to the database,
	// including the username, password, host, port, and database name.
	// Example: "user:password@tcp(localhost:3306)/dbname"
	Dsn string `validate:"required"`

	// The database driver to use for connecting to the database (e.g., "mysql", "postgres", etc.).
	Driver string `validate:"required"`

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

	Migrations Migrations `validate:"required"`
}

// Populate implements the go-config Config interface for Database.
// It reads values from environment variables providing sensible defaults.
func (d *Database) Populate() error {
	name, _ := params.GetEnvAsString("DB_NAME", "")
	dsn, _ := params.GetEnvAsString("DB_DSN", "")
	driver, _ := params.GetEnvAsString("DB_DRIVER", "mysql")
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

	d.Name = name
	d.Dsn = dsn
	d.Driver = driver
	d.MaxIdleConnections = maxIdleConnections
	d.MaxOpenConnections = maxOpenConnections
	d.ConnectionMaxIdleTime = connectionMaxIdleTime
	d.ConnectionMaxLifetime = connectionMaxLifetime

	return nil
}
