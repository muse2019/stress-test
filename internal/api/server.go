package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"stress-test/internal/store"
)

// Server HTTP 服务器
type Server struct {
	handler    *Handler
	router     *mux.Router
	addr       string
	httpServer *http.Server
	wsHub      *WebSocketHub
}

// NewServer 创建服务器
func NewServer(addr string, taskStore store.TaskStore, reportStore store.ReportStore) *Server {
	wsHub := NewWebSocketHub()
	go wsHub.Run()

	handler := NewHandler(taskStore, reportStore, wsHub)
	router := mux.NewRouter()

	s := &Server{
		handler: handler,
		router:  router,
		addr:    addr,
		wsHub:   wsHub,
	}

	s.setupRoutes()
	return s
}

// corsMiddleware CORS 中间件
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 应用 CORS 中间件
	s.router.Use(corsMiddleware)

	// 健康检查
	s.router.HandleFunc("/health", s.handler.HealthCheck).Methods("GET")

	api := s.router.PathPrefix("/api").Subrouter()

	// 任务管理
	api.HandleFunc("/tasks", s.handler.ListTasks).Methods("GET")
	api.HandleFunc("/tasks", s.handler.CreateTask).Methods("POST")
	api.HandleFunc("/tasks", s.handler.DeleteAllTasks).Methods("DELETE")
	api.HandleFunc("/tasks/groups", s.handler.GetGroups).Methods("GET")
	api.HandleFunc("/tasks/{id}", s.handler.GetTask).Methods("GET")
	api.HandleFunc("/tasks/{id}", s.handler.UpdateTask).Methods("PUT")
	api.HandleFunc("/tasks/{id}", s.handler.DeleteTask).Methods("DELETE")
	api.HandleFunc("/tasks/{id}/duplicate", s.handler.DuplicateTask).Methods("POST")

	// 压测控制
	api.HandleFunc("/tasks/{id}/start", s.handler.StartTask).Methods("POST")
	api.HandleFunc("/tasks/{id}/stop", s.handler.StopTask).Methods("POST")
	api.HandleFunc("/tasks/{id}/stats", s.handler.GetTaskStats).Methods("GET")

	// 报告管理
	api.HandleFunc("/reports", s.handler.ListReports).Methods("GET")
	api.HandleFunc("/reports/{id}", s.handler.GetReport).Methods("GET")
	api.HandleFunc("/reports/{id}/download", s.handler.DownloadReport).Methods("GET")
	api.HandleFunc("/reports/{id}", s.handler.DeleteReport).Methods("DELETE")
	api.HandleFunc("/reports/compare/{id1}/{id2}", s.handler.CompareReports).Methods("GET")

	// WebSocket 实时推送
	api.HandleFunc("/ws/stats/{id}", s.handler.wsHub.HandleWebSocket).Methods("GET")
}

// Handler 获取 HTTP Handler
func (s *Server) Handler() http.Handler {
	return s.router
}

// Router 获取路由器
func (s *Server) Router() *mux.Router {
	return s.router
}

// Start 启动服务器
func (s *Server) Start() error {
	s.httpServer = &http.Server{
		Addr:         s.addr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Server starting on %s", s.addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown 优雅关闭服务器
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Server shutting down...")

	// 关闭 WebSocket Hub
	if s.wsHub != nil {
		s.wsHub.Shutdown()
	}

	// 关闭 HTTP 服务器
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}
