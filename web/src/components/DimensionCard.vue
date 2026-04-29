<script setup lang="ts">
import type { DimensionAnalysis } from '../api/client'

defineProps<{
  dimension: DimensionAnalysis
}>()
</script>

<template>
  <div class="p-4 rounded-lg border border-gray-200 bg-white space-y-3">
    <div class="flex items-center justify-between">
      <h4 class="font-semibold text-gray-800">{{ dimension.name }}</h4>
      <span
        class="px-3 py-1 rounded-full text-sm font-medium"
        :class="
          dimension.score >= 80
            ? 'bg-green-100 text-green-700'
            : dimension.score >= 60
              ? 'bg-yellow-100 text-yellow-700'
              : 'bg-red-100 text-red-700'
        "
      >
        {{ dimension.score }}分
      </span>
    </div>
    <p class="text-sm text-gray-600">{{ dimension.comment }}</p>
    <div v-if="dimension.suggestions.length > 0" class="space-y-1">
      <div class="text-sm font-medium text-gray-500">改进建议：</div>
      <ul class="list-disc list-inside text-sm text-gray-600 space-y-1">
        <li v-for="(s, i) in dimension.suggestions" :key="i">{{ s }}</li>
      </ul>
    </div>
  </div>
</template>
