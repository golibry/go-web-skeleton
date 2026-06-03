# Testing

The skeleton keeps testing support split the same way as the application runtime:

- `framework/testing` contains small reusable test bootstrap helpers.
- each generated app owns its own test bootstrap, usually under `test/bootstrap.go`.

The framework testkit does not know about an application's config struct, database driver choice, migrations package, or cleanup strategy. The app-specific bootstrap provides those choices.

## Framework Testkit

`framework/testing` provides:

- `EnsureTestEnv()`: sets `APP_ENV=test` when no environment was selected
- `Setup(...)`: loads config through an app-provided function, runs `BeforeSetup` hooks, builds a framework container, and optionally calls an app-provided migration callback
- `Bootstrap.Cleanup(...)`: runs app-registered cleanup hooks for extra resources
- `Bootstrap.Close()`: closes resources created by the container

## App Bootstrap

An app bootstrap should decide:

- which config struct to load
- which common services the container should create
- whether migrations should run before integration tests
- how the full test environment is reset before a suite starts
- whether extra cleanup is needed after tests

The advanced example has a minimal app-specific bootstrap at:

```text
_examples/advanced/test/bootstrap.go
```

## Usage

Use the app bootstrap from tests that need integration resources:

```go
func TestSomething(t *testing.T) {
	bootstrap, err := test.Setup()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = bootstrap.Close() }()

	db := bootstrap.Container.DB()
	logger := bootstrap.Container.Logger()

	_ = db
	_ = logger
}
```

For package-level setup, wrap it in `TestMain`:

```go
func TestMain(m *testing.M) {
	bootstrap, err := test.Setup()
	if err != nil {
		panic(err)
	}

	code := m.Run()
	_ = bootstrap.Close()
	os.Exit(code)
}
```

## Environment Reset

The framework does not delete or recreate databases automatically. Reset is explicit: the app registers `BeforeSetup` hooks in its bootstrap. Those hooks run during `Setup(...)`, before the framework opens the normal app DB connection and before migrations run.

For integration and e2e suites, the preferred DB strategy is:

- drop the disposable test database
- recreate the disposable test database
- run migrations into the clean database
- run the suite

`framework/testing` exposes `SQLDatabaseResetter` for this. The app still decides the admin DSN, database name, and whether this reset is appropriate for the environment.

The SQL reset helper uses the normal app database config for driver and database name. The only reset env var normally needed is:

```text
TEST_RESET_DATABASE=true
```

The app bootstrap derives the admin connection from the normal app DSN before the framework opens the app database connection.

Because database reset is destructive, the helper refuses to run unless:

- `APP_ENV=test`
- the database name contains `test`

Those guardrails can be overridden only with explicit env vars:

```text
TEST_RESET_DATABASE_ALLOW_UNSAFE=true
TEST_RESET_DATABASE_ALLOW_ANY_ENV=true
TEST_RESET_DATABASE_ALLOW_ANY_DATABASE=true
```

Apps can also register extra cleanup hooks for cache keys, queues, files, search indexes, or any other state used by tests.

The intended lifecycle is:

- `BeforeSetup`: reset full environment before the container opens app resources
- `Migrate`: create the clean schema after the reset
- tests: run against the clean environment
- `Cleanup`: optionally clear non-DB resources after tests
- `Close`: close framework services

## Cleanup Helpers

`framework/testing` includes small opt-in cleanup helpers:

- `FilesystemCleaner`: removes configured paths
- `DirectoryCleaner`: empties a configured directory
- `RedisCleaner`: accepts a small Redis-compatible interface and either flushes the selected DB or deletes configured keys

The Redis helper does not import a Redis client package. Generated apps can adapt whichever Redis client they install.

The old global test bootstrap that dropped and recreated MySQL databases implicitly was removed because it was too application-specific. The new version keeps the operation explicit and app-owned.
