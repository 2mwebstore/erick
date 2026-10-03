import { fallbackContent } from '~/content/fallback'
import type { Project, SiteContentPayload } from '~/types'

/**
 * Typed views over the content loaded by plugins/00.site-content.ts.
 *
 * Synchronous on purpose: the plugin has already resolved the payload, so a
 * component can call this in setup without becoming async.
 */
export function useSiteContent() {
  const payload = useState<SiteContentPayload>('site-content', () => fallbackContent)
  const degraded = useState<boolean>('site-content-degraded', () => false)

  const site = computed(() => resolveSite(payload.value.settings))

  const projects = computed<Project[]>(() =>
    [...(payload.value.projects ?? [])].sort((a, b) => (a.order ?? 99) - (b.order ?? 99)),
  )

  return {
    payload,
    site,
    projects,
    featuredProjects: computed(() => projects.value.filter((p) => p.featured)),
    experience: computed(() => payload.value.experience ?? []),
    capabilities: computed(() => payload.value.capabilities ?? []),
    services: computed(() => payload.value.services ?? []),
    principles: computed(() => payload.value.principles ?? []),
    pillars: computed(() => payload.value.pillars ?? []),
    /** True when the page is showing bundled content because the API failed. */
    usingFallback: computed(() => degraded.value),
  }
}

/** Finds a project by slug within an already-loaded list. */
export function findProject(projects: Project[], slug: string): Project | undefined {
  return projects.find((p) => p.slug === slug)
}
