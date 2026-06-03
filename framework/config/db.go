package config

import (
	"path/filepath"
	"time"

	"github.com/golibry/go-params/params"
)

type Migrations struct {
	// MigrationsDirPath specifies the directory path containing database migration files.
	// This must be a valid directory path and is required for database migrations.
	MigrationsDirPath string `env:"DB_MIGRATIONS_DIR_PATH" validate:"required,dir"`

	// ExecutionsTable specifies the name of the table that stores migration execution information.
	ExecutionsTable string `env:"DB_MIGRATIONS_TABLE" default:"migrations_executions" validate:"required"`
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
	Name string `env:"DB_NAME" validate:"required"`

	// Dsn is the Data Source Name for the database connection.
	// It is a required field and is validated to ensure it is not empty.
	// The Dsn should contain all the necessary information to connect to the database,
	// including the username, password, host, port, and database name.
	// Example: "user:password@tcp(localhost:3306)/dbname"
	Dsn string `env:"DB_DSN" validate:"required"`

	// The database driver to use for connecting to the database (e.g., "mysql", "postgres", etc.).
	Driver string `env:"DB_DRIVER" default:"mysql" validate:"required"`

	// MaxIdleConnections sets the maximum number of connections in the idle connection pool.
	// Must be between 0 and 99. A value of 0 means no idle connections are retained.
	MaxIdleConnections int `env:"DB_MAX_IDLE_CONNECTIONS" default:"2" validate:"number,gte=0,lte=99"`

	// MaxOpenConnections sets the maximum number of open connections to the database.
	// Must be between 0 and 99. A value of 0 means unlimited connections.
	MaxOpenConnections int `env:"DB_MAX_OPEN_CONNECTIONS" default:"10" validate:"number,gte=0,lte=99"`

	// ConnectionMaxIdleTime sets the maximum amount of time a connection may be idle.
	// Expired connections may be closed lazily before reuse.
	ConnectionMaxIdleTime time.Duration `env:"DB_CONNECTION_MAX_IDLE_TIME" default:"3m"`

	// ConnectionMaxLifetime sets the maximum amount of time a connection may be reused.
	// Expired connections may be closed lazily before reuse.
	ConnectionMaxLifetime time.Duration `env:"DB_CONNECTION_MAX_LIFETIME" default:"3m"`

	Migrations Migrations `validate:"required"`
}
