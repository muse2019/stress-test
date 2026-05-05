<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { Plus, Delete } from '@element-plus/icons-vue'
import type { Task, Variable, Assertion } from '@/types'

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
const activeTab = ref('basic')

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
  duration: 60,
  thinkTime: 0,
  staircase: { start: 1, step: 1, stepTime: 10, max: 10 },
  warmup: { duration: 0, concurrency: 0 },
  retry: { count: 0, delay: 100 },
  variables: [],
  assertions: [],
})

// Headers 编辑
const headerList = ref<Array<{ key: string; value: string }>>([])

// 变量编辑
const variableList = ref<Variable[]>([])

// 断言编辑
const assertionList = ref<Assertion[]>([])

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
const variableTypes = [
  { label: '静态值', value: 'static' },
  { label: '随机整数', value: 'random_int' },
  { label: '随机字符串', value: 'random_string' },
  { label: 'UUID', value: 'uuid' },
]
const assertionTypes = [
  { label: '状态码', value: 'statusCode' },
  { label: '响应时间', value: 'responseTime' },
  { label: '响应体', value: 'body' },
]
const assertionOperators = [
  { label: '等于', value: 'eq' },
  { label: '不等于', value: 'ne' },
  { label: '小于', value: 'lt' },
  { label: '大于', value: 'gt' },
  { label: '小于等于', value: 'lte' },
  { label: '大于等于', value: 'gte' },
  { label: '包含', value: 'contains' },
  { label: '正则匹配', value: 'regex' },
]

const isEdit = computed(() => !!props.task?.id)
const dialogTitle = computed(() => isEdit.value ? '编辑任务' : '新建任务')

