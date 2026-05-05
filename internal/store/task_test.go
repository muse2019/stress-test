package store

import (
	"os"
	"testing"

	"stress-test/pkg/models"
)

func TestJSONTaskStore(t *testing.T) {
	// 创建临时目录
	dir, err := os.MkdirTemp("", "tasks")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	store, err := NewJSONTaskStore(dir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// 测试保存
	task := &models.Task{
		ID:          "task-001",
		Name:        "Test Task",
		Target:      "https://example.com",
		Method:      "GET",
		Mode:        "fixed",
		Concurrency: 100,
		Duration:    60,
	}

	if err := store.Save(task); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	// 测试获取
	loaded, err := store.Get("task-001")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}

	if loaded.Name != task.Name {
		t.Errorf("name mismatch: got %s, want %s", loaded.Name, task.Name)
	}

	// 测试列表
	tasks, err := store.List()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}

	// 测试删除
	if err := store.Delete("task-001"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	_, err = store.Get("task-001")
	if err == nil {
		t.Error("expected error for deleted task")
	}
}

func TestJSONTaskStore_Validation(t *testing.T) {
	dir, err := os.MkdirTemp("", "tasks")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	store, err := NewJSONTaskStore(dir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	t.Run("Save nil task", func(t *testing.T) {
		err := store.Save(nil)
		if err != ErrNilTask {
			t.Errorf("expected ErrNilTask, got %v", err)
		}
	})

	t.Run("Save empty ID", func(t *testing.T) {
		task := &models.Task{ID: "", Name: "test"}
		err := store.Save(task)
		if err != ErrInvalidID {
			t.Errorf("expected ErrInvalidID, got %v", err)
		}
	})

	t.Run("Save path traversal ID", func(t *testing.T) {
		task := &models.Task{ID: "../escape", Name: "test"}
		err := store.Save(task)
		if err != ErrInvalidID {
			t.Errorf("expected ErrInvalidID, got %v", err)
		}
	})

	t.Run("Get empty ID", func(t *testing.T) {
		_, err := store.Get("")
		if err != ErrInvalidID {
			t.Errorf("expected ErrInvalidID, got %v", err)
		}
	})

	t.Run("Get path traversal ID", func(t *testing.T) {
		_, err := store.Get("../../../etc/passwd")
		if err != ErrInvalidID {
			t.Errorf("expected ErrInvalidID, got %v", err)
		}
	})

	t.Run("Delete empty ID", func(t *testing.T) {
		err := store.Delete("")
		if err != ErrInvalidID {
			t.Errorf("expected ErrInvalidID, got %v", err)
		}
	})

	t.Run("Delete non-existent task", func(t *testing.T) {
		err := store.Delete("non-existent")
		if err != ErrTaskNotFound {
			t.Errorf("expected ErrTaskNotFound, got %v", err)
		}
	})

	t.Run("Get non-existent task", func(t *testing.T) {
		_, err := store.Get("non-existent")
		if err != ErrTaskNotFound {
			t.Errorf("expected ErrTaskNotFound, got %v", err)
		}
	})

	t.Run("Valid IDs", func(t *testing.T) {
		validIDs := []string{
			"task-001",
			"task_002",
			"TASK123",
			"my-task-name",
			"my_task_name",
			"simple123",
		}
		for _, id := range validIDs {
			task := &models.Task{ID: id, Name: "test"}
			if err := store.Save(task); err != nil {
				t.Errorf("ID %q should be valid, got error: %v", id, err)
			}
		}
	})

	t.Run("Invalid IDs", func(t *testing.T) {
		invalidIDs := []string{
			"../escape",
			"../../../etc/passwd",
			"task with spaces",
			"task/with/slashes",
			"task\\with\\backslash",
			"task:with:colon",
			"task.with.dot",
			"",
		}
		for _, id := range invalidIDs {
			task := &models.Task{ID: id, Name: "test"}
			if err := store.Save(task); err != ErrInvalidID {
				t.Errorf("ID %q should be invalid, got error: %v", id, err)
			}
		}
	})
}
