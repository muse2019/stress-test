package models

import (
	"errors"
	"fmt"
	"net/url"
	"time"
)

// Task 压测任务配置
type Task struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Protocol    string            `json:"protocol"`
	Target      string            `json:"target"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers,omitempty"`
	Body        string            `json:"body,omitempty"`
	Timeout     int               `json:"timeout"`     // 毫秒
	Mode        string            `json:"mode"`        // fixed/staircase/rate
	Concurrency int               `json:"concurrency"`
	Duration    int               `json:"duration"`    // 秒
	Rate        int               `json:"rate,omitempty"`
	Staircase   *Staircase        `json:"staircase,omitempty"`
	Variables   []Variable        `json:"variables,omitempty"`
	Assertions  []Assertion       `json:"assertions,omitempty"`
	PreScript   string            `json:"preScript,omitempty"`
	PostScript  string            `json:"postScript,omitempty"`
}

// Staircase 阶梯压测配置
type Staircase struct {
	Start    int `json:"start"`
	Step     int `json:"step"`
	StepTime int `json:"stepTime"`
	Max      int `json:"max"`
}

// Variable 变量定义
type Variable struct {
	Name   string `json:"name"`
	Type   string `json:"type"`   // static/random/csv
	Value  string `json:"value"`
	Min    int    `json:"min,omitempty"`
	Max    int    `json:"max,omitempty"`
	File   string `json:"file,omitempty"`
	Column string `json:"column,omitempty"`
}

// Assertion 断言规则
type Assertion struct {
	Type     string      `json:"type"`
	Operator string      `json:"operator"`
	Expected interface{} `json:"expected"`
}

// Validate 验证任务配置
func (t *Task) Validate() error {
	if t.Name == "" {
		return errors.New("name is required")
	}
	if t.Target == "" {
		return errors.New("target is required")
	}
	if _, err := url.Parse(t.Target); err != nil {
		return fmt.Errorf("invalid target URL: %w", err)
	}
	if t.Method == "" {
		t.Method = "GET"
	}
	if t.Mode == "" {
		t.Mode = "fixed"
	}
	if t.Mode == "fixed" && t.Concurrency <= 0 {
		return errors.New("concurrency must be greater than 0 for fixed mode")
	}
	if t.Duration <= 0 {
		return errors.New("duration must be greater than 0")
	}
	if t.Timeout <= 0 {
		t.Timeout = 30000
	}
	if t.Protocol == "" {
		t.Protocol = "http"
	}
	return nil
}

// GenerateID 生成任务 ID
func (t *Task) GenerateID() {
	if t.ID == "" {
		t.ID = fmt.Sprintf("task-%d", time.Now().UnixNano())
	}
}

// RealtimeStats 实时统计
type RealtimeStats struct {
	Timestamp     int64            `json:"timestamp"`
	TotalRequests int64            `json:"totalRequests"`
	SuccessCount  int64            `json:"successCount"`
	FailedCount   int64            `json:"failedCount"`
	QPS           float64          `json:"qps"`
	AvgRT         int64            `json:"avgRT"`
	MinRT         int64            `json:"minRT"`
	MaxRT         int64            `json:"maxRT"`
	P50           int64            `json:"p50"`
	P90           int64            `json:"p90"`
	P95           int64            `json:"p95"`
	P99           int64            `json:"p99"`
	Errors        map[string]int64 `json:"errors"`
}

// SuccessRate 计算成功率
func (s *RealtimeStats) SuccessRate() float64 {
	if s.TotalRequests == 0 {
		return 0
	}
	return float64(s.SuccessCount) / float64(s.TotalRequests) * 100
}

// Report 最终报告
type Report struct {
	ID         string          `json:"id"`
	TaskID     string          `json:"taskId"`
	TaskName   string          `json:"taskName"`
	StartTime  time.Time       `json:"startTime"`
	EndTime    time.Time       `json:"endTime"`
	Duration   int             `json:"duration"`
	Config     Task            `json:"config"`
	FinalStats RealtimeStats   `json:"finalStats"`
	Timeline   []RealtimeStats `json:"timeline"`
}

// GenerateID 生成报告 ID
func (r *Report) GenerateID() {
	if r.ID == "" {
		r.ID = fmt.Sprintf("report-%d", time.Now().UnixNano())
	}
}
