package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSocketHub WebSocket 连接管理器
type WebSocketHub struct {
	clients    map[string]map[*websocket.Conn]bool
	broadcast  chan *wsMessage
	register   chan *wsClient
	unregister chan *wsClient
	mu         sync.RWMutex
	stopCh     chan struct{}
}

type wsMessage struct {
	taskID string
	data   interface{}
}

type wsClient struct {
	taskID string
	conn   *websocket.Conn
}

// NewWebSocketHub 创建 WebSocket Hub
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[string]map[*websocket.Conn]bool),
		broadcast:  make(chan *wsMessage, 256),
		register:   make(chan *wsClient),
		unregister: make(chan *wsClient),
		stopCh:     make(chan struct{}),
	}
}

// Run 运行 Hub
func (h *WebSocketHub) Run() {
	for {
		select {
		case <-h.stopCh:
			// 关闭所有连接
			h.mu.Lock()
			for _, clients := range h.clients {
				for conn := range clients {
					conn.Close()
				}
			}
			h.clients = make(map[string]map[*websocket.Conn]bool)
			h.mu.Unlock()
			log.Println("WebSocket Hub stopped")
			return

		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.taskID] == nil {
				h.clients[client.taskID] = make(map[*websocket.Conn]bool)
			}
			h.clients[client.taskID][client.conn] = true
			h.mu.Unlock()
			log.Printf("WebSocket client connected for task %s", client.taskID)

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.taskID]; ok {
				if _, ok := clients[client.conn]; ok {
					delete(clients, client.conn)
				}
			}
			h.mu.Unlock()
			client.conn.Close() // Always close connection to prevent resource leak

		case msg := <-h.broadcast:
			h.mu.RLock()
			clients := h.clients[msg.taskID]
			h.mu.RUnlock()

			data, _ := json.Marshal(msg.data)
			for conn := range clients {
				if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
					log.Printf("WebSocket write error: %v", err)
					h.unregister <- &wsClient{taskID: msg.taskID, conn: conn}
				}
			}
		}
	}
}

// Shutdown 关闭 WebSocket Hub
func (h *WebSocketHub) Shutdown() {
	close(h.stopCh)
}

// Broadcast 广播消息
func (h *WebSocketHub) Broadcast(taskID string, data interface{}) {
	h.broadcast <- &wsMessage{taskID: taskID, data: data}
}

// HandleWebSocket 处理 WebSocket 连接
func (h *WebSocketHub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &wsClient{taskID: taskID, conn: conn}
	h.register <- client

	// 发送连接成功消息
	conn.WriteJSON(map[string]interface{}{
		"type": "connected",
		"data": map[string]string{"taskId": taskID},
	})

	// 保持连接
	defer func() {
		h.unregister <- client
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
