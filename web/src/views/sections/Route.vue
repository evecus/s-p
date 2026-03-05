<template>
  <div class="p-6 space-y-6 max-w-3xl">
    <h1 class="text-xl font-bold text-white">路由 Route</h1>

    <!-- Global -->
    <div class="card space-y-4">
      <p class="section-title">全局路由设置</p>
      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="label">兜底出站 (final)</label>
          <input v-model="route.final" type="text" class="input" placeholder="proxy" />
        </div>
        <div>
          <label class="label">默认域名解析器</label>
          <input v-model="defaultResolver" type="text" class="input" placeholder="google-dns" />
        </div>
        <div class="flex items-center gap-2">
          <input type="checkbox" id="autoiface" v-model="route.auto_detect_interface" class="w-4 h-4 rounded border-gray-600 bg-gray-700 text-indigo-500" />
          <label for="autoiface" class="text-sm text-gray-300 cursor-pointer">auto_detect_interface</label>
        </div>
      </div>
    </div>

    <!-- Rules -->
    <div class="card space-y-4">
      <div class="flex items-center justify-between">
        <p class="section-title mb-0">路由规则</p>
        <button class="btn-ghost text-xs" @click="addRule">+ 添加规则</button>
      </div>
      <div class="text-xs text-gray-500 bg-gray-800/50 rounded-lg p-2">规则按顺序匹配，第一条匹配即生效。可拖拽调整顺序（开发中）。</div>

      <div v-for="(r, i) in rules" :key="i" class="bg-gray-800 rounded-xl p-4 space-y-3">
        <div class="flex items-center justify-between">
          <span class="text-xs text-gray-500">规则 {{ i+1 }}</span>
          <div class="flex gap-2">
            <button class="text-gray-500 hover:text-gray-300 text-xs" :disabled="i===0" @click="rules.splice(i-1,2,...rules.slice(i-1,i+1).reverse())">↑</button>
            <button class="text-gray-500 hover:text-gray-300 text-xs" :disabled="i===rules.length-1" @click="rules.splice(i,2,...rules.slice(i,i+2).reverse())">↓</button>
            <button class="text-red-400 hover:text-red-300 text-xs" @click="rules.splice(i,1)">删除</button>
          </div>
        </div>
        <div class="grid grid-cols-2 gap-3">
          <!-- Action type -->
          <div>
            <label class="label">动作类型</label>
            <select v-model="r._actionType" class="input" @change="onActionTypeChange(r)">
              <option value="outbound">出站 (outbound)</option>
              <option value="hijack-dns">劫持DNS</option>
              <option value="resolve">解析 (resolve)</option>
            </select>
          </div>

          <template v-if="r._actionType === 'outbound'">
            <div>
              <label class="label">出站</label>
              <input v-model="r.outbound" type="text" class="input" placeholder="proxy / 直连 / 拦截" />
            </div>
          </template>
          <template v-if="r._actionType === 'resolve'">
            <div>
              <label class="label">解析策略</label>
              <select v-model="r.strategy" class="input">
                <option value="">默认</option>
                <option value="ipv4_only">ipv4_only</option>
                <option value="ipv6_only">ipv6_only</option>
                <option value="prefer_ipv4">prefer_ipv4</option>
                <option value="prefer_ipv6">prefer_ipv6</option>
              </select>
            </div>
            <div class="flex items-center gap-2 col-span-2">
              <input type="checkbox" :id="`mo_${i}`" v-model="r.match_only" class="w-4 h-4 rounded border-gray-600 bg-gray-700 text-indigo-500" />
              <label :for="`mo_${i}`" class="text-sm text-gray-300 cursor-pointer">match_only（仅解析，继续匹配后续规则）</label>
            </div>
          </template>
        </div>

        <!-- Match conditions -->
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="label">inbound（入站 tag）</label>
            <input :value="(r.inbound||[]).join(',')" @input="e => r.inbound = e.target.value.split(',').map(s=>s.trim()).filter(Boolean)"
              type="text" class="input" placeholder="tproxy-in" />
          </div>
          <div>
            <label class="label">rule_set（逗号分隔）</label>
            <input :value="(r.rule_set||[]).join(',')" @input="e => r.rule_set = e.target.value.split(',').map(s=>s.trim()).filter(Boolean)"
              type="text" class="input" placeholder="cn_domain,cn_ip" />
          </div>
          <div>
            <label class="label">protocol</label>
            <input :value="(r.protocol||[]).join(',')" @input="e => r.protocol = e.target.value.split(',').map(s=>s.trim()).filter(Boolean)"
              type="text" class="input" placeholder="dns,quic" />
          </div>
          <div>
            <label class="label">network</label>
            <input v-model="r.network" type="text" class="input" placeholder="tcp / udp" />
          </div>
        </div>
      </div>
    </div>

    <div class="flex gap-3">
      <button class="btn-primary" :disabled="saving" @click="save">{{ saving ? '保存中...' : '保存路由配置' }}</button>
    </div>
    <p v-if="msg" class="text-sm" :class="isErr ? 'text-red-400' : 'text-emerald-400'">{{ msg }}</p>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { configApi } from '@/api'

const route = ref({ final: 'proxy', auto_detect_interface: true })
const rules = ref([])
const saving = ref(false)
const msg = ref('')
const isErr = ref(false)

const defaultResolver = computed({
  get: () => route.value.default_domain_resolver?.server || '',
  set: v => { route.value.default_domain_resolver = v ? { server: v } : undefined }
})

function detectActionType(r) {
  if (r.action === 'hijack-dns') return 'hijack-dns'
  if (r.action === 'resolve') return 'resolve'
  return 'outbound'
}

function onActionTypeChange(r) {
  if (r._actionType === 'hijack-dns') { r.action = 'hijack-dns'; delete r.outbound }
  else if (r._actionType === 'resolve') { r.action = 'resolve'; delete r.outbound }
  else { delete r.action }
}

onMounted(async () => {
  try {
    const s = await configApi.getSections()
    if (s.route) {
      const { rules: r, ...rest } = s.route
      route.value = { auto_detect_interface: true, ...rest }
      rules.value = (r || []).map(rule => ({ ...rule, _actionType: detectActionType(rule) }))
    }
  } catch {}
})

function addRule() {
  rules.value.push({ _actionType: 'outbound', outbound: 'proxy', rule_set: [] })
}

async function save() {
  saving.value = true; msg.value = ''
  try {
    const cleanRules = rules.value.map(({ _actionType, ...r }) => r)
    await configApi.setSection('route', { ...route.value, rules: cleanRules })
    msg.value = '路由配置已保存 ✓'; isErr.value = false
  } catch (e) { msg.value = String(e); isErr.value = true }
  finally { saving.value = false; setTimeout(() => msg.value = '', 4000) }
}
</script>
