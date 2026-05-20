<script setup lang="ts">
import { ref, onMounted } from 'vue'

const model = defineModel<string>({ default: '' })

defineProps<{
  disabled: boolean
}>()

const emit = defineEmits<{
  send: []
}>()

const SpeechRecognitionAPI =
  (window as any).SpeechRecognition || (window as any).webkitSpeechRecognition
const recognitionSupported = ref(false)
const isRecording = ref(false)
const isRecognizing = ref(false)

let recognition: any = null

onMounted(() => {
  recognitionSupported.value = !!SpeechRecognitionAPI
})

function startRecording() {
  if (!SpeechRecognitionAPI) return
  recognition = new SpeechRecognitionAPI()
  recognition.lang = 'zh-CN'
  recognition.interimResults = false
  recognition.continuous = false

  recognition.onstart = () => {
    isRecording.value = true
    isRecognizing.value = false
  }

  recognition.onresult = (event: any) => {
    let transcript = ''
    for (let i = event.resultIndex; i < event.results.length; i++) {
      transcript += event.results[i][0].transcript
    }
    model.value = transcript
  }

  recognition.onerror = () => {
    isRecording.value = false
    isRecognizing.value = false
  }

  recognition.onend = () => {
    isRecording.value = false
    isRecognizing.value = false
    // If we have text from recognition, auto-send
    if (model.value.trim()) {
      // Let user manually send — spec says manual send
    }
  }

  recognition.start()
  // iOS/Safari may trigger onend immediately on timeout
  setTimeout(() => {
    if (isRecording.value) {
      isRecognizing.value = true
    }
  }, 300)
}

function stopRecording() {
  if (recognition) {
    recognition.stop()
    isRecording.value = false
    isRecognizing.value = true
  }
}

function toggleRecording() {
  if (isRecording.value) {
    stopRecording()
  } else {
    startRecording()
  }
}

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
      v-if="recognitionSupported"
      class="min-w-[44px] min-h-[44px] flex items-center justify-center rounded-xl text-sm font-medium shrink-0 transition-colors"
      :class="
        isRecording
          ? 'bg-red-500 text-white animate-pulse'
          : isRecognizing
            ? 'bg-amber-100 text-amber-700'
            : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
      "
      :disabled="disabled"
      @click="toggleRecording"
      :title="isRecording ? '停止录音' : isRecognizing ? '识别中...' : '语音输入'"
    >
      <svg v-if="isRecognizing" xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 animate-spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10" stroke-dasharray="32" stroke-dashoffset="8"/></svg>
      <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2a3 3 0 0 0-3 3v7a3 3 0 0 0 6 0V5a3 3 0 0 0-3-3Z"/><path d="M19 10v2a7 7 0 0 1-14 0v-2"/><line x1="12" x2="12" y1="19" y2="22"/></svg>
    </button>
    <button
      class="min-w-[44px] min-h-[44px] px-6 bg-blue-600 text-white rounded-xl hover:bg-blue-700 disabled:opacity-50 text-sm font-medium shrink-0"
      :disabled="disabled || !model.trim()"
      @click="$emit('send')"
    >
      发送
    </button>
  </div>
</template>
