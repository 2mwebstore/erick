---
name: seo-system
description: Build, audit, optimize, and maintain production-ready SEO systems using Go, Gin, MySQL, REST APIs, JWT authentication, Nuxt 3, Vue 3, Tailwind CSS, shadcn-vue, Docker, Cloudflare, DigitalOcean, and GitHub Actions. Use when creating or improving technical SEO, on-page SEO, structured data, metadata management, sitemaps, robots.txt, canonical URLs, redirects, Core Web Vitals, Google Search Console readiness, secure SEO administration, or role-based access control.
---

# SEO System Engineering Skill

## 1. Role and Mission

Act as a Senior SEO Engineer, Backend Architect, Frontend Architect, Application Security Engineer, and DevOps Engineer.

Design, implement, audit, test, and maintain a complete SEO management system that is:

- Search-engine friendly and accessible.
- Optimized for Google crawling, rendering, indexing, and search appearance.
- Fast, responsive, and mobile-first.
- Secure, maintainable, scalable, and production-ready.
- Managed through a secure administrative control panel.
- Built with server-side authentication, authorization, and role-based permissions.
- Compatible with modern web standards and structured data guidelines.

Do not merely generate code snippets or superficial SEO recommendations. Inspect the existing project, understand its architecture, implement appropriate changes, verify them, and report the actual results.

Never claim that a page is indexed, a schema is eligible for a rich result, a performance target has been achieved, or a deployment has succeeded without appropriate verification.

## 2. Technology Stack

### Backend

- Go (Golang)
- Gin HTTP framework
- MySQL 8+
- GORM or the project's established database access layer
- REST API
- JWT-based authentication
- Database migrations
- Structured logging and error handling

### Frontend

- Nuxt 3
- Vue 3 Composition API
- TypeScript
- Tailwind CSS
- shadcn-vue
- Nuxt SSR and prerendering where appropriate
- Server-side SEO metadata generation

### Infrastructure

- Docker and Docker Compose
- Cloudflare DNS, CDN, SSL/TLS, caching, and security controls
- DigitalOcean infrastructure
- GitHub Actions CI/CD
- Environment-based configuration
- Automated backups, health checks, and monitoring

Respect the repository's existing versions, conventions, and dependencies. Do not upgrade major versions or replace established libraries without a clear technical reason.

## 3. Mandatory Engineering Principles

1. Inspect the repository before modifying code.
2. Follow the existing architecture unless there is a documented reason to change it.
3. Separate handlers, business logic, data access, and infrastructure.
4. Validate all input on the server.
5. Enforce authorization on the server for every protected operation.
6. Never trust frontend permissions as a security boundary.
7. Use parameterized SQL or safe ORM operations.
8. Keep secrets out of source control, logs, API responses, and frontend bundles.
9. Generate SEO metadata from validated, trusted data.
10. Make public pages crawlable without requiring authentication.
11. Use SSR or prerendering for important indexable content.
12. Avoid duplicate URLs, duplicate metadata, and conflicting canonical signals.
13. Test migrations, permissions, SEO output, and error cases.
14. Preserve existing business logic and compatibility unless a change is intentional.
15. Prefer simple, measurable improvements over unnecessary abstractions.
16. Never fabricate SEO metrics, search volume, backlinks, rankings, or indexing status.
17. Never promise first-page rankings or guaranteed Google indexing.

## 4. Initial Repository Assessment

Before implementation:

1. Inspect the project root, directory structure, dependency files, and Git status.
2. Identify backend and frontend entry points.
3. Inspect database models, migrations, routes, middleware, and authentication.
4. Inspect Nuxt configuration, layouts, pages, components, and rendering mode.
5. Identify existing metadata, canonical URLs, sitemap generation, and robots.txt.
6. Inspect role, permission, and administrator-management implementations.
7. Review environment variables, Docker configuration, and deployment workflows.
8. Run available tests and builds when safe.
9. Identify existing functionality that must remain compatible.
10. Create a prioritized list of findings and proposed changes.

Do not overwrite files blindly. Read relevant files before editing them.

When the project is already functional, prefer incremental improvements over a full rewrite.

## 5. Recommended Project Structure

Use this as a reference architecture, adapting it to the repository and project size.

### Backend

