#!/usr/bin/env bash
#
# Applies SQL migrations to the database.
#
#   ./scripts/migrate.sh up            # apply every *.up.sql in order
#   ./scripts/migrate.sh down 0001     # roll one migration back
#   ./scripts/migrate.sh status        # show which tables exist
#
# Works against MySQL in Docker Compose or a locally installed MySQL — see
# scripts/_db.sh. Compose seeds a NEW database automatically; this script is for
# a database that already exists, including a local development one.
set -Eeuo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

# shellcheck disable=SC1091
source "$ROOT/scripts/_db.sh"
db_load_env "$ROOT"

DIRECTION="${1:-up}"
ONLY="${2:-}"
MIGRATIONS="$ROOT/backend/migrations"

log() { printf '[migrate] %s\n' "$*"; }

log "backend: $(db_backend), database: $DB_NAME"

if [[ "$DIRECTION" == "status" ]]; then
  log "tables currently in $DB_NAME:"
  db_mysql -N -B -e "SHOW TABLES FROM \`$DB_NAME\`;" | sed 's/^/  /' || true
  exit 0
fi

case "$DIRECTION" in
  up)   PATTERN='*.up.sql';   ORDER='sort' ;;
  down) PATTERN='*.down.sql'; ORDER='sort -r' ;; # newest first when rolling back
  *)    echo "usage: $0 [up|down|status] [migration-prefix]" >&2; exit 64 ;;
esac

APPLIED=0
while IFS= read -r file; do
  name="$(basename "$file")"
  [[ -n "$ONLY" && "$name" != "$ONLY"* ]] && continue

  log "applying $name"
  db_mysql "$DB_NAME" < "$file"
  APPLIED=$((APPLIED + 1))
done < <(find "$MIGRATIONS" -name "$PATTERN" -type f | $ORDER)

if (( APPLIED == 0 )); then
  log "no migrations matched${ONLY:+ prefix '$ONLY'}"
  exit 0
fi

log "$APPLIED migration(s) applied"
log "tables now in $DB_NAME:"
db_mysql -N -B -e "SHOW TABLES FROM \`$DB_NAME\`;" | sed 's/^/  /'
