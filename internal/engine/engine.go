package engine

import (
	"context"
	"time"
)

// Result 请求结果
type Result struct {
	Success    bool
	StatusCode int
	Latency    time.Duration
	Body       []byte
	Error      error
}

// Engine 压测引擎接口
type Engine interface {
	Execute(ctx context.Context, target, method string, headers map[string]string, body string) (*Result, error)
	Close() error
}