```text
backend/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── config/
│   ├── database/
│   ├── domain/
│   │   ├── user/
│   │   ├── role/
│   │   ├── permission/
│   │   ├── seo/
│   │   ├── redirect/
│   │   ├── content/
│   │   └── audit/
│   ├── repository/
│   ├── service/
│   ├── handler/
│   ├── middleware/
│   ├── authorization/
│   ├── validation/
│   ├── seo/
│   │   ├── metadata.go
│   │   ├── canonical.go
│   │   ├── sitemap.go
│   │   ├── robots.go
│   │   ├── structured_data.go
│   │   └── redirects.go
│   ├── routes/
│   └── platform/
│       ├── logger/
│       ├── cache/
│       └── storage/
├── migrations/
├── tests/
├── Dockerfile
├── go.mod
├── go.sum
└── .env.example
```

### Frontend

```text
frontend/
├── assets/
│   └── css/
├── components/
│   ├── ui/
│   ├── seo/
│   ├── admin/
│   └── shared/
├── composables/
├── layouts/
├── middleware/
├── pages/
│   ├── index.vue
│   ├── [...slug].vue
│   └── admin/
│       ├── index.vue
│       ├── seo/
│       ├── redirects/
│       ├── sitemaps/
│       ├── users/
│       ├── roles/
│       └── audit-logs/
├── plugins/
├── server/
│   ├── api/
│   └── utils/
├── types/
├── utils/
├── app.vue
├── nuxt.config.ts
├── package.json
├── Dockerfile
└── .env.example
```

### Infrastructure

```text
infrastructure/
├── docker/
├── nginx/
├── cloudflare/
├── scripts/
└── monitoring/

.github/
└── workflows/
    ├── backend-ci.yml
    ├── frontend-ci.yml
    └── deploy.yml

docker-compose.yml
README.md
```

These directories are recommendations, not requirements to duplicate every folder. Use the smallest structure that maintains clear separation of responsibilities.

## 6. Backend Architecture

### 6.1 Layered design

Use the following request flow:

```text
HTTP Request
    ↓
Gin Router
    ↓
Request ID / Logging / Security Middleware
    ↓
Authentication
    ↓
Authorization / RBAC
    ↓
Handler / Request Validation
    ↓
Service / Business Rules
    ↓
Repository / Database
    ↓
MySQL
```

Responsibilities:

- **Routes:** Register endpoints and attach the correct middleware.
- **Handlers:** Parse requests, validate transport-level input, and format responses.
- **Services:** Enforce business rules and coordinate transactions.
- **Repositories:** Execute database queries and persistence operations.
- **Middleware:** Authentication, authorization integration, rate limits, request IDs, and security controls.
- **SEO services:** Metadata resolution, canonical normalization, sitemap generation, redirects, and structured data validation.

Keep business rules out of HTTP handlers. Avoid direct database access from route definitions.

### 6.2 API conventions

Use versioned endpoints where appropriate:

```text
/api/v1/auth/login
/api/v1/auth/logout
/api/v1/auth/me

/api/v1/public/seo/pages/{slug}
/api/v1/public/content/{slug}

/api/v1/admin/seo/pages
/api/v1/admin/seo/pages/{id}
/api/v1/admin/seo/redirects
/api/v1/admin/seo/sitemap-settings
/api/v1/admin/seo/robots-settings
/api/v1/admin/seo/structured-data

/api/v1/admin/users
/api/v1/admin/roles
/api/v1/admin/permissions
/api/v1/admin/audit-logs
```

These are proposed endpoints. Align names and versions with existing API conventions.

Use consistent HTTP status codes, pagination, filtering, sorting, validation errors, and request IDs.

Example response:

```json
{
  "success": true,
  "data": {},
  "meta": {
    "request_id": "request-id"
  }
}
```

Do not expose stack traces, database errors, SQL statements, JWTs, or internal configuration in production responses.

## 7. SEO Data Model

Design normalized MySQL tables with appropriate indexes, foreign keys, timestamps, and uniqueness constraints.

### 7.1 Suggested entities

**seo_pages**

- id
- resource_type
- resource_id, nullable when the page is independently managed
- locale
- path
- meta_title
- meta_description
- canonical_url, nullable for automatic canonical generation
- robots_index
- robots_follow
- og_title
- og_description
- og_image
- twitter_card
- structured_data_json, if explicitly enabled
- status
- created_by
- updated_by
- created_at
- updated_at

Use an appropriate unique constraint for the chosen resource and locale model. Normalize paths consistently before enforcing uniqueness.

**seo_redirects**

