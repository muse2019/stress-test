package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"stress-test/internal/schedule"
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
	taskStore       store.TaskStore
	reportStore     store.ReportStore
	schedulers      map[string]*scheduler.Scheduler
	scheduleManager *schedule.Scheduler
	wsHub           *WebSocketHub
	mu              sync.RWMutex
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

// SetScheduleManager 设置定时调度器
func (h *Handler) SetScheduleManager(sm *schedule.Scheduler) {
	h.scheduleManager = sm
}

// ListTasks 获取任务列表
func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	// 支持按分组筛选
	group := r.URL.Query().Get("group")

	var tasks []*models.Task
	var err error

	if group != "" {
		tasks, err = h.taskStore.ListByGroup(group)
	} else {
		tasks, err = h.taskStore.List()
	}

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

// GetGroups 获取所有分组
func (h *Handler) GetGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := h.taskStore.GetGroups()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"groups": groups})
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

	// 如果启用了定时任务，添加到调度器
	if h.scheduleManager != nil && task.Schedule != nil && task.Schedule.Enabled {
		h.scheduleManager.ScheduleTask(&task)
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

	// 更新定时调度
	if h.scheduleManager != nil {
		if task.Schedule != nil && task.Schedule.Enabled {
			h.scheduleManager.ScheduleTask(&task)
		} else {
			h.scheduleManager.UnscheduleTask(task.ID)
		}
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

	// 从定时调度器中移除
	if h.scheduleManager != nil {
		h.scheduleManager.UnscheduleTask(id)
	}

	w.WriteHeader(http.StatusNoContent)
}

// StartTask 开始压测 (HTTP handler)
func (h *Handler) StartTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.StartTaskByID(id); err != nil {
		if err.Error() == "task not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if err.Error() == "task is already running" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "started",
		"taskId": id,
	})
}

// StartTaskByID 通过 ID 启动任务 (供定时调度器调用)
func (h *Handler) StartTaskByID(id string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	task, err := h.taskStore.Get(id)
	if err != nil {
		return fmt.Errorf("task not found")
	}

	if s, ok := h.schedulers[id]; ok && s.IsRunning() {
		return fmt.Errorf("task is already running")
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
			log.Printf("Failed to save report: %v", err)
		}

		// 广播完成消息
		h.wsHub.Broadcast(id, map[string]interface{}{
			"type": "completed",
			"data": finalStats,
		})
	}()

	return nil
}

// StopTask 停止压测 (HTTP handler)
func (h *Handler) StopTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	status := h.StopTaskByID(id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": status,
		"taskId": id,
	})
}

// StopTaskByID 通过 ID 停止任务 (供定时调度器调用)
func (h *Handler) StopTaskByID(id string) string {
	h.mu.RLock()
	s, ok := h.schedulers[id]
	if !ok || !s.IsRunning() {
		h.mu.RUnlock()
		return "not_running"
	}
	h.mu.RUnlock()

	s.Stop()
	return "stopped"
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

// CompareReports 对比报告
func (h *Handler) CompareReports(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id1 := vars["id1"]
	id2 := vars["id2"]

	report1, err := h.reportStore.Get(id1)
	if err != nil {
		writeError(w, "report 1 not found", http.StatusNotFound)
		return
	}

	report2, err := h.reportStore.Get(id2)
	if err != nil {
		writeError(w, "report 2 not found", http.StatusNotFound)
		return
	}

	// 计算对比数据
	comparison := map[string]interface{}{
		"report1": report1,
		"report2": report2,
		"diff": map[string]interface{}{
			"totalRequests": report2.FinalStats.TotalRequests - report1.FinalStats.TotalRequests,
			"successRate":   report2.FinalStats.SuccessRate() - report1.FinalStats.SuccessRate(),
			"qps":           report2.FinalStats.QPS - report1.FinalStats.QPS,
			"avgRT":         report2.FinalStats.AvgRT - report1.FinalStats.AvgRT,
			"p50":           report2.FinalStats.P50 - report1.FinalStats.P50,
			"p90":           report2.FinalStats.P90 - report1.FinalStats.P90,
			"p95":           report2.FinalStats.P95 - report1.FinalStats.P95,
			"p99":           report2.FinalStats.P99 - report1.FinalStats.P99,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comparison)
}
