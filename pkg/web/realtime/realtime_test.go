package realtime

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/auth"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/security/keyring"
)

// Mock logger for realtime testing
type mockRealtimeLogger struct {
	logs []map[string]interface{}
}

func (m *mockRealtimeLogger) Debug(msg string, fields ...map[string]interface{}) {
	var fieldsMap map[string]interface{}
	if len(fields) > 0 {
		fieldsMap = fields[0]
	}
	m.logs = append(m.logs, map[string]interface{}{"level": "debug", "msg": msg, "fields": fieldsMap})
}

func (m *mockRealtimeLogger) Info(msg string, fields ...map[string]interface{}) {
	var fieldsMap map[string]interface{}
	if len(fields) > 0 {
		fieldsMap = fields[0]
	}
	m.logs = append(m.logs, map[string]interface{}{"level": "info", "msg": msg, "fields": fieldsMap})
}

func (m *mockRealtimeLogger) Warn(msg string, fields ...map[string]interface{}) {
	var fieldsMap map[string]interface{}
	if len(fields) > 0 {
		fieldsMap = fields[0]
	}
	m.logs = append(m.logs, map[string]interface{}{"level": "warn", "msg": msg, "fields": fieldsMap})
}

func (m *mockRealtimeLogger) Error(msg string, fields ...map[string]interface{}) {
	var fieldsMap map[string]interface{}
	if len(fields) > 0 {
		fieldsMap = fields[0]
	}
	m.logs = append(m.logs, map[string]interface{}{"level": "error", "msg": msg, "fields": fieldsMap})
}

func (m *mockRealtimeLogger) Fatal(msg string, fields ...map[string]interface{}) {
	var fieldsMap map[string]interface{}
	if len(fields) > 0 {
		fieldsMap = fields[0]
	}
	m.logs = append(m.logs, map[string]interface{}{"level": "fatal", "msg": msg, "fields": fieldsMap})
}

func (m *mockRealtimeLogger) Debugf(msg string, data map[string]interface{}) {
	m.logs = append(m.logs, map[string]interface{}{"level": "debug", "msg": msg, "fields": data})
}

func (m *mockRealtimeLogger) Infof(msg string, data map[string]interface{}) {
	m.logs = append(m.logs, map[string]interface{}{"level": "info", "msg": msg, "fields": data})
}

func (m *mockRealtimeLogger) Warnf(msg string, data map[string]interface{}) {
	m.logs = append(m.logs, map[string]interface{}{"level": "warn", "msg": msg, "fields": data})
}

func (m *mockRealtimeLogger) Errorf(msg string, data map[string]interface{}) {
	m.logs = append(m.logs, map[string]interface{}{"level": "error", "msg": msg, "fields": data})
}

func (m *mockRealtimeLogger) SetLogLevel(level string) {}

func (m *mockRealtimeLogger) SetLogFile(path string) error {
	return nil
}

func (m *mockRealtimeLogger) SetTranslator(t logger.Translator) {}

func createTestConfig() *config.Config {
	return &config.Config{
		Auth: config.AuthConfig{
			CallbackPort: 8080,
		},
		Security: config.SecurityConfig{
			RequireHTTPS: false,
		},
		Letterboxd: config.LetterboxdConfig{
			ExportDir: "./test_exports",
		},
	}
}

func TestNewHub(t *testing.T) {
	log := &mockRealtimeLogger{}
	
	hub := NewHub(log)
	
	if hub == nil {
		t.Fatal("Expected hub to be created")
	}
	
	if hub.logger != log {
		t.Error("Expected logger to be set")
	}
	
	if hub.clients == nil {
		t.Error("Expected clients map to be initialized")
	}
	
	if hub.register == nil {
		t.Error("Expected register channel to be initialized")
	}
	
	if hub.unregister == nil {
		t.Error("Expected unregister channel to be initialized")
	}
	
	if hub.broadcast == nil {
		t.Error("Expected broadcast channel to be initialized")
	}
}