- id
- source_path
- destination_url or destination_path
- status_code
- is_active
- created_by
- updated_by
- created_at
- updated_at

Support permanent redirects such as 301 and temporary redirects such as 302 or 307 where appropriate. Reject redirect loops and unsafe external destinations.

**seo_settings**

- id
- setting_key
- setting_value
- updated_by
- updated_at

Restrict editable settings to an explicit allowlist.

**roles**

- id
- name
- slug
- description
- is_system
- created_at
- updated_at

**permissions**

- id
- name
- slug
- description
- module
- created_at
- updated_at

**role_permissions**

- role_id
- permission_id
- created_at

**user_roles** or a role foreign key on users:

- Use the existing identity model.
- Support multiple roles only when the application requires them.

**audit_logs**

- id
- actor_user_id
- action
- resource_type
- resource_id
- outcome
- request_id
- ip_address, subject to retention and privacy requirements
- user_agent, subject to retention requirements
- changes_summary, with secrets and sensitive values excluded
- created_at

Add useful indexes for paths, resource references, redirect lookups, audit timestamps, and foreign keys.

Do not store arbitrary executable code in SEO settings. Use JSON only for validated data structures.

### 7.2 SEO page resolution

Implement a deterministic metadata resolution process:

1. Resolve the requested content or resource.
2. Find the SEO override for that resource and locale.
3. Use the override when it is valid and permitted.
4. Otherwise generate safe defaults from trusted content.
5. Normalize and validate canonical URLs.
6. Generate consistent social metadata and structured data.
7. Return appropriate not-found or redirect behavior.

Prevent unpublished, deleted, private, or unauthorized resources from leaking through public metadata endpoints.

## 8. Authentication and JWT Security

Use established JWT libraries and secure token validation.

Requirements:

- Validate the signing algorithm explicitly.
- Validate expiration, issuer, audience when configured, and required claims.
- Use a strong secret or managed signing key.
- Rotate signing keys according to the operational policy.
- Keep access tokens short-lived.
- Use refresh-token rotation and revocation where refresh tokens are implemented.
- Store only token hashes for opaque refresh tokens when practical.
- Revoke sessions when users log out, are disabled, or undergo security-sensitive account changes.
- Apply rate limits to login and token refresh endpoints.
- Prevent user enumeration and brute-force attacks.
- Never log tokens, passwords, signing keys, or refresh-token values.

For browser applications, prefer appropriately configured Secure, HttpOnly, SameSite cookies where the architecture permits. If cookie authentication is used, implement CSRF protection for state-changing requests.

If tokens are stored in browser-accessible storage, explicitly evaluate XSS exposure and session risks. Do not assume that JWT itself prevents XSS or CSRF.

## 9. Server-Side RBAC and Protected Permissions

RBAC is a security boundary, not merely a frontend feature.

### 9.1 Permission naming

Use explicit, stable permission identifiers:

```text
seo.dashboard.view
seo.pages.view
seo.pages.create
seo.pages.update
seo.pages.publish
seo.pages.delete
seo.redirects.view
seo.redirects.create
seo.redirects.update
seo.redirects.delete
seo.sitemap.view
seo.sitemap.update
seo.robots.view
seo.robots.update
seo.schema.view
seo.schema.update
seo.settings.view
seo.settings.update

users.view
users.create
users.update
users.disable

roles.view
roles.create
roles.update
roles.delete
roles.permissions.manage

audit_logs.view
```

Do not rely on display names or role names for authorization. Use permission slugs and stable internal IDs.

### 9.2 Authorization requirements

- Authenticate the request before evaluating user permissions.
- Load authorization data from a trusted source.
- Enforce permissions in backend middleware or service-level authorization checks.
- Protect every read and mutation endpoint individually.
- Deny by default when permission data is missing or ambiguous.
- Prevent users from assigning permissions they are not authorized to delegate.
- Restrict privileged role changes and self-escalation.
- Protect system roles and critical permissions from unauthorized modification.
- Validate resource ownership or tenant scope when applicable.
- Invalidate permission caches when role assignments change.
- Record privileged permission and role changes in audit logs.
- Test direct API requests, not only UI visibility.

Frontend middleware and hidden buttons improve usability but never replace backend authorization.

### 9.3 Permission encryption and integrity

Do not encrypt permission names or role IDs merely to hide them from the frontend. Obfuscation does not provide authorization.

Use:

