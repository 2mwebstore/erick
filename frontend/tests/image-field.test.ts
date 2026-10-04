// @vitest-environment nuxt
import { beforeEach, describe, expect, it } from 'vitest'
import { mountSuspended, registerEndpoint } from '@nuxt/test-utils/runtime'
import { setResponseStatus } from 'h3'
import AdminImageField from '~/components/admin/ImageField.vue'

/**
 * An image field takes an upload (stored in R2 by the API) or a pasted link.
 * What the API accepts is tested in Go; this covers what the field sends and does.
 */

const received = { calls: 0 }
const reply: { status: number; body: Record<string, unknown> } = { status: 201, body: {} }

registerEndpoint('/api/admin/uploads/image', {
  method: 'POST',
  handler: (event) => {
    received.calls++
    if (reply.status !== 201) setResponseStatus(event, reply.status)
    return reply.body
  },
})

const settle = () => new Promise((resolve) => setTimeout(resolve, 30))

function uploadsOn(on: boolean) {
  useState('admin-upload-settings').value = on
    ? { enabled: true, maxBytes: 1024, types: ['image/jpeg', 'image/png', 'image/webp', 'image/gif', 'image/avif'] }
    : { enabled: false }
}

const mount = (modelValue = '') =>
  mountSuspended(AdminImageField, { props: { id: 'img', label: 'Image', folder: 'projects', modelValue } })

async function choose(wrapper: Awaited<ReturnType<typeof mount>>, file: File) {
  const input = wrapper.find('input[type="file"]')
  Object.defineProperty(input.element, 'files', { value: [file], configurable: true })
  await input.trigger('change')
  await settle()
}

beforeEach(() => {
  received.calls = 0
  reply.status = 201
  reply.body = { ok: true, url: 'https://cdn.example.com/projects/2026/10/abc.png' }
  useState('admin-user').value = { id: 1, email: 'someone@example.com', name: 'Someone', role: 'editor' }
  useState('admin-csrf').value = 'token'
})

describe('image field', () => {
  it('takes a pasted link when uploads are off, and says so', async () => {
    uploadsOn(false)
    const wrapper = await mount()

    expect(wrapper.find('input[type="file"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('Paste an image link')

    await wrapper.find('#img').setValue('  https://i.imgur.com/x.png ')
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual(['https://i.imgur.com/x.png'])
  })

  // The test environment's $fetch cannot carry FormData to a registered
  // endpoint (it arrives as a placeholder string), so the form is checked as it
  // is handed to $fetch. The bytes surviving the real proxy are checked end to
  // end instead.
  it('uploads the chosen file to its folder and fills in the stored image\'s URL', async () => {
    uploadsOn(true)
    const sent: { url?: string; method?: string; body?: unknown } = {}
    const original = globalThis.$fetch
    globalThis.$fetch = Object.assign(
      async (url: string, options: { method?: string; body?: unknown } = {}) => {
        Object.assign(sent, { url, method: options.method, body: options.body })
        return { ok: true, url: 'https://cdn.example.com/projects/2026/10/abc.png' }
      },
      original,
    ) as typeof $fetch

    try {
      const wrapper = await mount()
      await choose(wrapper, new File([new Uint8Array([0x89, 0x50, 0x4e, 0x47])], 'shot.png', { type: 'image/png' }))

      expect(sent.url).toBe('/api/admin/uploads/image')
      expect(sent.method).toBe('POST')
      const form = sent.body as FormData
      expect(form).toBeInstanceOf(FormData)
      expect(form.get('folder')).toBe('projects')
      expect((form.get('file') as File).name).toBe('shot.png')
      expect((form.get('file') as File).size).toBe(4)
      expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual(['https://cdn.example.com/projects/2026/10/abc.png'])
    } finally {
      globalThis.$fetch = original
    }
  })

  it('does not send a file over the limit', async () => {
    uploadsOn(true)
    const wrapper = await mount()

    await choose(wrapper, new File([new Uint8Array(2048)], 'big.png', { type: 'image/png' }))

    expect(received.calls).toBe(0)
    expect(wrapper.text()).toContain('Images can be up to 1 KB')
  })

  it('shows why the API refused a file', async () => {
    uploadsOn(true)
    reply.status = 400
    reply.body = { ok: false, message: 'Only images.', errors: { file: 'Only JPEG, PNG, WebP, GIF or AVIF images can be uploaded.' } }
    const wrapper = await mount()

    await choose(wrapper, new File(['<svg/>'], 'logo.png', { type: 'image/png' }))

    expect(wrapper.find('[role="alert"]').text()).toContain('Only JPEG, PNG, WebP, GIF or AVIF')
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
  })
})
