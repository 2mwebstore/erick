<script setup lang="ts">
import type { SiteContentPayload } from '~/types'

definePageMeta({ layout: 'admin', middleware: 'admin' })

const { user, loadContent, api } = useAdmin()

const content = ref<SiteContentPayload | null>(null)
const unread = ref(0)
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  try {
    content.value = await loadContent()
    const messages = await api<{ unread: number }>('/api/admin/messages?limit=1')
    unread.value = messages.unread
  } catch (e) {
    error.value = (e as { message?: string }).message ?? 'Could not load the dashboard.'
  } finally {
    loading.value = false
  }
})

const stats = computed(() => {
  const c = content.value
  if (!c) return []

  const drafts = c.projects.filter((p) => !p.published).length

  return [
    { label: 'Projects', value: c.projects.length, hint: drafts ? `${drafts} unpublished` : 'all published', to: '/admin/projects' },
    { label: 'Experience', value: c.experience.length, hint: 'timeline entries', to: '/admin/experience' },
    { label: 'Skills', value: c.capabilities.length, hint: 'capability groups', to: '/admin/capabilities' },
    { label: 'Positioning', value: c.pillars.length, hint: 'hero statements', to: '/admin/pillars' },
    { label: 'Principles', value: c.principles.length, hint: 'production cards', to: '/admin/principles' },
    { label: 'Services', value: c.services.length, hint: 'offerings', to: '/admin/services' },
    { label: 'Unread messages', value: unread.value, hint: 'in the inbox', to: '/admin/messages' },
  ]
})

/**
 * Counts placeholders still visible on the public site. The `TODO:` convention
 * survived the move into the database, so this still works as the content
 * completeness check it was before (§27).
 */
const pending = computed(() => {
  const c = content.value
  if (!c) return []

  const hits: { project: string; field: string }[] = []

  for (const project of c.projects) {
    if (isPending(project.description)) hits.push({ project: project.title, field: 'description' })
    for (const [field, value] of Object.entries(project.caseStudy ?? {})) {
      if (typeof value === 'string' && isPending(value)) {
        hits.push({ project: project.title, field: `caseStudy.${field}` })
      }
      if (Array.isArray(value)) {
        for (const item of value) {
          if (typeof item === 'string' && isPending(item)) {
            hits.push({ project: project.title, field: `caseStudy.${field}` })
          }
        }
      }
    }
  }

  return hits
})

useHead({ title: 'Overview' })
</script>

<template>
  <div class="space-y-8">
    <header>
      <h1 class="text-2xl font-semibold tracking-tight text-fg">
        Welcome back{{ user?.name ? `, ${user.name.split(' ')[0]}` : '' }}
      </h1>
      <p class="mt-2 text-sm text-fg-muted">
        Everything on the public site is edited here. Changes appear within a few minutes —
        pages are cached for five minutes to keep the database off the critical path.
      </p>
    </header>

    <p v-if="loading" class="text-sm text-fg-muted">Loading…</p>
    <p v-else-if="error" class="text-sm text-red-600 dark:text-red-400">{{ error }}</p>

    <template v-else>
      <dl class="grid gap-px overflow-hidden rounded-lg border border-border bg-border sm:grid-cols-2 lg:grid-cols-5">
        <NuxtLink
          v-for="stat in stats"
          :key="stat.label"
          :to="stat.to"
          class="group bg-bg p-5 transition-colors hover:bg-surface"
        >
          <dt class="text-xs font-medium text-fg-muted">{{ stat.label }}</dt>
          <dd class="mt-2 text-2xl font-semibold tracking-tight text-fg tabular-nums">
            {{ stat.value }}
          </dd>
          <dd class="mt-1 text-xs text-fg-subtle">{{ stat.hint }}</dd>
        </NuxtLink>
      </dl>

      <AdminPanel
        title="Content still pending"
        description="Fields marked TODO: render as a visible “content pending” note on the site rather than as text. Nothing is invented to fill them."
      >
        <p v-if="!pending.length" class="text-sm text-fg-muted">
          Nothing pending — every field is filled in.
        </p>
        <ul v-else class="space-y-1.5">
          <li v-for="(hit, i) in pending" :key="i" class="flex items-baseline gap-3 text-sm">
            <span class="font-medium text-fg">{{ hit.project }}</span>
            <code class="font-mono text-xs text-fg-subtle">{{ hit.field }}</code>
          </li>
        </ul>
      </AdminPanel>
    </template>
  </div>
</template>