func TestNewClient(t *testing.T) {
	clientID := "test-client-123"
	clientType := WebSocketClient
	
	client := NewClient(clientID, clientType)
	
	if client == nil {
		t.Fatal("Expected client to be created")
	}
	
	if client.ID != clientID {
		t.Errorf("Expected client ID '%s', got '%s'", clientID, client.ID)
	}
	
	if client.Type != clientType {
		t.Errorf("Expected client type %d, got %d", clientType, client.Type)
	}
	
	if client.Channel == nil {
		t.Error("Expected client channel to be initialized")
	}
	
	if client.ConnectedAt.IsZero() {
		t.Error("Expected client ConnectedAt to be set")
	}
}

func TestHubRegisterClient(t *testing.T) {
	log := &mockRealtimeLogger{}
	hub := NewHub(log)
	
	// Start hub in background
	go hub.Run()
	defer hub.Stop()
	
	client := NewClient("test-client", WebSocketClient)
	
	hub.RegisterClient(client)
	
	// Give some time for the registration to be processed
	time.Sleep(10 * time.Millisecond)
	
	// Check that client was registered
	if len(hub.clients) != 1 {
		t.Errorf("Expected 1 client to be registered, got %d", len(hub.clients))
	}
	
	if hub.clients[client.ID] != client {
		t.Error("Expected client to be in hub's clients map")
	}
}

func TestHubUnregisterClient(t *testing.T) {
	log := &mockRealtimeLogger{}
	hub := NewHub(log)
	
	go hub.Run()
	defer hub.Stop()
	
	client := NewClient("test-client", WebSocketClient)
	hub.RegisterClient(client)
	
	// Give some time for registration
	time.Sleep(10 * time.Millisecond)
	
	hub.UnregisterClient(client)
	
	// Give some time for unregistration
	time.Sleep(10 * time.Millisecond)
	
	if len(hub.clients) != 0 {
		t.Errorf("Expected 0 clients after unregistration, got %d", len(hub.clients))
	}
}

func TestHubBroadcastMessage(t *testing.T) {
	log := &mockRealtimeLogger{}
	hub := NewHub(log)
	
	go hub.Run()
	defer hub.Stop()
	
	client1 := NewClient("client1", WebSocketClient)
	client2 := NewClient("client2", SSEClient)
	
	hub.RegisterClient(client1)
	hub.RegisterClient(client2)
	
	// Give some time for registration
	time.Sleep(10 * time.Millisecond)
	
	message := Message{
		Type:      "test",
		Data:      "Hello World",
		Timestamp: time.Now(),
	}
	
	hub.BroadcastMessage(message)
	
	// Give some time for message delivery
	time.Sleep(10 * time.Millisecond)
	
	// Check that messages were sent to clients
	select {
	case receivedMsg := <-client1.Channel:
		if receivedMsg.Type != "test" {
			t.Errorf("Expected message type 'test', got '%s'", receivedMsg.Type)
		}
		if receivedMsg.Data != "Hello World" {
			t.Errorf("Expected message data 'Hello World', got '%v'", receivedMsg.Data)
		}
	default:
		t.Error("Expected message to be sent to client1")
	}
	
	select {
	case receivedMsg := <-client2.Channel:
		if receivedMsg.Type != "test" {
			t.Errorf("Expected message type 'test', got '%s'", receivedMsg.Type)
		}
	default:
		t.Error("Expected message to be sent to client2")
	}
}

func TestHubUpdateClientPing(t *testing.T) {
	log := &mockRealtimeLogger{}
	hub := NewHub(log)
	
	go hub.Run()
	defer hub.Stop()
	
	client := NewClient("test-client", WebSocketClient)
	hub.RegisterClient(client)
	
	// Give some time for registration
	time.Sleep(10 * time.Millisecond)
	
	originalLastPing := client.LastPing
	
	hub.UpdateClientPing("test-client")
	
	// Give some time for ping update
	time.Sleep(10 * time.Millisecond)
	
	if !client.LastPing.After(originalLastPing) {
		t.Error("Expected LastPing to be updated")
	}
}

