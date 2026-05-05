package scheduler

import (
	"context"
	"sync"
	"time"

	"stress-test/internal/engine"
	"stress-test/internal/stats"
	"stress-test/pkg/models"
)

// Broadcaster 广播接口（避免循环导入）
type Broadcaster interface {
	Broadcast(taskID string, data interface{})
}

// Scheduler 调度器
type Scheduler struct {
	task        *models.Task
	engine      engine.Engine
	stats       *stats.StatsCollector
	stopCh      chan struct{}
	doneCh      chan struct{}
	running     bool
	mu          sync.Mutex
	broadcaster Broadcaster // 使用接口
	timeline    []models.RealtimeStats
	startTime   time.Time
}

// NewScheduler 创建调度器
func NewScheduler(task *models.Task) *Scheduler {
	return &Scheduler{
		task:     task,
		engine:   engine.NewHTTPEngine(time.Duration(task.Timeout) * time.Millisecond),
		stats:    stats.NewStatsCollector(),
		stopCh:   make(chan struct{}, 1),
		doneCh:   make(chan struct{}),
		timeline: make([]models.RealtimeStats, 0),
	}
}

// NewSchedulerWithBroadcaster 创建带广播器的调度器
func NewSchedulerWithBroadcaster(task *models.Task, b Broadcaster) *Scheduler {
	return &Scheduler{
		task:        task,
		engine:      engine.NewHTTPEngine(time.Duration(task.Timeout) * time.Millisecond),
		stats:       stats.NewStatsCollector(),
		stopCh:      make(chan struct{}, 1),
		doneCh:      make(chan struct{}),
		broadcaster: b,
		timeline:    make([]models.RealtimeStats, 0),
	}
}

// Run 运行压测
func (s *Scheduler) Run(ctx context.Context) {
	s.mu.Lock()
	s.running = true
	s.startTime = time.Now()
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		close(s.doneCh)
	}()

	switch s.task.Mode {
	case "fixed":
		s.runFixed(ctx)
	}
}

// runFixed 固定并发模式（带 WebSocket 广播）
func (s *Scheduler) runFixed(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup

	for i := 0; i < s.task.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-s.stopCh:
					return
				case <-ctx.Done():
					return
				default:
					result, _ := s.engine.Execute(ctx, s.task.Target, s.task.Method, s.task.Headers, s.task.Body)
					s.stats.Record(result)
				}
			}
		}()
	}

	// 启动统计广播协程
	statsTicker := time.NewTicker(1 * time.Second)
	defer statsTicker.Stop()

	// 创建超时定时器
	timeout := time.NewTimer(time.Duration(s.task.Duration) * time.Second)
	defer timeout.Stop()

	for {
		select {
		case <-s.stopCh:
			cancel()
			wg.Wait()
			return
		case <-timeout.C:
			cancel()
			wg.Wait()
			return
		case <-statsTicker.C:
			if s.broadcaster != nil {
				snapshot := s.stats.Snapshot()
				s.mu.Lock()
				s.timeline = append(s.timeline, *snapshot)
				s.mu.Unlock()
				s.broadcaster.Broadcast(s.task.ID, map[string]interface{}{
					"type": "stats",
					"data": snapshot,
				})
			}
		}
	}
}

// Stop 停止压测
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	select {
	case s.stopCh <- struct{}{}:
	default:
	}
}

// GetStats 获取统计信息
func (s *Scheduler) GetStats() *models.RealtimeStats {
	return s.stats.Snapshot()
}

// GetTimeline 获取时间线数据
func (s *Scheduler) GetTimeline() []models.RealtimeStats {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]models.RealtimeStats, len(s.timeline))
	copy(result, s.timeline)
	return result
}

// StartTime 获取开始时间
func (s *Scheduler) StartTime() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.startTime
}

// IsRunning 是否正在运行
func (s *Scheduler) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// Close 关闭调度器
func (s *Scheduler) Close() {
	s.engine.Close()
}

// Done 返回完成通道
func (s *Scheduler) Done() <-chan struct{} {
	return s.doneCh
}
