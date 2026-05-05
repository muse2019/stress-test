<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  ElTable,
  ElTableColumn,
  ElButton,
  ElDialog,
  ElDescriptions,
  ElDescriptionsItem,
  ElTag,
  ElMessage,
  ElEmpty,
} from 'element-plus'
import { Download, View, Document, Timer, Connection } from '@element-plus/icons-vue'
import { api } from '@/api/client'
import StatsChart from '@/components/StatsChart.vue'
import type { Report, ReportSummary, RealtimeStats } from '@/types'

const reports = ref<ReportSummary[]>([])
const loading = ref(false)
const detailDialogVisible = ref(false)
const selectedReport = ref<Report | null>(null)
const detailLoading = ref(false)

// Fetch reports list
const fetchReports = async () => {
  loading.value = true
  try {
    reports.value = await api.listReports()
  } catch (error) {
    ElMessage.error('加载报告列表失败')
    console.error(error)
  } finally {
    loading.value = false
  }
}

// Format duration
const formatDuration = (seconds: number): string => {
  if (seconds < 60) return `${seconds}秒`
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return secs > 0 ? `${mins}分${secs}秒` : `${mins}分`
}

// Format date time
const formatDateTime = (dateStr: string): string => {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

// Get success rate tag type
const getSuccessRateType = (rate: number): 'success' | 'warning' | 'danger' => {
  if (rate >= 95) return 'success'
  if (rate >= 80) return 'warning'
  return 'danger'
}

// View report details
const viewReport = async (report: ReportSummary) => {
  detailDialogVisible.value = true
  detailLoading.value = true
  selectedReport.value = null

  try {
    selectedReport.value = await api.getReport(report.id)
  } catch (error) {
    ElMessage.error('加载报告详情失败')
    console.error(error)
    detailDialogVisible.value = false
  } finally {
    detailLoading.value = false
  }
}

// Download Markdown report
const downloadMarkdown = (reportId: string) => {
  api.downloadReport(reportId)
}

// Get mode text
const getModeText = (mode: string): string => {
  const modeMap: Record<string, string> = {
    fixed: '固定并发',
    staircase: '阶梯递增',
    rate: 'QPS限制',
  }
  return modeMap[mode] || mode
}

// Format stats for display
const formatStats = (stats: RealtimeStats) => {
  return {
    totalRequests: stats.totalRequests.toLocaleString(),
    successCount: stats.successCount.toLocaleString(),
    failedCount: stats.failedCount.toLocaleString(),
    qps: stats.qps.toFixed(2),
    avgRT: `${stats.avgRT}ms`,
    minRT: `${stats.minRT}ms`,
    maxRT: `${stats.maxRT}ms`,
    p50: `${stats.p50}ms`,
    p90: `${stats.p90}ms`,
    p95: `${stats.p95}ms`,
    p99: `${stats.p99}ms`,
  }
}

onMounted(() => {
  fetchReports()
})
</script>

<template>
  <div class="reports-page">
    <div class="page-header">
      <h2>
        <el-icon><Document /></el-icon>
        历史报告
      </h2>
      <el-button @click="fetchReports" :loading="loading">
        刷新
      </el-button>
    </div>

    <el-table
      :data="reports"
      v-loading="loading"
      stripe
      style="width: 100%"
      empty-text="暂无报告"
    >
      <el-table-column
        prop="taskName"
        label="任务名称"
        min-width="200"
      >
        <template #default="{ row }">
          <div class="task-name">
            <el-icon><Connection /></el-icon>
            <span>{{ row.taskName }}</span>
          </div>
        </template>
      </el-table-column>

      <el-table-column
        prop="startTime"
        label="开始时间"
        width="180"
      >
        <template #default="{ row }">
          <div class="time-cell">
            <el-icon><Timer /></el-icon>
            {{ formatDateTime(row.startTime) }}
          </div>
        </template>
      </el-table-column>

      <el-table-column
        prop="duration"
        label="持续时间"
        width="120"
        align="center"
      >
        <template #default="{ row }">
          {{ formatDuration(row.duration) }}
        </template>
      </el-table-column>

      <el-table-column
        prop="successRate"
        label="成功率"
        width="120"
        align="center"
      >
        <template #default="{ row }">
          <el-tag :type="getSuccessRateType(row.successRate)" effect="light">
            {{ row.successRate.toFixed(2) }}%
          </el-tag>
        </template>
      </el-table-column>

      <el-table-column
        label="操作"
        width="180"
        align="center"
        fixed="right"
      >
        <template #default="{ row }">
          <el-button
            type="primary"
            :icon="View"
            size="small"
            @click="viewReport(row)"
            link
          >
            查看
          </el-button>
          <el-button
            type="success"
            :icon="Download"
            size="small"
            @click="downloadMarkdown(row.id)"
            link
          >
            下载
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- Report Details Dialog -->
    <el-dialog
      v-model="detailDialogVisible"
      title="报告详情"
      width="90%"
      top="5vh"
      destroy-on-close
    >
      <div v-loading="detailLoading" class="report-detail">
        <template v-if="selectedReport">
          <!-- Basic Info -->
          <el-descriptions title="基本信息" :column="3" border>
            <el-descriptions-item label="任务名称">
              {{ selectedReport.taskName }}
            </el-descriptions-item>
            <el-descriptions-item label="协议">
              HTTP
            </el-descriptions-item>
            <el-descriptions-item label="目标地址">
              {{ selectedReport.config.target }}
            </el-descriptions-item>
            <el-descriptions-item label="请求方法">
              {{ selectedReport.config.method }}
            </el-descriptions-item>
            <el-descriptions-item label="压测模式">
              {{ getModeText(selectedReport.config.mode) }}
            </el-descriptions-item>
            <el-descriptions-item label="并发数">
              {{ selectedReport.config.concurrency }}
            </el-descriptions-item>
            <el-descriptions-item label="持续时间">
              {{ formatDuration(selectedReport.duration) }}
            </el-descriptions-item>
            <el-descriptions-item label="开始时间">
              {{ formatDateTime(selectedReport.startTime) }}
            </el-descriptions-item>
            <el-descriptions-item label="结束时间">
              {{ formatDateTime(selectedReport.endTime) }}
            </el-descriptions-item>
          </el-descriptions>

          <!-- Performance Stats -->
          <el-descriptions
            title="性能指标"
            :column="4"
            border
            class="stats-section"
          >
            <el-descriptions-item label="总请求数">
              {{ formatStats(selectedReport.finalStats).totalRequests }}
            </el-descriptions-item>
            <el-descriptions-item label="成功数">
              <span class="success-text">
                {{ formatStats(selectedReport.finalStats).successCount }}
              </span>
            </el-descriptions-item>
            <el-descriptions-item label="失败数">
              <span class="error-text">
                {{ formatStats(selectedReport.finalStats).failedCount }}
              </span>
            </el-descriptions-item>
            <el-descriptions-item label="QPS">
              {{ formatStats(selectedReport.finalStats).qps }}
            </el-descriptions-item>
            <el-descriptions-item label="平均响应时间">
              {{ formatStats(selectedReport.finalStats).avgRT }}
            </el-descriptions-item>
            <el-descriptions-item label="最小响应时间">
              {{ formatStats(selectedReport.finalStats).minRT }}
            </el-descriptions-item>
            <el-descriptions-item label="最大响应时间">
              {{ formatStats(selectedReport.finalStats).maxRT }}
            </el-descriptions-item>
            <el-descriptions-item label="成功率">
              <el-tag :type="getSuccessRateType(selectedReport.finalStats.totalRequests > 0 ? (selectedReport.finalStats.successCount / selectedReport.finalStats.totalRequests * 100) : 100)" effect="light">
                {{ selectedReport.finalStats.totalRequests > 0 ? (selectedReport.finalStats.successCount / selectedReport.finalStats.totalRequests * 100).toFixed(2) : '100.00' }}%
              </el-tag>
            </el-descriptions-item>
          </el-descriptions>

          <!-- Response Time Distribution -->
          <el-descriptions
            title="响应时间分布"
            :column="4"
            border
            class="stats-section"
          >
            <el-descriptions-item label="P50">
              {{ formatStats(selectedReport.finalStats).p50 }}
            </el-descriptions-item>
            <el-descriptions-item label="P90">
              {{ formatStats(selectedReport.finalStats).p90 }}
            </el-descriptions-item>
            <el-descriptions-item label="P95">
              {{ formatStats(selectedReport.finalStats).p95 }}
            </el-descriptions-item>
            <el-descriptions-item label="P99">
              {{ formatStats(selectedReport.finalStats).p99 }}
            </el-descriptions-item>
          </el-descriptions>

          <!-- Error Distribution -->
          <el-descriptions
            v-if="Object.keys(selectedReport.finalStats.errors).length > 0"
            title="错误分布"
            :column="1"
            border
            class="stats-section"
          >
            <el-descriptions-item
              v-for="(count, errorType) in selectedReport.finalStats.errors"
              :key="errorType"
              :label="String(errorType)"
            >
              {{ count }} 次
            </el-descriptions-item>
          </el-descriptions>

          <!-- Timeline Chart -->
          <div v-if="selectedReport.timeline && selectedReport.timeline.length > 0" class="chart-section">
            <h3>时间线图表</h3>
            <StatsChart
              :timeline="selectedReport.timeline"
              title="QPS、响应时间与成功率趋势"
              height="400px"
            />
          </div>

          <!-- Download Button -->
          <div class="action-buttons">
            <el-button
              type="success"
              :icon="Download"
              @click="downloadMarkdown(selectedReport.id)"
            >
              下载 Markdown 报告
            </el-button>
          </div>
        </template>
        <el-empty v-else-if="!detailLoading" description="暂无报告数据" />
      </div>
    </el-dialog>
  </div>
</template>

<style scoped>
.reports-page {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  font-size: 20px;
  color: #303133;
}

.task-name {
  display: flex;
  align-items: center;
  gap: 6px;
}

.time-cell {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #606266;
}

.report-detail {
  min-height: 200px;
}

.stats-section {
  margin-top: 24px;
}

.chart-section {
  margin-top: 24px;
}

.chart-section h3 {
  margin-bottom: 16px;
  font-size: 16px;
  color: #303133;
}

.success-text {
  color: #67c23a;
  font-weight: 500;
}

.error-text {
  color: #f56c6c;
  font-weight: 500;
}

.action-buttons {
  margin-top: 24px;
  display: flex;
  justify-content: center;
}

:deep(.el-descriptions__title) {
  font-size: 16px;
  color: #303133;
}
</style>
