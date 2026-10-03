<script setup lang="ts">
import type { AuditEntry } from '~/types'

definePageMeta({ layout: 'admin', middleware: 'admin' })

const { api, isAdmin } = useAdmin()

const entries = ref<AuditEntry[]>([])
const loading = ref(true)
const error = ref('')

// Paged by the API. The log is append-only and never pruned, so fetching it
// whole would get slower every day the site is used.
const { page, pageSize, total, offset } = usePager()

async function load() {
  loading.value = true
  error.value = ''

  try {
    const res = await api<{ entries: AuditEntry[]; total: number }>(
      `/api/admin/audit?limit=${pageSize.value}&offset=${offset.value}`,
    )
    entries.value = res.entries
    total.value = res.total
  } catch (e) {
    error.value = (e as { message?: string }).message ?? 'Could not load the audit log.'
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch([page, pageSize], load)

const when = (iso: string) =>
  new Date(iso).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' })

useHead({ title: 'Audit log' })
</script>

<template>
  <div class="space-y-6">
    <header>
      <h1 class="text-2xl font-semibold tracking-tight text-fg">Audit log</h1>
      <p class="mt-2 text-sm text-fg-muted">
        Who changed what, and when. Kept even after an account is deleted.
      </p>
    </header>

    <p v-if="!isAdmin" class="text-sm text-fg-muted">This section requires an admin account.</p>
    <p v-else-if="error" class="text-sm text-red-600 dark:text-red-400">{{ error }}</p>
    <p v-else-if="loading" class="text-sm text-fg-muted">Loading…</p>

    <p
      v-else-if="!entries.length"
      class="rounded-lg border border-dashed border-border-strong p-8 text-center text-sm text-fg-muted"
    >
      Nothing recorded yet.
    </p>

    <ul v-else class="divide-y divide-border overflow-hidden rounded-lg border border-border">
      <li v-for="entry in entries" :key="entry.id" class="flex flex-wrap items-baseline gap-x-3 gap-y-1 bg-bg px-5 py-3 text-sm">
        <span class="font-mono text-[0.6875rem] tracking-[0.1em] text-accent uppercase">
          {{ entry.action }}
        </span>
        <span class="text-fg">{{ entry.entity }}</span>
        <code v-if="entry.entity_id" class="font-mono text-xs text-fg-subtle">{{ entry.entity_id }}</code>
        <span class="ml-auto text-xs text-fg-subtle">
          {{ entry.actor_email }} · {{ when(entry.created_at) }}
        </span>
      </li>
    </ul>

    <AdminPagination
      v-if="isAdmin && !loading && !error"
      v-model:page="page"
      v-model:page-size="pageSize"
      :total="total"
      label="entries"
    />
  </div>
</template>
