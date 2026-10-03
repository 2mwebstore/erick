// @vitest-environment nuxt
import { describe, expect, it } from 'vitest'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import UiProfileLinks from '~/components/ui/ProfileLinks.vue'
import SectionsExperience from '~/components/sections/Experience.vue'
import { fallbackContent } from '~/content/fallback'
import type { Profile, TimelineEntry } from '~/types'
import { computed, ref } from 'vue'

const entries = ref<TimelineEntry[]>([])

mockNuxtImport('useSiteContent', () => () => ({
  payload: ref(fallbackContent),
  site: computed(() => resolveSite(fallbackContent.settings)),
  projects: computed(() => []),
  featuredProjects: computed(() => []),
  experience: computed(() => entries.value),
  capabilities: computed(() => []),
  services: computed(() => []),
  usingFallback: computed(() => false),
}))

describe('UiProfileLinks', () => {
  const mount = (profiles: Profile[]) => mountSuspended(UiProfileLinks, { props: { profiles } })

  it('renders nothing when there are no profiles (§27)', async () => {
    const wrapper = await mount([])
    expect(wrapper.find('ul').exists()).toBe(false)
  })

  it('shows the logo image and keeps the label for screen readers', async () => {
    const wrapper = await mount([
      { label: 'GitHub', href: 'https://github.com/example', image: '/logos/github.svg' },
    ])

    const img = wrapper.get('img')
    expect(img.attributes('src')).toBe('/logos/github.svg')
    // Decorative: the accessible name comes from the visually hidden label, so
    // the image must not repeat it.
    expect(img.attributes('alt')).toBe('')
    expect(wrapper.text()).toContain('GitHub')
  })

  it('opens external profiles safely', async () => {
    const wrapper = await mount([
      { label: 'LinkedIn', href: 'https://linkedin.com/in/example', image: '/l.svg' },
    ])
    const link = wrapper.get('a')
    expect(link.attributes('target')).toBe('_blank')
    expect(link.attributes('rel')).toContain('noopener')
  })

  /**
   * Profiles saved before logos were images still carry an Iconify name. They
   * have to keep rendering, or an upgrade silently blanks the row.
   */
  it('falls back to a stored icon name, then to an initial', async () => {
    const withIcon = await mount([
      { label: 'Telegram', href: 'https://t.me/example', icon: 'lucide:send' },
    ])
    expect(withIcon.find('img').exists()).toBe(false)
    expect(withIcon.text()).toContain('Telegram')

    const bare = await mount([{ label: 'Website', href: 'https://example.com' }])
    expect(bare.find('img').exists()).toBe(false)
    expect(bare.text()).toContain('W')
  })
})

describe('SectionsExperience', () => {
  it('renders nothing at all when there is no experience to show', async () => {
    entries.value = []
    const wrapper = await mountSuspended(SectionsExperience)
    expect(wrapper.find('section').exists()).toBe(false)
  })

  it('shows employer, period and place when an entry is a job', async () => {
    entries.value = [
      {
        id: 1,
        label: '',
        title: 'IT Manager & Full-Stack Developer',
        company: 'TANADA Property Co.,Ltd.',
        period: '2026 — Present',
        location: 'Phnom Penh, Cambodia',
        current: true,
        technologies: ['Laravel', 'Nuxt'],
        description: 'Leading development of scalable web applications.',
      },
    ]
    const wrapper = await mountSuspended(SectionsExperience)

    expect(wrapper.text()).toContain('IT Manager & Full-Stack Developer')
    expect(wrapper.text()).toContain('TANADA Property Co.,Ltd.')
    expect(wrapper.text()).toContain('2026 — Present')
    expect(wrapper.text()).toContain('Phnom Penh, Cambodia')
    expect(wrapper.text()).toContain('Laravel')
    expect(wrapper.text()).toContain('Current')
  })

  /**
   * The same list held a capability progression before it held jobs. Entries
   * with no employer still have to render, or upgrading blanks the section.
   */
  it('falls back to the label when an entry names no employer', async () => {
    entries.value = [
      { id: 2, label: 'Backend', title: 'Backend & API development', description: 'APIs.' },
    ]
    const wrapper = await mountSuspended(SectionsExperience)

    expect(wrapper.text()).toContain('Backend & API development')
    expect(wrapper.text()).toContain('Backend')
    expect(wrapper.text()).not.toContain('Current')
  })

  it('marks only the current role, and only when it says so', async () => {
    entries.value = [
      { id: 3, label: '', title: 'Older role', company: 'A', current: false },
      { id: 4, label: '', title: 'Newer role', company: 'B', current: true },
    ]
    const wrapper = await mountSuspended(SectionsExperience)
    expect(wrapper.findAll('article').length).toBe(2)
    expect(wrapper.text().match(/Current/g) ?? []).toHaveLength(1)
  })
})
