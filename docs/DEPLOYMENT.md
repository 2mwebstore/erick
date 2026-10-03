# Deployment

Target: a Linux VPS running Docker, behind Cloudflare.

```text
Cloudflare → Nginx → Nuxt (web) → Go API → MySQL
```

## Prerequisites

- Ubuntu 22.04+ with Docker Engine and the Compose plugin
- A domain with DNS on Cloudflare
- Ports 80 and 443 open; nothing else needs to be

## 1. DNS

In Cloudflare, for the apex and `www`:

| Type | Name | Content | Proxy |
| --- | --- | --- | --- |
| A | `@` | server IPv4 | Proxied (orange) |
| AAAA | `@` | server IPv6 | Proxied |
| CNAME | `www` | your-domain.com | Proxied |

Then set **SSL/TLS → Overview → Full (strict)**. "Flexible" leaves the hop
between Cloudflare and your server in plaintext; "Full (strict)" requires a valid
certificate at the origin, which step 3 installs.

Also enable: **Always Use HTTPS**, **Automatic HTTPS Rewrites**, **Brotli**, and
TLS minimum version 1.2.

## 2. Configure

```bash
git clone <your-repo> /srv/portfolio
cd /srv/portfolio
cp .env.example .env
```

Fill in `.env`. Generate each secret separately — never reuse one:

```bash
openssl rand -base64 32   # DB_PASSWORD
openssl rand -base64 32   # DB_ROOT_PASSWORD
```

`NUXT_PUBLIC_SITE_URL` must be the real origin with no trailing slash. It is
baked into prerendered canonical tags, Open Graph URLs, `robots.txt` and the
sitemap **at build time**, so changing it later requires rebuilding the web
image. While it points at `localhost`, `robots.txt` disallows all crawling — that
is intentional, so a staging deployment cannot be indexed by accident.

```bash
chmod 600 .env
```

## 3. Origin certificate

Easiest path is a Cloudflare Origin Certificate (15-year validity, valid only
for Cloudflare-proxied traffic):

1. Cloudflare → SSL/TLS → Origin Server → Create Certificate
2. Save the certificate as `docker/nginx/certs/fullchain.pem`
3. Save the private key as `docker/nginx/certs/privkey.pem`

```bash
chmod 600 docker/nginx/certs/*.pem
```

Prefer Let's Encrypt? Issue with certbot on the host and mount the live
directory into the nginx service instead.

## 4. Start

```bash
docker compose --profile edge up -d --build
```

Order is enforced by health checks: MySQL becomes healthy, then the API connects
and becomes healthy, then the web container starts.

When the API starts, it applies every migration in `backend/migrations/` that
the database has not recorded yet. That is the schema first, then the content
that ships with the repo. Each one is recorded in `schema_migrations` and never
runs twice, so a restart cannot overwrite content edited in `/admin`. Set
`DB_AUTO_MIGRATE=false` to turn this off.

Create the first admin account. Either set `SEED_ADMIN_EMAIL` and
`SEED_ADMIN_PASSWORD` in `.env` before the first start, which creates it only
while there are no accounts, or create it by hand:

```bash
docker compose exec api adminctl create -email you@example.com -name "Your Name"
docker compose exec api adminctl list
```

Then sign in at `https://your-domain.com/admin` and change the password. If
you used the seed, delete `SEED_ADMIN_PASSWORD` afterwards. The API logs a
warning on every start while it is still set.

There is no signup page on purpose: a self-service account form on an admin panel
is a liability. The password is prompted for, never passed as a flag, so it does
not land in shell history or the process list.

Verify:

```bash
docker compose ps
./scripts/healthcheck.sh https://your-domain.com
```

## 5. Firewall

```bash
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow OpenSSH
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

Neither MySQL (3306) nor the Go API (8080) is published to the host — they exist
only on the compose network. Do not open them. To inspect the database, use
`docker compose exec db mysql …` rather than exposing the port.

Optionally restrict 80/443 to [Cloudflare's IP ranges](https://www.cloudflare.com/ips/)
so the origin cannot be reached directly at all.

## 6. Cron

```bash
crontab -e
```

```cron
15 3 * * * cd /srv/portfolio && ./scripts/backup-db.sh 2>&1 | logger -t portfolio-backup
*/5 * * * * cd /srv/portfolio && ./scripts/healthcheck.sh https://your-domain.com >/dev/null
```

## Releasing a change

```bash
cd /srv/portfolio
git pull

# Content or frontend change — the origin must be passed again, since the build
# bakes it into the prerendered HTML.
docker compose up -d --build web

