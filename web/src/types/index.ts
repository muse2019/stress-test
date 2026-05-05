export interface Task {
  id: string
  name: string
  group?: string
  protocol: string
  target: string
  method: string
  headers?: Record<string, string>
  body?: string
  timeout: number
  mode: 'fixed' | 'staircase' | 'rate'
  concurrency: number
  duration: number
  thinkTime: number
  rate?: number
  staircase?: Staircase
  warmup?: Warmup
  retry?: RetryConfig
  variables?: Variable[]
  assertions?: Assertion[]
  schedule?: Schedule
  status?: 'idle' | 'running'
}

export interface Staircase {
  start: number
  step: number
  stepTime: number
  max: number
}

export interface Warmup {
  duration: number
  concurrency: number
}

export interface RetryConfig {
  count: number
  delay: number
}

export interface Variable {
  name: string
  type: 'static' | 'random_int' | 'random_string' | 'uuid' | 'csv'
  value?: string
  min?: number
  max?: number
  file?: string
  column?: string
}

export interface Assertion {
  type: 'statusCode' | 'responseTime' | 'body'
  operator: 'eq' | 'ne' | 'lt' | 'gt' | 'lte' | 'gte' | 'contains' | 'regex'
  expected: string | number
}

export interface Schedule {
  enabled: boolean
  cron: string
  nextRun?: string
  lastRun?: string
}

export interface RealtimeStats {
  timestamp: number
  totalRequests: number
  successCount: number
  failedCount: number
  qps: number
  avgRT: number
  minRT: number
  maxRT: number
  p50: number
  p90: number
  p95: number
  p99: number
  errors: Record<string, number>
}

export interface ReportSummary {
  id: string
  taskId: string
  taskName: string
  startTime: string
  duration: number
  successRate: number
}

export interface Report {
  id: string
  taskId: string
  taskName: string
  startTime: string
  endTime: string
  duration: number
  config: Task
  finalStats: RealtimeStats
  timeline: RealtimeStats[]
}

export interface ReportComparison {
  report1: Report
  report2: Report
  diff: {
    totalRequests: number
    successRate: number
    qps: number
    avgRT: number
    p50: number
    p90: number
    p95: number
    p99: number
  }
}

export interface WSMessage {
  type: 'connected' | 'started' | 'stats' | 'completed' | 'error' | 'warmup_stats'
  data: any
}
