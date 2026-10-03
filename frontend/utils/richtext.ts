/**
 * The small formatting subset the admin editor writes and the site renders.
 *
 * Content is stored as Markdown, not HTML, and rendered by escaping the whole
 * string *first* and only then putting back the handful of tags below. That
 * ordering is the entire security argument: by the time any pattern is matched,
 * every `<` in the author's text is already `&lt;`, so there is no path by which
 * an editor account — a lower privilege than admin — could store a script that
 * runs in an administrator's browser.
 *
 * Deliberately small. Headings, images, tables and raw HTML are not supported,
 * because the page designs have their own heading levels and their own image
 * handling, and a description that can restructure the page is a description
 * that can break it.
 */

/** Schemes a link may use. Anything else — `javascript:` above all — is dropped. */
const SAFE_SCHEME = /^(https?:\/\/|mailto:|\/|#)/i

const ESCAPES: Record<string, string> = {
  '&': '&amp;',
  '<': '&lt;',
  '>': '&gt;',
  '"': '&quot;',
  "'": '&#39;',
}

export function escapeHtml(value: string): string {
  return value.replace(/[&<>"']/g, (c) => ESCAPES[c]!)
}

/**
 * Inline formatting, applied to text that has already been escaped.
 *
 * Code spans are lifted out before anything else runs and put back at the end.
 * Emitting `<code>` first is not enough — the emphasis patterns that follow
 * happily match the asterisks *inside* it, so `` `a * b * c` `` came out with
 * an `<em>` in the middle of what is supposed to be literal.
 */
function inline(text: string): string {
  const code: string[] = []
  const stashed = text.replace(/`([^`\n]+)`/g, (_, body: string) => {
    code.push(body)
    return `${code.length - 1}`
  })

  return (
    stashed
      .replace(/\*\*([^*\n]+)\*\*/g, '<strong>$1</strong>')
      .replace(/(^|[^*])\*([^*\n]+)\*/g, '$1<em>$2</em>')
      .replace(/\[([^\]\n]+)\]\(([^)\s]+)\)/g, (whole, label: string, href: string) => {
        // The href was escaped with the rest of the string, so `&` is `&amp;`
        // here; it is checked in its original form and re-emitted as-is.
        const raw = href.replace(/&amp;/g, '&')
        if (!SAFE_SCHEME.test(raw)) return whole
        const external = /^https?:\/\//i.test(raw)
        const rel = external ? ' target="_blank" rel="noopener noreferrer"' : ''
        return `<a href="${escapeHtml(raw)}"${rel}>${label}</a>`
      })
      .replace(/(\d+)/g, (_, i: string) => `<code>${code[Number(i)]}</code>`)
  )
}

/**
 * Renders the subset to HTML.
 *
 * Blank lines separate paragraphs; a run of `-` or `1.` lines becomes a list.
 * Everything else is a paragraph with single newlines preserved as breaks.
 */
export function renderRichText(value: string | null | undefined): string {
  if (!value) return ''

  // U+E000 is the sentinel `inline` uses to park code spans. It is in a
  // private-use block, so it carries no meaning of its own, but it is stripped
  // from the author's text regardless — a sentinel that can be typed is not a
  // sentinel.
  const escaped = escapeHtml(value.replace(/\uE000/g, '').replace(/\r\n?/g, '\n'))
  const blocks = escaped.split(/\n{2,}/)
  const html: string[] = []

  for (const block of blocks) {
    const lines = block.split('\n').filter((line) => line.trim() !== '')
    if (!lines.length) continue

    const bulleted = lines.every((line) => /^\s*[-*]\s+/.test(line))
    const numbered = lines.every((line) => /^\s*\d+\.\s+/.test(line))

    if (bulleted || numbered) {
      const tag = bulleted ? 'ul' : 'ol'
      const items = lines
        .map((line) => `<li>${inline(line.replace(/^\s*(?:[-*]|\d+\.)\s+/, ''))}</li>`)
        .join('')
      html.push(`<${tag}>${items}</${tag}>`)
      continue
    }

    html.push(`<p>${inline(lines.join('<br>'))}</p>`)
  }

  return html.join('')
}

/** True when the value uses any of the subset, for deciding whether to preview. */
export function hasFormatting(value: string | null | undefined): boolean {
  if (!value) return false
  return /(\*\*|`|\[[^\]]+\]\(|^\s*[-*]\s+|^\s*\d+\.\s+|\n)/m.test(value)
}

/**
 * Strips the formatting back out, for places that can only take plain text.
 *
 * Meta descriptions, Open Graph tags and JSON-LD are read by machines and shown
 * in search results verbatim: once descriptions could carry Markdown, those
 * started rendering `**bold**` and `[label](url)` as literal characters in the
 * snippet. The prose is kept, the syntax is dropped.
 */
export function toPlainText(value: string | null | undefined): string {
  if (!value) return ''

  return value
    // Tag-shaped text is inert in an attribute, but a search snippet reading
    // "<script>" helps nobody. Only complete tag shapes go, so "a < b" stays.
    .replace(/<\/?[a-zA-Z][^>]*>/g, '')
    .replace(/`([^`\n]+)`/g, '$1')
    .replace(/\*\*([^*\n]+)\*\*/g, '$1')
    .replace(/(^|[^*])\*([^*\n]+)\*/g, '$1$2')
    // Keep the label, drop the address — a URL read aloud in a search snippet
    // is noise.
    .replace(/\[([^\]\n]+)\]\([^)\s]+\)/g, '$1')
    .replace(/^\s*(?:[-*]|\d+\.)\s+/gm, '')
    .replace(/\s*\n\s*/g, ' ')
    .trim()
}
