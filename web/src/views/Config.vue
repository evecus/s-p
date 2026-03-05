<template>
  <div class="p-6 space-y-4 h-screen flex flex-col">
    <div class="flex items-center justify-between flex-shrink-0">
      <h1 class="text-xl font-bold text-white">原始配置编辑器</h1>
      <div class="flex gap-3">
        <button class="btn-ghost text-xs" @click="validate">✓ 验证</button>
        <button class="btn-ghost text-xs" @click="format">⌥ 格式化</button>
        <button class="btn-primary text-xs" :disabled="saving" @click="save">{{ saving ? '保存中...' : '保存' }}</button>
      </div>
    </div>

    <p v-if="msg" class="text-sm flex-shrink-0" :class="isErr ? 'text-red-400' : 'text-emerald-400'">{{ msg }}</p>

    <div class="flex-1 relative" style="min-height:0">
      <textarea
        v-model="raw"
        class="absolute inset-0 w-full h-full bg-gray-950 border border-gray-800 rounded-xl p-4 font-mono text-xs text-gray-200 resize-none focus:outline-none focus:ring-2 focus:ring-indigo-500 leading-relaxed"
        spellcheck="false"
        @keydown.tab.prevent="insertTab"
      ></textarea>
    </div>

    <div class="flex-shrink-0 flex items-center justify-between text-xs text-gray-500">
      <span>{{ lineCount }} 行 · {{ byteCount }} 字节</span>
      <span>JSON · sing-box config</span>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { configApi } from '@/api'

const raw = ref('')
const saving = ref(false)
const msg = ref('')
const isErr = ref(false)

const lineCount = computed(() => raw.value.split('\n').length)
const byteCount = computed(() => new Blob([raw.value]).size)

onMounted(async () => {
  try { const r = await configApi.getRaw(); raw.value = r.config } catch {}
})

function insertTab(e) {
  const ta = e.target
  const start = ta.selectionStart
  const end = ta.selectionEnd
  raw.value = raw.value.slice(0,start) + '  ' + raw.value.slice(end)
  ta.selectionStart = ta.selectionEnd = start + 2
}

function format() {
  try {
    raw.value = JSON.stringify(JSON.parse(raw.value), null, 2)
    msg.value = '格式化完成'; isErr.value = false
  } catch (e) { msg.value = 'JSON解析失败: ' + e.message; isErr.value = true }
  setTimeout(() => msg.value = '', 3000)
}

async function validate() {
  try {
    JSON.parse(raw.value)
    const r = await configApi.validate()
    msg.value = r.valid ? '✓ 配置验证通过' : '✗ ' + r.error
    isErr.value = !r.valid
  } catch (e) { msg.value = 'JSON格式错误: ' + e.message; isErr.value = true }
  setTimeout(() => msg.value = '', 5000)
}

async function save() {
  saving.value = true; msg.value = ''
  try {
    await configApi.setRaw(raw.value)
    msg.value = '配置已保存 ✓'; isErr.value = false
  } catch (e) { msg.value = String(e); isErr.value = true }
  finally { saving.value = false; setTimeout(() => msg.value = '', 4000) }
}
</script>
