package alerts

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/monitoring"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a basic config
func createTestConfig() monitoring.AlertsConfig {
	return monitoring.AlertsConfig{
		WebhookURL:       "",
		EmailEnabled:     false,
		SlackEnabled:     false,
		RateLimitMinutes: 1,
	}
}

func TestNewAlertManager(t *testing.T) {
	logger := logrus.New()
	config := monitoring.AlertsConfig{
		WebhookURL:       "http://example.com/webhook",
		EmailEnabled:     false,
		SlackEnabled:     false,
		RateLimitMinutes: 1,
	}
	
	am := NewAlertManager(logger, config)
	
	assert.NotNil(t, am)
	assert.Equal(t, logger, am.logger)
	assert.Equal(t, config, am.config)
	assert.NotNil(t, am.channels)
	assert.NotNil(t, am.recentAlerts)
	assert.NotNil(t, am.alertHistory)
	
	// Should have webhook channel added since WebhookURL is provided
	assert.Len(t, am.channels, 1)
	assert.Equal(t, "webhook", am.channels[0].Name())
}

func TestAlertManager_AddChannel(t *testing.T) {
	logger := logrus.New()
	config := createTestConfig()
	am := NewAlertManager(logger, config)
	
	// Create a mock channel
	mockChannel := &MockNotificationChannel{
		name:    "test_channel",
		enabled: true,
	}
	
	initialCount := len(am.channels)
	am.AddChannel(mockChannel)
	
	assert.Len(t, am.channels, initialCount+1)
	assert.Contains(t, am.channels, mockChannel)
}

func TestAlertManager_RemoveChannel(t *testing.T) {
	logger := logrus.New()
	config := createTestConfig()
	am := NewAlertManager(logger, config)
	
	// Add a mock channel
	mockChannel := &MockNotificationChannel{
		name:    "test_channel",
		enabled: true,
	}
	am.AddChannel(mockChannel)
	
	initialCount := len(am.channels)
	am.RemoveChannel("test_channel")
	
	assert.Len(t, am.channels, initialCount-1)
	assert.NotContains(t, am.channels, mockChannel)
	
	// Test removing non-existent channel
	am.RemoveChannel("non_existent")
	assert.Len(t, am.channels, initialCount-1)
}

func TestAlertManager_CreateAlert(t *testing.T) {
	logger := logrus.New()
	config := createTestConfig()
	am := NewAlertManager(logger, config)
	
	metadata := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}
	
	alert := am.CreateAlert(
		monitoring.AlertLevelError,
		"Test Alert",
		"Test message",
		"test_source",
		metadata,
	)
	
	assert.NotEmpty(t, alert.ID)
	assert.Equal(t, monitoring.AlertLevelError, alert.Level)
	assert.Equal(t, "Test Alert", alert.Title)
	assert.Equal(t, "Test message", alert.Message)
	assert.Equal(t, "test_source", alert.Source)
	assert.Equal(t, metadata, alert.Metadata)
	assert.False(t, alert.Resolved)
	assert.WithinDuration(t, time.Now(), alert.Timestamp, time.Second)
}

func TestAlertManager_SendAlert(t *testing.T) {
	logger := logrus.New()
	config := createTestConfig()
	am := NewAlertManager(logger, config)
	
	// Add mock channels
	successChannel := &MockNotificationChannel{
		name:    "success_channel",
		enabled: true,
	}
	errorChannel := &MockNotificationChannel{
		name:     "error_channel",
		enabled:  true,
		sendError: fmt.Errorf("send failed"),
	}
	disabledChannel := &MockNotificationChannel{
		name:    "disabled_channel",
		enabled: false,
	}
	
	am.AddChannel(successChannel)
	am.AddChannel(errorChannel)
	am.AddChannel(disabledChannel)
	
	alert := am.CreateAlert(
		monitoring.AlertLevelError,
		"Test Alert",
		"Test message",
		"test_source",
		nil,
	)
	
	err := am.SendAlert(context.Background(), alert)
	
	// Should return error because one channel failed
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error_channel")
	
	// Check that alert was sent to enabled channels
	assert.Len(t, successChannel.sentAlerts, 1)
	assert.Len(t, errorChannel.sentAlerts, 1)
	assert.Len(t, disabledChannel.sentAlerts, 0)
	
	// Check alert history
	history := am.GetAlertHistory()
	assert.Len(t, history, 1)
	assert.Equal(t, alert.ID, history[0].ID)
}

