<script setup lang="ts">
const { t } = useI18n()
const localePath = useLocalePath()
const route = useRoute()
const { projects } = useSiteContent()

const allProjects = projects.value
const project = findProject(allProjects, String(route.params.slug))

// Slugs come from the database now, so an unknown one is a genuine 404 rather
// than a missing entry in a file.
if (!project) {
  throw createError({ statusCode: 404, statusMessage: 'Project not found', fatal: true })
}

const study = computed(() => project!.caseStudy ?? {})

// Next project, so a visitor who reaches the end has somewhere to go.
const next = computed(() => {
  const i = allProjects.findIndex((p) => p.slug === project!.slug)
  return allProjects[(i + 1) % allProjects.length]!
})

const seoDescription = hasContent(project.description)
  ? project.description
  : `${project.title} — ${project.category}. A case study covering what was built and the technical decisions behind it.`

usePageSeo({
  title: project.title,
  description: seoDescription,
  path: `/work/${project.slug}`,
  image: project.image,
  type: 'article',
})
useProjectSchema(project)
</script>

<template>
  <article>
    <!-- Header -->
    <header class="mx-auto max-w-6xl px-6 pt-12 pb-10 sm:pt-16 lg:px-8">
      <nav :aria-label="t('work.breadcrumb')">
        <NuxtLink
          :to="localePath('/work')"
          class="inline-flex items-center gap-1.5 font-mono text-[0.6875rem] tracking-[0.14em] text-fg-subtle uppercase transition-colors hover:text-accent"
        >
          <Icon name="lucide:arrow-left" class="size-3.5" aria-hidden="true" />
          Work
        </NuxtLink>
      </nav>

      <p class="eyebrow mt-8">{{ project!.category }}</p>

      <h1 class="mt-3 text-4xl font-semibold tracking-[-0.03em] text-fg sm:text-5xl">
        {{ project!.title }}
      </h1>

      <UiProse
        :value="project!.summary ?? project!.description"
        file="content/projects.ts"
        class="mt-6 max-w-2xl text-base sm:text-lg"
      />

      <div v-if="project!.liveUrl || project!.githubUrl" class="mt-8 flex flex-wrap items-center gap-3">
        <UiButton v-if="project!.liveUrl" :href="project!.liveUrl" size="sm">
          <Icon name="lucide:external-link" class="size-3.5" aria-hidden="true" />
          Live Website
        </UiButton>
        <UiButton v-if="project!.githubUrl" :href="project!.githubUrl" variant="secondary" size="sm">
          <Icon name="lucide:github" class="size-3.5" aria-hidden="true" />
          GitHub
        </UiButton>
      </div>
    </header>

    <!-- Visual -->
    <div class="mx-auto max-w-6xl px-6 lg:px-8">
      <div class="aspect-16/9 overflow-hidden rounded-lg">
        <NuxtImg
          v-if="project!.image"
          :src="project!.image"
          :alt="project!.imageAlt ?? `${project!.title} — ${project!.category}`"
          class="size-full object-cover"
        />
        <ProjectsThumb v-else :slug="project!.slug" :label="project!.category" class="size-full" />
      </div>
    </div>

    <!-- Body: each block renders only when it holds real content (§15) -->
    <div class="mx-auto max-w-6xl px-6 py-16 sm:py-20 lg:px-8">
      <ProjectsCaseBlock :title="t('work.overview')" :value="study.overview" />
      <ProjectsCaseBlock :title="t('work.context')" :value="study.context" />
      <ProjectsCaseBlock :title="t('work.problem')" :value="study.problem" />
      <ProjectsCaseBlock :title="t('work.role')" :value="study.role" />
      <ProjectsCaseBlock :title="t('work.approach')" :value="study.approach" />
      <ProjectsCaseBlock :title="t('work.architecture')" :value="study.architecture" />

      <section
        v-if="project!.technologies.length"
        class="reveal grid gap-4 border-t border-border py-8 sm:grid-cols-[10rem_1fr] sm:gap-10"
      >
        <h2 class="eyebrow pt-1">{{ t('work.technology') }}</h2>
        <ul class="flex max-w-2xl flex-wrap gap-1.5">
          <li v-for="tech in project!.technologies" :key="tech">
            <UiTag :label="tech" />
          </li>
        </ul>
      </section>

      <ProjectsCaseBlock :title="t('work.features')" :items="study.features" />
      <ProjectsCaseBlock :title="t('work.database')" :value="study.database" />
      <ProjectsCaseBlock :title="t('work.api')" :value="study.api" />
      <ProjectsCaseBlock :title="t('work.security')" :value="study.security" />
      <ProjectsCaseBlock :title="t('work.deployment')" :value="study.deployment" />
      <ProjectsCaseBlock :title="t('work.challenges')" :items="study.challenges" />

      <section
        v-if="study.decisions?.length"
        class="reveal grid gap-4 border-t border-border py-8 sm:grid-cols-[10rem_1fr] sm:gap-10"
      >
        <h2 class="eyebrow pt-1">{{ t('work.technicalDecisions') }}</h2>
        <ul class="max-w-2xl space-y-6">
          <li v-for="decision in study.decisions" :key="decision.decision">
            <template v-if="hasContent(decision.decision)">
              <p class="text-[0.9375rem] font-medium tracking-tight text-fg">{{ decision.decision }}</p>
              <UiProse :value="decision.rationale" file="content/projects.ts" class="mt-2 text-sm" />
            </template>
            <UiPending v-else :value="decision.decision" file="content/projects.ts" />
          </li>
        </ul>
      </section>

      <ProjectsCaseBlock :title="t('work.outcome')" :value="study.outcome" />
    </div>

    <!-- Next -->
    <div class="hairline">
      <div class="mx-auto max-w-6xl px-6 py-12 lg:px-8">
        <NuxtLink :to="localePath(`/work/${next.slug}`)" class="group flex items-center justify-between gap-6">
          <span>
            <span class="eyebrow">{{ t('actions.nextProject') }}</span>
            <span
              class="mt-2 block text-xl font-semibold tracking-tight text-fg transition-colors group-hover:text-accent sm:text-2xl"
            >
              {{ next.title }}
            </span>
          </span>
          <Icon
            name="lucide:arrow-right"
            class="size-5 shrink-0 text-fg-subtle transition-transform duration-200 group-hover:translate-x-1 group-hover:text-accent"
            aria-hidden="true"
          />
        </NuxtLink>
      </div>
    </div>
  </article>
</template>
