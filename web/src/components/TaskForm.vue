<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import type { Task } from '@/types'

const props = defineProps<{
  modelValue: boolean
  task?: Task
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'submit', data: Partial<Task>): void
}>()

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const formRef = ref<FormInstance>()
const formData = ref<Partial<Task>>({
  name: '',
  protocol: 'http',
  target: '',
  method: 'GET',
  headers: {},
  body: '',
  timeout: 30000,
  mode: 'fixed',
  concurrency: 10,
  duration: 60
})

const rules: FormRules = {
  name: [
    { required: true, message: '请输入任务名称', trigger: 'blur' },
    { min: 2, max: 50, message: '名称长度应为 2-50 个字符', trigger: 'blur' }
  ],
  target: [
    { required: true, message: '请输入目标 URL', trigger: 'blur' },
    { type: 'url', message: '请输入有效的 URL', trigger: 'blur' }
  ],
  method: [
    { required: true, message: '请选择 HTTP 方法', trigger: 'change' }
  ],
  concurrency: [
    { required: true, message: '请输入并发数', trigger: 'blur' },
    { type: 'number', min: 1, max: 10000, message: '并发数应为 1-10000', trigger: 'blur' }
  ],
  duration: [
    { required: true, message: '请输入持续时间', trigger: 'blur' },
    { type: 'number', min: 1, max: 86400, message: '持续时间应为 1-86400 秒', trigger: 'blur' }
  ]
}

const httpMethods = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS']

const isEdit = computed(() => !!props.task?.id)

const dialogTitle = computed(() => isEdit.value ? '编辑任务' : '新建任务')

// Watch for task prop changes to populate form
watch(() => props.task, (newTask) => {
  if (newTask) {
    formData.value = { ...newTask }
  } else {
    resetForm()
  }
}, { immediate: true })

function resetForm() {
  formData.value = {
    name: '',
    protocol: 'http',
    target: '',
    method: 'GET',
    headers: {},
    body: '',
    timeout: 30000,
    mode: 'fixed',
    concurrency: 10,
    duration: 60
  }
  formRef.value?.resetFields()
}

async function handleSubmit() {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    emit('submit', { ...formData.value })
    dialogVisible.value = false
    resetForm()
  } catch (error) {
    console.error('Form validation failed:', error)
  }
}

function handleCancel() {
  dialogVisible.value = false
  resetForm()
}
</script>

<template>
  <el-dialog
    v-model="dialogVisible"
    :title="dialogTitle"
    width="600px"
    :close-on-click-modal="false"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="rules"
      label-width="100px"
      label-position="right"
    >
      <el-form-item label="任务名称" prop="name">
        <el-input v-model="formData.name" placeholder="请输入任务名称" />
      </el-form-item>

      <el-form-item label="目标地址" prop="target">
        <el-input v-model="formData.target" placeholder="https://example.com/api" />
      </el-form-item>

      <el-form-item label="请求方法" prop="method">
        <el-select v-model="formData.method" placeholder="请选择方法" style="width: 100%">
          <el-option
            v-for="method in httpMethods"
            :key="method"
            :label="method"
            :value="method"
          />
        </el-select>
      </el-form-item>

      <el-form-item label="并发数" prop="concurrency">
        <el-input-number
          v-model="formData.concurrency"
          :min="1"
          :max="10000"
          style="width: 100%"
        />
      </el-form-item>

      <el-form-item label="持续时间" prop="duration">
        <el-input-number
          v-model="formData.duration"
          :min="1"
          :max="86400"
          style="width: 100%"
        />
        <span style="margin-left: 8px; color: #909399;">秒</span>
      </el-form-item>

      <el-form-item label="超时时间" prop="timeout">
        <el-input-number
          v-model="formData.timeout"
          :min="100"
          :max="300000"
          :step="1000"
          style="width: 100%"
        />
        <span style="margin-left: 8px; color: #909399;">毫秒</span>
      </el-form-item>

      <el-form-item label="压测模式">
        <el-radio-group v-model="formData.mode" disabled>
          <el-radio value="fixed">固定并发</el-radio>
          <el-radio value="staircase">阶梯递增</el-radio>
          <el-radio value="rate">QPS限制</el-radio>
        </el-radio-group>
        <div class="el-form-item__tip">当前仅支持固定并发模式</div>
      </el-form-item>

      <el-form-item label="请求体" prop="body" v-if="['POST', 'PUT', 'PATCH'].includes(formData.method || '')">
        <el-input
          v-model="formData.body"
          type="textarea"
          :rows="4"
          placeholder='{"key": "value"}'
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="handleCancel">取消</el-button>
      <el-button type="primary" @click="handleSubmit">
        {{ isEdit ? '更新' : '创建' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.el-form-item__tip {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
}
</style>