func TestHubGetConnectedClients(t *testing.T) {
	log := &mockRealtimeLogger{}
	hub := NewHub(log)
	
	go hub.Run()
	defer hub.Stop()
	
	client1 := NewClient("client1", WebSocketClient)
	client2 := NewClient("client2", SSEClient)
	
	hub.RegisterClient(client1)
	hub.RegisterClient(client2)
	
	// Give some time for registration
	time.Sleep(10 * time.Millisecond)
	
	clients := hub.GetConnectedClients()
	
	if len(clients) != 2 {
		t.Errorf("Expected 2 connected clients, got %d", len(clients))
	}
	
	// Check that both clients are in the list
	foundClient1 := false
	foundClient2 := false
	
	for _, client := range clients {
		if client.ID == "client1" {
			foundClient1 = true
		}
		if client.ID == "client2" {
			foundClient2 = true
		}
	}
	
	if !foundClient1 {
		t.Error("Expected client1 to be in connected clients list")
	}
	
	if !foundClient2 {
		t.Error("Expected client2 to be in connected clients list")
	}
}

func TestHubStop(t *testing.T) {
	log := &mockRealtimeLogger{}
	hub := NewHub(log)
	
	go hub.Run()
	
	client := NewClient("test-client", WebSocketClient)
	hub.RegisterClient(client)
	
	// Give some time for registration
	time.Sleep(10 * time.Millisecond)
	
	hub.Stop()
	
	// Give some time for cleanup
	time.Sleep(10 * time.Millisecond)
	
	// After stop, client channels should be closed
	select {
	case _, ok := <-client.Channel:
		if ok {
			t.Error("Expected client channel to be closed after hub stop")
		}
	default:
		// Channel might already be closed and drained
	}
}

func TestNewStatusBroadcaster(t *testing.T) {
	log := &mockRealtimeLogger{}
	cfg := createTestConfig()
	
	// Create test keyring manager
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	if err != nil {
		t.Fatalf("Failed to create keyring manager: %v", err)
	}
	
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)
	hub := NewHub(log)
	
	broadcaster := NewStatusBroadcaster(hub, cfg, log, tokenManager)
	
	if broadcaster == nil {
		t.Fatal("Expected status broadcaster to be created")
	}
	
	if broadcaster.hub != hub {
		t.Error("Expected hub to be set")
	}
	
	if broadcaster.config != cfg {
		t.Error("Expected config to be set")
	}
	
	if broadcaster.logger != log {
		t.Error("Expected logger to be set")
	}
	
	if broadcaster.tokenManager != tokenManager {
		t.Error("Expected token manager to be set")
	}
}

func TestStatusBroadcasterBroadcastStatus(t *testing.T) {
	log := &mockRealtimeLogger{}
	cfg := createTestConfig()
	
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	if err != nil {
		t.Fatalf("Failed to create keyring manager: %v", err)
	}
	
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)
	hub := NewHub(log)
	go hub.Run()
	defer hub.Stop()
	
	broadcaster := NewStatusBroadcaster(hub, cfg, log, tokenManager)
	
	// Register a client to receive the broadcast
	client := NewClient("test-client", WebSocketClient)
	hub.RegisterClient(client)
	
	// Give some time for registration
	time.Sleep(10 * time.Millisecond)
	
	broadcaster.BroadcastStatus()
	
	// Give some time for message delivery
	time.Sleep(10 * time.Millisecond)
	
	// Check that status message was sent
	select {
	case message := <-client.Channel:
		if message.Type != "status" {
			t.Errorf("Expected message type 'status', got '%s'", message.Type)
		}
		
		// Verify that data contains expected fields
		statusData, ok := message.Data.(map[string]interface{})
		if !ok {
			t.Error("Expected status data to be a map")
		} else {
			if statusData["server_status"] == nil {
				t.Error("Expected status data to contain server_status")
			}
			if statusData["timestamp"] == nil {
				t.Error("Expected status data to contain timestamp")
			}
		}
	default:
		t.Error("Expected status message to be sent to client")
	}
}

