<template>
  <div class="p-6 space-y-6 max-w-3xl">
    <div class="flex items-center justify-between">
      <h1 class="text-xl font-bold text-white">规则集 Rule Sets</h1>
      <div class="flex gap-2">
        <button class="btn-ghost text-xs" @click="updateAll" :disabled="updating">{{ updating ? '更新中...' : '⟳ 更新全部' }}</button>
        <button class="btn-primary text-xs" @click="addRuleSet">+ 添加</button>
      </div>
    </div>

    <!-- Preset templates -->
    <div class="card space-y-3">
      <p class="section-title">快速添加常用规则集</p>
      <div class="flex flex-wrap gap-2">
        <button v-for="t in presets" :key="t.tag" class="btn-ghost text-xs" @click="addPreset(t)">+ {{ t.tag }}</button>
      </div>
    </div>

    <div v-if="ruleSets.length === 0" class="card text-center py-10 text-gray-500">暂无规则集</div>

    <div v-for="(rs, i) in ruleSets" :key="i" class="card space-y-3">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <span class="badge badge-blue">{{ rs.type }}</span>
          <span class="font-medium text-white text-sm">{{ rs.tag }}</span>
          <span class="badge badge-yellow">{{ rs.format }}</span>
        </div>
        <button class="text-red-400 hover:text-red-300 text-xs" @click="ruleSets.splice(i,1)">删除</button>
      </div>
      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="label">Tag</label>
          <input v-model="rs.tag" type="text" class="input" />
        </div>
        <div>
          <label class="label">类型</label>
          <select v-model="rs.type" class="input">
            <option value="remote">remote（远程）</option>
            <option value="local">local（本地）</option>
            <option value="inline">inline（内联）</option>
          </select>
        </div>
        <div v-if="rs.type === 'remote'">
          <label class="label">格式</label>
          <select v-model="rs.format" class="input">
            <option value="binary">binary（.srs）</option>
            <option value="source">source（.json）</option>
          </select>
        </div>
        <div v-if="rs.type === 'remote'">
          <label class="label">更新间隔</label>
          <input v-model="rs.update_interval" type="text" class="input" placeholder="1d" />
        </div>
        <div v-if="rs.type === 'remote'" class="col-span-2">
          <label class="label">URL</label>
          <input v-model="rs.url" type="text" class="input" placeholder="https://..." />
        </div>
        <div v-if="rs.type === 'remote'">
          <label class="label">下载出站 (download_detour)</label>
          <input v-model="rs.download_detour" type="text" class="input" placeholder="直连" />
        </div>
      </div>
    </div>

    <div class="flex gap-3">
      <button class="btn-primary" :disabled="saving" @click="save">{{ saving ? '保存中...' : '保存规则集配置' }}</button>
    </div>
    <p v-if="msg" class="text-sm" :class="isErr ? 'text-red-400' : 'text-emerald-400'">{{ msg }}</p>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { rulesetsApi, configApi } from '@/api'

const ruleSets = ref([])
const saving = ref(false)
const updating = ref(false)
const msg = ref('')
const isErr = ref(false)

const BASE = 'https://fastly.jsdelivr.net/gh/evecus/rules_set@master/sing-box'
const presets = [
  { tag: 'cn_domain',     url: `${BASE}/cn_domain-lite.srs` },
  { tag: 'cn_ip',         url: `${BASE}/cn_ip-lite.srs` },
  { tag: 'foreign_domain',url: `${BASE}/foreign_domain-lite.srs` },
  { tag: 'ads',           url: `${BASE}/ads_domain-lite.srs` },
  { tag: 'medirect',      url: `${BASE}/direct.srs` },
]

onMounted(async () => {
  try { const r = await rulesetsApi.get(); ruleSets.value = r.rule_sets || [] } catch {}
})

function addRuleSet() {
  ruleSets.value.push({ tag: '', type: 'remote', format: 'binary', url: '', download_detour: '直连', update_interval: '1d' })
}

function addPreset(t) {
  if (ruleSets.value.find(r => r.tag === t.tag)) return
  ruleSets.value.push({ tag: t.tag, type: 'remote', format: 'binary', url: t.url, download_detour: '直连', update_interval: '1d' })
}

async function updateAll() {
  updating.value = true
  try { await rulesetsApi.update(); msg.value = '规则集已更新（需重启生效）'; isErr.value = false }
  catch (e) { msg.value = String(e); isErr.value = true }
  finally { updating.value = false; setTimeout(() => msg.value = '', 4000) }
}

async function save() {
  saving.value = true; msg.value = ''
  try {
    // Merge rule_sets back into route section
    const s = await configApi.getSections()
    const route = s.route || {}
    route.rule_set = ruleSets.value
    await configApi.setSection('route', route)
    msg.value = '规则集已保存 ✓'; isErr.value = false
  } catch (e) { msg.value = String(e); isErr.value = true }
  finally { saving.value = false; setTimeout(() => msg.value = '', 4000) }
}
</script>
