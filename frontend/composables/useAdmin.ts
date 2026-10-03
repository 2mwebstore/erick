import type { AdminUser, SiteContentPayload } from '~/types'

interface SessionResponse {
  ok: boolean
  /** API contract the server implements; absent on builds older than 3. */
  contract?: number
  user?: AdminUser
  csrf?: string
  message?: string
}

/**
 * The API contract this panel was built against.
 *
 * A server left running through a deploy answers the current request bodies
 * with "Malformed request.", which names the symptom and nothing else. Reading
 * the server's own number lets the panel say what is actually wrong, once, at
 * the top of the page, instead of failing on whichever form is used first.
 */
export const REQUIRED_API_CONTRACT = 5

export interface ApiError {
  message: string
  errors: Record<string, string>
  status: number
}

/**
 * Admin session and API access.
 *
 * The session itself is an HttpOnly cookie the browser never exposes to
 * JavaScript; what is held here is only the current user for display and the
 * CSRF token, which has to be readable in order to be echoed in a header.
 */
export function useAdmin() {
  const user = useState<AdminUser | null>('admin-user', () => null)
  const csrf = useState<string>('admin-csrf', () => '')
  const ready = useState<boolean>('admin-session-checked', () => false)
  const contract = useState<number>('admin-api-contract', () => 0)

  const isAdmin = computed(() => user.value?.role === 'admin')

  /** True when the API is older than the panel and saves will fail. */
  const staleApi = computed(() => ready.value && !!user.value && contract.value < REQUIRED_API_CONTRACT)

  /** Loads the session once; subsequent calls are cheap. */
  async function ensureSession(force = false): Promise<AdminUser | null> {
    if (ready.value && !force) return user.value

    try {
      const res = await $fetch<SessionResponse>('/api/auth/session')
      user.value = res.user ?? null
      csrf.value = res.csrf ?? ''
      contract.value = res.contract ?? 0
    } catch {
      user.value = null
      csrf.value = ''
      contract.value = 0
    } finally {
      ready.value = true
    }

    return user.value
  }

  async function login(email: string, password: string) {
    const res = await $fetch<SessionResponse>('/api/auth/login', {
      method: 'POST',
      body: { email, password },
    })
    user.value = res.user ?? null
    csrf.value = res.csrf ?? ''
    contract.value = res.contract ?? 0
    ready.value = true
    return res
  }

  async function logout() {
    try {
      await api('/api/auth/logout', { method: 'POST' })
    } finally {
      user.value = null
      csrf.value = ''
      ready.value = true
      await navigateTo('/admin/login')
    }
  }

  /**
   * Wrapper around $fetch that attaches the CSRF token and normalises errors
   * into a shape the forms can render field by field.
   */
  async function api<T>(path: string, options: Record<string, unknown> = {}): Promise<T> {
    try {
      // $fetch narrows its own return type from the path; the caller's generic
      // is the contract that matters here.
      return (await $fetch(path, {
        ...options,
        headers: {
          ...(options.headers as Record<string, string> | undefined),
          ...(csrf.value ? { 'X-CSRF-Token': csrf.value } : {}),
        },
      })) as T
    } catch (error) {
      const err = error as { status?: number; statusCode?: number; data?: { message?: string; errors?: Record<string, string> } }
      const status = err.status ?? err.statusCode ?? 0

      // A 401 means the session lapsed while the tab was open; send them back to
      // the login screen rather than showing a confusing inline error.
      if (status === 401) {
        user.value = null
        csrf.value = ''
        await navigateTo('/admin/login')
      }

      throw {
        status,
        message: err.data?.message ?? 'Something went wrong. Please try again.',
        errors: err.data?.errors ?? {},
      } satisfies ApiError
    }
  }

  /** Full content payload including unpublished projects. */
  function loadContent() {
    return api<SiteContentPayload>('/api/admin/content')
  }

  return { user, csrf, ready, isAdmin, contract, staleApi, ensureSession, login, logout, api, loadContent }
}
