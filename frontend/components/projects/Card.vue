<script setup lang="ts">
import type { Project } from '~/types'

/** §14 — editorial case-study preview, not a small tile. */
const props = withDefaults(
  defineProps<{
    project: Project
    index: number
    /**
     * Heading level for the project title. On the homepage the card sits under
     * the "Selected Work" h2, so h3 is correct; on /work the page title is the
     * h1 and the cards are the next level down. Getting this wrong skips a
     * heading level, which screen-reader users navigate by.
     */
    headingLevel?: 'h2' | 'h3'
  }>(),
  { headingLevel: 'h3' },
)

const reversed = computed(() => props.index % 2 === 1)
const hasLinks = computed(() => Boolean(props.project.liveUrl || props.project.githubUrl))

// In the reader's language: a card on /km links to /km/work/…, not the English page.
const localePath = useLocalePath()
const caseStudy = computed(() => localePath(`/work/${props.project.slug}`))
</script>

<template>
  <!-- Cards alternate their entrance with the row they sit in, matching the
       reversed layout so a card always arrives from its own side. -->
  <article
    class="reveal group hairline pt-10 first:border-t-0 first:pt-0 sm:pt-14"
    :class="reversed ? 'reveal-right' : 'reveal-left'"
  >
    <div class="grid gap-8 lg:grid-cols-12 lg:items-center lg:gap-14">
      <!-- Visual -->
      <NuxtLink
        :to="caseStudy"
        class="block lg:col-span-7"
        :class="reversed ? 'lg:order-2' : ''"
        :aria-label="`View the ${project.title} case study`"
        tabindex="-1"
      >
        <div class="aspect-16/10 overflow-hidden rounded-lg">
          <NuxtImg
            v-if="project.image"
            :src="project.image"
            :alt="project.imageAlt ?? `${project.title} — ${project.category}`"
            class="w-full aspect-video object-cover transition-transform duration-500 group-hover:scale-[1.02]"
            loading="lazy"
          />
          <ProjectsThumb v-else :slug="project.slug" :label="project.category" class="size-full" />
        </div>
      </NuxtLink>

      <!-- Copy -->
      <div class="lg:col-span-5" :class="reversed ? 'lg:order-1' : ''">
        <p class="eyebrow">{{ project.category }}</p>

        <component
          :is="headingLevel"
          class="mt-3 text-2xl font-semibold tracking-tight text-fg sm:text-[1.75rem]"
        >
          <NuxtLink
            :to="caseStudy"
            class="transition-colors hover:text-accent focus-visible:text-accent"
          >
            {{ project.title }}
          </NuxtLink>
        </component>

        <UiProse
          :value="project.description"
          file="content/projects.ts"
          class="mt-4 text-[0.9375rem]"
        />

        <ul v-if="project.technologies.length" class="mt-6 flex flex-wrap gap-1.5">
          <li v-for="tech in project.technologies" :key="tech">
            <UiTag :label="tech" />
          </li>
        </ul>

        <div class="mt-7 flex flex-wrap items-center gap-x-6 gap-y-3">
          <NuxtLink
            :to="caseStudy"
            class="inline-flex items-center gap-1.5 text-sm font-medium text-accent transition-colors hover:text-accent-hover"
          >
            View Case Study
            <Icon
              name="lucide:arrow-right"
              class="size-4 transition-transform duration-200 group-hover:translate-x-0.5"
              aria-hidden="true"
            />
          </NuxtLink>

          <template v-if="hasLinks">
            <a
              v-if="project.liveUrl"
              :href="project.liveUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex items-center gap-1.5 text-sm text-fg-muted transition-colors hover:text-fg"
            >
              <Icon name="lucide:external-link" class="size-3.5" aria-hidden="true" />
              Live Website
            </a>
            <a
              v-if="project.githubUrl"
              :href="project.githubUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex items-center gap-1.5 text-sm text-fg-muted transition-colors hover:text-fg"
            >
              <Icon name="lucide:github" class="size-3.5" aria-hidden="true" />
              GitHub
            </a>
          </template>
        </div>
      </div>
    </div>
  </article>
</template>
