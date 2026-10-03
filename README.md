# Portfolio — Kong Chansila

Personal portfolio for a Full-Stack Software Developer: a Nuxt 4 frontend, a Go
API, MySQL, and an admin panel at `/admin` that manages every piece of content on
the site.

```text
Cloudflare → Nginx → Nuxt 4 (Node) → Go API → MySQL 8.4
```

## Quick start

```bash
# Frontend
cd frontend
npm install
npm run dev                     # http://localhost:3000

# Backend (optional — only the contact form needs it)
cd backend
cp .env.example .env            # loaded automatically; set DB_ENABLED=false to skip MySQL
go run ./cmd/api                # http://localhost:8080/health

# The admin panel is at http://localhost:3000/admin — see "Signing in" below.
# Keep the password in a password manager, not in this file.

# Using a MySQL you already run locally rather than Docker? Create the database,
# point backend/.env at it, then apply the schema:
./scripts/migrate.sh status     # what is there now
./scripts/migrate.sh up         # create contact_messages
```

Or the whole stack as it runs in production:

```bash
cp .env.example .env            # fill in the secrets
docker compose up -d --build
```

## Editing the site

Content lives in MySQL and is edited at `/admin` — projects and case studies,
experience, skills, services, settings, plus the contact inbox and an audit log.

### Signing in

The panel is at [`/admin`](http://localhost:3000/admin). There is no signup
page, by design — accounts are created from the command line:

```bash
cd backend && go run ./cmd/adminctl create -email you@example.com -name "Your Name"
# in Docker:
docker compose exec api adminctl create -email you@example.com -name "Your Name"
```

`adminctl` prompts for the password rather than taking it as a flag, which
would leave it in your shell history and in the process list.

Changing your own password is done in the panel, under **Your account**. It
asks for the current one first, so a borrowed open tab cannot be used to take
the account over.

Locked out instead, or rotating someone else's:

```bash
cd backend && go run ./cmd/adminctl passwd -email you@example.com
go run ./cmd/adminctl list          # which accounts exist, and their roles
```

Changing a password revokes every existing session for that account, so
anything already signed in elsewhere is signed out.

**Passwords are not written down in this repository.** README.md is tracked by
git, and a credential committed once stays in the history — and in every clone
and fork — after it is deleted. Keep the current one in a password manager. The
rules the API enforces: at least 12 characters, and no more than 72 bytes,
because bcrypt silently truncates past that rather than failing.

If the panel shows a banner saying the API is running an older build, restart
the Go API — it is refusing to trust a server older than the page. See
[docs/SECURITY.md](docs/SECURITY.md) for sessions, CSRF and roles.

`frontend/content/*.ts` is no longer the live source. It seeds the database and
acts as the fallback if the content API is unreachable, so the site stays
readable during an outage.

**The site ships with placeholders on purpose.** Nothing about the projects has
been invented. Any string starting with `TODO:` renders as a visible "content
pending" note rather than as prose, so nothing untrue is ever displayed — the
admin Overview page lists every one. See [docs/CONTENT.md](docs/CONTENT.md).

## Commands

```bash
# frontend/
npm run dev              # dev server
npm run build            # production build (prerenders every content route)
npm run preview          # serve the build
npm run lint             # eslint
npm run typecheck        # vue-tsc
npm run test             # vitest
npm run verify           # lint + typecheck + test + build
npm run content:check    # list unfilled content placeholders
npm run og               # regenerate the Open Graph card
node scripts/export-seed.mjs   # regenerate the seed migration from content/

# backend/
go run ./cmd/api
go run ./cmd/adminctl create -email … -name "…"   # first admin account
go run ./cmd/adminctl list
go run ./cmd/adminctl passwd -email …
go test ./...
go test -race ./...
go vet ./...

# repo root
./scripts/backup-db.sh
./scripts/restore-db.sh <archive>
./scripts/migrate.sh up
./scripts/healthcheck.sh https://your-domain.com
```

## Stack

**Frontend** — Nuxt 4, Vue 3, TypeScript, Tailwind CSS 4, Lucide icons via
`@nuxt/icon`, self-hosted fonts via `@nuxt/fonts`, images via `@nuxt/image`.

**Backend** — Go 1.24, standard library plus the MySQL driver and bcrypt. No web
framework: `net/http` with `ServeMux` method patterns covers the whole API.

**Infrastructure** — Docker multi-stage builds, MySQL 8.4, Nginx, Cloudflare.

## How it is put together

- Content is **cached with stale-while-revalidate** and falls back to bundled
  modules if the API is down, so a database outage degrades the site rather than
  taking it offline. Verified: with the API stopped, every route still returns 200
  with real content.
- The admin panel is **client-rendered, `noindex`, never cached**, and every write
  is session-authenticated, CSRF-checked, role-enforced and audited.
- The hero background is a **CSS-only drifting gradient mesh** — no canvas, no
  scroll listener, no JavaScript — that stops moving entirely under
  `prefers-reduced-motion`.
- The **Go API is not exposed to the internet**. The browser posts to the Nuxt
  server, which forwards over the container network — so there is no CORS policy
  and the API cannot be probed directly.
- **Validation exists in both languages** and is tested for parity. The server is
  authoritative; the client copy only saves a round trip.
- **Both containers run as non-root**, with health checks and bounded resources.

## Documentation

| | |
| --- | --- |
| [ARCHITECTURE.md](docs/ARCHITECTURE.md) | How the system fits together, and the CMS trade-off |
| [CONTENT.md](docs/CONTENT.md) | Editing content, filling placeholders, adding projects |
| [DEPLOYMENT.md](docs/DEPLOYMENT.md) | DNS, TLS, Docker, firewall, releases, rollback |
| [SECURITY.md](docs/SECURITY.md) | Controls per layer, secrets, pre-production checklist |
| [BACKUPS.md](docs/BACKUPS.md) | Schedule, off-site copies, restore drill |
| [MONITORING.md](docs/MONITORING.md) | Health endpoints, what to watch, incident triage |
| [TESTING.md](docs/TESTING.md) | Test layout, manual checklist, Lighthouse, CI |

## Git workflow

```text
main        production
develop     integration
feature/*   new work
fix/*       corrections
```

```text
feat: add project case study pages
fix: correct mobile navigation focus order
chore: rebuild images on patched base
docs: document the restore drill
```

## Before going live

`NUXT_PUBLIC_SITE_URL` must be the real origin — it is baked into canonical tags,
Open Graph URLs, `robots.txt` and the sitemap at build time. While it points at
localhost, `robots.txt` disallows all crawling, so a staging deployment cannot be
indexed by accident.

Full checklist: [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md#pre-launch-checklist).
