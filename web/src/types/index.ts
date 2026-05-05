export interface Task {
  id: string
  name: string
  protocol: string
  target: string
  method: string
  headers?: Record<string, string>
  body?: string
  timeout: number
  mode: 'fixed' | 'staircase' | 'rate'
  concurrency: number
  duration: number
  rate?: number
  status?: 'idle' | 'running'
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

export interface WSMessage {
  type: 'connected' | 'started' | 'stats' | 'completed' | 'error'
  data: any
}
