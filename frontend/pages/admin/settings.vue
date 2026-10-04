<script setup lang="ts">
import type { Profile, SiteContentPayload } from '~/types'

definePageMeta({ layout: 'admin', middleware: 'admin' })

const { loadContent, api } = useAdmin()
const toast = useToast()

const settings = ref<Record<string, string>>({})
// Khmer overlay for the translatable settings. The shape mirrors every other
// entity: locale → field → text, sent back with the save.
const translations = ref<Record<string, Record<string, string>>>({ km: {} })
const heroStack = ref<string[]>([])
const wordpressSeoStack = ref<string[]>([])
const aboutParagraphs = ref<string[]>([])

/**
 * Khmer about paragraphs.
 *
 * A plain list, like the English one, rather than a computed writing straight
 * into `translations`. That version filtered blank rows on every emit, and the
 * list editor adds a row by emitting one — so clicking Add stored nothing, the
 * getter read nothing back, and the new row vanished before it could be typed
 * into. Blanks are dropped on save instead, which is where the English list
 * drops them too.
 */
const aboutParagraphsKm = ref<string[]>([])

const profiles = ref<Profile[]>([])

const loading = ref(true)
const state = ref<'idle' | 'saving' | 'saved' | 'error'>('idle')
const message = ref('')
const errors = ref<Record<string, string>>({})

/** Lists are stored as JSON strings, so they are parsed in and stringified out. */
function parseList<T>(raw: string | undefined, fallback: T[]): T[] {
  if (!raw) return fallback
  try {
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) ? parsed : fallback
  } catch {
    return fallback
  }
}

onMounted(async () => {
  try {
    const content: SiteContentPayload = await loadContent()
    settings.value = { ...content.settings }
    translations.value = { km: {}, ...(content.settingTranslations ?? {}) }
    heroStack.value = parseList<string>(settings.value.hero_stack, [])
    wordpressSeoStack.value = parseList<string>(settings.value.wordpress_seo_stack, [])
    aboutParagraphs.value = parseList<string>(settings.value.about_paragraphs, [])
    aboutParagraphsKm.value = parseList<string>(translations.value.km?.about_paragraphs, [])
    profiles.value = parseList(settings.value.profiles, [])
  } catch (e) {
    state.value = 'error'
    message.value = (e as { message?: string }).message ?? 'Could not load settings.'
  } finally {
    loading.value = false
  }
})

function addProfile() {
  profiles.value = [...profiles.value, { label: '', href: '', image: '' }]
}

function removeProfile(index: number) {
  profiles.value = profiles.value.filter((_, i) => i !== index)
}

async function save() {
  state.value = 'saving'
  errors.value = {}

  // Drop blank profiles so an unfinished row never ships as a dead link.
  const cleanProfiles = profiles.value.filter((p) => p.label.trim() && p.href.trim())
  const khmerParagraphs = aboutParagraphsKm.value.map((p) => p.trim()).filter(Boolean)

  // Settings and their translations are written together: the endpoint takes
  // both so a half-saved language is not possible.
  const body = {
    settings: {
      ...settings.value,
      hero_stack: JSON.stringify(heroStack.value.filter(Boolean)),
      wordpress_seo_stack: JSON.stringify(wordpressSeoStack.value.filter(Boolean)),
      about_paragraphs: JSON.stringify(aboutParagraphs.value.filter(Boolean)),
      profiles: JSON.stringify(cleanProfiles),
    },
    translations: {
      ...translations.value,
      km: {
        ...translations.value.km,
        // Empty means "no translation", which is what lets the English show
        // through; storing "[]" would blank the section instead.
        about_paragraphs: khmerParagraphs.length ? JSON.stringify(khmerParagraphs) : '',
      },
    },
  }

  try {
    await api('/api/admin/settings', { method: 'PUT', body })
    state.value = 'saved'
    message.value = 'Settings saved.'
    toast.success('Settings saved.')
  } catch (e) {
    const err = e as { message?: string; errors?: Record<string, string> }
    state.value = 'error'
    errors.value = err.errors ?? {}
    message.value = err.message ?? 'Could not save settings.'
    toast.error(message.value)
  }
}

useHead({ title: 'Settings' })
</script>

