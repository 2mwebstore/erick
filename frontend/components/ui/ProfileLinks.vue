<script setup lang="ts">
import type { Profile } from '~/types'

/**
 * Row of profile links.
 *
 * Each link shows an image now rather than an Iconify name. The icon sets do
 * not carry every platform, and a logo is a picture of a brand, not a glyph the
 * site should be restyling. `icon` is still honoured for links saved before the
 * change, so nothing that already worked stops working; a link with neither
 * falls back to its own initial, which is at least legible.
 */
withDefaults(defineProps<{ profiles: Profile[]; size?: 'sm' | 'md' }>(), { size: 'md' })
</script>

<template>
  <ul v-if="profiles.length" class="flex flex-wrap items-center gap-2.5">
    <li v-for="profile in profiles" :key="profile.href">
      <a
        :href="profile.href"
        target="_blank"
        rel="noopener noreferrer me"
        :title="profile.label"
        :class="[
          'group inline-flex items-center justify-center rounded-lg border border-border bg-surface/40',
          'transition-[transform,border-color,background-color] duration-200',
          'hover:-translate-y-0.5 hover:border-accent/50 hover:bg-surface',
          'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent',
          'motion-reduce:transition-none motion-reduce:hover:translate-y-0',
          size === 'sm' ? 'size-9' : 'size-11',
        ]"
      >
        <span class="sr-only">{{ profile.label }}</span>

        <!--
          Not NuxtImg: these are small remote logos of unpredictable origin, and
          a plain img needs no host on an allowlist to render.
        -->
        <img
          v-if="profile.image"
          :src="profile.image"
          alt=""
          aria-hidden="true"
          loading="lazy"
          decoding="async"
          :class="[
            'object-contain transition-opacity duration-200',
            'opacity-80 group-hover:opacity-100',
            size === 'sm' ? 'size-4' : 'size-5',
          ]"
        >
        <Icon
          v-else-if="profile.icon"
          :name="profile.icon"
          :class="[
            'text-fg-muted transition-colors group-hover:text-fg',
            size === 'sm' ? 'size-4' : 'size-[1.125rem]',
          ]"
          aria-hidden="true"
        />
        <span
          v-else
          class="font-mono text-xs font-medium text-fg-muted group-hover:text-fg"
          aria-hidden="true"
        >
          {{ profile.label.slice(0, 1).toUpperCase() }}
        </span>
      </a>
    </li>
  </ul>
</template>