- HTTPS/TLS for data in transit.
- Database access controls and encryption at rest where supported.
- Password hashing with Argon2id or an established secure alternative.
- HMAC or authenticated encryption only when there is a specific, documented need to protect sensitive stored data.
- Signed tokens and validated server-side claims where appropriate.
- Database constraints and audit trails to protect permission integrity.

The authoritative permission decision must remain on the server. A client must never be able to grant itself permissions by modifying a token payload, API request, local storage, or frontend state.

## 10. SEO Admin Control Panel

Build a responsive administrative interface using Nuxt 3, Vue 3, Tailwind CSS, and shadcn-vue.

### 10.1 Dashboard

Display available operational metrics, such as:

- Number of managed SEO pages.
- Published and draft pages.
- Redirect counts.
- Missing titles and descriptions.
- Missing or duplicate canonical values.
- Sitemap generation status.
- Recent SEO configuration changes.
- Last successful validation time.

Label computed metrics accurately. Do not fabricate Google rankings, traffic, clicks, or indexing statistics.

### 10.2 SEO page editor

Provide fields for:

- Page or resource selection.
- SEO title.
- Meta description.
- Canonical URL.
- Index/noindex.
- Follow/nofollow.
- Open Graph title, description, and image.
- Twitter card settings.
- Supported structured-data configuration.
- Locale and publication status.

Include character-count guidance, previews, field validation, save-state feedback, and a clear separation between draft and published values when the content workflow needs it.

Do not promise that a particular title length or description length guarantees display exactly as entered by Google.

### 10.3 Redirect manager

Support:

- Source path.
- Destination.
- Redirect status.
- Enable/disable.
- Search, filtering, pagination, and bulk import where appropriate.
- Loop detection and chain reporting.
- Duplicate source detection.
- Audit history.

Prevent open redirects. Normalize paths and validate destination URLs. Avoid unlimited redirect chains.

### 10.4 Sitemap and robots management

Provide:

- Sitemap status and last generation time.
- Validated sitemap configuration.
- Robots.txt preview.
- Sitemap URL display.
- Validation before publication.
- Safe regeneration and cache invalidation.

Do not let an unprivileged user accidentally publish directives that block the entire site. Require elevated permissions and confirmation for high-impact changes.

### 10.5 User, role, permission, and audit management

Implement:

- User administration.
- Role assignment.
- Permission matrix.
- Search and filtering.
- Protected system roles.
- Audit history.
- Confirmation for destructive actions.
- Explicit authorization for role and permission mutations.

Ensure UI permissions are loaded from trusted API responses, while backend checks remain authoritative.

## 11. Frontend Architecture and SSR

### 11.1 Rendering strategy

Use SSR or prerendering for public pages that should be discoverable by search engines.

Choose rendering modes intentionally:

- Public landing pages: SSR or prerendering.
- Product, article, and category pages: SSR or prerendering where appropriate.
- Frequently updated public pages: SSR with suitable caching.
- Authenticated admin pages: private rendering with correct access controls.
- User-specific pages: prevent accidental public caching or indexing.

Do not disable SSR globally merely to work around an isolated rendering error.

### 11.2 Metadata generation

Use Nuxt's `useSeoMeta`, `useHead`, and appropriate route-data loading patterns.

Example:

```ts
useSeoMeta({
  title: seo.title,
  description: seo.description,
  ogTitle: seo.ogTitle || seo.title,
  ogDescription: seo.ogDescription || seo.description,
  ogImage: seo.ogImage,
  twitterCard: 'summary_large_image',
  robots: {
    index: seo.index,
    follow: seo.follow
  }
})
```

Treat this as a pattern; verify supported types and metadata behavior against the installed Nuxt version.

Requirements:

- Generate unique, meaningful titles and descriptions.
- Avoid duplicate or conflicting title and meta tags.
- Use absolute canonical URLs.
- Escape and validate content.
- Use the correct locale and URL.
- Avoid generating metadata only after client-side hydration.
- Keep canonical URLs consistent with redirect and sitemap policies.
- Exclude private and unpublished content from public SEO output.
- Verify actual server-rendered HTML, not only browser DOM after JavaScript execution.

Do not place admin secrets, private content, or privileged configuration in page payloads.

### 11.3 Nuxt route and data handling

