<script setup lang="ts">
import { computed, ref, watch } from 'vue'

const props = defineProps<{
  role: 'user' | 'ai'
  content: string
  roleName?: string
  startTime?: Date
  endTime: Date
}>()

const ttsSupported = ref('speechSynthesis' in window)
const isPlaying = ref(false)

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

watch(() => props.role, () => {
  isPlaying.value = false
})

function togglePlay() {
  if (!ttsSupported.value || !props.content) return

  if (isPlaying.value) {
    speechSynthesis.cancel()
    isPlaying.value = false
    return
  }

  speechSynthesis.cancel()
  const utter = new SpeechSynthesisUtterance(props.content)
  utter.lang = 'zh-CN'
  utter.onend = () => { isPlaying.value = false }
  utter.onerror = () => { isPlaying.value = false }
  isPlaying.value = true
  speechSynthesis.speak(utter)
}
</script>

<template>
  <div class="flex flex-col" :class="role === 'user' ? 'items-end' : 'items-start'">
    <span class="text-xs text-gray-400 mb-1 px-1">
      {{ role === 'user' ? '我' : roleName || 'AI' }}
    </span>
    <div class="flex items-end gap-1 max-w-[80%]">
      <div
        class="flex-1 px-4 py-3 rounded-2xl text-sm leading-relaxed whitespace-pre-wrap"
        :class="
          role === 'user'
            ? 'bg-blue-600 text-white rounded-br-md'
            : 'bg-white border border-gray-200 text-gray-800 rounded-bl-md'
        "
      >
        {{ content }}
      </div>
      <button
        v-if="role === 'ai' && ttsSupported && content"
        class="min-w-[44px] min-h-[44px] flex items-center justify-center rounded-full shrink-0 transition-colors"
        :class="isPlaying ? 'bg-blue-100 text-blue-600' : 'bg-gray-100 text-gray-500 hover:bg-gray-200'"
        @click="togglePlay"
        :title="isPlaying ? '停止播放' : '播放语音'"
      >
        <svg v-if="isPlaying" xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 24 24" fill="currentColor"><rect x="6" y="4" width="4" height="16"/><rect x="14" y="4" width="4" height="16"/></svg>
        <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 24 24" fill="currentColor"><polygon points="5,3 19,12 5,21"/></svg>
      </button>
    </div>
    <div class="text-xs text-gray-400 mt-1 px-1 flex gap-2">
      <span>开始 {{ startDisplay }}</span>
      <span>结束 {{ endDisplay }}</span>
      <span>耗时 {{ durationDisplay }}</span>
    </div>
  </div>
</template>
