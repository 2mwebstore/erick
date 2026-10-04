# Editing content

**Content lives in MySQL and is edited at `/admin`.** Sign in, change what you
need, and it appears on the site within five minutes — pages are cached for that
long to keep the database off the critical path.

```text
/admin
├── Overview      counts, and which fields are still pending
├── Projects      Selected Work and every case study
├── Experience    the About timeline
├── Skills        the Technical Capabilities grid
├── Services      What I Build
├── Messages      contact inbox: read, archive, delete, CSV export
├── Settings      identity, positioning, photo, technology lists, profiles
├── Users         accounts and roles (admin only)
└── Audit log     who changed what (admin only)
```

Create the first account on the server — there is no signup page:

```bash
cd backend && go run ./cmd/adminctl create -email you@example.com -name "Your Name"
# or, in Docker:
docker compose exec api adminctl create -email you@example.com -name "Your Name"
```

## What `frontend/content/` is for now

Those files are no longer the live source. They have two remaining jobs:

1. **They seed the database.** `node scripts/export-seed.mjs` regenerates
   `backend/migrations/0003_seed_content.up.sql` from them.
2. **They are the outage fallback.** If the content API is unreachable, the site
   renders from them instead of failing. Edits made in `/admin` are not written
   back, so after a significant content change, regenerate the seed to stop the
   fallback drifting from reality.

Editing them by hand only changes what a fresh database gets seeded with, and what
visitors see during an outage.

| File | What it holds |
| --- | --- |
| `site.ts` | Name, role, tagline, positioning, email, location, profile links, hero stack |
| `navigation.ts` | The five nav items |
| `experience.ts` | About paragraphs and the capability timeline |
| `projects.ts` | Selected Work and every case study |
| `capabilities.ts` | Technical Capabilities grid, WordPress/SEO stack |
| `services.ts` | What I Build, engineering principles, contact project types |

After editing those files by hand: `npm run typecheck && npm run test`, then
`node scripts/export-seed.mjs` to regenerate the seed.

## The no-fabrication rule

The portfolio must never claim something that is not true — no invented clients,
employers, revenue, user counts, awards, certifications, education, repositories,
URLs, metrics or features. An empty section costs less credibility than an
invented one.

So one convention applies, in the database and in the files alike. **A string that
begins with `TODO:` is an unfilled placeholder.** It renders as a visible bordered
note reading "Content pending" plus your own reminder text — never as prose, and
never as something a visitor could mistake for a claim. The admin Overview page
lists every field still pending.

```ts
// Renders as a "Content pending" note:
description: 'TODO: one sentence describing what this application does.',

// Renders as prose:
description: 'A stock and sales system for a chain of three cafés.',

// Renders nothing at all — the whole section disappears:
// description: undefined,
```

Find every placeholder at any time:

```bash
cd frontend
npm run content:check          # lists them, with line numbers
node scripts/content-check.mjs --strict   # exits 1 if any still render (for CI)
```

`RENDERS` means it is visible on the site right now. `inactive` means it is
commented out — a scaffold waiting for you, invisible to visitors.

## Filling in a case study

Open `frontend/content/projects.ts`. Each project has a `caseStudy` object whose
fields map to the sections of `/work/[slug]`, in render order:

```text
Overview · Context · Problem · My Role · Approach · Architecture
Technology · Key Features · Database · API · Security · Deployment
Challenges · Technical Decisions · Outcome
```

Three rules:

1. **Delete what you cannot fill.** Remove the field and the section vanishes.
   A case study with six strong sections reads better than fifteen thin ones.
2. **`features` lists only shipped functionality.** Not planned, not partial.
3. **Add `liveUrl` or `githubUrl` only when the URL resolves.** A dead link is
   worse than no link. `frontend/tests/projects.test.ts` fails the build if a URL
   is a placeholder or is not `https://`.
4. **`technologies` may be empty.** Buffet System ships that way, because its
   stack was not known. An empty list simply hides the tag row and the
   Technology section; a guessed list would be a fabricated claim.

The `portfolio` entry is a worked example — it is this site, so every section in
it is verifiable.

## Adding a project

```ts
{
  slug: 'inventory-system',        // becomes /work/inventory-system
  title: 'Inventory System',
  category: 'Web Application',
  description: 'One sentence, used on the card and as the meta description.',
  technologies: ['Go', 'Vue', 'MySQL'],
  featured: true,                  // show it on the homepage
  order: 6,                        // position within Selected Work
  caseStudy: { overview: '…' },
}
```

The sitemap, the work index, the resume's project list, the "next project" link
and the structured data all pick it up automatically. Nothing else needs editing.

## Images: upload or link

Every image field in `/admin`, meaning the portrait, the profile logos and each
project's image, takes one of two things:

- **An upload.** Click **Upload** and choose a JPEG, PNG, WebP, GIF or AVIF
  file, up to 5 MB. It is stored in Cloudflare R2, and its public address fills
  the field. The button only appears once R2 is set up (below).
- **A link.** Paste any `https://` image address, or a path to a file in
  `frontend/public/` such as `/portrait.jpg`.

