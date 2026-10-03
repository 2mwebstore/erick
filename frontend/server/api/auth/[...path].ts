import type { H3Event } from 'h3'

/**
 * Auth proxy, kept separate from /api/admin/* because login has to be reachable
 * before a session exists.
 */
const SKIP_REQUEST_HEADERS = new Set(['host', 'connection', 'content-length', 'accept-encoding'])
const SKIP_RESPONSE_HEADERS = new Set([
  'content-encoding',
  'content-length',
  'transfer-encoding',
  'connection',
])

function forwardedHeaders(event: H3Event): Record<string, string> {
  const headers: Record<string, string> = {}
  for (const [key, value] of Object.entries(getRequestHeaders(event))) {
    if (!value || SKIP_REQUEST_HEADERS.has(key.toLowerCase())) continue
    headers[key] = value
  }
  const ip =
    getRequestHeader(event, 'cf-connecting-ip') ?? getRequestIP(event, { xForwardedFor: true }) ?? ''
  if (ip) headers['x-real-ip'] = ip
  return headers
}

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
      headers: forwardedHeaders(event),
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
