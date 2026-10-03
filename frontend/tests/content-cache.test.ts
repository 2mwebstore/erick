// @vitest-environment nuxt
import { describe, expect, it } from 'vitest'
import { isStaleContentKey } from '~/server/utils/content-cache'

/**
 * These are the real keys observed in the cache storage. If a key shape changes
 * upstream, an edit silently stops appearing on the site until it expires, so
 * the shapes are pinned here rather than rediscovered by hand.
 */
describe('isStaleContentKey', () => {
  it('drops the cached /api/content response', () => {
    expect(isStaleContentKey('nitro:handlers:site-content:sitecontent.json')).toBe(true)
  })

  it('drops cached page HTML and payloads from the swr route rules', () => {
    expect(isStaleContentKey('nitro:routes:_:index.il7asoJjJE.json')).toBe(true)
    expect(isStaleContentKey('nitro:routes:_:workbuffetsystem.n2XIAPouvv.json')).toBe(true)
    expect(isStaleContentKey('nitro:routes:_:_payloadjson.b0Yn1PkV8M.json')).toBe(true)
  })

  it("drops Nuxt's own payload cache, which client-side navigation reads", () => {
    expect(isStaleContentKey('nuxt:payload')).toBe(true)
    expect(isStaleContentKey('nuxt:payload:work:bubble-white')).toBe(true)
    // The dev remount surfaces the same entries under a second name.
    expect(isStaleContentKey('nuxt-root:payload')).toBe(true)
  })

  it('leaves caches that have nothing to do with content alone', () => {
    expect(isStaleContentKey('nuxt:icon:lucide_accessibility_25_abc.json')).toBe(false)
    expect(isStaleContentKey('nuxt-root:icon:simpleicons_android_21_abc.json')).toBe(false)
    expect(isStaleContentKey('nitro:functions:something:else')).toBe(false)
  })
})