func TestStatusBroadcasterGetSystemStatus(t *testing.T) {
	log := &mockRealtimeLogger{}
	cfg := createTestConfig()
	
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	if err != nil {
		t.Fatalf("Failed to create keyring manager: %v", err)
	}
	
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)
	hub := NewHub(log)
	
	broadcaster := NewStatusBroadcaster(hub, cfg, log, tokenManager)
	
	status := broadcaster.getSystemStatus()
	
	if status["server_status"] == nil {
		t.Error("Expected status to contain server_status")
	}
	
	if status["timestamp"] == nil {
		t.Error("Expected status to contain timestamp")
	}
	
	if status["connected_clients"] == nil {
		t.Error("Expected status to contain connected_clients")
	}
	
	if status["token_status"] == nil {
		t.Error("Expected status to contain token_status")
	}
}

func TestStatusBroadcasterGetExportStatus(t *testing.T) {
	log := &mockRealtimeLogger{}
	cfg := createTestConfig()
	
	// Create test export directory
	err := os.MkdirAll(cfg.Letterboxd.ExportDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test export directory: %v", err)
	}
	defer os.RemoveAll(cfg.Letterboxd.ExportDir)
	
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	if err != nil {
		t.Fatalf("Failed to create keyring manager: %v", err)
	}
	
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)
	hub := NewHub(log)
	
	broadcaster := NewStatusBroadcaster(hub, cfg, log, tokenManager)
	
	exportStatus := broadcaster.getExportStatus()
	
	if exportStatus["export_count"] == nil {
		t.Error("Expected export status to contain export_count")
	}
	
	if exportStatus["last_export"] == nil {
		t.Error("Expected export status to contain last_export")
	}
	
	if exportStatus["total_size"] == nil {
		t.Error("Expected export status to contain total_size")
	}
}

func TestStatusBroadcasterStartAndStop(t *testing.T) {
	log := &mockRealtimeLogger{}
	cfg := createTestConfig()
	
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	if err != nil {
		t.Fatalf("Failed to create keyring manager: %v", err)
	}
	
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)
	hub := NewHub(log)
	go hub.Run()
	defer hub.Stop()
	
	broadcaster := NewStatusBroadcaster(hub, cfg, log, tokenManager)
	
	// Start broadcaster
	go broadcaster.Start()
	
	// Let it run for a short time
	time.Sleep(50 * time.Millisecond)
	
	// Stop broadcaster
	broadcaster.Stop()
	
	// Check that stop was logged
	found := false
	for _, logEntry := range log.logs {
		if logEntry["msg"] == "realtime.status_broadcaster_stopped" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected status broadcaster stop to be logged")
	}
}

func TestNewSimpleWebSocketHandler(t *testing.T) {
	log := &mockRealtimeLogger{}
	hub := NewHub(log)
	
	handler := NewSimpleWebSocketHandler(hub, log)
	
	if handler == nil {
		t.Fatal("Expected WebSocket handler to be created")
	}
	
	if handler.hub != hub {
		t.Error("Expected hub to be set")
	}
	
	if handler.logger != log {
		t.Error("Expected logger to be set")
	}
}

func TestSimpleWebSocketHandlerHandleWebSocket(t *testing.T) {
	log := &mockRealtimeLogger{}
	hub := NewHub(log)
	go hub.Run()
	defer hub.Stop()
	
	handler := NewSimpleWebSocketHandler(hub, log)
	
	// Create a test HTTP request (won't be a real WebSocket upgrade)
	req := httptest.NewRequest("GET", "/ws", nil)
	w := httptest.NewRecorder()
	
	// This will fail because it's not a real WebSocket upgrade request
	handler.HandleWebSocket(w, req)
	
	// Should get a bad request response since WebSocket upgrade will fail
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for non-WebSocket request, got %d", w.Code)
	}
}

