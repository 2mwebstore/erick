#!/usr/bin/env bash
#
# Shared database access for the operational scripts.
#
# The stack can run two ways: MySQL inside Docker Compose, or a MySQL the
# developer already runs locally. These helpers pick whichever is actually
# present, so the same scripts work in both cases.
#
# Override the choice with DB_BACKEND=docker|local.
#
# A Railway database is reached the same way, with its variables passed in —
# see db_from_railway for the address:
#
#   railway run --no-local --service MySQL ./scripts/migrate.sh status

# Resolve configuration. Highest priority first:
#
#   1. the environment the script started with — `DB_HOST=… ./scripts/…`, or
#      whatever `railway run` passed in;
#   2. a Railway MySQL service's own variables (MYSQL*, under `railway run`);
#   3. the root .env (deployment);
#   4. backend/.env (local development).
#
# The environment has to win. When the files were sourced over it, `railway run
# ./scripts/migrate.sh up` reported success against the local database named in
# backend/.env, not the Railway one it was pointed at. The root file still wins
# over backend/.env, so on a server a stray backend/.env cannot redirect the
# scripts.
db_load_env() {
  local root="$1"
  local file line key saved
  local preset=""

  db_from_railway

  # Remember every key the files would set that is already set, before either
  # file can replace it. Keys are shell identifiers, so a space-separated list
  # is safe and avoids empty-array errors under `set -u` in macOS's bash 3.2.
  for file in "$root/backend/.env" "$root/.env"; do
    [[ -f "$file" ]] || continue
    while IFS= read -r line || [[ -n "$line" ]]; do
      [[ "$line" =~ ^[[:space:]]*(export[[:space:]]+)?([A-Za-z_][A-Za-z0-9_]*)= ]] || continue
      key="${BASH_REMATCH[2]}"
      [[ -n "${!key:-}" ]] || continue
      preset="$preset $key"
      printf -v "_db_saved_$key" '%s' "${!key}"
    done < "$file"
  done

  for file in "$root/backend/.env" "$root/.env"; do
    [[ -f "$file" ]] || continue
    set -a
    # shellcheck disable=SC1091
    source "$file"
    set +a
  done

  for key in $preset; do
    saved="_db_saved_$key"
    printf -v "$key" '%s' "${!saved}"
    export "$key"
  done

  DB_NAME="${DB_NAME:-portfolio}"
  DB_HOST="${DB_HOST:-127.0.0.1}"
  DB_PORT="${DB_PORT:-3306}"
  DB_SERVICE="${DB_SERVICE:-db}"

  # The backend service's DB_HOST is Railway's private address. It resolves only
  # inside Railway, so from here the client would fail with a bare "Unknown
  # MySQL server host" that does not say why.
  if [[ "$DB_HOST" == *.railway.internal && "$(db_backend)" != docker ]]; then
    echo "error: DB_HOST is $DB_HOST, which only resolves inside Railway." >&2
    echo "       Pass the database's own variables instead:" >&2
    echo "       railway run --no-local --service MySQL $0 …" >&2
    exit 64
  fi
}

# `railway run --service MySQL` passes the database's own variables, whose names
# are not ours. They are mapped onto DB_*, filling only what the caller has not
# set.
#
# The address depends on the service's public networking. With a TCP proxy on,
# the public URL or proxy domain works from anywhere. With it off, the only
# address is private to Railway, so open a tunnel and pass its address — the
# credentials still come from the service:
#
#   railway connect MySQL --tunnel-only --port 3306        # in another terminal
#   DB_HOST=127.0.0.1 DB_PORT=3306 \
#     railway run --no-local --service MySQL ./scripts/migrate.sh status
db_from_railway() {
  # Neither under `railway run` nor given a Railway URL: nothing to map.
  [[ -n "${RAILWAY_PROJECT_ID:-}${MYSQL_PUBLIC_URL:-}" ]] || return 0

  local url="${MYSQL_PUBLIC_URL:-}"

  if [[ -n "$url" ]]; then
    if [[ ! "$url" =~ ^mysql://([^:/@]+):([^@]*)@([^:/]+):([0-9]+)/([^?]+) ]]; then
      echo "error: MYSQL_PUBLIC_URL is not mysql://user:password@host:port/database" >&2
      exit 64
    fi
    DB_USER="${DB_USER:-${BASH_REMATCH[1]}}"
    DB_PASSWORD="${DB_PASSWORD:-${BASH_REMATCH[2]}}"
    DB_HOST="${DB_HOST:-${BASH_REMATCH[3]}}"
    DB_PORT="${DB_PORT:-${BASH_REMATCH[4]}}"
    DB_NAME="${DB_NAME:-${BASH_REMATCH[5]}}"
  else
    DB_USER="${DB_USER:-${MYSQLUSER:-}}"
    DB_PASSWORD="${DB_PASSWORD:-${MYSQLPASSWORD:-}}"
    DB_NAME="${DB_NAME:-${MYSQLDATABASE:-}}"
    DB_HOST="${DB_HOST:-${RAILWAY_TCP_PROXY_DOMAIN:-}}"
    DB_PORT="${DB_PORT:-${RAILWAY_TCP_PROXY_PORT:-}}"
  fi

  # With no address here, the scripts would fall through to backend/.env and act
  # on the local database while looking like a Railway run.
  if [[ -z "$DB_HOST" ]]; then
    echo "error: Railway passed no address this machine can reach — the database's" >&2
    echo "       public networking is off. Open a tunnel in another terminal:" >&2
    echo "         railway connect MySQL --tunnel-only --port 3306" >&2
    echo "       then pass its address:" >&2
    echo "         DB_HOST=127.0.0.1 DB_PORT=3306 railway run --no-local --service MySQL $0 …" >&2
    exit 64
  fi

  # A remote server is reached with the mysql client here, never through a
  # Compose container that happens to be running.
  DB_BACKEND="${DB_BACKEND:-local}"
}

# Where the scripts are about to connect, so every run says so before it acts.
db_target() {
  case "$(db_backend)" in
    docker) echo "compose service $DB_SERVICE, database $DB_NAME" ;;
    *)      echo "$DB_HOST:$DB_PORT, database $DB_NAME" ;;
  esac
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
