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

ask_database_driver() {
	local prompt="${1:-Database service}"
	local answer

	while true; do
		read -r -p "$prompt (mysql/postgres) [mysql] " answer
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
	local docker_enabled="${3:-false}"
	local docker_database_driver="${4:-}"

	{
		printf 'MIGRATIONS_ENABLED=%s\n' "$migrations_enabled"
		if [ -n "$migrations_driver" ]; then
			printf 'MIGRATIONS_DRIVER=%s\n' "$migrations_driver"
		fi
		printf 'DOCKER_DEV_SERVICES=%s\n' "$docker_enabled"
		if [ -n "$docker_database_driver" ]; then
			printf 'DOCKER_DATABASE_DRIVER=%s\n' "$docker_database_driver"
		fi
	} > "$ROOT_DIR/scripts/.app.env"
}

upsert_env_value() {
	local file="$1"
	local key="$2"
	local value="$3"
	local escaped_value

	mkdir -p "$(dirname "$file")"
	touch "$file"

	escaped_value="${value//\\/\\\\}"
	escaped_value="${escaped_value//&/\\&}"

	if grep -q "^${key}=" "$file"; then
		sed -i.bak "s#^${key}=.*#${key}=${escaped_value}#" "$file"
		rm -f "$file.bak"
	else
		printf '%s=%s\n' "$key" "$value" >> "$file"
	fi
}

write_docker_env_defaults() {
	local driver="$1"
	local env_file="$ROOT_DIR/.env.local"

	upsert_env_value "$env_file" "DB_DRIVER" "$driver"
	upsert_env_value "$env_file" "DB_HOST" "127.0.0.1"
	upsert_env_value "$env_file" "DB_NAME" "go_web_skeleton"
	upsert_env_value "$env_file" "DB_USER" "app_user"
	upsert_env_value "$env_file" "DB_PASSWORD" "app_password"

	case "$driver" in
		mysql)
			upsert_env_value "$env_file" "DB_PORT" "3306"
			upsert_env_value "$env_file" "DB_ROOT_PASSWORD" "root_password"
			upsert_env_value "$env_file" "DB_DSN" '${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}?parseTime=true&charset=utf8mb4'
			;;
		postgres)
			upsert_env_value "$env_file" "DB_PORT" "5432"
			upsert_env_value "$env_file" "DB_DSN" 'postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable'
			;;
	esac
}

write_mysql_compose() {
	mkdir -p "$ROOT_DIR/.docker"
	cat > "$ROOT_DIR/.docker/docker-compose.yml" <<'YAML'
name: go-web-skeleton

services:
  mysql:
    image: mysql:8.4
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: ${DB_ROOT_PASSWORD:-root_password}
      MYSQL_DATABASE: ${DB_NAME:-go_web_skeleton}
      MYSQL_USER: ${DB_USER:-app_user}
      MYSQL_PASSWORD: ${DB_PASSWORD:-app_password}
    ports:
      - "${DB_PORT:-3306}:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD-SHELL", "mysqladmin ping -h 127.0.0.1 -u root -p$${MYSQL_ROOT_PASSWORD}"]
      interval: 10s
      timeout: 5s
      retries: 10
      start_period: 30s

volumes:
  mysql_data:
YAML
}

write_postgres_compose() {
	mkdir -p "$ROOT_DIR/.docker"
	cat > "$ROOT_DIR/.docker/docker-compose.yml" <<'YAML'
name: go-web-skeleton

services:
  postgres:
    image: postgres:17
    restart: unless-stopped
    environment:
      POSTGRES_DB: ${DB_NAME:-go_web_skeleton}
      POSTGRES_USER: ${DB_USER:-app_user}
      POSTGRES_PASSWORD: ${DB_PASSWORD:-app_password}
    ports:
      - "${DB_PORT:-5432}:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U $${POSTGRES_USER} -d $${POSTGRES_DB}"]
      interval: 10s
      timeout: 5s
      retries: 10
      start_period: 15s

volumes:
  postgres_data:
YAML
}

install_docker_setup() {
	local driver="$1"

	case "$driver" in
		mysql)
			write_mysql_compose
			;;
		postgres)
			write_postgres_compose
			;;
	esac

	write_docker_env_defaults "$driver"
	printf 'Docker dev services enabled with %s.\n' "$driver"
}

remove_docker_setup() {
	rm -rf "$ROOT_DIR/.docker"
	printf 'Docker dev services disabled.\n'
}

migrations_enabled="true"
migrations_driver=""

if ! ask_yes_no "Enable database migrations?" "y"; then
	rm -rf "$ROOT_DIR/framework/migrations"
	rm -rf "$ROOT_DIR/cli/commands/migrations"
	migrations_enabled="false"
	printf 'Migrations disabled.\n'
else
	migrations_driver="$(ask_migrations_driver)"

	case "$migrations_driver" in
		mysql)
			rm -f "$ROOT_DIR/framework/migrations/repository_postgres.go"
			;;
		postgres)
			rm -f "$ROOT_DIR/framework/migrations/repository_mysql.go"
			;;
	esac

	printf 'Migrations enabled with %s.\n' "$migrations_driver"
fi

docker_enabled="false"
docker_database_driver=""
if ask_yes_no "Install local Docker dev services?" "y"; then
	docker_enabled="true"
	if [ -n "$migrations_driver" ]; then
		docker_database_driver="$migrations_driver"
	else
		docker_database_driver="$(ask_database_driver "Docker database service")"
	fi
	install_docker_setup "$docker_database_driver"
else
	remove_docker_setup
fi

write_app_env "$migrations_enabled" "$migrations_driver" "$docker_enabled" "$docker_database_driver"
