# Architecture

## Shape of the system

```text
Browser
   │  HTTPS
Cloudflare            DNS, TLS at the edge, CDN, WAF
   │
Nginx                 TLS at the origin, compression, asset caching, edge rate limit
   │
Nuxt 4 (Node)         prerendered pages + one server route
   │  container network only
Go API                validation, rate limiting, business logic
   │
MySQL 8.4             contact_messages
```

## Why it is built this way

**Content lives in MySQL and is edited at `/admin`.** That is a deliberate change
from the original design, where content was typed TypeScript prerendered at build
time. Moving it into the database bought runtime editing and cost the guarantee
that the site could not be affected by a database outage. Two mitigations bring
most of that back:

- **Stale-while-revalidate caching.** Content routes are cached for five minutes
  and refreshed behind the visitor, so all but the first request in each window
  is served without touching MySQL. The content API adds its own 60-second cache
  with a 10-minute stale window on top.
- **A bundled fallback.** `frontend/content/*.ts` still ships in the build. If the
  content API is unreachable, the site renders from those modules instead of
  failing. Pages stay up and readable; only edits made since the last seed are
  missing. This is verified: with the API down, every route still returns 200 with
  real content.

The files therefore have two jobs now: they seed the database (via
`frontend/scripts/export-seed.mjs`, which generates the seed migration) and they
are the outage fallback.

**The Go API is not reachable from the internet.** The browser posts the contact
form to `/api/contact` on the Nuxt server, which forwards it to the Go service
over the container network. Two consequences: there is no CORS policy to get
wrong, and the API cannot be probed directly. The cost is one internal hop on a
route that runs a handful of times a day.

**The Go API is the only writer.** Nothing in the browser talks to MySQL, and
nothing in the frontend issues SQL. The admin panel posts to this origin, the Nuxt
server forwards to the Go API, and the API validates before anything reaches the
database.

**The admin panel is a separate rendering mode.** `/admin/**` is client-rendered
(`ssr: false`), sends `noindex` and `no-store`, and never appears in the sitemap.
It needs no SEO, and rendering it on the server would mean forwarding session
cookies through SSR for no benefit.

## Repository layout

```text
portfolio/
├── frontend/              Nuxt 4 · Vue 3 · TypeScript · Tailwind CSS 4
│   ├── assets/css/        design tokens and base layer
│   ├── components/
│   │   ├── admin/         form input, list editor, panel, save status
│   │   ├── layout/        header, footer
│   │   ├── navigation/    desktop nav, mobile drawer, theme toggle
│   │   ├── sections/      one component per homepage section
│   │   ├── projects/      card, generated thumbnail, case-study block
│   │   └── ui/            button, section, tag, aurora, portrait, architecture
│   ├── composables/       theme, scroll spy, reveal, SEO, content, admin
│   ├── content/           seed + outage fallback (no longer the live source)
│   ├── layouts/           default and admin layouts
│   ├── middleware/        admin route guard
│   ├── pages/
│   │   ├── admin/         login, overview, projects, experience, skills,
│   │   │                  services, messages, settings, users, audit
│   │   ├── work/          index and [slug]
│   │   ├── index.vue      homepage
│   │   └── resume.vue
│   ├── plugins/           loads site content before first render
│   ├── public/            favicon, generated OG image
│   ├── server/
│   │   ├── api/           content, contact, admin and auth proxies
│   │   └── routes/        sitemap.xml, robots.txt, health
│   ├── scripts/           OG image, placeholder check, seed generator
│   ├── tests/             vitest — units, components, form, navigation, CMS
│   ├── types/             content model and CMS payload types
│   └── utils/             placeholders, formatting, validation, settings, icons
│
├── backend/               Go 1.24, standard library only apart from the driver
│   ├── cmd/api/           entrypoint, graceful shutdown
│   ├── cmd/adminctl/      create accounts, list them, reset a password
│   ├── internal/
│   │   ├── config/        environment loading and validation
│   │   ├── database/      MySQL pool
│   │   ├── handlers/      HTTP transport
│   │   ├── middleware/    request id, logging, recover, headers, rate limit
│   │   ├── models/        shared structs
│   │   ├── repositories/  SQL
│   │   ├── routes/        router assembly
│   │   └── services/      validation and business logic
│   ├── migrations/        numbered up/down SQL
│   └── tests/             go test — service, API, config
│
├── docker/nginx/          reverse proxy configuration
├── docs/                  this directory
└── scripts/               backup, restore, migrate, healthcheck
```

The frontend keeps a flat layout (`components/`, `pages/` at the root) rather
than Nuxt 4's nested `app/` directory. That is set explicitly in
`nuxt.config.ts` via `srcDir: '.'`.

## Layering rules

The Go service is layered so each piece has one reason to change:

```text
handlers/       HTTP in, HTTP out. No business rules, no SQL.
services/       Validation and business logic. No knowledge of HTTP or SQL.
repositories/   SQL only. No business rules.
```

`services.ContactRepository` is an interface, which is why the whole contact
path is tested without a database.

## Request paths

**A page view** is a static file. Nginx serves it from the Nuxt output; `/_nuxt/`
assets are immutable and cached for a year.

**A content read:** the Nuxt plugin fetches `/api/content` once per request, which
proxies to the Go API's `GET /v1/content`. That returns settings, projects with
their technologies and case studies, experience, capabilities and services in one
payload — five queries rather than one join, because joining multiplies every
project row by its technology count for no benefit at this size. The result is
cached at the Nitro route, at the proxy, and in the render payload.

**An admin write:** the browser sends a session cookie plus an `X-CSRF-Token`
header to `/api/admin/*` on this origin. The Nuxt server forwards both to the Go
API, which resolves the session, checks CSRF, enforces the role, validates, writes
inside a transaction, and records an audit entry.

**A contact submission:**

1. The browser validates client-side (`frontend/utils/validation.ts`) purely to
   save a round trip.
2. `POST /api/contact` hits the Nuxt server route, which caps the body at 16 kB,
   silently accepts and discards honeypot hits, and forwards the rest with the
   real client IP in `X-Real-IP`.
3. The Go service applies a per-IP fixed-window limit, validates authoritatively,
   and inserts with bound parameters.
4. A `4xx` comes back as per-field errors the form displays. A `5xx` is logged in
   full server-side and reported to the visitor generically.

Both validators enforce the same limits. `frontend/tests/validation.test.ts`
asserts the numbers match `backend/internal/services/contact.go`, so the two
cannot drift silently.

## Content model

`frontend/types/index.ts` defines the shape. Every optional field is optional on
purpose: a case-study section renders only when it holds verified content, and a
string prefixed `TODO:` renders as a visible "content pending" note rather than
invented prose. See [CONTENT.md](CONTENT.md).
