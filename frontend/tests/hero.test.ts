// @vitest-environment nuxt
import { describe, expect, it } from 'vitest'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { computed, ref } from 'vue'
import SectionsHero from '~/components/sections/Hero.vue'
import SectionsAbout from '~/components/sections/About.vue'
import { fallbackContent } from '~/content/fallback'
import type { SiteContentPayload } from '~/types'

const settings = ref<Record<string, string>>({})

mockNuxtImport('useSiteContent', () => () => {
  const payload = computed<SiteContentPayload>(() => ({
    ...fallbackContent,
    settings: { ...fallbackContent.settings, ...settings.value },
  }))
  return {
    payload,
    site: computed(() => resolveSite(payload.value.settings)),
    projects: computed(() => []),
    featuredProjects: computed(() => []),
    experience: computed(() => []),
    capabilities: computed(() => []),
    services: computed(() => []),
    principles: computed(() => []),
    pillars: computed(() => []),
    usingFallback: computed(() => false),
  }
})

describe('hero headline', () => {
  it('animates the half that carries the name and leaves the opening solid', async () => {
    settings.value = { headline: "Hi, I'm", headline_tail: 'KONG CHANSILA' }
    const wrapper = await mountSuspended(SectionsHero)

    const h1 = wrapper.get('h1')
    const spans = h1.findAll('span')
    expect(spans).toHaveLength(1)
    expect(spans[0]!.text()).toBe('KONG CHANSILA')
    expect(spans[0]!.classes()).toContain('headline-gradient')
    // The opening words are the heading's own text, and stay solid ink.
    expect(h1.classes()).not.toContain('headline-gradient')
  })

  /**
   * The quieter half is optional in Settings, and the animation is not allowed
   * to vanish with it — it would leave the homepage looking half-finished.
   */
  it('animates the whole line when there is no quieter half', async () => {
    settings.value = { headline: 'One line only', headline_tail: '' }
    const wrapper = await mountSuspended(SectionsHero)

    const h1 = wrapper.get('h1')
    expect(h1.findAll('span')).toHaveLength(0)
    // Nothing left to wrap, so the heading itself takes the gradient.
    expect(h1.classes()).toContain('headline-gradient')
  })

  /**
   * Whatever the paint does, the headline has to stay one readable heading:
   * real text, in one h1, with no part of it hidden from assistive technology.
   */
  it('keeps the headline as text in a single heading', async () => {
    settings.value = { headline: "Hi, I'm", headline_tail: 'KONG CHANSILA' }
    const wrapper = await mountSuspended(SectionsHero)

    const h1 = wrapper.findAll('h1')
    expect(h1).toHaveLength(1)
    expect(h1[0]!.text().replace(/\s+/g, ' ')).toBe("Hi, I'm KONG CHANSILA")
    expect(h1[0]!.find('[aria-hidden="true"]').exists()).toBe(false)
    expect(h1[0]!.find('img').exists()).toBe(false)
  })
})

describe('about heading', () => {
  it('heads the section with the name from Settings', async () => {
    settings.value = { name: 'KONG CHANSILA' }
    const wrapper = await mountSuspended(SectionsAbout)

    expect(wrapper.get('h2').text()).toBe('KONG CHANSILA')
  })

  /**
   * The name is painted with the same gradient in both places it appears, and
   * the heading must not also carry `text-fg`: that is a utility, and it would
   * outrank the transparent colour the gradient needs to show through.
   */
  it('paints the name with the gradient, and nothing else colours it', async () => {
    settings.value = { name: 'KONG CHANSILA' }
    const wrapper = await mountSuspended(SectionsAbout)
    const heading = wrapper.get('h2')

    expect(heading.classes()).toContain('headline-gradient')
    expect(heading.classes()).not.toContain('text-fg')
  })

  /**
   * The name is a translated setting, so the Khmer page must follow the Khmer
   * spelling rather than a slogan in the locale file — which is why the old key
   * is gone from both of them.
   */
  it('follows the name it is given, rather than a fixed string', async () => {
    settings.value = { name: 'គង់ ចន្ទសិលា' }
    const wrapper = await mountSuspended(SectionsAbout)

    expect(wrapper.get('h2').text()).toBe('គង់ ចន្ទសិលា')
    expect(wrapper.text()).not.toContain('Full-stack, end to end')
  })
})
