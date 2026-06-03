package migrations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/golibry/go-cli-command/cli"
	migrationscli "github.com/golibry/go-migrations/cli"
	"github.com/golibry/go-migrations/execution"
	gomigration "github.com/golibry/go-migrations/migration"
	"github.com/golibry/go-web-skeleton/framework/config"
)

const (
	DriverMySQL             = "mysql"
	DriverPostgres          = "postgres"
	defaultExecutionsTable  = "migration_executions"
	defaultMigrationsDriver = DriverMySQL
)

var (
	ErrMissingDriver          = errors.New("missing migrations driver")
	ErrMissingSQLDriver       = errors.New("missing migrations SQL driver")
	ErrMissingDSN             = errors.New("missing migrations database DSN")
	ErrMissingMigrationsDir   = errors.New("missing migrations directory")
	ErrMissingExecutionsTable = errors.New("missing migrations executions table")
	ErrUnsupportedDriver      = errors.New("unsupported migrations driver")
)

type Migrations struct {
	cli.CommandWithoutFlags
	Options Options
}

type Options struct {
	Context context.Context

	Database config.Database
	DB       *sql.DB
	CloseDB  bool

	Driver          string
	SQLDriver       string
	DSN             string
	MigrationsDir   string
	ExecutionsTable string

	// MigrationsFactory must return migrations that need direct access to the SQL handle.
	// Auto-registered migrations are used when this is nil.
	MigrationsFactory func(db *sql.DB, ctx context.Context) []gomigration.Migration
}

type Runtime struct {
	Context    context.Context
	DB         *sql.DB
	Repository execution.Repository
	Registry   gomigration.MigrationsRegistry
	DirPath    gomigration.MigrationsDirPath
	closeDB    bool
}

type MigrationsExitError struct {
	Code int
}

var newRuntime = NewRuntime

func (e *MigrationsExitError) Error() string {
	return fmt.Sprintf("migrations command failed with exit code %d", e.Code)
}

func (e *MigrationsExitError) ExitCode() int {
	return e.Code
}

func (c *Migrations) Id() string {
	return "migrations"
}

func (c *Migrations) Description() string {
	return "Handles database migrations"
}

func (c *Migrations) Exec(stdWriter io.Writer) error {
	return Run(c.Options, os.Args[2:], stdWriter)
}

func Run(options Options, args []string, writer io.Writer) error {
	runtime, err := newRuntime(options)
	if err != nil {
		return err
	}
	defer func() { _ = runtime.Close() }()

	exitCode := 0
	migrationscli.Bootstrap(
		runtime.Context,
		runtime.DB,
		args,
		runtime.Registry,
		runtime.Repository,
		runtime.DirPath,
		nil,
		writer,
		func(code int) {
			exitCode = code
		},
		nil,
	)

	if exitCode != 0 {
		return &MigrationsExitError{Code: exitCode}
	}

	return nil
}

func NewRuntime(options Options) (*Runtime, error) {
	options = options.withDefaults()
	if err := options.Validate(); err != nil {
		return nil, err
	}

	db, closeDB, err := options.databaseHandle()
	if err != nil {
		return nil, fmt.Errorf("could not build migrations runtime: failed to build database connection: %w", err)
	}

	dirPath, err := gomigration.NewMigrationsDirPath(options.MigrationsDir)
	if err != nil {
		if closeDB {
			_ = db.Close()
		}
		return nil, fmt.Errorf("could not build migrations runtime: invalid migrations directory path: %w", err)
	}

	repo, err := newRepository(options, db)
	if err != nil {
		if closeDB {
			_ = db.Close()
		}
		return nil, fmt.Errorf("could not build migrations runtime: failed to build executions repository: %w", err)
	}

	return &Runtime{
		Context:    options.Context,
		DB:         db,
		Repository: repo,
		Registry:   buildRegistry(dirPath, db, options.Context, options.MigrationsFactory),
		DirPath:    dirPath,
		closeDB:    closeDB,
	}, nil
}

func (r *Runtime) Close() error {
	if r == nil || !r.closeDB || r.DB == nil {
		return nil
	}

	return r.DB.Close()
}

func (o Options) Validate() error {
	if o.Driver == "" {
		return ErrMissingDriver
	}
	if o.SQLDriver == "" && o.DB == nil {
		return ErrMissingSQLDriver
	}
	if o.DSN == "" && o.DB == nil {
		return ErrMissingDSN
	}
	if o.MigrationsDir == "" {
		return ErrMissingMigrationsDir
	}
	if o.ExecutionsTable == "" {
		return ErrMissingExecutionsTable
	}

	return nil
}

func (o Options) withDefaults() Options {
	if o.Context == nil {
		o.Context = context.Background()
	}
	if o.Driver == "" {
		o.Driver = o.Database.Driver
	}
	if o.Driver == "" {
		o.Driver = defaultMigrationsDriver
	}
	if o.SQLDriver == "" {
		o.SQLDriver = defaultSQLDriverName(o.Driver)
	}
	o.Driver = canonicalDriverName(o.Driver)
	if o.DSN == "" {
		o.DSN = o.Database.Dsn
	}
	if o.MigrationsDir == "" {
		o.MigrationsDir = o.Database.Migrations.MigrationsDirPath
	}
	if o.ExecutionsTable == "" {
		o.ExecutionsTable = o.Database.Migrations.ExecutionsTable
	}
	if o.ExecutionsTable == "" {
		o.ExecutionsTable = defaultExecutionsTable
	}

	return o
}

func (o Options) databaseHandle() (*sql.DB, bool, error) {
	if o.DB != nil {
		return o.DB, o.CloseDB, nil
	}

	db, err := sql.Open(o.SQLDriver, o.DSN)
	if err != nil {
		return nil, false, err
	}

	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(1)
	db.SetConnMaxIdleTime(0)
	db.SetConnMaxLifetime(0)

	return db, true, nil
}

func buildRegistry(
	dirPath gomigration.MigrationsDirPath,
	db *sql.DB,
	ctx context.Context,
	factory func(db *sql.DB, ctx context.Context) []gomigration.Migration,
) gomigration.MigrationsRegistry {
	if factory == nil {
		return gomigration.NewAutoDirMigrationsRegistry(dirPath)
	}

	migrations := factory(db, ctx)
	if gomigration.DefaultRegistry.Count() > 0 {
		migrations = append(migrations, gomigration.DefaultRegistry.OrderedMigrations()...)
	}

	return gomigration.NewDirMigrationsRegistry(dirPath, migrations)
}

func unsupportedDriverError(driverName string) error {
	return fmt.Errorf("%w %q", ErrUnsupportedDriver, canonicalDriverName(driverName))
}

func canonicalDriverName(driverName string) string {
	switch strings.ToLower(strings.TrimSpace(driverName)) {
	case "mariadb", "mysql":
		return DriverMySQL
	case "postgres", "postgresql", "pgx":
		return DriverPostgres
	default:
		return strings.ToLower(strings.TrimSpace(driverName))
	}
}

func defaultSQLDriverName(driverName string) string {
	switch strings.ToLower(strings.TrimSpace(driverName)) {
	case "mariadb":
		return DriverMySQL
	case "postgresql":
		return DriverPostgres
	default:
		return strings.ToLower(strings.TrimSpace(driverName))
	}
}
