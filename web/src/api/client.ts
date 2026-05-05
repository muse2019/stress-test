import { ElMessage } from 'element-plus'
import type { Task, Report, ReportSummary, RealtimeStats, WSMessage, ReportComparison } from '@/types'

const API_BASE = '/api'

// API 错误类型
class APIError extends Error {
  code: number
  constructor(message: string, code: number) {
    super(message)
    this.code = code
    this.name = 'APIError'
  }
}

// 通用请求方法
async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const res = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  })

  if (!res.ok) {
    let message = `请求失败: ${res.status}`
    try {
      const data = await res.json()
      message = data.error || data.message || message
    } catch {
      // ignore
    }
    throw new APIError(message, res.status)
  }

  // 204 No Content
  if (res.status === 204) {
    return undefined as T
  }

  return res.json()
}

class APIClient {
  private ws: WebSocket | null = null
  private wsReconnectTimer: ReturnType<typeof setTimeout> | null = null
  private wsReconnectAttempts = 0
  private wsMaxReconnectAttempts = 5
  private wsReconnectDelay = 2000

  // Task APIs
  async listTasks(): Promise<Task[]> {
    const data = await request<{ tasks: Task[] }>(`${API_BASE}/tasks`)
    return data.tasks || []
  }

  async createTask(task: Partial<Task>): Promise<Task> {
    return request<Task>(`${API_BASE}/tasks`, {
      method: 'POST',
      body: JSON.stringify(task),
    })
  }

  async getTask(id: string): Promise<Task> {
    return request<Task>(`${API_BASE}/tasks/${id}`)
  }

  async updateTask(id: string, task: Partial<Task>): Promise<Task> {
    return request<Task>(`${API_BASE}/tasks/${id}`, {
      method: 'PUT',
      body: JSON.stringify(task),
    })
  }

  async deleteTask(id: string): Promise<void> {
    await request<void>(`${API_BASE}/tasks/${id}`, { method: 'DELETE' })
  }

  async duplicateTask(id: string): Promise<Task> {
    return request<Task>(`${API_BASE}/tasks/${id}/duplicate`, { method: 'POST' })
  }

  async deleteAllTasks(): Promise<void> {
    await request<void>(`${API_BASE}/tasks`, { method: 'DELETE' })
  }

  async startTask(id: string): Promise<void> {
    await request<void>(`${API_BASE}/tasks/${id}/start`, { method: 'POST' })
  }

  async stopTask(id: string): Promise<void> {
    await request<void>(`${API_BASE}/tasks/${id}/stop`, { method: 'POST' })
  }

  async getTaskStats(id: string): Promise<RealtimeStats> {
    return request<RealtimeStats>(`${API_BASE}/tasks/${id}/stats`)
  }

  // Report APIs
  async listReports(): Promise<ReportSummary[]> {
    const data = await request<{ reports: ReportSummary[] }>(`${API_BASE}/reports`)
    return data.reports || []
  }

  async getReport(id: string): Promise<Report> {
    return request<Report>(`${API_BASE}/reports/${id}`)
  }

  async deleteReport(id: string): Promise<void> {
    await request<void>(`${API_BASE}/reports/${id}`, { method: 'DELETE' })
  }

  async downloadReport(id: string): Promise<void> {
    window.open(`${API_BASE}/reports/${id}/download`, '_blank')
  }

  async compareReports(id1: string, id2: string): Promise<ReportComparison> {
    return request<ReportComparison>(`${API_BASE}/reports/compare/${id1}/${id2}`)
  }

  // Health check
  async healthCheck(): Promise<boolean> {
    try {
      await request<{ status: string }>('/health')
      return true
    } catch {
      return false
    }
  }

  // WebSocket with auto-reconnect
  connectWebSocket(taskId: string, callback: (msg: WSMessage) => void): void {
    this.disconnectWebSocket()
    this.wsReconnectAttempts = 0
    this.createWebSocket(taskId, callback)
  }

  private createWebSocket(taskId: string, callback: (msg: WSMessage) => void): void {
    const wsUrl = `${location.protocol === 'https:' ? 'wss:' : 'ws:'}//${location.host}/api/ws/stats/${taskId}`

    try {
      this.ws = new WebSocket(wsUrl)

      this.ws.onopen = () => {
        console.log('WebSocket connected')
        this.wsReconnectAttempts = 0
        callback({ type: 'connected', data: null })
      }

      this.ws.onmessage = (event) => {
        try {
          const msg: WSMessage = JSON.parse(event.data)
          callback(msg)
        } catch (e) {
          console.error('Failed to parse WebSocket message:', e)
        }
      }

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error)
      }

      this.ws.onclose = (event) => {
        console.log('WebSocket closed:', event.code, event.reason)
        this.ws = null

        // 尝试重连
        if (this.wsReconnectAttempts < this.wsMaxReconnectAttempts) {
          this.wsReconnectAttempts++
          console.log(`Attempting to reconnect (${this.wsReconnectAttempts}/${this.wsMaxReconnectAttempts})...`)

          this.wsReconnectTimer = setTimeout(() => {
            this.createWebSocket(taskId, callback)
          }, this.wsReconnectDelay)
        } else {
          callback({ type: 'error', data: { message: 'WebSocket 连接已断开' } })
        }
      }
    } catch (error) {
      console.error('Failed to create WebSocket:', error)
    }
  }

  disconnectWebSocket(): void {
    if (this.wsReconnectTimer) {
      clearTimeout(this.wsReconnectTimer)
      this.wsReconnectTimer = null
    }

    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  isWebSocketConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN
  }
}

export const api = new APIClient()

// 全局错误处理包装器
export async function withErrorHandling<T>(
  fn: () => Promise<T>,
  errorMsg: string
): Promise<T | null> {
  try {
    return await fn()
  } catch (error) {
    const message = error instanceof APIError ? error.message : errorMsg
    ElMessage.error(message)
    console.error(error)
    return null
  }
}
