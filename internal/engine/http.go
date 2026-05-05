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
	client     *http.Client
	retryCount int
	retryDelay time.Duration
}

// HTTPEngineOption 引擎配置选项
type HTTPEngineOption func(*HTTPEngine)

// WithRetry 设置重试配置
func WithRetry(count int, delay time.Duration) HTTPEngineOption {
	return func(e *HTTPEngine) {
		e.retryCount = count
		e.retryDelay = delay
	}
}

// NewHTTPEngine 创建 HTTP 引擎
func NewHTTPEngine(timeout time.Duration, opts ...HTTPEngineOption) *HTTPEngine {
	e := &HTTPEngine{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        1000,
				MaxIdleConnsPerHost: 1000,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// Execute 执行 HTTP 请求
func (e *HTTPEngine) Execute(ctx context.Context, target, method string, headers map[string]string, body string) (*Result, error) {
	var result *Result
	var err error

	for attempt := 0; attempt <= e.retryCount; attempt++ {
		result, err = e.doRequest(ctx, target, method, headers, body)
		if err != nil || result.Success {
			return result, err
		}

		// 如果还有重试机会，等待后重试
		if attempt < e.retryCount && e.retryDelay > 0 {
			select {
			case <-time.After(e.retryDelay):
			case <-ctx.Done():
				return result, ctx.Err()
			}
		}
	}

	return result, err
}

// doRequest 执行单次请求
func (e *HTTPEngine) doRequest(ctx context.Context, target, method string, headers map[string]string, body string) (*Result, error) {
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
