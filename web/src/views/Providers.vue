<template>
  <div class="p-6 space-y-6 max-w-3xl">
    <div class="flex items-center justify-between">
      <h1 class="text-xl font-bold text-white">Providers 订阅管理</h1>
      <button class="btn-primary" @click="addProvider">+ 添加订阅</button>
    </div>

    <div v-if="providers.length === 0" class="card text-center py-12 text-gray-500">
      暂无订阅，点击右上角添加
    </div>

    <div v-for="(p, i) in providers" :key="i" class="card space-y-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <span class="w-7 h-7 rounded-lg bg-indigo-600/20 border border-indigo-500/30 flex items-center justify-center text-indigo-400 text-xs font-bold">{{ i+1 }}</span>
          <span class="font-medium text-white">{{ p.tag || '未命名' }}</span>
          <span :class="p.type === 'remote' ? 'badge-blue' : 'badge-yellow'">{{ p.type }}</span>
        </div>
        <div class="flex gap-2">
          <button class="btn-ghost text-xs" @click="updateProvider(p.tag)">⟳ 更新</button>
          <button class="btn-danger text-xs" @click="removeProvider(i)">删除</button>
        </div>
      </div>

      <div class="grid grid-cols-1 gap-3">
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="label">Tag（标签）</label>
            <input v-model="p.tag" type="text" class="input" placeholder="订阅" />
          </div>
          <div>
            <label class="label">类型</label>
            <select v-model="p.type" class="input">
              <option value="remote">remote（远程）</option>
              <option value="local">local（本地）</option>
            </select>
          </div>
        </div>

        <div v-if="p.type === 'remote'">
          <label class="label">订阅 URL</label>
          <input v-model="p.url" type="text" class="input" placeholder="https://..." />
        </div>

        <div>
          <label class="label">本地缓存路径</label>
          <input v-model="p.path" type="text" class="input" placeholder="/etc/singbox-panel/providers/sub.json" />
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="label">User-Agent</label>
            <input v-model="p.user_agent" type="text" class="input" placeholder="clash.meta" />
          </div>
          <div>
            <label class="label">更新间隔</label>
            <input v-model="p.update_interval" type="text" class="input" placeholder="12h0m0s" />
          </div>
        </div>

        <!-- Health check -->
        <div class="bg-gray-800 rounded-lg p-3 space-y-3">
          <div class="flex items-center justify-between">
            <span class="text-xs text-gray-400 font-medium">健康检查</span>
            <button
              :class="p.health_check?.enabled ? 'bg-indigo-600' : 'bg-gray-600'"
              class="w-9 h-5 rounded-full transition-colors relative"
              @click="toggleHealthCheck(p)"
            >
              <span :class="p.health_check?.enabled ? 'translate-x-4' : 'translate-x-1'"
                class="inline-block w-3 h-3 rounded-full bg-white transition-transform absolute top-1"></span>
            </button>
          </div>
          <template v-if="p.health_check?.enabled">
            <div class="grid grid-cols-3 gap-3">
              <div>
                <label class="label">检测 URL</label>
                <input v-model="p.health_check.url" type="text" class="input text-xs" placeholder="https://..." />
              </div>
              <div>
                <label class="label">间隔</label>
                <input v-model="p.health_check.interval" type="text" class="input" placeholder="10m0s" />
              </div>
              <div>
                <label class="label">超时</label>
                <input v-model="p.health_check.timeout" type="text" class="input" placeholder="3s" />
              </div>
            </div>
          </template>
        </div>

        <!-- Override TLS -->
        <div class="flex items-center gap-3">
          <input type="checkbox" :id="`tls_insecure_${i}`"
            :checked="p.override_tls?.insecure"
            @change="e => setTLSInsecure(p, e.target.checked)"
            class="w-4 h-4 rounded border-gray-600 bg-gray-700 text-indigo-500" />
          <label :for="`tls_insecure_${i}`" class="text-sm text-gray-300 cursor-pointer">跳过 TLS 证书验证（不安全）</label>
        </div>
      </div>
    </div>

    <div class="flex gap-3">
      <button class="btn-primary" :disabled="saving" @click="save">{{ saving ? '保存中...' : '保存配置' }}</button>
    </div>
    <p v-if="msg" class="text-sm" :class="isErr ? 'text-red-400' : 'text-emerald-400'">{{ msg }}</p>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { providersApi } from '@/api'

const providers = ref([])
const saving = ref(false)
const msg = ref('')
const isErr = ref(false)

onMounted(async () => {
  try {
    const res = await providersApi.get()
    providers.value = res.providers || []
  } catch {}
})

function addProvider() {
  providers.value.push({
    type: 'remote', tag: '订阅' + (providers.value.length + 1),
    url: '', path: `/etc/singbox-panel/providers/sub${providers.value.length + 1}.json`,
    user_agent: 'clash.meta', update_interval: '12h0m0s',
    health_check: { enabled: true, url: 'https://www.gstatic.com/generate_204', interval: '10m0s', timeout: '3s' },
    override_tls: { enabled: true, insecure: false }
  })
}

function removeProvider(i) { providers.value.splice(i, 1) }

function toggleHealthCheck(p) {
  if (!p.health_check) p.health_check = {}
  p.health_check.enabled = !p.health_check.enabled
}

function setTLSInsecure(p, val) {
  if (!p.override_tls) p.override_tls = { enabled: true }
  p.override_tls.insecure = val
  p.override_tls.enabled = val
}

async function updateProvider(tag) {
  try { await providersApi.update(tag); msg.value = `${tag} 已更新`; isErr.value = false }
  catch (e) { msg.value = String(e); isErr.value = true }
}

async function save() {
  saving.value = true; msg.value = ''
  try {
    await providersApi.set(providers.value)
    msg.value = '保存成功 ✓'; isErr.value = false
  } catch (e) { msg.value = String(e); isErr.value = true }
  finally { saving.value = false; setTimeout(() => msg.value = '', 4000) }
}
</script>
