package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"stress-test/internal/api"
	"stress-test/internal/store"
)

func TestFullWorkflow(t *testing.T) {
	// Create temporary storage
	taskDir, err := os.MkdirTemp("", "tasks")
	if err != nil {
		t.Fatal(err)
	}
	reportDir, err := os.MkdirTemp("", "reports")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(taskDir)
	defer os.RemoveAll(reportDir)

	taskStore, err := store.NewJSONTaskStore(taskDir)
	if err != nil {
		t.Fatal(err)
	}
	reportStore, err := store.NewJSONReportStore(reportDir)
	if err != nil {
		t.Fatal(err)
	}

	// Create test server using the actual router
	server := api.NewServer(":0", taskStore, reportStore)
	defer server.Shutdown(nil)

	// Use httptest with the server's handler
	router := httptest.NewServer(server.Handler())
	defer router.Close()

	// 1. Create a task
	taskReq := map[string]interface{}{
		"name":        "Integration Test",
		"target":      "https://httpbin.org/get",
		"method":      "GET",
		"mode":        "fixed",
		"concurrency": 10,
		"duration":    3,
	}

	taskReqBody, err := json.Marshal(taskReq)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Post(router.URL+"/api/tasks", "application/json", bytes.NewReader(taskReqBody))
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	var task map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&task)
	resp.Body.Close()

	taskID, ok := task["id"].(string)
	if !ok || taskID == "" {
		t.Fatal("task ID should not be empty")
	}

	t.Logf("Created task: %s", taskID)

	// 2. Start the stress test
	resp, err = http.Post(router.URL+"/api/tasks/"+taskID+"/start", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d for start, got %d", http.StatusOK, resp.StatusCode)
	}

	var startResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&startResp)
	resp.Body.Close()

	t.Logf("Started task: %v", startResp)

	// 3. Wait for completion (duration is 3 seconds, wait a bit longer)
	time.Sleep(5 * time.Second)

	// 4. Stop the task (if still running)
	resp, err = http.Post(router.URL+"/api/tasks/"+taskID+"/stop", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	// 5. Verify task exists
	resp, err = http.Get(router.URL + "/api/tasks/" + taskID)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d for get task, got %d", http.StatusOK, resp.StatusCode)
	}
	resp.Body.Close()

	// 6. List tasks
	resp, err = http.Get(router.URL + "/api/tasks")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d for list tasks, got %d", http.StatusOK, resp.StatusCode)
	}

	var listResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&listResp)
	resp.Body.Close()

	tasks, ok := listResp["tasks"].([]interface{})
	if !ok {
		t.Fatal("expected tasks array in response")
	}
	if len(tasks) == 0 {
		t.Fatal("expected at least one task in list")
	}

	t.Logf("Found %d tasks", len(tasks))

	// 7. List reports
	resp, err = http.Get(router.URL + "/api/reports")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d for list reports, got %d", http.StatusOK, resp.StatusCode)
	}

	var reportsResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&reportsResp)
	resp.Body.Close()

	reports, ok := reportsResp["reports"].([]interface{})
	if !ok {
		t.Fatal("expected reports array in response")
	}

	t.Logf("Found %d reports", len(reports))

	// 8. Delete the task
	req, err := http.NewRequest("DELETE", router.URL+"/api/tasks/"+taskID, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	t.Log("Integration test completed successfully")
}

func TestCreateTaskValidation(t *testing.T) {
	// Create temporary storage
	taskDir, err := os.MkdirTemp("", "tasks")
	if err != nil {
		t.Fatal(err)
	}
	reportDir, err := os.MkdirTemp("", "reports")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(taskDir)
	defer os.RemoveAll(reportDir)

	taskStore, err := store.NewJSONTaskStore(taskDir)
	if err != nil {
		t.Fatal(err)
	}
	reportStore, err := store.NewJSONReportStore(reportDir)
	if err != nil {
		t.Fatal(err)
	}

	// Create test server
	server := api.NewServer(":0", taskStore, reportStore)
	defer server.Shutdown(nil)

	router := httptest.NewServer(server.Handler())
	defer router.Close()

	// Test missing required fields
	taskReq := map[string]interface{}{
		"name": "", // empty name should fail
	}

	taskReqBody, err := json.Marshal(taskReq)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Post(router.URL+"/api/tasks", "application/json", bytes.NewReader(taskReqBody))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d for invalid task, got %d", http.StatusBadRequest, resp.StatusCode)
	}

	t.Log("Validation test passed")
}

func TestTaskNotFound(t *testing.T) {
	// Create temporary storage
	taskDir, err := os.MkdirTemp("", "tasks")
	if err != nil {
		t.Fatal(err)
	}
	reportDir, err := os.MkdirTemp("", "reports")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(taskDir)
	defer os.RemoveAll(reportDir)

	taskStore, err := store.NewJSONTaskStore(taskDir)
	if err != nil {
		t.Fatal(err)
	}
	reportStore, err := store.NewJSONReportStore(reportDir)
	if err != nil {
		t.Fatal(err)
	}

	// Create test server
	server := api.NewServer(":0", taskStore, reportStore)
	defer server.Shutdown(nil)

	router := httptest.NewServer(server.Handler())
	defer router.Close()

	// Test getting non-existent task
	resp, err := http.Get(router.URL + "/api/tasks/nonexistent-id")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status %d for non-existent task, got %d", http.StatusNotFound, resp.StatusCode)
	}

	// Test starting non-existent task
	resp, err = http.Post(router.URL+"/api/tasks/nonexistent-id/start", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status %d for starting non-existent task, got %d", http.StatusNotFound, resp.StatusCode)
	}

	t.Log("Task not found test passed")
}