func TestNewSSEHandler(t *testing.T) {
	log := &mockRealtimeLogger{}
	hub := NewHub(log)
	
	handler := NewSSEHandler(hub, log)
	
	if handler == nil {
		t.Fatal("Expected SSE handler to be created")
	}
	
	if handler.hub != hub {
		t.Error("Expected hub to be set")
	}
	
	if handler.logger != log {
		t.Error("Expected logger to be set")
	}
}

func TestSSEHandlerSSE(t *testing.T) {
	log := &mockRealtimeLogger{}
	hub := NewHub(log)
	go hub.Run()
	defer hub.Stop()
	
	handler := NewSSEHandler(hub, log)
	
	req := httptest.NewRequest("GET", "/sse", nil)
	w := httptest.NewRecorder()
	
	// Start SSE in a goroutine since it's a blocking operation
	done := make(chan bool)
	go func() {
		handler.HandleSSE(w, req)
		done <- true
	}()
	
	// Give some time for SSE setup
	time.Sleep(50 * time.Millisecond)
	
	// Send a message through the hub
	message := Message{
		Type:      "test",
		Data:      "SSE test message",
		Timestamp: time.Now(),
	}
	hub.BroadcastMessage(message)
	
	// Give some time for message delivery
	time.Sleep(50 * time.Millisecond)
	
	// Check response headers for SSE
	if w.Header().Get("Content-Type") != "text/event-stream" {
		t.Errorf("Expected Content-Type 'text/event-stream', got '%s'", w.Header().Get("Content-Type"))
	}
	
	if w.Header().Get("Cache-Control") != "no-cache" {
		t.Error("Expected Cache-Control 'no-cache'")
	}
	
	if w.Header().Get("Connection") != "keep-alive" {
		t.Error("Expected Connection 'keep-alive'")
	}
	
	select {
	case <-done:
		// SSE handler completed
	case <-time.After(200 * time.Millisecond):
		// This is expected - SSE handler should be still running
	}
}

func TestSSEHandlerStatusAndExports(t *testing.T) {
	log := &mockRealtimeLogger{}
	hub := NewHub(log)
	go hub.Run()
	defer hub.Stop()
	
	handler := NewSSEHandler(hub, log)
	
	// Test HandleSSEStatus
	req := httptest.NewRequest("GET", "/sse/status", nil)
	w := httptest.NewRecorder()
	
	go func() {
		handler.HandleSSEStatus(w, req)
	}()
	
	time.Sleep(10 * time.Millisecond)
	
	// Check SSE headers
	if w.Header().Get("Content-Type") != "text/event-stream" {
		t.Error("Expected SSE Content-Type for status endpoint")
	}
	
	// Test HandleSSEExports
	req2 := httptest.NewRequest("GET", "/sse/exports", nil)
	w2 := httptest.NewRecorder()
	
	go func() {
		handler.HandleSSEExports(w2, req2)
	}()
	
	time.Sleep(10 * time.Millisecond)
	
	if w2.Header().Get("Content-Type") != "text/event-stream" {
		t.Error("Expected SSE Content-Type for exports endpoint")
	}
}

func TestMessage(t *testing.T) {
	// Test Message struct
	now := time.Now()
	message := Message{
		Type:      "test_type",
		Data:      "test_data",
		Timestamp: now,
	}
	
	if message.Type != "test_type" {
		t.Errorf("Expected type 'test_type', got '%s'", message.Type)
	}
	
	if message.Data != "test_data" {
		t.Errorf("Expected data 'test_data', got '%v'", message.Data)
	}
	
	if message.Timestamp != now {
		t.Error("Expected timestamp to match")
	}
}

func TestClientTypeConstants(t *testing.T) {
	// Test ClientType constants
	if WebSocketClient != 0 {
		t.Errorf("Expected WebSocketClient to be 0, got %d", WebSocketClient)
	}
	
	if SSEClient != 1 {
		t.Errorf("Expected SSEClient to be 1, got %d", SSEClient)
	}
}

