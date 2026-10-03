import { describe, expect, it } from 'vitest'
import { absoluteUrl, slugSeed, yearRange, yearsOfExperience } from '~/utils/format'

describe('url building', () => {
  it('keeps the root path trailing slash so canonical matches the sitemap', () => {
    expect(absoluteUrl('/', 'https://example.com')).toBe('https://example.com/')
  })

  it('strips duplicate slashes from the origin', () => {
    expect(absoluteUrl('/work', 'https://example.com/')).toBe('https://example.com/work')
    expect(absoluteUrl('/work', 'https://example.com///')).toBe('https://example.com/work')
  })

  it('accepts a path with no leading slash', () => {
    expect(absoluteUrl('resume', 'https://example.com')).toBe('https://example.com/resume')
  })
})

describe('formatting helpers', () => {
  it('renders the experience range', () => {
    expect(yearRange(2020)).toBe('2020 — Present')
  })

  it('counts years of experience from a fixed date', () => {
    expect(yearsOfExperience(2020, new Date('2026-09-23'))).toBe(6)
    expect(yearsOfExperience(2030, new Date('2026-09-23'))).toBe(0)
  })

  it('derives a stable seed from a slug', () => {
    expect(slugSeed('bubble-white')).toBe(slugSeed('bubble-white'))
    expect(slugSeed('bubble-white')).not.toBe(slugSeed('coffee-app'))
  })
})
