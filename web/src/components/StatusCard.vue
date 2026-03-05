<template>
  <div class="card p-4">
    <div class="flex items-start justify-between">
      <div class="flex-1 min-w-0">
        <p class="text-xs text-gray-500 mb-1">{{ title }}</p>
        <p class="text-base font-semibold truncate" :class="colorClass">{{ value }}</p>
        <p class="text-xs text-gray-500 mt-1 truncate">{{ sub }}</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
const props = defineProps({ title: String, value: String, color: String, sub: String, icon: String })
const colorMap = { green: 'text-emerald-400', red: 'text-red-400', blue: 'text-blue-400', yellow: 'text-yellow-400', gray: 'text-gray-300' }
const colorClass = computed(() => colorMap[props.color] || 'text-gray-300')
</script>
