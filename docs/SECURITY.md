# Security

The attack surface used to be one POST endpoint. Adding the admin panel changed
that: there is now an authenticated, cookie-backed, state-changing API, which is
a materially larger surface than a contact form. What follows covers both.

## What protects what

| Layer | Control |
| --- | --- |
| Cloudflare | TLS at the edge, WAF, DDoS absorption, bot filtering |
| Nginx | TLS at the origin, HSTS, CSP, request limit on `/api/contact`, 1 MB body cap |
| Nuxt server route | Body caps, honeypot handling, forwards real client IP, `no-store` on admin |
| Go API | Sessions, CSRF, role checks, per-IP rate limits, authoritative validation |
| MySQL | Bound parameters only, least-privilege application user, no published port |

## The admin panel

**Sessions are server-side.** The cookie holds a 256-bit random token; the
database stores only its SHA-256. A database leak therefore yields no usable
sessions, and a logout, password change or account deactivation revokes access
immediately — a stateless token could not be withdrawn.

**Passwords are bcrypt at cost 12** (~250 ms per verification): cheap for a
handful of logins, expensive for an offline attacker. Minimum 12 characters, and
anything over 72 bytes is refused rather than silently truncated, which is what
bcrypt would otherwise do.

**Login does not reveal which emails exist.** A wrong password and an unknown
account return the same message, and the unknown-account path still runs a bcrypt
comparison so the response timing matches.

**CSRF is enforced.** This replaces the earlier claim in this document that no
cookie-authenticated state-changing route existed — that stopped being true when
the admin panel was added. Two layers now apply: `SameSite=Strict` on the session
cookie, and a double-submit token that must be echoed in `X-CSRF-Token`. A
cross-site form post can send the cookie but cannot read it to set the header.

**Roles are enforced server-side.** Editors change content; only admins manage
accounts and read the audit log. The frontend hides links an editor cannot use,
but the API is what actually refuses — verified by test.

**Login is rate limited separately** and far more tightly than the contact form,
because this is the endpoint where guessing is the attack.

**Every change is audited**: actor, action, entity, id, IP and timestamp. The
actor email is denormalised so the log still reads correctly after an account is
deleted. An audit write failing is logged loudly but never fails the operation —
turning a logging problem into an outage would be worse.

**Settings keys are a closed set.** The table is key/value so new settings need no
migration, but accepting arbitrary keys would let any editor write unbounded rows.

## Decisions worth knowing

**The Go API has no CORS policy.** It is not reachable from a browser — the Nuxt
server proxies to it over the container network. A CORS policy would only widen
the surface.

**`X-Forwarded-For` is trusted only behind a proxy.** The header is
attacker-controlled on a directly exposed service, and trusting it would let one
client defeat rate limiting by rotating the value. `TRUSTED_PROXY` defaults to
`false` in `backend/.env.example` and is set to `true` only in compose, where
Nginx and the Nuxt server genuinely set it.

**Validation runs server-side and is authoritative.** The client-side copy in
`frontend/utils/validation.ts` exists to save a round trip, nothing more. Both
enforce identical limits, and a test fails if they drift.

**Unknown JSON fields are rejected**, not ignored. A padded payload is a signal,
not something to silently accept.

**Error responses never carry internals.** A `5xx` is logged in full with a
request id and returned to the visitor as one generic sentence. There is a test
asserting the response body contains no SQL, no package names and no error text.

**Logs contain no message bodies and no email addresses.** Operational
visibility should not create a second copy of personal data in the log stream.

## Secrets

Everything comes from the environment. Nothing is committed and nothing is baked
into an image.

```bash
openssl rand -base64 32
chmod 600 .env
```

The process refuses to start if `DB_USER` or `DB_PASSWORD` is missing while the
database is enabled — starting with a guessable default would be worse than not
starting.

Rotating the database password:

```bash
docker compose exec db mysql -u root -p"$DB_ROOT_PASSWORD" \
  -e "ALTER USER 'portfolio_app'@'%' IDENTIFIED BY 'new-password';"
# update DB_PASSWORD in .env, then
docker compose up -d api
```

## Data handling

`contact_messages` holds a name, an email address, a message, and the submitting
IP and user agent. The IP and user agent exist so abuse can be investigated after
the fact.

Keep it deliberate: decide a retention period and enforce it.

```sql
-- e.g. drop submissions older than two years
DELETE FROM contact_messages WHERE created_at < NOW() - INTERVAL 2 YEAR;
```

## Pre-production checklist

```text
[ ] HTTPS enforced, Cloudflare SSL mode = Full (strict)
[ ] HSTS present on responses
[ ] Secrets in .env only, chmod 600, never committed
[ ] Database and API ports not published to the host
[ ] Application database user is not root
[ ] Input validated server-side, length-bounded before storage
[ ] SQL uses bound parameters only — no string-built queries
[ ] XSS: no v-html anywhere, no innerHTML from content
[ ] CSRF: double-submit token enforced on every admin write
[ ] Sessions stored hashed, revoked on logout and password change
[ ] SECURE_COOKIES=true in production (the API refuses to start otherwise)
[ ] Admin roles enforced in the API, not only hidden in the UI
[ ] First admin created with adminctl, not through a signup page
[ ] Audit log reviewed periodically
[ ] CORS: none needed, none configured
[ ] Rate limiting active at both Nginx and the API
[ ] Request bodies capped at every layer
[ ] Error responses free of internal detail
[ ] Logs free of secrets and message contents
[ ] Firewall denies inbound except 22, 80, 443
[ ] Backups running and encrypted at rest off-site
[ ] Restore tested from an actual archive
[ ] Containers run as non-root (both do)
[ ] Base images rebuilt regularly for patches
```

## Verifying it

```bash
# Security headers
curl -sI https://your-domain.com | grep -iE 'strict-transport|content-security|x-frame|x-content-type'

# Rate limiting — the 6th request in an hour from one IP should be 429
for i in $(seq 1 6); do
  curl -s -o /dev/null -w "%{http_code}\n" -X POST https://your-domain.com/api/contact \
    -H 'content-type: application/json' \
    -d '{"name":"Test","email":"t@example.com","projectType":"Other","message":"A test message here."}'
done

# The API must not be reachable from outside
curl --max-time 5 http://your-domain.com:8080/health   # must fail

# Oversized body is rejected
curl -s -o /dev/null -w "%{http_code}\n" -X POST https://your-domain.com/api/contact \
  -H 'content-type: application/json' --data-binary "@/dev/zero" 2>/dev/null
```

## Dependency patching

```bash
cd frontend && npm audit && npx nuxi upgrade
cd backend && go get -u ./... && go mod tidy && go test ./...
docker compose build --pull        # rebase on patched base images
```

http://localhost:3000/admin
membertsd@gmail.com
VerifyPass2026!x