Either way the field holds one address, and the site shows the image from it.
The API refuses anything else: a `javascript:` address, `//host`, or a bare
file name. An uploaded file is checked by its content, so an HTML page renamed
`photo.jpg` is refused, and SVG is refused because it can carry script.

Images from a link or from R2 are shown exactly as uploaded: `@nuxt/image` only
resizes and converts files in `public/`. Upload WebP or AVIF at about the size
it is shown (800 px wide is plenty for the portrait).

### When uploaded files are deleted

An uploaded file is deleted from the bucket once nothing shows it any more:

- **A project is deleted:** its image goes with it.
- **An image is replaced or cleared and saved:** the old file goes, for a
  project image, the portrait or a profile logo.
- **An upload is replaced or cleared before it was ever saved,** for example
  when you upload twice in a row: the unsaved file goes at once.

A file is only deleted when all of these are true. Otherwise it is kept:

- **This site uploaded it.** Pasted links, files in `public/`, and anything put
  in the bucket another way are never touched, including another app's files if
  the bucket is shared.
- **Nothing else uses it.** The same file can be the portrait and a project
  image; it stays until neither uses it.
- **The save succeeded.** A save that fails deletes nothing, and a bucket that
  cannot be reached never fails a save; the file is just left behind.

Every deletion is recorded in the audit log, with its reason.

Two cases leave a file behind: an upload you never save (you upload, then leave
the page), and files replaced before this feature existed. Remove those in the
Cloudflare dashboard if you want the space back. On a custom domain, Cloudflare
may keep serving a deleted file from its cache for a while to anyone who has the
exact link; nothing on the site links to it any more.

### Setting up R2

1. Cloudflare → **R2** → **Create bucket**, e.g. `kongchansila-media`.
2. Make it public: bucket → **Settings** → **Custom Domains** → connect a
   subdomain such as `cdn.kongchansila.com`. The **r2.dev** development URL also
   works, but Cloudflare rate-limits it and does not cache it, so use it for
   testing only.
3. R2 → **Manage API tokens** → **Create API token**: permission **Object Read
   & Write**, applied to this bucket only. Copy the Access Key ID and Secret
   Access Key; the secret is shown once.
4. On the API service (Railway: **backend-portfolio** → **Variables**), set:

   | Variable | Value |
   | --- | --- |
   | `R2_ACCOUNT_ID` | Your Cloudflare account ID (R2 overview page) |
   | `R2_ACCESS_KEY_ID` | From step 3 |
   | `R2_SECRET_ACCESS_KEY` | From step 3 |
   | `R2_BUCKET` | `kongchansila-media` |
   | `R2_PUBLIC_URL` | `https://cdn.kongchansila.com` (or the r2.dev URL) |

   Optional: `UPLOAD_MAX_BYTES` (default 5242880, that is 5 MB).

The API logs `image uploads on` when it starts. If only some variables are set,
uploads stay off, and the log names the ones missing; the site keeps working.
Uploads go through the API, so the bucket needs no CORS rules.

## Your photo

The hero has a portrait column. Until you fill it, it shows an editable frame
that says what to do — never a stock face.

Set it in `/admin` → **Settings** → **Photo**: upload the image or paste its
link, then write the alt text.

Use a 4:5 or square crop at 800px wide or more. The frame is 4:5, so a
portrait-orientation crop fills it without letterboxing.

Write `portraitAlt` describing the person, not the file. "Kong Chansila,
Full-Stack Software Developer" is useful; "portrait.jpg" is not.

To drop the photo entirely, set `portrait: null` — the frame only appears
because the value is unset, so it will not ship as a broken image.

## Technology icons

Technology tags carry a brand mark where one fits, rendered monochrome so the
result reads as a labelled tag rather than a logo wall.

The mapping is in `frontend/utils/tech-icons.ts`. Names match
case-insensitively and ignore a trailing version, so `Nuxt`, `Nuxt 4` and `nuxt`
all resolve to the same icon. A technology with no entry — "Schema design",
"Indexing" — renders as plain text, which is the expected case.

To add one, find the slug at [simpleicons.org](https://simpleicons.org) and add
a line:

```ts
'svelte': 'simple-icons:svelte',
```

`frontend/tests/tech-icons.test.ts` fails if an icon name does not exist in the
installed collection, so a typo cannot ship as an invisible empty box.

## Project images

Projects with no `image` render a generated geometric thumbnail derived from the
slug — deliberate, on-brand, and honest about being a placeholder rather than a
stock photo.

To use a real screenshot, open the project in `/admin` and upload it, or paste
its link, in the **Image** field (see "Images: upload or link" above). Then
describe it in **Image alt text**, e.g. "The inventory dashboard showing stock
levels by branch". Always write the alt text — it is a real accessibility
requirement, not metadata.

## Things you will want to change early

- **`siteConfig.email`** — currently your personal address. Swap it for a
  dedicated one if you would rather not publish that.
- **`siteConfig.profiles`** — empty, so no social links render anywhere.
  Uncomment and fill in once you have the URLs.
- **`siteConfig.resumeFile`** — `null`, so the resume page offers "Print / Save
  as PDF" instead of a download. Drop a PDF in `frontend/public/` and set the
  path to add a download button.
- **The OG card** — regenerate after changing your name, role or tagline:
  `npm run og`.
