/**
 * Where a trailing-slash URL should redirect, or null when it should be served.
 *
 * Every page has one URL, without a trailing slash (the home page is "/").
 * Before this, /work/ and /km/ answered 200 as well — a second copy of each
 * page that any stray link could get crawled, competing with the canonical one.
 *
 * The result always starts with exactly one slash: "//evil.example/" must not
 * become the protocol-relative "//evil.example", which a browser would follow
 * off-site. API routes and build assets are left alone.
 */
export function trailingSlashTarget(pathname: string): string | null {
  if (pathname === '/' || !pathname.endsWith('/')) return null
  if (pathname.startsWith('/api/') || pathname.startsWith('/_')) return null
  return `/${pathname.replace(/^\/+|\/+$/g, '')}`
}
