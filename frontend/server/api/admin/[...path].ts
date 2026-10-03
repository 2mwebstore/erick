/**
 * Authenticated admin proxy.
 *
 * Every /api/admin/* request is forwarded to the Go API with the session and CSRF
 * cookies attached. Keeping the API unreachable from the browser means the admin
 * panel needs no CORS policy and no token in JavaScript-accessible storage — the
 * session stays in an HttpOnly cookie that only this server ever reads.
 */

const SKIP_RESPONSE_HEADERS = new Set([
  'content-encoding',
  'content-length',
  'transfer-encoding',
  'connection',
])

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const path = getRouterParam(event, 'path') ?? ''
  const query = getQuery(event)

  // Reads of the admin surface must never be cached by a browser or proxy.
  setHeader(event, 'cache-control', 'no-store')

  const method = event.method
  const mutating = method !== 'GET' && method !== 'HEAD' && method !== 'OPTIONS'
  const body = mutating ? await readRawBody(event, 'utf8') : undefined

  try {
    const response = await $fetch.raw(`/v1/admin/${path}`, {
      baseURL: config.apiBaseUrl,
      method: method as 'GET',
      query,
      body,
      headers: upstreamHeaders(event),
      timeout: Number(config.apiTimeoutMs) || 8000,
      // Non-2xx is handled below rather than thrown, so field errors and 401s
      // reach the client unchanged.
      ignoreResponseError: true,
      responseType: path.endsWith('export') ? 'text' : 'json',
    })

    for (const [key, value] of response.headers.entries()) {
      if (SKIP_RESPONSE_HEADERS.has(key.toLowerCase())) continue
      // set-cookie only ever originates from the auth routes.
      if (key.toLowerCase() === 'set-cookie') {
        appendResponseHeader(event, 'set-cookie', value)
        continue
      }
      setResponseHeader(event, key, value)
    }

    // A save that succeeded upstream has changed what the site should show, so
    // the cached copies go now rather than expiring on their own minutes later.
    // Done here rather than in each admin page: a page that forgot the call
    // would look like it had saved nothing.
    if (mutating && response.status >= 200 && response.status < 300) {
      await invalidateContentCache()
    }

    setResponseStatus(event, response.status)
    return response._data
  } catch (error) {
    console.error('[admin] upstream unreachable', {
      path,
      message: (error as Error)?.message,
    })

    setResponseStatus(event, 502)
    return { ok: false, message: 'The admin API is unreachable. Check that the service is running.' }
  }
})
