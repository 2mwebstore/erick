<script setup lang="ts">
import type { Project, SiteContentPayload, TechnicalDecision } from '~/types'

definePageMeta({ layout: 'admin', middleware: 'admin' })

const route = useRoute()
const { loadContent, api } = useAdmin()
const toast = useToast()

const isNew = computed(() => route.params.id === 'new')

const blank = (): Project => ({
  slug: '',
  title: '',
  category: '',
  description: '',
  summary: '',
  technologies: [],
  featured: true,
  published: false, // a new project starts as a draft, not live
  order: 0, // replaced on mount with a slot above every existing project
  caseStudy: {
    overview: '', context: '', problem: '', role: '', approach: '', architecture: '',
    database: '', api: '', security: '', deployment: '', outcome: '',
    features: [], challenges: [], decisions: [],
  },
})

const project = ref<Project>(blank())
const loading = ref(true)
const state = ref<'idle' | 'saving' | 'saved' | 'error'>('idle')
const message = ref('')
const errors = ref<Record<string, string>>({})

onMounted(async () => {
  try {
    const content: SiteContentPayload = await loadContent()

    if (isNew.value) {
      // A new project belongs at the top of the list, where it can be seen and
      // finished. Taking one below the lowest existing order gets it there
      // without rewriting every other project's order on the way in; the list's
      // move controls renumber from 1 the first time anything is reordered.
      project.value.order = Math.min(1, ...content.projects.map((p) => p.order ?? 99)) - 1
      return
    }

    const found = content.projects.find((p) => String(p.id) === String(route.params.id))
    if (!found) {
      message.value = 'That project no longer exists.'
      state.value = 'error'
      return
    }
    // Ensure every case-study field exists so v-model has something to bind to.
    project.value = { ...blank(), ...found, caseStudy: { ...blank().caseStudy, ...found.caseStudy } }
  } catch (e) {
    // A new project must still be creatable if the list could not be read; it
    // just lands at order 0 instead of strictly first.
    if (isNew.value) return
    state.value = 'error'
    message.value = (e as { message?: string }).message ?? 'Could not load that project.'
  } finally {
    loading.value = false
  }
})

