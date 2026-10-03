<script setup lang="ts">
/**
 * One section of a case study. Renders nothing when the content is absent, and
 * a visible pending marker when it is an unfilled placeholder (§15, §27).
 */
const props = defineProps<{
  title: string
  value?: string | null
  items?: string[] | null
}>()

const showProse = computed(() => hasContent(props.value) || isPending(props.value ?? ''))
const list = computed(() => realItems(props.items))
const pendingItems = computed(() => (props.items ?? []).filter((i) => isPending(i)))
const visible = computed(() => showProse.value || list.value.length > 0 || pendingItems.value.length > 0)
</script>

<template>
  <section v-if="visible" class="reveal grid gap-4 border-t border-border py-8 sm:grid-cols-[10rem_1fr] sm:gap-10">
    <h2 class="eyebrow pt-1">{{ title }}</h2>

    <div class="max-w-2xl space-y-4">
      <UiProse v-if="showProse" :value="props.value" file="content/projects.ts" />

      <ul v-if="list.length" class="space-y-2.5">
        <li v-for="item in list" :key="item" class="flex gap-3 text-[0.9375rem] leading-relaxed text-fg-muted">
          <Icon name="lucide:check" class="mt-1 size-3.5 shrink-0 text-accent" aria-hidden="true" />
          <span>{{ item }}</span>
        </li>
      </ul>

      <UiPending
        v-for="item in pendingItems"
        :key="item"
        :value="item"
        file="content/projects.ts"
      />

      <slot />
    </div>
  </section>
</template>
