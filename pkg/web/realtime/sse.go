package realtime

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

// SSEHandler handles Server-Sent Events connections
type SSEHandler struct {
	hub    *Hub
	logger logger.Logger
}

// NewSSEHandler creates a new SSE handler
func NewSSEHandler(hub *Hub, logger logger.Logger) *SSEHandler {
	return &SSEHandler{
		hub:    hub,
		logger: logger,
	}
}

// HandleSSE handles Server-Sent Events connection
func (ssh *SSEHandler) HandleSSE(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// Check if the connection supports flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		ssh.logger.Error("realtime.sse_no_flusher", map[string]interface{}{
			"remote_addr": r.RemoteAddr,
		})
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// Generate client ID
	clientID := fmt.Sprintf("sse_%d_%s", time.Now().Unix(), r.RemoteAddr)
	
	// Create client
	client := NewClient(clientID, SSEClient)
	
	// Register client with hub
	ssh.hub.RegisterClient(client)
	defer ssh.hub.UnregisterClient(client)
	
	ssh.logger.Info("realtime.sse_connected", map[string]interface{}{
		"client_id": clientID,
		"remote_addr": r.RemoteAddr,
		"user_agent": r.UserAgent(),
	})

	// Send initial connection message
	ssh.sendSSEMessage(w, flusher, "connection", map[string]interface{}{
		"status": "connected",
		"client_id": clientID,
		"server_time": time.Now().Format(time.RFC3339),
	})

	// Create context for cleanup
	ctx := r.Context()
	
	// Start keep-alive routine
	keepAliveTicker := time.NewTicker(30 * time.Second)
	defer keepAliveTicker.Stop()
	
	// Message processing loop
	for {
		select {
		case <-ctx.Done():
			// Client disconnected
			ssh.logger.Info("realtime.sse_disconnected", map[string]interface{}{
				"client_id": clientID,
				"duration": time.Since(client.Connected).String(),
			})
			return
			
		case message, ok := <-client.Channel:
			if !ok {
				// Channel closed
				return
			}
			
			// Send message via SSE
			ssh.sendSSEMessage(w, flusher, string(message.Type), message.Payload)
			
		case <-keepAliveTicker.C:
			// Send keep-alive ping
			ssh.sendSSEMessage(w, flusher, "ping", map[string]interface{}{
				"timestamp": time.Now().Format(time.RFC3339),
			})
			
			// Update client ping time
			ssh.hub.UpdateClientPing(clientID)
		}
	}
}

// sendSSEMessage sends a Server-Sent Event message
func (ssh *SSEHandler) sendSSEMessage(w http.ResponseWriter, flusher http.Flusher, eventType string, data interface{}) {
	// Marshal data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		ssh.logger.Error("realtime.sse_marshal_error", map[string]interface{}{
			"error": err.Error(),
			"event_type": eventType,
		})
		return
	}
	
	// Write SSE format
	// Format: event: eventType\ndata: jsonData\n\n
	fmt.Fprintf(w, "event: %s\n", eventType)
	fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
	
	// Flush the data to the client
	flusher.Flush()
}

// HandleSSEStatus provides a specialized SSE endpoint for status updates only
func (ssh *SSEHandler) HandleSSEStatus(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// Generate client ID
	clientID := fmt.Sprintf("sse_status_%d_%s", time.Now().Unix(), r.RemoteAddr)
	
	// Create client
	client := NewClient(clientID, SSEClient)
	
	// Register client with hub
	ssh.hub.RegisterClient(client)
	defer ssh.hub.UnregisterClient(client)

	ctx := r.Context()
	keepAliveTicker := time.NewTicker(30 * time.Second)
	defer keepAliveTicker.Stop()

	// Send initial status
	ssh.sendSSEMessage(w, flusher, "status_update", map[string]interface{}{
		"connected": true,
		"timestamp": time.Now().Format(time.RFC3339),
	})

	for {
		select {
		case <-ctx.Done():
			return
			
		case message, ok := <-client.Channel:
			if !ok {
				return
			}
			
			// Only send status-related messages
			if message.Type == StatusUpdate || message.Type == ServerHealth || message.Type == TokenUpdate {
				ssh.sendSSEMessage(w, flusher, string(message.Type), message.Payload)
			}
			
		case <-keepAliveTicker.C:
			ssh.sendSSEMessage(w, flusher, "ping", map[string]interface{}{
				"timestamp": time.Now().Format(time.RFC3339),
			})
			ssh.hub.UpdateClientPing(clientID)
		}
	}
}

// HandleSSEExports provides a specialized SSE endpoint for export progress updates
func (ssh *SSEHandler) HandleSSEExports(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// Generate client ID
	clientID := fmt.Sprintf("sse_exports_%d_%s", time.Now().Unix(), r.RemoteAddr)
	
	// Create client
	client := NewClient(clientID, SSEClient)
	
	// Register client with hub
	ssh.hub.RegisterClient(client)
	defer ssh.hub.UnregisterClient(client)

	ctx := r.Context()
	keepAliveTicker := time.NewTicker(30 * time.Second)
	defer keepAliveTicker.Stop()

	// Send initial connection message
	ssh.sendSSEMessage(w, flusher, "export_connection", map[string]interface{}{
		"connected": true,
		"timestamp": time.Now().Format(time.RFC3339),
	})

	for {
		select {
		case <-ctx.Done():
			return
			
		case message, ok := <-client.Channel:
			if !ok {
				return
			}
			
			// Only send export-related messages
			if message.Type == ExportProgress || message.Type == Alert {
				ssh.sendSSEMessage(w, flusher, string(message.Type), message.Payload)
			}
			
		case <-keepAliveTicker.C:
			ssh.sendSSEMessage(w, flusher, "ping", map[string]interface{}{
				"timestamp": time.Now().Format(time.RFC3339),
			})
			ssh.hub.UpdateClientPing(clientID)
		}
	}
}