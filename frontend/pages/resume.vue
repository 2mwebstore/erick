<script setup lang="ts">
const { t } = useI18n()
const sectionLink = useSectionLink()
const localePath = useLocalePath()
const {
  site,
  capabilities,
  experience: timeline,
  projects: allProjects,
  services,
} = useSiteContent()

const aboutParagraphs = computed(() => site.value.aboutParagraphs)
const wordpressSeoStack = computed(() => site.value.wordpressSeoStack)

const description = `Resume of ${site.value.name}, ${site.value.role} — profile, experience, technical skills, selected projects, and services.`

usePageSeo({ title: 'Resume', description, path: '/resume' })
useWebPageSchema({ name: 'Resume', description, path: '/resume' })

/**
 * §21 — no education, certifications, employers, job titles or awards are
 * listed, because none have been supplied. Add them to the content files when
 * you have them; do not fill this space with approximations.
 */
function print() {
  if (import.meta.client) window.print()
}
</script>

<template>
  <div class="mx-auto max-w-4xl px-6 py-14 sm:py-20 lg:px-8">
    <!-- Header -->
    <header class="flex flex-col gap-8 sm:flex-row sm:items-start sm:justify-between">
      <div>
        <h1 class="text-3xl font-semibold tracking-tight text-fg sm:text-4xl">{{ site.name }}</h1>
        <p class="mt-2 text-base text-fg-muted">{{ site.role }}</p>
        <p class="mt-4 font-mono text-[0.6875rem] tracking-[0.12em] text-fg-subtle uppercase">
          {{ yearRange(site.startYear) }}<template v-if="site.location"> · {{ site.location }}</template>
        </p>
        <a
          :href="`mailto:${site.email}`"
          class="mt-2 inline-block text-sm text-accent transition-colors hover:text-accent-hover"
        >
          {{ site.email }}
        </a>
      </div>

      <div class="flex flex-wrap gap-3 print:hidden">
        <UiButton v-if="site.resumeFile" :href="site.resumeFile!" size="sm" external>
          <Icon name="lucide:download" class="size-3.5" aria-hidden="true" />
          Download Resume
        </UiButton>
        <UiButton variant="secondary" size="sm" @click="print">
          <Icon name="lucide:printer" class="size-3.5" aria-hidden="true" />
          Print / Save as PDF
        </UiButton>
      </div>
    </header>

    <!-- Profile -->
    <section class="hairline mt-12 pt-10">
      <h2 class="eyebrow">{{ t('resume.profile') }}</h2>
      <div class="mt-5 max-w-2xl space-y-4">
        <p v-for="(p, i) in aboutParagraphs" :key="i" class="text-[0.9375rem] leading-relaxed text-fg-muted">
          {{ p }}
        </p>
      </div>
    </section>

    <!-- Experience -->
    <section class="hairline mt-10 pt-10">
      <h2 class="eyebrow">{{ t('resume.experience') }}</h2>
      <p class="mt-4 max-w-2xl text-sm text-fg-subtle">
        A progression of responsibility rather than a list of employers.
      </p>
      <ol class="mt-6 space-y-5">
        <li
          v-for="entry in timeline"
          :key="entry.title"
          class="grid gap-2 sm:grid-cols-[6rem_1fr] sm:gap-6"
        >
          <span class="font-mono text-[0.6875rem] tracking-[0.1em] text-fg-subtle uppercase sm:pt-1">
            {{ entry.label }}
          </span>
          <div>
            <p class="text-[0.9375rem] font-medium tracking-tight text-fg">{{ entry.title }}</p>
            <p v-if="entry.description" class="mt-1 text-sm leading-relaxed text-fg-muted">
              {{ toPlainText(entry.description) }}
            </p>
          </div>
        </li>
      </ol>
    </section>

    <!-- Technical skills -->
    <section class="hairline mt-10 pt-10">
      <h2 class="eyebrow">{{ t('resume.technicalSkills') }}</h2>
      <dl class="mt-6 space-y-5">
        <div v-for="capability in capabilities" :key="capability.slug" class="grid gap-2 sm:grid-cols-[10rem_1fr] sm:gap-6">
          <dt class="text-[0.9375rem] font-medium tracking-tight text-fg">{{ capability.title }}</dt>
          <dd class="text-sm leading-relaxed text-fg-muted">{{ capability.items.join(' · ') }}</dd>
        </div>
        <div class="grid gap-2 sm:grid-cols-[10rem_1fr] sm:gap-6">
          <dt class="text-[0.9375rem] font-medium tracking-tight text-fg">{{ t('work.wordpressSeo') }}</dt>
          <dd class="text-sm leading-relaxed text-fg-muted">{{ wordpressSeoStack.join(' · ') }}</dd>
        </div>
      </dl>
    </section>

    <!-- Selected projects -->
    <section class="hairline mt-10 pt-10">
      <h2 class="eyebrow">{{ t('resume.selectedProjects') }}</h2>
      <ul class="mt-6 space-y-6">
        <li v-for="project in allProjects" :key="project.slug">
          <div class="flex flex-wrap items-baseline gap-x-3">
            <NuxtLink
              :to="localePath(`/work/${project.slug}`)"
              class="text-[0.9375rem] font-medium tracking-tight text-fg transition-colors hover:text-accent"
            >
              {{ project.title }}
            </NuxtLink>
            <span class="font-mono text-[0.6875rem] tracking-[0.1em] text-fg-subtle uppercase">
              {{ project.category }}
            </span>
          </div>
          <p v-if="hasContent(project.description)" class="mt-1.5 max-w-2xl text-sm leading-relaxed text-fg-muted">
            {{ toPlainText(project.description) }}
          </p>
          <p class="mt-1.5 font-mono text-xs text-fg-subtle">{{ project.technologies.join(' · ') }}</p>
        </li>
      </ul>
    </section>

    <!-- Services -->
    <section class="hairline mt-10 pt-10">
      <h2 class="eyebrow">{{ t('nav.services') }}</h2>
      <ul class="mt-6 grid gap-x-10 gap-y-4 sm:grid-cols-2">
        <li v-for="service in services" :key="service.slug">
          <p class="text-[0.9375rem] font-medium tracking-tight text-fg">{{ service.title }}</p>
          <p class="mt-1 text-sm leading-relaxed text-fg-muted">{{ toPlainText(service.description) }}</p>
        </li>
      </ul>
    </section>

    <!-- Contact -->
    <section class="hairline mt-10 pt-10">
      <h2 class="eyebrow">{{ t('nav.contact') }}</h2>
      <ul class="mt-5 space-y-2 text-sm">
        <li>
          <a :href="`mailto:${site.email}`" class="text-accent transition-colors hover:text-accent-hover">
            {{ site.email }}
          </a>
        </li>
        <li v-for="profile in site.profiles" :key="profile.href">
          <a
            :href="profile.href"
            target="_blank"
            rel="noopener noreferrer"
            class="text-fg-muted transition-colors hover:text-accent"
          >
            {{ profile.label }}
          </a>
        </li>
      </ul>

      <div class="mt-8 print:hidden">
        <UiButton :to="sectionLink('contact')" variant="secondary" size="sm">
          Start a conversation
          <Icon name="lucide:arrow-right" class="size-3.5" aria-hidden="true" />
        </UiButton>
      </div>
    </section>
  </div>
</template>

<style>
@media print {
  :root {
    --bg: #ffffff;
    --surface: #ffffff;
    --fg: #000000;
    --fg-muted: #333333;
    --fg-subtle: #555555;
    --border: #cccccc;
    --accent: #000000;
  }

  header[class*='sticky'],
  footer {
    display: none !important;
  }

  a {
    text-decoration: none;
  }
}
</style>
