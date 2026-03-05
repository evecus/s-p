<template>
  <div class="min-h-screen bg-gray-950 flex items-center justify-center p-4">
    <div class="w-full max-w-sm">
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-indigo-600/20 border border-indigo-500/30 mb-4">
          <svg class="w-8 h-8 text-indigo-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/>
          </svg>
        </div>
        <h1 class="text-2xl font-bold text-white">Singbox Panel</h1>
        <p class="text-gray-400 mt-1 text-sm">请登录以继续</p>
      </div>
      <div class="card">
        <div class="space-y-4">
          <div>
            <label class="label">管理密码</label>
            <input v-model="password" type="password" class="input" placeholder="输入密码" @keyup.enter="submit" autofocus />
          </div>
          <p v-if="error" class="text-red-400 text-sm">{{ error }}</p>
          <button class="btn-primary w-full justify-center" :disabled="loading" @click="submit">
            <span v-if="loading">登录中...</span>
            <span v-else>登 录</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()
const password = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    await auth.login(password.value)
    router.push('/dashboard')
  } catch (e) {
    error.value = '密码错误'
  } finally {
    loading.value = false
  }
}
</script>
