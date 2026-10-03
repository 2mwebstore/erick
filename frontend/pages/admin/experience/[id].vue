<script setup lang="ts">
import type { SiteContentPayload, TimelineEntry } from '~/types'

definePageMeta({ layout: 'admin', middleware: 'admin' })

const route = useRoute()
const { loadContent, api } = useAdmin()
const toast = useToast()

const isNew = computed(() => route.params.id === 'new')

const blank = (): TimelineEntry => ({
  label: '',
  title: '',
  company: '',
  period: '',
  location: '',
  current: false,
  technologies: [],
  description: '',
  order: 0,
})

const entry = ref<TimelineEntry>(blank())
const loading = ref(true)
const state = ref<'idle' | 'saving' | 'saved' | 'error'>('idle')
const message = ref('')
const errors = ref<Record<string, string>>({})

onMounted(async () => {
  try {
    const content: SiteContentPayload = await loadContent()

    if (isNew.value) {
      // A new entry belongs at the top, where the most recent role lives.
      // One below the lowest existing order gets it there without rewriting
      // every other row on the way in.
      entry.value.order = Math.min(0, ...content.experience.map((e) => e.order ?? 0)) - 1
      return
    }

    const found = content.experience.find((e) => String(e.id) === String(route.params.id))
    if (!found) {
      state.value = 'error'
      message.value = 'That entry no longer exists.'
      return
    }
    entry.value = { ...blank(), ...found }
  } catch (e) {
    if (isNew.value) return
    state.value = 'error'
    message.value = (e as { message?: string }).message ?? 'Could not load that entry.'
  } finally {
    loading.value = false
  }
})

async function save() {
  if (state.value === 'saving') return

  state.value = 'saving'
  errors.value = {}
  message.value = ''

  try {
    const response = await api<{ entry: TimelineEntry }>(
      isNew.value ? '/api/admin/experience' : `/api/admin/experience/${entry.value.id}`,
      { method: isNew.value ? 'POST' : 'PUT', body: entry.value },
    )

    state.value = 'saved'
    message.value = 'Saved.'
    // Creating navigates away from this page, so the inline status disappears
    // with it — the toast is what survives to say it worked.
    toast.success(isNew.value ? 'Created.' : 'Saved.')

    if (isNew.value) await navigateTo(`/admin/experience/${response.entry.id}`)
    else entry.value = { ...blank(), ...response.entry }
  } catch (e) {
    const err = e as { message?: string; errors?: Record<string, string> }
    state.value = 'error'
    errors.value = err.errors ?? {}
    message.value = err.message ?? 'Could not save.'
    toast.error(message.value)
  }
}

useHead({ title: computed(() => (isNew.value ? 'New entry' : entry.value.title || 'Experience')) })
</script>

<template>
  <div class="space-y-6">
    <header class="flex flex-wrap items-end justify-between gap-4">
      <div>
        <NuxtLink
          to="/admin/experience"
          class="inline-flex items-center gap-1.5 font-mono text-[0.6875rem] tracking-[0.14em] text-fg-subtle uppercase transition-colors hover:text-accent"
        >
          <Icon name="lucide:arrow-left" class="size-3.5" aria-hidden="true" />
          Experience
        </NuxtLink>
        <h1 class="mt-3 text-2xl font-semibold tracking-tight text-fg">
          {{ isNew ? 'New entry' : entry.title || 'Untitled' }}
        </h1>
      </div>

      <div class="flex items-center gap-4">
        <AdminStatus :state="state" :message="message" />
        <UiButton size="sm" :disabled="state === 'saving'" @click="save">Save</UiButton>
      </div>
    </header>

    <p v-if="loading" class="text-sm text-fg-muted">Loading…</p>

    <form v-else class="space-y-6" novalidate @submit.prevent="save">
      <AdminPanel title="Role" description="What the entry is, and where it happened.">
        <div class="space-y-5">
          <AdminBilingual
            id="exp-title"
            v-model="entry.title"
            v-model:translations="entry.translations"
            field="title"
            label="Title"
            required
            :error="errors.title"
          />
          <AdminBilingual
            id="exp-company"
            v-model="entry.company"
            v-model:translations="entry.translations"
            field="company"
            label="Company"
            :error="errors.company"
            hint="Leave empty for a career stage rather than a job."
          />
          <AdminBilingual
            id="exp-period"
            v-model="entry.period"
            v-model:translations="entry.translations"
            field="period"
            label="Period"
            :error="errors.period"
            hint="e.g. 2020 — 2023"
          />
          <AdminBilingual
            id="exp-location"
            v-model="entry.location"
            v-model:translations="entry.translations"
            field="location"
            label="Location"
            :error="errors.location"
          />
          <AdminBilingual
            id="exp-label"
            v-model="entry.label"
            v-model:translations="entry.translations"
            field="label"
            label="Label"
            required
            :error="errors.label"
            hint="Shown instead of the company line when there is no company. e.g. Backend"
          />

          <label for="exp-current" class="flex items-center gap-2.5 text-[0.8125rem] text-fg">
            <input
              id="exp-current"
              v-model="entry.current"
              type="checkbox"
              class="size-4 rounded border-border-strong accent-accent"
            >
            Current role — shows the “Current” badge and highlights the card
          </label>
        </div>
      </AdminPanel>

      <AdminPanel title="Detail" description="What you did, and what it was built with.">
        <div class="space-y-5">
          <AdminBilingual
            id="exp-description"
            v-model="entry.description"
            v-model:translations="entry.translations"
            field="description"
            label="Description"
            :rows="4"
            :error="errors.description"
            rich
          />
          <AdminListEditor
            v-model="entry.technologies"
            label="Technologies"
            placeholder="e.g. Laravel"
          />
        </div>
      </AdminPanel>
    </form>
  </div>
</template>
