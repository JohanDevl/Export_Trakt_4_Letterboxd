package logger

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/monitoring"
)

func TestVisualFormatterAllCases(t *testing.T) {
	formatter := &VisualFormatter{isQuietMode: false}
	
	tests := []struct {
		name         string
		level        logrus.Level
		message      string
		expectedIcon string
		expectedLevel string
	}{
		{
			name:         "error level",
			level:        logrus.ErrorLevel,
			message:      "error message",
			expectedIcon: "❌",
			expectedLevel: "ERROR",
		},
		{
			name:         "warn level",
			level:        logrus.WarnLevel,
			message:      "warning message",
			expectedIcon: "⚠️ ",
			expectedLevel: "WARN ",
		},
		{
			name:         "info level - success",
			level:        logrus.InfoLevel,
			message:      "Successfully exported data",
			expectedIcon: "✅",
			expectedLevel: "INFO ",
		},
		{
			name:         "info level - retrieved movies",
			level:        logrus.InfoLevel,
			message:      "Retrieved 100 movies from API",
			expectedIcon: "📥",
			expectedLevel: "INFO ",
		},
		{
			name:         "info level - scheduler",
			level:        logrus.InfoLevel,
			message:      "Scheduler is running",
			expectedIcon: "⏰",
			expectedLevel: "INFO ",
		},
		{
			name:         "info level - starting",
			level:        logrus.InfoLevel,
			message:      "Starting application",
			expectedIcon: "🚀",
			expectedLevel: "INFO ",
		},
		{
			name:         "info level - initializing",
			level:        logrus.InfoLevel,
			message:      "Initializing service",
			expectedIcon: "🚀",
			expectedLevel: "INFO ",
		},
		{
			name:         "info level - generic",
			level:        logrus.InfoLevel,
			message:      "generic info message",
			expectedIcon: "ℹ️ ",
			expectedLevel: "INFO ",
		},
		{
			name:         "debug level",
			level:        logrus.DebugLevel,
			message:      "debug message",
			expectedIcon: "🔧",
			expectedLevel: "DEBUG",
		},
		{
			name:         "unknown level",
			level:        logrus.Level(10),
			message:      "unknown message",
			expectedIcon: "📝",
			expectedLevel: "LOG  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &logrus.Entry{
				Time:    time.Now(),
				Level:   tt.level,
				Message: tt.message,
			}

			result, err := formatter.Format(entry)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			output := string(result)
			if !strings.Contains(output, tt.expectedIcon) {
				t.Errorf("Expected output to contain icon '%s', got: %s", tt.expectedIcon, output)
			}
			if !strings.Contains(output, tt.expectedLevel) {
				t.Errorf("Expected output to contain level '%s', got: %s", tt.expectedLevel, output)
			}
			if !strings.Contains(output, tt.message) {
				t.Errorf("Expected output to contain message '%s', got: %s", tt.message, output)
			}
		})
	}
}

func TestVisualFormatterQuietMode(t *testing.T) {
	formatter := &VisualFormatter{isQuietMode: true}
	
	tests := []struct {
		name           string
		message        string
		expectFormatted bool
	}{
		{
			name:           "success message in quiet mode",
			message:        "Successfully exported data",
			expectFormatted: true,
		},
		{
			name:           "retrieved movies in quiet mode",
			message:        "Retrieved 100 movies from API",
			expectFormatted: true,
		},
		{
			name:           "scheduler running in quiet mode",
			message:        "Scheduler is running",
			expectFormatted: true,
		},
		{
			name:           "regular message in quiet mode",
			message:        "regular info message",
			expectFormatted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &logrus.Entry{
				Time:    time.Now(),
				Level:   logrus.InfoLevel,
				Message: tt.message,
			}

			result, err := formatter.Format(entry)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			output := string(result)
			if tt.expectFormatted {
				// Quiet mode messages should have special formatting
				if strings.Contains(tt.message, "Successfully exported") {
					if !strings.Contains(output, "SUCCESS:") {
						t.Errorf("Expected success message formatting, got: %s", output)
					}
				} else if strings.Contains(tt.message, "Retrieved") {
					if !strings.Contains(output, "DATA:") {
						t.Errorf("Expected data message formatting, got: %s", output)
					}
				} else if strings.Contains(tt.message, "Scheduler is running") {
					if !strings.Contains(output, "STATUS:") {
						t.Errorf("Expected status message formatting, got: %s", output)
					}
				}
			} else {
				// Regular formatting in quiet mode
				if !strings.Contains(output, "[INFO ]") {
					t.Errorf("Expected standard formatting, got: %s", output)
				}
			}
		})
	}
}

