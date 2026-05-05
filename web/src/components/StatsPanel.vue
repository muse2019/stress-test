<script setup lang="ts">
import { computed } from 'vue'
import type { RealtimeStats } from '@/types'
import { CaretTop, Timer, Document, SuccessFilled, WarningFilled } from '@element-plus/icons-vue'

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
</script>

<template>
  <div class="stats-panel">
    <el-row :gutter="16">
      <!-- Total Requests -->
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon total">
              <el-icon :size="28"><Document /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats ? formatNumber(stats.totalRequests) : '0' }}</div>
              <div class="stat-label">总请求数</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- QPS -->
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon qps">
              <el-icon :size="28"><CaretTop /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats ? stats.qps.toFixed(2) : '0' }}</div>
              <div class="stat-label">QPS</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- Average RT -->
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon rt">
              <el-icon :size="28"><Timer /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats ? formatMs(stats.avgRT) : '0ms' }}</div>
              <div class="stat-label">平均响应时间</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- Success Rate -->
      <el-col :span="6">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon" :class="Number(successRate) >= 99 ? 'success' : 'warning'">
              <el-icon :size="28">
                <SuccessFilled v-if="Number(successRate) >= 99" />
                <WarningFilled v-else />
              </el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ successRate }}%</div>
              <div class="stat-label">成功率</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Percentile Section -->
    <el-card class="percentile-card" shadow="hover" v-if="stats">
      <template #header>
        <div class="card-header">
          <span>响应时间百分位</span>
        </div>
      </template>
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="percentile-item">
            <div class="percentile-label">P50</div>
            <div class="percentile-value">{{ formatMs(stats.p50) }}</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="percentile-item">
            <div class="percentile-label">P90</div>
            <div class="percentile-value">{{ formatMs(stats.p90) }}</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="percentile-item">
            <div class="percentile-label">P95</div>
            <div class="percentile-value">{{ formatMs(stats.p95) }}</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="percentile-item">
            <div class="percentile-label">P99</div>
            <div class="percentile-value">{{ formatMs(stats.p99) }}</div>
          </div>
        </el-col>
      </el-row>
    </el-card>

    <!-- Error Distribution -->
    <el-card class="error-card" shadow="hover" v-if="stats && Object.keys(stats.errors).length > 0">
      <template #header>
        <div class="card-header">
          <span>错误分布</span>
        </div>
      </template>
      <div class="error-list">
        <div class="error-item" v-for="(count, error) in stats.errors" :key="error">
          <span class="error-type">{{ error }}</span>
          <el-tag type="danger" size="small">{{ count }}</el-tag>
        </div>
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.stats-panel {
  margin-bottom: 20px;
}

.stat-card {
  margin-bottom: 16px;
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.stat-icon.total {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.stat-icon.qps {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.stat-icon.rt {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}

.stat-icon.success {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
}

.stat-icon.warning {
  background: linear-gradient(135deg, #fa709a 0%, #fee140 100%);
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 28px;
  font-weight: 600;
  color: #303133;
  line-height: 1.2;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 4px;
}

.percentile-card {
  margin-bottom: 16px;
}

.card-header {
  font-size: 16px;
  font-weight: 500;
}

.percentile-item {
  text-align: center;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 8px;
}

.percentile-label {
  font-size: 14px;
  color: #909399;
  margin-bottom: 8px;
}

.percentile-value {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
}

.error-card {
  margin-bottom: 16px;
}

.error-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.error-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #fef0f0;
  border-radius: 4px;
}

.error-type {
  font-family: monospace;
  font-size: 13px;
  color: #f56c6c;
}
</style>
