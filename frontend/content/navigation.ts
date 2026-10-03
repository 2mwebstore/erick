import type { NavItem } from '~/types'

/**
 * The primary navigation.
 *
 * §9 asked for five; Work Experience was added as a sixth on request. Every
 * entry here must correspond to a section id on the homepage — the scroll spy
 * observes exactly this list, so a section missing from it is a stretch of page
 * where nothing is highlighted.
 *
 * Labels live in the locale files, not here: this list is the structure of the
 * site, which is the same in every language.
 */
export const navItems: NavItem[] = [
  { key: 'about', hash: 'about', section: 'about' },
  { key: 'experience', hash: 'experience', section: 'experience' },
  { key: 'work', hash: 'work', section: 'work' },
  { key: 'capabilities', hash: 'capabilities', section: 'capabilities' },
  { key: 'services', hash: 'services', section: 'services' },
  { key: 'contact', hash: 'contact', section: 'contact' },
]

export const sectionIds = navItems.map((i) => i.section)
