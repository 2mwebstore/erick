/**
 * Renders public/og-default.png (1200×630) for Open Graph and Twitter cards.
 *
 * Social platforms do not render SVG previews, so the card has to ship as a
 * raster. sharp is already present via @nuxt/image, so this needs no extra
 * dependency. Run it whenever the name, role, or tagline changes:
 *
 *   node scripts/generate-og.mjs
 */
import { mkdir, writeFile } from 'node:fs/promises'
import { dirname, resolve } from 'node:path'
import sharp from 'sharp'

const NAME = 'Kong Chansila'
const ROLE = 'Full-Stack Software Developer'
const TAGLINE = 'Production-ready web, mobile, and business systems since 2020'
const STACK = 'Go  ·  Laravel  ·  Vue  ·  Nuxt  ·  Flutter  ·  MySQL  ·  Docker  ·  Cloudflare'

// System font stack: librsvg resolves by family name, and Inter may not be
// installed on the machine running this script.
const SANS = 'Helvetica Neue, Helvetica, Arial, sans-serif'
const MONO = 'SFMono-Regular, Menlo, Consolas, monospace'

const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="1200" height="630" viewBox="0 0 1200 630">
  <rect width="1200" height="630" fill="#0a0a0b"/>

  <g stroke="#1c1c20" stroke-width="1">
    ${Array.from({ length: 12 }, (_, i) => `<line x1="0" y1="${(i + 1) * 52}" x2="1200" y2="${(i + 1) * 52}"/>`).join('\n    ')}
    ${Array.from({ length: 23 }, (_, i) => `<line x1="${(i + 1) * 52}" y1="0" x2="${(i + 1) * 52}" y2="630"/>`).join('\n    ')}
  </g>

  <rect x="0" y="0" width="1200" height="4" fill="#1d4ed8"/>

  <text x="88" y="150" font-family="${MONO}" font-size="20" letter-spacing="4" fill="#71717a">${ROLE.toUpperCase()}</text>

  <text x="88" y="268" font-family="${SANS}" font-size="84" font-weight="600" letter-spacing="-3" fill="#fafafa">${NAME}</text>

  <text x="88" y="348" font-family="${SANS}" font-size="32" fill="#a1a1aa">${TAGLINE}</text>

  <line x1="88" y1="430" x2="1112" y2="430" stroke="#26262a" stroke-width="1"/>

  <text x="88" y="482" font-family="${MONO}" font-size="21" letter-spacing="1" fill="#71717a">${STACK}</text>

  <text x="88" y="556" font-family="${MONO}" font-size="19" letter-spacing="3" fill="#60a5fa">ARCHITECTURE → DEPLOYMENT → PRODUCTION</text>
</svg>`

const out = resolve(import.meta.dirname, '../public/og-default.png')
await mkdir(dirname(out), { recursive: true })
const png = await sharp(Buffer.from(svg)).png({ compressionLevel: 9 }).toBuffer()
await writeFile(out, png)
console.log(`og image written: ${out} (${(png.length / 1024).toFixed(1)} kB)`)
