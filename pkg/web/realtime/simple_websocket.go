package realtime

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

// SimpleWebSocketHandler provides basic WebSocket functionality without external dependencies
type SimpleWebSocketHandler struct {
	hub    *Hub
	logger logger.Logger
}

// NewSimpleWebSocketHandler creates a new simple WebSocket handler
func NewSimpleWebSocketHandler(hub *Hub, logger logger.Logger) *SimpleWebSocketHandler {
	return &SimpleWebSocketHandler{
		hub:    hub,
		logger: logger,
	}
}

// HandleWebSocket handles WebSocket upgrade and connection using basic implementation
func (swsh *SimpleWebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Check if it's a WebSocket upgrade request
	if !isWebSocketUpgrade(r) {
		http.Error(w, "Expected WebSocket upgrade", http.StatusBadRequest)
		return
	}

	// Perform WebSocket handshake
	conn, err := swsh.upgradeConnection(w, r)
	if err != nil {
		swsh.logger.Error("realtime.websocket_upgrade_failed", map[string]interface{}{
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
	swsh.hub.RegisterClient(client)
	defer swsh.hub.UnregisterClient(client)
	
	swsh.logger.Info("realtime.websocket_connected", map[string]interface{}{
		"client_id": clientID,
		"remote_addr": r.RemoteAddr,
	})

	// Handle messages
	swsh.handleMessages(conn, client)
}

// isWebSocketUpgrade checks if the request is a WebSocket upgrade
func isWebSocketUpgrade(r *http.Request) bool {
	return strings.ToLower(r.Header.Get("Connection")) == "upgrade" &&
		strings.ToLower(r.Header.Get("Upgrade")) == "websocket" &&
		r.Header.Get("Sec-WebSocket-Key") != ""
}

// upgradeConnection performs the WebSocket handshake
func (swsh *SimpleWebSocketHandler) upgradeConnection(w http.ResponseWriter, r *http.Request) (net.Conn, error) {
	// Get the WebSocket key
	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		return nil, fmt.Errorf("missing Sec-WebSocket-Key header")
	}

	// Generate accept key
	acceptKey := generateAcceptKey(key)

	// Hijack the connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return nil, fmt.Errorf("connection hijacking not supported")
	}

	conn, bufrw, err := hijacker.Hijack()
	if err != nil {
		return nil, fmt.Errorf("hijack failed: %w", err)
	}

	// Send handshake response
	response := fmt.Sprintf(
		"HTTP/1.1 101 Switching Protocols\r\n"+
			"Upgrade: websocket\r\n"+
			"Connection: Upgrade\r\n"+
			"Sec-WebSocket-Accept: %s\r\n\r\n",
		acceptKey)

	if _, err := bufrw.WriteString(response); err != nil {
		conn.Close()
		return nil, fmt.Errorf("handshake response failed: %w", err)
	}

	if err := bufrw.Flush(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("handshake flush failed: %w", err)
	}

	return conn, nil
}

// generateAcceptKey generates the WebSocket accept key
func generateAcceptKey(key string) string {
	const websocketGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	hash := sha1.Sum([]byte(key + websocketGUID))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// handleMessages handles the WebSocket message loop
func (swsh *SimpleWebSocketHandler) handleMessages(conn net.Conn, client *Client) {
	reader := bufio.NewReader(conn)
	
	// Set connection timeout
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	
	// Start message sender goroutine
	go swsh.sendMessages(conn, client)
	
	// Read messages (basic implementation for keepalive)
	for {
		// Try to read a frame (simplified)
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		
		// For simplicity, we'll just read bytes and update ping time
		buffer := make([]byte, 1024)
		_, err := reader.Read(buffer)
		if err != nil {
			swsh.logger.Debug("realtime.websocket_read_error", map[string]interface{}{
				"client_id": client.ID,
				"error": err.Error(),
			})
			break
		}
		
		// Update ping time
		swsh.hub.UpdateClientPing(client.ID)
	}
}

// sendMessages sends messages from client channel to WebSocket
func (swsh *SimpleWebSocketHandler) sendMessages(conn net.Conn, client *Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case message, ok := <-client.Channel:
			if !ok {
				return
			}
			
			// Convert message to JSON
			jsonData, err := json.Marshal(message)
			if err != nil {
				swsh.logger.Error("realtime.websocket_marshal_error", map[string]interface{}{
					"client_id": client.ID,
					"error": err.Error(),
				})
				continue
			}
			
			// Send as WebSocket text frame (simplified)
			if err := swsh.sendTextFrame(conn, jsonData); err != nil {
				swsh.logger.Error("realtime.websocket_send_error", map[string]interface{}{
					"client_id": client.ID,
					"error": err.Error(),
				})
				return
			}
			
		case <-ticker.C:
			// Send ping frame
			if err := swsh.sendPingFrame(conn); err != nil {
				swsh.logger.Error("realtime.websocket_ping_error", map[string]interface{}{
					"client_id": client.ID,
					"error": err.Error(),
				})
				return
			}
		}
	}
}

// sendTextFrame sends a WebSocket text frame (simplified implementation)
func (swsh *SimpleWebSocketHandler) sendTextFrame(conn net.Conn, data []byte) error {
	frame := swsh.createFrame(0x1, data) // 0x1 = text frame
	_, err := conn.Write(frame)
	return err
}

// sendPingFrame sends a WebSocket ping frame
func (swsh *SimpleWebSocketHandler) sendPingFrame(conn net.Conn) error {
	frame := swsh.createFrame(0x9, []byte{}) // 0x9 = ping frame
	_, err := conn.Write(frame)
	return err
}

// createFrame creates a basic WebSocket frame
func (swsh *SimpleWebSocketHandler) createFrame(opcode byte, data []byte) []byte {
	dataLen := len(data)
	
	var frame []byte
	frame = append(frame, 0x80|opcode) // FIN=1, opcode
	
	if dataLen < 126 {
		frame = append(frame, byte(dataLen))
	} else if dataLen < 65536 {
		frame = append(frame, 126)
		frame = append(frame, byte(dataLen>>8), byte(dataLen))
	} else {
		frame = append(frame, 127)
		for i := 7; i >= 0; i-- {
			frame = append(frame, byte(dataLen>>(i*8)))
		}
	}
	
	frame = append(frame, data...)
	return frame
}