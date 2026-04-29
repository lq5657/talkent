<script setup lang="ts">
import type { TrainingGoal } from '../api/client'

const props = defineProps<{
  modelValue: TrainingGoal[]
  recommended: TrainingGoal[]
  loading: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [goals: TrainingGoal[]]
}>()

const customInput = defineModel<string>('customInput', { default: '' })

function toggleGoal(goal: TrainingGoal) {
  const selected = [...props.modelValue]
  const idx = selected.findIndex((g) => g.name === goal.name)
  if (idx >= 0) {
    selected.splice(idx, 1)
  } else {
    selected.push(goal)
  }
  emit('update:modelValue', selected)
}

function addCustomGoal() {
  const name = customInput.value.trim()
  if (!name) return
  if (props.modelValue.some((g) => g.name === name)) return
  const updated = [...props.modelValue, { name, description: '' }]
  emit('update:modelValue', updated)
  customInput.value = ''
}

function removeGoal(goal: TrainingGoal) {
  emit('update:modelValue', props.modelValue.filter((g) => g.name !== goal.name))
}
</script>

<template>
  <div class="space-y-4">
    <h3 class="text-lg font-semibold text-gray-700">训练目标</h3>

    <!-- Recommended goals -->
    <div v-if="loading" class="text-gray-500 text-sm">正在推荐训练目标...</div>
    <div v-else-if="recommended.length > 0" class="space-y-2">
      <div class="text-sm text-gray-500">推荐目标（点击选择）：</div>
      <div
        v-for="goal in recommended"
        :key="goal.name"
        class="flex items-start gap-2 p-3 rounded-lg border cursor-pointer transition-colors"
        :class="
          modelValue.some((g) => g.name === goal.name)
            ? 'border-blue-500 bg-blue-50'
            : 'border-gray-200 hover:border-gray-300'
        "
        @click="toggleGoal(goal)"
      >
        <input
          type="checkbox"
          :checked="modelValue.some((g) => g.name === goal.name)"
          class="mt-1"
          tabindex="-1"
          readonly
        />
        <div>
          <div class="font-medium text-gray-800">{{ goal.name }}</div>
          <div v-if="goal.description" class="text-sm text-gray-500">{{ goal.description }}</div>
        </div>
      </div>
    </div>

    <!-- Selected goals -->
    <div v-if="modelValue.length > 0" class="space-y-2">
      <div class="text-sm text-gray-500">已选目标：</div>
      <div class="flex flex-wrap gap-2">
        <span
          v-for="goal in modelValue"
          :key="goal.name"
          class="inline-flex items-center gap-1 px-3 py-1 rounded-full bg-blue-100 text-blue-800 text-sm"
        >
          {{ goal.name }}
          <button
            class="ml-1 text-blue-500 hover:text-blue-700"
            @click="removeGoal(goal)"
          >
            &times;
          </button>
        </span>
      </div>
    </div>

    <!-- Add custom goal -->
    <div class="flex gap-2">
      <input
        v-model="customInput"
        type="text"
        placeholder="添加自定义目标"
        class="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
        @keydown.enter="addCustomGoal"
      />
      <button
        class="px-4 py-2 text-sm bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 disabled:opacity-50"
        :disabled="!customInput.trim()"
        @click="addCustomGoal"
      >
        添加
      </button>
    </div>
  </div>
</template>
