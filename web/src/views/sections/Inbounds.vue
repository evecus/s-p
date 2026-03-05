<template>
  <div class="p-6 space-y-6 max-w-3xl">
    <div class="flex items-center justify-between">
      <h1 class="text-xl font-bold text-white">入站 Inbounds</h1>
      <div class="flex gap-2">
        <select v-model="newType" class="input w-36 text-xs">
          <option value="tproxy">TProxy</option>
          <option value="tun">TUN</option>
          <option value="mixed">Mixed</option>
          <option value="socks">SOCKS</option>
          <option value="http">HTTP</option>
          <option value="direct">Direct</option>
        </select>
        <button class="btn-primary text-xs" @click="addInbound">+ 添加</button>
      </div>
    </div>

    <div v-for="(ib, i) in inbounds" :key="i" class="card space-y-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <span class="badge badge-blue">{{ ib.type }}</span>
          <span class="text-sm font-medium text-white">{{ ib.tag }}</span>
        </div>
        <button class="text-red-400 hover:text-red-300 text-xs" @click="inbounds.splice(i,1)">删除</button>
      </div>
      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="label">Tag</label>
          <input v-model="ib.tag" type="text" class="input" />
        </div>
        <div>
          <label class="label">类型</label>
          <select v-model="ib.type" class="input">
            <option value="tproxy">tproxy</option>
            <option value="tun">tun</option>
            <option value="mixed">mixed</option>
            <option value="socks">socks</option>
            <option value="http">http</option>
            <option value="direct">direct</option>
            <option value="redirect">redirect</option>
          </select>
        </div>
        <div>
          <label class="label">监听地址</label>
          <input v-model="ib.listen" type="text" class="input" placeholder="::" />
        </div>
        <div>
          <label class="label">端口</label>
          <input v-model.number="ib.listen_port" type="number" class="input" />
        </div>
      </div>

      <!-- TProxy specific -->
      <template v-if="ib.type === 'tproxy' || ib.type === 'mixed' || ib.type === 'socks'">
        <div class="grid grid-cols-2 gap-3">
          <div class="flex items-center gap-2">
            <input type="checkbox" :id="`sniff_${i}`" v-model="ib.sniff" class="w-4 h-4 rounded border-gray-600 bg-gray-700 text-indigo-500" />
            <label :for="`sniff_${i}`" class="text-sm text-gray-300 cursor-pointer">启用协议嗅探</label>
          </div>
          <div class="flex items-center gap-2">
            <input type="checkbox" :id="`sniffoverride_${i}`" v-model="ib.sniff_override_destination" class="w-4 h-4 rounded border-gray-600 bg-gray-700 text-indigo-500" />
            <label :for="`sniffoverride_${i}`" class="text-sm text-gray-300 cursor-pointer">覆盖目标地址</label>
          </div>
        </div>
      </template>

      <!-- TUN specific -->
      <template v-if="ib.type === 'tun'">
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="label">接口名称</label>
            <input v-model="ib.interface_name" type="text" class="input" placeholder="tun0" />
          </div>
          <div>
            <label class="label">MTU</label>
            <input v-model.number="ib.mtu" type="number" class="input" placeholder="9000" />
          </div>
          <div>
            <label class="label">IPv4 地址</label>
            <input :value="(ib.address||[]).filter(a=>a.includes('.')).join(',')"
              @input="e => setTunAddr(ib, e.target.value, 4)"
              type="text" class="input" placeholder="172.19.0.1/30" />
          </div>
          <div>
            <label class="label">IPv6 地址</label>
            <input :value="(ib.address||[]).filter(a=>a.includes(':')).join(',')"
              @input="e => setTunAddr(ib, e.target.value, 6)"
              type="text" class="input" placeholder="fdfe:dcba:9876::1/126" />
          </div>
          <div class="flex items-center gap-2">
            <input type="checkbox" :id="`autoroute_${i}`" v-model="ib.auto_route" class="w-4 h-4 rounded border-gray-600 bg-gray-700 text-indigo-500" />
            <label :for="`autoroute_${i}`" class="text-sm text-gray-300 cursor-pointer">auto_route</label>
          </div>
          <div class="flex items-center gap-2">
            <input type="checkbox" :id="`strictroute_${i}`" v-model="ib.strict_route" class="w-4 h-4 rounded border-gray-600 bg-gray-700 text-indigo-500" />
            <label :for="`strictroute_${i}`" class="text-sm text-gray-300 cursor-pointer">strict_route</label>
          </div>
        </div>
      </template>
    </div>

    <div class="flex gap-3">
      <button class="btn-primary" :disabled="saving" @click="save">{{ saving ? '保存中...' : '保存入站配置' }}</button>
    </div>
    <p v-if="msg" class="text-sm" :class="isErr ? 'text-red-400' : 'text-emerald-400'">{{ msg }}</p>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { configApi } from '@/api'

const inbounds = ref([])
const newType = ref('tproxy')
const saving = ref(false)
const msg = ref('')
const isErr = ref(false)

const defaults = {
  tproxy: { tag: 'tproxy-in', type: 'tproxy', listen: '::', listen_port: 7899, sniff: true, sniff_override_destination: false },
  tun:    { tag: 'tun-in',    type: 'tun',    interface_name: 'tun0', mtu: 9000, address: ['172.19.0.1/30','fdfe:dcba:9876::1/126'], auto_route: true, strict_route: true },
  mixed:  { tag: 'mixed-in',  type: 'mixed',  listen: '127.0.0.1', listen_port: 2080, sniff: true },
  socks:  { tag: 'socks-in',  type: 'socks',  listen: '127.0.0.1', listen_port: 1080 },
  http:   { tag: 'http-in',   type: 'http',   listen: '127.0.0.1', listen_port: 8080 },
  direct: { tag: 'dns-in',    type: 'direct', listen: '::', listen_port: 1153 },
}

onMounted(async () => {
  try {
    const s = await configApi.getSections()
    inbounds.value = s.inbounds || []
  } catch {}
})

function addInbound() { inbounds.value.push({ ...defaults[newType.value] }) }

function setTunAddr(ib, val, ver) {
  if (!ib.address) ib.address = []
  const others = ib.address.filter(a => ver === 4 ? a.includes(':') : a.includes('.'))
  ib.address = [...others, ...val.split(',').map(s=>s.trim()).filter(Boolean)]
}

async function save() {
  saving.value = true; msg.value = ''
  try {
    await configApi.setSection('inbounds', inbounds.value)
    msg.value = '入站配置已保存 ✓'; isErr.value = false
  } catch (e) { msg.value = String(e); isErr.value = true }
  finally { saving.value = false; setTimeout(() => msg.value = '', 4000) }
}
</script>
