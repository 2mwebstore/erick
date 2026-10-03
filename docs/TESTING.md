# Testing

```bash
# Frontend — 98 tests
cd frontend && npm run test

# Backend — 109 tests
cd backend && go test ./...
cd backend && go test -race ./...

# Everything the build gates on
cd frontend && npm run verify     # lint + typecheck + test + build
```

## Frontend

Vitest with the Nuxt environment (`@nuxt/test-utils`).

| File | Covers |
| --- | --- |
| `tests/content.test.ts` | Placeholder detection — the rule that stops invented prose shipping |
| `tests/validation.test.ts` | Contact validation, including parity with the Go limits |
| `tests/projects.test.ts` | Content integrity: unique slugs, no placeholder URLs, ordering |
| `tests/seo.test.ts` | URL building, canonical/sitemap consistency, formatting |
| `tests/components.test.ts` | Button, tag, prose, pending marker, project card |
| `tests/form.test.ts` | Contact form: validation, a11y wiring, honeypot, submit, reset |
| `tests/navigation.test.ts` | Header, mobile drawer, footer, theme toggle and persistence |
| `tests/server.test.ts` | Sitemap and robots contracts |
| `tests/settings.test.ts` | Settings resolution, malformed-JSON tolerance, fallback completeness |
| `tests/tech-icons.test.ts` | Icon existence, brand colours, contrast overrides |

Two of these are guard rails rather than unit tests:

- `validation.test.ts` asserts the limit numbers match
  `backend/internal/services/contact.go`. If one side changes, this fails — which
  is the only reliable way to keep two validators in two languages aligned.
- `projects.test.ts` fails if a project URL is a placeholder or is not `https://`,
  so a dead link cannot reach production.

## Backend

| File | Covers |
| --- | --- |
| `tests/contact_service_test.go` | Validation table, rune vs byte counting, trimming, persistence, error classification |
| `tests/api_test.go` | Status codes, field errors, unknown-field rejection, body cap, rate limiting per IP, security headers, request ids, health, routing |
| `tests/config_test.go` | Refusal to start without credentials, overrides, DSN contents |
| `tests/dotenv_test.go` | .env parsing, real-environment precedence, malformed-line rejection |
| `tests/auth_test.go` | Session middleware, role enforcement, CSRF, password policy, hashing |
| `tests/content_service_test.go` | Slug rules, URL rules, list cleaning, settings key allowlist |

No database is needed: `services.ContactRepository` is an interface, and the
tests substitute a stub. `go test -race` passes, which matters because the rate
limiter is shared mutable state across goroutines.

Notable assertions:

- **`TestContactNeverLeaksInternalErrors`** — the 500 response body must contain
  no SQL, package names or error text.
- **`TestRateLimitIsPerClientAddress`** — one address being blocked must not block
  another.
- **`TestHealthIsNotRateLimited`** — an uptime monitor polling every minute must
  never be throttled into a false alarm.
- **`TestNormaliseCountsRunesNotBytes`** — a Khmer message of 10 characters is 30
  bytes; a byte-length check would reject valid input.
- **`TestAdminRoutesAreUnreachableWithoutASession`** — every admin route, by method
  and path, must answer 401 without a session. This is the test that would catch a
  route accidentally registered outside the protected chain.
- **`TestRequireAdminBlocksEditors`** — the API refuses, not just the UI.
- **`TestCSRFRejectsWritesWithoutAMatchingToken`** — covers the missing-cookie,
  missing-header and mismatched cases separately.
- **`TestSaveProjectDeduplicatesAndTrimsTechnologies`** — the table has a unique
  key, so without this a repeated technology would surface as a 500.

## Manual checks before a release

Things automated tests do not cover:

```text
[ ] Every route loads: / /work /work/<each slug> /resume, and a 404
[ ] Every nav item scrolls to its section, and the active item updates
[ ] Mobile drawer opens, closes on navigation, closes on Escape
[ ] Layout at 375px, 768px, 1024px, 1440px — no horizontal scroll
[ ] Dark, light and system themes; preference survives a reload
[ ] Keyboard only: tab through the whole page, skip link works, focus visible
[ ] Contact form: field errors, successful send, message arrives in the database
[ ] Reduced motion enabled — content is visible, nothing animates
[ ] View source: title, canonical, OG tags, JSON-LD present per page
[ ] Social preview renders: Twitter Card Validator, Facebook Sharing Debugger
[ ] Structured data passes validator.schema.org and Google Rich Results
[ ] Lighthouse on the production build, mobile preset
[ ] Print the resume page — it should look like a resume
[ ] Sign in at /admin, edit a project, confirm it appears on the site
[ ] Editor account: Users and Audit log are refused
[ ] Stop the API — every page still renders from the bundled fallback
[ ] Restart the API — edits reappear within the cache window
```

## Lighthouse

Run against the built output, not the dev server — dev ships unminified code and
an HMR client, which makes the numbers meaningless.

```bash
cd frontend
NUXT_PUBLIC_SITE_URL=https://your-domain.com npm run build
node .output/server/index.mjs &
npx lighthouse http://localhost:3000 --preset=desktop --view
npx lighthouse http://localhost:3000 --view          # mobile
```

## CI

```yaml
# .github/workflows/ci.yml
name: ci
on: [push, pull_request]

jobs:
  frontend:
    runs-on: ubuntu-latest
    defaults: { run: { working-directory: frontend } }
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: 22, cache: npm, cache-dependency-path: frontend/package-lock.json }
      - run: npm ci
      - run: npm run lint
      - run: npm run typecheck
      - run: npm run test
      - run: npm run build
        env: { NUXT_PUBLIC_SITE_URL: https://example.com }

  backend:
    runs-on: ubuntu-latest
    defaults: { run: { working-directory: backend } }
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.24', cache-dependency-path: backend/go.sum }
      - run: go vet ./...
      - run: go test -race ./...
```

Add `npm run content:check -- --strict` to the frontend job once the placeholders
are filled in — before that it would fail every build by design.
