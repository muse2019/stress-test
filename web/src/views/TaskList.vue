<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, VideoPlay, VideoPause, Edit, Delete, CopyDocument, Folder } from '@element-plus/icons-vue'
import TaskForm from '@/components/TaskForm.vue'
import type { Task } from '@/types'
import { api } from '@/api/client'

const tasks = ref<Task[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const editingTask = ref<Task | undefined>()
const groups = ref<string[]>([])
const selectedGroup = ref<string>('')

// 按分组筛选后的任务
const filteredTasks = computed(() => {
  if (!selectedGroup.value) return tasks.value
  return tasks.value.filter(t => t.group === selectedGroup.value)
})

// 按分组统计任务数量
const groupStats = computed(() => {
  const stats: Record<string, number> = { '全部': tasks.value.length }
  for (const task of tasks.value) {
    const group = task.group || '未分组'
    stats[group] = (stats[group] || 0) + 1
  }
  return stats
})

async function fetchTasks() {
  loading.value = true
  try {
    tasks.value = await api.listTasks()
    groups.value = await api.getGroups()
  } catch (error) {
    ElMessage.error('获取任务列表失败')
    console.error(error)
  } finally {
    loading.value = false
  }
}

function handleCreate() {
  editingTask.value = undefined
  dialogVisible.value = true
}

function handleEdit(task: Task) {
  editingTask.value = { ...task }
  dialogVisible.value = true
}

async function handleSubmit(data: Partial<Task>) {
  try {
    if (editingTask.value?.id) {
      await api.updateTask(editingTask.value.id, data)
      ElMessage.success('任务更新成功')
    } else {
      await api.createTask(data)
      ElMessage.success('任务创建成功')
    }
    await fetchTasks()
  } catch (error) {
    ElMessage.error(editingTask.value?.id ? '更新任务失败' : '创建任务失败')
    console.error(error)
  }
}

async function handleStart(task: Task) {
  if (!task.id) return
  try {
    await api.startTask(task.id)
    ElMessage.success('任务已启动')
    await fetchTasks()
  } catch (error) {
    ElMessage.error('启动任务失败')
    console.error(error)
  }
}

async function handleStop(task: Task) {
  if (!task.id) return
  try {
    await api.stopTask(task.id)
    ElMessage.success('任务已停止')
    await fetchTasks()
  } catch (error) {
    ElMessage.error('停止任务失败')
    console.error(error)
  }
}

async function handleDelete(task: Task) {
  if (!task.id) return

  try {
    await ElMessageBox.confirm(
      `确定要删除任务 "${task.name}" 吗？`,
      '删除任务',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await api.deleteTask(task.id)
    ElMessage.success('任务已删除')
    await fetchTasks()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除任务失败')
      console.error(error)
    }
  }
}

async function handleDuplicate(task: Task) {
  if (!task.id) return
  try {
    await api.duplicateTask(task.id)
    ElMessage.success('任务已复制')
    await fetchTasks()
  } catch (error) {
    ElMessage.error('复制任务失败')
    console.error(error)
  }
}

function getStatusType(status?: string) {
  switch (status) {
    case 'running':
      return 'success'
    case 'idle':
    default:
      return 'info'
  }
}

function getStatusText(status?: string) {
  switch (status) {
    case 'running':
      return '运行中'
    case 'idle':
    default:
      return '空闲'
  }
}

onMounted(() => {
  fetchTasks()
})
</script>

<template>
  <div class="task-list">
    <div class="header">
      <h1>任务管理</h1>
      <div class="header-actions">
        <el-select
          v-model="selectedGroup"
          placeholder="全部分组"
          clearable
          style="width: 150px; margin-right: 12px;"
        >
          <el-option label="全部" value="" />
          <el-option label="未分组" value="__none__" />
          <el-option
            v-for="group in groups"
            :key="group"
            :label="`${group} (${groupStats[group] || 0})`"
            :value="group"
          />
        </el-select>
        <el-button type="primary" :icon="Plus" @click="handleCreate">
          新建任务
        </el-button>
      </div>
    </div>

    <el-table
      :data="selectedGroup === '__none__' ? tasks.filter(t => !t.group) : filteredTasks"
      v-loading="loading"
      stripe
      style="width: 100%"
    >
      <el-table-column prop="name" label="任务名称" min-width="150">
        <template #default="{ row }">
          <div class="task-name">
            <span>{{ row.name }}</span>
          </div>
        </template>
      </el-table-column>

      <el-table-column prop="group" label="分组" width="120" align="center">
        <template #default="{ row }">
          <el-tag v-if="row.group" size="small" type="info">
            <el-icon style="margin-right: 4px;"><Folder /></el-icon>
            {{ row.group }}
          </el-tag>
          <span v-else class="no-group">-</span>
        </template>
      </el-table-column>

      <el-table-column prop="target" label="目标地址" min-width="200" show-overflow-tooltip />

      <el-table-column prop="method" label="方法" width="80" align="center">
        <template #default="{ row }">
          <el-tag size="small" :type="row.method === 'GET' ? 'success' : row.method === 'POST' ? 'primary' : 'warning'">
            {{ row.method }}
          </el-tag>
        </template>
      </el-table-column>

      <el-table-column prop="concurrency" label="并发数" width="80" align="center" />

      <el-table-column prop="duration" label="持续时间" width="90" align="center">
        <template #default="{ row }">
          {{ row.duration }}秒
        </template>
      </el-table-column>

      <el-table-column prop="status" label="状态" width="90" align="center">
        <template #default="{ row }">
          <el-tag :type="getStatusType(row.status)">
            {{ getStatusText(row.status) }}
          </el-tag>
        </template>
      </el-table-column>

      <el-table-column label="操作" width="260" align="center" fixed="right">
        <template #default="{ row }">
          <el-button-group>
            <el-button
              v-if="row.status !== 'running'"
              size="small"
              type="success"
              :icon="VideoPlay"
              @click="handleStart(row)"
            >
              启动
            </el-button>
            <el-button
              v-else
              size="small"
              type="warning"
              :icon="VideoPause"
              @click="handleStop(row)"
            >
              停止
            </el-button>
            <el-button
              size="small"
              type="primary"
              :icon="Edit"
              @click="handleEdit(row)"
              :disabled="row.status === 'running'"
            >
              编辑
            </el-button>
            <el-button
              size="small"
              type="info"
              :icon="CopyDocument"
              @click="handleDuplicate(row)"
              :disabled="row.status === 'running'"
            >
              复制
            </el-button>
            <el-button
              size="small"
              type="danger"
              :icon="Delete"
              @click="handleDelete(row)"
              :disabled="row.status === 'running'"
            >
              删除
            </el-button>
          </el-button-group>
        </template>
      </el-table-column>
    </el-table>

    <TaskForm
      v-model="dialogVisible"
      :task="editingTask"
      @submit="handleSubmit"
    />
  </div>
</template>

<style scoped>
.task-list {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header h1 {
  margin: 0;
  font-size: 24px;
  color: var(--el-text-color-primary);
}

.header-actions {
  display: flex;
  align-items: center;
}

.task-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.no-group {
  color: #c0c4cc;
}

.el-button-group {
  display: flex;
  gap: 0;
}
</style>
