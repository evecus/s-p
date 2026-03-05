<template>
  <div class="p-6 space-y-6 max-w-3xl">
    <div class="flex items-center justify-between">
      <h1 class="text-xl font-bold text-white">出站 Outbounds</h1>
      <div class="flex gap-2">
        <select v-model="newType" class="input w-36 text-xs">
          <option value="selector">selector</option>
          <option value="urltest">urltest</option>
          <option value="direct">direct</option>
          <option value="block">block</option>
          <option value="dns">dns</option>
        </select>
        <button class="btn-primary text-xs" @click="addOutbound">+ 添加</button>
      </div>
    </div>

    <div v-for="(ob, i) in outbounds" :key="i" class="card space-y-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <span :class="typeColor(ob.type)" class="badge">{{ ob.type }}</span>
          <span class="text-sm font-medium text-white">{{ ob.tag }}</span>
        </div>
        <button class="text-red-400 hover:text-red-300 text-xs" @click="outbounds.splice(i,1)">删除</button>
      </div>

      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="label">Tag</label>
          <input v-model="ob.tag" type="text" class="input" />
        </div>
        <div>
          <label class="label">类型</label>
          <select v-model="ob.type" class="input">
            <option value="selector">selector（手动选择）</option>
            <option value="urltest">urltest（自动测速）</option>
            <option value="direct">direct（直连）</option>
            <option value="block">block（拦截）</option>
            <option value="dns">dns（DNS出站）</option>
          </select>
        </div>
      </div>

      <!-- selector / urltest -->
      <template v-if="ob.type === 'selector' || ob.type === 'urltest'">
        <div>
          <label class="label">outbounds（逗号分隔，providers 节点会自动注入）</label>
          <input :value="(ob.outbounds||[]).join(',')"
            @input="e => ob.outbounds = e.target.value.split(',').map(s=>s.trim()).filter(Boolean)"
            type="text" class="input" placeholder="direct,proxy" />
        </div>
        <div>
          <label class="label">providers（绑定的订阅 tag）</label>
          <input :value="(ob.providers||[]).join(',')"
            @input="e => ob.providers = e.target.value.split(',').map(s=>s.trim()).filter(Boolean)"
            type="text" class="input" placeholder="订阅" />
        </div>
        <template v-if="ob.type === 'urltest'">
          <div class="grid grid-cols-3 gap-3">
            <div>
              <label class="label">测速 URL</label>
              <input v-model="ob.url" type="text" class="input" placeholder="https://www.gstatic.com/generate_204" />
            </div>
            <div>
              <label class="label">测速间隔</label>
              <input v-model="ob.interval" type="text" class="input" placeholder="3m" />
            </div>
            <div>
              <label class="label">容差（ms）</label>
              <input v-model.number="ob.tolerance" type="number" class="input" placeholder="50" />
            </div>
          </div>
        </template>
      </template>
    </div>

    <div class="flex gap-3">
      <button class="btn-primary" :disabled="saving" @click="save">{{ saving ? '保存中...' : '保存出站配置' }}</button>
    </div>
    <p v-if="msg" class="text-sm" :class="isErr ? 'text-red-400' : 'text-emerald-400'">{{ msg }}</p>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { configApi } from '@/api'

const outbounds = ref([])
const newType = ref('selector')
const saving = ref(false)
const msg = ref('')
const isErr = ref(false)

const typeColors = { selector: 'badge-blue', urltest: 'badge-yellow', direct: 'badge-green', block: 'badge-red', dns: 'badge-blue' }
const typeColor = t => typeColors[t] || 'badge-blue'

const defaults = {
  selector: { tag: '默认代理', type: 'selector', outbounds: [], providers: [] },
  urltest:  { tag: '自动选择', type: 'urltest',  outbounds: [], providers: [], url: 'https://www.gstatic.com/generate_204', interval: '3m', tolerance: 50 },
  direct:   { tag: '直连',     type: 'direct' },
  block:    { tag: '拦截',     type: 'block' },
  dns:      { tag: 'dns-out',  type: 'dns' },
}

onMounted(async () => {
  try { const s = await configApi.getSections(); outbounds.value = s.outbounds || [] } catch {}
})

function addOutbound() { outbounds.value.push({ ...defaults[newType.value] }) }

async function save() {
  saving.value = true; msg.value = ''
  try {
    await configApi.setSection('outbounds', outbounds.value)
    msg.value = '出站配置已保存 ✓'; isErr.value = false
  } catch (e) { msg.value = String(e); isErr.value = true }
  finally { saving.value = false; setTimeout(() => msg.value = '', 4000) }
}
</script>
