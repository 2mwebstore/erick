import type { H3Event } from 'h3'
import { getRequestHeader, getRequestHeaders, getRequestIP } from 'h3'

/**
 * Headers the Go API reads from a proxied request. Nothing else is forwarded.
 *
 * The auth and admin proxies used to copy every incoming header except a few.
 * That passed on what Cloudflare and Railway's edge add on the way in —
 * cf-connecting-ip, cf-ray, cdn-loop, x-forwarded-for — and with the API also
 * behind Cloudflare, Cloudflare took those for a request it had already handled
 * coming round again. It answered 403 "DNS points to prohibited IP" (error
 * 1000) without reaching the API, so signing in failed every time.
 */
const FORWARDED = new Set(['accept', 'content-type', 'cookie', 'user-agent', 'x-csrf-token'])

/**
 * Keeps the allowed headers and sets x-real-ip from the visitor's address.
 * x-real-ip is never taken from the client: the API rate-limits and audits by it.
 */
export function pickUpstreamHeaders(
  incoming: Record<string, string | undefined>,
  visitorIp: string,
): Record<string, string> {
  const headers: Record<string, string> = {}
  for (const [key, value] of Object.entries(incoming)) {
    const name = key.toLowerCase()
    if (value && FORWARDED.has(name)) headers[name] = value
  }
  if (visitorIp) headers['x-real-ip'] = visitorIp
  return headers
}

/** The headers to send the API for this request. */
export function upstreamHeaders(event: H3Event): Record<string, string> {
  const visitorIp =
    getRequestHeader(event, 'cf-connecting-ip') ?? getRequestIP(event, { xForwardedFor: true }) ?? ''
  return pickUpstreamHeaders(getRequestHeaders(event), visitorIp)
}
