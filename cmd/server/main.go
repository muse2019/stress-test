package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"stress-test/internal/api"
	"stress-test/internal/store"
)

func main() {
	log.Println("Stress Test Tool starting...")

	// 创建数据目录
	dataDir := getEnv("DATA_DIR", "./data")
	if err := os.MkdirAll(dataDir+"/tasks", 0755); err != nil {
		log.Fatal("Failed to create tasks directory:", err)
	}
	if err := os.MkdirAll(dataDir+"/reports", 0755); err != nil {
		log.Fatal("Failed to create reports directory:", err)
	}

	// 初始化存储
	taskStore, err := store.NewJSONTaskStore(dataDir + "/tasks")
	if err != nil {
		log.Fatal("Failed to create task store:", err)
	}
	reportStore, err := store.NewJSONReportStore(dataDir + "/reports")
	if err != nil {
		log.Fatal("Failed to create report store:", err)
	}

	// 启动服务器
	addr := getEnv("SERVER_ADDR", ":8080")
	server := api.NewServer(addr, taskStore, reportStore)

	// 设置信号处理
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// 在 goroutine 中启动服务器
	go func() {
		if err := server.Start(); err != nil {
			log.Printf("Server stopped: %v", err)
		}
	}()

	// 等待中断信号
	<-stop

	// 创建关闭超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 优雅关闭
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped gracefully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
