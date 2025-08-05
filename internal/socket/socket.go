package socket

import (
	"bridge-serial/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Message represents the websocket message format matching the client
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

// Client represents a connected websocket client
type Client struct {
	id       string
	conn     *websocket.Conn
	send     chan Message
	server   *Server
	lastPong time.Time
	mu       sync.RWMutex
}

// Server represents the websocket server
type Server struct {
	clients    map[*Client]bool
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	upgrader   websocket.Upgrader
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewServer creates a new websocket server
func NewServer() *Server {
	ctx, cancel := context.WithCancel(context.Background())

	return &Server{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// Allow connections from any origin for development
				// In production, you should implement proper origin checking
				return true
			},
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start starts the websocket server
func (s *Server) Start() {
	go s.handleConnections()
	logger.Info("WebSocket server started")
}

// Stop stops the websocket server
func (s *Server) Stop() {
	s.cancel()
	logger.Info("WebSocket server stopped")
}

// handleConnections manages client connections and message broadcasting
func (s *Server) handleConnections() {
	for {
		select {
		case <-s.ctx.Done():
			return

		case client := <-s.register:
			s.mu.Lock()
			s.clients[client] = true
			s.mu.Unlock()
			logger.Info("Client %s connected. Total clients: %d", client.id, len(s.clients))

		case client := <-s.unregister:
			s.mu.Lock()
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.send)
				logger.Info("Client %s disconnected. Total clients: %d", client.id, len(s.clients))
			}
			s.mu.Unlock()

		case message := <-s.broadcast:
			s.mu.RLock()
			for client := range s.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(s.clients, client)
				}
			}
			s.mu.RUnlock()
		}
	}
}

// ServeWS handles websocket connections
func (s *Server) ServeWS(w http.ResponseWriter, r *http.Request) {
	// Get token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token parameter", http.StatusUnauthorized)
		return
	}

	// Validate token (implement your authentication logic here)
	if !s.validateToken(token) {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Failed to upgrade connection: %v", err)
		return
	}

	// Create new client
	client := &Client{
		id:       generateClientID(),
		conn:     conn,
		send:     make(chan Message, 256),
		server:   s,
		lastPong: time.Now(),
	}

	// Register client
	s.register <- client

	// Start client goroutines
	go client.writePump()
	go client.readPump()
}

// validateToken validates the authentication token
// Implement your authentication logic here
func (s *Server) validateToken(token string) bool {
	// For demo purposes, accept any non-empty token
	// In production, implement proper token validation
	return len(token) > 0
}

// BroadcastMessage broadcasts a message to all connected clients
func (s *Server) BroadcastMessage(msgType string, payload interface{}) {
	message := Message{
		Type:    msgType,
		Payload: payload,
	}

	select {
	case s.broadcast <- message:
	default:
		logger.Error("Broadcast channel full, message dropped")
	}
}

// GetConnectedClientsCount returns the number of connected clients
func (s *Server) GetConnectedClientsCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}

// readPump handles reading messages from the websocket connection
func (c *Client) readPump() {
	defer func() {
		c.server.unregister <- c
		c.conn.Close()
	}()

	// Set connection limits
	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.mu.Lock()
		c.lastPong = time.Now()
		c.mu.Unlock()
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, messageData, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("WebSocket error for client %s: %v", c.id, err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(messageData, &msg); err != nil {
			logger.Error("Failed to unmarshal message from client %s: %v", c.id, err)
			continue
		}

		logger.Info("Received message from client %s: type=%s, payload=%v", c.id, msg.Type, msg.Payload)
		c.handleMessage(msg)
	}
}

// writePump handles writing messages to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			jsonData, err := json.Marshal(message)
			if err != nil {
				logger.Error("Failed to marshal message for client %s: %v", c.id, err)
				continue
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
				logger.Error("Failed to write message to client %s: %v", c.id, err)
				return
			}

			logger.Info("Sent message to client %s: type=%s, payload=%v", c.id, message.Type, message.Payload)

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming messages from clients
func (c *Client) handleMessage(msg Message) {
	switch msg.Type {
	case "ping":
		// Respond with pong
		response := Message{
			Type:    "pong",
			Payload: msg.Payload,
		}
		select {
		case c.send <- response:
			logger.Info("Sent pong response to client %s", c.id)
		default:
			logger.Error("Failed to send pong response to client %s", c.id)
		}

	case "sync-to-self":
		// Handle sync-to-self messages
		logger.Info("Received sync-to-self from client %s with payload: %v", c.id, msg.Payload)

		// Auto-respond with sync-from-self message
		response := Message{
			Type:    "sync-from-self",
			Payload: "pong",
		}
		select {
		case c.send <- response:
			logger.Info("Sent sync-from-self response to client %s", c.id)
		default:
			logger.Error("Failed to send sync-from-self response to client %s", c.id)
		}

	case "sync-from-self":
		// Handle sync-from-self messages (informational)
		logger.Info("Received sync-from-self from client %s with payload: %v", c.id, msg.Payload)

	default:
		logger.Info("Received unknown message type '%s' from client %s with payload: %v", msg.Type, c.id, msg.Payload)
	}
}

// SendMessage sends a message to this specific client
func (c *Client) SendMessage(msgType string, payload interface{}) {
	message := Message{
		Type:    msgType,
		Payload: payload,
	}

	select {
	case c.send <- message:
	default:
		logger.Error("Client %s send channel full, message dropped", c.id)
	}
}

// generateClientID generates a unique client ID
func generateClientID() string {
	return fmt.Sprintf("client_%d", time.Now().UnixNano())
}
