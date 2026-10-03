/**
 * Transient confirmations for actions that change something.
 *
 * A save already shows its state next to the button, but a delete, or a save
 * that navigates away, leaves nothing behind to say it worked. These do: one
 * line, top-right, gone on its own.
 *
 * Deliberately not used for validation errors — those belong beside the field
 * that is wrong, where the person is already looking.
 */
export type ToastKind = 'success' | 'error' | 'info'

export interface Toast {
  id: number
  kind: ToastKind
  message: string
  /** Milliseconds before it removes itself; 0 keeps it until dismissed. */
  duration: number
}

/** Errors stay longer: they are read, not just noticed. */
const DEFAULT_DURATION: Record<ToastKind, number> = {
  success: 4000,
  info: 5000,
  error: 8000,
}

export function useToast() {
  const toasts = useState<Toast[]>('toasts', () => [])
  const nextId = useState<number>('toast-id', () => 0)

  function dismiss(id: number) {
    toasts.value = toasts.value.filter((toast) => toast.id !== id)
  }

  function notify(kind: ToastKind, message: string, duration?: number): number {
    const id = nextId.value++
    const toast: Toast = { id, kind, message, duration: duration ?? DEFAULT_DURATION[kind] }

    // Newest first, and capped: a failed bulk save can fire one per row, and a
    // column of twenty identical messages is noise, not information.
    toasts.value = [toast, ...toasts.value].slice(0, 4)

    if (import.meta.client && toast.duration > 0) {
      setTimeout(() => dismiss(id), toast.duration)
    }
    return id
  }

  return {
    toasts,
    dismiss,
    notify,
    success: (message: string, duration?: number) => notify('success', message, duration),
    error: (message: string, duration?: number) => notify('error', message, duration),
    info: (message: string, duration?: number) => notify('info', message, duration),
  }
}
