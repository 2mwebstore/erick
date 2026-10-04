<script setup lang="ts">
const { t } = useI18n()
const { site, projects: allProjects } = useSiteContent()
const sectionLink = useSectionLink()

const description = `Applications and systems built by ${site.value.name} — e-commerce, mobile, business systems, and web development, with the technical decisions behind each one.`

usePageSeo({ title: 'Work', description, path: '/work' })
useWebPageSchema({ name: 'Work', description, path: '/work' })
</script>

<template>
  <UiSection
    :divider="false"
    wide
    title-tag="h1"
    :eyebrow="t('work.title')"
    :title="t('work.title')"
    subtitle="Applications and systems I've designed, built, and operated. Each case study covers what was built, why, and the technical decisions involved."
  >
    <div class="space-y-14 sm:space-y-20">
      <ProjectsCard
        v-for="(project, i) in allProjects"
        :key="project.slug"
        :project="project"
        :index="i"
        heading-level="h2"
      />
    </div>

    <div class="reveal hairline mt-20 pt-10">
      <p class="text-base text-fg-muted">
        Want the detail behind any of these?
        <NuxtLink :to="sectionLink('contact')" class="text-accent transition-colors hover:text-accent-hover">
          Get in touch
        </NuxtLink>
        and I'll walk you through the architecture.
      </p>
    </div>
  </UiSection>
</template>
