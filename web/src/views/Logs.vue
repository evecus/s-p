<template>
  <div class="p-6 space-y-4 h-screen flex flex-col">
    <div class="flex items-center justify-between flex-shrink-0">
      <h1 class="text-xl font-bold text-white">实时日志</h1>
      <div class="flex items-center gap-3">
        <label class="flex items-center gap-2 text-sm text-gray-400 cursor-pointer">
          <input type="checkbox" v-model="autoScroll" class="rounded border-gray-600 bg-gray-700 text-indigo-500" />
          自动滚动
        </label>
        <select v-model.number="lineCount" class="input w-28 text-xs" @change="fetchLogs">
          <option :value="100">最近100行</option>
          <option :value="300">最近300行</option>
          <option :value="500">最近500行</option>
        </select>
        <input v-model="filter" type="text" class="input w-40 text-xs" placeholder="过滤关键词" />
        <button class="btn-ghost text-xs" @click="clearDisplay">清屏</button>
      </div>
    </div>

    <div ref="logBox"
      class="flex-1 bg-gray-950 border border-gray-800 rounded-xl p-4 font-mono text-xs overflow-y-auto leading-relaxed"
      style="min-height: 0"
    >
      <div v-for="(line, i) in filteredLogs" :key="i"
        :class="lineClass(line)"
        class="whitespace-pre-wrap break-all"
      >{{ line }}</div>
      <div v-if="filteredLogs.length === 0" class="text-gray-600 text-center mt-8">暂无日志</div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { coreApi } from '@/api'

const logs = ref([])
const filter = ref('')
const autoScroll = ref(true)
const lineCount = ref(200)
const logBox = ref(null)
const cleared = ref(false)
let timer = null

const filteredLogs = computed(() => {
  if (cleared.value) return []
  if (!filter.value) return logs.value
  return logs.value.filter(l => l.toLowerCase().includes(filter.value.toLowerCase()))
})

function lineClass(line) {
  const l = line.toLowerCase()
  if (l.includes('error') || l.includes('fatal')) return 'text-red-400'
  if (l.includes('warn')) return 'text-yellow-400'
  if (l.includes('info')) return 'text-blue-300'
  if (l.includes('[panel]')) return 'text-purple-400'
  return 'text-gray-300'
}

async function fetchLogs() {
  try {
    const res = await coreApi.logs(lineCount.value)
    logs.value = res.logs || []
    cleared.value = false
  } catch {}
}

function clearDisplay() { cleared.value = true }

watch(filteredLogs, async () => {
  if (autoScroll.value) {
    await nextTick()
    if (logBox.value) logBox.value.scrollTop = logBox.value.scrollHeight
  }
})

onMounted(() => { fetchLogs(); timer = setInterval(fetchLogs, 2000) })
onUnmounted(() => clearInterval(timer))
</script>