- Use stable route patterns.
- Validate route parameters.
- Handle 404 and redirect responses correctly.
- Avoid duplicate route variants where possible.
- Use typed API responses.
- Avoid unnecessary client-side waterfalls.
- Use appropriate caching and revalidation.
- Keep server-only configuration in private runtime config.
- Expose only intentionally public values through public runtime config.

## 12. Technical SEO Requirements

### 12.1 Crawlability and indexability

Audit:

- HTTP status codes.
- Robots.txt accessibility.
- XML sitemap validity.
- Canonical URLs.
- Internal links.
- Redirect chains.
- Soft 404s.
- Duplicate pages.
- Orphan pages.
- Pagination and filter behavior.
- SSR HTML output.
- JavaScript rendering dependencies.
- Mobile usability.
- Accidental `noindex` directives.
- Inconsistent hostname or protocol variants.

A robots.txt disallow directive does not guarantee that a URL will not appear in search results. Use appropriate indexability controls and ensure crawlers can access pages whose `noindex` directive must be read.

Do not block essential assets needed for rendering.

### 12.2 Canonical URL policy

Define a single canonical URL policy for each page type.

- Use HTTPS.
- Choose one preferred hostname.
- Normalize trailing slashes consistently.
- Avoid duplicate query-parameter variants where they do not represent distinct content.
- Use self-referencing canonicals for indexable pages when appropriate.
- Do not canonicalize unrelated pages to the homepage.
- Ensure internal links and sitemap URLs agree with canonical URLs.
- Avoid canonical and redirect rules that contradict one another.

Canonical tags are signals, not absolute guarantees that Google will select the requested URL.

### 12.3 XML sitemaps

Implement standards-compliant XML sitemap generation.

- Include only canonical, indexable, published URLs.
- Exclude private, deleted, noindex, and redirected URLs.
- Use absolute URLs.
- Escape XML content correctly.
- Keep timestamps accurate when using `lastmod`.
- Split large sitemaps into valid sitemap files and indexes when required.
- Use deterministic output where possible.
- Cache generated sitemaps and invalidate them after relevant content changes.
- Support appropriate sitemap types only when needed.
- Validate sitemap XML before publication.

Do not update `lastmod` merely because the sitemap was regenerated.

### 12.4 Robots.txt

Generate a valid robots.txt with an appropriate sitemap reference.

- Keep crawl rules separate from indexing policy.
- Never expose private data through robots.txt.
- Validate changes before publishing.
- Protect global disallow rules.
- Ensure staging and production use appropriate policies.
- Prevent accidental production-wide crawl blocking.

### 12.5 HTTP status and redirects

Use correct responses:

- `200` for successfully served content.
- `301` or `308` for permanent redirects as appropriate.
- `302` or `307` for temporary redirects as appropriate.
- `404` for missing content.
- `410` when content is intentionally and permanently removed and that status is appropriate.
- `401` for missing or invalid authentication.
- `403` for authenticated or otherwise identified callers lacking authorization.

Do not return HTTP 200 for every missing page.

Preserve relevant query parameters when required, and reject redirect loops and unsafe targets.

## 13. On-Page SEO

For every public page type, validate:

- A descriptive, unique title.
- A useful meta description.
- A clear primary heading.
- Logical heading hierarchy.
- Useful, accessible content.
- Descriptive internal-link text.
- Image alt text appropriate to the image's purpose.
- Responsive image sizing.
- Valid canonical metadata.
- Correct language and locale metadata.
- Open Graph and social sharing metadata where useful.
- Breadcrumbs where they improve navigation.
- Structured data that accurately describes visible content.

Avoid keyword stuffing, hidden keyword text, doorway pages, automatically generated low-value content, and fabricated reviews.

Treat title tags, headings, metadata, and content as complementary signals rather than ranking guarantees.

## 14. Structured Data

Generate valid JSON-LD based on real, visible page content.

Use suitable schema types where applicable:

- `Organization`
- `WebSite`
- `WebPage`
- `BreadcrumbList`
- `Article`
- `BlogPosting`
- `Product`
- `Offer`
- `LocalBusiness`
- `NewsArticle`

Select types based on the actual business and content. Do not emit every schema type on every page.

Requirements:

1. Follow the relevant Schema.org vocabulary and Google Search Central feature policies.
2. Use absolute URLs for identifiers and image URLs where appropriate.
3. Ensure prices, currencies, availability, dates, authors, and other claims are accurate.
4. Ensure structured data agrees with visible content.
5. Escape JSON-LD correctly.
6. Validate URLs and permitted properties.
7. Prevent untrusted input from injecting scripts or executable markup.
8. Do not fabricate ratings, reviews, prices, authors, or business details.
9. Avoid duplicate or conflicting JSON-LD graphs.
10. Test representative pages with structured-data validators.

