/**
 * Lists every unfilled content placeholder (§27).
 *
 * A placeholder is any string literal in content/ that starts with `TODO:`.
 * Those render as a visible "content pending" note rather than invented prose,
 * so this script is how you find them all before publishing.
 *
 *   node scripts/content-check.mjs            # report
 *   node scripts/content-check.mjs --strict   # exit 1 if any remain (for CI)
 */
import { readdir, readFile } from 'node:fs/promises'
import { join, resolve } from 'node:path'

const CONTENT_DIR = resolve(import.meta.dirname, '../content')
const strict = process.argv.includes('--strict')

const files = (await readdir(CONTENT_DIR)).filter((f) => f.endsWith('.ts')).sort()

let total = 0
let commented = 0

for (const file of files) {
  const source = await readFile(join(CONTENT_DIR, file), 'utf8')
  const lines = source.split('\n')
  const hits = []

  lines.forEach((line, i) => {
    if (!/['"`]TODO:/.test(line)) return
    const isComment = /^\s*(\/\/|\*)/.test(line)
    if (isComment) commented++
    else total++
    hits.push({
      line: i + 1,
      commented: isComment,
      text: line.trim().replace(/\s+/g, ' ').slice(0, 96),
    })
  })

  if (!hits.length) continue

  console.log(`\ncontent/${file}`)
  for (const hit of hits) {
    const mark = hit.commented ? 'inactive' : 'RENDERS '
    console.log(`  ${String(hit.line).padStart(4)}  ${mark}  ${hit.text}`)
  }
}

console.log(
  `\n${total} placeholder${total === 1 ? '' : 's'} currently render on the site` +
    (commented ? `, ${commented} more commented out` : '') +
    '.',
)

if (!total && !commented) console.log('All content filled in.')

if (strict && total > 0) {
  console.error('\nRefusing to pass: fill or remove the rendering placeholders above.')
  process.exit(1)
}
