/** §22 — robots.txt generated so the sitemap URL always matches the deployment. */
export default defineEventHandler((event) => {
  const { public: pub } = useRuntimeConfig(event)
  const origin = pub.siteUrl.replace(/\/+$/, '')
  const isProduction = !origin.includes('localhost')

  setHeader(event, 'content-type', 'text/plain; charset=utf-8')
  setHeader(event, 'cache-control', 'public, max-age=3600')

  // A non-production origin must never invite indexing.
  if (!isProduction) {
    return ['User-agent: *', 'Disallow: /', ''].join('\n')
  }

  return [
    'User-agent: *',
    'Allow: /',
    '',
    'Disallow: /api/',
    '',
    `Sitemap: ${origin}/sitemap.xml`,
    '',
  ].join('\n')
})
