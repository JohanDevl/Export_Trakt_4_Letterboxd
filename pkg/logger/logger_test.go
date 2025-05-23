package logger

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

// MockTranslator implements the Translator interface for testing
type MockTranslator struct {
	translations map[string]string
}

func NewMockTranslator() *MockTranslator {
	return &MockTranslator{
		translations: map[string]string{
			"test.info":    "Test info message",
			"test.error":   "Test error message",
			"test.warn":    "Test warning message",
			"test.debug":   "Test debug message",
			"test.with_data": "Message with data: {data}",
		},
	}
}

func (m *MockTranslator) Translate(messageID string, templateData map[string]interface{}) string {
	msg, ok := m.translations[messageID]
	if !ok {
		return messageID
	}
	if templateData != nil {
		for key, value := range templateData {
			msg = strings.Replace(msg, "{"+key+"}", fmt.Sprintf("%v", value), -1)
		}
	}
	return msg
}

func TestNewLogger(t *testing.T) {
	logger := NewLogger()
	if logger == nil {
		t.Error("Expected non-nil logger")
	}
	
	// We can't access the internal logger directly in the interface-based implementation
	// Instead, test that basic logging works
	var buf bytes.Buffer
	stdLogger, ok := logger.(*StandardLogger)
	if !ok {
		t.Fatal("Expected StandardLogger implementation")
	}
	stdLogger.SetOutput(&buf)
	
	logger.Info("test message")
	if !strings.Contains(buf.String(), "test message") {
		t.Error("Expected log to contain the test message")
	}
}

func TestSetLogLevel(t *testing.T) {
	tests := []struct {
		name          string
		level         string
		expectedLevel logrus.Level
	}{
		{
			name:          "debug level",
			level:         "debug",
			expectedLevel: logrus.DebugLevel,
		},
		{
			name:          "info level",
			level:         "info",
			expectedLevel: logrus.InfoLevel,
		},
		{
			name:          "warn level",
			level:         "warn",
			expectedLevel: logrus.WarnLevel,
		},
		{
			name:          "error level",
			level:         "error",
			expectedLevel: logrus.ErrorLevel,
		},
		{
			name:          "invalid level defaults to info",
			level:         "invalid",
			expectedLevel: logrus.InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger()
			stdLogger, ok := logger.(*StandardLogger)
			if !ok {
				t.Fatal("Expected StandardLogger implementation")
			}
			
			logger.SetLogLevel(tt.level)
			if stdLogger.GetLevel() != tt.expectedLevel {
				t.Errorf("Expected level %v, got %v", tt.expectedLevel, stdLogger.GetLevel())
			}
		})
	}
}

func TestSetLogFile(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "logger_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name        string
		filePath    string
		expectError bool
	}{
		{
			name:        "valid file path",
			filePath:    filepath.Join(tmpDir, "test.log"),
			expectError: false,
		},
		{
			name:        "invalid directory",
			filePath:    filepath.Join(tmpDir, "nonexistent", "test.log"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger()
			err := logger.SetLogFile(tt.filePath)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				// Check if file was created
				if _, err := os.Stat(tt.filePath); os.IsNotExist(err) {
					t.Error("Expected log file to be created")
				}
			}
		})
	}
}

func TestLoggingWithTranslation(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger()
	stdLogger, ok := logger.(*StandardLogger)
	if !ok {
		t.Fatal("Expected StandardLogger implementation")
	}
	stdLogger.SetOutput(&buf)
	logger.SetLogLevel("debug")

	mockTranslator := NewMockTranslator()
	logger.SetTranslator(mockTranslator)

	tests := []struct {
		name        string
		logFunc     func(messageID string, data map[string]interface{})
		messageID   string
		data        map[string]interface{}
		expectInLog string
	}{
		{
			name: "info message",
			logFunc: func(m string, d map[string]interface{}) {
				logger.Info(m, d)
			},
			messageID:   "test.info",
			data:        nil,
			expectInLog: "Test info message",
		},
		{
			name: "error message",
			logFunc: func(m string, d map[string]interface{}) {
				logger.Error(m, d)
			},
			messageID:   "test.error",
			data:        nil,
			expectInLog: "Test error message",
		},
		{
			name: "warning message",
			logFunc: func(m string, d map[string]interface{}) {
				logger.Warn(m, d)
			},
			messageID:   "test.warn",
			data:        nil,
			expectInLog: "Test warning message",
		},
		{
			name: "debug message",
			logFunc: func(m string, d map[string]interface{}) {
				logger.Debug(m, d)
			},
			messageID:   "test.debug",
			data:        nil,
			expectInLog: "Test debug message",
		},
		{
			name: "message with template data",
			logFunc: func(m string, d map[string]interface{}) {
				logger.Info(m, d)
			},
			messageID: "test.with_data",
			data: map[string]interface{}{
				"data": "test value",
			},
			expectInLog: "Message with data: test value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc(tt.messageID, tt.data)
			logOutput := buf.String()

			if !strings.Contains(logOutput, tt.expectInLog) {
				t.Errorf("Expected log to contain '%s', got '%s'", tt.expectInLog, logOutput)
			}
		})
	}
}

func TestLoggingLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger()
	stdLogger, ok := logger.(*StandardLogger)
	if !ok {
		t.Fatal("Expected StandardLogger implementation")
	}
	stdLogger.SetOutput(&buf)
	logger.SetLogLevel("info")

	mockTranslator := NewMockTranslator()
	logger.SetTranslator(mockTranslator)

	// Debug message should not appear
	logger.Debug("test.debug", nil)
	if strings.Contains(buf.String(), "Test debug message") {
		t.Error("Debug message should not appear when log level is info")
	}

	// Info message should appear
	buf.Reset()
	logger.Info("test.info", nil)
	if !strings.Contains(buf.String(), "Test info message") {
		t.Error("Info message should appear when log level is info")
	}
}

func TestLoggingWithoutTranslator(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger()
	stdLogger, ok := logger.(*StandardLogger)
	if !ok {
		t.Fatal("Expected StandardLogger implementation")
	}
	stdLogger.SetOutput(&buf)
	logger.SetLogLevel("info")

	// Log without translator should use message ID directly
	logger.Info("direct.message", nil)
	if !strings.Contains(buf.String(), "direct.message") {
		t.Error("Message ID should be used directly when no translator is set")
	}
}

// TestFormattingMethods tests all the formatting methods that use the same interface
func TestFormattingMethods(t *testing.T) {
	// Create a new logger and capture its output
	logger := NewLogger()
	stdLogger, ok := logger.(*StandardLogger)
	if !ok {
		t.Fatal("Expected StandardLogger implementation")
	}
	
	var buf bytes.Buffer
	stdLogger.SetOutput(&buf)
	
	// Set the log level to "debug" to ensure all messages are logged
	logger.SetLogLevel("debug")
	
	// Set a mock translator
	mockTranslator := NewMockTranslator()
	logger.SetTranslator(mockTranslator)
	
	tests := []struct {
		name        string
		method      func(string, map[string]interface{})
		messageID   string
		data        map[string]interface{}
		expectInLog string
	}{
		{
			name: "Infof method",
			method: func(m string, d map[string]interface{}) {
				logger.Infof(m, d)
			},
			messageID:   "test.info",
			data:        map[string]interface{}{"data": "formatted info"},
			expectInLog: "Test info message",
		},
		{
			name: "Errorf method",
			method: func(m string, d map[string]interface{}) {
				logger.Errorf(m, d)
			},
			messageID:   "test.error",
			data:        map[string]interface{}{"data": "formatted error"},
			expectInLog: "Test error message",
		},
		{
			name: "Warnf method",
			method: func(m string, d map[string]interface{}) {
				logger.Warnf(m, d)
			},
			messageID:   "test.warn",
			data:        map[string]interface{}{"data": "formatted warning"},
			expectInLog: "Test warning message",
		},
		{
			name: "Debugf method",
			method: func(m string, d map[string]interface{}) {
				logger.Debugf(m, d)
			},
			messageID:   "test.debug",
			data:        map[string]interface{}{"data": "formatted debug"},
			expectInLog: "Test debug message",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.method(tt.messageID, tt.data)
			logOutput := buf.String()
			
			if !strings.Contains(logOutput, tt.expectInLog) {
				t.Errorf("Expected log to contain '%s', got '%s'", tt.expectInLog, logOutput)
			}
		})
	}
	
	// Test with nil translator
	logger.SetTranslator(nil)
	buf.Reset()
	logger.Infof("direct.message", nil)
	if !strings.Contains(buf.String(), "direct.message") {
		t.Error("Expected log to contain the direct message when no translator is set")
	}
} 