func TestDualWriter(t *testing.T) {
	var fileBuffer, stdoutBuffer bytes.Buffer
	
	tests := []struct {
		name      string
		quietMode bool
		message   string
		expectStdout bool
	}{
		{
			name:         "quiet mode - important message",
			quietMode:    true,
			message:      "Successfully exported data",
			expectStdout: true,
		},
		{
			name:         "quiet mode - error message",
			quietMode:    true,
			message:      "❌ Error occurred",
			expectStdout: true,
		},
		{
			name:         "quiet mode - level=error",
			quietMode:    true,
			message:      "level=error Something failed",
			expectStdout: true,
		},
		{
			name:         "quiet mode - retrieved message",
			quietMode:    true,
			message:      "Retrieved data from API",
			expectStdout: true,
		},
		{
			name:         "quiet mode - scheduler message",
			quietMode:    true,
			message:      "Scheduler is running",
			expectStdout: true,
		},
		{
			name:         "quiet mode - regular message",
			quietMode:    true,
			message:      "regular debug message",
			expectStdout: false,
		},
		{
			name:         "non-quiet mode - any message",
			quietMode:    false,
			message:      "any message",
			expectStdout: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileBuffer.Reset()
			stdoutBuffer.Reset()
			
			dw := &DualWriter{
				fileWriter:   &fileBuffer,
				stdoutWriter: &stdoutBuffer,
				quietMode:    tt.quietMode,
			}

			n, err := dw.Write([]byte(tt.message))
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if n != len(tt.message) {
				t.Errorf("Expected to write %d bytes, wrote %d", len(tt.message), n)
			}

			// File should always contain the message
			if !strings.Contains(fileBuffer.String(), tt.message) {
				t.Error("File buffer should contain the message")
			}

			// Stdout depends on quiet mode and message type
			stdoutContains := strings.Contains(stdoutBuffer.String(), tt.message)
			if tt.expectStdout && !stdoutContains {
				t.Errorf("Expected stdout to contain message, but it didn't: %s", stdoutBuffer.String())
			}
			if !tt.expectStdout && stdoutContains {
				t.Errorf("Expected stdout to NOT contain message, but it did: %s", stdoutBuffer.String())
			}
		})
	}
}

func TestDualWriterNilFileWriter(t *testing.T) {
	var stdoutBuffer bytes.Buffer
	
	dw := &DualWriter{
		fileWriter:   nil, // Test with nil file writer
		stdoutWriter: &stdoutBuffer,
		quietMode:    false,
	}

	message := "test message"
	n, err := dw.Write([]byte(message))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if n != len(message) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(message), n)
	}

	if !strings.Contains(stdoutBuffer.String(), message) {
		t.Error("Stdout buffer should contain the message")
	}
}

func TestNewLoggerWithQuietMode(t *testing.T) {
	// Test with quiet mode enabled
	os.Setenv("EXPORT_QUIET_MODE", "true")
	defer os.Unsetenv("EXPORT_QUIET_MODE")
	
	logger := NewLogger()
	stdLogger, ok := logger.(*StandardLogger)
	if !ok {
		t.Fatal("Expected StandardLogger implementation")
	}
	
	if !stdLogger.isQuietMode {
		t.Error("Expected quiet mode to be enabled")
	}
	
	// Test with quiet mode disabled
	os.Setenv("EXPORT_QUIET_MODE", "false")
	logger2 := NewLogger()
	stdLogger2, ok := logger2.(*StandardLogger)
	if !ok {
		t.Fatal("Expected StandardLogger implementation")
	}
	
	if stdLogger2.isQuietMode {
		t.Error("Expected quiet mode to be disabled")
	}
}

