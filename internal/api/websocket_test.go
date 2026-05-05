package api

import (
	"testing"
)

func TestWebSocketHubBroadcast(t *testing.T) {
	hub := NewWebSocketHub()
	go hub.Run()

	// 测试消息广播
	msg := map[string]interface{}{
		"type": "stats",
		"data": map[string]interface{}{
			"qps": 100.5,
		},
	}

	hub.Broadcast("task-001", msg)

	// 由于没有实际连接，这个测试主要验证不会 panic
}
