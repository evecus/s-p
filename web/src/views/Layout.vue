<template>
  <div class="min-h-screen bg-gray-950 flex">
    <!-- Sidebar -->
    <aside class="w-56 bg-gray-900 border-r border-gray-800 flex flex-col fixed inset-y-0 z-10">
      <!-- Logo -->
      <div class="flex items-center gap-3 px-5 py-5 border-b border-gray-800">
        <div class="w-8 h-8 rounded-lg bg-indigo-600 flex items-center justify-center text-white font-bold text-sm">SB</div>
        <span class="font-semibold text-white">Singbox Panel</span>
      </div>

      <!-- Nav -->
      <nav class="flex-1 p-3 space-y-0.5 overflow-y-auto">
        <NavItem to="/dashboard" icon="dashboard">仪表盘</NavItem>

        <div class="pt-3 pb-1 px-2">
          <span class="text-xs text-gray-500 font-medium uppercase tracking-wider">代理控制</span>
        </div>
        <NavItem to="/proxy" icon="proxy">代理模式</NavItem>
        <NavItem to="/core" icon="core">核心管理</NavItem>
        <NavItem to="/logs" icon="logs">实时日志</NavItem>

        <div class="pt-3 pb-1 px-2">
          <span class="text-xs text-gray-500 font-medium uppercase tracking-wider">配置</span>
        </div>
        <NavItem to="/providers" icon="providers">Providers</NavItem>
        <NavItem to="/dns" icon="dns">DNS</NavItem>
        <NavItem to="/inbounds" icon="inbound">入站 Inbounds</NavItem>
        <NavItem to="/outbounds" icon="outbound">出站 Outbounds</NavItem>
        <NavItem to="/route" icon="route">路由 Route</NavItem>
        <NavItem to="/rulesets" icon="ruleset">规则集</NavItem>
        <NavItem to="/config" icon="config">原始配置</NavItem>
      </nav>

      <!-- Bottom status -->
      <div class="p-4 border-t border-gray-800 space-y-2">
        <div class="flex items-center justify-between text-xs">
          <span class="text-gray-500">sing-box</span>
          <span :class="coreRunning ? 'badge-green' : 'badge-red'">{{ coreRunning ? '运行中' : '已停止' }}</span>
        </div>
        <div class="flex items-center justify-between text-xs">
          <span class="text-gray-500">代理规则</span>
          <span :class="proxyEnabled ? 'badge-green' : 'badge-red'">{{ proxyEnabled ? '已应用' : '未应用' }}</span>
        </div>
        <button class="btn-ghost w-full text-xs justify-center mt-2" @click="logout">退出登录</button>
      </div>
    </aside>

    <!-- Main content -->
    <main class="ml-56 flex-1 min-h-screen">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { systemApi } from '@/api'
import NavItem from '@/components/NavItem.vue'

const router = useRouter()
const auth = useAuthStore()
const coreRunning = ref(false)
const proxyEnabled = ref(false)
let timer = null

async function fetchStatus() {
  try {
    const s = await systemApi.status()
    coreRunning.value = s.core?.running || false
    proxyEnabled.value = s.proxy?.enabled || false
  } catch {}
}

onMounted(() => {
  fetchStatus()
  timer = setInterval(fetchStatus, 5000)
})
onUnmounted(() => clearInterval(timer))

function logout() {
  auth.logout()
  router.push('/login')
}
</script>
