<template>
  <div
    :class="[
      'relative border-2 rounded-xl p-4 cursor-pointer transition-all duration-150',
      modelValue === value
        ? 'border-indigo-500 bg-indigo-600/10'
        : 'border-gray-700 bg-gray-800/50 hover:border-gray-600'
    ]"
    @click="$emit('update:modelValue', value)"
  >
    <div v-if="recommended" class="absolute top-2 right-2">
      <span class="badge badge-blue text-xs">推荐</span>
    </div>
    <div class="text-2xl mb-2">{{ icon }}</div>
    <p class="font-medium text-sm text-white">{{ title }}</p>
    <p class="text-xs text-gray-400 mt-0.5">{{ desc }}</p>
  </div>
</template>

<script setup>
defineProps({ modelValue: String, value: String, title: String, desc: String, icon: String, recommended: Boolean })
defineEmits(['update:modelValue'])
</script>
