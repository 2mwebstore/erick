// @vitest-environment nuxt
import { describe, expect, it } from 'vitest'
import { mountSuspended, registerEndpoint } from '@nuxt/test-utils/runtime'
// h3 utilities are not auto-imported inside a test file.
import { readBody } from 'h3'
import SectionsContact from '~/components/sections/Contact.vue'

// Held on an object rather than in a bare `let`: TypeScript narrows a reassigned
// local to `null` at the point of use, since the write happens in a callback.
const captured: { body: Record<string, unknown> | null } = { body: null }

registerEndpoint('/api/contact', {
  method: 'POST',
  handler: async (event) => {
    captured.body = await readBody(event)
    return { ok: true, message: 'Thanks — your message has been sent.' }
  },
})

async function fillValid(wrapper: Awaited<ReturnType<typeof mountSuspended>>) {
  await wrapper.find('#contact-name').setValue('Test Visitor')
  await wrapper.find('#contact-email').setValue('visitor@example.com')
  await wrapper.find('#contact-projectType').setValue('Web Application')
  await wrapper
    .find('#contact-message')
    .setValue('We need an internal ordering system for about forty staff.')
}

describe('contact form', () => {
  it('renders every required field and the project type options', async () => {
    const wrapper = await mountSuspended(SectionsContact)

    expect(wrapper.find('#contact-name').exists()).toBe(true)
    expect(wrapper.find('#contact-email').exists()).toBe(true)
    expect(wrapper.find('#contact-projectType').exists()).toBe(true)
    expect(wrapper.find('#contact-message').exists()).toBe(true)

    const options = wrapper.findAll('#contact-projectType option').map((o) => o.text())
    expect(options).toContain('Web Application')
    expect(options).toContain('WordPress / SEO')
    expect(options).toContain('DevOps')
  })

  it('labels every field for assistive technology (§24)', async () => {
    const wrapper = await mountSuspended(SectionsContact)

    for (const id of ['contact-name', 'contact-email', 'contact-projectType', 'contact-message']) {
      const label = wrapper.find(`label[for="${id}"]`)
      expect(label.exists(), `${id} has no label`).toBe(true)
      expect(label.text().trim().length).toBeGreaterThan(0)
    }
  })

  it('blocks submission and marks fields invalid when empty', async () => {
    const wrapper = await mountSuspended(SectionsContact)
    captured.body = null

    await wrapper.find('form').trigger('submit')
    await new Promise((r) => setTimeout(r, 0))

    expect(captured.body).toBeNull()
    expect(wrapper.text()).toContain('Please correct the highlighted fields.')
    expect(wrapper.find('#contact-name').attributes('aria-invalid')).toBe('true')
    expect(wrapper.find('#contact-email').attributes('aria-invalid')).toBe('true')
  })

  it('ties each error message to its field with aria-describedby', async () => {
    const wrapper = await mountSuspended(SectionsContact)

    await wrapper.find('form').trigger('submit')
    await new Promise((r) => setTimeout(r, 0))

    const describedBy = wrapper.find('#contact-email').attributes('aria-describedby')
    expect(describedBy).toContain('contact-email-error')
    expect(wrapper.find('#contact-email-error').exists()).toBe(true)
  })

  it('clears a field error as soon as the visitor edits it', async () => {
    const wrapper = await mountSuspended(SectionsContact)

    await wrapper.find('form').trigger('submit')
    await new Promise((r) => setTimeout(r, 0))
    expect(wrapper.find('#contact-name').attributes('aria-invalid')).toBe('true')

    await wrapper.find('#contact-name').setValue('Test Visitor')
    expect(wrapper.find('#contact-name').attributes('aria-invalid')).toBeUndefined()
  })

  it('submits a valid message and reports success', async () => {
    const wrapper = await mountSuspended(SectionsContact)
    captured.body = null

    await fillValid(wrapper)
    await wrapper.find('form').trigger('submit')
    await new Promise((r) => setTimeout(r, 50))

    expect(captured.body).toMatchObject({
      name: 'Test Visitor',
      email: 'visitor@example.com',
      projectType: 'Web Application',
    })
    expect(wrapper.text()).toContain('your message has been sent')
  })

  it('trims whitespace before sending', async () => {
    const wrapper = await mountSuspended(SectionsContact)
    captured.body = null

    await wrapper.find('#contact-name').setValue('   Test Visitor   ')
    await wrapper.find('#contact-email').setValue('  visitor@example.com  ')
    await wrapper.find('#contact-projectType').setValue('Other')
    await wrapper.find('#contact-message').setValue('   A message with padding around it.   ')

    await wrapper.find('form').trigger('submit')
    await new Promise((r) => setTimeout(r, 50))

    expect(captured.body).toMatchObject({
      name: 'Test Visitor',
      email: 'visitor@example.com',
      message: 'A message with padding around it.',
    })
  })

  it('clears the form after a successful send', async () => {
    const wrapper = await mountSuspended(SectionsContact)

    await fillValid(wrapper)
    await wrapper.find('form').trigger('submit')
    await new Promise((r) => setTimeout(r, 50))

    expect((wrapper.find('#contact-name').element as HTMLInputElement).value).toBe('')
    expect((wrapper.find('#contact-message').element as HTMLTextAreaElement).value).toBe('')
  })

  it('hides the honeypot from sight and from assistive technology (§20)', async () => {
    const wrapper = await mountSuspended(SectionsContact)

    const honeypot = wrapper.find('#contact-company')
    expect(honeypot.exists()).toBe(true)
    expect(honeypot.attributes('tabindex')).toBe('-1')
    expect(honeypot.attributes('autocomplete')).toBe('off')

    // The wrapper is off-screen and aria-hidden, so a real visitor never sees it.
    const container = honeypot.element.closest('[aria-hidden="true"]')
    expect(container).not.toBeNull()
  })

  it('announces status changes politely rather than stealing focus', async () => {
    const wrapper = await mountSuspended(SectionsContact)
    const live = wrapper.find('[aria-live="polite"]')
    expect(live.exists()).toBe(true)
    expect(live.attributes('role')).toBe('status')
  })

  it('counts characters against the server limit', async () => {
    const wrapper = await mountSuspended(SectionsContact)
    await wrapper.find('#contact-message').setValue('Twelve chars')
    expect(wrapper.text()).toContain('12 / 4000')
  })
})
