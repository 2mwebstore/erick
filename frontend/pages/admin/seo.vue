<script setup lang="ts">
import type { SeoAuditStatus, SeoCheck } from '~/types'

definePageMeta({ layout: 'admin', middleware: 'admin' })

const { api } = useAdmin()

const status = ref<SeoAuditStatus | null>(null)
const error = ref('')
const starting = ref(false)

const running = computed(() => starting.value || Boolean(status.value?.running))
const report = computed(() => status.value?.report)

async function load() {
  try {
    status.value = await api<SeoAuditStatus>('/api/admin/seo/audit')
    error.value = ''
  } catch (e) {
    error.value = (e as { message?: string }).message ?? 'Could not load the SEO check.'
  }
}

// The crawl runs on the server and takes a few seconds — longer than a request
// may stay open through the proxy — so starting it returns at once and the page
// polls until it is done.
let timer: ReturnType<typeof setTimeout> | undefined
function poll() {
  clearTimeout(timer)
  timer = setTimeout(async () => {
    await load()
    if (status.value?.running) poll()
  }, 1500)
}

async function run() {
  starting.value = true
  error.value = ''
  try {
    status.value = await api<SeoAuditStatus>('/api/admin/seo/audit', { method: 'POST' })
    poll()
  } catch (e) {
    const err = e as { status?: number; message?: string }
    // Someone else started one: follow it rather than reporting a failure.
    if (err.status === 409) poll()
    else error.value = err.message ?? 'Could not start the check.'
  } finally {
    starting.value = false
  }
}

onMounted(async () => {
  await load()
  if (status.value?.running) poll()
})
onBeforeUnmount(() => clearTimeout(timer))

const groups: { id: SeoCheck['group']; title: string; intro: string }[] = [
  {
    id: 'sitelinks',
    title: 'Sitelinks readiness',
    intro: 'What Google reads to understand how the site is organised. Google alone decides whether to show sitelinks; this only checks that nothing stands in the way.',
  },
  { id: 'technical', title: 'Technical', intro: 'Whether every page can be found, fetched and indexed.' },
  { id: 'onpage', title: 'On the page', intro: 'Titles, descriptions, headings and structured data.' },
]
const checksIn = (group: SeoCheck['group']) => report.value?.checks.filter((c) => c.group === group) ?? []

const badge: Record<SeoCheck['status'], { label: string; icon: string; class: string }> = {
  pass: { label: 'Pass', icon: 'lucide:circle-check', class: 'text-emerald-700 dark:text-emerald-400' },
  warn: { label: 'Warning', icon: 'lucide:triangle-alert', class: 'text-amber-700 dark:text-amber-400' },
  fail: { label: 'Problem', icon: 'lucide:circle-x', class: 'text-red-600 dark:text-red-400' },
}

const counts = computed(() => {
  const all = report.value?.checks ?? []
  return {
    fail: all.filter((c) => c.status === 'fail').length,
    warn: all.filter((c) => c.status === 'warn').length,
  }
})

const path = (url: string) => {
  try {
    const u = new URL(url)
    return u.pathname + u.search
  } catch {
    return url
  }
}
const when = (iso: string) => new Date(iso).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' })

useHead({ title: 'SEO check' })
</script>

