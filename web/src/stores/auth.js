import { defineStore } from 'pinia'
import { ref } from 'vue'
import { authApi } from '@/api'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('sb_token') || '')

  async function checkSetup() {
    return await authApi.status()
  }

  async function setup(password) {
    const res = await authApi.setup(password)
    token.value = res.token
    localStorage.setItem('sb_token', res.token)
  }

  async function login(password) {
    const res = await authApi.login(password)
    token.value = res.token
    localStorage.setItem('sb_token', res.token)
  }

  function logout() {
    token.value = ''
    localStorage.removeItem('sb_token')
  }

  return { token, checkSetup, setup, login, logout }
})
