<script setup lang="ts">
const { t } = useI18n()
const { experience } = useSiteContent()

const timeline = computed(() => experience.value)

/**
 * An entry is a job when it names an employer.
 *
 * The same list held career stages before it held employment, and both shapes
 * still render: an entry with a company gets the meta row, one without falls
 * back to its label. Nothing is invented to fill the gap (§27).
 */
const hasMeta = (entry: (typeof timeline.value)[number]) =>
  hasContent(entry.company) || hasContent(entry.period) || hasContent(entry.location)
</script>

<template>
  <UiSection
    v-if="timeline.length"
    id="experience"
    :eyebrow="t('sections.experience.eyebrow')"
    :title="t('sections.experience.title')"
    :subtitle="t('sections.experience.subtitle')"
  >
    <ol class="relative mx-auto max-w-4xl">
      <li
        v-for="(entry, i) in timeline"
        :key="entry.id ?? entry.title"
        class="reveal reveal-left relative grid grid-cols-[1.75rem_1fr] gap-x-5 pb-10 last:pb-0 sm:grid-cols-[2.25rem_1fr] sm:gap-x-7"
        :style="{ '--reveal-delay': `${Math.min(i, 6) * 70}ms` }"
      >
        <!-- Rail: one continuous line, stopped short of the last dot so it
             does not trail off the end of the list. -->
        <span
          v-if="i < timeline.length - 1"
          class="absolute top-4 bottom-0 left-[0.4375rem] w-px bg-gradient-to-b from-accent/45 to-border sm:left-[0.6875rem]"
          aria-hidden="true"
        />

        <span
          class="relative mt-3 size-3.5 rounded-full ring-4 ring-bg sm:mt-3.5 sm:size-4"
          :class="entry.current ? 'bg-accent-solid' : 'bg-border-strong'"
          aria-hidden="true"
        >
          <span
            v-if="entry.current"
            class="absolute inset-0 animate-ping rounded-full bg-accent-solid/60 motion-reduce:hidden"
          />
        </span>

        <article
          class="rounded-xl border p-5 transition-[transform,border-color,background-color] duration-300 sm:p-6"
          :class="
            entry.current
              ? 'border-accent/35 bg-accent/[0.06]'
              : 'border-border bg-surface/40 hover:-translate-y-0.5 hover:border-border-strong'
          "
        >
          <div class="flex flex-wrap items-start justify-between gap-3">
            <h3 class="text-base font-semibold tracking-tight text-fg sm:text-lg">
              {{ entry.title }}
            </h3>
            <span
              v-if="entry.current"
              class="shrink-0 rounded-full border border-accent/30 bg-accent/10 px-2.5 py-1 text-[0.6875rem] font-medium text-accent"
            >
              {{ t('experience.current') }}
            </span>
          </div>

          <p
            v-if="hasMeta(entry)"
            class="mt-2.5 flex flex-wrap items-center gap-x-5 gap-y-1.5 text-[0.8125rem] text-fg-muted"
          >
            <span v-if="hasContent(entry.company)" class="inline-flex items-center gap-1.5">
              <Icon name="lucide:building-2" class="size-3.5 text-fg-subtle" aria-hidden="true" />
              {{ entry.company }}
            </span>
            <span v-if="hasContent(entry.period)" class="inline-flex items-center gap-1.5">
              <Icon name="lucide:calendar" class="size-3.5 text-fg-subtle" aria-hidden="true" />
              {{ entry.period }}
            </span>
            <span v-if="hasContent(entry.location)" class="inline-flex items-center gap-1.5">
              <Icon name="lucide:map-pin" class="size-3.5 text-fg-subtle" aria-hidden="true" />
              {{ entry.location }}
            </span>
          </p>

          <!-- No employer yet: the label still says where this sits. -->
          <p
            v-else-if="hasContent(entry.label)"
            class="mt-2.5 font-mono text-[0.6875rem] tracking-[0.12em] text-fg-subtle uppercase"
          >
            {{ entry.label }}
          </p>

          <UiProse
            v-if="entry.description"
            :value="entry.description"
            file="Admin → Experience"
            class="mt-4 text-sm"
          />

          <ul v-if="entry.technologies?.length" class="mt-4 flex flex-wrap gap-1.5">
            <li v-for="tech in entry.technologies" :key="tech">
              <UiTag :label="tech" />
            </li>
          </ul>
        </article>
      </li>
    </ol>
  </UiSection>
</template>
