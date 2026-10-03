/**
 * Guards every admin route except the login page.
 *
 * This is a convenience, not the security boundary — the Go API rejects any
 * request without a valid session regardless of what the client does.
 */
export default defineNuxtRouteMiddleware(async (to) => {
  const { ensureSession } = useAdmin()
  const user = await ensureSession()

  if (!user) {
    return navigateTo({ path: '/admin/login', query: { next: to.fullPath } })
  }
})
