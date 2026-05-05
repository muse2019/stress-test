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
		task:   task,
		engine: createEngine(task),
		stats:  stats.NewStatsCollector(),
		stopCh: make(chan struct{}, 1),
		doneCh: make(chan struct{}),
		timeline: make([]models.RealtimeStats, 0),
	}
}

// NewSchedulerWithBroadcaster 创建带广播器的调度器
func NewSchedulerWithBroadcaster(task *models.Task, b Broadcaster) *Scheduler {
	return &Scheduler{
		task:        task,
		engine:      createEngine(task),
		stats:       stats.NewStatsCollector(),
		stopCh:      make(chan struct{}, 1),
		doneCh:      make(chan struct{}),
		broadcaster: b,
		timeline:    make([]models.RealtimeStats, 0),
	}
}

// createEngine 创建引擎
func createEngine(task *models.Task) engine.Engine {
	timeout := time.Duration(task.Timeout) * time.Millisecond

	// 检查是否配置了重试
	if task.Retry != nil && task.Retry.Count > 0 {
		retryDelay := time.Duration(task.Retry.Delay) * time.Millisecond
		return engine.NewHTTPEngine(timeout, engine.WithRetry(task.Retry.Count, retryDelay))
	}

	return engine.NewHTTPEngine(timeout)
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

	// 预热阶段
	if s.task.Warmup != nil && s.task.Warmup.Duration > 0 {
		s.runWarmup(ctx)
	}

	switch s.task.Mode {
	case "fixed":
		s.runFixed(ctx)
	case "staircase":
		s.runStaircase(ctx)
	case "rate":
		s.runRate(ctx)
	}
}

// runWarmup 预热阶段
func (s *Scheduler) runWarmup(ctx context.Context) {
	warmup := s.task.Warmup
	warmupConcurrency := warmup.Concurrency
	if warmupConcurrency <= 0 {
		// 默认使用正式并发数的 10%
		warmupConcurrency = s.task.Concurrency / 10
		if warmupConcurrency < 1 {
			warmupConcurrency = 1
		}
	}

	thinkTime := time.Duration(s.task.ThinkTime) * time.Millisecond

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 启动预热工作协程
	for i := 0; i < warmupConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					result, _ := s.engine.Execute(ctx, s.task.Target, s.task.Method, s.task.Headers, s.task.Body)
					s.stats.Record(result)

					if thinkTime > 0 {
						select {
						case <-time.After(thinkTime):
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}()
	}

	// 预热定时器
	warmupTimer := time.NewTimer(time.Duration(warmup.Duration) * time.Second)
	defer warmupTimer.Stop()

	// 预热统计广播
	statsTicker := time.NewTicker(1 * time.Second)
	defer statsTicker.Stop()

	for {
		select {
		case <-s.stopCh:
			cancel()
			wg.Wait()
			return
		case <-warmupTimer.C:
			// 预热完成，停止预热协程
			cancel()
			wg.Wait()
			// 重置统计（预热数据不计入正式统计）
			s.stats.Reset()
			return
		case <-statsTicker.C:
			if s.broadcaster != nil {
				snapshot := s.stats.Snapshot()
				s.broadcaster.Broadcast(s.task.ID, map[string]interface{}{
					"type": "warmup_stats",
					"data": snapshot,
				})
			}
		}
	}
}

// runFixed 固定并发模式（带 WebSocket 广播）
func (s *Scheduler) runFixed(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	thinkTime := time.Duration(s.task.ThinkTime) * time.Millisecond

	for i := 0; i < s.task.Concurrency; i++ {
		wg.Add(1)
		go s.worker(ctx, &wg, thinkTime)
	}

	s.runMainLoop(ctx, cancel, &wg)
}