/** Suggests a slug from the title, but never overwrites one already typed. */
function suggestSlug() {
  if (project.value.slug) return
  project.value.slug = project.value.title
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

const decisions = computed({
  get: () => project.value.caseStudy?.decisions ?? [],
  set: (value: TechnicalDecision[]) => {
    project.value.caseStudy = { ...project.value.caseStudy, decisions: value }
  },
})

function addDecision() {
  decisions.value = [...decisions.value, { decision: '', rationale: '' }]
}

function removeDecision(index: number) {
  decisions.value = decisions.value.filter((_, i) => i !== index)
}

async function save() {
  if (state.value === 'saving') return

  state.value = 'saving'
  errors.value = {}
  message.value = ''

  try {
    const response = await api<{ project: Project }>(
      isNew.value ? '/api/admin/projects' : `/api/admin/projects/${project.value.id}`,
      { method: isNew.value ? 'POST' : 'PUT', body: project.value },
    )

    state.value = 'saved'
    message.value = 'Saved.'
    // Creating navigates away from this page, so the inline status disappears
    // with it — the toast is what survives to say it worked.
    toast.success(isNew.value ? 'Created.' : 'Saved.')

    if (isNew.value) {
      await navigateTo(`/admin/projects/${response.project.id}`)
    } else {
      project.value = { ...blank(), ...response.project, caseStudy: { ...blank().caseStudy, ...response.project.caseStudy } }
    }
  } catch (e) {
    const err = e as { message?: string; errors?: Record<string, string> }
    state.value = 'error'
    errors.value = err.errors ?? {}
    message.value = err.message ?? 'Could not save.'
    toast.error(message.value)
  }
}

useHead({ title: computed(() => (isNew.value ? 'New project' : project.value.title || 'Project')) })
</script>

<template>
  <div class="space-y-6">
    <header class="flex flex-wrap items-end justify-between gap-4">
      <div>
        <NuxtLink
          to="/admin/projects"
          class="inline-flex items-center gap-1.5 font-mono text-[0.6875rem] tracking-[0.14em] text-fg-subtle uppercase transition-colors hover:text-accent"
        >
          <Icon name="lucide:arrow-left" class="size-3.5" aria-hidden="true" />
          Projects
        </NuxtLink>
        <h1 class="mt-3 text-2xl font-semibold tracking-tight text-fg">
          {{ isNew ? 'New project' : project.title || 'Untitled' }}
        </h1>
      </div>

      <div class="flex items-center gap-4">
        <AdminStatus :state="state" :message="message" />
        <UiButton size="sm" :disabled="state === 'saving'" @click="save">Save</UiButton>
      </div>
    </header>

    <p v-if="loading" class="text-sm text-fg-muted">Loading…</p>

    <form v-else class="space-y-6" novalidate @submit.prevent="save">
      <AdminPanel title="Card" description="What appears in Selected Work and on the work index.">
        <div class="space-y-5">
          <div class="grid gap-5 sm:grid-cols-2">
            <AdminBilingual
              id="p-title"
              v-model="project.title"
              v-model:translations="project.translations"
              field="title"
              label="Title"
              required
              :error="errors.title"
              class="sm:col-span-2"
              @focusout="suggestSlug"
            />
            <AdminInput
              id="p-slug"
              v-model="project.slug"
              label="Slug"
              required
              :error="errors.slug"
              hint="Becomes /work/<slug>. Lowercase, hyphens only."
            />
          </div>

          <AdminBilingual
            id="p-category"
            v-model="project.category"
            v-model:translations="project.translations"
            field="category"
            label="Category"
            required
            :error="errors.category"
            hint="e.g. Restaurant / POS System"
          />

          <AdminBilingual
            id="p-description"
            v-model="project.description"
            v-model:translations="project.translations"
            field="description"
            label="Description"
            :rows="3"
            :error="errors.description"
            hint="One sentence. Used on the card and as the page meta description. Prefix with TODO: to show a “content pending” note instead."
            rich
          />

          <AdminBilingual
            id="p-summary"
            v-model="project.summary"
            v-model:translations="project.translations"
            field="summary"
            label="Summary"
            :rows="4"
            :error="errors.summary"
            hint="Longer lede at the top of the case study. Optional."
            rich
          />

          <AdminListEditor
            v-model="project.technologies"
            label="Technologies"
            placeholder="e.g. Flutter"
            :error="errors.technologies"
            hint="Leave empty if the stack is not confirmed — better an empty list than a guessed one."
          />

          <div class="grid gap-5 sm:grid-cols-2">
            <AdminInput
              id="p-live"
              v-model="project.liveUrl"
              label="Live URL"
              :error="errors.liveUrl"
              hint="Only add it once the address actually resolves."
            />
            <AdminInput
              id="p-github"
              v-model="project.githubUrl"
              label="GitHub URL"
              :error="errors.githubUrl"
            />
          </div>

          <div class="grid gap-5 sm:grid-cols-2">
            <AdminImageField
              id="p-image"
              v-model="project.image"
              label="Image"
              folder="projects"
              :error="errors.image"
            />
            <AdminInput
              id="p-image-alt"
              v-model="project.imageAlt"
              label="Image alt text"
              :error="errors.imageAlt"
              hint="Describe the screenshot, not the file."
            />
          </div>

          <div class="flex flex-wrap items-center gap-6 border-t border-border pt-5">
            <label class="flex items-center gap-2 text-sm text-fg">
              <input v-model="project.published" type="checkbox" class="size-4 accent-[var(--accent)]">
              Published
            </label>
            <label class="flex items-center gap-2 text-sm text-fg">
              <input v-model="project.featured" type="checkbox" class="size-4 accent-[var(--accent)]">
              Show on homepage
            </label>
            <label class="flex items-center gap-2 text-sm text-fg">
              Order
              <input
                v-model.number="project.order"
                type="number"
                min="1"
                class="w-16 rounded-md border border-border-strong bg-bg px-2 py-1 text-sm tabular-nums"
              >
            </label>
          </div>
        </div>
      </AdminPanel>

      <AdminPanel
        title="Case study"
        description="Every section renders only when it holds real text. Leave a field empty to hide it entirely; prefix with TODO: to show a visible pending note."
      >
        <div v-if="project.caseStudy" class="space-y-5">
          <AdminBilingual
            id="cs-overview"
            v-model="project.caseStudy.overview"
            v-model:translations="project.translations"
            field="caseStudy.overview"
            label="Overview"
            :error="errors['caseStudy.overview']"
            rich
          />
          <AdminBilingual
            id="cs-context"
            v-model="project.caseStudy.context"
            v-model:translations="project.translations"
            field="caseStudy.context"
            label="Context"
            :error="errors['caseStudy.context']"
            rich
          />
          <AdminBilingual
            id="cs-problem"
            v-model="project.caseStudy.problem"
            v-model:translations="project.translations"
            field="caseStudy.problem"
            label="Problem"
            :error="errors['caseStudy.problem']"
            rich
          />
          <AdminBilingual
            id="cs-role"
            v-model="project.caseStudy.role"
            v-model:translations="project.translations"
            field="caseStudy.role"
            label="My role"
            :error="errors['caseStudy.role']"
            rich
          />
          <AdminBilingual
            id="cs-approach"
            v-model="project.caseStudy.approach"
            v-model:translations="project.translations"
            field="caseStudy.approach"
            label="Approach"
            :error="errors['caseStudy.approach']"
            rich
          />
          <AdminBilingual
            id="cs-architecture"
            v-model="project.caseStudy.architecture"
            v-model:translations="project.translations"
            field="caseStudy.architecture"
            label="Architecture"
            :error="errors['caseStudy.architecture']"
            rich
          />

          <AdminListEditor
            v-model="project.caseStudy.features"
            label="Key features"
            placeholder="A shipped feature"
            hint="Only functionality that is actually live."
          />

          <AdminBilingual
            id="cs-database"
            v-model="project.caseStudy.database"
            v-model:translations="project.translations"
            field="caseStudy.database"
            label="Database"
            :error="errors['caseStudy.database']"
            rich
          />
          <AdminBilingual
            id="cs-api"
            v-model="project.caseStudy.api"
            v-model:translations="project.translations"
            field="caseStudy.api"
            label="API"
            :error="errors['caseStudy.api']"
            rich
          />
          <AdminBilingual
            id="cs-security"
            v-model="project.caseStudy.security"
            v-model:translations="project.translations"
            field="caseStudy.security"
            label="Security"
            :error="errors['caseStudy.security']"
            rich
          />
          <AdminBilingual
            id="cs-deployment"
            v-model="project.caseStudy.deployment"
            v-model:translations="project.translations"
            field="caseStudy.deployment"
            label="Deployment"
            :error="errors['caseStudy.deployment']"
            rich
          />

          <AdminListEditor
            v-model="project.caseStudy.challenges"
            label="Challenges"
            placeholder="A real constraint you hit"
          />

          <fieldset>
            <legend class="text-[0.8125rem] font-medium text-fg">Technical decisions</legend>
            <p class="mt-1 text-xs text-fg-subtle">
              What you chose, and what you traded away for it.
            </p>

            <ul class="mt-3 space-y-4">
              <li
                v-for="(decision, i) in decisions"
                :key="i"
                class="rounded-md border border-border p-4"
              >
                <div class="space-y-3">
                  <AdminInput
                    :id="`decision-${i}`"
                    :model-value="decision.decision"
                    label="Decision"
                    @update:model-value="decision.decision = $event"
                  />
                  <AdminInput
                    :id="`rationale-${i}`"
                    :model-value="decision.rationale"
                    label="Rationale"
                    multiline
                    :rows="3"
                    @update:model-value="decision.rationale = $event"
                  />
                </div>
                <UiButton variant="ghost" size="sm" class="mt-3" @click="removeDecision(i)">
                  <Icon name="lucide:trash-2" class="size-3.5" aria-hidden="true" />
                  Remove
                </UiButton>
              </li>
            </ul>

            <UiButton variant="secondary" size="sm" class="mt-3" @click="addDecision">
              <Icon name="lucide:plus" class="size-3.5" aria-hidden="true" />
              Add decision
            </UiButton>
          </fieldset>

          <AdminInput
            id="cs-outcome"
            v-model="project.caseStudy.outcome"
            label="Outcome"
            multiline
            :error="errors['caseStudy.outcome']"
            hint="Real, citable results only — no invented metrics."
          />
        </div>
      </AdminPanel>

      <div class="flex items-center gap-4">
        <UiButton type="submit" :disabled="state === 'saving'">Save project</UiButton>
        <AdminStatus :state="state" :message="message" />
      </div>
    </form>
  </div>
</template>
