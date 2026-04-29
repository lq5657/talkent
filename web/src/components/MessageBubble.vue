<script setup lang="ts">
defineProps<{
  role: 'user' | 'ai'
  content: string
}>()
</script>

<template>
  <div class="flex" :class="role === 'user' ? 'justify-end' : 'justify-start'">
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
