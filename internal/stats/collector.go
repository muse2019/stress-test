package stats

import (
	"sort"
	"strconv"
	"sync"
	"time"

	"stress-test/internal/engine"
	"stress-test/pkg/models"
)

// StatsCollector 统计收集器
type StatsCollector struct {
	mu            sync.RWMutex
	totalRequests int64
	successCount  int64
	failedCount   int64
	latencies     []int64
	writeIndex    int // ring buffer write position
	count         int // current number of elements in buffer
	errors        map[string]int64
	startTime     time.Time
}

// NewStatsCollector 创建统计收集器
func NewStatsCollector() *StatsCollector {
	return &StatsCollector{
		latencies: make([]int64, 10000), // fixed-size ring buffer
		errors:    make(map[string]int64),
		startTime: time.Now(),
	}
}

// Record 记录请求结果
func (c *StatsCollector) Record(result *engine.Result) {
	if result == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.totalRequests++
	latencyMs := result.Latency.Milliseconds()

	if result.Success {
		c.successCount++
	} else {
		c.failedCount++
		if result.Error != nil {
			key := classifyError(result.Error)
			c.errors[key]++
		} else if result.StatusCode >= 400 {
			key := classifyHTTPError(result.StatusCode)
			c.errors[key]++
		}
	}

	// Use ring buffer with modular indexing
	c.latencies[c.writeIndex] = latencyMs
	c.writeIndex = (c.writeIndex + 1) % 10000
	if c.count < 10000 {
		c.count++
	}
}

// Snapshot 获取当前统计快照
func (c *StatsCollector) Snapshot() *models.RealtimeStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	elapsed := time.Since(c.startTime).Seconds()

	stats := &models.RealtimeStats{
		Timestamp:     time.Now().Unix(),
		TotalRequests: c.totalRequests,
		SuccessCount:  c.successCount,
		FailedCount:   c.failedCount,
		Errors:        make(map[string]int64),
	}

	// Guard against division by zero for QPS
	if elapsed > 0 {
		stats.QPS = float64(c.totalRequests) / elapsed
	}

	// 复制错误统计
	for k, v := range c.errors {
		stats.Errors[k] = v
	}

	// 计算延迟统计
	if c.count > 0 {
		sorted := make([]int64, c.count)
		copy(sorted, c.latencies[:c.count])
		sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

		stats.MinRT = sorted[0]
		stats.MaxRT = sorted[len(sorted)-1]
		// Calculate average from sorted slice (not latencySum which is now removed)
		var sum int64
		for _, v := range sorted {
			sum += v
		}
		stats.AvgRT = sum / int64(len(sorted))
		stats.P50 = sorted[len(sorted)*50/100]
		stats.P90 = sorted[len(sorted)*90/100]
		stats.P95 = sorted[len(sorted)*95/100]
		stats.P99 = sorted[len(sorted)*99/100]
	}

	return stats
}

// Reset 重置统计
func (c *StatsCollector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.totalRequests = 0
	c.successCount = 0
	c.failedCount = 0
	c.latencies = make([]int64, 10000) // reset to fixed size
	c.writeIndex = 0
	c.count = 0
	c.errors = make(map[string]int64)
	c.startTime = time.Now()
}

// classifyError 分类错误
func classifyError(err error) string {
	return "error:" + err.Error()
}

// classifyHTTPError 分类 HTTP 错误
func classifyHTTPError(statusCode int) string {
	return "http:" + strconv.Itoa(statusCode)
}
