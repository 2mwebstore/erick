<script setup lang="ts">
import type { Capability, SiteContentPayload } from '~/types'

definePageMeta({ layout: 'admin', middleware: 'admin' })

const route = useRoute()
const { loadContent, api } = useAdmin()
const toast = useToast()

const isNew = computed(() => route.params.id === 'new')

const blank = (): Capability => ({
  slug: '',
  title: '',
  description: '',
  icon: 'lucide:layers',
  items: [],
  order: 0,
})

const item = ref<Capability>(blank())
const loading = ref(true)
const state = ref<'idle' | 'saving' | 'saved' | 'error'>('idle')
const message = ref('')
const errors = ref<Record<string, string>>({})

onMounted(async () => {
  try {
    const content: SiteContentPayload = await loadContent()

    if (isNew.value) {
      // New entries go to the top, where they are visible and get finished.
      item.value.order = Math.min(0, ...content.capabilities.map((c) => c.order ?? 0)) - 1
      return
    }

    const found = content.capabilities.find((c) => String(c.id) === String(route.params.id))
    if (!found) {
      state.value = 'error'
      message.value = 'That entry no longer exists.'
      return
    }
    item.value = { ...blank(), ...found }
  } catch (e) {
    if (isNew.value) return
    state.value = 'error'
    message.value = (e as { message?: string }).message ?? 'Could not load that entry.'
  } finally {
    loading.value = false
  }
})

/** Suggests a slug from the title, but never overwrites one already typed. */
function suggestSlug() {
  if (item.value.slug) return
  item.value.slug = item.value.title
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

async function save() {
  if (state.value === 'saving') return

  state.value = 'saving'
  errors.value = {}
  message.value = ''

  try {
    const response = await api<{ capability: Capability }>(
      isNew.value ? '/api/admin/capabilities' : `/api/admin/capabilities/${item.value.id}`,
      { method: isNew.value ? 'POST' : 'PUT', body: item.value },
    )

    state.value = 'saved'
    message.value = 'Saved.'
    // Creating navigates away from this page, so the inline status disappears
    // with it — the toast is what survives to say it worked.
    toast.success(isNew.value ? 'Created.' : 'Saved.')

    if (isNew.value) await navigateTo(`/admin/capabilities/${response.capability.id}`)
    else item.value = { ...blank(), ...response.capability }
  } catch (e) {
    const err = e as { message?: string; errors?: Record<string, string> }
    state.value = 'error'
    errors.value = err.errors ?? {}
    message.value = err.message ?? 'Could not save.'
    toast.error(message.value)
  }
}

useHead({ title: computed(() => (isNew.value ? 'New' : item.value.title || 'Skills')) })
</script>

<template>
  <div class="space-y-6">
    <header class="flex flex-wrap items-end justify-between gap-4">
      <div>
        <NuxtLink
          to="/admin/capabilities"
          class="inline-flex items-center gap-1.5 font-mono text-[0.6875rem] tracking-[0.14em] text-fg-subtle uppercase transition-colors hover:text-accent"
        >
          <Icon name="lucide:arrow-left" class="size-3.5" aria-hidden="true" />
          Skills
        </NuxtLink>
        <h1 class="mt-3 text-2xl font-semibold tracking-tight text-fg">
          {{ isNew ? 'New entry' : item.title || 'Untitled' }}
        </h1>
      </div>

      <div class="flex items-center gap-4">
        <AdminStatus :state="state" :message="message" />
        <UiButton size="sm" :disabled="state === 'saving'" @click="save">Save</UiButton>
      </div>
    </header>

    <p v-if="loading" class="text-sm text-fg-muted">Loading…</p>

    <form v-else class="space-y-6" novalidate @submit.prevent="save">
      <AdminPanel title="Content">
        <div class="space-y-5">
          <AdminBilingual
            id="e-title"
            v-model="item.title"
            v-model:translations="item.translations"
            field="title"
            label="Title"
            required
            :error="errors.title"
            @focusout="suggestSlug"
          />
          <AdminInput
            id="e-slug"
            v-model="item.slug"
            label="Slug"
            required
            :error="errors.slug"
            hint="Lowercase, hyphens only. Used as the stable key."
          />
          <AdminBilingual
            id="e-description"
            v-model="item.description"
            v-model:translations="item.translations"
            field="description"
            label="Description"
            :rows="3"
            required
            :error="errors.description"
            rich
          />
          <AdminInput
            id="e-icon"
            v-model="item.icon"
            label="Icon"
            required
            :error="errors.icon"
            hint="An Iconify name, e.g. lucide:layers. See icones.js.org."
          />
          <AdminListEditor
            v-model="item.items"
            label="Tools"
            placeholder="e.g. MySQL"
            :error="errors.items"
          />
        </div>
      </AdminPanel>
    </form>
  </div>
</template>
