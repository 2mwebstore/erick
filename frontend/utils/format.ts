/** Deterministic seed from a slug, so generated project visuals never change between builds. */
export function slugSeed(slug: string): number {
  let h = 0
  for (let i = 0; i < slug.length; i++) h = (h * 31 + slug.charCodeAt(i)) % 9973
  return h
}

export function yearRange(startYear: number): string {
  return `${startYear} — Present`
}

export function yearsOfExperience(startYear: number, now = new Date()): number {
  return Math.max(0, now.getFullYear() - startYear)
}

/**
 * Builds an absolute URL. The root path keeps its trailing slash so canonical
 * tags and sitemap entries are byte-identical.
 */
export function absoluteUrl(path: string, origin: string): string {
  const base = origin.replace(/\/+$/, '')
  const suffix = path.startsWith('/') ? path : `/${path}`
  return `${base}${suffix}`
}

/**
 * The profile links fit for JSON-LD sameAs: absolute http(s) URLs on another
 * site. sameAs names the same person elsewhere — a relative link, or one back
 * to this site, is not that, and placeholders like "/" made the Person node
 * invalid.
 */
export function sameAsUrls(hrefs: string[], origin: string): string[] {
  const own = origin.replace(/\/+$/, '').toLowerCase()
  const urls = hrefs
    .map((href) => href.trim())
    .filter((href) => /^https?:\/\/[^/]+/i.test(href) && !href.toLowerCase().startsWith(own))
  return [...new Set(urls)]
}
