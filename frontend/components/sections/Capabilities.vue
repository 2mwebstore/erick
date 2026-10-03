<script setup lang="ts">
const { t } = useI18n()
const { capabilities } = useSiteContent()
</script>

<template>
  <UiSection
    id="capabilities"
    wide
    :eyebrow="t('sections.capabilities.eyebrow')"
    :title="t('sections.capabilities.title')"
    :subtitle="t('sections.capabilities.subtitle')"
  >
    <div class="grid gap-px overflow-hidden rounded-lg border border-border bg-border sm:grid-cols-2 lg:grid-cols-4">
      <article
        v-for="(capability, i) in capabilities"
        :key="capability.slug"
        class="reveal reveal-scale flex flex-col bg-bg p-6 sm:p-7"
        :style="{ '--reveal-delay': `${Math.min(i, 8) * 60}ms` }"
      >
        <Icon :name="capability.icon" class="size-5 text-accent" aria-hidden="true" />

        <h3 class="mt-5 text-base font-semibold tracking-tight text-fg">
          {{ capability.title }}
        </h3>

        <UiProse :value="capability.description" class="mt-2.5 flex-1 text-sm leading-relaxed text-fg-muted" />

        <ul class="mt-5 flex flex-wrap gap-1.5">
          <li v-for="item in capability.items" :key="item">
            <UiTag :label="item" />
          </li>
        </ul>
      </article>
    </div>
  </UiSection>
</template>