func TestTranslateEdgeCases(t *testing.T) {
	logger := NewLogger().(*StandardLogger)
	
	// Test with nil translator (already tested but good to have explicit test)
	result := logger.translate("test.message", nil)
	if result != "test.message" {
		t.Errorf("Expected 'test.message', got '%s'", result)
	}
	
	// Test with empty message ID
	result = logger.translate("", map[string]interface{}{"key": "value"})
	if result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}
	
	// Test with translation_failed message ID (should prevent recursion)
	result = logger.translate("errors.translation_failed", nil)
	if result != "errors.translation_failed" {
		t.Errorf("Expected 'errors.translation_failed', got '%s'", result)
	}
}

func TestInfoSpecialMessages(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger().(*StandardLogger)
	logger.SetOutput(&buf)
	
	tests := []struct {
		messageID       string
		expectedMessage string
	}{
		{
			messageID:       "scheduler.started",
			expectedMessage: "Scheduler started successfully!",
		},
		{
			messageID:       "scheduler.waiting",
			expectedMessage: "Scheduler is running. Press Ctrl+C to stop...",
		},
		{
			messageID:       "regular.message",
			expectedMessage: "regular.message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.messageID, func(t *testing.T) {
			buf.Reset()
			logger.Info(tt.messageID)
			
			if !strings.Contains(buf.String(), tt.expectedMessage) {
				t.Errorf("Expected '%s', got '%s'", tt.expectedMessage, buf.String())
			}
		})
	}
}

func TestWarnQuietModeFiltering(t *testing.T) {
	var buf bytes.Buffer
	
	// Test with quiet mode enabled
	os.Setenv("EXPORT_QUIET_MODE", "true")
	defer os.Unsetenv("EXPORT_QUIET_MODE")
	
	logger := NewLogger().(*StandardLogger)
	logger.SetOutput(&buf)
	
	// Translation warning should be filtered in quiet mode
	logger.Warn("translation_not_found.test")
	if buf.String() != "" {
		t.Error("Translation warnings should be filtered in quiet mode")
	}
	
	// Regular warning should still appear
	buf.Reset()
	logger.Warn("regular.warning")
	if !strings.Contains(buf.String(), "regular.warning") {
		t.Error("Regular warnings should appear in quiet mode")
	}
	
	// Test Warnf method
	buf.Reset()
	logger.Warnf("translation_not_found.test", nil)
	if buf.String() != "" {
		t.Error("Translation warnings should be filtered in quiet mode via Warnf")
	}
	
	buf.Reset()
	logger.Warnf("regular.warning", nil)
	if !strings.Contains(buf.String(), "regular.warning") {
		t.Error("Regular warnings should appear in quiet mode via Warnf")
	}
}

func TestStructuredLoggerConfigureOutput(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "logger_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name           string
		config         monitoring.LoggingConfig
		expectedOutput string
	}{
		{
			name: "stdout output",
			config: monitoring.LoggingConfig{
				Level:  "info",
				Format: "json",
				Output: "stdout",
			},
		},
		{
			name: "stderr output",
			config: monitoring.LoggingConfig{
				Level:  "info",
				Format: "json",
				Output: "stderr",
			},
		},
		{
			name: "file output with rotation",
			config: monitoring.LoggingConfig{
				Level:           "info",
				Format:          "json",
				Output:          "file",
				RotationEnabled: true,
			},
		},
		{
			name: "default output",
			config: monitoring.LoggingConfig{
				Level:  "info",
				Format: "json",
				Output: "unknown",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewStructuredLogger(tt.config)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if logger == nil {
				t.Error("Logger should not be nil")
			}
		})
	}
}

