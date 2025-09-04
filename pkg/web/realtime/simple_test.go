package realtime

import (
	"testing"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
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
		Letterboxd: config.LetterboxdConfig{
			ExportDir: "./test_exports",
		},
	}
}

// Test basic hub creation
func TestNewHub(t *testing.T) {
	log := &mockRealtimeLogger{}
	hub := NewHub(log)

	if hub == nil {
		t.Fatal("Expected hub to be created")
	}

	if len(hub.clients) != 0 {
		t.Errorf("Expected no clients initially, got %d", len(hub.clients))
	}
}

// Test client creation
func TestNewClient(t *testing.T) {
	client := NewClient("test-123", WebSocketClient)

	if client == nil {
		t.Fatal("Expected client to be created")
	}

	if client.ID != "test-123" {
		t.Errorf("Expected client ID 'test-123', got '%s'", client.ID)
	}

	if client.Type != WebSocketClient {
		t.Errorf("Expected client type WebSocketClient, got %s", client.Type)
	}

	if client.Channel == nil {
		t.Error("Expected client channel to be initialized")
	}

	if client.Connected.IsZero() {
		t.Error("Expected client Connected to be set")
	}
}

// Test message creation
func TestMessage(t *testing.T) {
	msg := Message{
		Type:      StatusUpdate,
		Payload:   "test payload",
		Timestamp: time.Now(),
		ID:        "test-id",
	}

	if msg.Type != StatusUpdate {
		t.Errorf("Expected message type StatusUpdate, got %s", msg.Type)
	}

	if msg.Payload != "test payload" {
		t.Errorf("Expected payload 'test payload', got %v", msg.Payload)
	}
}

// Test BroadcastMessage
func TestHubBroadcastMessage(t *testing.T) {
	log := &mockRealtimeLogger{}
	hub := NewHub(log)

	// Test that BroadcastMessage doesn't panic
	hub.BroadcastMessage(StatusUpdate, "test message")

	// Should not panic, which means test passes
}

// Test client registration
func TestHubRegisterClient(t *testing.T) {
	log := &mockRealtimeLogger{}
	hub := NewHub(log)
	client := NewClient("test-client", WebSocketClient)

	// Test RegisterClient doesn't panic
	hub.RegisterClient(client)

	// Should not panic, which means test passes
}