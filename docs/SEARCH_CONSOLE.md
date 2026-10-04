# Google Search Console

How to connect the site to Google Search Console and submit its sitemap.

## 1. Prove you own the site

Pick one of the two property types.

### Domain property (recommended)

Covers every version of the site at once: `https://kongchansila.com`, `/km`,
`www.` if you add it later, and `http://`. Nothing in the code is needed.

1. In Search Console, **Add property** → **Domain** → `kongchansila.com`.
2. Copy the `TXT` record it shows (`google-site-verification=…`).
3. In Cloudflare → **DNS** → **Add record**: type `TXT`, name `@`, content the
   value you copied.
4. Back in Search Console, click **Verify**. DNS can take a few minutes.

Keep the TXT record. Deleting it removes your access to the property.

### URL-prefix property (HTML tag)

Covers only `https://kongchansila.com/` and the pages under it.

1. In Search Console, **Add property** → **URL prefix** →
   `https://kongchansila.com/`, then choose **HTML tag**.
2. Copy the code from the tag it shows. Pasting the whole
   `<meta name="google-site-verification" content="…">` tag also works.
3. On the frontend service in Railway, set
   `NUXT_PUBLIC_GOOGLE_SITE_VERIFICATION` to it. Saving restarts the service.
   No rebuild is needed.
4. Check that the tag is on the page, then click **Verify**:
   ```bash
   curl -s https://kongchansila.com/ | grep -o '<meta name="google-site-verification"[^>]*>'
   ```

Keep the variable set. Removing it removes your access to the property. A
second owner can add their own code to the same variable, separated by a comma.

## 2. Submit the sitemap

In the property, open **Sitemaps**, enter `sitemap.xml`, and click **Submit**.

`https://kongchansila.com/sitemap.xml` lists every page in both languages:
home, work, résumé and each published project. Each entry names its English,
Khmer and default (`x-default`) versions, so Google treats them as
translations, not duplicates. `robots.txt` points to it too, so Google finds it
even before you submit it.

`lastmod` is only set where the content records a real edit date: each project
page has its own, and `/work` has the newest of them. Home and résumé have none,
because a wrong date teaches Google to ignore `lastmod` everywhere.

Draft (unpublished) projects are never listed.

## 3. What to expect

- **Status: Success** with about 18 discovered URLs: 3 pages plus 6 projects,
  times 2 languages. The number changes as projects are published.
- Indexing takes days to weeks. Submitting a sitemap does not guarantee that
  every page is indexed.
- Every page declares itself as its own canonical, including the Khmer ones
  (`/km/work` → `/km/work`). If **Pages** reports *Alternate page with proper
  canonical tag* for a `/km` URL, the page is being served by an old deploy.
- `/admin` is never listed and sends `noindex`. Missing pages answer `404`.

## Check the site yourself

Search Console reports what Google has already seen, days later. To check the
site now, open **/admin → SEO check** and click **Run check**. It crawls the
public site the way a search engine does: robots.txt, the sitemap, and every
page that can be reached by following links from the home page. It takes a few
seconds, and then reports:

- **Sitelinks readiness:** whether the navigation is crawlable and leads to
  real pages, whether any sitemap page has no link leading to it, how many
  clicks each page is from the home page, which pages have only one link in, and
  whether project pages carry breadcrumbs.
- **Technical:** HTTPS, robots.txt, the sitemap, broken links, links that go
  through a redirect, canonicals and hreflang.
- **On the page:** titles, descriptions, headings, structured data and image
  alt text.

The scores are the site's own diagnostic, not anything Google reports. Google
alone decides whether to show sitelinks. The check shows whether anything stands
in the way.

It only ever crawls the address in `SITE_URL` on the API (or
`NUXT_PUBLIC_SITE_URL` when that is set instead), one run at a time, and each
run is recorded in the audit log. The latest result is kept in memory, so a
restart clears it.

## Things you do not need to do

- **Resubmit after editing content.** Google re-reads the sitemap on its own,
  and `lastmod` tells it which projects changed.
- **Ping Google.** Google retired its sitemap ping URL in 2023, so the site
  does not ping, and nothing should.

## If the sitemap shows *Couldn't fetch*

1. Open `https://kongchansila.com/sitemap.xml`. It must load as XML, not an
   error page.
2. In Cloudflare → **Security** → **Events**, check that Googlebot is not being
   challenged or blocked (Bot Fight Mode, firewall rules).
3. Check that `https://kongchansila.com/robots.txt` still ends with the
   `Sitemap:` line.

The sitemap never fails because of the API: if the content API is down, it is
built from the bundled content instead.
