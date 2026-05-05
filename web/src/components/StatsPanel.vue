<script setup lang="ts">
import { computed } from 'vue'
import type { RealtimeStats } from '@/types'
import { TrendCharts, Timer, Document, CircleCheck, Warning } from '@element-plus/icons-vue'

const props = defineProps<{
  stats: RealtimeStats | null
}>()

const successRate = computed(() => {
  if (!props.stats || props.stats.totalRequests === 0) return 0
  return ((props.stats.successCount / props.stats.totalRequests) * 100).toFixed(2)
})

const formatNumber = (num: number): string => {
  if (num >= 1000000) {
    return (num / 1000000).toFixed(2) + 'M'
  }
  if (num >= 1000) {
    return (num / 1000).toFixed(2) + 'K'
  }
  return num.toFixed(0)
}

const formatMs = (ms: number): string => {
  if (ms >= 1000) {
    return (ms / 1000).toFixed(2) + 's'
  }
  return ms.toFixed(2) + 'ms'
}

const getSuccessRateStatus = computed(() => {
  const rate = Number(successRate.value)
  if (rate >= 99) return 'excellent'
  if (rate >= 95) return 'good'
  if (rate >= 80) return 'warning'
  return 'danger'
})

const getPercentileColor = (label: string) => {
  const colors: Record<string, string> = {
    P50: 'from-indigo-500 to-purple-500',
    P90: 'from-purple-500 to-pink-500',
    P95: 'from-pink-500 to-rose-500',
    P99: 'from-rose-500 to-orange-500',
  }
  return colors[label] || 'from-gray-500 to-gray-600'
}
</script>

