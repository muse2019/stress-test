<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
import { Plus, Delete, Connection, InfoFilled, Key, Check, Sunny, RefreshRight, Collection, CircleCheck, Clock, Timer } from '@element-plus/icons-vue'
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
  group: '',
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
  schedule: { enabled: false, cron: '' },
})

// Headers 编辑
const headerList = ref<Array<{ key: string; value: string }>>([])

// Query 参数编辑
const queryParams = ref<Array<{ key: string; value: string; enabled: boolean }>>([])

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

const httpMethodOptions = [
  { label: 'GET', value: 'GET' },
  { label: 'POST', value: 'POST' },
  { label: 'PUT', value: 'PUT' },
  { label: 'DELETE', value: 'DELETE' },
  { label: 'PATCH', value: 'PATCH' },
]
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

// 解析 curl 命令
function parseCurlCommand(curlCmd: string): { url: string; method?: string; headers?: Record<string, string>; body?: string } {
  const result: { url: string; method?: string; headers?: Record<string, string>; body?: string } = { url: '' }

  // 移除换行符和多余空格
  let cmd = curlCmd.replace(/\\\s*\n/g, ' ').replace(/\s+/g, ' ')

  // 提取 URL
  const urlMatch = cmd.match(/(?:^|\s)(?:'([^']+)'|"([^"]+)"|(\S+))(?:\s|$)/)
  if (urlMatch) {
    // 查找 curl 后面的第一个参数作为 URL（不是选项的参数）
    const args = cmd.split(/\s+/)
    for (let i = 1; i < args.length; i++) {
      const arg = args[i]
      if (!arg.startsWith('-') && (arg.startsWith('http://') || arg.startsWith('https://'))) {
        result.url = arg.replace(/^['"]|['"]$/g, '')
        break
      }
    }
  }

  // 提取请求方法
  const methodMatch = cmd.match(/-X\s+(['"]?)(\w+)\1/i)
  if (methodMatch) {
    result.method = methodMatch[2].toUpperCase()
  }

  // 提取请求头
  const headers: Record<string, string> = {}
  const headerRegex = /-H\s+(['"])([^'"]+)\1/g
  let headerMatch
  while ((headerMatch = headerRegex.exec(cmd)) !== null) {
    const headerStr = headerMatch[2]
    const colonIdx = headerStr.indexOf(':')
    if (colonIdx > 0) {
      const key = headerStr.substring(0, colonIdx).trim()
      const value = headerStr.substring(colonIdx + 1).trim()
      headers[key] = value
    }
  }
  if (Object.keys(headers).length > 0) {
    result.headers = headers
  }

  // 提取请求体
  const dataMatch = cmd.match(/(?:--data|-d)\s+(['"])([^'"]*)\1/)
  if (dataMatch) {
    result.body = dataMatch[2]
    // 如果有 data 但没有指定方法，默认为 POST
    if (!result.method) {
      result.method = 'POST'
    }
  }

  // 提取 JSON 数据
  const jsonMatch = cmd.match(/(?:--json)\s+(['"])([^'"]*)\1/)
  if (jsonMatch) {
    result.body = jsonMatch[2]
    result.method = result.method || 'POST'
    if (!result.headers) {
      result.headers = {}
    }
    result.headers['Content-Type'] = 'application/json'
  }

  return result
}

// 处理目标地址粘贴
function handleTargetPaste(event: ClipboardEvent) {
  const pastedText = event.clipboardData?.getData('text') || ''
  if (!pastedText.trim()) return

  // 检测是否为 curl 命令
  if (pastedText.trim().startsWith('curl ')) {
    event.preventDefault()
    const parsed = parseCurlCommand(pastedText)

    if (parsed.url) {
      formData.value.target = parsed.url
    }
    if (parsed.method) {
      formData.value.method = parsed.method
    }
    if (parsed.headers) {
      // 合并已有的 headers
      const existingHeaders = headerList.value.filter(h => h.key.trim())
      for (const [key, value] of Object.entries(parsed.headers)) {
        const existing = existingHeaders.find(h => h.key.toLowerCase() === key.toLowerCase())
        if (existing) {
          existing.value = value
        } else {
          headerList.value.push({ key, value })
        }
      }
    }
    if (parsed.body) {
      formData.value.body = parsed.body
    }

    ElMessage.success('已解析 curl 命令')
  }
}

// 解析 URL 查询参数
function parseUrlParams(url: string) {
  try {
    const urlObj = new URL(url)
    const params: Array<{ key: string; value: string; enabled: boolean }> = []

    urlObj.searchParams.forEach((value, key) => {
      params.push({ key, value, enabled: true })
    })

    return {
      baseUrl: urlObj.origin + urlObj.pathname,
      params
    }
  } catch {
    return { baseUrl: url, params: [] }
  }
}

// 从 URL 同步参数到列表，并分离基础 URL
function syncParamsFromUrl() {
  if (!formData.value.target) return

  const result = parseUrlParams(formData.value.target)
  if (result.params.length > 0) {
    queryParams.value = result.params
    // 将目标地址更新为不含参数的基础 URL
    formData.value.target = result.baseUrl
  }
}

// 构建完整的 URL（基础 URL + 参数）
function buildFullUrl(): string {
  if (!formData.value.target) return ''

  const enabledParams = queryParams.value.filter(p => p.enabled && p.key.trim())
  if (enabledParams.length === 0) {
    return formData.value.target
  }

  const queryString = enabledParams
    .map(p => `${encodeURIComponent(p.key.trim())}=${encodeURIComponent(p.value)}`)
    .join('&')

  const separator = formData.value.target.includes('?') ? '&' : '?'
  return `${formData.value.target}${separator}${queryString}`
}

// 添加查询参数
function addQueryParam() {
  queryParams.value.push({ key: '', value: '', enabled: true })
}

// 删除查询参数
function removeQueryParam(index: number) {
  queryParams.value.splice(index, 1)
}

// 监听目标地址变化，解析参数
watch(() => formData.value.target, (newUrl) => {
  if (newUrl && newUrl.includes('?')) {
    syncParamsFromUrl()
  }
}, { immediate: true })

const isEdit = computed(() => !!props.task?.id)
const dialogTitle = computed(() => isEdit.value ? '编辑任务' : '新建任务')

// Watch for task prop changes to populate form
watch(() => props.task, (newTask) => {
  if (newTask) {
    formData.value = {
      ...newTask,
      // 确保嵌套对象有默认值
      warmup: newTask.warmup || { duration: 0, concurrency: 0 },
      retry: newTask.retry || { count: 0, delay: 100 },
      staircase: newTask.staircase || { start: 1, step: 1, stepTime: 10, max: 10 },
      schedule: newTask.schedule || { enabled: false, cron: '' },
    }
    headerList.value = Object.entries(newTask.headers || {}).map(([key, value]) => ({ key, value }))
    variableList.value = newTask.variables || []
    assertionList.value = newTask.assertions || []
    // 解析 URL 参数
    if (newTask.target) {
      syncParamsFromUrl()
    }
  } else {
    resetForm()
  }
}, { immediate: true })

function resetForm() {
  formData.value = {
    name: '',
    group: '',
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
    schedule: { enabled: false, cron: '' },
  }
  headerList.value = []
  queryParams.value = []
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
      target: buildFullUrl(), // 合并基础 URL 和参数
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
    width="850px"
    :close-on-click-modal="false"
    class="task-form-dialog"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="rules"
      label-width="90px"
      label-position="left"
      class="task-form"
    >
      <el-tabs v-model="activeTab" type="border-card">
        <!-- 基本配置 -->
        <el-tab-pane label="基本配置" name="basic">
          <div class="form-section">
            <div class="section-title">请求配置</div>
            <div class="form-row">
              <el-form-item label="任务名称" prop="name" class="form-item-half">
                <el-input v-model="formData.name" placeholder="请输入任务名称" />
              </el-form-item>
              <el-form-item label="分组" class="form-item-half">
                <el-input v-model="formData.group" placeholder="可选分组" />
              </el-form-item>
            </div>

            <el-form-item label="目标地址" prop="target">
              <el-input
                v-model="formData.target"
                placeholder="https://api.example.com/endpoint 或粘贴 curl 命令"
                @paste="handleTargetPaste"
                @blur="syncParamsFromUrl"
                class="target-input"
              >
                <template #prepend>
                  <el-icon><Connection /></el-icon>
                </template>
              </el-input>
              <div class="input-tip">
                <el-icon><InfoFilled /></el-icon>
                支持粘贴 curl 命令自动解析
              </div>
            </el-form-item>

            <!-- Query 参数 -->
            <el-form-item label="Query 参数" v-if="queryParams.length > 0">
              <div class="params-editor">
                <div class="editor-header">
                  <el-tag size="small" type="info">{{ queryParams.filter(p => p.enabled).length }} 个参数</el-tag>
                  <el-button type="primary" :icon="Plus" size="small" text @click="addQueryParam">添加</el-button>
                </div>
                <div class="params-list">
                  <div v-for="(param, index) in queryParams" :key="index" class="param-row">
                    <el-checkbox v-model="param.enabled" />
                    <el-input v-model="param.key" placeholder="参数名" class="param-key" />
                    <span class="param-equal">=</span>
                    <el-input v-model="param.value" placeholder="参数值" class="param-value" />
                    <el-button type="danger" :icon="Delete" circle size="small" text @click="removeQueryParam(index)" />
                  </div>
                </div>
              </div>
            </el-form-item>

            <div class="form-row">
              <el-form-item label="请求方法" prop="method" class="form-item-half">
                <el-segmented v-model="formData.method" :options="httpMethodOptions" block />
              </el-form-item>
              <el-form-item label="超时时间" prop="timeout" class="form-item-half">
                <el-input-number v-model="formData.timeout" :min="100" :max="300000" :step="1000" style="width: 100%">
                  <template #suffix>ms</template>
                </el-input-number>
              </el-form-item>
            </div>

            <el-form-item label="请求头">
              <div class="headers-editor">
                <div v-for="(header, index) in headerList" :key="index" class="header-row">
                  <el-input v-model="header.key" placeholder="Header Name" class="header-key">
                    <template #prefix>
                      <el-icon><Key /></el-icon>
                    </template>
                  </el-input>
                  <el-input v-model="header.value" placeholder="Header Value" class="header-value" />
                  <el-button type="danger" :icon="Delete" circle size="small" text @click="removeHeader(index)" />
                </div>
                <el-button type="primary" :icon="Plus" size="small" @click="addHeader">添加请求头</el-button>
              </div>
            </el-form-item>

            <el-form-item label="请求体" prop="body" v-if="['POST', 'PUT', 'PATCH'].includes(formData.method || '')">
              <el-input v-model="formData.body" type="textarea" :rows="4" placeholder='{"key": "value"}' class="body-input" />
            </el-form-item>
          </div>
        </el-tab-pane>

        <!-- 压测模式 -->
        <el-tab-pane label="压测模式" name="mode">
          <div class="form-section">
            <div class="section-title">并发策略</div>
            <el-form-item label="压测模式">
              <el-radio-group v-model="formData.mode" class="mode-radio-group">
                <el-radio-button value="fixed">
                  <div class="mode-option">
                    <span class="mode-label">固定并发</span>
                    <span class="mode-desc">恒定并发数</span>
                  </div>
                </el-radio-button>
                <el-radio-button value="staircase">
                  <div class="mode-option">
                    <span class="mode-label">阶梯递增</span>
                    <span class="mode-desc">逐步加压</span>
                  </div>
                </el-radio-button>
                <el-radio-button value="rate">
                  <div class="mode-option">
                    <span class="mode-label">QPS 限制</span>
                    <span class="mode-desc">流量控制</span>
                  </div>
                </el-radio-button>
              </el-radio-group>
            </el-form-item>

            <!-- 固定并发 -->
            <div v-if="formData.mode === 'fixed'" class="mode-config">
              <el-form-item label="并发数" prop="concurrency">
                <el-slider v-model="formData.concurrency" :min="1" :max="1000" show-input :show-input-controls="false" />
              </el-form-item>
            </div>

            <!-- 阶梯递增 -->
            <div v-if="formData.mode === 'staircase' && formData.staircase" class="mode-config">
              <div class="form-row">
                <el-form-item label="起始并发" class="form-item-half">
                  <el-input-number v-model="formData.staircase.start" :min="1" :max="10000" style="width: 100%" />
                </el-form-item>
                <el-form-item label="递增步长" class="form-item-half">
                  <el-input-number v-model="formData.staircase.step" :min="1" :max="1000" style="width: 100%" />
                </el-form-item>
              </div>
              <div class="form-row">
                <el-form-item label="每步时间" class="form-item-half">
                  <el-input-number v-model="formData.staircase.stepTime" :min="1" :max="3600" style="width: 100%">
                    <template #suffix>秒</template>
                  </el-input-number>
                </el-form-item>
                <el-form-item label="最大并发" class="form-item-half">
                  <el-input-number v-model="formData.staircase.max" :min="1" :max="10000" style="width: 100%" />
                </el-form-item>
              </div>
            </div>

            <!-- QPS 限制 -->
            <div v-if="formData.mode === 'rate'" class="mode-config">
              <div class="form-row">
                <el-form-item label="QPS 限制" class="form-item-half">
                  <el-input-number v-model="formData.rate" :min="1" :max="100000" style="width: 100%">
                    <template #suffix>req/s</template>
                  </el-input-number>
                </el-form-item>
                <el-form-item label="并发数" class="form-item-half">
                  <el-input-number v-model="formData.concurrency" :min="1" :max="10000" style="width: 100%" />
                </el-form-item>
              </div>
            </div>

            <div class="section-title" style="margin-top: 24px;">时间配置</div>
            <div class="form-row">
              <el-form-item label="持续时间" prop="duration" class="form-item-half">
                <el-input-number v-model="formData.duration" :min="1" :max="86400" style="width: 100%">
                  <template #suffix>秒</template>
                </el-input-number>
              </el-form-item>
              <el-form-item label="思考时间" class="form-item-half">
                <el-input-number v-model="formData.thinkTime" :min="0" :max="60000" style="width: 100%">
                  <template #suffix>ms</template>
                </el-input-number>
              </el-form-item>
            </div>
          </div>
        </el-tab-pane>

        <!-- 高级配置 -->
        <el-tab-pane label="高级配置" name="advanced">
          <div class="form-section">
            <div class="section-title">
              <el-icon><Sunny /></el-icon>
              预热阶段
            </div>
            <div class="form-row">
              <el-form-item label="预热时间" class="form-item-half">
                <el-input-number v-model="formData.warmup!.duration" :min="0" :max="300" style="width: 100%">
                  <template #suffix>秒</template>
                </el-input-number>
                <div class="form-item-tip">0 表示不预热</div>
              </el-form-item>
              <el-form-item label="预热并发" class="form-item-half" v-if="formData.warmup?.duration">
                <el-input-number v-model="formData.warmup!.concurrency" :min="0" :max="10000" style="width: 100%" />
                <div class="form-item-tip">0 表示正式并发的 10%</div>
              </el-form-item>
            </div>

            <div class="section-title">
              <el-icon><RefreshRight /></el-icon>
              错误重试
            </div>
            <div class="form-row">
              <el-form-item label="重试次数" class="form-item-half">
                <el-input-number v-model="formData.retry!.count" :min="0" :max="10" style="width: 100%" />
              </el-form-item>
              <el-form-item label="重试间隔" class="form-item-half" v-if="formData.retry?.count">
                <el-input-number v-model="formData.retry!.delay" :min="0" :max="10000" :step="100" style="width: 100%">
                  <template #suffix>ms</template>
                </el-input-number>
              </el-form-item>
            </div>
          </div>
        </el-tab-pane>

        <!-- 变量配置 -->
        <el-tab-pane label="变量" name="variables">
          <div class="form-section">
            <div class="section-title">
              <el-icon><Collection /></el-icon>
              参数化变量
            </div>
            <div class="variables-editor">
              <div v-for="(variable, index) in variableList" :key="index" class="variable-row">
                <el-input v-model="variable.name" placeholder="变量名" style="width: 140px">
                  <template #prefix>$</template>
                </el-input>
                <el-select v-model="variable.type" style="width: 130px">
                  <el-option v-for="t in variableTypes" :key="t.value" :label="t.label" :value="t.value" />
                </el-select>
                <el-input v-if="variable.type === 'static'" v-model="variable.value" placeholder="静态值" style="flex: 1" />
                <template v-else-if="variable.type === 'random_int'">
                  <el-input-number v-model="variable.min" placeholder="最小" style="width: 100px" />
                  <span class="range-separator">~</span>
                  <el-input-number v-model="variable.max" placeholder="最大" style="width: 100px" />
                </template>
                <el-input-number v-else-if="variable.type === 'random_string'" v-model="variable.min" placeholder="长度" style="width: 120px" />
                <el-button type="danger" :icon="Delete" circle size="small" text @click="removeVariable(index)" />
              </div>
              <el-button type="primary" :icon="Plus" size="small" @click="addVariable">添加变量</el-button>
              <div class="editor-tip">
                <el-icon><InfoFilled /></el-icon>
                使用方式：在 URL、请求头、请求体中使用 <code>${"{varName}"}</code>
              </div>
            </div>
          </div>
        </el-tab-pane>

        <!-- 断言配置 -->
        <el-tab-pane label="断言" name="assertions">
          <div class="form-section">
            <div class="section-title">
              <el-icon><CircleCheck /></el-icon>
              结果验证
            </div>
            <div class="assertions-editor">
              <div v-for="(assertion, index) in assertionList" :key="index" class="assertion-row">
                <el-select v-model="assertion.type" style="width: 120px">
                  <el-option v-for="t in assertionTypes" :key="t.value" :label="t.label" :value="t.value" />
                </el-select>
                <el-select v-model="assertion.operator" style="width: 110px">
                  <el-option v-for="o in assertionOperators" :key="o.value" :label="o.label" :value="o.value" />
                </el-select>
                <el-input v-model.number="assertion.expected" placeholder="期望值" style="flex: 1" />
                <el-button type="danger" :icon="Delete" circle size="small" text @click="removeAssertion(index)" />
              </div>
              <el-button type="primary" :icon="Plus" size="small" @click="addAssertion">添加断言</el-button>
            </div>
          </div>
        </el-tab-pane>

        <!-- 定时计划 -->
        <el-tab-pane label="定时计划" name="schedule">
          <div class="form-section">
            <div class="section-title">
              <el-icon><Clock /></el-icon>
              定时执行
            </div>
            <el-form-item label="启用定时">
              <el-switch v-model="formData.schedule!.enabled" active-text="开启" inactive-text="关闭" />
            </el-form-item>
            <template v-if="formData.schedule?.enabled">
              <el-form-item label="Cron 表达式">
                <el-input v-model="formData.schedule!.cron" placeholder="0 0 * * * *">
                  <template #prefix>
                    <el-icon><Timer /></el-icon>
                  </template>
                </el-input>
                <div class="cron-tip">
                  格式: 秒 分 时 日 月 周 (例: <code>0 30 9 * * *</code> 每天9:30执行)
                </div>
              </el-form-item>
              <el-form-item label="常用预设">
                <div class="preset-buttons">
                  <el-button size="small" @click="formData.schedule!.cron = '0 0 * * * *'">每小时</el-button>
                  <el-button size="small" @click="formData.schedule!.cron = '0 0 9 * * *'">每天9点</el-button>
                  <el-button size="small" @click="formData.schedule!.cron = '0 0 9 * * 1-5'">工作日9点</el-button>
                  <el-button size="small" @click="formData.schedule!.cron = '0 0 0 * * *'">每天凌晨</el-button>
                </div>
              </el-form-item>
              <el-form-item label="下次执行" v-if="formData.schedule?.nextRun">
                <el-tag type="success">{{ formData.schedule.nextRun }}</el-tag>
              </el-form-item>
            </template>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleCancel" size="large">取消</el-button>
        <el-button type="primary" @click="handleSubmit" size="large">
          <el-icon><Check /></el-icon>
          {{ isEdit ? '更新任务' : '创建任务' }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<style scoped>
.task-form {
  padding: 0;
}

.form-section {
  padding: 8px 0;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin-bottom: 16px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.section-title .el-icon {
  color: var(--el-color-primary);
}

.form-row {
  display: flex;
  gap: 24px;
  margin-bottom: 0;
}

.form-item-half {
  flex: 1;
  min-width: 0;
}

.form-item-tip {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
}

.target-input {
  font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
}

.input-tip {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 6px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.input-tip .el-icon {
  font-size: 14px;
}

/* Params Editor */
.params-editor {
  width: 100%;
  background: var(--el-fill-color-lighter);
  border-radius: 8px;
  padding: 12px;
}

.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.params-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.param-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.param-key {
  flex: 1;
}

.param-equal {
  color: var(--el-text-color-secondary);
  font-weight: 500;
}

.param-value {
  flex: 2;
}

/* Headers Editor */
.headers-editor {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.header-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.header-key {
  width: 180px;
}

.header-value {
  flex: 1;
}

/* Variables Editor */
.variables-editor {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.variable-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.range-separator {
  color: var(--el-text-color-secondary);
  font-weight: 500;
}

/* Assertions Editor */
.assertions-editor {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.assertion-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

/* Body Input */
.body-input {
  font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
}

/* Mode Config */
.mode-config {
  background: var(--el-fill-color-lighter);
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
}

.mode-radio-group {
  width: 100%;
}

.mode-radio-group :deep(.el-radio-button) {
  flex: 1;
}

.mode-radio-group :deep(.el-radio-button__inner) {
  width: 100%;
  padding: 12px 8px;
}

.mode-option {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}

.mode-label {
  font-weight: 500;
}

.mode-desc {
  font-size: 11px;
  color: var(--el-text-color-secondary);
}

/* Editor Tip */
.editor-tip {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
  padding: 8px 12px;
  background: var(--el-color-info-light-9);
  border-radius: 6px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.editor-tip code {
  background: var(--el-fill-color);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'SF Mono', Monaco, monospace;
  color: var(--el-color-primary);
}

/* Cron Tip */
.cron-tip {
  margin-top: 6px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.cron-tip code {
  background: var(--el-fill-color-light);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'SF Mono', Monaco, monospace;
}

/* Preset Buttons */
.preset-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

/* Dialog Footer */
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

/* Tabs */
:deep(.el-tabs--border-card) {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  overflow: hidden;
}

:deep(.el-tabs__content) {
  max-height: 450px;
  overflow-y: auto;
  padding: 16px;
}

:deep(.el-tabs__header) {
  background: var(--el-fill-color-lighter);
}

:deep(.el-tabs__item) {
  padding: 0 20px;
  height: 44px;
  line-height: 44px;
}

:deep(.el-tabs__item.is-active) {
  font-weight: 600;
}

/* Form Items */
:deep(.el-form-item) {
  margin-bottom: 18px;
}

:deep(.el-form-item__label) {
  font-weight: 500;
}

/* Number Input */
:deep(.el-input-number) {
  width: 100%;
}

:deep(.el-input-number__decrease),
:deep(.el-input-number__increase) {
  width: 32px;
}

/* Slider */
:deep(.el-slider) {
  padding: 0 12px;
}

:deep(.el-slider__input) {
  width: 100px;
}
</style>
