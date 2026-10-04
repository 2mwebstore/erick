import type { UploadSettings } from '~/types'

let pending: Promise<void> | null = null

/**
 * Whether image fields can upload (R2 configured on the API) and what they
 * accept. Asked once per session however many image fields are on the page;
 * until the answer arrives, and if asking fails, fields offer a link only.
 */
export function useUploadSettings() {
  const state = useState<UploadSettings | null>('admin-upload-settings', () => null)
  const { api } = useAdmin()

  if (!state.value && !pending && import.meta.client) {
    pending = api<UploadSettings>('/api/admin/uploads')
      .then((settings) => {
        state.value = settings
      })
      .catch(() => {
        state.value = { enabled: false }
      })
      .finally(() => {
        pending = null
      })
  }

  return computed<UploadSettings>(() => state.value ?? { enabled: false })
}