func TestAlertManager_SendExportAlert(t *testing.T) {
	logger := logrus.New()
	config := createTestConfig()
	am := NewAlertManager(logger, config)
	
	mockChannel := &MockNotificationChannel{
		name:    "test_channel",
		enabled: true,
	}
	am.AddChannel(mockChannel)
	
	t.Run("successful export", func(t *testing.T) {
		mockChannel.sentAlerts = nil // Reset
		
		am.SendExportAlert(context.Background(), "movies", "success", 30*time.Second, nil)
		
		assert.Len(t, mockChannel.sentAlerts, 1)
		alert := mockChannel.sentAlerts[0]
		assert.Equal(t, monitoring.AlertLevelInfo, alert.Level)
		assert.Contains(t, alert.Title, "Export Completed")
		assert.Contains(t, alert.Message, "completed successfully")
	})
	
	t.Run("slow export (warning)", func(t *testing.T) {
		mockChannel.sentAlerts = nil // Reset
		
		am.SendExportAlert(context.Background(), "movies", "success", 10*time.Minute, nil)
		
		assert.Len(t, mockChannel.sentAlerts, 1)
		alert := mockChannel.sentAlerts[0]
		assert.Equal(t, monitoring.AlertLevelWarning, alert.Level)
	})
	
	t.Run("failed export", func(t *testing.T) {
		mockChannel.sentAlerts = nil // Reset
		
		testError := fmt.Errorf("export failed")
		am.SendExportAlert(context.Background(), "movies", "error", 1*time.Minute, testError)
		
		assert.Len(t, mockChannel.sentAlerts, 1)
		alert := mockChannel.sentAlerts[0]
		assert.Equal(t, monitoring.AlertLevelError, alert.Level)
		assert.Contains(t, alert.Title, "Export Failed")
		assert.Contains(t, alert.Message, "export failed")
	})
}

func TestAlertManager_SendHealthAlert(t *testing.T) {
	logger := logrus.New()
	config := createTestConfig()
	am := NewAlertManager(logger, config)
	
	mockChannel := &MockNotificationChannel{
		name:    "test_channel",
		enabled: true,
	}
	am.AddChannel(mockChannel)
	
	testCases := []struct {
		status        monitoring.HealthStatus
		expectedLevel monitoring.AlertLevel
		titleContains string
	}{
		{monitoring.HealthStatusUnhealthy, monitoring.AlertLevelCritical, "Unhealthy"},
		{monitoring.HealthStatusDegraded, monitoring.AlertLevelWarning, "Degraded"},
		{monitoring.HealthStatusHealthy, monitoring.AlertLevelInfo, "Recovered"},
	}
	
	for _, tc := range testCases {
		t.Run(string(tc.status), func(t *testing.T) {
			mockChannel.sentAlerts = nil // Reset
			
			am.SendHealthAlert(context.Background(), "test_component", tc.status, "test message")
			
			assert.Len(t, mockChannel.sentAlerts, 1)
			alert := mockChannel.sentAlerts[0]
			assert.Equal(t, tc.expectedLevel, alert.Level)
			assert.Contains(t, alert.Title, tc.titleContains)
		})
	}
}

func TestWebhookChannel(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		
		var payload map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.NoError(t, err)
		
		assert.Contains(t, payload, "id")
		assert.Contains(t, payload, "level")
		assert.Contains(t, payload, "title")
		assert.Contains(t, payload, "message")
		
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	
	logger := logrus.New()
	channel := NewWebhookChannel(logger, server.URL)
	
	assert.Equal(t, "webhook", channel.Name())
	assert.True(t, channel.IsEnabled())
	
	alert := monitoring.Alert{
		ID:        "test_id",
		Level:     monitoring.AlertLevelError,
		Title:     "Test Alert",
		Message:   "Test message",
		Timestamp: time.Now(),
		Source:    "test",
	}
	
	err := channel.SendAlert(context.Background(), alert)
	assert.NoError(t, err)
}

func TestWebhookChannel_Errors(t *testing.T) {
	logger := logrus.New()
	
	t.Run("invalid URL", func(t *testing.T) {
		channel := NewWebhookChannel(logger, "invalid-url")
		
		alert := monitoring.Alert{
			ID:    "test_id",
			Level: monitoring.AlertLevelError,
		}
		
		err := channel.SendAlert(context.Background(), alert)
		assert.Error(t, err)
	})
	
	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()
		
		channel := NewWebhookChannel(logger, server.URL)
		
		alert := monitoring.Alert{
			ID:    "test_id",
			Level: monitoring.AlertLevelError,
		}
		
		err := channel.SendAlert(context.Background(), alert)
		assert.Error(t, err)
	})
}

