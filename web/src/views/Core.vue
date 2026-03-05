<template>
  <div class="p-6 space-y-6 max-w-2xl">
    <h1 class="text-xl font-bold text-white">核心管理</h1>

    <!-- Current status -->
    <div class="card space-y-4">
      <p class="section-title">当前状态</p>
      <div class="grid grid-cols-2 gap-4 text-sm">
        <div>
          <p class="text-gray-500 mb-1">安装状态</p>
          <span :class="info.installed ? 'badge-green' : 'badge-red'">{{ info.installed ? '已安装' : '未安装' }}</span>
        </div>
        <div>
          <p class="text-gray-500 mb-1">运行状态</p>
          <span :class="info.running ? 'badge-green' : 'badge-red'">{{ info.running ? '运行中' : '已停止' }}</span>
        </div>
        <div>
          <p class="text-gray-500 mb-1">当前版本</p>
          <span class="text-gray-200 font-mono">{{ info.version || '-' }}</span>
        </div>
        <div>
          <p class="text-gray-500 mb-1">最新版本</p>
          <span class="text-gray-200 font-mono">{{ info.latest_version || '获取中...' }}</span>
        </div>
        <div class="col-span-2">
          <p class="text-gray-500 mb-1">二进制路径</p>
          <span class="text-gray-400 font-mono text-xs">{{ info.path }}</span>
        </div>
      </div>
      <div v-if="info.update_available" class="bg-yellow-900/30 border border-yellow-700/50 rounded-lg p-3 text-sm text-yellow-300">
        ⬆ 发现新版本 {{ info.latest_version }}，建议升级
      </div>
    </div>

    <!-- Download -->
    <div class="card space-y-4">
      <p class="section-title">下载 / 升级</p>
      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="label">版本（空=最新）</label>
          <input v-model="dlVersion" type="text" class="input" placeholder="e.g. 1.9.0" />
        </div>
        <div>
          <label class="label">架构（空=自动）</label>
          <select v-model="dlArch" class="input">
            <option value="">自动检测</option>
            <option value="amd64">amd64 (x86_64)</option>
            <option value="arm64">arm64 (aarch64)</option>
            <option value="armv7">armv7</option>
            <option value="386">386 (i686)</option>
          </select>
        </div>
      </div>
      <button class="btn-primary" :disabled="downloading" @click="startDownload">
        {{ downloading ? '下载中...' : (info.installed ? '⬆ 升级核心' : '⬇ 下载核心') }}
      </button>

      <!-- Progress -->
      <div v-if="downloading || progress.status === 'done' || progress.status === 'error'" class="space-y-2">
        <div class="flex justify-between text-xs text-gray-400">
          <span>{{ progress.message || '准备中...' }}</span>
          <span>{{ Math.round(progress.progress || 0) }}%</span>
        </div>
        <div class="h-2 bg-gray-700 rounded-full overflow-hidden">
          <div class="h-full bg-indigo-500 rounded-full transition-all duration-300"
            :style="{ width: (progress.progress || 0) + '%' }"
            :class="progress.status === 'error' ? 'bg-red-500' : progress.status === 'done' ? 'bg-emerald-500' : 'bg-indigo-500'"
          ></div>
        </div>
        <p v-if="progress.status === 'done'" class="text-emerald-400 text-sm">✓ 下载完成！</p>
        <p v-if="progress.status === 'error'" class="text-red-400 text-sm">✗ 下载失败</p>
      </div>
    </div>

    <!-- Service control -->
    <div class="card space-y-4">
      <p class="section-title">服务控制</p>
      <div class="flex gap-3">
        <button class="btn-success" :disabled="!info.installed || info.running || actionLoading" @click="start">▶ 启动</button>
        <button class="btn-danger"  :disabled="!info.running || actionLoading" @click="stop">■ 停止</button>
        <button class="btn-ghost"   :disabled="!info.installed || actionLoading" @click="restart">↺ 重启</button>
      </div>
      <p v-if="actionMsg" class="text-sm" :class="actionErr ? 'text-red-400' : 'text-emerald-400'">{{ actionMsg }}</p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { coreApi } from '@/api'

const info = ref({})
const dlVersion = ref('')
const dlArch = ref('')
const downloading = ref(false)
const progress = ref({})
const actionLoading = ref(false)
const actionMsg = ref('')
const actionErr = ref(false)
let pollTimer = null

async function fetchInfo() {
  try { info.value = await coreApi.info() } catch {}
}

async function startDownload() {
  downloading.value = true
  progress.value = { status: 'downloading', progress: 0 }
  try {
    await coreApi.download(dlVersion.value, dlArch.value)
    pollTimer = setInterval(async () => {
      const p = await coreApi.downloadProgress()
      progress.value = p
      if (p.status === 'done' || p.status === 'error') {
        clearInterval(pollTimer)
        downloading.value = false
        await fetchInfo()
      }
    }, 800)
  } catch (e) {
    progress.value = { status: 'error', message: String(e) }
    downloading.value = false
  }
}

async function withAction(fn, ok) {
  actionLoading.value = true; actionMsg.value = ''
  try { await fn(); actionMsg.value = ok; actionErr.value = false; await fetchInfo() }
  catch (e) { actionMsg.value = String(e); actionErr.value = true }
  finally { actionLoading.value = false; setTimeout(() => actionMsg.value = '', 4000) }
}

const start   = () => withAction(coreApi.start,   'sing-box 已启动 ✓')
const stop    = () => withAction(coreApi.stop,    'sing-box 已停止')
const restart = () => withAction(coreApi.restart, 'sing-box 已重启 ✓')

onMounted(fetchInfo)
onUnmounted(() => clearInterval(pollTimer))
</script>
