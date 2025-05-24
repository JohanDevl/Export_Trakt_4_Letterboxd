package audit

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
	}{
		{
			name: "valid json config",
			config: Config{
				LogLevel:         "info",
				OutputFormat:     "json",
				IncludeSensitive: false,
				RetentionDays:    30,
			},
			expectError: false,
		},
		{
			name: "valid text config",
			config: Config{
				LogLevel:         "debug",
				OutputFormat:     "text",
				IncludeSensitive: true,
				RetentionDays:    90,
			},
			expectError: false,
		},
		{
			name: "invalid log level",
			config: Config{
				LogLevel:     "invalid",
				OutputFormat: "json",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.config)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if logger == nil {
				t.Fatal("Logger should not be nil")
			}
			
			if logger.outputFormat != tt.config.OutputFormat {
				t.Errorf("Expected output format %s, got %s", tt.config.OutputFormat, logger.outputFormat)
			}
			
			if logger.includeSensitive != tt.config.IncludeSensitive {
				t.Errorf("Expected includeSensitive %v, got %v", tt.config.IncludeSensitive, logger.includeSensitive)
			}
		})
	}
}

func TestNewLoggerWithFile(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "audit_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	logFile := filepath.Join(tempDir, "audit.log")
	config := Config{
		LogLevel:     "info",
		OutputFormat: "json",
		LogFile:      logFile,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	if logger.logFile == nil {
		t.Error("Log file should be set")
	}

	// Check if file was created
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file should have been created")
	}
}

func TestLogEvent(t *testing.T) {
	var buf bytes.Buffer
	
	config := Config{
		LogLevel:     "info",
		OutputFormat: "json",
	}
	
	logger, err := NewLogger(config)
	if err != nil {
		t.Fatal(err)
	}
	
	logger.SetOutput(&buf)

	event := AuditEvent{
		EventType: AuthSuccess,
		Severity:  SeverityMedium,
		UserID:    "user123",
		Source:    "auth_system",
		Action:    "login",
		Result:    "success",
		Message:   "User logged in successfully",
		Details: map[string]interface{}{
			"method": "password",
		},
	}

	logger.LogEvent(event)

	// Check that something was logged
	output := buf.String()
	if output == "" {
		t.Fatal("No output was logged")
	}

	// Parse the logged JSON
	var loggedEvent map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &loggedEvent); err != nil {
		t.Fatalf("Failed to parse logged JSON: %v, output was: %s", err, output)
	}

	// Verify key fields
	if loggedEvent["audit_event_type"] != string(AuthSuccess) {
		t.Errorf("Expected event type %s, got %v", AuthSuccess, loggedEvent["audit_event_type"])
	}
	
	if loggedEvent["severity"] != string(SeverityMedium) {
		t.Errorf("Expected severity %s, got %v", SeverityMedium, loggedEvent["severity"])
	}
	
	if loggedEvent["user_id"] != "user123" {
		t.Errorf("Expected user_id user123, got %v", loggedEvent["user_id"])
	}
}

func TestLogEventWithTimestamp(t *testing.T) {
	var buf bytes.Buffer
	
	config := Config{
		LogLevel:     "debug",
		OutputFormat: "json",
	}
	
	logger, err := NewLogger(config)
	if err != nil {
		t.Fatal(err)
	}
	
	logger.SetOutput(&buf)

	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	event := AuditEvent{
		Timestamp: fixedTime,
		EventType: SystemStart,
		Severity:  SeverityLow,
		Source:    "system",
		Action:    "startup",
		Result:    "success",
		Message:   "System started",
	}
	
	logger.LogEvent(event)

	// Check that something was logged
	output := buf.String()
	if output == "" {
		t.Fatal("No output was logged")
	}

	// Parse the logged JSON
	var loggedEvent map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &loggedEvent); err != nil {
		t.Fatalf("Failed to parse logged JSON: %v, output was: %s", err, output)
	}

	// Verify timestamp (should be the fixed time we set)
	expectedTimestamp := fixedTime.Format(time.RFC3339Nano)
	actualTimestamp, ok := loggedEvent["timestamp"].(string)
	if !ok {
		t.Errorf("Timestamp should be a string, got %T", loggedEvent["timestamp"])
	} else if actualTimestamp != expectedTimestamp {
		t.Errorf("Expected timestamp %s, got %s", expectedTimestamp, actualTimestamp)
	}
}

