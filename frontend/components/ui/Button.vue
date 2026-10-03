<script setup lang="ts">
type Variant = 'primary' | 'secondary' | 'ghost'
type Size = 'md' | 'sm'

const props = withDefaults(
  defineProps<{
    variant?: Variant
    size?: Size
    to?: string
    href?: string
    type?: 'button' | 'submit'
    disabled?: boolean
    external?: boolean
    block?: boolean
  }>(),
  {
    variant: 'primary',
    size: 'md',
    type: 'button',
    to: undefined,
    href: undefined,
    disabled: false,
    external: undefined,
    block: false,
  },
)

const variants: Record<Variant, string> = {
  primary:
    'bg-accent-solid text-accent-fg border-accent-solid hover:bg-accent-solid-hover hover:border-accent-solid-hover',
  secondary:
    'bg-transparent text-fg border-border-strong hover:border-fg hover:bg-surface',
  ghost:
    'bg-transparent text-fg-muted border-transparent hover:text-fg hover:bg-surface',
}

const sizes: Record<Size, string> = {
  md: 'h-11 px-5 text-sm',
  sm: 'h-9 px-3.5 text-[0.8125rem]',
}

const classes = computed(() => [
  'inline-flex items-center justify-center gap-2 rounded-md border font-medium tracking-tight',
  'transition-colors duration-150 select-none',
  'disabled:pointer-events-none disabled:opacity-50',
  variants[props.variant],
  sizes[props.size],
  props.block ? 'w-full' : '',
])

const isExternal = computed(() => props.external ?? /^https?:\/\//.test(props.href ?? ''))
</script>

<template>
  <NuxtLink v-if="to" :to="to" :class="classes">
    <slot />
  </NuxtLink>
  <a
    v-else-if="href"
    :href="href"
    :class="classes"
    :target="isExternal ? '_blank' : undefined"
    :rel="isExternal ? 'noopener noreferrer' : undefined"
  >
    <slot />
  </a>
  <button v-else :type="type" :disabled="disabled" :class="classes">
    <slot />
  </button>
</template>
