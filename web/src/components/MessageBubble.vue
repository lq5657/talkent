<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  role: 'user' | 'ai'
  content: string
  roleName?: string
  startTime?: Date
  endTime: Date
}>()

function pad(n: number): string {
  return n.toString().padStart(2, '0')
}

function formatTime(d: Date): string {
  return `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

function formatDuration(ms: number): string {
  const totalSec = Math.round(ms / 1000)
  if (totalSec >= 60) {
    const min = Math.floor(totalSec / 60)
    const sec = totalSec % 60
    return sec > 0 ? `${min}分${sec}秒` : `${min}分`
  }
  return `${totalSec}秒`
}

const startDisplay = computed(() => {
  return props.startTime ? formatTime(props.startTime) : '—'
})

const endDisplay = computed(() => {
  return formatTime(props.endTime)
})

const durationDisplay = computed(() => {
  if (!props.startTime) return '—'
  return formatDuration(props.endTime.getTime() - props.startTime.getTime())
})
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
    <div class="text-xs text-gray-400 mt-1 px-1 flex gap-2">
      <span>开始 {{ startDisplay }}</span>
      <span>结束 {{ endDisplay }}</span>
      <span>耗时 {{ durationDisplay }}</span>
    </div>
  </div>
</template>
