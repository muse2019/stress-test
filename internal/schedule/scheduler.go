package schedule

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"stress-test/internal/store"
	"stress-test/pkg/models"
)

// TaskExecutor 任务执行器接口
type TaskExecutor interface {
	StartTask(taskID string) error
	StopTask(taskID string) error
}

// Scheduler 定时调度器
type Scheduler struct {
	cron     *cron.Cron
	store    store.TaskStore
	executor TaskExecutor
	entryMap map[string]cron.EntryID
	mu       sync.RWMutex
}

// NewScheduler 创建定时调度器
func NewScheduler(store store.TaskStore, executor TaskExecutor) *Scheduler {
	return &Scheduler{
		cron:     cron.New(cron.WithLocation(time.Local)),
		store:    store,
		executor: executor,
		entryMap: make(map[string]cron.EntryID),
	}
}

// Start 启动调度器
func (s *Scheduler) Start() {
	s.cron.Start()
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.cron.Stop()
}

// ScheduleTask 调度任务
func (s *Scheduler) ScheduleTask(task *models.Task) error {
	if task.Schedule == nil || !task.Schedule.Enabled {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 如果已经调度，先移除
	if entryID, ok := s.entryMap[task.ID]; ok {
		s.cron.Remove(entryID)
		delete(s.entryMap, task.ID)
	}

	// 添加新调度
	entryID, err := s.cron.AddFunc(task.Schedule.Cron, func() {
		s.executeTask(task.ID)
	})
	if err != nil {
		return err
	}

	s.entryMap[task.ID] = entryID
	return nil
}

// UnscheduleTask 取消任务调度
func (s *Scheduler) UnscheduleTask(taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, ok := s.entryMap[taskID]; ok {
		s.cron.Remove(entryID)
		delete(s.entryMap, taskID)
	}
}

// executeTask 执行任务
func (s *Scheduler) executeTask(taskID string) {
	log.Printf("Executing scheduled task: %s", taskID)

	if err := s.executor.StartTask(taskID); err != nil {
		log.Printf("Failed to start scheduled task %s: %v", taskID, err)
	}

	// 更新任务的下次运行时间
	task, err := s.store.Get(taskID)
	if err != nil {
		return
	}

	if task.Schedule != nil {
		entryID, ok := s.entryMap[taskID]
		if ok {
			entry := s.cron.Entry(entryID)
			task.Schedule.NextRun = entry.Next.Format(time.RFC3339)
			task.Schedule.LastRun = time.Now().Format(time.RFC3339)
			s.store.Save(task)
		}
	}
}

// GetNextRun 获取任务下次运行时间
func (s *Scheduler) GetNextRun(taskID string) (time.Time, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entryID, ok := s.entryMap[taskID]
	if !ok {
		return time.Time{}, false
	}

	entry := s.cron.Entry(entryID)
	return entry.Next, true
}

// LoadScheduledTasks 加载所有已调度的任务
func (s *Scheduler) LoadScheduledTasks(ctx context.Context) error {
	tasks, err := s.store.List()
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if task.Schedule != nil && task.Schedule.Enabled {
			if err := s.ScheduleTask(task); err != nil {
				log.Printf("Failed to schedule task %s: %v", task.ID, err)
			}
		}
	}

	return nil
}
