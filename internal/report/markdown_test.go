package report

import (
	"testing"
	"time"

	"stress-test/pkg/models"
)

func TestGenerateMarkdown(t *testing.T) {
	report := &models.Report{
		ID:        "report-001",
		TaskID:    "task-001",
		TaskName:  "Test API",
		StartTime: time.Date(2024, 5, 4, 10, 30, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 5, 4, 10, 31, 0, 0, time.UTC),
		Duration:  60,
		Config: models.Task{
			Target:      "https://api.example.com/users",
			Method:      "GET",
			Mode:        "fixed",
			Concurrency: 100,
			Duration:    60,
		},
		FinalStats: models.RealtimeStats{
			TotalRequests: 75000,
			SuccessCount:  74850,
			FailedCount:   150,
			QPS:           1250.5,
			AvgRT:         23,
			MinRT:         5,
			MaxRT:         320,
			P50:           20,
			P90:           45,
			P95:           78,
			P99:           156,
		},
	}

	md := GenerateMarkdown(report)

	if !contains(md, "# 压测报告: Test API") {
		t.Error("missing report title")
	}
	if !contains(md, "https://api.example.com/users") {
		t.Error("missing target URL")
	}
	if !contains(md, "75000") {
		t.Error("missing total requests")
	}
	if !contains(md, "1250.5") {
		t.Error("missing QPS")
	}
	if !contains(md, "P50") {
		t.Error("missing P50")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
