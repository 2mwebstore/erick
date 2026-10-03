#!/usr/bin/env bash
#
# Daily MySQL backup with compression, retention and verification (§32).
#
#   ./scripts/backup-db.sh
#
# Cron, 03:15 daily, logging to syslog:
#   15 3 * * * cd /srv/portfolio && ./scripts/backup-db.sh 2>&1 | logger -t portfolio-backup
#
# Restore with scripts/restore-db.sh. A backup that has never been restored is
# not a backup — see docs/BACKUPS.md for the quarterly drill.
set -Eeuo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

# shellcheck disable=SC1091
source "$ROOT/scripts/_db.sh"
db_load_env "$ROOT"

BACKUP_DIR="${BACKUP_DIR:-$ROOT/backups}"
RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-30}"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
TARGET="$BACKUP_DIR/${DB_NAME}_${STAMP}.sql.gz"

mkdir -p "$BACKUP_DIR"
chmod 700 "$BACKUP_DIR"

log() { printf '[backup] %s %s\n' "$(date -u +%H:%M:%S)" "$*"; }

log "dumping $(db_target) (backend: $(db_backend))"

# --single-transaction keeps the dump consistent without locking the table,
# so a submission arriving mid-backup is neither blocked nor half-captured.
db_mysqldump \
    --single-transaction \
    --quick \
    --routines \
    --triggers \
    --default-character-set=utf8mb4 \
    "$DB_NAME" \
  | gzip -9 > "$TARGET.partial"

# Only promote the file once the dump has completed, so an interrupted run never
# leaves a truncated backup that looks valid.
mv "$TARGET.partial" "$TARGET"
chmod 600 "$TARGET"

log "verifying archive integrity"
gzip -t "$TARGET"

# A dump that failed early can still be valid gzip, so check for the marker
# mysqldump writes as its final line.
if ! gzip -dc "$TARGET" | tail -5 | grep -q 'Dump completed'; then
  log "ERROR: dump is incomplete, removing $TARGET"
  rm -f "$TARGET"
  exit 1
fi

SIZE="$(du -h "$TARGET" | cut -f1)"
log "wrote $TARGET ($SIZE)"

if [[ -n "${BACKUP_REMOTE:-}" ]]; then
  log "copying off-site to $BACKUP_REMOTE"
  # rclone, rsync or aws s3 cp — whichever is installed. Off-site matters: a
  # backup on the same disk does not survive the failure it exists for.
  if command -v rclone >/dev/null 2>&1; then
    rclone copy "$TARGET" "$BACKUP_REMOTE"
  else
    rsync -az "$TARGET" "$BACKUP_REMOTE"
  fi
  log "off-site copy complete"
else
  log "WARNING: BACKUP_REMOTE is unset — this backup exists only on this host"
fi

log "pruning backups older than ${RETENTION_DAYS}d"
find "$BACKUP_DIR" -name "${DB_NAME}_*.sql.gz" -type f -mtime "+$RETENTION_DAYS" -print -delete

log "done ($(find "$BACKUP_DIR" -name "${DB_NAME}_*.sql.gz" | wc -l | tr -d ' ') backups retained)"