func TestSanitizeEvent(t *testing.T) {
	config := Config{
		LogLevel:         "info",
		OutputFormat:     "json",
		IncludeSensitive: false,
	}
	
	logger, err := NewLogger(config)
	if err != nil {
		t.Fatal(err)
	}

	event := AuditEvent{
		EventType: CredentialAccess,
		Severity:  SeverityHigh,
		Source:    "credential_manager",
		Action:    "retrieve",
		Result:    "success",
		Message:   "Retrieved password=secret123 for user",
		Details: map[string]interface{}{
			"password":      "secret123",
			"access_token":  "token456",
			"normal_field":  "normal_value",
		},
	}

	sanitized := logger.sanitizeEvent(event)

	// Check that sensitive data is redacted
	if !strings.Contains(sanitized.Message, "[REDACTED]") {
		t.Error("Sensitive data in message should be redacted")
	}

	if sanitized.Details["password"] != "[REDACTED]" {
		t.Errorf("Password should be redacted, got %v", sanitized.Details["password"])
	}

	if sanitized.Details["access_token"] != "[REDACTED]" {
		t.Errorf("Access token should be redacted, got %v", sanitized.Details["access_token"])
	}

	if sanitized.Details["normal_field"] != "normal_value" {
		t.Errorf("Normal field should not be redacted, got %v", sanitized.Details["normal_field"])
	}
}

func TestLogCredentialAccess(t *testing.T) {
	var buf bytes.Buffer
	
	config := Config{
		LogLevel:     "info",
		OutputFormat: "json",
	}
	
	logger, err := NewLogger(config)
	if err != nil {
		t.Fatal(err)
	}
	
	logger.SetOutput(&buf)

	logger.LogCredentialAccess("user123", "api_key", "success")

	// Parse the logged JSON
	var loggedEvent map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &loggedEvent); err != nil {
		t.Fatalf("Failed to parse logged JSON: %v", err)
	}

	// Verify key fields
	if loggedEvent["audit_event_type"] != string(CredentialAccess) {
		t.Errorf("Expected event type %s, got %v", CredentialAccess, loggedEvent["audit_event_type"])
	}
	
	if loggedEvent["user_id"] != "user123" {
		t.Errorf("Expected user_id user123, got %v", loggedEvent["user_id"])
	}
	
	if loggedEvent["severity"] != string(SeverityMedium) {
		t.Errorf("Expected severity %s for success, got %v", SeverityMedium, loggedEvent["severity"])
	}
}

func TestLogCredentialAccessFailure(t *testing.T) {
	var buf bytes.Buffer
	
	config := Config{
		LogLevel:     "info",
		OutputFormat: "json",
	}
	
	logger, err := NewLogger(config)
	if err != nil {
		t.Fatal(err)
	}
	
	logger.SetOutput(&buf)

	logger.LogCredentialAccess("user123", "api_key", "failed")

	// Parse the logged JSON
	var loggedEvent map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &loggedEvent); err != nil {
		t.Fatalf("Failed to parse logged JSON: %v", err)
	}

	// Verify severity is high for failure
	if loggedEvent["severity"] != string(SeverityHigh) {
		t.Errorf("Expected severity %s for failure, got %v", SeverityHigh, loggedEvent["severity"])
	}
}

func TestLogAuthEvent(t *testing.T) {
	var buf bytes.Buffer
	
	config := Config{
		LogLevel:     "info",
		OutputFormat: "json",
	}
	
	logger, err := NewLogger(config)
	if err != nil {
		t.Fatal(err)
	}
	
	logger.SetOutput(&buf)

	logger.LogAuthEvent(AuthSuccess, "user123", "success", "192.168.1.1")

	// Parse the logged JSON
	var loggedEvent map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &loggedEvent); err != nil {
		t.Fatalf("Failed to parse logged JSON: %v", err)
	}

	// Verify key fields
	if loggedEvent["audit_event_type"] != string(AuthSuccess) {
		t.Errorf("Expected event type %s, got %v", AuthSuccess, loggedEvent["audit_event_type"])
	}
	
	if loggedEvent["remote_addr"] != "192.168.1.1" {
		t.Errorf("Expected remote_addr 192.168.1.1, got %v", loggedEvent["remote_addr"])
	}
}

