#!/usr/bin/env bash
#
# Applies SQL migrations to the database.
#
#   ./scripts/migrate.sh up            # apply every migration not yet recorded
#   ./scripts/migrate.sh down 0001     # roll one migration back
#   ./scripts/migrate.sh status        # show tables and recorded migrations
#
# The API applies pending migrations itself when it starts, so this is for
# applying one without a restart, or rolling one back. Both record what ran in
# schema_migrations, so neither applies a file the other already has.
#
# Works against MySQL in Docker Compose, a locally installed MySQL, or a hosted
# one through `railway run` — see scripts/_db.sh.
set -Eeuo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

# shellcheck disable=SC1091
source "$ROOT/scripts/_db.sh"
db_load_env "$ROOT"

DIRECTION="${1:-up}"
ONLY="${2:-}"
MIGRATIONS="$ROOT/backend/migrations"

# Same definition as backend/internal/database/migrate.go.
TRACKING='CREATE TABLE IF NOT EXISTS schema_migrations (
    version    VARCHAR(191) NOT NULL,
    applied_at DATETIME     NOT NULL,
    PRIMARY KEY (version)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci'

log() { printf '[migrate] %s\n' "$*"; }
query() { db_mysql -N -B -e "$1" "$DB_NAME"; }
table_exists() {
  [[ "$(query "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = '$1'")" == 1 ]]
}

log "backend: $(db_backend), target: $(db_target)"

# Stop if the server cannot be reached. Carrying on prints an empty table list
# and "no schema_migrations table yet", which reads like an empty database
# rather than a tunnel that is not open.
if ! db_mysql -e 'SELECT 1' >/dev/null; then
  echo "error: cannot connect to $(db_target) — nothing was read or changed." >&2
  exit 69
fi

if [[ "$DIRECTION" == "status" ]]; then
  log "tables currently in $DB_NAME:"
  db_mysql -N -B -e "SHOW TABLES FROM \`$DB_NAME\`;" | sed 's/^/  /' || true
  if table_exists schema_migrations; then
    log "recorded migrations:"
    query "SELECT CONCAT(version, '  ', applied_at) FROM schema_migrations ORDER BY version" | sed 's/^/  /'
  else
    log "no schema_migrations table yet — the API creates it the next time it starts"
  fi
  exit 0
fi

case "$DIRECTION" in
  up)   PATTERN='*.up.sql';   ORDER='sort' ;;
  down) PATTERN='*.down.sql'; ORDER='sort -r' ;; # newest first when rolling back
  *)    echo "usage: $0 [up|down|status] [migration-prefix]" >&2; exit 64 ;;
esac

query "$TRACKING"

# Tables with no record of what made them were migrated by hand, before the
# record existed. Applying everything again would fail on a column added twice,
# or re-run the seed over content edited since. The API works out what is
# there and records it, so it goes first.
if [[ "$DIRECTION" == up && "$(query "SELECT COUNT(*) FROM schema_migrations")" == 0 ]] \
  && table_exists contact_messages; then
  echo "error: $DB_NAME has tables but no record of which migrations made them." >&2
  echo "       Start the API once — it records what is already there — then run this again." >&2
  exit 65
fi

APPLIED=0
while IFS= read -r file; do
  name="$(basename "$file")"
  [[ -n "$ONLY" && "$name" != "$ONLY"* ]] && continue

  if [[ "$DIRECTION" == up ]]; then
    version="${name%.up.sql}"
    [[ "$(query "SELECT COUNT(*) FROM schema_migrations WHERE version = '$version'")" == 0 ]] || continue

    log "applying $name"
    db_mysql "$DB_NAME" < "$file"
    query "INSERT INTO schema_migrations (version, applied_at) VALUES ('$version', UTC_TIMESTAMP())"
  else
    version="${name%.down.sql}"
    log "rolling back $name"
    db_mysql "$DB_NAME" < "$file"
    query "DELETE FROM schema_migrations WHERE version = '$version'"
  fi
  APPLIED=$((APPLIED + 1))
done < <(find "$MIGRATIONS" -name "$PATTERN" -type f | $ORDER)

if (( APPLIED == 0 )); then
  log "nothing to do${ONLY:+ for prefix '$ONLY'}"
  exit 0
fi

log "$APPLIED migration(s) $( [[ "$DIRECTION" == up ]] && echo applied || echo rolled back )"
log "tables now in $DB_NAME:"
db_mysql -N -B -e "SHOW TABLES FROM \`$DB_NAME\`;" | sed 's/^/  /'
