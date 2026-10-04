import type { NavItem } from '~/types'

/**
 * The primary navigation.
 *
 * Five entries. About was taken out on request: it sits straight under the hero,
 * so it is reached by scrolling, and its section keeps id="about" so links to
 * /#about still land on it. Every entry here must correspond to a section id on
 * the homepage. The scroll spy observes exactly this list, so a section missing
 * from it — About included — is a stretch of page where nothing is highlighted.
 *
 * Labels live in the locale files, not here: this list is the structure of the
 * site, which is the same in every language.
 */
export const navItems: NavItem[] = [
  { key: 'experience', hash: 'experience', section: 'experience' },
  { key: 'work', hash: 'work', section: 'work' },
  { key: 'capabilities', hash: 'capabilities', section: 'capabilities' },
  { key: 'services', hash: 'services', section: 'services' },
  { key: 'contact', hash: 'contact', section: 'contact' },
]

export const sectionIds = navItems.map((i) => i.section)
