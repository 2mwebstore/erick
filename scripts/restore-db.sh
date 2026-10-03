#!/usr/bin/env bash
#
# Restores a backup produced by scripts/backup-db.sh (§32).
#
#   ./scripts/restore-db.sh backups/portfolio_20260923T031500Z.sql.gz
#   ./scripts/restore-db.sh --dry-run backups/latest.sql.gz
#
# Restoring OVERWRITES the target database. The script refuses to run without
# an explicit confirmation, because the most likely reason you are reading this
# is that something has already gone wrong.
set -Eeuo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

# shellcheck disable=SC1091
source "$ROOT/scripts/_db.sh"
db_load_env "$ROOT"

DRY_RUN=false
if [[ "${1:-}" == "--dry-run" ]]; then
  DRY_RUN=true
  shift
fi

ARCHIVE="${1:-}"

if [[ -z "$ARCHIVE" ]]; then
  echo "usage: $0 [--dry-run] <backup.sql.gz>" >&2
  echo >&2
  echo "available backups:" >&2
  ls -1t "${BACKUP_DIR:-$ROOT/backups}"/*.sql.gz 2>/dev/null | head -10 >&2 || echo "  (none found)" >&2
  exit 64
fi

[[ -f "$ARCHIVE" ]] || { echo "error: $ARCHIVE not found" >&2; exit 66; }

log() { printf '[restore] %s %s\n' "$(date -u +%H:%M:%S)" "$*"; }

log "checking archive"
gzip -t "$ARCHIVE"

TABLES="$(gzip -dc "$ARCHIVE" | grep -c '^CREATE TABLE' || true)"
log "archive contains $TABLES table definition(s)"

if [[ "$DRY_RUN" == true ]]; then
  log "dry run: archive is readable and complete, nothing was written"
  exit 0
fi

echo
echo "This will REPLACE the contents of database '$DB_NAME'."
read -r -p "Type the database name to confirm: " CONFIRM
[[ "$CONFIRM" == "$DB_NAME" ]] || { echo "aborted" >&2; exit 1; }

# Snapshot the current state first: a restore from the wrong archive is
# recoverable, a restore over an un-backed-up database is not.
SAFETY="${BACKUP_DIR:-$ROOT/backups}/pre-restore_$(date -u +%Y%m%dT%H%M%SZ).sql.gz"
log "snapshotting current database to $SAFETY"
mkdir -p "$(dirname "$SAFETY")"
db_mysqldump --single-transaction --quick "$DB_NAME" | gzip -9 > "$SAFETY"

log "restoring $ARCHIVE"
gzip -dc "$ARCHIVE" | db_mysql "$DB_NAME"

log "verifying"
db_mysql -N -B -e \
  "SELECT CONCAT('contact_messages rows: ', COUNT(*)) FROM \`$DB_NAME\`.contact_messages;"

log "restore complete — restart the API so it reconnects: docker compose restart api"