func TestStructuredLoggerLevelConfiguration(t *testing.T) {
	tests := []struct {
		level         string
		expectedLevel logrus.Level
	}{
		{"debug", logrus.DebugLevel},
		{"info", logrus.InfoLevel},
		{"warn", logrus.WarnLevel},
		{"warning", logrus.WarnLevel},
		{"error", logrus.ErrorLevel},
		{"fatal", logrus.FatalLevel},
		{"unknown", logrus.InfoLevel}, // default
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			config := monitoring.LoggingConfig{
				Level:  tt.level,
				Format: "json",
				Output: "stdout",
			}
			
			logger, err := NewStructuredLogger(config)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if logger.Logger.GetLevel() != tt.expectedLevel {
				t.Errorf("Expected level %v, got %v", tt.expectedLevel, logger.Logger.GetLevel())
			}
		})
	}
}

func TestStructuredLoggerWithCorrelationID(t *testing.T) {
	config := monitoring.LoggingConfig{
		Level:         "info",
		Format:        "json",
		Output:        "stdout",
		CorrelationID: true,
	}
	
	logger, err := NewStructuredLogger(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	ctx := context.Background()
	ctxWithID := logger.WithCorrelationID(ctx)
	
	correlationID := ctxWithID.Value(CorrelationIDKey)
	if correlationID == nil {
		t.Error("Correlation ID should be set in context")
	}
	
	if correlationIDStr, ok := correlationID.(string); !ok || correlationIDStr == "" {
		t.Error("Correlation ID should be a non-empty string")
	}
}

func TestStructuredLoggerContextMethods(t *testing.T) {
	var buf bytes.Buffer
	
	config := monitoring.LoggingConfig{
		Level:  "debug",
		Format: "json",
		Output: "stdout",
	}
	
	logger, err := NewStructuredLogger(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	logger.Logger.SetOutput(&buf)
	
	// Create context with correlation ID
	ctx := context.Background()
	ctx = logger.WithCorrelationID(ctx)
	
	// Test all context methods
	logger.InfoWithContext(ctx, "test.info")
	logger.ErrorWithContext(ctx, "test.error")
	logger.WarnWithContext(ctx, "test.warn")
	logger.DebugWithContext(ctx, "test.debug")
	
	output := buf.String()
	
	// Should contain correlation ID
	if !strings.Contains(output, "correlation_id") {
		t.Error("Output should contain correlation ID")
	}
	
	// Should contain all log messages
	if !strings.Contains(output, "test.info") {
		t.Error("Output should contain info message")
	}
	if !strings.Contains(output, "test.error") {
		t.Error("Output should contain error message")
	}
	if !strings.Contains(output, "test.warn") {
		t.Error("Output should contain warn message")
	}
	if !strings.Contains(output, "test.debug") {
		t.Error("Output should contain debug message")
	}
}

func TestStructuredLoggerWithContextEntry(t *testing.T) {
	config := monitoring.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	
	logger, err := NewStructuredLogger(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Test with correlation ID
	ctx := context.Background()
	ctx = context.WithValue(ctx, CorrelationIDKey, "test-correlation-id")
	
	entry := logger.WithContext(ctx)
	if entry == nil {
		t.Error("Entry should not be nil")
	}
	
	// Test with empty context
	emptyCtx := context.Background()
	entry2 := logger.WithContext(emptyCtx)
	if entry2 == nil {
		t.Error("Entry should not be nil even with empty context")
	}
}

func TestCorrelationIDHook(t *testing.T) {
	hook := &CorrelationIDHook{}
	
	// Test levels - should return all logrus levels
	levels := hook.Levels()
	allLevels := logrus.AllLevels
	
	if len(levels) != len(allLevels) {
		t.Errorf("Expected %d levels, got %d", len(allLevels), len(levels))
	}
	
	// Test Fire method
	entry := &logrus.Entry{
		Data: make(logrus.Fields),
		Context: context.WithValue(context.Background(), CorrelationIDKey, "test-id"),
	}
	
	err := hook.Fire(entry)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if entry.Data["correlation_id"] != "test-id" {
		t.Errorf("Expected correlation_id 'test-id', got %v", entry.Data["correlation_id"])
	}
	
	// Test with entry without context
	entry2 := &logrus.Entry{
		Data: make(logrus.Fields),
	}
	
	err = hook.Fire(entry2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestTracingHook(t *testing.T) {
	hook := &TracingHook{}
	
	// Test levels
	levels := hook.Levels()
	if len(levels) == 0 {
		t.Error("TracingHook should support some levels")
	}
	
	// Test Fire method with entry without context
	entry := &logrus.Entry{
		Data: make(logrus.Fields),
	}
	
	err := hook.Fire(entry)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Test with context but no span
	entry2 := &logrus.Entry{
		Data: make(logrus.Fields),
		Context: context.Background(),
	}
	
	err = hook.Fire(entry2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestSecurityHook(t *testing.T) {
	sensitiveFields := []string{"password", "token", "secret"}
	hook := NewSecurityHook(sensitiveFields)
	
	if hook == nil {
		t.Fatal("Hook should not be nil")
	}
	
	// Test levels
	levels := hook.Levels()
	if len(levels) == 0 {
		t.Error("SecurityHook should support some levels")
	}
	
	// Test Fire method with sensitive data
	entry := &logrus.Entry{
		Data: logrus.Fields{
			"password": "secret123",
			"token":    "abc123",
			"username": "john",
			"secret":   "mysecret",
		},
	}
	
	err := hook.Fire(entry)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Check that sensitive fields are masked
	if entry.Data["password"] != "[REDACTED]" {
		t.Errorf("Expected password to be redacted, got %v", entry.Data["password"])
	}
	if entry.Data["token"] != "[REDACTED]" {
		t.Errorf("Expected token to be redacted, got %v", entry.Data["token"])
	}
	if entry.Data["secret"] != "[REDACTED]" {
		t.Errorf("Expected secret to be redacted, got %v", entry.Data["secret"])
	}
	if entry.Data["username"] != "john" {
		t.Errorf("Expected username to remain unchanged, got %v", entry.Data["username"])
	}
}

func TestPerformanceHook(t *testing.T) {
	threshold := 100 * time.Millisecond
	hook := NewPerformanceHook(threshold)
	
	if hook == nil {
		t.Fatal("Hook should not be nil")
	}
	
	// Test levels
	levels := hook.Levels()
	if len(levels) == 0 {
		t.Error("PerformanceHook should support some levels")
	}
	
	// Test Fire method - the hook doesn't actually check for duration field
	// It measures its own execution time which will be very fast
	entry := &logrus.Entry{
		Data: logrus.Fields{
			"test": "value",
		},
	}
	
	err := hook.Fire(entry)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// The hook doesn't add slow_operation flag based on input duration
	// It measures its own execution time, which will be fast
	// So we just check that it doesn't error
	
	// Test with another entry
	entry2 := &logrus.Entry{
		Data: logrus.Fields{
			"another": "test",
		},
	}
	
	err = hook.Fire(entry2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestAsyncLogger(t *testing.T) {
	config := monitoring.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	
	asyncLogger, err := NewAsyncLogger(config, 10)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if asyncLogger == nil {
		t.Fatal("AsyncLogger should not be nil")
	}
	
	// Test async logging
	fields := logrus.Fields{
		"test": "value",
	}
	
	asyncLogger.AsyncInfo("test message", fields)
	
	// Give some time for async processing
	time.Sleep(10 * time.Millisecond)
	
	// Close the async logger
	asyncLogger.Close()
}

func TestStructuredLoggerAddHook(t *testing.T) {
	config := monitoring.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	
	logger, err := NewStructuredLogger(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	initialHookCount := len(logger.hooks)
	
	// Add a hook
	hook := &CorrelationIDHook{}
	logger.AddHook(hook)
	
	if len(logger.hooks) != initialHookCount+1 {
		t.Errorf("Expected %d hooks, got %d", initialHookCount+1, len(logger.hooks))
	}
} 