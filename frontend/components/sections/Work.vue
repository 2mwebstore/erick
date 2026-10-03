<script setup lang="ts">
const { t } = useI18n()
const { featuredProjects } = useSiteContent()
</script>

<template>
  <UiSection
    id="work"
    wide
    :eyebrow="t('sections.work.eyebrow')"
    :title="t('sections.work.title')"
    :subtitle="t('sections.work.subtitle')"
  >
    <div class="space-y-14 sm:space-y-20">
      <ProjectsCard
        v-for="(project, i) in featuredProjects"
        :key="project.slug"
        :project="project"
        :index="i"
      />
    </div>

    <div class="reveal mt-16 flex justify-center">
      <UiButton to="/work" variant="secondary">
        All work
        <Icon name="lucide:arrow-right" class="size-4" aria-hidden="true" />
      </UiButton>
    </div>
  </UiSection>
</template>
