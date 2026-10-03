<script setup lang="ts" generic="T extends { id?: number }">
/**
 * The shared admin list: position, name, reorder, edit, delete, pager.
 *
 * Extracted from the projects page once experience, skills and principles
 * wanted the same thing. Editing happens on its own page rather than in a stack
 * of open panels — a list is for finding the row you want, and a form is for
 * changing it. It also means a long list stays scannable instead of becoming a
 * column of forms.
 *
 * The row's own wording is a slot, because that is the only part that differs.
 */
const props = defineProps<{
  items: T[]
  /** Where an item is edited: `${basePath}/${id}`. */
  basePath: string
  /** Plural noun for the pager, e.g. "entries". */
  label: string
  loading?: boolean
  error?: string
  emptyText?: string
  /** Extra consequence shown in the delete dialog. */
  deleteNote?: string
}>()

const emit = defineEmits<{
  move: [index: number, direction: -1 | 1]
  remove: [item: T]
}>()

// Named `paged` so it cannot be mistaken for the `items` prop, which stays
// the whole list and is what the reorder bounds are checked against.
const { page, pageSize, total, offset, items: paged } = usePagedList(() => props.items)

/**
 * Reordering works on the position in the whole list, not on the page, and the
 * page follows the row so it does not vanish when it crosses a boundary.
 */
function move(index: number, direction: -1 | 1) {
  const target = index + direction
  if (target < 0 || target >= props.items.length) return
  emit('move', index, direction)
  page.value = Math.floor(target / pageSize.value) + 1
}
</script>

<template>
  <div class="space-y-6">
    <p v-if="error" class="text-sm text-red-600 dark:text-red-400">{{ error }}</p>
    <p v-if="loading" class="text-sm text-fg-muted">Loading…</p>

    <template v-else>
      <p
        v-if="!props.items.length"
        class="rounded-lg border border-dashed border-border-strong p-8 text-center text-sm text-fg-muted"
      >
        {{ emptyText ?? 'Nothing here yet.' }}
      </p>

      <ul v-else class="divide-y divide-border overflow-hidden rounded-lg border border-border">
        <li
          v-for="(item, i) in paged"
          :key="item.id ?? i"
          class="flex flex-wrap items-center gap-4 bg-bg px-5 py-4"
        >
          <span class="w-8 shrink-0 font-mono text-xs text-fg-subtle tabular-nums">
            {{ offset + i + 1 }}
          </span>

          <span class="min-w-0 flex-1">
            <slot name="row" :item="item" :index="offset + i" />
          </span>

          <span class="flex shrink-0 gap-1">
            <button
              type="button"
              class="inline-flex size-8 items-center justify-center rounded-md border border-border text-fg-muted transition-colors hover:text-fg disabled:opacity-40"
              :disabled="offset + i === 0"
              aria-label="Move up"
              @click="move(offset + i, -1)"
            >
              <Icon name="lucide:chevron-up" class="size-4" aria-hidden="true" />
            </button>
            <button
              type="button"
              class="inline-flex size-8 items-center justify-center rounded-md border border-border text-fg-muted transition-colors hover:text-fg disabled:opacity-40"
              :disabled="offset + i === props.items.length - 1"
              aria-label="Move down"
              @click="move(offset + i, 1)"
            >
              <Icon name="lucide:chevron-down" class="size-4" aria-hidden="true" />
            </button>

            <slot name="actions" :item="item" />

            <NuxtLink
              :to="`${basePath}/${item.id}`"
              class="inline-flex size-8 items-center justify-center rounded-md border border-border text-fg-muted transition-colors hover:border-accent hover:text-accent"
              aria-label="Edit"
            >
              <Icon name="lucide:pencil" class="size-3.5" aria-hidden="true" />
            </NuxtLink>

            <AdminConfirmDelete
              :label="(item as { title?: string }).title ?? 'this entry'"
              :note="deleteNote"
              @confirm="emit('remove', item)"
            />
          </span>
        </li>
      </ul>

      <AdminPagination
        v-model:page="page"
        v-model:page-size="pageSize"
        :total="total"
        :label="label"
      />
    </template>
  </div>
</template>
