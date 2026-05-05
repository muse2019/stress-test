package stats

import (
	"testing"
	"time"

	"stress-test/internal/engine"
)

func TestStatsCollectorRecord(t *testing.T) {
	collector := NewStatsCollector()

	// 记录成功请求
	collector.Record(&engine.Result{
		Success:    true,
		Latency:    20 * time.Millisecond,
		StatusCode: 200,
	})

	// 记录失败请求
	collector.Record(&engine.Result{
		Success:    false,
		Latency:    50 * time.Millisecond,
		StatusCode: 500,
	})

	snapshot := collector.Snapshot()

	if snapshot.TotalRequests != 2 {
		t.Errorf("expected 2 requests, got %d", snapshot.TotalRequests)
	}
	if snapshot.SuccessCount != 1 {
		t.Errorf("expected 1 success, got %d", snapshot.SuccessCount)
	}
	if snapshot.FailedCount != 1 {
		t.Errorf("expected 1 failure, got %d", snapshot.FailedCount)
	}
}

func TestStatsCollectorPercentiles(t *testing.T) {
	collector := NewStatsCollector()

	// 记录 100 个请求，延迟从 1ms 到 100ms
	for i := 1; i <= 100; i++ {
		collector.Record(&engine.Result{
			Success: true,
			Latency: time.Duration(i) * time.Millisecond,
		})
	}

	snapshot := collector.Snapshot()

	if snapshot.MinRT != 1 {
		t.Errorf("expected min 1ms, got %d", snapshot.MinRT)
	}
	if snapshot.MaxRT != 100 {
		t.Errorf("expected max 100ms, got %d", snapshot.MaxRT)
	}
	// P50 应该在 50ms 左右
	if snapshot.P50 < 45 || snapshot.P50 > 55 {
		t.Errorf("P50 should be around 50ms, got %d", snapshot.P50)
	}
}

func TestStatsCollectorQPS(t *testing.T) {
	collector := NewStatsCollector()
	time.Sleep(100 * time.Millisecond) // 让经过一些时间

	for i := 0; i < 100; i++ {
		collector.Record(&engine.Result{Success: true, Latency: time.Millisecond})
	}

	snapshot := collector.Snapshot()

	if snapshot.QPS <= 0 {
		t.Error("QPS should be positive")
	}
}
