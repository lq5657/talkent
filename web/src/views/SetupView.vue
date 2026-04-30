<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { api, type TrainingGoal, type Dimension } from '../api/client'
import GoalSelector from '../components/GoalSelector.vue'
import DimensionList from '../components/DimensionList.vue'

const router = useRouter()

// Form state
const roleDescription = ref('')
const scenario = ref('')
const roundLimit = ref(10)
const selectedGoals = ref<TrainingGoal[]>([])
const selectedDimensions = ref<Dimension[]>([])

// API state
const recommendedGoals = ref<TrainingGoal[]>([])
const goalsLoading = ref(false)
const dimsLoading = ref(false)
const creating = ref(false)
const error = ref('')
const dimsWarning = ref('')

const ROLE_TYPE_CUSTOM = 'custom'
const DIMENSION_MODE_DERIVE = 'derive'

// Recommend goals when role description is submitted
async function recommendGoals() {
  if (!roleDescription.value.trim()) {
    error.value = '请输入对话者角色'
    return
  }
  error.value = ''
  dimsWarning.value = ''
  goalsLoading.value = true
  recommendedGoals.value = []
  selectedGoals.value = []
  selectedDimensions.value = []
  try {
    const res = await api.recommendGoals(roleDescription.value.trim())
    recommendedGoals.value = res.goals
  } catch (e) {
    error.value = e instanceof Error ? e.message : '推荐目标失败'
  } finally {
    goalsLoading.value = false
  }
}

// Recommend dimensions when goals change (with debounce)
let dimsTimer: ReturnType<typeof setTimeout> | null = null
watch(selectedGoals, async (goals) => {
  if (dimsTimer) clearTimeout(dimsTimer)
  if (goals.length === 0) {
    selectedDimensions.value = []
    dimsWarning.value = ''
    return
  }
  dimsLoading.value = true
  selectedDimensions.value = []
  dimsWarning.value = ''
  dimsTimer = setTimeout(async () => {
    try {
      const res = await api.recommendDimensions({
        role_type: ROLE_TYPE_CUSTOM,
        goals,
        mode: DIMENSION_MODE_DERIVE,
        role_desc: roleDescription.value.trim(),
      })
      selectedDimensions.value = res.dimensions
    } catch {
      dimsWarning.value = '维度推荐失败，您可以继续对话但将缺少维度分析'
    } finally {
      dimsLoading.value = false
    }
  }, 300)
})

// Create session and navigate to chat
async function createSession() {
  if (selectedGoals.value.length === 0) {
    error.value = '请至少选择一个训练目标'
    return
  }
  error.value = ''
  creating.value = true
  try {
    const res = await api.createSession({
      role_description: roleDescription.value.trim(),
      scenario: scenario.value.trim(),
      role_type: ROLE_TYPE_CUSTOM,
      goals: selectedGoals.value,
      dimensions: selectedDimensions.value,
      round_limit: roundLimit.value,
    })
    router.push(`/chat/${res.session_id}`)
  } catch (e) {
    error.value = e instanceof Error ? e.message : '创建会话失败'
  } finally {
    creating.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <div class="max-w-2xl mx-auto px-4 py-6 md:py-8 space-y-6 md:space-y-8">
      <h1 class="text-2xl font-bold text-gray-900">Talkent - 角色扮演训练</h1>

      <!-- Error alert -->
      <div
        v-if="error"
        class="p-4 rounded-lg bg-red-50 border border-red-200 text-red-700 text-sm"
      >
        {{ error }}
      </div>

      <!-- Dimension recommendation warning -->
      <div
        v-if="dimsWarning"
        class="p-4 rounded-lg bg-amber-50 border border-amber-200 text-amber-700 text-sm"
      >
        {{ dimsWarning }}
      </div>

      <!-- Step 1: Role & Scenario -->
      <section class="space-y-4">
        <h2 class="text-lg font-semibold text-gray-700">第一步：设定对话者与场景</h2>
        <div>
          <label class="block text-sm font-medium text-gray-600 mb-1">对话者角色</label>
          <textarea
            v-model="roleDescription"
            rows="3"
            placeholder="例如：一位资深技术面试官、一个难缠的客户、一位严厉但关心学生的大学教授"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-600 mb-1">场景描述（你的身份和情境）</label>
          <textarea
            v-model="scenario"
            rows="2"
            placeholder="例如：我是一名初级程序员，正在参加技术面试"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
        <button
          class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 text-sm"
          :disabled="goalsLoading || !roleDescription.trim()"
          @click="recommendGoals"
        >
          {{ goalsLoading ? '推荐中...' : '获取推荐目标' }}
        </button>
      </section>

      <!-- Step 2: Goals -->
      <section v-if="recommendedGoals.length > 0 || selectedGoals.length > 0">
        <GoalSelector
          v-model="selectedGoals"
          :recommended="recommendedGoals"
          :loading="goalsLoading"
        />
      </section>

      <!-- Step 3: Dimensions (auto-loaded when goals change) -->
      <section v-if="selectedGoals.length > 0">
        <DimensionList
          v-model="selectedDimensions"
          :loading="dimsLoading"
        />
      </section>

      <!-- Step 4: Round limit & Start -->
      <section v-if="selectedGoals.length > 0" class="space-y-4">
        <h3 class="text-lg font-semibold text-gray-700">对话设置</h3>
        <div>
          <label class="block text-sm font-medium text-gray-600 mb-1">对话轮数限制</label>
          <input
            v-model.number="roundLimit"
            type="number"
            min="1"
            max="50"
            class="w-24 px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <span class="ml-2 text-sm text-gray-500">轮（1-50）</span>
        </div>
        <button
          class="w-full px-6 py-3 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50 text-sm font-medium"
          :disabled="creating || selectedGoals.length === 0"
          @click="createSession"
        >
          {{ creating ? '创建中...' : '开始对话' }}
        </button>
      </section>
    </div>
  </div>
</template>
