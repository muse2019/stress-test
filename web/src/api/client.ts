import type { Task, Report, ReportSummary, RealtimeStats, WSMessage } from '@/types'

const API_BASE = '/api'

class APIClient {
  private ws: WebSocket | null = null
  private wsCallbacks: Map<string, (msg: WSMessage) => void> = new Map()

  // Task APIs
  async listTasks(): Promise<Task[]> {
    const res = await fetch(`${API_BASE}/tasks`)
    const data = await res.json()
    return data.tasks
  }

  async createTask(task: Partial<Task>): Promise<Task> {
    const res = await fetch(`${API_BASE}/tasks`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(task),
    })
    return res.json()
  }

  async getTask(id: string): Promise<Task> {
    const res = await fetch(`${API_BASE}/tasks/${id}`)
    return res.json()
  }

  async updateTask(id: string, task: Partial<Task>): Promise<Task> {
    const res = await fetch(`${API_BASE}/tasks/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(task),
    })
    return res.json()
  }

  async deleteTask(id: string): Promise<void> {
    await fetch(`${API_BASE}/tasks/${id}`, { method: 'DELETE' })
  }

  async startTask(id: string): Promise<void> {
    await fetch(`${API_BASE}/tasks/${id}/start`, { method: 'POST' })
  }

  async stopTask(id: string): Promise<void> {
    await fetch(`${API_BASE}/tasks/${id}/stop`, { method: 'POST' })
  }

  async getTaskStats(id: string): Promise<RealtimeStats> {
    const res = await fetch(`${API_BASE}/tasks/${id}/stats`)
    return res.json()
  }

  // Report APIs
  async listReports(): Promise<ReportSummary[]> {
    const res = await fetch(`${API_BASE}/reports`)
    const data = await res.json()
    return data.reports
  }

  async getReport(id: string): Promise<Report> {
    const res = await fetch(`${API_BASE}/reports/${id}`)
    return res.json()
  }

  async downloadReport(id: string): Promise<void> {
    window.open(`${API_BASE}/reports/${id}/download`, '_blank')
  }

  // WebSocket
  connectWebSocket(taskId: string, callback: (msg: WSMessage) => void): void {
    const wsUrl = `${location.protocol === 'https:' ? 'wss:' : 'ws:'}//${location.host}/api/ws/stats/${taskId}`
    this.ws = new WebSocket(wsUrl)
    this.wsCallbacks.set(taskId, callback)

    this.ws.onmessage = (event) => {
      const msg: WSMessage = JSON.parse(event.data)
      callback(msg)
    }

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error)
    }

    this.ws.onclose = () => {
      console.log('WebSocket closed')
    }
  }

  disconnectWebSocket(): void {
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }
}

export const api = new APIClient()
