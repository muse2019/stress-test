package models

import (
	"encoding/json"
	"testing"
)

func TestTaskJSONMarshal(t *testing.T) {
	task := &Task{
		ID:          "task-001",
		Name:        "Test API",
		Protocol:    "http",
		Target:      "https://api.example.com/users",
		Method:      "GET",
		Mode:        "fixed",
		Concurrency: 100,
		Duration:    60,
		Timeout:     30000,
	}

	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("failed to marshal task: %v", err)
	}

	var unmarshaled Task
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal task: %v", err)
	}

	if unmarshaled.ID != task.ID {
		t.Errorf("ID mismatch: got %s, want %s", unmarshaled.ID, task.ID)
	}
	if unmarshaled.Concurrency != task.Concurrency {
		t.Errorf("Concurrency mismatch: got %d, want %d", unmarshaled.Concurrency, task.Concurrency)
	}
}

func TestTaskValidation(t *testing.T) {
	task := &Task{
		Name:        "Test",
		Target:      "https://example.com",
		Method:      "GET",
		Mode:        "fixed",
		Concurrency: 100,
		Duration:    60,
	}

	if err := task.Validate(); err != nil {
		t.Errorf("valid task failed validation: %v", err)
	}

	invalidTask := &Task{Name: "Invalid"}
	if err := invalidTask.Validate(); err == nil {
		t.Error("invalid task should fail validation")
	}
}

func TestRealtimeStatsCalculation(t *testing.T) {
	stats := &RealtimeStats{
		TotalRequests: 1000,
		SuccessCount:  990,
		FailedCount:   10,
		AvgRT:         25,
		P50:           20,
		P90:           45,
		P95:           78,
		P99:           120,
	}

	if stats.SuccessRate() != 99.0 {
		t.Errorf("SuccessRate: got %f, want 99.0", stats.SuccessRate())
	}
}
