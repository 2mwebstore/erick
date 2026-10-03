<script setup lang="ts">
const { choice, isDark, cycle } = useTheme()

const icon = computed(() =>
  choice.value === 'system' ? 'lucide:monitor-smartphone' : isDark.value ? 'lucide:moon' : 'lucide:sun',
)

const label = computed(() => {
  const current = choice.value === 'system' ? `system (${isDark.value ? 'dark' : 'light'})` : choice.value
  return `Theme: ${current}. Activate to change.`
})
</script>

<template>
  <button
    type="button"
    :title="label"
    :aria-label="label"
    class="inline-flex size-9 items-center justify-center rounded-md border border-transparent text-fg-muted transition-colors hover:border-border hover:bg-surface hover:text-fg"
    @click="cycle"
  >
    <Icon :name="icon" class="size-[1.05rem]" aria-hidden="true" />
  </button>
</template>
