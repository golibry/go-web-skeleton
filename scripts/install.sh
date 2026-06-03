#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

ask_yes_no() {
	local prompt="$1"
	local default="${2:-y}"
	local answer

	while true; do
		if [ "$default" = "y" ]; then
			read -r -p "$prompt [Y/n] " answer
			answer="${answer:-y}"
		else
			read -r -p "$prompt [y/N] " answer
			answer="${answer:-n}"
		fi

		case "$answer" in
			y|Y|yes|YES) return 0 ;;
			n|N|no|NO) return 1 ;;
		esac
	done
}

ask_migrations_driver() {
	local answer

	while true; do
		read -r -p "Migrations driver (mysql/postgres) [mysql] " answer
		answer="${answer:-mysql}"

		case "$answer" in
			mysql|postgres)
				printf '%s\n' "$answer"
				return 0
				;;
		esac
	done
}

write_app_env() {
	local migrations_enabled="$1"
	local migrations_driver="${2:-}"

	{
		printf 'MIGRATIONS_ENABLED=%s\n' "$migrations_enabled"
		if [ -n "$migrations_driver" ]; then
			printf 'MIGRATIONS_DRIVER=%s\n' "$migrations_driver"
		fi
	} > "$ROOT_DIR/scripts/.app.env"
}

if ! ask_yes_no "Enable database migrations?" "y"; then
	rm -rf "$ROOT_DIR/framework/migrations"
	rm -rf "$ROOT_DIR/cli/commands/migrations"
	write_app_env "false"
	printf 'Migrations disabled.\n'
	exit 0
fi

driver="$(ask_migrations_driver)"

case "$driver" in
	mysql)
		rm -f "$ROOT_DIR/framework/migrations/repository_postgres.go"
		;;
	postgres)
		rm -f "$ROOT_DIR/framework/migrations/repository_mysql.go"
		;;
esac

write_app_env "true" "$driver"
printf 'Migrations enabled with %s.\n' "$driver"
