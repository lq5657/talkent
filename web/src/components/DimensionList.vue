<script setup lang="ts">
import type { Dimension } from '../api/client'

const props = defineProps<{
  modelValue: Dimension[]
  loading: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [dims: Dimension[]]
}>()

function removeDimension(dim: Dimension) {
  emit('update:modelValue', props.modelValue.filter((d: Dimension) => d.name !== dim.name))
}
</script>

<template>
  <div class="space-y-4">
    <h3 class="text-lg font-semibold text-gray-700">分析维度</h3>

    <div v-if="loading" class="text-gray-500 text-sm">正在推荐分析维度...</div>
    <div v-else-if="modelValue.length > 0" class="space-y-2">
      <div
        v-for="dim in modelValue"
        :key="dim.name"
        class="flex items-start gap-2 p-3 rounded-lg border border-blue-200 bg-blue-50"
      >
        <div>
          <div class="font-medium text-gray-800">{{ dim.name }}</div>
          <div v-if="dim.description" class="text-sm text-gray-500">{{ dim.description }}</div>
        </div>
        <button
          class="ml-auto text-gray-400 hover:text-red-500"
          @click="removeDimension(dim)"
        >
          &times;
        </button>
      </div>
    </div>

    <div v-else class="text-sm text-gray-400">请先选择训练目标后获取推荐维度</div>
  </div>
</template>
