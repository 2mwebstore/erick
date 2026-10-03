import { describe, expect, it } from 'vitest'
import { LIMITS, validateContact } from '~/utils/validation'
import type { ContactPayload } from '~/types'

const valid = (): ContactPayload => ({
  name: 'Test Visitor',
  email: 'visitor@example.com',
  projectType: 'Web Application',
  message: 'We need an internal ordering system for about forty staff.',
})

const errorKeys = (payload: ContactPayload) =>
  Object.entries(validateContact(payload))
    .filter(([, v]) => Boolean(v))
    .map(([k]) => k)

describe('contact validation', () => {
  it('accepts a complete submission', () => {
    expect(errorKeys(valid())).toEqual([])
  })

  it('rejects a missing name', () => {
    expect(errorKeys({ ...valid(), name: '' })).toContain('name')
    expect(errorKeys({ ...valid(), name: 'A' })).toContain('name')
  })

  it('rejects malformed email addresses', () => {
    for (const email of ['', 'nope', 'a@b', 'a b@example.com', 'visitor@localhost']) {
      expect(errorKeys({ ...valid(), email })).toContain('email')
    }
  })

  it('rejects a project type outside the list', () => {
    expect(errorKeys({ ...valid(), projectType: '' })).toContain('projectType')
    expect(errorKeys({ ...valid(), projectType: 'Blockchain' as never })).toContain('projectType')
  })

  it('rejects a message that is too short or too long', () => {
    expect(errorKeys({ ...valid(), message: 'hi' })).toContain('message')
    expect(errorKeys({ ...valid(), message: 'a'.repeat(LIMITS.message.max + 1) })).toContain('message')
  })

  it('ignores surrounding whitespace when measuring length', () => {
    expect(errorKeys({ ...valid(), message: `   ${'a'.repeat(9)}   ` })).toContain('message')
    expect(errorKeys({ ...valid(), message: `   ${'a'.repeat(10)}   ` })).not.toContain('message')
  })

  it('reports every invalid field at once rather than stopping at the first', () => {
    const keys = errorKeys({ name: '', email: 'nope', projectType: '', message: '' })
    expect(keys.sort()).toEqual(['email', 'message', 'name', 'projectType'])
  })

  it('keeps limits aligned with the Go service', () => {
    // backend/internal/services/contact.go declares the same numbers. If one
    // side changes, this test is the reminder to change the other.
    expect(LIMITS).toEqual({
      name: { min: 2, max: 100 },
      email: { max: 254 },
      message: { min: 10, max: 4000 },
      phone: { max: 40 },
      subject: { max: 160 },
    })
  })
})
