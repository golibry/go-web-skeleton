package testkit

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/golibry/go-params/params"
)

const (
	EnvTestResetDatabase              = "TEST_RESET_DATABASE"
	EnvTestResetDatabaseAllowUnsafe   = "TEST_RESET_DATABASE_ALLOW_UNSAFE"
	EnvTestResetDatabaseAllowAnyEnv   = "TEST_RESET_DATABASE_ALLOW_ANY_ENV"
	EnvTestResetDatabaseAllowAnyDB    = "TEST_RESET_DATABASE_ALLOW_ANY_DATABASE"
	requiredResetDatabaseNameFragment = "test"
)

type SQLDatabaseResetOptions struct {
	Enabled                 bool
	Driver                  string
	AdminDSN                string
	DatabaseName            string
	AllowUnsafe             bool
	AllowNonTestEnvironment bool
	AllowNonTestDatabase    bool
}

func SQLDatabaseResetter[C any](
	options func(config C) SQLDatabaseResetOptions,
) EnvironmentResetFunc[C] {
	return func(ctx context.Context, config C) error {
		if options == nil {
			return nil
		}

		return ResetSQLDatabase(ctx, options(config))
	}
}

func SQLDatabaseResetOptionsFromEnv(defaults SQLDatabaseResetOptions) SQLDatabaseResetOptions {
	enabled, _ := params.GetEnvAsBool(EnvTestResetDatabase, defaults.Enabled)
	allowUnsafe, _ := params.GetEnvAsBool(EnvTestResetDatabaseAllowUnsafe, defaults.AllowUnsafe)
	allowNonTestEnvironment, _ := params.GetEnvAsBool(
		EnvTestResetDatabaseAllowAnyEnv,
		defaults.AllowNonTestEnvironment,
	)
	allowNonTestDatabase, _ := params.GetEnvAsBool(
		EnvTestResetDatabaseAllowAnyDB,
		defaults.AllowNonTestDatabase,
	)

	return SQLDatabaseResetOptions{
		Enabled:                 enabled,
		Driver:                  defaults.Driver,
		AdminDSN:                defaults.AdminDSN,
		DatabaseName:            defaults.DatabaseName,
		AllowUnsafe:             allowUnsafe,
		AllowNonTestEnvironment: allowNonTestEnvironment,
		AllowNonTestDatabase:    allowNonTestDatabase,
	}
}

func ResetSQLDatabase(ctx context.Context, options SQLDatabaseResetOptions) error {
	if !options.Enabled {
		return nil
	}
	if err := ValidateSQLDatabaseResetOptions(options); err != nil {
		return err
	}

	db, err := sql.Open(options.Driver, options.AdminDSN)
	if err != nil {
		return fmt.Errorf("could not open SQL admin connection: %w", err)
	}
	defer func() { _ = db.Close() }()

	if err := DropSQLDatabase(ctx, db, options.Driver, options.DatabaseName); err != nil {
		return err
	}
	if err := CreateSQLDatabase(ctx, db, options.Driver, options.DatabaseName); err != nil {
		return err
	}

	return nil
}

func ValidateSQLDatabaseResetOptions(options SQLDatabaseResetOptions) error {
	appEnv, _ := params.GetEnvAsString("APP_ENV", "")

	if strings.TrimSpace(options.Driver) == "" {
		return fmt.Errorf("missing SQL database reset driver")
	}
	if strings.TrimSpace(options.AdminDSN) == "" {
		return fmt.Errorf("missing SQL database reset admin DSN")
	}
	if strings.TrimSpace(options.DatabaseName) == "" {
		return fmt.Errorf("missing SQL database reset database name")
	}
	if !options.AllowUnsafe && !options.AllowNonTestEnvironment && appEnv != "test" {
		return fmt.Errorf("refusing SQL database reset outside APP_ENV=test")
	}
	if !options.AllowUnsafe && !options.AllowNonTestDatabase &&
		!strings.Contains(strings.ToLower(options.DatabaseName), requiredResetDatabaseNameFragment) {
		return fmt.Errorf(
			"refusing SQL database reset for database %q because its name does not contain %q",
			options.DatabaseName,
			requiredResetDatabaseNameFragment,
		)
	}

	return nil
}

func DropSQLDatabase(ctx context.Context, db *sql.DB, driverName, databaseName string) error {
	if db == nil {
		return fmt.Errorf("missing SQL admin connection")
	}
	if isPostgresDriver(driverName) {
		if err := TerminatePostgresDatabaseConnections(ctx, db, databaseName); err != nil {
			return err
		}
	}

	_, err := db.ExecContext(
		ctx,
		fmt.Sprintf(
			"DROP DATABASE IF EXISTS %s",
			quoteSQLDatabaseIdentifier(driverName, databaseName),
		),
	)
	if err != nil {
		return fmt.Errorf("could not drop SQL test database: %w", err)
	}

	return nil
}

func TerminatePostgresDatabaseConnections(ctx context.Context, db *sql.DB, databaseName string) error {
	if db == nil {
		return fmt.Errorf("missing SQL admin connection")
	}

	_, err := db.ExecContext(
		ctx,
		`SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = $1
AND pid <> pg_backend_pid()`,
		databaseName,
	)
	if err != nil {
		return fmt.Errorf("could not terminate PostgreSQL test database connections: %w", err)
	}

	return nil
}

func CreateSQLDatabase(ctx context.Context, db *sql.DB, driverName, databaseName string) error {
	if db == nil {
		return fmt.Errorf("missing SQL admin connection")
	}

	_, err := db.ExecContext(
		ctx,
		fmt.Sprintf(
			"CREATE DATABASE %s",
			quoteSQLDatabaseIdentifier(driverName, databaseName),
		),
	)
	if err != nil {
		return fmt.Errorf("could not create SQL test database: %w", err)
	}

	return nil
}

func quoteSQLDatabaseIdentifier(driverName, identifier string) string {
	identifier = strings.TrimSpace(identifier)
	if isPostgresDriver(driverName) {
		return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
	}

	return "`" + strings.ReplaceAll(identifier, "`", "``") + "`"
}

func MySQLAdminDSN(databaseDSN, databaseName string) string {
	_ = databaseName
	databaseDSN = strings.TrimSpace(databaseDSN)
	if databaseDSN == "" {
		return ""
	}

	slashIndex := strings.LastIndex(databaseDSN, "/")
	if slashIndex < 0 {
		return databaseDSN
	}

	tail := databaseDSN[slashIndex+1:]
	queryIndex := strings.Index(tail, "?")
	if queryIndex < 0 {
		return databaseDSN[:slashIndex+1]
	}

	return databaseDSN[:slashIndex+1] + tail[queryIndex:]
}

func PostgresAdminDSN(databaseDSN, adminDatabaseName string) string {
	databaseDSN = strings.TrimSpace(databaseDSN)
	if databaseDSN == "" {
		return ""
	}
	if strings.TrimSpace(adminDatabaseName) == "" {
		adminDatabaseName = "postgres"
	}

	slashIndex := strings.LastIndex(databaseDSN, "/")
	if slashIndex < 0 {
		return databaseDSN
	}

	tail := databaseDSN[slashIndex+1:]
	queryIndex := strings.Index(tail, "?")
	if queryIndex < 0 {
		return databaseDSN[:slashIndex+1] + adminDatabaseName
	}

	return databaseDSN[:slashIndex+1] + adminDatabaseName + tail[queryIndex:]
}

func isPostgresDriver(driverName string) bool {
	switch strings.ToLower(strings.TrimSpace(driverName)) {
	case "postgres", "postgresql", "pgx":
		return true
	default:
		return false
	}
}
