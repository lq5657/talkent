<script setup lang="ts">
const model = defineModel<string>({ default: '' })

defineProps<{
  disabled: boolean
}>()

const emit = defineEmits<{
  send: []
}>()

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    if (model.value.trim()) {
      emit('send')
    }
  }
}
</script>

<template>
  <div class="flex gap-2 items-end">
    <textarea
      v-model="model"
      rows="1"
      placeholder="输入消息...（Shift+Enter 换行）"
      class="flex-1 px-4 py-3 border border-gray-300 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
      :disabled="disabled"
      @keydown="handleKeydown"
    />
    <button
      class="px-6 py-3 bg-blue-600 text-white rounded-xl hover:bg-blue-700 disabled:opacity-50 text-sm font-medium shrink-0"
      :disabled="disabled || !model.trim()"
      @click="$emit('send')"
    >
      发送
    </button>
  </div>
</template>
