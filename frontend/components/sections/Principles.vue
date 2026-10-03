<script setup lang="ts">
const { t } = useI18n()
const { principles } = useSiteContent()
</script>

<template>
  <UiSection v-if="principles.length" :eyebrow="t('sections.principles.eyebrow')" :title="t('sections.principles.title')">
    <dl class="grid gap-x-12 gap-y-10 sm:grid-cols-2 lg:grid-cols-4">
      <div
        v-for="(principle, i) in principles"
        :key="principle.id ?? principle.slug"
        class="reveal"
        :style="{ '--reveal-delay': `${i * 70}ms` }"
      >
        <Icon :name="principle.icon" class="size-4.5 text-accent" aria-hidden="true" />
        <dt class="mt-4 text-[0.9375rem] font-semibold tracking-tight text-fg">
          {{ principle.title }}
        </dt>
        <dd class="mt-2"><UiProse :value="principle.description" class="text-sm" /></dd>
      </div>
    </dl>
  </UiSection>
</template>
