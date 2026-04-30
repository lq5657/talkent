<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { api, ApiError, type ReportResponse, type ReportSummary } from '../api/client'
import DimensionCard from '../components/DimensionCard.vue'
import MarkdownRenderer from '../components/MarkdownRenderer.vue'

const route = useRoute()
const sessionId = route.params.id as string

const report = ref<ReportResponse | null>(null)
const historyReports = ref<ReportSummary[]>([])
const loading = ref(true)
const analyzing = ref(false)
const error = ref('')
const noReportYet = ref(false)

async function loadReport() {
  loading.value = true
  error.value = ''
  noReportYet.value = false
  try {
    report.value = await api.getReport(sessionId)
  } catch (e) {
    if (e instanceof ApiError && e.status === 404) {
      report.value = null
      noReportYet.value = true
    } else {
      error.value = e instanceof Error ? e.message : '加载报告失败'
      report.value = null
    }
  } finally {
    loading.value = false
  }
}

async function loadHistory() {
  try {
    historyReports.value = await api.getReports(sessionId)
  } catch {
    // silently fail
  }
}

async function triggerAnalysis() {
  analyzing.value = true
  error.value = ''
  try {
    const res = await api.analyze(sessionId)
    report.value = res
    noReportYet.value = false
    await loadHistory()
  } catch (e) {
    error.value = e instanceof Error ? e.message : '生成分析失败'
  } finally {
    analyzing.value = false
  }
}

onMounted(() => {
  loadReport()
  loadHistory()
})
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <div class="max-w-3xl mx-auto px-4 py-6 md:py-8 space-y-6">
      <h1 class="text-xl md:text-2xl font-bold text-gray-900">分析报告</h1>

      <!-- Error -->
      <div
        v-if="error"
        class="p-3 md:p-4 rounded-lg bg-red-50 border border-red-200 text-red-700 text-sm flex items-center justify-between gap-2"
      >
        <span>{{ error }}</span>
        <button
          class="shrink-0 px-3 py-1 bg-red-600 text-white rounded-md hover:bg-red-700 text-xs font-medium"
          :disabled="analyzing"
          @click="triggerAnalysis"
        >
          重试
        </button>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="text-gray-500 text-sm">加载报告中...</div>

      <!-- No report — generate button -->
      <div v-else-if="noReportYet && !report" class="text-center py-12 space-y-4">
        <p class="text-gray-500">对话已结束，尚未生成分析报告</p>
        <button
          class="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 text-sm font-medium"
          :disabled="analyzing"
          @click="triggerAnalysis"
        >
          {{ analyzing ? '分析生成中...' : '生成分析报告' }}
        </button>
      </div>

      <!-- Report content -->
      <template v-else-if="report">
        <!-- Dimension cards -->
        <section v-if="report.dimensions.length > 0" class="space-y-4">
          <h2 class="text-lg font-semibold text-gray-700">维度分析</h2>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
            <DimensionCard
              v-for="dim in report.dimensions"
              :key="dim.name"
              :dimension="dim"
            />
          </div>
        </section>

        <!-- Markdown report -->
        <section class="space-y-3">
          <h2 class="text-lg font-semibold text-gray-700">详细报告</h2>
          <div class="p-4 md:p-6 bg-white rounded-lg border border-gray-200">
            <MarkdownRenderer :content="report.markdown" />
          </div>
        </section>

        <!-- Meta info -->
        <div class="text-xs text-gray-400">
          分析模型：{{ report.model_used }} · 生成时间：{{ report.created_at }}
        </div>
      </template>

      <!-- History reports -->
      <section v-if="historyReports.length > 1" class="space-y-3 pt-6 border-t border-gray-200">
        <h2 class="text-lg font-semibold text-gray-700">历史报告</h2>
        <div class="space-y-2">
          <div
            v-for="hr in historyReports"
            :key="hr.report_id"
            class="flex items-center justify-between p-3 rounded-lg border border-gray-200 bg-white cursor-pointer hover:bg-gray-50"
            :class="report && hr.report_id === report.report_id ? 'ring-2 ring-blue-500' : ''"
            @click="loadReport"
          >
            <div class="text-sm text-gray-600">
              报告 #{{ hr.report_id }} · {{ hr.model_used }}
            </div>
            <div class="text-xs text-gray-400">{{ hr.created_at }}</div>
          </div>
        </div>
      </section>

      <!-- Back link -->
      <div class="pt-4">
        <router-link
          to="/"
          class="text-blue-600 hover:text-blue-800 text-sm"
        >
          &larr; 返回设定页
        </router-link>
      </div>
    </div>
  </div>
</template>
