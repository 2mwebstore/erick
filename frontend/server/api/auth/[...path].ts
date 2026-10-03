/**
 * Auth proxy, kept separate from /api/admin/* because login has to be reachable
 * before a session exists.
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

  setHeader(event, 'cache-control', 'no-store')

  const method = event.method
  const body = method === 'GET' || method === 'HEAD' ? undefined : await readRawBody(event, 'utf8')

  try {
    const response = await $fetch.raw(`/v1/auth/${path}`, {
      baseURL: config.apiBaseUrl,
      method: method as 'GET',
      body,
      headers: upstreamHeaders(event),
      timeout: Number(config.apiTimeoutMs) || 8000,
      ignoreResponseError: true,
    })

    for (const [key, value] of response.headers.entries()) {
      if (SKIP_RESPONSE_HEADERS.has(key.toLowerCase())) continue
      if (key.toLowerCase() === 'set-cookie') {
        appendResponseHeader(event, 'set-cookie', value)
        continue
      }
      setResponseHeader(event, key, value)
    }

    setResponseStatus(event, response.status)
    return response._data
  } catch (error) {
    console.error('[auth] upstream unreachable', { path, message: (error as Error)?.message })
    setResponseStatus(event, 502)
    return { ok: false, message: 'The authentication service is unreachable.' }
  }
})