func TestStatusBroadcasterGetExportStatusWithFiles(t *testing.T) {
	log := &mockRealtimeLogger{}
	cfg := createTestConfig()
	
	// Create test export directory with some files
	err := os.MkdirAll(cfg.Letterboxd.ExportDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test export directory: %v", err)
	}
	defer os.RemoveAll(cfg.Letterboxd.ExportDir)
	
	// Create a test export subdirectory with files
	exportSubDir := filepath.Join(cfg.Letterboxd.ExportDir, "test_export_2023-01-01_12-00-00")
	err = os.MkdirAll(exportSubDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create export subdirectory: %v", err)
	}
	
	// Create test files
	testFile := filepath.Join(exportSubDir, "watched_movies.csv")
	err = os.WriteFile(testFile, []byte("test,csv,content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	if err != nil {
		t.Fatalf("Failed to create keyring manager: %v", err)
	}
	
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)
	hub := NewHub(log)
	broadcaster := NewStatusBroadcaster(hub, cfg, log, tokenManager)
	
	exportStatus := broadcaster.getExportStatus()
	
	// Should find the export
	exportCount, ok := exportStatus["export_count"].(int)
	if !ok || exportCount != 1 {
		t.Errorf("Expected export_count to be 1, got %v", exportStatus["export_count"])
	}
	
	// Should have calculated total size
	totalSize, ok := exportStatus["total_size"].(string)
	if !ok || totalSize == "" {
		t.Error("Expected total_size to be calculated")
	}
}

// Integration test for full realtime system
func TestRealtimeIntegration(t *testing.T) {
	log := &mockRealtimeLogger{}
	cfg := createTestConfig()
	
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	if err != nil {
		t.Fatalf("Failed to create keyring manager: %v", err)
	}
	
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)
	
	// Create and start hub
	hub := NewHub(log)
	go hub.Run()
	defer hub.Stop()
	
	// Create and start status broadcaster
	broadcaster := NewStatusBroadcaster(hub, cfg, log, tokenManager)
	go broadcaster.Start()
	defer broadcaster.Stop()
	
	// Create handlers
	wsHandler := NewSimpleWebSocketHandler(hub, log)
	sseHandler := NewSSEHandler(hub, log)
	
	// Register test clients
	wsClient := NewClient("ws-client", WebSocketClient)
	sseClient := NewClient("sse-client", SSEClient)
	
	hub.RegisterClient(wsClient)
	hub.RegisterClient(sseClient)
	
	// Give some time for setup
	time.Sleep(50 * time.Millisecond)
	
	// Test broadcast functionality
	testMessage := Message{
		Type:      "integration_test",
		Data:      map[string]string{"message": "Hello from integration test"},
		Timestamp: time.Now(),
	}
	
	hub.BroadcastMessage(testMessage)
	
	// Give some time for message delivery
	time.Sleep(50 * time.Millisecond)
	
	// Verify messages were received
	select {
	case receivedMsg := <-wsClient.Channel:
		if receivedMsg.Type != "integration_test" {
			t.Errorf("Expected message type 'integration_test', got '%s'", receivedMsg.Type)
		}
	default:
		t.Error("Expected message to be received by WebSocket client")
	}
	
	select {
	case receivedMsg := <-sseClient.Channel:
		if receivedMsg.Type != "integration_test" {
			t.Errorf("Expected message type 'integration_test', got '%s'", receivedMsg.Type)
		}
	default:
		t.Error("Expected message to be received by SSE client")
	}
	
	// Test client count
	clients := hub.GetConnectedClients()
	if len(clients) != 2 {
		t.Errorf("Expected 2 connected clients, got %d", len(clients))
	}
	
	// Test handler creation
	if wsHandler == nil {
		t.Error("Expected WebSocket handler to be created")
	}
	
	if sseHandler == nil {
		t.Error("Expected SSE handler to be created")
	}
	
	// Verify logging occurred
	if len(log.logs) == 0 {
		t.Error("Expected some log entries to be created during integration test")
	}
}