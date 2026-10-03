// @vitest-environment node
import { describe, expect, it } from 'vitest'
import { escapeHtml, hasFormatting, renderRichText, toPlainText } from '~/utils/richtext'

describe('renderRichText', () => {
  it('wraps a plain line in a paragraph', () => {
    expect(renderRichText('Hello there.')).toBe('<p>Hello there.</p>')
  })

  it('splits paragraphs on a blank line and keeps single newlines as breaks', () => {
    expect(renderRichText('One\ntwo\n\nThree')).toBe('<p>One<br>two</p><p>Three</p>')
  })

  it('renders the emphasis it claims to support', () => {
    expect(renderRichText('**bold** and *italic* and `code`')).toBe(
      '<p><strong>bold</strong> and <em>italic</em> and <code>code</code></p>',
    )
  })

  it('does not read emphasis inside code', () => {
    expect(renderRichText('`a * b * c`')).toBe('<p><code>a * b * c</code></p>')
  })

  it('renders both kinds of list', () => {
    expect(renderRichText('- one\n- two')).toBe('<ul><li>one</li><li>two</li></ul>')
    expect(renderRichText('1. one\n2. two')).toBe('<ol><li>one</li><li>two</li></ol>')
  })

  it('returns nothing for nothing', () => {
    expect(renderRichText('')).toBe('')
    expect(renderRichText(null)).toBe('')
    expect(renderRichText(undefined)).toBe('')
  })
})

/**
 * The whole safety argument is that the string is escaped before any pattern is
 * matched, so the only tags in the output are ones this module emitted. An
 * editor account is a lower privilege than an admin; a script stored here would
 * run in an administrator's browser.
 */
describe('renderRichText is not an HTML passthrough', () => {
  it('escapes tags an author types', () => {
    expect(renderRichText('<script>alert(1)</script>')).toBe(
      '<p>&lt;script&gt;alert(1)&lt;/script&gt;</p>',
    )
  })

  it('escapes an event handler smuggled into markup', () => {
    const out = renderRichText('<img src=x onerror=alert(1)>')
    // The text stays readable — escaped, it is inert — but no tag is emitted.
    expect(out).toBe('<p>&lt;img src=x onerror=alert(1)&gt;</p>')
  })

  it('refuses a javascript: link and leaves the text visible', () => {
    const out = renderRichText('[click](javascript:alert(1))')
    expect(out).not.toContain('<a ')
    // Shown as the literal text it was, which is inert and tells the author
    // their link was rejected rather than silently swallowing it.
    expect(out).toContain('[click](javascript:alert(1))')
  })

  it.each([
    'JavaScript:alert(1)',
    ' javascript:alert(1)',
    'data:text/html;base64,PHNjcmlwdD4=',
    'vbscript:msgbox(1)',
  ])('refuses the scheme in %s', (href) => {
    const out = renderRichText(`[x](${href})`)
    expect(out).not.toContain('<a ')
  })

  it.each(['https://example.com', 'http://example.com', 'mailto:a@b.co', '/work', '#contact'])(
    'allows %s',
    (href) => {
      expect(renderRichText(`[x](${href})`)).toContain('<a href=')
    },
  )

  it('opens external links safely and leaves internal ones alone', () => {
    expect(renderRichText('[x](https://example.com)')).toContain('rel="noopener noreferrer"')
    expect(renderRichText('[x](/work)')).not.toContain('target=')
  })

  it('cannot be escaped out of by a quote in a link', () => {
    const out = renderRichText('[x](/a"onmouseover="alert(1))')
    expect(out).not.toContain('onmouseover="alert')
  })

  it('escapes ampersands and quotes in ordinary prose', () => {
    expect(renderRichText('Tom & "Jerry"')).toBe('<p>Tom &amp; &quot;Jerry&quot;</p>')
  })
})

describe('escapeHtml', () => {
  it('covers every character that can start markup or break an attribute', () => {
    expect(escapeHtml(`&<>"'`)).toBe('&amp;&lt;&gt;&quot;&#39;')
  })
})

describe('hasFormatting', () => {
  it('is false for a single plain sentence', () => {
    expect(hasFormatting('Just a sentence.')).toBe(false)
  })

  it.each(['**bold**', '- item', '1. item', '[a](/b)', 'two\nlines', '`code`'])(
    'is true for %s',
    (value) => expect(hasFormatting(value)).toBe(true),
  )
})

/**
 * Meta descriptions, Open Graph tags and JSON-LD are shown verbatim in search
 * results. Once descriptions could carry Markdown, those began rendering
 * `**bold**` and `[label](url)` as literal characters in the snippet.
 */
describe('toPlainText', () => {
  it('drops the syntax and keeps the words', () => {
    expect(toPlainText('A **bold** claim with `code` and *emphasis*')).toBe(
      'A bold claim with code and emphasis',
    )
  })

  it('keeps a link label and drops its address', () => {
    expect(toPlainText('See [the case study](https://example.com/x) for more')).toBe(
      'See the case study for more',
    )
  })

  it('flattens lists and line breaks into one line', () => {
    expect(toPlainText('Intro\n\n- one\n- two')).toBe('Intro one two')
  })

  it('leaves plain prose exactly as it was', () => {
    expect(toPlainText('Nothing to strip here.')).toBe('Nothing to strip here.')
  })

  it('returns an empty string for nothing', () => {
    expect(toPlainText(null)).toBe('')
    expect(toPlainText(undefined)).toBe('')
    expect(toPlainText('')).toBe('')
  })
})

describe('toPlainText and markup', () => {
  it('removes tag-shaped text, which a search snippet should never show', () => {
    expect(toPlainText('A <script>alert(1)</script> claim')).toBe('A alert(1) claim')
  })

  it('leaves a lone angle bracket alone', () => {
    expect(toPlainText('when a < b holds')).toBe('when a < b holds')
  })
})
