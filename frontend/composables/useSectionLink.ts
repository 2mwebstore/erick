/**
 * Link to a homepage section in the language being shown: /#work in English,
 * /km#work in Khmer.
 *
 * Not /km/#work: the home page has no trailing slash, so that form costs a 301
 * on every full page load. And not a hard-coded /#contact, which sent a reader
 * on a Khmer page to the English home page.
 */
export function useSectionLink() {
  const localePath = useLocalePath()
  return (hash: string) => `${localePath('/')}#${hash}`
}
