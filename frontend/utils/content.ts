/**
 * Placeholder handling (§27).
 *
 * Content that has not been verified yet is written as a string prefixed with
 * `TODO:`. Nothing fabricated ever reaches the page — such a string renders as
 * a visible "content pending" note, and `scripts/content-check.mjs` lists them
 * all so none ship by accident.
 */
export const PENDING_PREFIX = 'TODO:'

export function isPending(value?: string | null): boolean {
  return typeof value === 'string' && value.trimStart().startsWith(PENDING_PREFIX)
}

/** The author's note, minus the marker — shown only as editor guidance. */
export function pendingHint(value: string): string {
  return value.trimStart().slice(PENDING_PREFIX.length).trim()
}

/** True when the value is present and is not a placeholder. */
export function hasContent(value?: string | null): value is string {
  return typeof value === 'string' && value.trim().length > 0 && !isPending(value)
}

/** Filters a list down to real entries, dropping placeholders. */
export function realItems(items?: string[] | null): string[] {
  return (items ?? []).filter((i) => hasContent(i))
}

/** A list is renderable only when at least one entry is real. */
export function hasItems(items?: string[] | null): boolean {
  return realItems(items).length > 0
}
