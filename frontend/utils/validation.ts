import type { ContactPayload } from '~/types'
import { projectTypes } from '~/content/services'

/**
 * Client-side mirror of the Go validator in
 * backend/internal/services/contact.go. The server is authoritative — this only
 * saves the visitor a round trip.
 */
export const LIMITS = {
  name: { min: 2, max: 100 },
  email: { max: 254 },
  message: { min: 10, max: 4000 },
  phone: { max: 40 },
  subject: { max: 160 },
} as const

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[a-z]{2,}$/i

export type ContactErrors = Partial<Record<keyof ContactPayload, string>>

export function validateContact(payload: ContactPayload): ContactErrors {
  const errors: ContactErrors = {}
  const name = payload.name.trim()
  const email = payload.email.trim()
  const message = payload.message.trim()

  if (name.length < LIMITS.name.min) errors.name = 'Please enter your name.'
  else if (name.length > LIMITS.name.max) errors.name = `Keep this under ${LIMITS.name.max} characters.`

  if (!email) errors.email = 'Please enter your email address.'
  else if (email.length > LIMITS.email.max) errors.email = 'That email address is too long.'
  else if (!EMAIL_RE.test(email)) errors.email = 'Please enter a valid email address.'

  if (!payload.projectType) errors.projectType = 'Please choose a project type.'
  else if (!projectTypes.includes(payload.projectType as (typeof projectTypes)[number]))
    errors.projectType = 'Please choose a project type from the list.'

  if (message.length < LIMITS.message.min) errors.message = 'Please add a few details about the project.'
  else if (message.length > LIMITS.message.max)
    errors.message = `Keep this under ${LIMITS.message.max} characters.`

  // Phone and subject are optional — only their length is checked, and the
  // numbers match backend/internal/services/contact.go.
  const phone = payload.phone?.trim() ?? ''
  if (phone.length > LIMITS.phone.max) errors.phone = `Keep this under ${LIMITS.phone.max} characters.`

  const subject = payload.subject?.trim() ?? ''
  if (subject.length > LIMITS.subject.max)
    errors.subject = `Keep this under ${LIMITS.subject.max} characters.`

  return errors
}