# Backend change — new migrations are applied when the API starts
docker compose up -d --build api
```

`restart: unless-stopped` plus health checks means a failed container is
restarted rather than silently left down.

## Schema changes

Add a numbered file pair to `backend/migrations/` (`0008_….up.sql` and
`0008_….down.sql`) and deploy. The API applies it on start. MySQL cannot roll
back DDL, so if a migration fails partway the API refuses to start and logs
which file failed, rather than serving on a half-changed schema. Take a backup
before deploying one (`./scripts/backup-db.sh`).

The API only ever applies files. To roll one back, or to apply one without a
restart, use the script. It records what it does in the same
`schema_migrations` table:

```bash
./scripts/migrate.sh status     # tables, and which migrations are recorded
./scripts/migrate.sh up
./scripts/migrate.sh down 0008  # roll one back
```

A database migrated by hand before `schema_migrations` existed is recognised the
first time the API starts: it checks which of 0001–0007 are already there and
records them without running them again. Until then, `migrate.sh up` refuses to
run on it.

`migrate.sh`, `backup-db.sh` and `restore-db.sh` work against either MySQL in
Compose or a MySQL installed on the host — they detect which is running (see
`scripts/_db.sh`). Force the choice with `DB_BACKEND=docker` or `DB_BACKEND=local`.

### Railway

Nothing needs to run by hand: the backend service migrates its database when it
deploys. For the first admin account, set `SEED_ADMIN_EMAIL`, `SEED_ADMIN_NAME`
and `SEED_ADMIN_PASSWORD` on the backend service before the first deploy. After
signing in and changing the password, delete `SEED_ADMIN_PASSWORD`.

To run the scripts against the Railway database, for a backup say, note that
variables already in the environment win over both `.env` files. Pass the
**MySQL** service's variables, not the backend's: the backend's `DB_HOST` is a
private address that only resolves inside Railway. With the database's public
networking off (the safer setting), open a private tunnel first. This needs an
SSH key registered once with `railway ssh keys add`.

```bash
# terminal 1 — keep it open
railway connect MySQL --tunnel-only --port 3307

# terminal 2
export DB_HOST=127.0.0.1 DB_PORT=3307
railway run --no-local --service MySQL ./scripts/backup-db.sh
railway run --no-local --service MySQL ./scripts/migrate.sh status
```

With a TCP proxy turned on for the database, skip the tunnel and the `export`.
The scripts use the proxy address that Railway passes in.

Every script prints the host and database it is about to use before it acts.
Check that line. `127.0.0.1:3306, database portfolio_app` is your local
database. The Railway one shows database `railway`.

## Rollback

```bash
git checkout <previous-tag>
docker compose up -d --build
```

If a release involved a schema change, restore the backup taken before it:

```bash
./scripts/restore-db.sh --dry-run backups/portfolio_<stamp>.sql.gz
./scripts/restore-db.sh backups/portfolio_<stamp>.sql.gz
docker compose restart api
```

## Deploying without Docker

The frontend output is a plain Node server:

```bash
cd frontend
NUXT_PUBLIC_SITE_URL=https://your-domain.com npm ci && npm run build
NUXT_API_BASE_URL=http://127.0.0.1:8080 node .output/server/index.mjs
```

The backend is a single static binary:

```bash
cd backend
CGO_ENABLED=0 go build -ldflags="-s -w" -o api ./cmd/api
```

Run both under systemd with `EnvironmentFile=/etc/portfolio.env` and
`Restart=always`. Everything in `.env.example` applies unchanged.

## Pre-launch checklist

- [ ] `NUXT_PUBLIC_SITE_URL` is the production origin
- [ ] `.env` is `chmod 600` and not in git
- [ ] `curl https://your-domain.com/robots.txt` allows crawling
- [ ] `curl https://your-domain.com/sitemap.xml` lists every page
- [ ] Canonical tags in page source point at the real domain
- [ ] `./scripts/healthcheck.sh https://your-domain.com` passes
- [ ] A contact submission arrives and appears in `contact_messages`
- [ ] `./scripts/backup-db.sh` succeeds and `BACKUP_REMOTE` is set
- [ ] `./scripts/restore-db.sh --dry-run <latest>` passes
- [ ] `npm run content:check` — you accept every remaining placeholder
- [ ] Cloudflare SSL mode is Full (strict)
- [ ] `SECURE_COOKIES=true` and `APP_ENV=production`
- [ ] An admin account exists and `/admin` requires it
- [ ] `/admin` returns `x-robots-tag: noindex` and is absent from the sitemap
- [ ] Content edited in `/admin` appears on the site within the cache window
