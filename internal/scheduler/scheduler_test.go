package scheduler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"stress-test/pkg/models"
)

func TestSchedulerFixedMode(t *testing.T) {
	var requestCount int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	task := &models.Task{
		ID:          "test-task",
		Target:      server.URL,
		Method:      "GET",
		Mode:        "fixed",
		Concurrency: 5,
		Duration:    2, // 2 seconds
		Timeout:     5000,
	}

	scheduler := NewScheduler(task)
	defer scheduler.Stop()

	done := make(chan struct{})
	go func() {
		scheduler.Run(context.Background())
		close(done)
	}()

	// Wait for completion
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("scheduler did not complete in time")
	}

	if requestCount < 10 {
		t.Errorf("expected at least 10 requests, got %d", requestCount)
	}

	stats := scheduler.GetStats()
	if stats.TotalRequests < 10 {
		t.Errorf("expected at least 10 total requests in stats, got %d", stats.TotalRequests)
	}
}

func TestSchedulerStop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	task := &models.Task{
		ID:          "test-task",
		Target:      server.URL,
		Method:      "GET",
		Mode:        "fixed",
		Concurrency: 10,
		Duration:    60, // 60 seconds
		Timeout:     5000,
	}

	scheduler := NewScheduler(task)

	done := make(chan struct{})
	go func() {
		scheduler.Run(context.Background())
		close(done)
	}()

	// Wait a while then stop
	time.Sleep(500 * time.Millisecond)
	scheduler.Stop()

	select {
	case <-done:
		// Normal termination
	case <-time.After(2 * time.Second):
		t.Fatal("scheduler did not stop in time")
	}
}
