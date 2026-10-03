import type { H3Event } from 'h3'
import type { ContactPayload, ContactResponse } from '~/types'

/**
 * Contact endpoint (§20).
 *
 * This route is a server-side proxy, not the business logic. The Go API holds
 * validation, rate limiting and persistence, and it is never exposed to the
 * public internet — the browser only ever talks to this origin, so the API
 * needs no CORS surface at all.
 */

const MAX_BODY_BYTES = 16 * 1024

function clientIp(event: H3Event): string {
  const header = getRequestHeader(event, 'cf-connecting-ip') ?? getRequestHeader(event, 'x-forwarded-for')
  if (header) return header.split(',')[0]!.trim()
  return getRequestIP(event, { xForwardedFor: true }) ?? ''
}

export default defineEventHandler(async (event): Promise<ContactResponse> => {
  const config = useRuntimeConfig(event)

  if (config.public.contactEnabled === false) {
    throw createError({
      statusCode: 503,
      data: { ok: false, message: 'The contact form is currently disabled.' } satisfies ContactResponse,
    })
  }

  const raw = await readRawBody(event, 'utf8')
  if (!raw || Buffer.byteLength(raw, 'utf8') > MAX_BODY_BYTES) {
    throw createError({
      statusCode: 413,
      data: { ok: false, message: 'That message is too large to send.' } satisfies ContactResponse,
    })
  }

  let body: ContactPayload
  try {
    body = JSON.parse(raw) as ContactPayload
  } catch {
    throw createError({
      statusCode: 400,
      data: { ok: false, message: 'Malformed request.' } satisfies ContactResponse,
    })
  }

  // Honeypot: accept and discard silently so bots get no signal to adapt to.
  if (typeof body.company === 'string' && body.company.trim() !== '') {
    return { ok: true, message: 'Thanks — your message has been sent.' }
  }

  const ip = clientIp(event)

  try {
    const response = await $fetch<ContactResponse>('/v1/contact', {
      baseURL: config.apiBaseUrl,
      method: 'POST',
      timeout: Number(config.apiTimeoutMs) || 8000,
      headers: {
        'content-type': 'application/json',
        ...(ip ? { 'x-forwarded-for': ip } : {}),
        'x-real-ip': ip,
        'user-agent': getRequestHeader(event, 'user-agent') ?? 'portfolio-frontend',
      },
      body: {
        name: body.name,
        email: body.email,
        phone: body.phone ?? '',
        subject: body.subject ?? '',
        project_type: body.projectType,
        message: body.message,
      },
    })

    return response
  } catch (error) {
    const status = (error as { status?: number; statusCode?: number }).status
      ?? (error as { statusCode?: number }).statusCode
      ?? 502
    const data = (error as { data?: { message?: string; errors?: Record<string, string> } }).data

    // 4xx from the API is a real validation result and is passed through.
    // Anything else is logged here and reported generically — internals never
    // reach the browser (§34).
    if (status >= 400 && status < 500 && data) {
      throw createError({
        statusCode: status,
        data: {
          ok: false,
          message: data.message ?? 'Please check the form and try again.',
          errors: mapErrorKeys(data.errors),
        } satisfies ContactResponse,
      })
    }

    console.error('[contact] upstream API error', {
      status,
      message: (error as Error)?.message,
    })

    throw createError({
      statusCode: 502,
      data: {
        ok: false,
        message: 'The message could not be delivered right now. Please email me directly.',
      } satisfies ContactResponse,
    })
  }
})

/** The Go API speaks snake_case; the form speaks camelCase. */
function mapErrorKeys(errors?: Record<string, string>): Record<string, string> | undefined {
  if (!errors) return undefined
  const map: Record<string, string> = { project_type: 'projectType' }
  return Object.fromEntries(Object.entries(errors).map(([k, v]) => [map[k] ?? k, v]))
}
