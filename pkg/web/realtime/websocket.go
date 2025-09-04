package realtime

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub      *Hub
	logger   logger.Logger
	upgrader websocket.Upgrader
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *Hub, logger logger.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		hub:    hub,
		logger: logger,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// In production, you should validate the origin properly
				return true
			},
		},
	}
}

// HandleWebSocket handles WebSocket upgrade and connection
func (wsh *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade connection to WebSocket
	conn, err := wsh.upgrader.Upgrade(w, r, nil)
	if err != nil {
		wsh.logger.Error("realtime.websocket_upgrade_failed", map[string]interface{}{
			"error": err.Error(),
			"remote_addr": r.RemoteAddr,
		})
		return
	}
	defer conn.Close()

	// Generate client ID
	clientID := fmt.Sprintf("ws_%d_%s", time.Now().Unix(), r.RemoteAddr)
	
	// Create client
	client := NewClient(clientID, WebSocketClient)
	
	// Register client with hub
	wsh.hub.RegisterClient(client)
	defer wsh.hub.UnregisterClient(client)
	
	// Configure connection
	conn.SetReadLimit(512)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		wsh.hub.UpdateClientPing(clientID)
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	// Start ping routine
	go wsh.startPingRoutine(conn, clientID)
	
	// Start message sender
	go wsh.startMessageSender(conn, client)
	
	// Start message reader (handles incoming messages and keep-alive)
	wsh.startMessageReader(conn, clientID)
}

// startMessageSender sends messages from client channel to WebSocket
func (wsh *WebSocketHandler) startMessageSender(conn *websocket.Conn, client *Client) {
	ticker := time.NewTicker(54 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case message, ok := <-client.Channel:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Channel closed
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			// Send message as JSON
			if err := conn.WriteJSON(message); err != nil {
				wsh.logger.Error("realtime.websocket_write_error", map[string]interface{}{
					"client_id": client.ID,
					"error": err.Error(),
				})
				return
			}
			
		case <-ticker.C:
			// Send ping
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				wsh.logger.Error("realtime.websocket_ping_error", map[string]interface{}{
					"client_id": client.ID,
					"error": err.Error(),
				})
				return
			}
		}
	}
}

// startMessageReader reads messages from WebSocket (mainly for keep-alive)
func (wsh *WebSocketHandler) startMessageReader(conn *websocket.Conn, clientID string) {
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				wsh.logger.Error("realtime.websocket_read_error", map[string]interface{}{
					"client_id": clientID,
					"error": err.Error(),
				})
			}
			break
		}
		
		// Handle different message types
		switch messageType {
		case websocket.TextMessage:
			wsh.handleTextMessage(clientID, message)
		case websocket.PongMessage:
			wsh.hub.UpdateClientPing(clientID)
		}
	}
}

// handleTextMessage handles incoming text messages from client
func (wsh *WebSocketHandler) handleTextMessage(clientID string, message []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		wsh.logger.Warn("realtime.invalid_client_message", map[string]interface{}{
			"client_id": clientID,
			"error": err.Error(),
		})
		return
	}
	
	// Handle client commands
	if cmd, ok := msg["command"].(string); ok {
		switch cmd {
		case "ping":
			wsh.hub.UpdateClientPing(clientID)
		case "subscribe":
			// Future: Handle subscription to specific message types
			wsh.logger.Debug("realtime.client_subscribe", map[string]interface{}{
				"client_id": clientID,
				"channels": msg["channels"],
			})
		default:
			wsh.logger.Debug("realtime.unknown_client_command", map[string]interface{}{
				"client_id": clientID,
				"command": cmd,
			})
		}
	}
}

// startPingRoutine sends periodic pings to keep connection alive
func (wsh *WebSocketHandler) startPingRoutine(conn *websocket.Conn, clientID string) {
	ticker := time.NewTicker(54 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			wsh.logger.Error("realtime.ping_routine_error", map[string]interface{}{
				"client_id": clientID,
				"error": err.Error(),
			})
			return
		}
	}
}