#!/usr/bin/env bash
#
# Shared database access for the operational scripts.
#
# The stack can run two ways: MySQL inside Docker Compose, or a MySQL the
# developer already runs locally. These helpers pick whichever is actually
# present, so the same scripts work in both cases.
#
# Override the choice with DB_BACKEND=docker|local.

# Resolve configuration from backend/.env (local development) and then the root
# .env (deployment). The root file is loaded last so that on a server it wins,
# even if a stray backend/.env has been left behind.
db_load_env() {
  local root="$1"
  local file

  for file in "$root/backend/.env" "$root/.env"; do
    [[ -f "$file" ]] || continue
    set -a
    # shellcheck disable=SC1091
    source "$file"
    set +a
  done

  DB_NAME="${DB_NAME:-portfolio}"
  DB_HOST="${DB_HOST:-127.0.0.1}"
  DB_PORT="${DB_PORT:-3306}"
  DB_SERVICE="${DB_SERVICE:-db}"
}

# Decide how to reach MySQL.
db_backend() {
  if [[ -n "${DB_BACKEND:-}" ]]; then
    echo "$DB_BACKEND"
    return
  fi

  if command -v docker >/dev/null 2>&1 \
    && docker compose ps --status running --services 2>/dev/null | grep -qx "${DB_SERVICE:-db}"; then
    echo docker
    return
  fi

  if command -v mysql >/dev/null 2>&1; then
    echo local
    return
  fi

  echo none
}

# Credentials differ by backend: the Compose container has a root password,
# a local server is usually reached as the application user.
db_credentials() {
  case "$(db_backend)" in
    docker)
      : "${DB_ROOT_PASSWORD:?DB_ROOT_PASSWORD must be set in .env}"
      DB_CLI_USER="root"
      DB_CLI_PASSWORD="$DB_ROOT_PASSWORD"
      ;;
    local)
      : "${DB_USER:?DB_USER must be set in .env}"
      DB_CLI_USER="$DB_USER"
      DB_CLI_PASSWORD="${DB_PASSWORD:-}"
      ;;
    *)
      echo "error: no way to reach MySQL — start Docker Compose, or install the mysql client" >&2
      exit 69
      ;;
  esac
}

# Run the mysql client against the configured database.
# MYSQL_PWD is used rather than --password so the secret never appears in the
# process list, where any other user on the host could read it.
#
# --default-character-set is not optional. The client ships with it set to
# `auto`, which resolves from the shell's locale, and on a machine that reports
# no UTF-8 locale it lands on latin1. Every migration in this repository is a
# UTF-8 file, so a latin1 connection hands MySQL three latin1 characters where
# the file had one em dash and stores the mojibake permanently — which is how
# "maintenance — not just writing code" became "maintenance â€” not just
# writing code" on the live site. The same applies to anything in Khmer.
db_mysql() {
  db_credentials
  case "$(db_backend)" in
    docker)
      docker compose exec -T -e MYSQL_PWD="$DB_CLI_PASSWORD" "$DB_SERVICE" \
        mysql --default-character-set=utf8mb4 --user="$DB_CLI_USER" "$@"
      ;;
    local)
      MYSQL_PWD="$DB_CLI_PASSWORD" mysql --default-character-set=utf8mb4 \
        --host="$DB_HOST" --port="$DB_PORT" --user="$DB_CLI_USER" "$@"
      ;;
  esac
}

# Run mysqldump against the configured database.
#
# Same reasoning as db_mysql, and it matters more here: a dump taken over a
# latin1 connection is a backup that cannot be restored faithfully.
db_mysqldump() {
  db_credentials
  case "$(db_backend)" in
    docker)
      docker compose exec -T -e MYSQL_PWD="$DB_CLI_PASSWORD" "$DB_SERVICE" \
        mysqldump --default-character-set=utf8mb4 --user="$DB_CLI_USER" "$@"
      ;;
    local)
      MYSQL_PWD="$DB_CLI_PASSWORD" mysqldump --default-character-set=utf8mb4 \
        --host="$DB_HOST" --port="$DB_PORT" --user="$DB_CLI_USER" "$@"
      ;;
  esac
}