// runStaircase 阶梯递增模式
func (s *Scheduler) runStaircase(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	st := s.task.Staircase
	if st == nil {
		// 如果未配置，使用默认值
		st = &models.Staircase{
			Start:    1,
			Step:     1,
			StepTime: 10,
			Max:      s.task.Concurrency,
		}
	}

	var wg sync.WaitGroup
	thinkTime := time.Duration(s.task.ThinkTime) * time.Millisecond

	// 当前并发数
	currentConcurrency := st.Start

	// 工作协程管理
	type workerCtrl struct {
		stopCh chan struct{}
	}
	workers := make([]*workerCtrl, 0, st.Max)

	// 启动初始并发
	for i := 0; i < currentConcurrency; i++ {
		ctrl := &workerCtrl{stopCh: make(chan struct{})}
		workers = append(workers, ctrl)
		wg.Add(1)
		go s.staircaseWorker(ctx, &wg, thinkTime, ctrl.stopCh)
	}

	// 阶梯递增定时器
	stepTicker := time.NewTicker(time.Duration(st.StepTime) * time.Second)
	defer stepTicker.Stop()

	// 统计广播定时器
	statsTicker := time.NewTicker(1 * time.Second)
	defer statsTicker.Stop()

	// 总超时定时器
	timeout := time.NewTimer(time.Duration(s.task.Duration) * time.Second)
	defer timeout.Stop()

	for {
		select {
		case <-s.stopCh:
			// 停止所有工作协程
			for _, w := range workers {
				close(w.stopCh)
			}
			cancel()
			wg.Wait()
			return

		case <-timeout.C:
			for _, w := range workers {
				close(w.stopCh)
			}
			cancel()
			wg.Wait()
			return

		case <-stepTicker.C:
			// 增加并发
			if currentConcurrency < st.Max {
				addCount := st.Step
				if currentConcurrency+addCount > st.Max {
					addCount = st.Max - currentConcurrency
				}
				for i := 0; i < addCount; i++ {
					ctrl := &workerCtrl{stopCh: make(chan struct{})}
					workers = append(workers, ctrl)
					wg.Add(1)
					go s.staircaseWorker(ctx, &wg, thinkTime, ctrl.stopCh)
				}
				currentConcurrency += addCount
			}

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

// staircaseWorker 阶梯模式工作协程
func (s *Scheduler) staircaseWorker(ctx context.Context, wg *sync.WaitGroup, thinkTime time.Duration, stopCh chan struct{}) {
	defer wg.Done()
	for {
		select {
		case <-stopCh:
			return
		case <-ctx.Done():
			return
		default:
			result, _ := s.engine.Execute(ctx, s.task.Target, s.task.Method, s.task.Headers, s.task.Body)
			s.stats.Record(result)

			if thinkTime > 0 {
				select {
				case <-time.After(thinkTime):
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

// runRate QPS 限制模式
func (s *Scheduler) runRate(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	thinkTime := time.Duration(s.task.ThinkTime) * time.Millisecond
	rateLimit := s.task.Rate
	if rateLimit <= 0 {
		rateLimit = 100
	}

	// 计算每个请求的时间间隔
	interval := time.Second / time.Duration(rateLimit)

	// 请求通道
	requestCh := make(chan struct{}, rateLimit*2)

	// 启动请求生成器
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				close(requestCh)
				return
			case <-ticker.C:
				select {
				case requestCh <- struct{}{}:
				default:
					// 队列满了，跳过
				}
			}
		}
	}()

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < s.task.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case _, ok := <-requestCh:
					if !ok {
						return
					}
					result, _ := s.engine.Execute(ctx, s.task.Target, s.task.Method, s.task.Headers, s.task.Body)
					s.stats.Record(result)

					if thinkTime > 0 {
						select {
						case <-time.After(thinkTime):
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}()
	}

	s.runMainLoop(ctx, cancel, &wg)
}

// worker 普通工作协程
func (s *Scheduler) worker(ctx context.Context, wg *sync.WaitGroup, thinkTime time.Duration) {
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

			if thinkTime > 0 {
				select {
				case <-time.After(thinkTime):
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

// runMainLoop 主循环
func (s *Scheduler) runMainLoop(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) {
	statsTicker := time.NewTicker(1 * time.Second)
	defer statsTicker.Stop()

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
