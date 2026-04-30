<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const onLine = ref(navigator.onLine)

function handleOnline() {
  onLine.value = true
}

function handleOffline() {
  onLine.value = false
}

onMounted(() => {
  window.addEventListener('online', handleOnline)
  window.addEventListener('offline', handleOffline)
})

onUnmounted(() => {
  window.removeEventListener('online', handleOnline)
  window.removeEventListener('offline', handleOffline)
})
</script>

<template>
  <div
    v-if="!onLine"
    class="sticky top-0 z-50 px-4 py-2 bg-amber-500 text-white text-center text-sm font-medium"
  >
    当前处于离线状态，部分功能不可用
  </div>
  <router-view />
</template>
