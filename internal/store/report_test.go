package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"stress-test/pkg/models"
)

func TestJSONReportStore_SaveAndGet(t *testing.T) {
	// Create temp directory
	dir, err := os.MkdirTemp("", "reports")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	store, err := NewJSONReportStore(dir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Create test report
	report := &models.Report{
		ID:        "report-001",
		TaskID:    "task-001",
		TaskName:  "Test Task",
		StartTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 1, 10, 1, 0, 0, time.UTC),
		Duration:  60,
		Config: models.Task{
			ID:          "task-001",
			Name:        "Test Task",
			Target:      "https://example.com",
			Method:      "GET",
			Concurrency: 100,
			Duration:    60,
		},
		FinalStats: models.RealtimeStats{
			TotalRequests: 1000,
			SuccessCount:  990,
			FailedCount:   10,
			QPS:           16.67,
			AvgRT:         50,
			MinRT:         10,
			MaxRT:         200,
			P50:           45,
			P90:           80,
			P95:           100,
			P99:           150,
		},
	}

	// Test Save
	if err := store.Save(report); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	// Verify JSON file exists
	jsonPath := filepath.Join(dir, "report-001.json")
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		t.Error("JSON file was not created")
	}

	// Verify MD file exists
	mdPath := filepath.Join(dir, "report-001.md")
	if _, err := os.Stat(mdPath); os.IsNotExist(err) {
		t.Error("Markdown file was not created")
	}

	// Test Get
	loaded, err := store.Get("report-001")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}

	if loaded.ID != report.ID {
		t.Errorf("ID mismatch: got %s, want %s", loaded.ID, report.ID)
	}
	if loaded.TaskName != report.TaskName {
		t.Errorf("TaskName mismatch: got %s, want %s", loaded.TaskName, report.TaskName)
	}
	if loaded.FinalStats.TotalRequests != report.FinalStats.TotalRequests {
		t.Errorf("TotalRequests mismatch: got %d, want %d", loaded.FinalStats.TotalRequests, report.FinalStats.TotalRequests)
	}
}

func TestJSONReportStore_List(t *testing.T) {
	// Create temp directory
	dir, err := os.MkdirTemp("", "reports")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	store, err := NewJSONReportStore(dir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Create multiple reports with different times
	reports := []*models.Report{
		{
			ID:        "report-old",
			TaskID:    "task-001",
			TaskName:  "Old Report",
			StartTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2024, 1, 1, 10, 1, 0, 0, time.UTC),
			Duration:  60,
		},
		{
			ID:        "report-new",
			TaskID:    "task-002",
			TaskName:  "New Report",
			StartTime: time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2024, 1, 2, 10, 1, 0, 0, time.UTC),
			Duration:  60,
		},
		{
			ID:        "report-middle",
			TaskID:    "task-003",
			TaskName:  "Middle Report",
			StartTime: time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2024, 1, 1, 15, 1, 0, 0, time.UTC),
			Duration:  60,
		},
	}

	for _, r := range reports {
		if err := store.Save(r); err != nil {
			t.Fatalf("save failed: %v", err)
		}
	}

	// Test List - should be sorted by StartTime descending (newest first)
	list, err := store.List()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	if len(list) != 3 {
		t.Fatalf("expected 3 reports, got %d", len(list))
	}

	// Verify order: newest first
	if list[0].ID != "report-new" {
		t.Errorf("first report should be 'report-new', got %s", list[0].ID)
	}
	if list[1].ID != "report-middle" {
		t.Errorf("second report should be 'report-middle', got %s", list[1].ID)
	}
	if list[2].ID != "report-old" {
		t.Errorf("third report should be 'report-old', got %s", list[2].ID)
	}
}

