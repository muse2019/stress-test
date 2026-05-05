<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
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
  ElCheckbox,
} from 'element-plus'
import { Download, View, Document, Timer, Connection, Sort } from '@element-plus/icons-vue'
import { api } from '@/api/client'
import StatsChart from '@/components/StatsChart.vue'
import type { Report, ReportSummary, RealtimeStats, ReportComparison } from '@/types'

const reports = ref<ReportSummary[]>([])
const loading = ref(false)
const detailDialogVisible = ref(false)
const selectedReport = ref<Report | null>(null)
const detailLoading = ref(false)

// 报告对比
const compareMode = ref(false)
const selectedReportIds = ref<string[]>([])
const compareDialogVisible = ref(false)
const comparison = ref<ReportComparison | null>(null)
const compareLoading = ref(false)

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

// Toggle compare mode
const toggleCompareMode = () => {
  compareMode.value = !compareMode.value
  if (!compareMode.value) {
    selectedReportIds.value = []
  }
}

// Handle selection change
const handleSelectionChange = (selection: ReportSummary[]) => {
  if (selection.length > 2) {
    ElMessage.warning('最多只能选择两个报告进行对比')
    return
  }
  selectedReportIds.value = selection.map(r => r.id)
}

// Compare selected reports
const compareSelectedReports = async () => {
  if (selectedReportIds.value.length !== 2) {
    ElMessage.warning('请选择两个报告进行对比')
    return
  }

  compareDialogVisible.value = true
  compareLoading.value = true
  comparison.value = null

  try {
    comparison.value = await api.compareReports(selectedReportIds.value[0], selectedReportIds.value[1])
  } catch (error) {
    ElMessage.error('对比报告失败')
    console.error(error)
    compareDialogVisible.value = false
  } finally {
    compareLoading.value = false
  }
}

// Format diff value
const formatDiff = (value: number, unit: string = ''): string => {
  const prefix = value >= 0 ? '+' : ''
  return `${prefix}${value.toFixed(2)}${unit}`
}

// Get diff class
const getDiffClass = (value: number): string => {
  if (value > 0) return 'diff-positive'
  if (value < 0) return 'diff-negative'
  return 'diff-neutral'
}

// Get avgRT diff class (lower is better)
const getAvgRTDiffClass = (value: number): string => {
  if (value < 0) return 'diff-positive'
  if (value > 0) return 'diff-negative'
  return 'diff-neutral'
}

const canCompare = computed(() => selectedReportIds.value.length === 2)

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
      <div class="header-actions">
        <el-button
          :type="compareMode ? 'primary' : 'default'"
          :icon="Sort"
          @click="toggleCompareMode"
        >
          {{ compareMode ? '取消对比' : '报告对比' }}
        </el-button>
        <el-button
          v-if="compareMode"
          type="success"
          :disabled="!canCompare"
          @click="compareSelectedReports"
        >
          对比选中报告 ({{ selectedReportIds.length }}/2)
        </el-button>
        <el-button @click="fetchReports" :loading="loading">
          刷新
        </el-button>
      </div>
    </div>

    <el-table
      :data="reports"
      v-loading="loading"
      stripe
      style="width: 100%"
      empty-text="暂无报告"
      @selection-change="handleSelectionChange"
    >
      <el-table-column
        v-if="compareMode"
        type="selection"
        width="55"
        :selectable="() => selectedReportIds.length < 2 || selectedReportIds.includes(reports.find(r => r.id === selectedReportIds[0])?.id || '')"
      />

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

    <!-- Compare Dialog -->
    <el-dialog
      v-model="compareDialogVisible"
      title="报告对比"
      width="90%"
      top="5vh"
      destroy-on-close
    >
      <div v-loading="compareLoading" class="compare-content">
        <template v-if="comparison">
          <div class="compare-header">
            <div class="compare-report">
              <h4>报告 1</h4>
              <div>{{ comparison.report1.taskName }}</div>
              <div class="compare-time">{{ formatDateTime(comparison.report1.startTime) }}</div>
            </div>
            <div class="compare-arrow">→</div>
            <div class="compare-report">
              <h4>报告 2</h4>
              <div>{{ comparison.report2.taskName }}</div>
              <div class="compare-time">{{ formatDateTime(comparison.report2.startTime) }}</div>
            </div>
          </div>

          <el-table :data="[
            { label: '总请求数', v1: comparison.report1.finalStats.totalRequests, v2: comparison.report2.finalStats.totalRequests, diff: comparison.diff.totalRequests, unit: '' },
            { label: '成功率', v1: comparison.report1.finalStats.successRate(), v2: comparison.report2.finalStats.successRate(), diff: comparison.diff.successRate, unit: '%' },
            { label: 'QPS', v1: comparison.report1.finalStats.qps, v2: comparison.report2.finalStats.qps, diff: comparison.diff.qps, unit: '' },
            { label: '平均响应时间', v1: comparison.report1.finalStats.avgRT, v2: comparison.report2.finalStats.avgRT, diff: comparison.diff.avgRT, unit: 'ms' },
            { label: 'P50', v1: comparison.report1.finalStats.p50, v2: comparison.report2.finalStats.p50, diff: comparison.diff.p50, unit: 'ms' },
            { label: 'P90', v1: comparison.report1.finalStats.p90, v2: comparison.report2.finalStats.p90, diff: comparison.diff.p90, unit: 'ms' },
            { label: 'P95', v1: comparison.report1.finalStats.p95, v2: comparison.report2.finalStats.p95, diff: comparison.diff.p95, unit: 'ms' },
            { label: 'P99', v1: comparison.report1.finalStats.p99, v2: comparison.report2.finalStats.p99, diff: comparison.diff.p99, unit: 'ms' },
          ]" stripe>
            <el-table-column prop="label" label="指标" width="150" />
            <el-table-column label="报告 1" width="150">
              <template #default="{ row }">
                {{ row.v1.toFixed(2) }}{{ row.unit }}
              </template>
            </el-table-column>
            <el-table-column label="报告 2" width="150">
              <template #default="{ row }">
                {{ row.v2.toFixed(2) }}{{ row.unit }}
              </template>
            </el-table-column>
            <el-table-column label="差异">
              <template #default="{ row }">
                <span :class="row.label.includes('响应时间') || row.label.startsWith('P') ? getAvgRTDiffClass(row.diff) : getDiffClass(row.diff)">
                  {{ formatDiff(row.diff, row.unit) }}
                </span>
              </template>
            </el-table-column>
          </el-table>

          <div class="compare-tip">
            <p>绿色表示改善，红色表示下降（响应时间相反：减少为绿色，增加为红色）</p>
          </div>
        </template>
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

.header-actions {
  display: flex;
  gap: 12px;
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

/* Compare styles */
.compare-content {
  min-height: 200px;
}

.compare-header {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 40px;
  margin-bottom: 24px;
  padding: 20px;
  background: #f5f7fa;
  border-radius: 8px;
}

.compare-report {
  text-align: center;
}

.compare-report h4 {
  margin: 0 0 8px 0;
  color: #303133;
}

.compare-time {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.compare-arrow {
  font-size: 24px;
  color: #909399;
}

.diff-positive {
  color: #67c23a;
  font-weight: 500;
}

.diff-negative {
  color: #f56c6c;
  font-weight: 500;
}

.diff-neutral {
  color: #909399;
}

.compare-tip {
  margin-top: 16px;
  text-align: center;
  color: #909399;
  font-size: 12px;
}
</style>
