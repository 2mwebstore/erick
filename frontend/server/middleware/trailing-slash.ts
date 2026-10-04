/** 301s /work/ to /work, keeping the query string. See trailingSlashTarget. */
export default defineEventHandler((event) => {
  if (event.method !== 'GET' && event.method !== 'HEAD') return

  const url = getRequestURL(event)
  const target = trailingSlashTarget(url.pathname)
  if (target) return sendRedirect(event, target + url.search, 301)
})