func TestLogDataEvent(t *testing.T) {
	var buf bytes.Buffer
	
	config := Config{
		LogLevel:     "info",
		OutputFormat: "json",
	}
	
	logger, err := NewLogger(config)
	if err != nil {
		t.Fatal(err)
	}
	
	logger.SetOutput(&buf)

	logger.LogDataEvent(DataExport, "user123", "movie_data", "export", "success")

	// Parse the logged JSON
	var loggedEvent map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &loggedEvent); err != nil {
		t.Fatalf("Failed to parse logged JSON: %v", err)
	}

	// Verify key fields
	if loggedEvent["audit_event_type"] != string(DataExport) {
		t.Errorf("Expected event type %s, got %v", DataExport, loggedEvent["audit_event_type"])
	}
	
	if loggedEvent["target"] != "movie_data" {
		t.Errorf("Expected target movie_data, got %v", loggedEvent["target"])
	}

	// Export actions should have high severity
	if loggedEvent["severity"] != string(SeverityHigh) {
		t.Errorf("Expected severity %s for export action, got %v", SeverityHigh, loggedEvent["severity"])
	}
}

func TestLogSecurityViolation(t *testing.T) {
	var buf bytes.Buffer
	
	config := Config{
		LogLevel:     "info",
		OutputFormat: "json",
	}
	
	logger, err := NewLogger(config)
	if err != nil {
		t.Fatal(err)
	}
	
	logger.SetOutput(&buf)

	logger.LogSecurityViolation("path_traversal", "file_system", "Attempted access to /etc/passwd", "192.168.1.100")

	// Parse the logged JSON
	var loggedEvent map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &loggedEvent); err != nil {
		t.Fatalf("Failed to parse logged JSON: %v", err)
	}

	// Verify key fields
	if loggedEvent["audit_event_type"] != string(SecurityViolation) {
		t.Errorf("Expected event type %s, got %v", SecurityViolation, loggedEvent["audit_event_type"])
	}
	
	if loggedEvent["severity"] != string(SeverityCritical) {
		t.Errorf("Expected severity %s for security violation, got %v", SeverityCritical, loggedEvent["severity"])
	}
	
	if loggedEvent["remote_addr"] != "192.168.1.100" {
		t.Errorf("Expected remote_addr 192.168.1.100, got %v", loggedEvent["remote_addr"])
	}
}

func TestLogSystemEvent(t *testing.T) {
	var buf bytes.Buffer
	
	config := Config{
		LogLevel:     "info",
		OutputFormat: "json",
	}
	
	logger, err := NewLogger(config)
	if err != nil {
		t.Fatal(err)
	}
	
	logger.SetOutput(&buf)

	logger.LogSystemEvent(SystemStart, "security_manager", "initialize", "success")

	// Parse the logged JSON
	var loggedEvent map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &loggedEvent); err != nil {
		t.Fatalf("Failed to parse logged JSON: %v", err)
	}

	// Verify key fields
	if loggedEvent["audit_event_type"] != string(SystemStart) {
		t.Errorf("Expected event type %s, got %v", SystemStart, loggedEvent["audit_event_type"])
	}
	
	if loggedEvent["source"] != "security_manager" {
		t.Errorf("Expected source security_manager, got %v", loggedEvent["source"])
	}
}

