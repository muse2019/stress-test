package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"stress-test/internal/store"
	"stress-test/pkg/models"
)

func TestTaskAPI(t *testing.T) {
	// 创建临时目录
	taskDir, _ := os.MkdirTemp("", "tasks")
	reportDir, _ := os.MkdirTemp("", "reports")
	defer os.RemoveAll(taskDir)
	defer os.RemoveAll(reportDir)

	taskStore, _ := store.NewJSONTaskStore(taskDir)
	reportStore, _ := store.NewJSONReportStore(reportDir)
	wsHub := NewWebSocketHub()
	go wsHub.Run()
	handler := NewHandler(taskStore, reportStore, wsHub)

	// 测试创建任务
	task := map[string]interface{}{
		"name":        "Test Task",
		"target":      "https://example.com",
		"method":      "GET",
		"mode":        "fixed",
		"concurrency": 100,
		"duration":    60,
	}

	body, _ := json.Marshal(task)
	req := httptest.NewRequest("POST", "/api/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateTask(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("create task: expected 201, got %d, body: %s", w.Code, w.Body.String())
	}

	var created models.Task
	json.Unmarshal(w.Body.Bytes(), &created)

	if created.Name != "Test Task" {
		t.Errorf("name mismatch: got %s", created.Name)
	}

	// 测试获取任务列表
	req = httptest.NewRequest("GET", "/api/tasks", nil)
	w = httptest.NewRecorder()
	handler.ListTasks(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("list tasks: expected 200, got %d", w.Code)
	}

	// 测试获取单个任务
	req = httptest.NewRequest("GET", "/api/tasks/"+created.ID, nil)
	req = mux.SetURLVars(req, map[string]string{"id": created.ID})
	w = httptest.NewRecorder()
	handler.GetTask(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("get task: expected 200, got %d", w.Code)
	}

	// 测试更新任务
	updatedTask := map[string]interface{}{
		"name":        "Updated Task",
		"target":      "https://updated.example.com",
		"method":      "POST",
		"mode":        "fixed",
		"concurrency": 200,
		"duration":    120,
	}
	body, _ = json.Marshal(updatedTask)
	req = httptest.NewRequest("PUT", "/api/tasks/"+created.ID, bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": created.ID})
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.UpdateTask(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("update task: expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var updated models.Task
	json.Unmarshal(w.Body.Bytes(), &updated)

	if updated.Name != "Updated Task" {
		t.Errorf("updated name mismatch: got %s", updated.Name)
	}
	if updated.ID != created.ID {
		t.Errorf("updated task ID should remain the same: got %s, expected %s", updated.ID, created.ID)
	}

	// 测试删除任务
	req = httptest.NewRequest("DELETE", "/api/tasks/"+created.ID, nil)
	req = mux.SetURLVars(req, map[string]string{"id": created.ID})
	w = httptest.NewRecorder()
	handler.DeleteTask(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("delete task: expected 204, got %d", w.Code)
	}
}
