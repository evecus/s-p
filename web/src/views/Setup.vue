<template>
  <div class="min-h-screen bg-gray-950 flex items-center justify-center p-4">
    <div class="w-full max-w-sm">
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-indigo-600/20 border border-indigo-500/30 mb-4">
          <svg class="w-8 h-8 text-indigo-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"/>
          </svg>
        </div>
        <h1 class="text-2xl font-bold text-white">Singbox Panel</h1>
        <p class="text-gray-400 mt-1 text-sm">首次使用，请设置管理密码</p>
      </div>
      <div class="card">
        <div class="space-y-4">
          <div>
            <label class="label">设置密码（至少6位）</label>
            <input v-model="password" type="password" class="input" placeholder="请输入密码" @keyup.enter="submit" />
          </div>
          <div>
            <label class="label">确认密码</label>
            <input v-model="confirm" type="password" class="input" placeholder="再次输入密码" @keyup.enter="submit" />
          </div>
          <p v-if="error" class="text-red-400 text-sm">{{ error }}</p>
          <button class="btn-primary w-full justify-center" :disabled="loading" @click="submit">
            <span v-if="loading">设置中...</span>
            <span v-else>完成设置</span>
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
const confirm = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  if (password.value.length < 6) { error.value = '密码至少6位'; return }
  if (password.value !== confirm.value) { error.value = '两次密码不一致'; return }
  loading.value = true
  try {
    await auth.setup(password.value)
    router.push('/dashboard')
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}
</script>
