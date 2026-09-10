import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const token = ref('')
  const user = ref('')

  function load() {
    token.value = localStorage.getItem('admin_token') || ''
    user.value = localStorage.getItem('admin_user') || ''
  }

  function setToken(t: string) {
    token.value = t
    localStorage.setItem('admin_token', t)
  }

  function setUser(u: string) {
    user.value = u
    localStorage.setItem('admin_user', u)
  }

  function clear() {
    token.value = ''
    user.value = ''
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_user')
  }

  return { token, user, load, setToken, setUser, clear }
})