package realtime

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

// MessageType represents the type of real-time message
type MessageType string

const (
	StatusUpdate    MessageType = "status_update"
	ExportProgress  MessageType = "export_progress"
	LogEntry        MessageType = "log_entry"
	Alert           MessageType = "alert"
	ServerHealth    MessageType = "server_health"
	TokenUpdate     MessageType = "token_update"
)

// Message represents a real-time message
type Message struct {
	Type      MessageType `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
	ID        string      `json:"id"`
}

// Client represents a connected client
type Client struct {
	ID        string
	Type      ClientType
	Channel   chan Message
	Connected time.Time
	LastPing  time.Time
}

// ClientType represents the type of client connection
type ClientType string

const (
	WebSocketClient ClientType = "websocket"
	SSEClient       ClientType = "sse"
)

// Hub manages all connected clients and message broadcasting
type Hub struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan Message
	logger     logger.Logger
	mutex      sync.RWMutex
	
	// Statistics
	stats HubStats
}

// HubStats tracks hub statistics
type HubStats struct {
	TotalClients     int
	WebSocketClients int
	SSEClients      int
	MessagesTotal    int64
	BytesSent        int64
	StartTime        time.Time
}

// NewHub creates a new real-time message hub
func NewHub(logger logger.Logger) *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client, 100),
		unregister: make(chan *Client, 100),
		broadcast:  make(chan Message, 1000),
		logger:     logger,
		stats: HubStats{
			StartTime: time.Now(),
		},
	}
}

// Start begins processing hub operations
func (h *Hub) Start() {
	h.logger.Info("realtime.hub_starting", nil)
	
	// Start cleanup goroutine
	go h.startCleanupRoutine()
	
	// Main hub loop
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)
			
		case client := <-h.unregister:
			h.unregisterClient(client)
			
		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// RegisterClient adds a new client to the hub
func (h *Hub) RegisterClient(client *Client) {
	select {
	case h.register <- client:
	default:
		h.logger.Warn("realtime.register_channel_full", map[string]interface{}{
			"client_id": client.ID,
		})
	}
}

// UnregisterClient removes a client from the hub
func (h *Hub) UnregisterClient(client *Client) {
	select {
	case h.unregister <- client:
	default:
		h.logger.Warn("realtime.unregister_channel_full", map[string]interface{}{
			"client_id": client.ID,
		})
	}
}

// BroadcastMessage sends a message to all connected clients
func (h *Hub) BroadcastMessage(msgType MessageType, payload interface{}) {
	message := Message{
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now(),
		ID:        generateMessageID(),
	}
	
	select {
	case h.broadcast <- message:
	default:
		h.logger.Warn("realtime.broadcast_channel_full", map[string]interface{}{
			"message_type": string(msgType),
		})
	}
}

// registerClient handles client registration
func (h *Hub) registerClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	h.clients[client.ID] = client
	h.updateStats()
	
	h.logger.Info("realtime.client_connected", map[string]interface{}{
		"client_id":    client.ID,
		"client_type":  string(client.Type),
		"total_clients": len(h.clients),
	})
	
	// Send welcome message
	welcomeMessage := Message{
		Type: StatusUpdate,
		Payload: map[string]interface{}{
			"message": "Connected to real-time updates",
			"server_time": time.Now().Format(time.RFC3339),
		},
		Timestamp: time.Now(),
		ID:        generateMessageID(),
	}
	
	select {
	case client.Channel <- welcomeMessage:
	default:
		h.logger.Warn("realtime.welcome_message_failed", map[string]interface{}{
			"client_id": client.ID,
		})
	}
}

// unregisterClient handles client disconnection
func (h *Hub) unregisterClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	if _, exists := h.clients[client.ID]; exists {
		delete(h.clients, client.ID)
		close(client.Channel)
		h.updateStats()
		
		h.logger.Info("realtime.client_disconnected", map[string]interface{}{
			"client_id":     client.ID,
			"client_type":   string(client.Type),
			"connected_duration": time.Since(client.Connected).String(),
			"total_clients": len(h.clients),
		})
	}
}

// broadcastMessage sends a message to all connected clients
func (h *Hub) broadcastMessage(message Message) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	if len(h.clients) == 0 {
		return
	}
	
	// Marshal message to JSON for size calculation
	messageData, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("realtime.message_marshal_error", map[string]interface{}{
			"error": err.Error(),
			"message_type": string(message.Type),
		})
		return
	}
	
	deliveredCount := 0
	failedCount := 0
	
	for clientID, client := range h.clients {
		select {
		case client.Channel <- message:
			deliveredCount++
		default:
			failedCount++
			h.logger.Warn("realtime.message_delivery_failed", map[string]interface{}{
				"client_id": clientID,
				"message_type": string(message.Type),
			})
		}
	}
	
	// Update statistics
	h.stats.MessagesTotal++
	h.stats.BytesSent += int64(len(messageData) * deliveredCount)
	
	h.logger.Debug("realtime.message_broadcast", map[string]interface{}{
		"message_type": string(message.Type),
		"total_clients": len(h.clients),
		"delivered": deliveredCount,
		"failed": failedCount,
		"message_size": len(messageData),
	})
}

// updateStats updates client statistics
func (h *Hub) updateStats() {
	h.stats.TotalClients = len(h.clients)
	h.stats.WebSocketClients = 0
	h.stats.SSEClients = 0
	
	for _, client := range h.clients {
		switch client.Type {
		case WebSocketClient:
			h.stats.WebSocketClients++
		case SSEClient:
			h.stats.SSEClients++
		}
	}
}

// startCleanupRoutine starts a goroutine to clean up stale connections
func (h *Hub) startCleanupRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		h.cleanupStaleClients()
	}
}

// cleanupStaleClients removes clients that haven't pinged recently
func (h *Hub) cleanupStaleClients() {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	staleThreshold := time.Now().Add(-2 * time.Minute)
	staleClients := make([]*Client, 0)
	
	for clientID, client := range h.clients {
		if client.LastPing.Before(staleThreshold) {
			staleClients = append(staleClients, client)
			delete(h.clients, clientID)
		}
	}
	
	// Close stale client channels
	for _, client := range staleClients {
		close(client.Channel)
		h.logger.Info("realtime.stale_client_removed", map[string]interface{}{
			"client_id": client.ID,
			"last_ping": client.LastPing.Format(time.RFC3339),
		})
	}
	
	if len(staleClients) > 0 {
		h.updateStats()
	}
}

// GetStats returns hub statistics
func (h *Hub) GetStats() HubStats {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	stats := h.stats
	stats.TotalClients = len(h.clients)
	return stats
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}

// UpdateClientPing updates the last ping time for a client
func (h *Hub) UpdateClientPing(clientID string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	if client, exists := h.clients[clientID]; exists {
		client.LastPing = time.Now()
	}
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	return time.Now().Format("20060102-150405.000000")
}

// NewClient creates a new client
func NewClient(id string, clientType ClientType) *Client {
	return &Client{
		ID:        id,
		Type:      clientType,
		Channel:   make(chan Message, 100),
		Connected: time.Now(),
		LastPing:  time.Now(),
	}
}