<template>
  <div class="space-y-8">
    <header class="flex flex-wrap items-start justify-between gap-4">
      <div class="max-w-2xl">
        <h1 class="text-2xl font-semibold tracking-tight text-fg">SEO check</h1>
        <p class="mt-2 text-sm text-fg-muted">
          Crawls
          <code v-if="status?.origin" class="font-mono text-xs">{{ status.origin }}</code>
          <template v-else>the public site</template>
          the way a search engine does: robots.txt, the sitemap, and every page that can be reached by
          following links from the home page. The scores are this site's own diagnostic. They are not
          data from Google, and no score makes Google show sitelinks.
        </p>
      </div>
      <UiButton v-if="status?.configured" :disabled="running" @click="run">
        <Icon
          :name="running ? 'lucide:loader-circle' : 'lucide:scan-search'"
          :class="['size-4', running && 'animate-spin']"
          aria-hidden="true"
        />
        {{ running ? 'Checking…' : report ? 'Run again' : 'Run check' }}
      </UiButton>
    </header>

    <p v-if="error" role="alert" class="text-sm text-red-600 dark:text-red-400">{{ error }}</p>

    <p v-if="!status && !error" class="text-sm text-fg-muted">Loading…</p>

    <p
      v-else-if="status && !status.configured"
      class="rounded-lg border border-dashed border-border-strong p-6 text-sm text-fg-muted"
    >
      The check is off: set <code class="font-mono text-xs">SITE_URL</code> on the API to the public
      site address, for example <code class="font-mono text-xs">https://kongchansila.com</code>.
    </p>

    <p v-if="status?.error" role="alert" class="text-sm text-red-600 dark:text-red-400">{{ status.error }}</p>

    <p
      v-if="status?.configured && !report && !running && !status.error"
      class="rounded-lg border border-dashed border-border-strong p-8 text-center text-sm text-fg-muted"
    >
      No check has run since the API last started. Run one to see the results.
    </p>

    <template v-if="report">
      <section aria-label="Scores" class="grid gap-4 sm:grid-cols-3">
        <div class="rounded-lg border border-border bg-surface p-5">
          <p class="text-xs tracking-wide text-fg-muted uppercase">Overall</p>
          <p class="mt-1 text-3xl font-semibold text-fg tabular-nums">{{ report.score }}<span class="text-base text-fg-muted">/100</span></p>
        </div>
        <div class="rounded-lg border border-border bg-surface p-5">
          <p class="text-xs tracking-wide text-fg-muted uppercase">Sitelinks readiness</p>
          <p class="mt-1 text-3xl font-semibold text-fg tabular-nums">{{ report.sitelinksScore }}<span class="text-base text-fg-muted">/100</span></p>
        </div>
        <div class="rounded-lg border border-border bg-surface p-5 text-sm text-fg-muted">
          <p>
            <strong class="text-fg">{{ counts.fail }}</strong> problem(s) ·
            <strong class="text-fg">{{ counts.warn }}</strong> warning(s)
          </p>
          <p class="mt-1">{{ report.pages.length }} pages in {{ (report.durationMs / 1000).toFixed(1) }} s</p>
          <p class="mt-1">Checked {{ when(report.startedAt) }}</p>
        </div>
      </section>

      <p v-if="report.truncated" class="text-sm text-amber-700 dark:text-amber-400">
        The crawl stopped at its page limit, so some pages were not checked.
      </p>

      <section v-for="group in groups" :key="group.id" :aria-labelledby="`seo-${group.id}`" class="space-y-3">
        <div>
          <h2 :id="`seo-${group.id}`" class="text-lg font-semibold text-fg">{{ group.title }}</h2>
          <p class="mt-1 text-sm text-fg-muted">{{ group.intro }}</p>
        </div>
        <ul class="divide-y divide-border overflow-hidden rounded-lg border border-border">
          <li v-for="check in checksIn(group.id)" :key="check.id" class="bg-bg px-5 py-4" :data-check="check.id">
            <div class="flex items-start gap-3">
              <span :class="['inline-flex shrink-0 items-center gap-1.5 text-xs font-medium', badge[check.status].class]">
                <Icon :name="badge[check.status].icon" class="size-4" aria-hidden="true" />
                <span class="w-16">{{ badge[check.status].label }}</span>
              </span>
              <div class="min-w-0 flex-1">
                <p class="text-sm font-medium text-fg">{{ check.title }}</p>
                <p class="mt-1 text-sm text-fg-muted">{{ check.detail }}</p>
                <details v-if="check.urls?.length" class="mt-2 text-xs">
                  <summary class="cursor-pointer text-accent">{{ check.urls.length }} item(s)</summary>
                  <ul class="mt-2 space-y-1 font-mono break-all text-fg-muted">
                    <li v-for="item in check.urls" :key="item">{{ item }}</li>
                  </ul>
                </details>
              </div>
            </div>
          </li>
        </ul>
      </section>

      <details class="rounded-lg border border-border">
        <summary class="cursor-pointer px-5 py-3 text-sm font-medium text-fg">
          Every page crawled ({{ report.pages.length }})
        </summary>
        <div class="overflow-x-auto border-t border-border">
          <table class="w-full text-left text-xs">
            <thead class="text-fg-muted">
              <tr>
                <th scope="col" class="px-4 py-2 font-medium">Page</th>
                <th scope="col" class="px-4 py-2 font-medium">Status</th>
                <th scope="col" class="px-4 py-2 font-medium">Clicks from home</th>
                <th scope="col" class="px-4 py-2 font-medium">Links in</th>
                <th scope="col" class="px-4 py-2 font-medium">Indexable</th>
                <th scope="col" class="px-4 py-2 font-medium">In sitemap</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border">
              <tr v-for="page in report.pages" :key="page.url">
                <td class="px-4 py-2 font-mono break-all text-fg">{{ path(page.url) }}</td>
                <td class="px-4 py-2 tabular-nums" :class="page.status === 200 ? 'text-fg-muted' : 'text-red-600 dark:text-red-400'">
                  {{ page.status || 'error' }}
                </td>
                <td class="px-4 py-2 tabular-nums text-fg-muted">{{ page.depth < 0 ? 'no link' : page.depth }}</td>
                <td class="px-4 py-2 tabular-nums text-fg-muted">{{ page.incoming }}</td>
                <td class="px-4 py-2 text-fg-muted">{{ page.indexable ? 'Yes' : 'No' }}</td>
                <td class="px-4 py-2 text-fg-muted">{{ page.inSitemap ? 'Yes' : '—' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </details>
    </template>
  </div>
</template>
