import type { Directive } from 'vue'
import { useAuthStore } from '@/stores/auth'

// v-permission directive: removes the element when the current user lacks the
// listed permission(s). Front-end hiding is a convenience only — the backend
// enforces authorization on every request.
export const permissionDirective: Directive<HTMLElement, string | string[]> = {
  mounted(el, binding) {
    const auth = useAuthStore()
    const value = binding.value
    const required = Array.isArray(value) ? value : [value]
    const ok = required.some((p) => auth.hasPermission(p))
    if (!ok) {
      el.parentNode?.removeChild(el)
    }
  },
}