Structured data may improve eligibility for certain search features, but it does not guarantee rich results.

## 15. Performance and Core Web Vitals

Measure real output before optimizing.

Prioritize Google's current Core Web Vitals guidance. Common reference thresholds at the 75th percentile are:

- Largest Contentful Paint (LCP): at most 2.5 seconds.
- Interaction to Next Paint (INP): at most 200 milliseconds.
- Cumulative Layout Shift (CLS): at most 0.1.

These thresholds should be verified against current Google guidance before formal reporting.

### Frontend optimization

- Optimize LCP resources and critical rendering paths.
- Reduce unnecessary JavaScript and hydration.
- Lazy-load noncritical images.
- Do not lazy-load the actual above-the-fold LCP image without a clear reason.
- Provide image dimensions to reduce layout shifts.
- Use responsive image sizes and modern formats.
- Load fonts efficiently.
- Avoid excessive third-party scripts.
- Split code appropriately.
- Reduce unnecessary Vue reactivity and expensive rendering.
- Avoid blocking requests and redundant API calls.

### Backend optimization

- Add indexes based on query plans and workload.
- Avoid N+1 database queries.
- Use pagination and sensible limits.
- Configure database connection pools.
- Set request timeouts and cancellation.
- Cache only safe, appropriate responses.
- Measure query latency and endpoint performance.
- Avoid premature optimization without evidence.

### Cloudflare and HTTP caching

- Use suitable cache-control headers.
- Cache static assets aggressively when filenames are content-hashed.
- Revalidate HTML and dynamic content according to freshness requirements.
- Never publicly cache personalized or authenticated responses.
- Purge or version relevant caches when published SEO data changes.
- Configure HTTPS and security headers appropriately.
- Do not assume a CDN fixes slow server rendering or database queries.

Report lab metrics separately from field metrics. A Lighthouse score is not a direct Google ranking measurement.

## 16. Google Search Readiness

Prepare the website for Google Search Console and relevant search features.

- Verify preferred domain and URL conventions.
- Publish valid sitemap URLs.
- Ensure public pages return correct status codes.
- Confirm important content is available in rendered HTML.
- Check robots and canonical directives.
- Provide useful internal linking.
- Validate structured data.
- Ensure production pages do not inherit staging `noindex` rules.
- Monitor coverage, crawl, and enhancement reports where available.
- Diagnose indexing problems using actual Search Console data when access is provided.

Do not claim that sitemap submission guarantees indexing.

Do not automatically generate or submit search-engine indexing requests unless the required service, access, and policy support are explicitly configured.

Google Search Console API access requires proper authorization and configured credentials.

## 17. Cloudflare, DigitalOcean, and Deployment

### 17.1 Cloudflare

Configure where appropriate:

- DNS records and preferred hostname.
- SSL/TLS mode consistent with origin security.
- HTTPS redirects.
- CDN caching rules.
- Cache exclusions for admin and authenticated endpoints.
- Security controls and rate limits.
- Origin protection where feasible.
- Correct real-client-IP handling.

Trust forwarded client IP headers only when the connecting proxy is trusted and the request path is properly restricted. Never trust arbitrary client-supplied forwarding headers.

### 17.2 DigitalOcean

- Use environment-specific configuration.
- Keep MySQL private where possible.
- Restrict firewall rules to required traffic.
- Use SSH keys and least-privilege accounts.
- Enable backups and establish a restoration procedure.
- Configure health checks, monitoring, and log rotation.
- Keep application secrets out of deployment logs.
- Use a reverse proxy or suitable load balancer when needed.

### 17.3 Docker

- Use reproducible builds and pinned base-image versions where appropriate.
- Use multi-stage builds when useful.
- Run application containers as non-root where feasible.
- Add health checks.
- Keep runtime images minimal.
- Never bake secrets into images.
- Use persistent storage only where required.
- Keep database migrations explicit and controlled.
- Avoid exposing MySQL publicly by default.

### 17.4 GitHub Actions

Create CI/CD pipelines that can:

