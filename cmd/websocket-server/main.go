package main

import (
	"bridge-serial/internal/socket"
	"bridge-serial/pkg/logger"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Create WebSocket server
	wsServer := socket.NewServer()

	// Start the WebSocket server
	wsServer.Start()

	// Create HTTP server
	mux := http.NewServeMux()

	// WebSocket endpoint
	mux.HandleFunc("/ws", wsServer.ServeWS)

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		clientCount := wsServer.GetConnectedClientsCount()
		fmt.Fprintf(w, `{"status":"ok","connected_clients":%d}`, clientCount)
	})

	// Create HTTP server
	server := &http.Server{
		Addr:    ":8001",
		Handler: mux,
	}

	// Start HTTP server in goroutine
	go func() {
		logger.Info("Starting HTTP server on :8001")
		logger.Info("WebSocket endpoint: ws://localhost:8001/ws")
		logger.Info("Test page: http://localhost:8001/")
		logger.Info("Health check: http://localhost:8001/health")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Stop WebSocket server
	wsServer.Stop()

	// Shutdown HTTP server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown: %v", err)
	}

	logger.Info("Server stopped")
}
