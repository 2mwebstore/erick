/**
 * Google Search Console verification codes, from NUXT_PUBLIC_GOOGLE_SITE_VERIFICATION.
 *
 * Accepts what Search Console hands out: the bare code, or the whole
 * <meta name="google-site-verification" content="…"> tag pasted as it is. Several
 * are allowed, separated by commas or spaces, because every owner who verifies
 * gets a code of their own. Anything that is not a code is dropped, so a typo
 * in the variable cannot put arbitrary text into the <head> of every page.
 */
export function siteVerificationTokens(raw: string | undefined): string[] {
  if (!raw) return []

  // Codes inside pasted tags first, then whatever bare codes are left once the
  // tags are cut out — a pasted tag and a typed code can sit side by side.
  const tags = /<meta\b[^>]*>/gi
  const fromTags = (raw.match(tags) ?? []).flatMap((tag) =>
    [...tag.matchAll(/content\s*=\s*["']([^"']*)["']/gi)].map((match) => match[1] ?? ''),
  )
  const bare = raw.replace(tags, ' ').split(/[\s,]+/)

  const tokens = [...fromTags, ...bare]
    .map((token) => token.trim())
    .filter((token) => /^[\w-]{10,100}$/.test(token))
  return [...new Set(tokens)]
}