func TestJSONReportStore_Delete(t *testing.T) {
	// Create temp directory
	dir, err := os.MkdirTemp("", "reports")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	store, err := NewJSONReportStore(dir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Create and save a report
	report := &models.Report{
		ID:        "report-to-delete",
		TaskID:    "task-001",
		TaskName:  "Test Task",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Minute),
		Duration:  60,
	}

	if err := store.Save(report); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	// Verify files exist
	jsonPath := filepath.Join(dir, "report-to-delete.json")
	mdPath := filepath.Join(dir, "report-to-delete.md")
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		t.Fatal("JSON file should exist before delete")
	}
	if _, err := os.Stat(mdPath); os.IsNotExist(err) {
		t.Fatal("MD file should exist before delete")
	}

	// Delete the report
	if err := store.Delete("report-to-delete"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	// Verify files are deleted
	if _, err := os.Stat(jsonPath); !os.IsNotExist(err) {
		t.Error("JSON file should be deleted")
	}
	if _, err := os.Stat(mdPath); !os.IsNotExist(err) {
		t.Error("MD file should be deleted")
	}

	// Verify Get returns error
	_, err = store.Get("report-to-delete")
	if err == nil {
		t.Error("expected error for deleted report")
	}
}

func TestJSONReportStore_GetMarkdown(t *testing.T) {
	// Create temp directory
	dir, err := os.MkdirTemp("", "reports")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	store, err := NewJSONReportStore(dir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	report := &models.Report{
		ID:        "report-001",
		TaskID:    "task-001",
		TaskName:  "Test Task",
		StartTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 1, 1, 10, 1, 0, 0, time.UTC),
		Duration:  60,
		Config: models.Task{
			Target:      "https://example.com",
			Method:      "GET",
			Concurrency: 100,
		},
		FinalStats: models.RealtimeStats{
			TotalRequests: 1000,
			SuccessCount:  990,
			FailedCount:   10,
			QPS:           16.67,
			AvgRT:         50,
		},
	}

	t.Run("GetMarkdown from file", func(t *testing.T) {
		if err := store.Save(report); err != nil {
			t.Fatalf("save failed: %v", err)
		}

		md, err := store.GetMarkdown("report-001")
		if err != nil {
			t.Fatalf("GetMarkdown failed: %v", err)
		}

		if !strings.Contains(md, "Test Task") {
			t.Error("markdown should contain task name")
		}
		if !strings.Contains(md, "1000") {
			t.Error("markdown should contain total requests")
		}
	})

	t.Run("GetMarkdown fallback when MD file missing", func(t *testing.T) {
		// Remove the MD file to test fallback
		mdPath := filepath.Join(dir, "report-001.md")
		if err := os.Remove(mdPath); err != nil {
			t.Fatalf("failed to remove MD file: %v", err)
		}

		// GetMarkdown should still work by regenerating from JSON
		md, err := store.GetMarkdown("report-001")
		if err != nil {
			t.Fatalf("GetMarkdown fallback failed: %v", err)
		}

		if !strings.Contains(md, "Test Task") {
			t.Error("markdown should contain task name")
		}
		if !strings.Contains(md, "1000") {
			t.Error("markdown should contain total requests")
		}
	})

	t.Run("GetMarkdown for non-existent report", func(t *testing.T) {
		_, err := store.GetMarkdown("non-existent")
		if err == nil {
			t.Error("expected error for non-existent report")
		}
	})
}

