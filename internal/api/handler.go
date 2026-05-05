package api

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"stress-test/internal/scheduler"
	"stress-test/internal/store"
	"stress-test/pkg/models"
)

// APIError 统一错误响应
type APIError struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

// writeError 写入错误响应
func writeError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(APIError{
		Error:   message,
		Code:    code,
		Message: message,
	})
}

// Handler API 处理器
type Handler struct {
	taskStore   store.TaskStore
	reportStore store.ReportStore
	schedulers  map[string]*scheduler.Scheduler
	wsHub       *WebSocketHub
	mu          sync.RWMutex
}

// NewHandler 创建处理器
func NewHandler(taskStore store.TaskStore, reportStore store.ReportStore, wsHub *WebSocketHub) *Handler {
	return &Handler{
		taskStore:   taskStore,
		reportStore: reportStore,
		schedulers:  make(map[string]*scheduler.Scheduler),
		wsHub:       wsHub,
	}
}

// ListTasks 获取任务列表
func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.taskStore.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 添加运行状态
	type TaskWithStatus struct {
		*models.Task
		Status string `json:"status"`
	}

	h.mu.RLock()
	// Initialize as empty slice to avoid null in JSON
	result := make([]TaskWithStatus, 0)
	for _, t := range tasks {
		status := "idle"
		if s, ok := h.schedulers[t.ID]; ok && s.IsRunning() {
			status = "running"
		}
		result = append(result, TaskWithStatus{Task: t, Status: status})
	}
	h.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"tasks": result})
}

// CreateTask 创建任务
func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := task.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	task.GenerateID()
	if err := h.taskStore.Save(&task); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// GetTask 获取任务详情
func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	task, err := h.taskStore.Get(id)
	if err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// UpdateTask 更新任务
func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// 检查任务是否正在运行
	h.mu.RLock()
	if s, ok := h.schedulers[id]; ok && s.IsRunning() {
		h.mu.RUnlock()
		http.Error(w, "cannot update running task", http.StatusConflict)
		return
	}
	h.mu.RUnlock()

	existing, err := h.taskStore.Get(id)
	if err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	task.ID = existing.ID
	if err := task.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.taskStore.Save(&task); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// DeleteTask 删除任务
func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// 检查任务是否正在运行
	h.mu.RLock()
	if s, ok := h.schedulers[id]; ok && s.IsRunning() {
		h.mu.RUnlock()
		http.Error(w, "cannot delete running task", http.StatusConflict)
		return
	}
	h.mu.RUnlock()

	if err := h.taskStore.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// StartTask 开始压测
func (h *Handler) StartTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	h.mu.Lock()
	defer h.mu.Unlock()

	task, err := h.taskStore.Get(id)
	if err != nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	if s, ok := h.schedulers[id]; ok && s.IsRunning() {
		http.Error(w, "task is already running", http.StatusConflict)
		return
	}

	// 创建带广播器的调度器
	s := scheduler.NewSchedulerWithBroadcaster(task, h.wsHub)
	h.schedulers[id] = s

	// 广播开始消息
	h.wsHub.Broadcast(id, map[string]interface{}{
		"type": "started",
		"data": map[string]interface{}{
			"taskId":    id,
			"startTime": time.Now().Format(time.RFC3339),
		},
	})

	// 保存任务副本用于报告
	taskCopy := *task

	// 启动调度器并在完成后保存报告
	go func() {
		s.Run(context.Background())

		// 测试完成，生成报告
		endTime := time.Now()
		startTime := s.StartTime()
		finalStats := s.GetStats()
		timeline := s.GetTimeline()

		report := &models.Report{
			TaskID:     id,
			TaskName:   taskCopy.Name,
			StartTime:  startTime,
			EndTime:    endTime,
			Duration:   int(endTime.Sub(startTime).Seconds()),
			Config:     taskCopy,
			FinalStats: *finalStats,
			Timeline:   timeline,
		}
		report.GenerateID()

		if err := h.reportStore.Save(report); err != nil {
			// 记录错误但不影响响应
			// TODO: 添加日志
		}

		// 广播完成消息
		h.wsHub.Broadcast(id, map[string]interface{}{
			"type": "completed",
			"data": finalStats,
		})
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "started",
		"taskId": id,
	})
}

// StopTask 停止压测
func (h *Handler) StopTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	h.mu.RLock()
	s, ok := h.schedulers[id]
	if !ok || !s.IsRunning() {
		h.mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "not_running",
			"taskId": id,
		})
		return
	}
	h.mu.RUnlock()

	s.Stop()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "stopped",
		"taskId": id,
	})
}

// ListReports 获取报告列表
func (h *Handler) ListReports(w http.ResponseWriter, r *http.Request) {
	reports, err := h.reportStore.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 简化响应
	type ReportSummary struct {
		ID          string  `json:"id"`
		TaskID      string  `json:"taskId"`
		TaskName    string  `json:"taskName"`
		StartTime   string  `json:"startTime"`
		Duration    int     `json:"duration"`
		SuccessRate float64 `json:"successRate"`
	}

	// Initialize as empty slice to avoid null in JSON
	result := make([]ReportSummary, 0)
	for _, report := range reports {
		result = append(result, ReportSummary{
			ID:          report.ID,
			TaskID:      report.TaskID,
			TaskName:    report.TaskName,
			StartTime:   report.StartTime.Format("2006-01-02T15:04:05Z"),
			Duration:    report.Duration,
			SuccessRate: report.FinalStats.SuccessRate(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"reports": result})
}

// GetReport 获取报告详情
func (h *Handler) GetReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	report, err := h.reportStore.Get(id)
	if err != nil {
		http.Error(w, "report not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// DownloadReport 下载 Markdown 报告
func (h *Handler) DownloadReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	md, err := h.reportStore.GetMarkdown(id)
	if err != nil {
		http.Error(w, "report not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/markdown")
	w.Header().Set("Content-Disposition", "attachment; filename="+id+".md")
	w.Write([]byte(md))
}

// GetTaskStats 获取实时统计
func (h *Handler) GetTaskStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	h.mu.RLock()
	s, ok := h.schedulers[id]
	h.mu.RUnlock()

	if !ok {
		writeError(w, "task not running", http.StatusNotFound)
		return
	}

	stats := s.GetStats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// HealthCheck 健康检查
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
	})
}

// DeleteReport 删除报告
func (h *Handler) DeleteReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.reportStore.Delete(id); err != nil {
		writeError(w, "failed to delete report", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DuplicateTask 复制任务
func (h *Handler) DuplicateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	existing, err := h.taskStore.Get(id)
	if err != nil {
		writeError(w, "task not found", http.StatusNotFound)
		return
	}

	// 创建副本
	newTask := *existing
	newTask.ID = ""
	newTask.Name = existing.Name + " (副本)"
	newTask.GenerateID()

	if err := h.taskStore.Save(&newTask); err != nil {
		writeError(w, "failed to duplicate task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTask)
}

// DeleteAllTasks 删除所有任务
func (h *Handler) DeleteAllTasks(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 检查是否有运行中的任务
	for _, s := range h.schedulers {
		if s.IsRunning() {
			writeError(w, "cannot delete tasks while one is running", http.StatusConflict)
			return
		}
	}

	tasks, err := h.taskStore.List()
	if err != nil {
		writeError(w, "failed to list tasks", http.StatusInternalServerError)
		return
	}

	for _, t := range tasks {
		h.taskStore.Delete(t.ID)
	}

	w.WriteHeader(http.StatusNoContent)
}
