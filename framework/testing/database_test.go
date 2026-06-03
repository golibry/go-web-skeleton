package testkit

import "testing"

func TestSQLDatabaseResetOptionsFromEnvOverridesDefaults(t *testing.T) {
	t.Setenv(EnvTestResetDatabase, "true")
	t.Setenv(EnvTestResetDatabaseAllowUnsafe, "true")
	t.Setenv(EnvTestResetDatabaseAllowAnyEnv, "true")
	t.Setenv(EnvTestResetDatabaseAllowAnyDB, "true")

	options := SQLDatabaseResetOptionsFromEnv(
		SQLDatabaseResetOptions{
			Enabled:      false,
			Driver:       "mysql",
			AdminDSN:     "mysql-admin",
			DatabaseName: "mysql_test",
		},
	)

	if !options.Enabled {
		t.Fatal("Enabled = false, want true")
	}
	if options.Driver != "mysql" {
		t.Fatalf("Driver = %q, want mysql", options.Driver)
	}
	if options.AdminDSN != "mysql-admin" {
		t.Fatalf("AdminDSN = %q, want mysql admin DSN", options.AdminDSN)
	}
	if options.DatabaseName != "mysql_test" {
		t.Fatalf("DatabaseName = %q, want mysql_test", options.DatabaseName)
	}
	if !options.AllowUnsafe || !options.AllowNonTestEnvironment || !options.AllowNonTestDatabase {
		t.Fatal("unsafe override flags were not enabled")
	}
}

func TestResetSQLDatabaseDisabledSkipsValidation(t *testing.T) {
	if err := ResetSQLDatabase(nil, SQLDatabaseResetOptions{}); err != nil {
		t.Fatalf("ResetSQLDatabase() error = %v", err)
	}
}

func TestValidateSQLDatabaseResetOptionsRejectsNonTestEnvironment(t *testing.T) {
	t.Setenv("APP_ENV", "dev")

	err := ValidateSQLDatabaseResetOptions(
		SQLDatabaseResetOptions{
			Enabled:      true,
			Driver:       "mysql",
			AdminDSN:     "root@tcp(localhost:3306)/",
			DatabaseName: "app_test",
		},
	)
	if err == nil {
		t.Fatal("ValidateSQLDatabaseResetOptions() error = nil, want error")
	}
}

func TestValidateSQLDatabaseResetOptionsRejectsNonTestDatabaseName(t *testing.T) {
	t.Setenv("APP_ENV", "test")

	err := ValidateSQLDatabaseResetOptions(
		SQLDatabaseResetOptions{
			Enabled:      true,
			Driver:       "mysql",
			AdminDSN:     "root@tcp(localhost:3306)/",
			DatabaseName: "production",
		},
	)
	if err == nil {
		t.Fatal("ValidateSQLDatabaseResetOptions() error = nil, want error")
	}
}

func TestValidateSQLDatabaseResetOptionsAllowsUnsafeOverride(t *testing.T) {
	t.Setenv("APP_ENV", "dev")

	err := ValidateSQLDatabaseResetOptions(
		SQLDatabaseResetOptions{
			Enabled:      true,
			Driver:       "mysql",
			AdminDSN:     "root@tcp(localhost:3306)/",
			DatabaseName: "production",
			AllowUnsafe:  true,
		},
	)
	if err != nil {
		t.Fatalf("ValidateSQLDatabaseResetOptions() error = %v", err)
	}
}

func TestMySQLAdminDSN(t *testing.T) {
	tests := map[string]string{
		"user:pass@tcp(localhost:3306)/app_test":                        "user:pass@tcp(localhost:3306)/",
		"user:pass@tcp(localhost:3306)/app_test?parseTime=true":         "user:pass@tcp(localhost:3306)/?parseTime=true",
		"user:pass@tcp(localhost:3306)/app_test?charset=utf8mb4&x=true": "user:pass@tcp(localhost:3306)/?charset=utf8mb4&x=true",
	}

	for input, expected := range tests {
		if got := MySQLAdminDSN(input, "app_test"); got != expected {
			t.Fatalf("MySQLAdminDSN(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestPostgresAdminDSN(t *testing.T) {
	tests := map[string]string{
		"postgres://user:pass@localhost:5432/app_test":                 "postgres://user:pass@localhost:5432/postgres",
		"postgres://user:pass@localhost:5432/app_test?sslmode=disable": "postgres://user:pass@localhost:5432/postgres?sslmode=disable",
	}

	for input, expected := range tests {
		if got := PostgresAdminDSN(input, "postgres"); got != expected {
			t.Fatalf("PostgresAdminDSN(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestQuoteSQLDatabaseIdentifier(t *testing.T) {
	if got := quoteSQLDatabaseIdentifier("mysql", "a`b"); got != "`a``b`" {
		t.Fatalf("mysql identifier = %q", got)
	}
	if got := quoteSQLDatabaseIdentifier("postgres", `a"b`); got != `"a""b"` {
		t.Fatalf("postgres identifier = %q", got)
	}
}
