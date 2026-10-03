<script setup lang="ts">
import type { SiteContentPayload, TimelineEntry } from '~/types'

definePageMeta({ layout: 'admin', middleware: 'admin' })

const { loadContent, api } = useAdmin()
const toast = useToast()

const entries = ref<TimelineEntry[]>([])
const loading = ref(true)
const error = ref('')

async function load() {
  loading.value = true
  try {
    const content: SiteContentPayload = await loadContent()
    entries.value = [...content.experience]
  } catch (e) {
    error.value = (e as { message?: string }).message ?? 'Could not load experience.'
  } finally {
    loading.value = false
  }
}

onMounted(load)

/** Reordering saves immediately: a separate save button here would be a trap. */
async function move(index: number, direction: -1 | 1) {
  const list = [...entries.value]
  const target = index + direction
  ;[list[index], list[target]] = [list[target]!, list[index]!]
  list.forEach((entry, i) => (entry.order = i))
  entries.value = list

  try {
    await Promise.all(
      list.map((entry) =>
        api(`/api/admin/experience/${entry.id}`, { method: 'PUT', body: entry }),
      ),
    )
    toast.success('Order saved.')
  } catch (e) {
    error.value = (e as { message?: string }).message ?? 'Could not save the new order.'
    toast.error(error.value)
    await load()
  }
}

async function remove(entry: TimelineEntry) {
  try {
    await api(`/api/admin/experience/${entry.id}`, { method: 'DELETE' })
    await load()
    toast.success(`Deleted “${entry.title}”.`)
  } catch (e) {
    error.value = (e as { message?: string }).message ?? 'Could not delete that entry.'
    toast.error(error.value)
  }
}

useHead({ title: 'Experience' })
</script>

<template>
  <div class="space-y-6">
    <header class="flex flex-wrap items-end justify-between gap-4">
      <div>
        <h1 class="text-2xl font-semibold tracking-tight text-fg">Experience</h1>
        <p class="mt-2 text-sm text-fg-muted">
          The Work Experience timeline. Order here is the order on the site.
        </p>
      </div>
      <UiButton to="/admin/experience/new" size="sm">
        <Icon name="lucide:plus" class="size-3.5" aria-hidden="true" />
        New entry
      </UiButton>
    </header>

    <AdminEntityList
      :items="entries"
      :loading="loading"
      :error="error"
      base-path="/admin/experience"
      label="entries"
      empty-text="No experience yet. Add the first entry."
      delete-note="The entry and its translations are removed. This cannot be undone."
      @move="move"
      @remove="remove"
    >
      <template #row="{ item }">
        <span class="block text-sm font-medium text-fg">{{ item.title }}</span>
        <span class="mt-0.5 flex flex-wrap items-center gap-x-3 text-xs text-fg-subtle">
          <span v-if="item.company">{{ item.company }}</span>
          <span v-if="item.period">{{ item.period }}</span>
          <span v-if="item.current" class="text-accent">Current</span>
          <span v-if="!item.company" class="font-mono">{{ item.label }}</span>
        </span>
      </template>
    </AdminEntityList>
  </div>
</template>
