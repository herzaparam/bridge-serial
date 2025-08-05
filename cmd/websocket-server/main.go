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

	// Simple web page for testing
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>WebSocket Test</title>
</head>
<body>
    <h1>WebSocket Server Test</h1>
    <div id="status">Disconnected</div>
    <div>
        <input type="text" id="token" placeholder="Enter token" value="test-token">
        <button onclick="connect()">Connect</button>
        <button onclick="disconnect()">Disconnect</button>
    </div>
    <div>
        <input type="text" id="message" placeholder="Enter message">
        <button onclick="sendMessage()">Send Message</button>
        <button onclick="sendPing()">Send Ping</button>
        <button onclick="sendSyncToSelf()">Send Sync-To-Self</button>
    </div>
    <div id="messages" style="border: 1px solid #ccc; height: 300px; overflow-y: scroll; padding: 10px; margin-top: 10px;"></div>

    <script>
        let ws = null;
        const status = document.getElementById('status');
        const messages = document.getElementById('messages');

        function updateStatus(text, color = 'black') {
            status.textContent = text;
            status.style.color = color;
        }

        function addMessage(msg) {
            const div = document.createElement('div');
            div.textContent = new Date().toLocaleTimeString() + ': ' + msg;
            messages.appendChild(div);
            messages.scrollTop = messages.scrollHeight;
        }

        function connect() {
            const token = document.getElementById('token').value;
            if (!token) {
                alert('Please enter a token');
                return;
            }

            const wsUrl = 'ws://localhost:8001/ws?token=' + encodeURIComponent(token);
            ws = new WebSocket(wsUrl);

            ws.onopen = function() {
                updateStatus('Connected', 'green');
                addMessage('Connected to server');
            };

            ws.onmessage = function(event) {
                addMessage('Received: ' + event.data);
            };

            ws.onclose = function() {
                updateStatus('Disconnected', 'red');
                addMessage('Disconnected from server');
            };

            ws.onerror = function(error) {
                updateStatus('Error', 'red');
                addMessage('Error: ' + error);
            };
        }

        function disconnect() {
            if (ws) {
                ws.close();
                ws = null;
            }
        }

        function sendMessage() {
            const messageText = document.getElementById('message').value;
            if (!ws || ws.readyState !== WebSocket.OPEN) {
                alert('Not connected');
                return;
            }
            if (!messageText) {
                alert('Please enter a message');
                return;
            }

            const message = {
                type: 'custom',
                payload: messageText
            };
            ws.send(JSON.stringify(message));
            addMessage('Sent: ' + JSON.stringify(message));
            document.getElementById('message').value = '';
        }

        function sendPing() {
            if (!ws || ws.readyState !== WebSocket.OPEN) {
                alert('Not connected');
                return;
            }

            const message = {
                type: 'ping',
                payload: null
            };
            ws.send(JSON.stringify(message));
            addMessage('Sent: ' + JSON.stringify(message));
        }

        function sendSyncToSelf() {
            if (!ws || ws.readyState !== WebSocket.OPEN) {
                alert('Not connected');
                return;
            }

            const message = {
                type: 'sync-to-self',
                payload: 'test data'
            };
            ws.send(JSON.stringify(message));
            addMessage('Sent: ' + JSON.stringify(message));
        }
    </script>
</body>
</html>
		`))
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
