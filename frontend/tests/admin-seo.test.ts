// @vitest-environment nuxt
import { beforeEach, describe, expect, it } from 'vitest'
import { mountSuspended, registerEndpoint } from '@nuxt/test-utils/runtime'
import AdminSeo from '~/pages/admin/seo.vue'
import type { SeoAuditStatus, SeoReport } from '~/types'

/**
 * The SEO check page. The crawl itself is tested in Go
 * (backend/internal/seoaudit); this covers what the page shows and sends.
 */

const report: SeoReport = {
  origin: 'https://kongchansila.com',
  startedAt: '2026-10-04T10:00:00Z',
  durationMs: 3400,
  score: 94,
  sitelinksScore: 90,
  truncated: false,
  checks: [
    { id: 'navigation', group: 'sitelinks', title: 'The main navigation is crawlable', status: 'warn', detail: 'Almost every link is a jump to a section.' },
    { id: 'orphans', group: 'sitelinks', title: 'Every page is reachable by links', status: 'pass', detail: 'Every sitemap page can be reached.' },
    {
      id: 'broken_links', group: 'technical', title: 'No internal link leads to a missing page', status: 'fail',
      detail: '1 link(s) lead to a page that is missing.', urls: ['https://kongchansila.com/missing ← linked from https://kongchansila.com/'],
    },
    { id: 'titles', group: 'onpage', title: 'Every page has its own title', status: 'pass', detail: 'Every indexable page has a title of its own.' },
  ],
  pages: [
    {
      url: 'https://kongchansila.com/', status: 200, depth: 0, inSitemap: true, indexable: true, title: 'Home', description: 'Home',
      canonical: 'https://kongchansila.com/', h1: 1, imagesWithoutAlt: 0, structuredData: ['WebSite'], incoming: 9, outgoing: 8,
    },
    {
      url: 'https://kongchansila.com/missing', status: 404, depth: 1, inSitemap: false, indexable: false, title: '', description: '',
      canonical: '', h1: 0, imagesWithoutAlt: 0, structuredData: [], incoming: 1, outgoing: 0,
    },
  ],
}

const server: { status: SeoAuditStatus; posts: number } = { status: { configured: true, running: false }, posts: 0 }

registerEndpoint('/api/admin/seo/audit', {
  method: 'GET',
  handler: () => server.status,
})
registerEndpoint('/api/admin/seo/audit', {
  method: 'POST',
  handler: () => {
    server.posts++
    server.status = { ...server.status, running: true }
    return server.status
  },
})

const settle = () => new Promise((resolve) => setTimeout(resolve, 20))

beforeEach(() => {
  server.posts = 0
  useState('admin-user').value = { id: 1, email: 'someone@example.com', name: 'Someone', role: 'editor' }
  useState('admin-csrf').value = 'token'
})

describe('SEO check page', () => {
  it('explains how to turn the check on when the API has no site address', async () => {
    server.status = { configured: false, running: false }
    const wrapper = await mountSuspended(AdminSeo)
    await settle()

    expect(wrapper.text()).toContain('SITE_URL')
    expect(wrapper.find('button').exists()).toBe(false)
  })

  it('shows both scores, labelled as the site\'s own, and every check with its status', async () => {
    server.status = { configured: true, origin: report.origin, running: false, report }
    const wrapper = await mountSuspended(AdminSeo)
    await settle()

    const text = wrapper.text()
    expect(text).toContain('94')
    expect(text).toContain('Sitelinks readiness')
    expect(text).toContain('not data from Google')

    expect(wrapper.find('[data-check="broken_links"]').text()).toContain('Problem')
    expect(wrapper.find('[data-check="navigation"]').text()).toContain('Warning')
    expect(wrapper.find('[data-check="orphans"]').text()).toContain('Pass')
    // Where the broken link was found is what makes it fixable.
    expect(wrapper.find('[data-check="broken_links"]').text()).toContain('linked from')
  })

  it('starts a check and shows it running', async () => {
    server.status = { configured: true, origin: report.origin, running: false }
    const wrapper = await mountSuspended(AdminSeo)
    await settle()

    const button = wrapper.find('button')
    expect(button.text()).toContain('Run check')
    await button.trigger('click')
    await settle()

    expect(server.posts).toBe(1)
    expect(wrapper.find('button').text()).toContain('Checking')
    expect(wrapper.find('button').attributes('disabled')).toBeDefined()
  })
})
