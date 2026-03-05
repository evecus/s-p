<template>
  <div class="p-6 space-y-6 max-w-3xl">
    <h1 class="text-xl font-bold text-white">DNS 配置</h1>

    <!-- Global strategy -->
    <div class="card space-y-4">
      <p class="section-title">全局策略</p>
      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="label">全局解析策略</label>
          <select v-model="dns.strategy" class="input">
            <option value="">默认</option>
            <option value="prefer_ipv4">prefer_ipv4（优先IPv4）</option>
            <option value="prefer_ipv6">prefer_ipv6（优先IPv6）</option>
            <option value="ipv4_only">ipv4_only</option>
            <option value="ipv6_only">ipv6_only</option>
          </select>
        </div>
        <div>
          <label class="label">兜底 DNS</label>
          <input v-model="dns.final" type="text" class="input" placeholder="google-dns" />
        </div>
        <div class="flex items-center gap-2">
          <input type="checkbox" id="ic" v-model="dns.independent_cache" class="w-4 h-4 rounded border-gray-600 bg-gray-700 text-indigo-500" />
          <label for="ic" class="text-sm text-gray-300 cursor-pointer">independent_cache</label>
        </div>
        <div class="flex items-center gap-2">
          <input type="checkbox" id="rm" v-model="dns.reverse_mapping" class="w-4 h-4 rounded border-gray-600 bg-gray-700 text-indigo-500" />
          <label for="rm" class="text-sm text-gray-300 cursor-pointer">reverse_mapping（IP反查域名）</label>
        </div>
      </div>
    </div>

    <!-- DNS Servers -->
    <div class="card space-y-4">
      <div class="flex items-center justify-between">
        <p class="section-title mb-0">DNS 服务器</p>
        <button class="btn-ghost text-xs" @click="addServer">+ 添加</button>
      </div>
      <div v-for="(s, i) in dns.servers" :key="i" class="bg-gray-800 rounded-xl p-4 space-y-3">
        <div class="flex items-center justify-between">
          <span class="text-sm font-medium text-gray-200">{{ s.tag || '未命名' }}</span>
          <button class="text-red-400 hover:text-red-300 text-xs" @click="dns.servers.splice(i,1)">删除</button>
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="label">Tag</label>
            <input v-model="s.tag" type="text" class="input" placeholder="ali-dns" />
          </div>
          <div>
            <label class="label">类型</label>
            <select v-model="s.type" class="input">
              <option value="udp">UDP</option>
              <option value="tcp">TCP</option>
              <option value="tls">TLS (DoT)</option>
              <option value="https">HTTPS (DoH)</option>
              <option value="quic">QUIC (DoQ)</option>
              <option value="local">Local</option>
              <option value="fakeip">FakeIP</option>
            </select>
          </div>
          <div>
            <label class="label">服务器地址</label>
            <input v-model="s.server" type="text" class="input" placeholder="8.8.8.8" />
          </div>
          <div>
            <label class="label">端口（空=默认）</label>
            <input v-model.number="s.server_port" type="number" class="input" placeholder="53" />
          </div>
          <div>
            <label class="label">出站（detour）</label>
            <input v-model="s.detour" type="text" class="input" placeholder="直连 / proxy" />
          </div>
          <div v-if="s.type === 'https'">
            <label class="label">DoH Path</label>
            <input v-model="s.path" type="text" class="input" placeholder="/dns-query" />
          </div>
        </div>
      </div>
    </div>

    <!-- DNS Rules -->
    <div class="card space-y-4">
      <div class="flex items-center justify-between">
        <p class="section-title mb-0">DNS 规则</p>
        <button class="btn-ghost text-xs" @click="addRule">+ 添加规则</button>
      </div>
      <div v-for="(r, i) in dns.rules" :key="i" class="bg-gray-800 rounded-xl p-4 space-y-3">
        <div class="flex items-center justify-between">
          <span class="text-xs text-gray-400">规则 {{ i+1 }}</span>
          <button class="text-red-400 hover:text-red-300 text-xs" @click="dns.rules.splice(i,1)">删除</button>
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="label">rule_set（逗号分隔）</label>
            <input :value="(r.rule_set||[]).join(',')" @input="e => r.rule_set = e.target.value.split(',').map(s=>s.trim()).filter(Boolean)" type="text" class="input" placeholder="cn_domain,medirect" />
          </div>
          <div>
            <label class="label">DNS 服务器</label>
            <input v-model="r.server" type="text" class="input" placeholder="ali-dns" />
          </div>
          <div>
            <label class="label">解析策略</label>
            <select v-model="r.strategy" class="input">
              <option value="">继承全局</option>
              <option value="prefer_ipv4">prefer_ipv4</option>
              <option value="prefer_ipv6">prefer_ipv6</option>
              <option value="ipv4_only">ipv4_only</option>
              <option value="ipv6_only">ipv6_only</option>
            </select>
          </div>
        </div>
      </div>
    </div>

    <div class="flex gap-3">
      <button class="btn-primary" :disabled="saving" @click="save">{{ saving ? '保存中...' : '保存 DNS 配置' }}</button>
    </div>
    <p v-if="msg" class="text-sm" :class="isErr ? 'text-red-400' : 'text-emerald-400'">{{ msg }}</p>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { configApi } from '@/api'

const dns = ref({ servers: [], rules: [], strategy: 'prefer_ipv4', final: '', independent_cache: true, reverse_mapping: false })
const saving = ref(false)
const msg = ref('')
const isErr = ref(false)

onMounted(async () => {
  try {
    const s = await configApi.getSections()
    if (s.dns) dns.value = { ...dns.value, ...s.dns }
    if (!dns.value.servers) dns.value.servers = []
    if (!dns.value.rules) dns.value.rules = []
  } catch {}
})

function addServer() {
  dns.value.servers.push({ tag: '', type: 'udp', server: '', detour: '' })
}
function addRule() {
  dns.value.rules.push({ rule_set: [], server: '', strategy: '' })
}

async function save() {
  saving.value = true; msg.value = ''
  try {
    await configApi.setSection('dns', dns.value)
    msg.value = 'DNS 配置已保存 ✓'; isErr.value = false
  } catch (e) { msg.value = String(e); isErr.value = true }
  finally { saving.value = false; setTimeout(() => msg.value = '', 4000) }
}
</script>
