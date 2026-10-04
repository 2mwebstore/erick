import { describe, expect, it } from 'vitest'
import { trailingSlashTarget } from '~/server/utils/trailing-slash'
import { absoluteUrl, sameAsUrls, slugSeed, yearRange, yearsOfExperience } from '~/utils/format'
import { localizedPath } from '~/utils/locales'
import { siteVerificationTokens } from '~/utils/search-console'

describe('siteVerificationTokens', () => {
  const code = 'aBcD-1234_efGH5678ijKL9012mnOP3456qrST7890uvw'

  it('reads a bare code', () => {
    expect(siteVerificationTokens(code)).toEqual([code])
  })

  it('reads the whole <meta> tag pasted from Search Console', () => {
    expect(siteVerificationTokens(`<meta name="google-site-verification" content="${code}" />`)).toEqual([code])
  })

  it('reads one code per owner, separated by commas or spaces', () => {
    expect(siteVerificationTokens(`${code}, second-owner-code-123 ${code}`)).toEqual([code, 'second-owner-code-123'])
  })

  it('reads a pasted tag and a bare code side by side', () => {
    expect(
      siteVerificationTokens(`<meta name="google-site-verification" content="${code}" />, second-owner-code-123`),
    ).toEqual([code, 'second-owner-code-123'])
  })

  it('is empty when the variable is unset', () => {
    expect(siteVerificationTokens(undefined)).toEqual([])
    expect(siteVerificationTokens('')).toEqual([])
  })

  // The value lands in the <head> of every page.
  it('drops anything that is not a code', () => {
    expect(siteVerificationTokens('"><script>alert(1)</script>')).toEqual([])
    expect(siteVerificationTokens('short')).toEqual([])
  })
})

describe('localizedPath', () => {
  it('leaves English paths unprefixed', () => {
    expect(localizedPath('/', 'en')).toBe('/')
    expect(localizedPath('/work/bubble-white', 'en')).toBe('/work/bubble-white')
  })

  // Every /km page used to declare its English twin as canonical.
  it('puts Khmer pages under /km, the home page without a trailing slash', () => {
    expect(localizedPath('/', 'km')).toBe('/km')
    expect(localizedPath('/work', 'km')).toBe('/km/work')
    expect(localizedPath('/work/bubble-white', 'km')).toBe('/km/work/bubble-white')
  })

  it('answers an unknown language in English rather than inventing a prefix', () => {
    expect(localizedPath('/work', 'fr')).toBe('/work')
  })

  it('builds the same absolute URLs the sitemap lists', () => {
    expect(absoluteUrl(localizedPath('/', 'en'), 'https://example.com')).toBe('https://example.com/')
    expect(absoluteUrl(localizedPath('/', 'km'), 'https://example.com')).toBe('https://example.com/km')
  })
})

describe('sameAsUrls', () => {
  const origin = 'https://kongchansila.com'

  it('keeps real profiles on other sites', () => {
    expect(sameAsUrls(['https://github.com/kong', 'https://www.linkedin.com/in/kong/'], origin)).toEqual([
      'https://github.com/kong',
      'https://www.linkedin.com/in/kong/',
    ])
  })

  // What production emitted: four "/" placeholders, an invalid sameAs.
  it('drops placeholders, relative links and links back to this site', () => {
    expect(sameAsUrls(['/', '/', '#', 'mailto:a@b.c', 'https://kongchansila.com/work'], origin)).toEqual([])
  })

  it('lists a profile once', () => {
    expect(sameAsUrls(['https://github.com/kong', ' https://github.com/kong '], origin)).toEqual(['https://github.com/kong'])
  })
})

describe('trailingSlashTarget', () => {
  it('redirects a trailing-slash page to the one canonical URL', () => {
    expect(trailingSlashTarget('/work/')).toBe('/work')
    expect(trailingSlashTarget('/km/')).toBe('/km')
    expect(trailingSlashTarget('/work/bubble-white//')).toBe('/work/bubble-white')
  })

  it('serves the home page and slash-less paths as they are', () => {
    expect(trailingSlashTarget('/')).toBeNull()
    expect(trailingSlashTarget('/work')).toBeNull()
  })

  it('leaves the API and build assets alone', () => {
    expect(trailingSlashTarget('/api/content/')).toBeNull()
    expect(trailingSlashTarget('/_nuxt/')).toBeNull()
  })

  it('never produces a protocol-relative, off-site redirect', () => {
    expect(trailingSlashTarget('//evil.example/')).toBe('/evil.example')
    expect(trailingSlashTarget('///evil.example/path/')).toBe('/evil.example/path')
  })
})

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
