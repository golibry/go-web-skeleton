package migrations

import (
	"context"
	"database/sql"
	"os"
	"testing"

	gomigration "github.com/golibry/go-migrations/migration"
	"github.com/golibry/go-web-skeleton/framework/config"
)

func TestCanonicalDriverName(t *testing.T) {
	tests := map[string]string{
		"mysql":      DriverMySQL,
		"mariadb":    DriverMySQL,
		"postgres":   DriverPostgres,
		"postgresql": DriverPostgres,
		"pgx":        DriverPostgres,
		" custom ":   "custom",
	}

	for input, expected := range tests {
		if got := canonicalDriverName(input); got != expected {
			t.Fatalf("CanonicalDriverName(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestDefaultSQLDriverName(t *testing.T) {
	tests := map[string]string{
		"mysql":      "mysql",
		"mariadb":    "mysql",
		"postgres":   "postgres",
		"postgresql": "postgres",
		"pgx":        "pgx",
	}

	for input, expected := range tests {
		if got := defaultSQLDriverName(input); got != expected {
			t.Fatalf("DefaultSQLDriverName(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestOptionsWithDefaultsUsesDatabaseConfig(t *testing.T) {
	options := Options{
		Database: config.Database{
			Dsn:    "db-dsn",
			Driver: "postgresql",
			Migrations: config.Migrations{
				MigrationsDirPath: "migrations",
				ExecutionsTable:   "executions",
			},
		},
	}

	resolved := options.withDefaults()
	if resolved.DSN != "db-dsn" {
		t.Fatalf("DSN = %q, want db-dsn", resolved.DSN)
	}
	if resolved.Driver != DriverPostgres {
		t.Fatalf("Driver = %q, want %q", resolved.Driver, DriverPostgres)
	}
	if resolved.SQLDriver != DriverPostgres {
		t.Fatalf("SQLDriver = %q, want %q", resolved.SQLDriver, DriverPostgres)
	}
	if resolved.MigrationsDir != "migrations" {
		t.Fatalf("MigrationsDir = %q, want migrations", resolved.MigrationsDir)
	}
	if resolved.ExecutionsTable != "executions" {
		t.Fatalf("ExecutionsTable = %q, want executions", resolved.ExecutionsTable)
	}
}

func TestBuildRegistryUsesMigrationFactory(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(dir+"/version_1.go", []byte("package migrations\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	dirPath, err := gomigration.NewMigrationsDirPath(dir)
	if err != nil {
		t.Fatalf("NewMigrationsDirPath() error = %v", err)
	}

	registry := buildRegistry(
		dirPath,
		&sql.DB{},
		context.Background(),
		func(db *sql.DB, ctx context.Context) []gomigration.Migration {
			return []gomigration.Migration{testMigration{}}
		},
	)

	if registry.Count() != 1 {
		t.Fatalf("registry.Count() = %d, want 1", registry.Count())
	}
}

type testMigration struct{}

func (testMigration) Version() uint64 { return 1 }
func (testMigration) Up(context.Context, any) error {
	return nil
}
func (testMigration) Down(context.Context, any) error {
	return nil
}
