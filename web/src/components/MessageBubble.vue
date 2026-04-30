<script setup lang="ts">
defineProps<{
  role: 'user' | 'ai'
  content: string
  roleName?: string
}>()
</script>

<template>
  <div class="flex flex-col" :class="role === 'user' ? 'items-end' : 'items-start'">
    <span class="text-xs text-gray-400 mb-1 px-1">
      {{ role === 'user' ? '我' : roleName || 'AI' }}
    </span>
    <div
      class="max-w-[80%] px-4 py-3 rounded-2xl text-sm leading-relaxed whitespace-pre-wrap"
      :class="
        role === 'user'
          ? 'bg-blue-600 text-white rounded-br-md'
          : 'bg-white border border-gray-200 text-gray-800 rounded-bl-md'
      "
    >
      {{ content }}
    </div>
  </div>
</template>
