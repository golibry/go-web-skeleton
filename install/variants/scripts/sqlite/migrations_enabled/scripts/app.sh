#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

DATABASE_DRIVER="${DATABASE_DRIVER:-sqlite}"
GO_TAGS="${GO_TAGS:-$DATABASE_DRIVER}"
APP_ENTRY="${APP_ENTRY:-./cli}"
APP_OUTPUT="${APP_OUTPUT:-./build/app}"
MIGRATIONS_ENABLED="true"

usage() {
	printf 'Usage: %s {test|build|run|migrations|docker} [args...]\n' "$0"
	printf 'Env: DATABASE_DRIVER=mysql|postgres|sqlite GO_TAGS=<tags> APP_ENTRY=<main package> APP_OUTPUT=<binary path>\n'
	printf 'Docker: SQLite apps do not install local Docker database files.\n'
}

case "${1:-}" in
	test)
		shift
		if [ "$#" -eq 0 ]; then
			set -- ./...
		fi
		go test -tags="$GO_TAGS" "$@"
		;;
	build)
		shift
		mkdir -p "$(dirname "$APP_OUTPUT")"
		go build -tags="$GO_TAGS" -o "$APP_OUTPUT" "$@" "$APP_ENTRY"
		;;
	run)
		shift
		go run -tags="$GO_TAGS" "$APP_ENTRY" "$@"
		;;
	migrations)
		shift
		if [ "$MIGRATIONS_ENABLED" != "true" ]; then
			printf 'Migrations are disabled for this app.\n' >&2
			exit 2
		fi
		go run -tags="$GO_TAGS" "$APP_ENTRY" migrations "$@"
		;;
	docker)
		printf 'SQLite apps do not install local Docker database files.\n' >&2
		exit 2
		;;
	*)
		usage
		exit 2
		;;
esac
