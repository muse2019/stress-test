package engine

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"
)

// HTTPEngine HTTP 压测引擎
type HTTPEngine struct {
	client *http.Client
}

// NewHTTPEngine 创建 HTTP 引擎
func NewHTTPEngine(timeout time.Duration) *HTTPEngine {
	return &HTTPEngine{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        1000,
				MaxIdleConnsPerHost: 1000,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// Execute 执行 HTTP 请求
func (e *HTTPEngine) Execute(ctx context.Context, target, method string, headers map[string]string, body string) (*Result, error) {
	req, err := http.NewRequestWithContext(ctx, method, target, strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	start := time.Now()
	resp, err := e.client.Do(req)
	latency := time.Since(start)

	if err != nil {
		return &Result{
			Success: false,
			Latency: latency,
			Error:   err,
		}, nil
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	return &Result{
		Success:    resp.StatusCode < 400,
		StatusCode: resp.StatusCode,
		Latency:    latency,
		Body:       respBody,
	}, nil
}

// Close 关闭引擎
func (e *HTTPEngine) Close() error {
	e.client.CloseIdleConnections()
	return nil
}