// Watch for task prop changes to populate form
watch(() => props.task, (newTask) => {
  if (newTask) {
    formData.value = { ...newTask }
    headerList.value = Object.entries(newTask.headers || {}).map(([key, value]) => ({ key, value }))
    variableList.value = newTask.variables || []
    assertionList.value = newTask.assertions || []
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
    duration: 60,
    thinkTime: 0,
    staircase: { start: 1, step: 1, stepTime: 10, max: 10 },
    warmup: { duration: 0, concurrency: 0 },
    retry: { count: 0, delay: 100 },
    variables: [],
    assertions: [],
  }
  headerList.value = []
  variableList.value = []
  assertionList.value = []
  formRef.value?.resetFields()
}

function addHeader() {
  headerList.value.push({ key: '', value: '' })
}

function removeHeader(index: number) {
  headerList.value.splice(index, 1)
}

function headersToObject(): Record<string, string> {
  const headers: Record<string, string> = {}
  for (const h of headerList.value) {
    if (h.key.trim()) {
      headers[h.key.trim()] = h.value
    }
  }
  return headers
}

function addVariable() {
  variableList.value.push({ name: '', type: 'static', value: '' })
}

function removeVariable(index: number) {
  variableList.value.splice(index, 1)
}

function addAssertion() {
  assertionList.value.push({ type: 'statusCode', operator: 'eq', expected: 200 })
}

function removeAssertion(index: number) {
  assertionList.value.splice(index, 1)
}

async function handleSubmit() {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    const data = {
      ...formData.value,
      headers: headersToObject(),
      variables: variableList.value.length > 0 ? variableList.value : undefined,
      assertions: assertionList.value.length > 0 ? assertionList.value : undefined,
    }
    emit('submit', data)
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
    width="750px"
    :close-on-click-modal="false"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="rules"
      label-width="100px"
      label-position="right"
    >
      <el-tabs v-model="activeTab">
        <!-- 基本配置 -->
        <el-tab-pane label="基本配置" name="basic">
          <el-form-item label="任务名称" prop="name">
            <el-input v-model="formData.name" placeholder="请输入任务名称" />
          </el-form-item>

          <el-form-item label="目标地址" prop="target">
            <el-input v-model="formData.target" placeholder="https://example.com/api" />
          </el-form-item>

          <el-form-item label="请求方法" prop="method">
            <el-select v-model="formData.method" placeholder="请选择方法" style="width: 100%">
              <el-option v-for="method in httpMethods" :key="method" :label="method" :value="method" />
            </el-select>
          </el-form-item>

          <el-form-item label="请求头">
            <div class="headers-editor">
              <div v-for="(header, index) in headerList" :key="index" class="header-row">
                <el-input v-model="header.key" placeholder="Header Name" class="header-key" />
                <el-input v-model="header.value" placeholder="Header Value" class="header-value" />
                <el-button type="danger" :icon="Delete" circle size="small" @click="removeHeader(index)" />
              </div>
              <el-button type="primary" :icon="Plus" size="small" @click="addHeader">添加请求头</el-button>
            </div>
          </el-form-item>

          <el-form-item label="请求体" prop="body" v-if="['POST', 'PUT', 'PATCH'].includes(formData.method || '')">
            <el-input v-model="formData.body" type="textarea" :rows="4" placeholder='{"key": "value"}' />
          </el-form-item>

          <el-form-item label="超时时间" prop="timeout">
            <el-input-number v-model="formData.timeout" :min="100" :max="300000" :step="1000" style="width: 100%" />
            <span class="unit-label">毫秒</span>
          </el-form-item>
        </el-tab-pane>

        <!-- 压测模式 -->
        <el-tab-pane label="压测模式" name="mode">
          <el-form-item label="模式">
            <el-radio-group v-model="formData.mode">
              <el-radio value="fixed">固定并发</el-radio>
              <el-radio value="staircase">阶梯递增</el-radio>
              <el-radio value="rate">QPS限制</el-radio>
            </el-radio-group>
          </el-form-item>

          <!-- 固定并发 -->
          <template v-if="formData.mode === 'fixed'">
            <el-form-item label="并发数" prop="concurrency">
              <el-input-number v-model="formData.concurrency" :min="1" :max="10000" style="width: 100%" />
            </el-form-item>
          </template>

          <!-- 阶梯递增 -->
          <template v-if="formData.mode === 'staircase' && formData.staircase">
            <el-form-item label="起始并发">
              <el-input-number v-model="formData.staircase.start" :min="1" :max="10000" style="width: 100%" />
            </el-form-item>
            <el-form-item label="递增步长">
              <el-input-number v-model="formData.staircase.step" :min="1" :max="1000" style="width: 100%" />
            </el-form-item>
            <el-form-item label="每步时间">
              <el-input-number v-model="formData.staircase.stepTime" :min="1" :max="3600" style="width: 100%" />
              <span class="unit-label">秒</span>
            </el-form-item>
            <el-form-item label="最大并发">
              <el-input-number v-model="formData.staircase.max" :min="1" :max="10000" style="width: 100%" />
            </el-form-item>
          </template>

          <!-- QPS 限制 -->
          <template v-if="formData.mode === 'rate'">
            <el-form-item label="QPS 限制">
              <el-input-number v-model="formData.rate" :min="1" :max="100000" style="width: 100%" />
              <span class="unit-label">请求/秒</span>
            </el-form-item>
            <el-form-item label="并发数">
              <el-input-number v-model="formData.concurrency" :min="1" :max="10000" style="width: 100%" />
            </el-form-item>
          </template>

          <el-form-item label="持续时间" prop="duration">
            <el-input-number v-model="formData.duration" :min="1" :max="86400" style="width: 100%" />
            <span class="unit-label">秒</span>
          </el-form-item>

          <el-form-item label="思考时间">
            <el-input-number v-model="formData.thinkTime" :min="0" :max="60000" style="width: 100%" />
            <span class="unit-label">毫秒</span>
          </el-form-item>
        </el-tab-pane>

        <!-- 高级配置 -->
        <el-tab-pane label="高级配置" name="advanced">
          <el-divider content-position="left">预热阶段</el-divider>
          <el-form-item label="预热时间">
            <el-input-number v-model="formData.warmup!.duration" :min="0" :max="300" style="width: 100%" />
            <span class="unit-label">秒 (0 表示不预热)</span>
          </el-form-item>
          <el-form-item label="预热并发" v-if="formData.warmup!.duration > 0">
            <el-input-number v-model="formData.warmup!.concurrency" :min="0" :max="10000" style="width: 100%" />
            <span class="unit-label">0 表示正式并发的 10%</span>
          </el-form-item>

          <el-divider content-position="left">错误重试</el-divider>
          <el-form-item label="重试次数">
            <el-input-number v-model="formData.retry!.count" :min="0" :max="10" style="width: 100%" />
          </el-form-item>
          <el-form-item label="重试间隔" v-if="formData.retry!.count > 0">
            <el-input-number v-model="formData.retry!.delay" :min="0" :max="10000" :step="100" style="width: 100%" />
            <span class="unit-label">毫秒</span>
          </el-form-item>
        </el-tab-pane>

        <!-- 变量配置 -->
        <el-tab-pane label="变量" name="variables">
          <div class="variables-editor">
            <div v-for="(variable, index) in variableList" :key="index" class="variable-row">
              <el-input v-model="variable.name" placeholder="变量名" style="width: 120px" />
              <el-select v-model="variable.type" style="width: 120px">
                <el-option v-for="t in variableTypes" :key="t.value" :label="t.label" :value="t.value" />
              </el-select>
              <el-input v-if="variable.type === 'static'" v-model="variable.value" placeholder="值" style="flex: 1" />
              <template v-else-if="variable.type === 'random_int'">
                <el-input-number v-model="variable.min" placeholder="最小值" style="width: 100px" />
                <el-input-number v-model="variable.max" placeholder="最大值" style="width: 100px" />
              </template>
              <el-input-number v-else-if="variable.type === 'random_string'" v-model="variable.min" placeholder="长度" style="width: 100px" />
              <el-button type="danger" :icon="Delete" circle size="small" @click="removeVariable(index)" />
            </div>
            <el-button type="primary" :icon="Plus" size="small" @click="addVariable">添加变量</el-button>
            <div class="tip">变量使用方式：在 URL、请求头、请求体中使用 {{ '${varName}' }}</div>
          </div>
        </el-tab-pane>

        <!-- 断言配置 -->
        <el-tab-pane label="断言" name="assertions">
          <div class="assertions-editor">
            <div v-for="(assertion, index) in assertionList" :key="index" class="assertion-row">
              <el-select v-model="assertion.type" style="width: 120px">
                <el-option v-for="t in assertionTypes" :key="t.value" :label="t.label" :value="t.value" />
              </el-select>
              <el-select v-model="assertion.operator" style="width: 100px">
                <el-option v-for="o in assertionOperators" :key="o.value" :label="o.label" :value="o.value" />
              </el-select>
              <el-input v-model.number="assertion.expected" placeholder="期望值" style="flex: 1" />
              <el-button type="danger" :icon="Delete" circle size="small" @click="removeAssertion(index)" />
            </div>
            <el-button type="primary" :icon="Plus" size="small" @click="addAssertion">添加断言</el-button>
          </div>
        </el-tab-pane>
      </el-tabs>
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
.headers-editor, .variables-editor, .assertions-editor {
  width: 100%;
}

.header-row, .variable-row, .assertion-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  align-items: center;
}

.header-key {
  width: 150px;
}

.header-value {
  flex: 1;
}

.unit-label {
  margin-left: 8px;
  color: #909399;
  font-size: 12px;
}

.tip {
  margin-top: 8px;
  font-size: 12px;
  color: #909399;
}

:deep(.el-tabs__content) {
  max-height: 400px;
  overflow-y: auto;
}
</style>