func TestEmailChannel(t *testing.T) {
	logger := logrus.New()
	config := EmailConfig{
		SMTPHost:    "smtp.example.com",
		SMTPPort:    587,
		Username:    "test@example.com",
		Password:    "password",
		FromAddress: "alerts@example.com",
		FromName:    "Alert System",
		ToAddresses: []string{"admin@example.com"},
	}
	
	channel := NewEmailChannel(logger, config)
	
	assert.Equal(t, "email", channel.Name())
	assert.True(t, channel.IsEnabled())
	assert.NotNil(t, channel.dialer)
	
	// Note: We can't easily test actual email sending without a real SMTP server
	// In a real test environment, you might use a mock SMTP server
}

func TestSlackChannel(t *testing.T) {
	// Create a test server that mimics Slack webhook
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		
		var payload map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.NoError(t, err)
		
		assert.Contains(t, payload, "username")
		assert.Contains(t, payload, "attachments")
		
		// Check that attachments contain the text
		attachments, ok := payload["attachments"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, attachments, 1)
		
		attachment, ok := attachments[0].(map[string]interface{})
		assert.True(t, ok)
		assert.Contains(t, attachment, "text")
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()
	
	logger := logrus.New()
	config := SlackConfig{
		WebhookURL: server.URL,
		Channel:    "#alerts",
		Username:   "Alert Bot",
	}
	
	channel := NewSlackChannel(logger, config)
	
	assert.Equal(t, "slack", channel.Name())
	assert.True(t, channel.IsEnabled())
	
	alert := monitoring.Alert{
		ID:        "test_id",
		Level:     monitoring.AlertLevelError,
		Title:     "Test Alert",
		Message:   "Test message",
		Timestamp: time.Now(),
		Source:    "test",
		Metadata: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
	}
	
	err := channel.SendAlert(context.Background(), alert)
	assert.NoError(t, err)
}

func TestFormatMetadata(t *testing.T) {
	metadata := map[string]interface{}{
		"string_key": "string_value",
		"int_key":    42,
		"bool_key":   true,
		"float_key":  3.14,
	}
	
	formatted := formatMetadata(metadata)
	
	assert.Contains(t, formatted, "string_key: string_value")
	assert.Contains(t, formatted, "int_key: 42")
	assert.Contains(t, formatted, "bool_key: true")
	assert.Contains(t, formatted, "float_key: 3.14")
}

// MockNotificationChannel for testing
type MockNotificationChannel struct {
	name       string
	enabled    bool
	sendError  error
	sentAlerts []monitoring.Alert
}

func (m *MockNotificationChannel) Name() string {
	return m.name
}

func (m *MockNotificationChannel) IsEnabled() bool {
	return m.enabled
}

func (m *MockNotificationChannel) SendAlert(ctx context.Context, alert monitoring.Alert) error {
	if m.sendError != nil {
		// Still add to sent alerts for testing purposes
		m.sentAlerts = append(m.sentAlerts, alert)
		return m.sendError
	}
	
	m.sentAlerts = append(m.sentAlerts, alert)
	return nil
}

func TestAlertHistory(t *testing.T) {
	logger := logrus.New()
	config := createTestConfig()
	am := NewAlertManager(logger, config)
	
	// Create and send multiple alerts
	for i := 0; i < 3; i++ {
		alert := am.CreateAlert(
			monitoring.AlertLevelInfo,
			fmt.Sprintf("Alert %d", i),
			fmt.Sprintf("Message %d", i),
			"test",
			nil,
		)
		am.addToHistory(alert)
	}
	
	history := am.GetAlertHistory()
	
	// Should have all alerts in history
	assert.Len(t, history, 3)
}

func TestConcurrentAlertSending(t *testing.T) {
	logger := logrus.New()
	config := createTestConfig()
	am := NewAlertManager(logger, config)
	
	mockChannel := &MockNotificationChannel{
		name:    "test_channel",
		enabled: true,
	}
	am.AddChannel(mockChannel)
	
	// Send alerts concurrently
	numGoroutines := 5 // Reduced for simpler test
	done := make(chan bool, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			alert := am.CreateAlert(
				monitoring.AlertLevelInfo,
				fmt.Sprintf("Concurrent Alert %d", id),
				fmt.Sprintf("Message from goroutine %d", id),
				"test",
				nil,
			)
			am.SendAlert(context.Background(), alert)
			done <- true
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	
	// All alerts should have been sent (no race conditions)
	assert.Len(t, mockChannel.sentAlerts, numGoroutines)
	
	// Check alert history has all alerts
	history := am.GetAlertHistory()
	assert.Len(t, history, numGoroutines)
} 