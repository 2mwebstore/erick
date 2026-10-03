import { describe, expect, it } from 'vitest'
import { hasContent, isPending, pendingHint, realItems, hasItems } from '~/utils/content'

describe('placeholder handling', () => {
  it('recognises a TODO string as pending', () => {
    expect(isPending('TODO: write this')).toBe(true)
    expect(isPending('  TODO: leading space')).toBe(true)
  })

  it('does not treat real prose as pending', () => {
    expect(isPending('A real sentence.')).toBe(false)
    expect(isPending('This mentions a TODO mid-sentence.')).toBe(false)
  })

  it('treats empty and missing values as absent, not pending', () => {
    expect(isPending('')).toBe(false)
    expect(isPending(undefined)).toBe(false)
    expect(isPending(null)).toBe(false)
    expect(hasContent('')).toBe(false)
    expect(hasContent('   ')).toBe(false)
    expect(hasContent(undefined)).toBe(false)
  })

  it('accepts real content only', () => {
    expect(hasContent('Real text')).toBe(true)
    expect(hasContent('TODO: not yet')).toBe(false)
  })

  it('strips the marker for the editor hint', () => {
    expect(pendingHint('TODO:  what the app does')).toBe('what the app does')
  })

  it('filters placeholder entries out of lists', () => {
    expect(realItems(['Real one', 'TODO: later', 'Real two'])).toEqual(['Real one', 'Real two'])
    expect(realItems(undefined)).toEqual([])
    expect(hasItems(['TODO: only a placeholder'])).toBe(false)
    expect(hasItems(['Something real'])).toBe(true)
  })
})
