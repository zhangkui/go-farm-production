import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { AuthAPI, MeAPI, type User, type Role } from '@/api/modules'
import { api } from '@/api/client'

// Auth store: holds the current user, roles and permission set; performs
// login/logout and exposes a hasPermission helper.
export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const roles = ref<Role[]>([])
  const permissions = ref<string[]>([])
  const initialized = ref(false)

  const isLoggedIn = computed(() => !!user.value)
  const fullName = computed(() => user.value?.full_name || user.value?.username || '')

  async function login(username: string, password: string) {
    const res = await AuthAPI.login(username, password)
    api.setSession(res.access_token, res.refresh_token)
    await fetchMe()
  }

  async function fetchMe() {
    const me = await MeAPI.me()
    user.value = me.user
    roles.value = me.roles
    permissions.value = me.permissions
    initialized.value = true
  }

  async function logout() {
    const rt = localStorage.getItem('refresh_token')
    if (rt) {
      try { await AuthAPI.logout(rt) } catch { /* ignore */ }
    }
    api.clearSession()
    user.value = null
    roles.value = []
    permissions.value = []
  }

  function hasPermission(perm: string): boolean {
    if (!perm) return true
    return permissions.value.includes(perm)
  }

  return { user, roles, permissions, initialized, isLoggedIn, fullName, login, fetchMe, logout, hasPermission }
})