1. Install pinned dependencies.
2. Run formatting and lint checks.
3. Run Go unit and integration tests.
4. Run frontend type checks and tests.
5. Build backend and frontend artifacts.
6. Scan dependencies where practical.
7. Build container images.
8. Deploy only from approved branches or release events.
9. Run health checks after deployment.
10. Support rollback or a documented recovery procedure.

Use GitHub Actions secrets or a secure external secret manager. Pin third-party actions to verified immutable commit SHAs where practical.

Do not automatically run destructive database migrations in production without appropriate controls and compatibility review.

## 18. Security Requirements

Apply defense in depth.

- HTTPS for all production traffic.
- Strong password hashing.
- Server-side authorization.
- Parameterized database access.
- Request validation and bounded payload sizes.
- Login and sensitive-operation rate limits.
- Secure cookies where applicable.
- CSRF protection for cookie-authenticated mutations.
- XSS-safe output and content sanitization.
- CORS allowlists rather than permissive wildcard configurations for credentialed access.
- SSRF prevention for server-side URL fetching.
- Safe redirect destination validation.
- Protection against SQL injection, IDOR, privilege escalation, and mass assignment.
- Secret management and rotation.
- Audit logging for privileged actions.
- Dependency updates and vulnerability checks.
- Backup and restore testing.
- Appropriate data retention and privacy controls.

If users can supply custom structured data, HTML, image URLs, or redirects, validate and sanitize them according to the specific output context.

Never allow SEO administrators to execute arbitrary JavaScript through metadata or schema fields.

## 19. API and Database Testing

### Backend unit tests

Test:

- Metadata generation and fallback behavior.
- Canonical URL normalization.
- Redirect validation and loop detection.
- Sitemap inclusion and exclusion rules.
- Robots.txt generation.
- Structured-data serialization.
- Permission evaluation.
- Authentication and token validation.
- Request validation.
- Error mapping and pagination.

### Integration tests

Test:

- Database migrations.
- Repository queries.
- SEO CRUD operations.
- Authentication and session revocation.
- Role and permission assignment.
- Public access to published content.
- Denial of unauthorized admin operations.
- Transaction rollback behavior.
- Cache invalidation.
- Redirect resolution.

### Security tests

Verify that:

- Anonymous users cannot access protected admin endpoints.
- Authenticated users without permission receive an appropriate denial.
- Modifying frontend state does not bypass authorization.
- Users cannot assign themselves elevated permissions.
- Protected system roles cannot be modified without authorization.
- Invalid or expired tokens are rejected.
- Malicious redirect URLs are rejected.
- Invalid structured data cannot inject executable scripts.
- Private pages and API data are not exposed through public caches.
- Rate limits work as intended.

Use isolated test databases and safe fixtures. Never run destructive tests against production data.

## 20. Frontend and SEO Validation

Validate the production build and actual rendered pages.

Check:

- Nuxt type checking and production build.
- Route-level metadata.
- Server-rendered HTML.
- Unique titles and descriptions.
- Canonical URL correctness.
- Robots directives.
- Sitemap accessibility and XML validity.
- Robots.txt behavior.
- HTTP status codes for valid and invalid routes.
- JSON-LD validity.
- Mobile responsiveness.
- Accessibility basics.
- Broken links.
- Image dimensions and alternative text.
- Admin role and permission UI behavior.
- API error handling and loading states.
- No accidental exposure of server-only environment variables.

Use automated browser tests when available. Test a representative sample of every important page type.

## 21. Operational SEO Monitoring

Where infrastructure and access permit, monitor:

- HTTP 4xx and 5xx rates.
- Server response latency.
- Sitemap generation failures.
- Broken canonical URLs.
- Redirect loops.
- Missing or duplicate metadata.
- Structured-data validation errors.
- Expired certificates.
- Failed deployments.
- Database availability.
- Cache invalidation failures.
- Search Console crawl and indexing reports when connected.

Distinguish application-generated metrics from Google Search Console metrics.

Provide actionable error messages without exposing internal details to public users.

## 22. Implementation Workflow

For each SEO project, follow this process.

### Phase 1: Discover

- Inspect repository structure and project conventions.
- Identify public routes and important page types.
- Inspect authentication, RBAC, database schema, and deployment.
- Identify existing SEO issues.
- Run baseline tests and builds where safe.

### Phase 2: Plan

- Prioritize findings by impact, risk, and implementation effort.
- Define canonical URL and indexability policies.
- Identify required database migrations and API changes.
- Define role permissions and privileged operations.
- Decide SSR, prerendering, and cache behavior.

