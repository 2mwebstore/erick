<script setup lang="ts">
import type { AdminMessage } from '~/types'

definePageMeta({ layout: 'admin', middleware: 'admin' })

const { api } = useAdmin()
const toast = useToast()

const messages = ref<AdminMessage[]>([])
const unread = ref(0)
const loading = ref(true)
const error = ref('')
const view = ref<'inbox' | 'unread' | 'archived'>('inbox')
const expanded = ref<number | null>(null)

// The inbox is paged by the API, not in the browser: it is the one list here
// that grows on its own, from strangers, without an upper bound.
const { page, pageSize, total, offset, clamp } = usePager()

async function load() {
  loading.value = true
  error.value = ''

  const query = new URLSearchParams({
    limit: String(pageSize.value),
    offset: String(offset.value),
  })
  if (view.value === 'unread') query.set('unread', '1')
  if (view.value === 'archived') query.set('archived', '1')

  try {
    const res = await api<{ messages: AdminMessage[]; total: number; unread: number }>(
      `/api/admin/messages?${query}`,
    )
    messages.value = res.messages
    total.value = res.total
    unread.value = res.unread

    // Deleting or archiving the last message on the last page leaves this page
    // past the end of the list. Stepping back changes `page`, and the watcher
    // below refetches — calling load() here as well would double the request.
    clamp()
  } catch (e) {
    error.value = (e as { message?: string }).message ?? 'Could not load messages.'
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch([page, pageSize], load)

// A different view is a different list, so it starts at the beginning. Resetting
// the page already triggers a reload when it was not page 1 to begin with.
watch(view, () => {
  expanded.value = null
  if (page.value === 1) load()
  else page.value = 1
})

async function update(message: AdminMessage, body: Record<string, boolean>) {
  try {
    await api(`/api/admin/messages/${message.id}`, { method: 'PATCH', body })
    await load()
  } catch (e) {
    error.value = (e as { message?: string }).message ?? 'Could not update that message.'
    toast.error(error.value)
  }
}

async function remove(message: AdminMessage) {
  try {
    await api(`/api/admin/messages/${message.id}`, { method: 'DELETE' })
    await load()
    toast.success(`Deleted the message from ${message.name}.`)
  } catch (e) {
    error.value = (e as { message?: string }).message ?? 'Could not delete that message.'
    toast.error(error.value)
  }
}

/** Opening a message marks it read, which is what a reader expects. */
async function toggle(message: AdminMessage) {
  expanded.value = expanded.value === message.id ? null : message.id
  if (expanded.value === message.id && !message.read_at) {
    await update(message, { read: true })
  }
}

const when = (iso: string) =>
  new Date(iso).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' })

useHead({ title: 'Messages' })
</script>

<template>
  <div class="space-y-6">
    <header class="flex flex-wrap items-end justify-between gap-4">
      <div>
        <h1 class="text-2xl font-semibold tracking-tight text-fg">Messages</h1>
        <p class="mt-2 text-sm text-fg-muted">
          {{ total }} in this view<span v-if="unread"> · {{ unread }} unread</span>
        </p>
      </div>

      <div class="flex items-center gap-2">
        <div role="tablist" aria-label="Message view" class="flex rounded-md border border-border p-0.5">
          <button
            v-for="tab in (['inbox', 'unread', 'archived'] as const)"
            :key="tab"
            role="tab"
            :aria-selected="view === tab"
            class="rounded px-3 py-1.5 text-xs font-medium capitalize transition-colors"
            :class="view === tab ? 'bg-surface text-fg' : 'text-fg-muted hover:text-fg'"
            @click="view = tab"
          >
            {{ tab }}
          </button>
        </div>

        <UiButton href="/api/admin/messages/export" variant="secondary" size="sm">
          <Icon name="lucide:download" class="size-3.5" aria-hidden="true" />
          CSV
        </UiButton>
      </div>
    </header>

    <p v-if="error" class="text-sm text-red-600 dark:text-red-400">{{ error }}</p>
    <p v-if="loading" class="text-sm text-fg-muted">Loading…</p>

    <p v-else-if="!messages.length" class="rounded-lg border border-dashed border-border-strong p-8 text-center text-sm text-fg-muted">
      Nothing here yet.
    </p>

    <ul v-else class="divide-y divide-border overflow-hidden rounded-lg border border-border">
      <li v-for="message in messages" :key="message.id" class="bg-bg">
        <button
          type="button"
          class="flex w-full items-start gap-4 px-5 py-4 text-left transition-colors hover:bg-surface/60"
          :aria-expanded="expanded === message.id"
          @click="toggle(message)"
        >
          <span
            class="mt-1.5 size-2 shrink-0 rounded-full"
            :class="message.read_at ? 'bg-border-strong' : 'bg-accent-solid'"
            :aria-label="message.read_at ? 'Read' : 'Unread'"
          />

          <span class="min-w-0 flex-1">
            <span class="flex flex-wrap items-baseline gap-x-3">
              <span class="text-sm font-medium text-fg">{{ message.name }}</span>
              <span class="font-mono text-[0.6875rem] tracking-[0.1em] text-fg-subtle uppercase">
                {{ message.project_type }}
              </span>
            </span>
            <span class="mt-1 block truncate text-sm text-fg-muted">
              {{ message.subject || message.message }}
            </span>
          </span>

          <span class="shrink-0 text-xs whitespace-nowrap text-fg-subtle">{{ when(message.created_at) }}</span>
        </button>

        <div v-if="expanded === message.id" class="border-t border-border bg-surface/40 px-5 py-5">
          <dl class="grid gap-x-8 gap-y-2 text-sm sm:grid-cols-[7rem_1fr]">
            <dt class="text-fg-subtle">Email</dt>
            <dd>
              <a :href="`mailto:${message.email}`" class="text-accent hover:underline">{{ message.email }}</a>
            </dd>
            <dt v-if="message.phone" class="text-fg-subtle">Phone</dt>
            <dd v-if="message.phone">
              <a :href="`tel:${message.phone}`" class="text-accent hover:underline">{{ message.phone }}</a>
            </dd>
            <dt v-if="message.subject" class="text-fg-subtle">Subject</dt>
            <dd v-if="message.subject" class="text-fg">{{ message.subject }}</dd>
            <dt class="text-fg-subtle">Received</dt>
            <dd class="text-fg-muted">{{ when(message.created_at) }}</dd>
            <dt v-if="message.ip_address" class="text-fg-subtle">From</dt>
            <dd v-if="message.ip_address" class="font-mono text-xs text-fg-muted">
              {{ message.ip_address }}
            </dd>
          </dl>

          <p class="mt-4 text-sm leading-relaxed whitespace-pre-wrap text-fg">{{ message.message }}</p>

          <div class="mt-5 flex flex-wrap gap-2">
            <UiButton
              :href="`mailto:${message.email}?subject=Re: your enquiry`"
              size="sm"
            >
              <Icon name="lucide:reply" class="size-3.5" aria-hidden="true" />
              Reply
            </UiButton>
            <UiButton variant="secondary" size="sm" @click="update(message, { read: !message.read_at })">
              Mark {{ message.read_at ? 'unread' : 'read' }}
            </UiButton>
            <UiButton variant="secondary" size="sm" @click="update(message, { archived: !message.archived_at })">
              {{ message.archived_at ? 'Move to inbox' : 'Archive' }}
            </UiButton>
            <AdminConfirmDelete
              variant="button"
              :label="`the message from ${message.name}`"
              note="The enquiry is erased. If you have not replied yet, you will lose the address."
              @confirm="remove(message)"
            />
          </div>
        </div>
      </li>
    </ul>

    <AdminPagination
      v-if="!loading"
      v-model:page="page"
      v-model:page-size="pageSize"
      :total="total"
      label="messages"
    />
  </div>
</template>
