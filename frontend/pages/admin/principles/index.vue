<script setup lang="ts">
import type { Principle, SiteContentPayload } from '~/types'

definePageMeta({ layout: 'admin', middleware: 'admin' })

const { loadContent, api } = useAdmin()
const toast = useToast()

const items = ref<Principle[]>([])
const loading = ref(true)
const error = ref('')

async function load() {
  loading.value = true
  try {
    const content: SiteContentPayload = await loadContent()
    items.value = [...content.principles]
  } catch (e) {
    error.value = (e as { message?: string }).message ?? 'Could not load principles.'
  } finally {
    loading.value = false
  }
}

onMounted(load)

/** Reordering saves immediately: a separate save button here would be a trap. */
async function move(index: number, direction: -1 | 1) {
  const list = [...items.value]
  const target = index + direction
  ;[list[index], list[target]] = [list[target]!, list[index]!]
  list.forEach((item, i) => (item.order = i))
  items.value = list

  try {
    await Promise.all(
      list.map((item) => api(`/api/admin/principles/${item.id}`, { method: 'PUT', body: item })),
    )
    toast.success('Order saved.')
  } catch (e) {
    error.value = (e as { message?: string }).message ?? 'Could not save the new order.'
    toast.error(error.value)
    await load()
  }
}

async function remove(item: Principle) {
  try {
    await api(`/api/admin/principles/${item.id}`, { method: 'DELETE' })
    await load()
    toast.success(`Deleted “${item.title}”.`)
  } catch (e) {
    error.value = (e as { message?: string }).message ?? 'Could not delete that entry.'
    toast.error(error.value)
  }
}

useHead({ title: 'Principles' })
</script>

<template>
  <div class="space-y-6">
    <header class="flex flex-wrap items-end justify-between gap-4">
      <div>
        <h1 class="text-2xl font-semibold tracking-tight text-fg">Principles</h1>
        <p class="mt-2 text-sm text-fg-muted">The “Built for Production” cards. Four is the intended size — keep it small.</p>
      </div>
      <UiButton to="/admin/principles/new" size="sm">
        <Icon name="lucide:plus" class="size-3.5" aria-hidden="true" />
        New
      </UiButton>
    </header>

    <AdminEntityList
      :items="items"
      :loading="loading"
      :error="error"
      base-path="/admin/principles"
      label="principles"
      delete-note="The entry and its translations are removed. This cannot be undone."
      @move="move"
      @remove="remove"
    >
      <template #row="{ item }">
        <span class="flex items-center gap-2">
          <Icon :name="item.icon" class="size-4 shrink-0 text-accent" aria-hidden="true" />
          <span class="text-sm font-medium text-fg">{{ item.title }}</span>
        </span>
        <span class="mt-0.5 block font-mono text-xs text-fg-subtle">{{ item.slug }}</span>
      </template>
    </AdminEntityList>
  </div>
</template>