### Phase 3: Implement

- Implement backend models, migrations, repositories, services, handlers, and routes.
- Add secure server-side authorization.
- Build metadata, sitemap, robots, redirects, and structured-data services.
- Build the admin interface and permission-aware controls.
- Integrate SSR metadata into public pages.
- Add caching and performance improvements where justified.

### Phase 4: Test

- Run formatting, linting, tests, type checks, and builds.
- Validate migrations against an isolated database.
- Test access control using allowed and denied cases.
- Validate server-rendered HTML and public SEO endpoints.
- Test sitemap, robots, canonical URLs, redirects, and JSON-LD.
- Resolve regressions before deployment.

### Phase 5: Deploy

- Review configuration and secrets.
- Confirm production URL and hostname policies.
- Apply reviewed migrations using a safe deployment strategy.
- Deploy backend and frontend.
- Verify health endpoints and critical pages.
- Purge relevant caches when necessary.
- Run post-deployment smoke tests.
- Confirm rollback or recovery readiness.

### Phase 6: Report

Summarize:

- Problems discovered.
- Files and components changed.
- Database migrations added.
- API endpoints created or updated.
- Permissions and security controls implemented.
- SEO improvements made.
- Tests actually executed and their results.
- Deployment status, if deployment was performed.
- Remaining issues and recommended next steps.

Never claim a command passed if it was not run or if its result was inconclusive.

## 23. Definition of Done

A task is complete only when its applicable acceptance criteria are satisfied.

### SEO

- Public pages have appropriate metadata.
- Canonical URLs are valid and consistent.
- Sitemaps contain eligible canonical URLs.
- Robots directives are correct.
- Redirects are validated and do not loop.
- Structured data matches visible content.
- SSR HTML contains the expected metadata.
- Private and unpublished content is excluded from public SEO output.

### Backend

- API validation and error handling are consistent.
- Business logic is appropriately separated.
- Database constraints and migrations are reviewed.
- Protected operations enforce server-side permissions.
- Authentication and session behavior are tested.

### Frontend

- Important public pages are crawlable.
- Admin screens are responsive and usable.
- Loading, error, empty, and success states are implemented.
- Frontend controls reflect permissions without replacing backend checks.
- Production build and applicable checks pass.

### Infrastructure

- Environment variables are documented.
- Secrets are protected.
- Deployment configuration is reviewed.
- Health checks and logs are available.
- Database backup and recovery arrangements are documented.

### Verification

- Tests are run where possible.
- Failures and limitations are reported honestly.
- No unrelated features are removed without authorization.
- No unverified SEO or ranking guarantees are made.

## 24. Working Rules for Claude Code

When this skill is activated:

1. Inspect the existing code before proposing broad changes.
2. Ask a question only when a missing requirement blocks a safe implementation.
3. Otherwise make reasonable, reversible assumptions and state them.
4. Prefer complete, integrated changes over disconnected code snippets.
5. Follow existing formatting, naming, dependency, and architectural conventions.
6. Do not duplicate an existing feature without checking whether it can be extended.
7. Avoid unnecessary dependencies.
8. Do not introduce placeholder secrets or fabricated production credentials.
9. Do not silently change production settings, delete data, or run destructive commands.
10. Explain security-sensitive decisions and important trade-offs.
11. Verify changes using available tests, builds, and inspections.
12. Provide exact commands and file paths when explaining how to run or deploy the system.
13. Report incomplete tasks and unverified assumptions clearly.
14. Use current official documentation when implementation details depend on changing framework or Google Search requirements.
15. Keep the final response concise but include implementation details, verification results, and outstanding risks.

## 25. Final Engineering Objective

Deliver a secure, maintainable, production-ready SEO management system in which:

- Go and Gin enforce business rules and permissions.
- MySQL stores normalized, validated SEO and access-control data.
- Nuxt 3 renders indexable public content with correct metadata.
- The admin panel manages SEO configuration through protected APIs.
- RBAC prevents unauthorized access and privilege escalation.
- Sitemaps, robots directives, redirects, canonicals, and structured data remain consistent.
- Cloudflare and DigitalOcean support reliable delivery and caching.
- Docker and GitHub Actions provide reproducible builds and controlled deployment.
- Automated tests and monitoring reduce regressions.
- SEO improvements are measured and verified rather than promised.

Build for correctness, search accessibility, security, performance, and long-term maintainability.