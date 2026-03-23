package realtime

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

const (
	DefaultReadDeadline   = 60 * time.Second
	DefaultWriteDeadline  = 10 * time.Second
	DefaultPingInterval   = 54 * time.Second
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
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true
				}
				u, err := url.Parse(origin)
				if err != nil {
					return false
				}
				return u.Host == r.Host
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
	conn.SetReadDeadline(time.Now().Add(DefaultReadDeadline))
	conn.SetPongHandler(func(string) error {
		wsh.hub.UpdateClientPing(clientID)
		conn.SetReadDeadline(time.Now().Add(DefaultReadDeadline))
		return nil
	})
	
	// Start message sender (includes ping via its ticker)
	go wsh.startMessageSender(conn, client)

	// Start message reader (handles incoming messages and keep-alive)
	wsh.startMessageReader(conn, clientID)
}

// startMessageSender sends messages from client channel to WebSocket
func (wsh *WebSocketHandler) startMessageSender(conn *websocket.Conn, client *Client) {
	ticker := time.NewTicker(DefaultPingInterval)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-client.Channel:
			if !ok {
				// Channel closed
				conn.SetWriteDeadline(time.Now().Add(DefaultWriteDeadline))
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			conn.SetWriteDeadline(time.Now().Add(DefaultWriteDeadline))
			// Send message as JSON
			if err := conn.WriteJSON(message); err != nil {
				wsh.logger.Error("realtime.websocket_write_error", map[string]interface{}{
					"client_id": client.ID,
					"error":     err.Error(),
				})
				return
			}

		case <-ticker.C:
			// Send ping
			conn.SetWriteDeadline(time.Now().Add(DefaultWriteDeadline))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				wsh.logger.Error("realtime.websocket_ping_error", map[string]interface{}{
					"client_id": client.ID,
					"error":     err.Error(),
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

