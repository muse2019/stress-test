<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '@/api/client'
import type { RealtimeStats, WSMessage, Task } from '@/types'
import StatsPanel from '@/components/StatsPanel.vue'
import StatsChart from '@/components/StatsChart.vue'
import { VideoPause, VideoPlay, RefreshRight } from '@element-plus/icons-vue'

const route = useRoute()
const taskId = computed(() => route.params.id as string)

const task = ref<Task | null>(null)
const stats = ref<RealtimeStats | null>(null)
const timeline = ref<RealtimeStats[]>([])
const loading = ref(false)
const wsConnected = ref(false)
const testStatus = ref<'idle' | 'running' | 'completed'>('idle')

// WebSocket message handler
const handleWSMessage = (msg: WSMessage) => {
  switch (msg.type) {
    case 'connected':
      wsConnected.value = true
      break
    case 'started':
      testStatus.value = 'running'
      break
    case 'stats':
      stats.value = msg.data
      timeline.value.push(msg.data)
      // Keep only last 60 data points (1 minute at 1s interval)
      if (timeline.value.length > 60) {
        timeline.value.shift()
      }
      break
    case 'completed':
      testStatus.value = 'completed'
      stats.value = msg.data
      api.disconnectWebSocket()
      wsConnected.value = false
      break
    case 'error':
      console.error('Test error:', msg.data)
      testStatus.value = 'idle'
      break
  }
}

// Start WebSocket connection
const connectWebSocket = () => {
  if (!taskId.value) return
  api.connectWebSocket(taskId.value, handleWSMessage)
}

// Start the test
const startTest = async () => {
  if (!taskId.value) return
  loading.value = true
  try {
    await api.startTask(taskId.value)
    testStatus.value = 'running'
    connectWebSocket()
  } catch (error) {
    console.error('Failed to start test:', error)
  } finally {
    loading.value = false
  }
}

// Stop the test
const stopTest = async () => {
  if (!taskId.value) return
  loading.value = true
  try {
    await api.stopTask(taskId.value)
    testStatus.value = 'completed'
    api.disconnectWebSocket()
    wsConnected.value = false
  } catch (error) {
    console.error('Failed to stop test:', error)
  } finally {
    loading.value = false
  }
}

// Reset the monitor
const resetMonitor = () => {
  stats.value = null
  timeline.value = []
  testStatus.value = 'idle'
}

// Fetch task details
const fetchTask = async () => {
  if (!taskId.value) return
  try {
    task.value = await api.getTask(taskId.value)
    if (task.value?.status === 'running') {
      testStatus.value = 'running'
      connectWebSocket()
    }
  } catch (error) {
    console.error('Failed to fetch task:', error)
  }
}

onMounted(() => {
  fetchTask()
})

onUnmounted(() => {
  api.disconnectWebSocket()
})
</script>

<template>
  <div class="monitor-page">
    <!-- Header -->
    <div class="monitor-header">
      <div class="header-info">
        <h2>实时监控</h2>
        <div class="task-info" v-if="task">
          <el-tag :type="task.protocol === 'http' ? 'primary' : 'success'" size="small">
            {{ task.protocol.toUpperCase() }}
          </el-tag>
          <span class="task-name">{{ task.name }}</span>
          <span class="task-target">{{ task.target }}</span>
        </div>
      </div>
      <div class="header-actions">
        <el-tag :type="wsConnected ? 'success' : 'info'" effect="plain">
          <el-icon class="status-dot" :class="{ connected: wsConnected }">
            <span class="dot"></span>
          </el-icon>
          {{ wsConnected ? '已连接' : '未连接' }}
        </el-tag>
        <el-button
          v-if="testStatus === 'idle'"
          type="primary"
          :icon="VideoPlay"
          :loading="loading"
          @click="startTest"
        >
          开始测试
        </el-button>
        <el-button
          v-if="testStatus === 'running'"
          type="danger"
          :icon="VideoPause"
          :loading="loading"
          @click="stopTest"
        >
          停止测试
        </el-button>
        <el-button
          v-if="testStatus === 'completed'"
          :icon="RefreshRight"
          @click="resetMonitor"
        >
          重置
        </el-button>
      </div>
    </div>

    <!-- Status Alert -->
    <el-alert
      v-if="testStatus === 'running'"
      title="测试运行中..."
      type="info"
      :closable="false"
      show-icon
      class="status-alert"
    />
    <el-alert
      v-if="testStatus === 'completed'"
      title="测试完成"
      type="success"
      :closable="false"
      show-icon
      class="status-alert"
    />

    <!-- Stats Panel -->
    <StatsPanel :stats="stats" />

    <!-- Charts -->
    <StatsChart
      v-if="timeline.length > 0"
      :timeline="timeline"
      title="QPS与响应时间趋势"
      height="400px"
    />

    <!-- Empty State -->
    <el-empty
      v-if="!stats && testStatus === 'idle'"
      description="开始测试以查看实时指标"
    >
      <el-button type="primary" :icon="VideoPlay" @click="startTest">
        Start Test
      </el-button>
    </el-empty>

    <!-- Loading State -->
    <div class="loading-state" v-if="testStatus === 'running' && !stats">
      <el-progress :percentage="100" :indeterminate="true" :duration="1" />
      <p>等待指标数据...</p>
    </div>
  </div>
</template>

<style scoped>
.monitor-page {
  padding: 20px;
}

.monitor-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
}

.header-info h2 {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
}

.task-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.task-name {
  font-weight: 500;
  color: #303133;
}

.task-target {
  color: #909399;
  font-size: 14px;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.status-dot {
  display: inline-flex;
  align-items: center;
  margin-right: 4px;
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: #909399;
}

.status-dot.connected .dot {
  background-color: #67c23a;
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.status-alert {
  margin-bottom: 20px;
}

.loading-state {
  text-align: center;
  padding: 40px;
}

.loading-state p {
  margin-top: 16px;
  color: #909399;
}
</style>
