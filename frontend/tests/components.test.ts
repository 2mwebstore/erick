// @vitest-environment nuxt
import { describe, expect, it } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import UiButton from '~/components/ui/Button.vue'
import UiPending from '~/components/ui/Pending.vue'
import UiProse from '~/components/ui/Prose.vue'
import UiTag from '~/components/ui/Tag.vue'
import UiPortrait from '~/components/ui/Portrait.vue'
import UiSection from '~/components/ui/Section.vue'
import ProjectsCard from '~/components/projects/Card.vue'
import type { Project } from '~/types'

describe('UiButton', () => {
  it('renders a button element by default', async () => {
    const wrapper = await mountSuspended(UiButton, { slots: { default: () => 'Send' } })
    expect(wrapper.find('button').exists()).toBe(true)
    expect(wrapper.text()).toContain('Send')
  })

  it('renders an anchor when given href, and marks external links safely', async () => {
    const wrapper = await mountSuspended(UiButton, {
      props: { href: 'https://example.com' },
      slots: { default: () => 'Live' },
    })
    const anchor = wrapper.find('a')
    expect(anchor.exists()).toBe(true)
    expect(anchor.attributes('target')).toBe('_blank')
    expect(anchor.attributes('rel')).toBe('noopener noreferrer')
  })

  it('does not open internal hrefs in a new tab', async () => {
    const wrapper = await mountSuspended(UiButton, {
      props: { href: '/resume.pdf' },
      slots: { default: () => 'Resume' },
    })
    expect(wrapper.find('a').attributes('target')).toBeUndefined()
  })

  it('supports a submit type for forms', async () => {
    const wrapper = await mountSuspended(UiButton, {
      props: { type: 'submit' },
      slots: { default: () => 'Send Message' },
    })
    expect(wrapper.find('button').attributes('type')).toBe('submit')
  })
})

