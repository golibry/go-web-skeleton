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

The Bash installer asks what the app needs, then keeps only the files required by those choices. The generated app imports the framework packages through thin entrypoints, so application code does not need to know the lower-level bootstrap details.

For example, migrations are installed as a CLI command. If migrations are enabled, the installer asks for the database driver and keeps only one driver implementation:

- `framework/migrations/repository_mysql.go`
- `framework/migrations/repository_postgres.go`

The installed app should keep only the selected one.

## Install

Run the installer from the project root:

```bash
scripts/install.sh
```

The installer currently asks whether to enable database migrations. If enabled, it asks which driver to use:

- `mysql`
- `postgres`

The installer writes local choices to:

```text
scripts/.app.env
```

That file is intentionally local and ignored by Git.

## App Commands

Use the Bash app helper after installation:

```bash
scripts/app.sh test
scripts/app.sh build
scripts/app.sh run http:start
scripts/app.sh migrations up
```

`scripts/app.sh` reads `scripts/.app.env` and applies the selected build tags automatically. You can still override values when needed:

```bash
MIGRATIONS_DRIVER=postgres scripts/app.sh test
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
