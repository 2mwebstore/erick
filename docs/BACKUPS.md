# Backups

```text
Daily dump → gzip → verify → off-site copy → 30-day retention → restore drill
```

Only `contact_messages` needs backing up. Everything else — content, code,
configuration — is in git.

## Taking a backup

```bash
./scripts/backup-db.sh
```

The script:

1. Dumps with `--single-transaction`, so a submission arriving mid-backup is
   neither blocked nor half-captured.
2. Writes to `*.sql.gz.partial` and renames only on success, so an interrupted
   run never leaves a truncated file that looks valid.
3. Tests the gzip stream, then checks for mysqldump's `Dump completed` marker —
   a dump that failed early can still be valid gzip.
4. Copies off-site when `BACKUP_REMOTE` is set, and warns loudly when it is not.
5. Prunes archives older than `BACKUP_RETENTION_DAYS` (default 30).

## Off-site

A backup on the same disk does not survive the failure it exists for.

```bash
# .env
BACKUP_REMOTE=remote:portfolio-backups        # rclone
# BACKUP_REMOTE=user@host:/backups/portfolio  # rsync over ssh
BACKUP_RETENTION_DAYS=30
```

With rclone, use a remote with server-side encryption, or wrap it in an `rclone
crypt` remote. These dumps contain personal data.

## Schedule

```cron
15 3 * * * cd /srv/portfolio && ./scripts/backup-db.sh 2>&1 | logger -t portfolio-backup
```

Check it is actually running — a silent backup failure is the standard way to
discover you have no backups:

```bash
journalctl -t portfolio-backup --since '2 days ago'
ls -lh backups/ | tail -5
```

## Restoring

```bash
./scripts/restore-db.sh --dry-run backups/portfolio_20260923T031500Z.sql.gz
./scripts/restore-db.sh backups/portfolio_20260923T031500Z.sql.gz
docker compose restart api
```

The script requires you to type the database name to confirm, and snapshots the
current database to `backups/pre-restore_<stamp>.sql.gz` before overwriting
anything — a restore from the wrong archive is recoverable; a restore over an
un-backed-up database is not.

## Restore drill

Quarterly, and after any change to the schema or the backup script. An untested
backup is a hypothesis.

```bash
# 1. Stand up a throwaway database
docker run -d --name restore-test -e MYSQL_ROOT_PASSWORD=test \
  -e MYSQL_DATABASE=portfolio -p 3307:3306 mysql:8.4

# 2. Restore the newest archive into it
gzip -dc "$(ls -1t backups/*.sql.gz | head -1)" \
  | docker exec -i restore-test mysql -uroot -ptest portfolio

# 3. Confirm the data is there
docker exec restore-test mysql -uroot -ptest -e \
  "SELECT COUNT(*), MAX(created_at) FROM portfolio.contact_messages;"

# 4. Clean up
docker rm -f restore-test
```

Record the date of each successful drill. If you cannot remember the last one,
you are due.

## Recovery targets

| | Target | How |
| --- | --- | --- |
| RPO | 24 hours | Daily backup; at most one day of submissions lost |
| RTO | under 1 hour | `docker compose up -d --build` plus one restore |

To narrow the RPO, run the backup more often — the dataset is small enough that
hourly is cheap.