describe('UiProse and UiPending', () => {
  it('renders real prose as a paragraph', async () => {
    const wrapper = await mountSuspended(UiProse, { props: { value: 'A real description.' } })
    expect(wrapper.text()).toContain('A real description.')
    expect(wrapper.text()).not.toContain('Content pending')
  })

  it('renders a placeholder as a visible pending marker, never as prose', async () => {
    const wrapper = await mountSuspended(UiProse, {
      props: { value: 'TODO: describe the platform', file: 'content/projects.ts' },
    })
    expect(wrapper.find('[data-pending]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Content pending')
    expect(wrapper.text()).toContain('describe the platform')
    expect(wrapper.text()).not.toContain('TODO:')
  })

  it('renders nothing at all when a field is absent', async () => {
    const wrapper = await mountSuspended(UiProse, { props: { value: undefined } })
    expect(wrapper.text().trim()).toBe('')
  })

  it('shows the file to edit on the pending marker', async () => {
    const wrapper = await mountSuspended(UiPending, {
      props: { value: 'TODO: fill this in', file: 'content/site.ts' },
    })
    expect(wrapper.text()).toContain('content/site.ts')
  })
})

describe('UiTag', () => {
  it('renders the technology label', async () => {
    const wrapper = await mountSuspended(UiTag, { props: { label: 'Nuxt' } })
    expect(wrapper.text()).toBe('Nuxt')
  })
})

const project = (overrides: Partial<Project> = {}): Project => ({
  slug: 'example-system',
  title: 'Example System',
  category: 'Web Application',
  description: 'A one-sentence description of the system.',
  technologies: ['Nuxt', 'Go'],
  featured: true,
  ...overrides,
})

describe('ProjectsCard', () => {
  it('links the title and the case-study call to action to the project route', async () => {
    const wrapper = await mountSuspended(ProjectsCard, { props: { project: project(), index: 0 } })
    const hrefs = wrapper.findAll('a').map((a) => a.attributes('href'))
    expect(hrefs).toContain('/work/example-system')
    expect(wrapper.text()).toContain('View Case Study')
  })

  it('shows the category and every technology tag', async () => {
    const wrapper = await mountSuspended(ProjectsCard, { props: { project: project(), index: 0 } })
    expect(wrapper.text()).toContain('Web Application')
    expect(wrapper.text()).toContain('Nuxt')
    expect(wrapper.text()).toContain('Go')
  })

  it('omits Live Website and GitHub when no URL exists (§14)', async () => {
    const wrapper = await mountSuspended(ProjectsCard, { props: { project: project(), index: 0 } })
    expect(wrapper.text()).not.toContain('Live Website')
    expect(wrapper.text()).not.toContain('GitHub')
  })

  it('shows Live Website and GitHub only when the URLs are real', async () => {
    const wrapper = await mountSuspended(ProjectsCard, {
      props: {
        project: project({ liveUrl: 'https://example.com', githubUrl: 'https://github.com/example/repo' }),
        index: 0,
      },
    })
    expect(wrapper.text()).toContain('Live Website')
    expect(wrapper.text()).toContain('GitHub')
  })

  it('renders a generated thumbnail rather than a broken image when none is set (§11)', async () => {
    const wrapper = await mountSuspended(ProjectsCard, { props: { project: project(), index: 0 } })
    expect(wrapper.find('img').exists()).toBe(false)
    expect(wrapper.find('svg').exists()).toBe(true)
  })

  it('shows a pending marker instead of inventing a description', async () => {
    const wrapper = await mountSuspended(ProjectsCard, {
      props: { project: project({ description: 'TODO: describe this app' }), index: 0 },
    })
    expect(wrapper.find('[data-pending]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Content pending')
  })

  it('alternates the layout so the grid reads as an editorial list', async () => {
    const even = await mountSuspended(ProjectsCard, { props: { project: project(), index: 0 } })
    const odd = await mountSuspended(ProjectsCard, { props: { project: project(), index: 1 } })
    expect(even.html()).not.toContain('lg:order-2')
    expect(odd.html()).toContain('lg:order-2')
  })
})

describe('heading levels (§24)', () => {
  it('defaults to h3, for cards sitting under a section h2', async () => {
    const wrapper = await mountSuspended(ProjectsCard, { props: { project: project(), index: 0 } })
    expect(wrapper.find('h3').exists()).toBe(true)
    expect(wrapper.find('h2').exists()).toBe(false)
  })

  it('uses h2 when the page title is the h1, so no level is skipped', async () => {
    const wrapper = await mountSuspended(ProjectsCard, {
      props: { project: project(), index: 0, headingLevel: 'h2' },
    })
    expect(wrapper.find('h2').exists()).toBe(true)
    expect(wrapper.find('h3').exists()).toBe(false)
  })
})

describe('technology icons', () => {
  it('shows a brand mark alongside a known technology', async () => {
    const wrapper = await mountSuspended(UiTag, { props: { label: 'Nuxt' } })
    expect(wrapper.text()).toContain('Nuxt')
    expect(wrapper.find('svg').exists()).toBe(true)
  })

  it('matches a versioned name to the same icon', async () => {
    const plain = await mountSuspended(UiTag, { props: { label: 'Nuxt' } })
    const versioned = await mountSuspended(UiTag, { props: { label: 'Nuxt 4' } })
    expect(versioned.find('svg').exists()).toBe(plain.find('svg').exists())
  })

  it('renders as plain text when no icon fits', async () => {
    const wrapper = await mountSuspended(UiTag, { props: { label: 'Schema design' } })
    expect(wrapper.text()).toContain('Schema design')
    expect(wrapper.find('svg').exists()).toBe(false)
  })

  it('can be told not to show an icon', async () => {
    const wrapper = await mountSuspended(UiTag, { props: { label: 'Docker', icon: false } })
    expect(wrapper.find('svg').exists()).toBe(false)
  })

  it('omits the tag row entirely for a project with no known stack', async () => {
    const wrapper = await mountSuspended(ProjectsCard, {
      props: { project: project({ technologies: [] }), index: 0 },
    })
    expect(wrapper.findAll('[class*="font-mono"]').some((n) => n.text() === 'Nuxt')).toBe(false)
  })
})

describe('UiPortrait', () => {
  /**
   * Regression: `sizes` was written as `16rem`. Nuxt Image parses that string
   * itself, does not understand `rem`, and emits an <img> with an empty srcset
   * and no src rather than falling back — so the photo silently never appeared,
   * for local files and remote URLs alike. These assert the one thing that
   * matters: the image has somewhere to load from.
   */
  it('renders a remote portrait with a real src', async () => {
    const wrapper = await mountSuspended(UiPortrait, {
      props: { src: 'https://i.imgur.com/eX6Cu0Z.png', alt: 'Kong Chansila' },
    })

    const img = wrapper.get('img')
    expect(img.attributes('src')).toBe('https://i.imgur.com/eX6Cu0Z.png')
    expect(img.attributes('srcset')).toBeTruthy()
    expect(img.attributes('alt')).toBe('Kong Chansila')
  })

  it('never asks for a sizes width in units Nuxt Image cannot parse', async () => {
    const wrapper = await mountSuspended(UiPortrait, {
      props: { src: 'https://i.imgur.com/eX6Cu0Z.png' },
    })
    expect(wrapper.get('img').attributes('sizes') ?? '').not.toMatch(/\d\s*rem/)
  })

  it('shows an editable frame rather than a stock face when unset (§27)', async () => {
    const wrapper = await mountSuspended(UiPortrait, { props: { src: '' } })
    expect(wrapper.find('img').exists()).toBe(false)
    expect(wrapper.text()).toContain('Your photo')
  })

  it('treats a TODO placeholder as no photo', async () => {
    const wrapper = await mountSuspended(UiPortrait, { props: { src: 'TODO: add a portrait' } })
    expect(wrapper.find('img').exists()).toBe(false)
  })
})

/**
 * The section heading's colour.
 *
 * `titleClass` replaces `text-fg` rather than joining it, because both would be
 * utilities of equal specificity and the one Tailwind emitted last would win —
 * which is how the gradient on the About heading ended up painted over with
 * flat ink the first time.
 */
describe('UiSection title colour', () => {
  it('uses the default ink when no treatment is given', async () => {
    const wrapper = await mountSuspended(UiSection, { props: { title: 'A heading' } })
    expect(wrapper.get('h2').classes()).toContain('text-fg')
  })

  it('hands the colour over entirely when one is', async () => {
    const wrapper = await mountSuspended(UiSection, {
      props: { title: 'A heading', titleClass: 'headline-gradient' },
    })
    const heading = wrapper.get('h2')
    expect(heading.classes()).toContain('headline-gradient')
    expect(heading.classes()).not.toContain('text-fg')
  })
})