func TestIsSensitiveField(t *testing.T) {
	logger := &Logger{}

	tests := []struct {
		fieldName string
		expected  bool
	}{
		{"password", true},
		{"secret", true},
		{"token", true},
		{"api_key", true},
		{"access_token", true},
		{"client_secret", true},
		{"encryption_key", true},
		{"normal_field", false},
		{"username", false},
		{"email", false},
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			result := logger.isSensitiveField(tt.fieldName)
			if result != tt.expected {
				t.Errorf("isSensitiveField(%s) = %v, expected %v", tt.fieldName, result, tt.expected)
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	logger := &Logger{}

	tests := []struct {
		input    string
		expected string
	}{
		{"password=secret123", "password=[REDACTED]"},
		{"token: abc123", "token: [REDACTED]"},
		{"normal text", "normal text"},
		{"api_key=xyz789", "api_key=[REDACTED]"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := logger.sanitizeString(tt.input)
			if !strings.Contains(result, "[REDACTED]") && strings.Contains(tt.expected, "[REDACTED]") {
				t.Errorf("sanitizeString(%s) should contain [REDACTED], got %s", tt.input, result)
			}
		})
	}
}

func TestGetMetrics(t *testing.T) {
	config := Config{
		LogLevel:     "info",
		OutputFormat: "json",
	}
	
	logger, err := NewLogger(config)
	if err != nil {
		t.Fatal(err)
	}

	metrics := logger.GetMetrics()

	// Check that metrics are returned
	if metrics == nil {
		t.Error("Metrics should not be nil")
	}

	// Check for expected metric keys
	expectedKeys := []string{"retention_hours", "output_format", "include_sensitive", "log_level"}
	for _, key := range expectedKeys {
		if _, exists := metrics[key]; !exists {
			t.Errorf("Expected metric key %s not found", key)
		}
	}
}

func TestCleanupOldLogs(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "audit_cleanup_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	logFile := filepath.Join(tempDir, "audit.log")
	config := Config{
		LogLevel:      "info",
		OutputFormat:  "json",
		LogFile:       logFile,
		RetentionDays: 1, // 1 day retention for testing
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()

	// Create an old log file (simulated)
	oldLogFile := logFile + ".old"
	if err := os.WriteFile(oldLogFile, []byte("old log content"), 0640); err != nil {
		t.Fatal(err)
	}

	// Change the modification time to make it old
	oldTime := time.Now().Add(-48 * time.Hour) // 2 days ago
	if err := os.Chtimes(oldLogFile, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	// Run cleanup
	err = logger.CleanupOldLogs()
	if err != nil {
		t.Fatalf("CleanupOldLogs failed: %v", err)
	}

	// Note: The actual cleanup logic would need to be implemented in the logger
	// This test verifies the method can be called without error
}

func TestClose(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "audit_close_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	logFile := filepath.Join(tempDir, "audit.log")
	config := Config{
		LogLevel:     "info",
		OutputFormat: "json",
		LogFile:      logFile,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatal(err)
	}

	// Close should not return an error
	err = logger.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	// File should be closed (logFile should be nil after close)
	if logger.logFile != nil {
		// Note: This depends on the actual implementation of Close()
		t.Error("Log file should be closed")
	}
}

func TestTextFormatOutput(t *testing.T) {
	var buf bytes.Buffer
	
	config := Config{
		LogLevel:     "info",
		OutputFormat: "text",
	}
	
	logger, err := NewLogger(config)
	if err != nil {
		t.Fatal(err)
	}
	
	logger.SetOutput(&buf)

	event := AuditEvent{
		EventType: AuthSuccess,
		Severity:  SeverityMedium,
		Source:    "auth_system",
		Action:    "login",
		Result:    "success",
		Message:   "User logged in successfully",
	}

	logger.LogEvent(event)

	output := buf.String()
	
	// For text format, we should see readable text output
	if output == "" {
		t.Error("Text output should not be empty")
	}

	// Should not be JSON format
	var jsonTest map[string]interface{}
	if json.Unmarshal(buf.Bytes(), &jsonTest) == nil {
		t.Error("Output should not be valid JSON in text format")
	}
}

func TestEventTypeConstants(t *testing.T) {
	// Test that all event type constants are defined
	eventTypes := []EventType{
		AuthSuccess, AuthFailure, AuthLogout,
		CredentialAccess, CredentialStore, CredentialDelete, CredentialRotation,
		DataExport, DataEncrypt, DataDecrypt, DataAccess,
		SystemStart, SystemStop, SystemError, ConfigChange,
		SecurityViolation, RateLimitHit, PermissionDenied, IntrusionAttempt,
	}

	for _, eventType := range eventTypes {
		if string(eventType) == "" {
			t.Errorf("Event type %v should not be empty", eventType)
		}
	}
}

func TestSeverityConstants(t *testing.T) {
	// Test that all severity constants are defined
	severities := []Severity{
		SeverityLow, SeverityMedium, SeverityHigh, SeverityCritical,
	}

	for _, severity := range severities {
		if string(severity) == "" {
			t.Errorf("Severity %v should not be empty", severity)
		}
	}
} 