func TestJSONReportStore_Validation(t *testing.T) {
	// Create temp directory
	dir, err := os.MkdirTemp("", "reports")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	store, err := NewJSONReportStore(dir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	t.Run("Save nil report", func(t *testing.T) {
		err := store.Save(nil)
		if err == nil {
			t.Error("expected error for nil report")
		}
		if !strings.Contains(err.Error(), "nil") {
			t.Errorf("expected nil error, got: %v", err)
		}
	})

	t.Run("Save path traversal ID", func(t *testing.T) {
		report := &models.Report{
			ID:        "../escape",
			TaskID:    "task-001",
			TaskName:  "Test",
			StartTime: time.Now(),
		}
		err := store.Save(report)
		if err != ErrInvalidID {
			t.Errorf("expected ErrInvalidID, got: %v", err)
		}
	})

	t.Run("Get empty ID", func(t *testing.T) {
		_, err := store.Get("")
		if err != ErrInvalidID {
			t.Errorf("expected ErrInvalidID, got: %v", err)
		}
	})

	t.Run("Get path traversal ID", func(t *testing.T) {
		_, err := store.Get("../../../etc/passwd")
		if err != ErrInvalidID {
			t.Errorf("expected ErrInvalidID, got: %v", err)
		}
	})

	t.Run("Delete empty ID", func(t *testing.T) {
		err := store.Delete("")
		if err != ErrInvalidID {
			t.Errorf("expected ErrInvalidID, got: %v", err)
		}
	})

	t.Run("Delete path traversal ID", func(t *testing.T) {
		err := store.Delete("../escape")
		if err != ErrInvalidID {
			t.Errorf("expected ErrInvalidID, got: %v", err)
		}
	})

	t.Run("GetMarkdown empty ID", func(t *testing.T) {
		_, err := store.GetMarkdown("")
		if err != ErrInvalidID {
			t.Errorf("expected ErrInvalidID, got: %v", err)
		}
	})

	t.Run("GetMarkdown path traversal ID", func(t *testing.T) {
		_, err := store.GetMarkdown("../../etc/passwd")
		if err != ErrInvalidID {
			t.Errorf("expected ErrInvalidID, got: %v", err)
		}
	})

	t.Run("Get non-existent report", func(t *testing.T) {
		_, err := store.Get("non-existent")
		if err == nil {
			t.Error("expected error for non-existent report")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("expected not found error, got: %v", err)
		}
	})

	t.Run("Valid IDs", func(t *testing.T) {
		validIDs := []string{
			"report-001",
			"report_002",
			"REPORT123",
			"my-report-name",
			"my_report_name",
			"simple123",
		}
		for _, id := range validIDs {
			report := &models.Report{
				ID:        id,
				TaskID:    "task-001",
				TaskName:  "Test",
				StartTime: time.Now(),
			}
			if err := store.Save(report); err != nil {
				t.Errorf("ID %q should be valid, got error: %v", id, err)
			}
		}
	})

	t.Run("Invalid IDs", func(t *testing.T) {
		invalidIDs := []string{
			"../escape",
			"../../../etc/passwd",
			"report with spaces",
			"report/with/slashes",
			"report\\with\\backslash",
			"report:with:colon",
			"report.with.dot",
		}
		for _, id := range invalidIDs {
			report := &models.Report{
				ID:        id,
				TaskID:    "task-001",
				TaskName:  "Test",
				StartTime: time.Now(),
			}
			if err := store.Save(report); err != ErrInvalidID {
				t.Errorf("ID %q should be invalid, got error: %v", id, err)
			}
		}
	})

	t.Run("Empty ID auto-generates", func(t *testing.T) {
		report := &models.Report{
			ID:        "",
			TaskID:    "task-001",
			TaskName:  "Test",
			StartTime: time.Now(),
		}
		// Empty ID should be auto-generated, not rejected
		if err := store.Save(report); err != nil {
			t.Errorf("empty ID should be auto-generated, got error: %v", err)
		}
		if report.ID == "" {
			t.Error("ID should have been auto-generated")
		}
	})
}

func TestJSONReportStore_AutoGenerateID(t *testing.T) {
	// Create temp directory
	dir, err := os.MkdirTemp("", "reports")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	store, err := NewJSONReportStore(dir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Save report without ID
	report := &models.Report{
		TaskID:    "task-001",
		TaskName:  "Auto ID Test",
		StartTime: time.Now(),
	}

	if err := store.Save(report); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	// ID should have been auto-generated
	if report.ID == "" {
		t.Error("ID should have been auto-generated")
	}
	if !strings.HasPrefix(report.ID, "report-") {
		t.Errorf("auto-generated ID should start with 'report-', got %s", report.ID)
	}

	// Verify we can retrieve it
	loaded, err := store.Get(report.ID)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if loaded.TaskName != "Auto ID Test" {
		t.Errorf("TaskName mismatch: got %s, want Auto ID Test", loaded.TaskName)
	}
}

func TestJSONReportStore_EmptyList(t *testing.T) {
	// Create temp directory
	dir, err := os.MkdirTemp("", "reports")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	store, err := NewJSONReportStore(dir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// List should return nil or empty slice for empty directory
	list, err := store.List()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d items", len(list))
	}
}