<template>
  <div class="space-y-6">
    <header class="flex flex-wrap items-end justify-between gap-4">
      <div>
        <h1 class="text-2xl font-semibold tracking-tight text-fg">Settings</h1>
        <p class="mt-2 text-sm text-fg-muted">Identity, positioning, and contact details.</p>
      </div>
      <div class="flex items-center gap-4">
        <AdminStatus :state="state" :message="message" />
        <UiButton size="sm" :disabled="state === 'saving'" @click="save">Save</UiButton>
      </div>
    </header>

    <p v-if="loading" class="text-sm text-fg-muted">Loading…</p>

    <form v-else class="space-y-6" novalidate @submit.prevent="save">
      <AdminPanel title="Identity">
        <div class="space-y-4">
          <div class="grid gap-4 sm:grid-cols-2">
            <!--
              Bilingual since the About section heads with it: a Khmer page that
              headed an English name would be the one heading on it still in
              Latin script. Empty Khmer keeps the English, as everywhere else.
            -->
            <AdminBilingual
              id="s-name"
              v-model="settings.name"
              v-model:translations="translations"
              field="name"
              label="Name"
              required
              :error="errors.name"
            />
            <AdminBilingual id="s-role" v-model="settings.role" v-model:translations="translations" field="role" label="Role" required :error="errors.role" />
          </div>
          <div class="grid gap-4 sm:grid-cols-2">
            <AdminInput id="s-email" v-model="settings.email" label="Contact email" type="email" :error="errors.email" />
            <AdminInput id="s-phone" v-model="settings.phone" label="Phone" type="tel" :error="errors.phone" hint="Shown on the contact card. Leave empty to hide the row." />
            <AdminBilingual id="s-location" v-model="settings.location" v-model:translations="translations" field="location" label="Location" :error="errors.location" />
          </div>
          <div class="grid gap-4 sm:grid-cols-2">
            <AdminInput
              id="s-start"
              v-model="settings.start_year"
              label="Started in"
              type="number"
              :error="errors.start_year"
            />
            <AdminInput
              id="s-resume"
              v-model="settings.resume_file"
              label="Resume file"
              :error="errors.resume_file"
              hint="e.g. /resume.pdf — leave empty to offer Print instead."
            />
          </div>
        </div>
      </AdminPanel>

      <AdminPanel title="Positioning" description="The lines that carry the most weight on the homepage.">
        <div class="space-y-4">
          <AdminBilingual
            id="s-headline"
            v-model="settings.headline"
            v-model:translations="translations"
            field="headline"
            label="Hero headline"
            :error="errors.headline"
            hint="The large line at the top of the homepage."
          />
          <AdminBilingual
            id="s-headline-tail"
            v-model="settings.headline_tail"
            v-model:translations="translations"
            field="headline_tail"
            label="Hero headline — quieter half"
            :error="errors.headline_tail"
            hint="Continues the headline in the muted colour. Leave empty for a single-colour headline."
          />
          <AdminBilingual id="s-tagline" v-model="settings.tagline" v-model:translations="translations" field="tagline" label="Tagline" multiline :rows="2" :error="errors.tagline" />
          <AdminBilingual id="s-positioning" v-model="settings.positioning" v-model:translations="translations" field="positioning" label="Positioning" multiline :rows="3" :error="errors.positioning" />
          <AdminBilingual
            id="s-availability"
            v-model="settings.availability"
            v-model:translations="translations"
            field="availability"
            label="Availability note"
            multiline
            :rows="2"
            :error="errors.availability"
            hint="One line under the contact details. Leave empty to hide it."
          />
          <AdminBilingual
            id="s-description"
            v-model="settings.description"
            v-model:translations="translations"
            field="description"
            label="Meta description"
            multiline
            :rows="3"
            :error="errors.description"
            hint="Used as the homepage search-result snippet. Aim for 150–200 characters."
          />

          <div class="grid gap-3 lg:grid-cols-2">
            <AdminListEditor
              v-model="aboutParagraphs"
              label="About paragraphs"
              placeholder="A paragraph of the About section"
            />
            <AdminListEditor
              v-model="aboutParagraphsKm"
              label="About paragraphs (ខ្មែរ)"
              placeholder="កថាខណ្ឌនៃផ្នែកអំពីខ្ញុំ"
            />
          </div>
        </div>
      </AdminPanel>

      <AdminPanel title="Photo">
        <div class="grid gap-4 sm:grid-cols-2">
          <AdminImageField
            id="s-portrait"
            v-model="settings.portrait"
            label="Portrait"
            folder="portraits"
            :error="errors.portrait"
          />
          <AdminInput
            id="s-portrait-alt"
            v-model="settings.portrait_alt"
            label="Portrait alt text"
            :error="errors.portrait_alt"
            hint="Describe the person, not the file."
          />
        </div>
      </AdminPanel>

      <AdminPanel title="Technology lists">
        <div class="space-y-5">
          <AdminListEditor
            v-model="heroStack"
            label="Hero technology line"
            placeholder="e.g. Go"
            hint="Keep it short — this is not the full stack."
          />
          <AdminListEditor
            v-model="wordpressSeoStack"
            label="WordPress & SEO tags"
            placeholder="e.g. Elementor Pro"
          />
        </div>
      </AdminPanel>

      <AdminPanel
        title="Profile links"
        description="Only add a profile once the URL resolves — a dead link costs more credibility than an absent one."
      >
        <ul class="space-y-3">
          <li
            v-for="(profile, i) in profiles"
            :key="i"
            class="grid gap-3 sm:grid-cols-[1fr_2fr_3fr_auto] sm:items-start"
          >
            <AdminInput :id="`pr-label-${i}`" v-model="profile.label" label="Label" />
            <AdminInput :id="`pr-href-${i}`" v-model="profile.href" label="URL" />
            <!-- The field shows its own preview: the fastest way to see a logo address is wrong. -->
            <AdminImageField
              :id="`pr-image-${i}`"
              v-model="profile.image"
              label="Logo"
              folder="profiles"
              fit="contain"
            />
            <div class="flex items-end sm:pt-6">
              <UiButton variant="ghost" size="sm" @click="removeProfile(i)">
                <Icon name="lucide:trash-2" class="size-3.5" aria-hidden="true" />
              </UiButton>
            </div>
          </li>
        </ul>
        <UiButton variant="secondary" size="sm" class="mt-3" @click="addProfile">
          <Icon name="lucide:plus" class="size-3.5" aria-hidden="true" />
          Add profile
        </UiButton>
      </AdminPanel>

      <AdminPanel title="Contact form">
        <label class="flex items-center gap-2 text-sm text-fg">
          <input
            :checked="settings.contact_enabled !== 'false'"
            type="checkbox"
            class="size-4 accent-[var(--accent)]"
            @change="settings.contact_enabled = ($event.target as HTMLInputElement).checked ? 'true' : 'false'"
          >
          Accept messages through the contact form
        </label>
        <p class="mt-2 text-xs text-fg-subtle">
          When off, the section shows your email address instead of the form.
        </p>
      </AdminPanel>

      <div class="flex items-center gap-4">
        <UiButton type="submit" :disabled="state === 'saving'">Save settings</UiButton>
        <AdminStatus :state="state" :message="message" />
      </div>
    </form>
  </div>
</template>
