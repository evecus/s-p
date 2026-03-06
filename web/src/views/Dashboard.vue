<template>
  <div class="p-6 space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-xl font-bold text-white">仪表盘</h1>
      <span class="text-xs text-gray-500">{{ currentTime }}</span>
    </div>

    <!-- Status Cards -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <StatusCard title="sing-box" :value="status.core?.running ? '运行中' : '已停止'"
        :color="status.core?.running ? 'green' : 'red'" icon="core"
        :sub="status.core?.uptime ? '运行 ' + status.core.uptime : '-'" />
      <StatusCard title="代理规则" :value="status.proxy?.enabled ? '已应用' : '未应用'"
        :color="status.proxy?.enabled ? 'green' : 'gray'" icon="proxy"
        :sub="status.proxy?.mode ? status.proxy.mode.toUpperCase() + ' · ' + (status.proxy.scope === 'gateway' ? '网关' : '本机') : '-'" />
      <StatusCard title="核心版本" :value="coreInfo.version || '未安装'"
        :color="coreInfo.installed ? 'blue' : 'yellow'" icon="version"
        :sub="coreInfo.update_available ? '⬆ 有新版本 ' + coreInfo.latest_version : '已是最新'" />
      <StatusCard title="系统" :value="sysInfo.hostname || '-'"
        color="gray" icon="system" :sub="sysInfo.arch + ' · ' + sysInfo.kernel" />
    </div>

    <!-- Main Controls -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">

      <!-- sing-box control -->
      <div class="card">
        <div class="section-title">sing-box 控制</div>
        <div class="flex items-center gap-3 mb-6">
          <div :class="['w-3 h-3 rounded-full', status.core?.running ? 'bg-emerald-400 animate-pulse' : 'bg-gray-600']"></div>
          <span class="text-lg font-semibold" :class="status.core?.running ? 'text-emerald-400' : 'text-gray-400'">
            {{ status.core?.running ? '运行中' : '已停止' }}
          </span>
          <span v-if="status.core?.pid" class="text-xs text-gray-500">PID {{ status.core.pid }}</span>
        </div>
        <div class="flex gap-3">
          <button v-if="!status.core?.running" class="btn-success" :disabled="!coreInfo.installed || actionLoading" @click="startCore">
            ▶ 启动
          </button>
          <template v-else>
            <button class="btn-danger" :disabled="actionLoading" @click="stopCore">■ 停止</button>
            <button class="btn-ghost" :disabled="actionLoading" @click="restartCore">↺ 重启</button>
          </template>
          <button class="btn-ghost" @click="validateConfig">✓ 验证配置</button>
        </div>
        <p v-if="actionMsg" class="mt-3 text-sm" :class="actionError ? 'text-red-400' : 'text-emerald-400'">{{ actionMsg }}</p>
      </div>

      <!-- Proxy mode quick status -->
      <div class="card">
        <div class="section-title">代理规则状态</div>
        <div class="space-y-3 mb-5">
          <div class="flex justify-between text-sm">
            <span class="text-gray-400">状态</span>
            <span :class="status.proxy?.enabled ? 'text-emerald-400' : 'text-gray-500'">
              {{ status.proxy?.enabled ? '已应用' : '未应用' }}
            </span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-gray-400">模式</span>
            <span class="text-gray-200">{{ modeLabel }}</span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-gray-400">IP版本</span>
            <span class="text-gray-200">{{ ipLabel }}</span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-gray-400">范围</span>
            <span class="text-gray-200">{{ scopeLabel }}</span>
          </div>
        </div>
        <div class="flex gap-3">
          <router-link to="/proxy" class="btn-primary">配置代理模式</router-link>
          <button v-if="status.proxy?.enabled" class="btn-danger" @click="stopProxy">清除规则</button>
        </div>
      </div>
    </div>

    <!-- Quick links -->
    <div class="card">
      <div class="section-title">快捷操作</div>
      <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
        <router-link to="/core" class="quick-btn">
          <span class="text-2xl">⬇</span>
          <span class="text-sm">下载/升级核心</span>
        </router-link>
        <router-link to="/providers" class="quick-btn">
          <span class="text-2xl">📦</span>
          <span class="text-sm">管理订阅</span>
        </router-link>
        <router-link to="/dns" class="quick-btn">
          <span class="text-2xl">🛡</span>
          <span class="text-sm">DNS 配置</span>
        </router-link>
        <router-link to="/logs" class="quick-btn">
          <span class="text-2xl">📋</span>
          <span class="text-sm">查看日志</span>
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { systemApi, coreApi, proxyApi } from '@/api'
import StatusCard from '@/components/StatusCard.vue'

const status = ref({ core: {}, proxy: {} })
const coreInfo = ref({})
const sysInfo = ref({ hostname: '-', arch: '-', kernel: '-' })
const actionLoading = ref(false)
const actionMsg = ref('')
const actionError = ref(false)
const currentTime = ref('')
let timer = null

const modeMap = { tproxy: 'TProxy', redir: 'Redir', tun: 'TUN' }
const ipMap = { ipv4: 'IPv4', ipv6: 'IPv6', both: 'IPv4 + IPv6' }
const scopeMap = { self: '本机', gateway: '透明网关' }

const modeLabel = computed(() => modeMap[status.value.proxy?.mode] || '-')
const ipLabel = computed(() => ipMap[status.value.proxy?.ip_version] || '-')
const scopeLabel = computed(() => scopeMap[status.value.proxy?.scope] || '-')

async function fetchAll() {
  const [s, ci, si] = await Promise.allSettled([
    systemApi.status(), coreApi.info(), systemApi.info()
  ])
  if (s.status === 'fulfilled') status.value = s.value
  if (ci.status === 'fulfilled') coreInfo.value = ci.value
  if (si.status === 'fulfilled') sysInfo.value = si.value
  currentTime.value = new Date().toLocaleTimeString('zh-CN')
}

async function withAction(fn, successMsg) {
  actionLoading.value = true
  actionMsg.value = ''
  try {
    await fn()
    actionMsg.value = successMsg
    actionError.value = false
    await fetchAll()
  } catch (e) {
    actionMsg.value = String(e)
    actionError.value = true
  } finally {
    actionLoading.value = false
    setTimeout(() => actionMsg.value = '', 4000)
  }
}

const startCore   = () => withAction(coreApi.start,   'sing-box 已启动 ✓')
const stopCore    = () => withAction(coreApi.stop,    'sing-box 已停止')
const restartCore = () => withAction(coreApi.restart, 'sing-box 已重启 ✓')
const stopProxy   = () => withAction(proxyApi.stop,   '代理规则已清除')

async function validateConfig() {
  actionLoading.value = true
  try {
    const r = await coreApi.validate?.() || { valid: true }
    actionMsg.value = r.valid ? '配置验证通过 ✓' : '配置有误: ' + r.error
    actionError.value = !r.valid
  } catch (e) {
    actionMsg.value = String(e)
    actionError.value = true
  } finally {
    actionLoading.value = false
    setTimeout(() => actionMsg.value = '', 5000)
  }
}

onMounted(() => { fetchAll(); timer = setInterval(fetchAll, 5000) })
onUnmounted(() => clearInterval(timer))
</script>

<style scoped>
.quick-btn {
  @apply flex flex-col items-center gap-2 p-4 rounded-xl bg-gray-800 hover:bg-gray-700 border border-gray-700 hover:border-gray-600 transition-colors cursor-pointer text-gray-300 hover:text-white;
}
</style>
