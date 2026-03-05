<template>
  <div class="p-6 space-y-6 max-w-3xl">
    <h1 class="text-xl font-bold text-white">代理模式配置</h1>

    <!-- Master switch -->
    <div class="card flex items-center justify-between">
      <div>
        <p class="font-semibold text-white">代理规则总开关</p>
        <p class="text-sm text-gray-400 mt-0.5">开启后自动配置防火墙规则和策略路由</p>
      </div>
      <button
        :class="form.enabled ? 'bg-indigo-600 hover:bg-indigo-500' : 'bg-gray-700 hover:bg-gray-600'"
        class="relative w-14 h-7 rounded-full transition-colors duration-200 focus:outline-none"
        @click="form.enabled = !form.enabled"
      >
        <span :class="form.enabled ? 'translate-x-7' : 'translate-x-1'"
          class="inline-block w-5 h-5 rounded-full bg-white shadow transition-transform duration-200"></span>
      </button>
    </div>

    <template v-if="form.enabled">
      <!-- Scope -->
      <div class="card space-y-4">
        <p class="section-title">代理范围</p>
        <div class="grid grid-cols-2 gap-3">
          <ModeCard v-model="form.scope" value="self" title="本机代理" desc="仅代理本设备流量" icon="💻" />
          <ModeCard v-model="form.scope" value="gateway" title="透明网关" desc="代理整个局域网流量" icon="🌐" />
        </div>
      </div>

      <!-- Mode -->
      <div class="card space-y-4">
        <p class="section-title">代理模式</p>
        <div class="grid grid-cols-3 gap-3">
          <ModeCard v-model="form.mode" value="tproxy" title="TProxy" desc="支持TCP+UDP，需Linux内核4.18+" icon="⚡" recommended />
          <ModeCard v-model="form.mode" value="redir" title="Redirect" desc="仅TCP，兼容性最好" icon="🔀" />
          <ModeCard v-model="form.mode" value="tun" title="TUN" desc="虚拟网卡，sing-box内置处理" icon="🔧" />
        </div>
        <div class="bg-gray-800 rounded-lg p-3 text-xs text-gray-400 space-y-1">
          <p v-if="form.mode === 'tproxy'">⚡ <b class="text-gray-200">TProxy</b>：透明代理，TCP+UDP全支持，性能最佳。需要 nftables + 策略路由，内核≥4.18。</p>
          <p v-if="form.mode === 'redir'">🔀 <b class="text-gray-200">Redirect</b>：NAT重定向，仅支持TCP（UDP需配合TProxy）。兼容性好，适合旧内核。</p>
          <p v-if="form.mode === 'tun'">🔧 <b class="text-gray-200">TUN</b>：sing-box创建虚拟网卡接管全部流量，防火墙规则最简单，需要 /dev/tun 支持。</p>
        </div>
      </div>

      <!-- IP Version -->
      <div class="card space-y-4">
        <p class="section-title">IP 版本</p>
        <div class="grid grid-cols-3 gap-3">
          <ModeCard v-model="form.ip_version" value="ipv4" title="IPv4 only" desc="仅代理IPv4流量" icon="4️⃣" />
          <ModeCard v-model="form.ip_version" value="ipv6" title="IPv6 only" desc="仅代理IPv6流量" icon="6️⃣" />
          <ModeCard v-model="form.ip_version" value="both" title="IPv4 + IPv6" desc="同时代理两种流量" icon="🔢" />
        </div>
      </div>

      <!-- Advanced -->
      <div class="card space-y-4">
        <div class="flex items-center justify-between cursor-pointer" @click="showAdvanced = !showAdvanced">
          <p class="section-title mb-0">高级参数</p>
          <span class="text-gray-400 text-sm">{{ showAdvanced ? '▲ 收起' : '▼ 展开' }}</span>
        </div>
        <template v-if="showAdvanced">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="label">TProxy 端口</label>
              <input v-model.number="form.tproxy_port" type="number" class="input" />
            </div>
            <div>
              <label class="label">Redirect 端口</label>
              <input v-model.number="form.redir_port" type="number" class="input" />
            </div>
            <div>
              <label class="label">DNS 劫持端口</label>
              <input v-model.number="form.dns_port" type="number" class="input" />
            </div>
            <div>
              <label class="label">FW Mark</label>
              <input v-model.number="form.fwmark" type="number" class="input" />
            </div>
            <div>
              <label class="label">路由表编号</label>
              <input v-model.number="form.route_table" type="number" class="input" />
            </div>
            <div>
              <label class="label">出口网卡（空=自动）</label>
              <input v-model="form.interface" type="text" class="input" placeholder="auto" />
            </div>
          </div>
        </template>
      </div>
    </template>

    <!-- Apply -->
    <div class="flex gap-3">
      <button class="btn-primary" :disabled="loading" @click="apply">
        <span v-if="loading">应用中...</span>
        <span v-else>{{ form.enabled ? '✓ 应用规则' : '✓ 清除规则' }}</span>
      </button>
      <button v-if="currentConfig?.enabled" class="btn-danger" :disabled="loading" @click="clearRules">清除所有规则</button>
    </div>
    <p v-if="msg" class="text-sm" :class="isError ? 'text-red-400' : 'text-emerald-400'">{{ msg }}</p>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { proxyApi } from '@/api'
import ModeCard from '@/components/ModeCard.vue'

const form = ref({
  enabled: false, scope: 'self', mode: 'tproxy', ip_version: 'ipv4',
  tproxy_port: 7899, redir_port: 7898, dns_port: 1053,
  fwmark: 1, route_table: 100, interface: ''
})
const currentConfig = ref(null)
const showAdvanced = ref(false)
const loading = ref(false)
const msg = ref('')
const isError = ref(false)

onMounted(async () => {
  try {
    const cfg = await proxyApi.getMode()
    currentConfig.value = cfg
    Object.assign(form.value, cfg)
  } catch {}
})

async function apply() {
  loading.value = true; msg.value = ''
  try {
    await proxyApi.apply(form.value)
    msg.value = '规则已' + (form.value.enabled ? '应用 ✓' : '清除 ✓')
    isError.value = false
    currentConfig.value = { ...form.value }
  } catch (e) {
    msg.value = String(e); isError.value = true
  } finally {
    loading.value = false
    setTimeout(() => msg.value = '', 4000)
  }
}

async function clearRules() {
  loading.value = true
  try {
    await proxyApi.stop()
    form.value.enabled = false
    msg.value = '规则已清除'
    isError.value = false
  } catch (e) {
    msg.value = String(e); isError.value = true
  } finally {
    loading.value = false
  }
}
</script>
