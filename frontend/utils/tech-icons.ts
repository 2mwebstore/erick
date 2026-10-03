/**
 * Technology → icon and brand colour.
 *
 * Colours are the official Simple Icons brand hex. Several are unusable as-is on
 * one of the two backgrounds — Apple's black disappears on near-black, Firebase
 * amber washes out on white — so an entry can carry a `dark` override that is
 * used only in dark mode. Anything without a brand mark renders as plain text,
 * which is the expected case for concepts like "Schema design".
 *
 * Keys match case-insensitively after a trailing version number is stripped, so
 * "Nuxt", "Nuxt 4" and "nuxt" all resolve to the same entry.
 */
export interface TechIcon {
  icon: string
  /** Brand colour used in light mode. Omit to inherit the surrounding text. */
  color?: string
  /** Override used in dark mode when the brand colour is too dark there. */
  dark?: string
}

const ICONS: Record<string, TechIcon> = {
  // ── Languages and runtimes ────────────────────────────────────────────────
  'go': { icon: 'simple-icons:go', color: '#00ADD8' },
  'php': { icon: 'simple-icons:php', color: '#777BB4', dark: '#9CA0D8' },
  'typescript': { icon: 'simple-icons:typescript', color: '#3178C6', dark: '#4C9BE8' },
  'dart': { icon: 'simple-icons:dart', color: '#0175C2', dark: '#3FA9F5' },
  'css': { icon: 'simple-icons:css3', color: '#1572B6', dark: '#2695D4' },
  'html': { icon: 'simple-icons:html5', color: '#E34C26', dark: '#F06529' },
  // ── Frameworks ───────────────────────────────────────────────────────────
  'laravel': { icon: 'simple-icons:laravel', color: '#FF2D20' },
  'laravel / php': { icon: 'simple-icons:laravel', color: '#FF2D20' },
  'vue': { icon: 'simple-icons:vuedotjs', color: '#35A372', dark: '#4FC08D' },
  'nuxt': { icon: 'simple-icons:nuxt', color: '#00A968', dark: '#00DC82' },
  'flutter': { icon: 'simple-icons:flutter', color: '#0468D7', dark: '#47B6FF' },
  'tailwind css': { icon: 'simple-icons:tailwindcss', color: '#0891B2', dark: '#22D3EE' },
  'bootstrap': { icon: 'simple-icons:bootstrap', color: '#563D7C', dark: '#71509C' },
   'codeigniter': { icon: 'simple-icons:codeigniter', color: '#EF4223', dark: '#FF5C41' },
  // ── Data ─────────────────────────────────────────────────────────────────
  'mysql': { icon: 'simple-icons:mysql', color: '#4479A1', dark: '#7AB3D9' },
  'firebase': { icon: 'simple-icons:firebase', color: '#D9A400', dark: '#FFCA28' },

  // ── Infrastructure ───────────────────────────────────────────────────────
  'docker': { icon: 'simple-icons:docker', color: '#2496ED' },
  'nginx': { icon: 'simple-icons:nginx', color: '#009639', dark: '#2FBF63' },
  'cloudflare': { icon: 'simple-icons:cloudflare', color: '#F38020' },
  'linux': { icon: 'simple-icons:linux', color: '#B58900', dark: '#FCC624' },
  'ubuntu': { icon: 'simple-icons:ubuntu', color: '#E95420' },
  'linux / ubuntu': { icon: 'simple-icons:linux', color: '#B58900', dark: '#FCC624' },
  'railway': { icon: 'simple-icons:railway', color: '#2B2E33', dark: '#E5E7EB' },
  'digitalocean': { icon: 'simple-icons:digitalocean', color: '#0080FF' },
  'cpanel': { icon: 'simple-icons:cpanel', color: '#FF6C2C' },
  'vps': { icon: 'lucide:server' },

  // ── Platforms ────────────────────────────────────────────────────────────
  'android': { icon: 'simple-icons:android', color: '#2E9E4B', dark: '#3DDC84' },
  'ios': { icon: 'simple-icons:apple', color: '#111113', dark: '#F5F5F7' },
  'wordpress': { icon: 'simple-icons:wordpress', color: '#21759B', dark: '#4FA3CC' },
  'elementor': { icon: 'simple-icons:elementor', color: '#92003B', dark: '#E0407F' },
  'elementor pro': { icon: 'simple-icons:elementor', color: '#92003B', dark: '#E0407F' },

  // ── Concepts that earn a generic mark, in the accent colour ──────────────
  'rest': { icon: 'lucide:webhook' },
  'payment apis': { icon: 'lucide:credit-card' },
  'technical seo': { icon: 'lucide:search' },
  'on-page seo': { icon: 'lucide:search' },
  'mobile seo': { icon: 'lucide:smartphone' },
  'structured data': { icon: 'lucide:braces' },
  'sitemap': { icon: 'lucide:list-tree' },
  'performance': { icon: 'lucide:gauge' },
  'ssr / ssg': { icon: 'lucide:file-code' },
  'accessibility': { icon: 'lucide:accessibility' },
  'authentication': { icon: 'lucide:key-round' },
  'authorization': { icon: 'lucide:shield-check' },
  'auth & sessions': { icon: 'lucide:key-round' },
  'migrations': { icon: 'lucide:git-compare-arrows' },
  'backup & recovery': { icon: 'lucide:database-backup' },
  'monitoring': { icon: 'lucide:activity' },
  'logging': { icon: 'lucide:scroll-text' },

}

/** Strips a trailing major version so "Nuxt 4" matches "nuxt". */
function normalise(tech: string): string {
  return tech.trim().toLowerCase().replace(/\s+\d+(\.\d+)*$/, '')
}

/** Full entry for a technology, or undefined when none fits. */
export function techIconMeta(tech: string): TechIcon | undefined {
  return ICONS[normalise(tech)]
}

/** Icon name only, for callers that do not need the colour. */
export function techIcon(tech: string): string | undefined {
  return techIconMeta(tech)?.icon
}

/** Exposed for the test that guards against typos in the map. */
export const techIconMap = ICONS
