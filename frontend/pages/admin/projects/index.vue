<script setup lang="ts">
import type { Project, SiteContentPayload } from '~/types'

definePageMeta({ layout: 'admin', middleware: 'admin' })

const { loadContent, api } = useAdmin()
const toast = useToast()

const projects = ref<Project[]>([])
const loading = ref(true)
const error = ref('')


async function load() {
  loading.value = true
  try {
    const content: SiteContentPayload = await loadContent()
    projects.value = [...content.projects].sort((a, b) => (a.order ?? 99) - (b.order ?? 99))
  } catch (e) {
    error.value = (e as { message?: string }).message ?? 'Could not load projects.'
  } finally {
    loading.value = false
  }
}

onMounted(load)

async function remove(project: Project) {
  try {
    await api(`/api/admin/projects/${project.id}`, { method: 'DELETE' })
    await load()
    toast.success(`Deleted “${project.title}”.`)
  } catch (e) {
    error.value = (e as { message?: string }).message ?? 'Could not delete that project.'
    toast.error(error.value)
  }
}

/**
 * Reordering saves immediately: a separate save button here would be a trap.
 *
 * `index` is the position in the whole list, not on the current page, so moving
 * still works when the list is paged.
 */
async function move(index: number, direction: -1 | 1) {
  const target = index + direction
  if (target < 0 || target >= projects.value.length) return

  const list = [...projects.value]
  ;[list[index], list[target]] = [list[target]!, list[index]!]

  // Renumber from 1 so the values stay tidy rather than drifting apart. This is
  // also what pulls a newly created project's negative order back into line.
  list.forEach((p, i) => (p.order = i + 1))
  projects.value = list

  try {
    await Promise.all(
      list.map((p) => api(`/api/admin/projects/${p.id}`, { method: 'PUT', body: p })),
    )
    toast.success('Order saved.')
  } catch (e) {
    error.value = (e as { message?: string }).message ?? 'Could not save the new order.'
    toast.error(error.value)
    await load()
  }
}

useHead({ title: 'Projects' })
</script>

<template>
  <div class="space-y-6">
    <header class="flex flex-wrap items-end justify-between gap-4">
      <div>
        <h1 class="text-2xl font-semibold tracking-tight text-fg">Projects</h1>
        <p class="mt-2 text-sm text-fg-muted">
          Selected Work and every case study. Order here is the order on the site.
        </p>
      </div>
      <UiButton to="/admin/projects/new" size="sm">
        <Icon name="lucide:plus" class="size-3.5" aria-hidden="true" />
        New project
      </UiButton>
    </header>

    <AdminEntityList
      :items="projects"
      :loading="loading"
      :error="error"
      base-path="/admin/projects"
      label="projects"
      delete-note="The project and its case study are removed from the site. This cannot be undone."
      @move="move"
      @remove="remove"
    >
      <template #row="{ item }">
        <NuxtLink
          :to="`/admin/projects/${item.id}`"
          class="text-sm font-medium text-fg transition-colors hover:text-accent"
        >
          {{ item.title }}
        </NuxtLink>
        <span class="mt-0.5 flex flex-wrap items-center gap-x-3 text-xs text-fg-subtle">
          <span>{{ item.category }}</span>
          <span v-if="!item.published" class="text-amber-600 dark:text-amber-400">Draft</span>
          <span v-if="item.featured">Featured</span>
          <span v-if="isPending(item.description)">Content pending</span>
        </span>
      </template>

      <template #actions="{ item }">
        <NuxtLink
          :to="`/work/${item.slug}`"
          target="_blank"
          class="inline-flex size-8 items-center justify-center rounded-md border border-border text-fg-muted transition-colors hover:text-fg"
          :aria-label="`View ${item.title} on the site`"
        >
          <Icon name="lucide:external-link" class="size-3.5" aria-hidden="true" />
        </NuxtLink>
      </template>
    </AdminEntityList>
  </div>
</template>
