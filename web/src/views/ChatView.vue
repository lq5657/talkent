<script setup lang="ts">
import { ref, nextTick, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, chatStream, type ChatResponse } from '../api/client'
import MessageBubble from '../components/MessageBubble.vue'
import ChatInput from '../components/ChatInput.vue'

const route = useRoute()
const router = useRouter()
const sessionId = route.params.id as string

interface Message {
  role: 'user' | 'ai'
  content: string
  timestamp: Date
}

const messages = ref<Message[]>([])
const inputText = ref('')
const sending = ref(false)
const ending = ref(false)
const error = ref('')
const autoEndNotice = ref('')
const roundCurrent = ref(0)
const roundLimit = ref(0)
const lastUserMessage = ref('')
const messagesContainer = ref<HTMLElement | null>(null)
const roleName = ref('AI')
const streamAbort = ref<AbortController | null>(null)

function speakReply(text: string) {
  if (!('speechSynthesis' in window)) return
  speechSynthesis.cancel()
  const utter = new SpeechSynthesisUtterance(text)
  utter.lang = 'zh-CN'
  speechSynthesis.speak(utter)
}

function cancelStream() {
  streamAbort.value?.abort()
  streamAbort.value = null
}

onBeforeUnmount(() => {
  cancelStream()
  speechSynthesis.cancel()
})

async function scrollToBottom() {
  await nextTick()
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

async function sendMessage() {
  const content = inputText.value.trim()
  if (!content || sending.value) return

  error.value = ''
  autoEndNotice.value = ''
  lastUserMessage.value = content
  const userMsg: Message = { role: 'user', content, timestamp: new Date() }
  messages.value.push(userMsg)
  inputText.value = ''
  sending.value = true
  await scrollToBottom()

  // Streaming mode
  const aiMsg: Message = { role: 'ai', content: '', timestamp: new Date() }
  messages.value.push(aiMsg)
  const controller = new AbortController()
  streamAbort.value = controller
  let streamFailed = false

  try {
    for await (const event of chatStream(sessionId, content, controller.signal)) {
      if (event.type === 'token') {
        aiMsg.content += event.token!
        await scrollToBottom()
      } else if (event.type === 'done') {
        userMsg.timestamp = new Date(event.user_message_created_at!)
        aiMsg.timestamp = new Date(event.assistant_message_created_at!)
        roundCurrent.value = event.round_info!.current
        roundLimit.value = event.round_info!.limit
        await scrollToBottom()

        speakReply(aiMsg.content)

        if (event.round_info!.is_last) {
          autoEndNotice.value = '对话轮数已达上限，正在跳转到分析报告...'
          sending.value = false
          setTimeout(() => router.push(`/report/${sessionId}`), 1500)
          return
        }
      } else if (event.type === 'error') {
        streamFailed = true
        break
      }
    }
  } catch (e) {
    streamFailed = true
  }

  streamAbort.value = null

  if (streamFailed) {
    // Remove empty AI streaming message, fallback to non-streaming
    if (!aiMsg.content) {
      messages.value.pop()
    }
    try {
      const res: ChatResponse = await api.chat(sessionId, content)
      userMsg.timestamp = new Date(res.user_message_created_at)
      messages.value.push({ role: 'ai', content: res.reply, timestamp: new Date(res.assistant_message_created_at) })
      roundCurrent.value = res.round_info.current
      roundLimit.value = res.round_info.limit
      await scrollToBottom()
      speakReply(res.reply)

      if (res.round_info.is_last) {
        autoEndNotice.value = '对话轮数已达上限，正在跳转到分析报告...'
        sending.value = false
        setTimeout(() => router.push(`/report/${sessionId}`), 1500)
        return
      }
    } catch (e2) {
      error.value = e2 instanceof Error ? e2.message : '发送消息失败'
    } finally {
      sending.value = false
    }
    return
  }

  sending.value = false
}

async function endSession() {
  error.value = ''
  ending.value = true
  try {
    await api.endSession(sessionId)
    router.push(`/report/${sessionId}`)
  } catch (e) {
    error.value = e instanceof Error ? e.message : '结束对话失败'
    ending.value = false
  }
}

function retryLastMessage() {
  if (!lastUserMessage.value || sending.value) return
  if (error.value.toLowerCase().includes('completed')) {
    router.push(`/report/${sessionId}`)
    return
  }
  error.value = ''
  inputText.value = lastUserMessage.value
  sendMessage()
}

onMounted(async () => {
  try {
    const session = await api.getSession(sessionId)
    roundLimit.value = session.round_limit
    if (session.role_description) {
      roleName.value = session.role_description
    }
  } catch {
    // session info is supplementary — chat still works
  }
})
</script>

<template>
  <div class="h-screen flex flex-col bg-gray-50">
    <!-- Header -->
    <header class="shrink-0 px-3 md:px-4 py-2 md:py-3 bg-white border-b border-gray-200 flex items-center justify-between gap-2">
      <h1 class="text-base md:text-lg font-semibold text-gray-800 truncate">{{ roleName }}</h1>
      <div class="flex items-center gap-2 md:gap-3 shrink-0">
        <span v-if="roundLimit > 0" class="text-xs md:text-sm text-gray-500">
          第 {{ roundCurrent }} / {{ roundLimit }} 轮
        </span>
        <button
          class="px-3 py-1.5 text-xs md:text-sm bg-red-50 text-red-600 rounded-lg hover:bg-red-100 disabled:opacity-50"
          :disabled="ending || sending"
          @click="endSession"
        >
          {{ ending ? '结束中...' : '结束对话' }}
        </button>
      </div>
    </header>

    <!-- Auto-end notice -->
    <div
      v-if="autoEndNotice"
      class="shrink-0 mx-4 mt-2 p-3 rounded-lg bg-amber-50 border border-amber-200 text-amber-700 text-sm"
    >
      {{ autoEndNotice }}
    </div>

    <!-- Error -->
    <div
      v-if="error"
      class="shrink-0 mx-4 mt-2 p-3 rounded-lg bg-red-50 border border-red-200 text-red-700 text-sm flex items-center justify-between gap-2"
    >
      <span>{{ error }}</span>
      <button
        class="shrink-0 px-3 py-1 bg-red-600 text-white rounded-md hover:bg-red-700 text-xs font-medium"
        :disabled="sending"
        @click="retryLastMessage"
      >
        重试
      </button>
    </div>

    <!-- Messages -->
    <div
      ref="messagesContainer"
      class="flex-1 overflow-y-auto px-4 py-4 space-y-3"
    >
      <MessageBubble
        v-for="(msg, i) in messages"
        :key="i"
        :role="msg.role"
        :content="msg.content"
        :role-name="roleName"
        :start-time="i > 0 ? messages[i - 1]!.timestamp : undefined"
        :end-time="msg.timestamp"
      />
      <div v-if="sending" class="flex justify-start">
        <div class="px-4 py-3 rounded-2xl rounded-bl-md bg-white border border-gray-200 text-sm text-gray-400">
          AI 正在思考...
        </div>
      </div>
      <div v-if="messages.length === 0 && !sending" class="text-center text-gray-400 text-sm mt-12">
        开始对话吧！
      </div>
    </div>

    <!-- Input -->
    <div class="shrink-0 px-4 py-3 bg-white border-t border-gray-200">
      <ChatInput
        v-model="inputText"
        :disabled="sending || ending || !!autoEndNotice"
        @send="sendMessage"
      />
    </div>
  </div>
</template>
