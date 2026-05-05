<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, VideoPlay, VideoPause, Edit, Delete } from '@element-plus/icons-vue'
import TaskForm from '@/components/TaskForm.vue'
import type { Task } from '@/types'
import { api } from '@/api/client'

const tasks = ref<Task[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const editingTask = ref<Task | undefined>()

async function fetchTasks() {
  loading.value = true
  try {
    tasks.value = await api.listTasks()
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
      <el-button type="primary" :icon="Plus" @click="handleCreate">
        新建任务
      </el-button>
    </div>

    <el-table
      :data="tasks"
      v-loading="loading"
      stripe
      style="width: 100%"
    >
      <el-table-column prop="name" label="任务名称" min-width="150" />

      <el-table-column prop="target" label="目标地址" min-width="200" show-overflow-tooltip />

      <el-table-column prop="method" label="方法" width="100" align="center">
        <template #default="{ row }">
          <el-tag size="small" :type="row.method === 'GET' ? 'success' : row.method === 'POST' ? 'primary' : 'warning'">
            {{ row.method }}
          </el-tag>
        </template>
      </el-table-column>

      <el-table-column prop="concurrency" label="并发数" width="100" align="center" />

      <el-table-column prop="duration" label="持续时间" width="100" align="center">
        <template #default="{ row }">
          {{ row.duration }}秒
        </template>
      </el-table-column>

      <el-table-column prop="status" label="状态" width="100" align="center">
        <template #default="{ row }">
          <el-tag :type="getStatusType(row.status)">
            {{ getStatusText(row.status) }}
          </el-tag>
        </template>
      </el-table-column>

      <el-table-column label="操作" width="240" align="center" fixed="right">
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

.el-button-group {
  display: flex;
  gap: 0;
}
</style>
