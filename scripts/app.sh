#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

if [ -f "$ROOT_DIR/scripts/.app.env" ]; then
	# shellcheck disable=SC1091
	. "$ROOT_DIR/scripts/.app.env"
fi

MIGRATIONS_DRIVER="${MIGRATIONS_DRIVER:-mysql}"
GO_TAGS="${GO_TAGS:-$MIGRATIONS_DRIVER}"
APP_ENTRY="${APP_ENTRY:-./_examples/advanced/cli}"
APP_OUTPUT="${APP_OUTPUT:-./build/app}"

usage() {
	printf 'Usage: %s {test|build|run|migrations} [args...]\n' "$0"
	printf 'Env: MIGRATIONS_DRIVER=mysql|postgres GO_TAGS=<tags> APP_ENTRY=<main package> APP_OUTPUT=<binary path>\n'
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
		go run -tags="$GO_TAGS" "$APP_ENTRY" migrations "$@"
		;;
	*)
		usage
		exit 2
		;;
esac
