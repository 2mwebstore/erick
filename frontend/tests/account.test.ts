// @vitest-environment nuxt
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mountSuspended, mockNuxtImport, registerEndpoint } from '@nuxt/test-utils/runtime'
import { readBody, setResponseStatus } from 'h3'
import AdminAccount from '~/pages/admin/account.vue'

/**
 * Changing your own password.
 *
 * The API has had an endpoint for this since it was written; nothing called it,
 * so the only way to change a password was `adminctl passwd` on the server —
 * which is how one ended up written into the README.
 */

const sent: { body: Record<string, unknown> | null } = { body: null }
const reply: { status: number; body: Record<string, unknown> } = {
  status: 200,
  body: { ok: true, message: 'Password changed. Please sign in again.' },
}

registerEndpoint('/api/auth/password', {
  method: 'POST',
  handler: async (event) => {
    sent.body = await readBody(event)
    // A status and a plain JSON body, exactly as the Go API answers — not an
    // h3 error, which wraps the body a level deeper than $fetch reports it.
    if (reply.status !== 200) setResponseStatus(event, reply.status)
    return reply.body
  },
})

// Hoisted: mockNuxtImport's factory runs before the file's own statements.
const { navigate } = vi.hoisted(() => ({ navigate: vi.fn() }))
mockNuxtImport('navigateTo', () => navigate)

const mount = () => mountSuspended(AdminAccount)

async function fill(
  wrapper: Awaited<ReturnType<typeof mount>>,
  values: { current?: string; next?: string; confirm?: string },
) {
  if (values.current !== undefined) await wrapper.find('#pw-current').setValue(values.current)
  if (values.next !== undefined) await wrapper.find('#pw-next').setValue(values.next)
  if (values.confirm !== undefined) await wrapper.find('#pw-confirm').setValue(values.confirm)
}

const submit = async (wrapper: Awaited<ReturnType<typeof mount>>) => {
  await wrapper.find('form').trigger('submit')
  await new Promise((resolve) => setTimeout(resolve, 20))
}

beforeEach(() => {
  sent.body = null
  reply.status = 200
  reply.body = { ok: true, message: 'Password changed. Please sign in again.' }
  navigate.mockClear()
  useState('admin-user').value = { id: 1, email: 'someone@example.com', name: 'Someone', role: 'admin' }
  useState('admin-csrf').value = 'token'
})

describe('changing your own password', () => {
  it('sends the current and the new password, and nothing else', async () => {
    const wrapper = await mount()
    await fill(wrapper, { current: 'old passphrase 1', next: 'new passphrase 2', confirm: 'new passphrase 2' })
    await submit(wrapper)

    expect(sent.body).toEqual({ current: 'old passphrase 1', next: 'new passphrase 2' })
  })

  /**
   * The server revokes every session, this one included. Staying on the page
   * would leave a panel that 401s on the next thing you touch.
   */
  it('sends you back to sign in, holding no session behind', async () => {
    const wrapper = await mount()
    await fill(wrapper, { current: 'old passphrase 1', next: 'new passphrase 2', confirm: 'new passphrase 2' })
    await submit(wrapper)

    expect(navigate).toHaveBeenCalledWith('/admin/login')
    expect(useState('admin-user').value).toBeNull()
    expect(useState('admin-csrf').value).toBe('')
  })

  it('does not submit a new password that is too short', async () => {
    const wrapper = await mount()
    await fill(wrapper, { current: 'old passphrase 1', next: 'short', confirm: 'short' })
    await submit(wrapper)

    expect(sent.body).toBeNull()
    expect(wrapper.text()).toContain('Use at least 12 characters')
  })

  it('does not submit when the confirmation does not match', async () => {
    const wrapper = await mount()
    await fill(wrapper, { current: 'old passphrase 1', next: 'new passphrase 2', confirm: 'new passphrase 3' })
    await submit(wrapper)

    expect(sent.body).toBeNull()
    expect(wrapper.text()).toContain('does not match')
  })

  /** bcrypt truncates past 72 bytes rather than failing, so the server refuses it. */
  it('counts the limit in bytes, not characters', async () => {
    const wrapper = await mount()
    // 40 characters, but 120 bytes in UTF-8.
    const khmer = 'ពាក្យសម្ងាត់'.repeat(4)
    await fill(wrapper, { current: 'old passphrase 1', next: khmer, confirm: khmer })
    await submit(wrapper)

    expect(sent.body).toBeNull()
    expect(wrapper.text()).toContain('72 bytes')
  })

  it('refuses to set the password that is already in use', async () => {
    const wrapper = await mount()
    await fill(wrapper, { current: 'same passphrase 1', next: 'same passphrase 1', confirm: 'same passphrase 1' })
    await submit(wrapper)

    expect(sent.body).toBeNull()
    expect(wrapper.text()).toContain('password you already have')
  })

  /**
   * The server reports a rejected new password under `password`, because the
   * same validator runs when an account is created. Shown against `next`, it
   * would appear under no field at all.
   */
  it('shows a rejection from the server against the field it is about', async () => {
    reply.status = 422
    reply.body = { message: 'Please check the form.', errors: { password: 'Use a longer passphrase.' } }

    const wrapper = await mount()
    await fill(wrapper, { current: 'old passphrase 1', next: 'new passphrase 2', confirm: 'new passphrase 2' })
    await submit(wrapper)

    const field = wrapper.find('#pw-next-error')
    expect(field.exists()).toBe(true)
    expect(field.text()).toContain('Use a longer passphrase.')
    expect(navigate).not.toHaveBeenCalled()
  })

  it('says so when the current password is wrong, and stays put', async () => {
    reply.status = 422
    reply.body = { message: 'Please check the form.', errors: { current: 'That is not your current password.' } }

    const wrapper = await mount()
    await fill(wrapper, { current: 'wrong passphrase', next: 'new passphrase 2', confirm: 'new passphrase 2' })
    await submit(wrapper)

    expect(wrapper.find('#pw-current-error').text()).toContain('not your current password')
    expect(navigate).not.toHaveBeenCalled()
    expect(useState('admin-user').value).not.toBeNull()
  })

  /** A password manager can only offer to save the new one if it is told. */
  it('marks the fields for a password manager', async () => {
    const wrapper = await mount()

    expect(wrapper.find('#pw-current').attributes('autocomplete')).toBe('current-password')
    expect(wrapper.find('#pw-next').attributes('autocomplete')).toBe('new-password')
    expect(wrapper.find('#pw-confirm').attributes('autocomplete')).toBe('new-password')
    for (const id of ['#pw-current', '#pw-next', '#pw-confirm']) {
      expect(wrapper.find(id).attributes('type')).toBe('password')
    }
  })
})
