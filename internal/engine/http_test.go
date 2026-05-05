package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPEngineExecute(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("X-Test") != "value" {
			t.Errorf("expected X-Test header to be 'value'")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	engine := NewHTTPEngine(5 * time.Second)
	defer engine.Close()

	result, err := engine.Execute(context.Background(), server.URL, "GET", map[string]string{"X-Test": "value"}, "")
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}
	if result.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", result.StatusCode)
	}
	if result.Latency == 0 {
		t.Error("latency should be recorded")
	}
	if string(result.Body) != `{"status":"ok"}` {
		t.Errorf("unexpected body: %s", result.Body)
	}
}

func TestHTTPEngineTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	engine := NewHTTPEngine(10 * time.Millisecond)
	defer engine.Close()

	result, err := engine.Execute(context.Background(), server.URL, "GET", nil, "")
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if result.Success {
		t.Error("expected failure due to timeout")
	}
}

func TestHTTPEngineErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	engine := NewHTTPEngine(5 * time.Second)
	defer engine.Close()

	result, _ := engine.Execute(context.Background(), server.URL, "GET", nil, "")

	if result.Success {
		t.Error("500 status should not be success")
	}
	if result.StatusCode != 500 {
		t.Errorf("expected 500, got %d", result.StatusCode)
	}
}