<template>
  <div class="stats-panel">
    <!-- Main Stats Cards -->
    <div class="stats-grid">
      <!-- Total Requests -->
      <div class="stat-card total">
        <div class="stat-header">
          <span class="stat-title">总请求数</span>
          <div class="stat-icon">
            <el-icon :size="20"><Document /></el-icon>
          </div>
        </div>
        <div class="stat-body">
          <span class="stat-value">{{ stats ? formatNumber(stats.totalRequests) : '0' }}</span>
          <div class="stat-detail" v-if="stats">
            <span class="success">成功: {{ formatNumber(stats.successCount) }}</span>
            <span class="failed">失败: {{ formatNumber(stats.failedCount) }}</span>
          </div>
        </div>
      </div>

      <!-- QPS -->
      <div class="stat-card qps">
        <div class="stat-header">
          <span class="stat-title">QPS</span>
          <div class="stat-icon">
            <el-icon :size="20"><TrendCharts /></el-icon>
          </div>
        </div>
        <div class="stat-body">
          <span class="stat-value">{{ stats ? stats.qps.toFixed(2) : '0' }}</span>
          <span class="stat-unit">请求/秒</span>
        </div>
        <div class="stat-progress" v-if="stats && stats.qps > 0">
          <div class="progress-bar" :style="{ width: Math.min(stats.qps / 10, 100) + '%' }"></div>
        </div>
      </div>

      <!-- Average RT -->
      <div class="stat-card rt">
        <div class="stat-header">
          <span class="stat-title">平均响应时间</span>
          <div class="stat-icon">
            <el-icon :size="20"><Timer /></el-icon>
          </div>
        </div>
        <div class="stat-body">
          <span class="stat-value">{{ stats ? formatMs(stats.avgRT) : '0ms' }}</span>
          <div class="stat-range" v-if="stats">
            <span>Min: {{ formatMs(stats.minRT) }}</span>
            <span>Max: {{ formatMs(stats.maxRT) }}</span>
          </div>
        </div>
      </div>

      <!-- Success Rate -->
      <div class="stat-card" :class="getSuccessRateStatus">
        <div class="stat-header">
          <span class="stat-title">成功率</span>
          <div class="stat-icon">
            <el-icon :size="20">
              <CircleCheck v-if="getSuccessRateStatus !== 'danger'" />
              <Warning v-else />
            </el-icon>
          </div>
        </div>
        <div class="stat-body">
          <span class="stat-value">{{ successRate }}%</span>
          <div class="stat-badge" :class="getSuccessRateStatus">
            {{ getSuccessRateStatus === 'excellent' ? '优秀' : getSuccessRateStatus === 'good' ? '良好' : getSuccessRateStatus === 'warning' ? '警告' : '危险' }}
          </div>
        </div>
      </div>
    </div>

    <!-- Percentile Section -->
    <div class="percentile-section" v-if="stats">
      <div class="section-header">
        <span class="section-title">响应时间分布</span>
        <span class="section-subtitle">百分位统计</span>
      </div>
      <div class="percentile-grid">
        <div class="percentile-item" v-for="(label, index) in ['P50', 'P90', 'P95', 'P99']" :key="label">
          <div class="percentile-header">
            <span class="percentile-label">{{ label }}</span>
            <span class="percentile-desc">{{ index === 0 ? '中位数' : index === 1 ? '较慢10%' : index === 2 ? '较慢5%' : '最慢1%' }}</span>
          </div>
          <div class="percentile-bar" :class="getPercentileColor(label)">
            <div class="bar-fill" :style="{ width: Math.min((stats as any)[label.toLowerCase()] / stats.maxRT * 100, 100) + '%' }"></div>
          </div>
          <div class="percentile-value">
            {{ formatMs((stats as any)[label.toLowerCase()]) }}
          </div>
        </div>
      </div>
    </div>

    <!-- Error Distribution -->
    <div class="error-section" v-if="stats && Object.keys(stats.errors).length > 0">
      <div class="section-header">
        <span class="section-title">错误分布</span>
        <span class="error-count">共 {{ stats.failedCount }} 次错误</span>
      </div>
      <div class="error-list">
        <div class="error-item" v-for="(count, error) in stats.errors" :key="error">
          <div class="error-info">
            <span class="error-type">{{ error }}</span>
            <span class="error-percent">{{ ((count as number) / stats.failedCount * 100).toFixed(1) }}%</span>
          </div>
          <div class="error-bar">
            <div class="error-bar-fill" :style="{ width: ((count as number) / stats.failedCount * 100) + '%' }"></div>
          </div>
          <span class="error-count">{{ count }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.stats-panel {
  margin-bottom: 24px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 20px;
}

.stat-card {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.stat-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
}

.stat-card.total::before { background: linear-gradient(90deg, #667eea, #764ba2); }
.stat-card.qps::before { background: linear-gradient(90deg, #f093fb, #f5576c); }
.stat-card.rt::before { background: linear-gradient(90deg, #4facfe, #00f2fe); }
.stat-card.excellent::before { background: linear-gradient(90deg, #43e97b, #38f9d7); }
.stat-card.good::before { background: linear-gradient(90deg, #4facfe, #00f2fe); }
.stat-card.warning::before { background: linear-gradient(90deg, #fa709a, #fee140); }
.stat-card.danger::before { background: linear-gradient(90deg, #f5576c, #f093fb); }

.stat-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.stat-title {
  font-size: 13px;
  color: #6b7280;
  font-weight: 500;
}

.stat-icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.stat-card.total .stat-icon { background: linear-gradient(135deg, #667eea, #764ba2); }
.stat-card.qps .stat-icon { background: linear-gradient(135deg, #f093fb, #f5576c); }
.stat-card.rt .stat-icon { background: linear-gradient(135deg, #4facfe, #00f2fe); }
.stat-card.excellent .stat-icon { background: linear-gradient(135deg, #43e97b, #38f9d7); }
.stat-card.good .stat-icon { background: linear-gradient(135deg, #4facfe, #00f2fe); }
.stat-card.warning .stat-icon { background: linear-gradient(135deg, #fa709a, #fee140); }
.stat-card.danger .stat-icon { background: linear-gradient(135deg, #f5576c, #f093fb); }

.stat-body {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #1f2937;
  line-height: 1.2;
}

.stat-unit {
  font-size: 12px;
  color: #9ca3af;
}

.stat-detail {
  display: flex;
  gap: 12px;
  font-size: 12px;
  margin-top: 4px;
}

.stat-detail .success { color: #10b981; }
.stat-detail .failed { color: #ef4444; }

.stat-range {
  display: flex;
  gap: 12px;
  font-size: 11px;
  color: #9ca3af;
}

.stat-progress {
  height: 4px;
  background: #f3f4f6;
  border-radius: 2px;
  margin-top: 12px;
  overflow: hidden;
}

.stat-card.qps .progress-bar {
  height: 100%;
  background: linear-gradient(90deg, #f093fb, #f5576c);
  border-radius: 2px;
  transition: width 0.3s ease;
}

.stat-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
  margin-top: 4px;
}

.stat-badge.excellent { background: #d1fae5; color: #059669; }
.stat-badge.good { background: #dbeafe; color: #2563eb; }
.stat-badge.warning { background: #fef3c7; color: #d97706; }
.stat-badge.danger { background: #fee2e2; color: #dc2626; }

/* Percentile Section */
.percentile-section {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  margin-bottom: 16px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: #374151;
}

.section-subtitle {
  font-size: 12px;
  color: #9ca3af;
}

.percentile-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.percentile-item {
  padding: 12px;
  background: #f9fafb;
  border-radius: 8px;
}

.percentile-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.percentile-label {
  font-size: 14px;
  font-weight: 600;
  color: #374151;
}

.percentile-desc {
  font-size: 11px;
  color: #9ca3af;
}

.percentile-bar {
  height: 6px;
  background: #e5e7eb;
  border-radius: 3px;
  margin-bottom: 8px;
  overflow: hidden;
}

.percentile-bar.from-indigo-500 .bar-fill { background: linear-gradient(90deg, #6366f1, #8b5cf6); }
.percentile-bar.from-purple-500 .bar-fill { background: linear-gradient(90deg, #8b5cf6, #a855f7); }
.percentile-bar.from-pink-500 .bar-fill { background: linear-gradient(90deg, #a855f7, #d946ef); }
.percentile-bar.from-rose-500 .bar-fill { background: linear-gradient(90deg, #d946ef, #f97316); }

.bar-fill {
  height: 100%;
  border-radius: 3px;
  transition: width 0.3s ease;
}

.percentile-value {
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
}

/* Error Section */
.error-section {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.error-count {
  font-size: 12px;
  color: #ef4444;
  background: #fee2e2;
  padding: 2px 8px;
  border-radius: 4px;
}

.error-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.error-item {
  display: grid;
  grid-template-columns: 1fr 100px 60px;
  gap: 12px;
  align-items: center;
}

.error-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.error-type {
  font-family: monospace;
  font-size: 13px;
  color: #dc2626;
  background: #fef2f2;
  padding: 4px 8px;
  border-radius: 4px;
}

.error-percent {
  font-size: 12px;
  color: #9ca3af;
}

.error-bar {
  height: 8px;
  background: #fee2e2;
  border-radius: 4px;
  overflow: hidden;
}

.error-bar-fill {
  height: 100%;
  background: linear-gradient(90deg, #f87171, #ef4444);
  border-radius: 4px;
  transition: width 0.3s ease;
}

.error-item .error-count {
  font-size: 14px;
  font-weight: 600;
  color: #dc2626;
  text-align: right;
  background: none;
  padding: 0;
}

/* Responsive */
@media (max-width: 1200px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  .percentile-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  .percentile-grid {
    grid-template-columns: 1fr;
  }
}
</style>
