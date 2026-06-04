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
DOCKER_COMPOSE_FILE="${DOCKER_COMPOSE_FILE:-$ROOT_DIR/.docker/docker-compose.yml}"

usage() {
	printf 'Usage: %s {test|build|run|migrations|docker} [args...]\n' "$0"
	printf 'Env: MIGRATIONS_DRIVER=mysql|postgres GO_TAGS=<tags> APP_ENTRY=<main package> APP_OUTPUT=<binary path>\n'
	printf 'Docker: %s docker up -d | down | ps | logs\n' "$0"
}

run_docker_compose() {
	if docker compose version >/dev/null 2>&1; then
		docker compose -f "$DOCKER_COMPOSE_FILE" "$@"
		return
	fi

	if command -v docker-compose >/dev/null 2>&1; then
		docker-compose -f "$DOCKER_COMPOSE_FILE" "$@"
		return
	fi

	printf 'Docker Compose is not available. Install Docker Compose and try again.\n' >&2
	return 1
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
	docker)
		shift
		run_docker_compose "$@"
		;;
	*)
		usage
		exit 2
		;;
esac
