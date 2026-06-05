# go-web-skeleton

`go-web-skeleton` is the Golibry starter framework for building Go web applications with very little setup. The goal is to install a small application skeleton, choose the pieces your app needs, change the config, and start building application code.

Reusable framework code lives under:

- `framework/app`: application lifecycle, typed container, logger, SQL DB, and response builder helpers
- `framework/cli`: generic CLI bootstrap
- `framework/config`: config loading, validation, debug output, and common config structs
- `framework/http`: HTTP server runtime and HTTP CLI command adapter
- `framework/migrations`: migrations runtime and migrations CLI command adapter

The skeleton is organized for CQRS-style development:

- `cli`: thin app CLI entrypoints and command aliases
- `application`: application commands, queries, handlers, and use cases
- `domain`: domain models, value objects, and domain rules
- `infrastructure`: adapters for persistence, logging, external systems, and runtime services
- `presentation`: app-owned HTTP entrypoints such as controllers and routes; server bootstrap belongs in `framework/http`
- `migrations`: application migration files
- `config`: thin app config aliases and application-specific config extensions

`infrastructure/registry` is an app-facing registry/container location for generated applications. It should stay thin and typed. The reusable implementation lives in `framework/app`, while each app can wrap it with application-specific getters and services.

## Framework Approach

This module is not meant to behave like a large runtime framework that wires every choice dynamically. Instead, it is meant to be installed as a tailored app skeleton with reusable framework packages.

The installer is a Go command that copies real scaffold files from `install/app` and selected variant files from `install/variants`. The generated app imports the framework packages through thin entrypoints, so application code does not need to know the lower-level bootstrap details.

For example, the installer copies only one database driver variant:

- `install/variants/database/mysql/infrastructure/database/driver_mysql.go`
- `install/variants/database/postgres/infrastructure/database/driver_postgres.go`
- `install/variants/database/sqlite/infrastructure/database/driver_sqlite.go`

The installed app keeps only the selected one.

## Install

Run the Go installer directly:

```bash
go run github.com/golibry/go-web-skeleton/cmd/golibry-web@latest install
```

The installer asks for the application name, Go module path, database driver, whether to install migrations, and whether to install local Docker database files. For non-interactive installs, pass flags:

```bash
go run github.com/golibry/go-web-skeleton/cmd/golibry-web@latest install \
  -app-name my-app \
  -module-path example.com/my-app \
  -database-driver mysql \
  -migrations=true \
  -docker=true
```

Use `-database-driver sqlite` for a local SQLite app. SQLite installs the embedded database driver and skips local Docker database files.

The installer uses these real files as its source of truth:

```text
install/app/
install/variants/database/mysql/
install/variants/database/postgres/
install/variants/database/sqlite/
```

The installer mostly copies files as-is. The only rendered placeholder is the app module path needed by Go imports in the generated CLI entrypoint.

## App Commands

Use the Bash app helper after installation:

```bash
scripts/app.sh test
scripts/app.sh build
scripts/app.sh run http:start
scripts/app.sh migrations stats
```

`scripts/app.sh` applies the selected database build tag automatically. You can still override values when needed:

```bash
DATABASE_DRIVER=postgres scripts/app.sh test
DATABASE_DRIVER=sqlite scripts/app.sh test
APP_ENTRY=./cmd/app scripts/app.sh build
```

## Design Goal

The generated application should require minimal wiring from the consumer. Most common web-app concerns should already have a clear place:

- config loads once and is shared through the app container
- common services like logger, DB, and response builder are created by the framework container
- CLI commands act as process entrypoints
- HTTP server bootstrap is provided by the skeleton
- migrations run through the CLI command when installed
- optional modules are selected during installation instead of being configured at runtime

After installation, the main job of the app developer should be changing config, adding domain/application code, and adding presentation entrypoints.
