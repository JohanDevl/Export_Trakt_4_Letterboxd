package i18n

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

// MockLogger implements the Logger interface for testing
type MockLogger struct {
	infoMessages  []string
	errorMessages []string
	warnMessages  []string
	debugMessages []string
}

func NewMockLogger() *MockLogger {
	return &MockLogger{
		infoMessages:  []string{},
		errorMessages: []string{},
		warnMessages:  []string{},
		debugMessages: []string{},
	}
}

func (m *MockLogger) Info(messageID string, data ...map[string]interface{}) { 
	m.infoMessages = append(m.infoMessages, messageID) 
}
func (m *MockLogger) Error(messageID string, data ...map[string]interface{}) { 
	m.errorMessages = append(m.errorMessages, messageID) 
}
func (m *MockLogger) Warn(messageID string, data ...map[string]interface{}) { 
	m.warnMessages = append(m.warnMessages, messageID) 
}
func (m *MockLogger) Debug(messageID string, data ...map[string]interface{}) { 
	m.debugMessages = append(m.debugMessages, messageID) 
}
func (m *MockLogger) Infof(format string, data map[string]interface{}) {}
func (m *MockLogger) Errorf(format string, data map[string]interface{}) {}
func (m *MockLogger) Warnf(format string, data map[string]interface{}) {}
func (m *MockLogger) Debugf(format string, data map[string]interface{}) {}
func (m *MockLogger) SetLogLevel(level string) {}
func (m *MockLogger) SetLogFile(filePath string) error { return nil }
func (m *MockLogger) SetTranslator(t logger.Translator) {}

func TestNewTranslator(t *testing.T) {
	// Create a temporary directory for test translation files
	tempDir, err := os.MkdirTemp("", "i18n_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a basic English translation file
	enContent := `{
		"test": {
			"message": "This is a test message"
		}
	}`
	err = os.WriteFile(filepath.Join(tempDir, "en.json"), []byte(enContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test translation file: %v", err)
	}

	// Create a basic French translation file
	frContent := `{
		"test": {
			"message": "C'est un message de test"
		}
	}`
	err = os.WriteFile(filepath.Join(tempDir, "fr.json"), []byte(frContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test translation file: %v", err)
	}

	tests := []struct {
		name           string
		config         config.I18nConfig
		expectError    bool
		expectedLang   string
		fallbackLang   string
		expectLogEntry bool
	}{
		{
			name: "valid configuration",
			config: config.I18nConfig{
				DefaultLanguage: "en",
				Language:        "en",
				LocalesDir:      tempDir,
			},
			expectError:    false,
			expectedLang:   "en",
			fallbackLang:   "en",
			expectLogEntry: true,
		},
		{
			name: "invalid locale directory",
			config: config.I18nConfig{
				DefaultLanguage: "en",
				Language:        "en",
				LocalesDir:      "/nonexistent/dir",
			},
			expectError:    true,
			expectLogEntry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLog := NewMockLogger()
			translator, err := NewTranslator(&tt.config, mockLog)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if translator == nil {
				t.Fatal("Expected non-nil translator")
			}

			if translator.config.Language != tt.expectedLang {
				t.Errorf("Expected language %s, got %s", tt.expectedLang, translator.config.Language)
			}

			if translator.config.DefaultLanguage != tt.fallbackLang {
				t.Errorf("Expected default language %s, got %s", tt.fallbackLang, translator.config.DefaultLanguage)
			}

			if len(mockLog.debugMessages) == 0 && tt.expectLogEntry {
				t.Error("Expected debug log entries for loaded translation files")
			}
		})
	}
}

func TestTranslate(t *testing.T) {
	// Create a temporary directory for test translation files
	tempDir, err := os.MkdirTemp("", "i18n_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a basic English translation file
	enContent := `{
		"test.message": "This is a test message",
		"test.with_data": "Hello, {{.name}}!"
	}`
	err = os.WriteFile(filepath.Join(tempDir, "en.json"), []byte(enContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test translation file: %v", err)
	}

	mockLog := NewMockLogger()
	translator, err := NewTranslator(&config.I18nConfig{
		DefaultLanguage: "en",
		Language:        "en",
		LocalesDir:      tempDir,
	}, mockLog)

	if err != nil {
		t.Fatalf("Failed to create translator: %v", err)
	}

	tests := []struct {
		name         string
		messageID    string
		templateData map[string]interface{}
		expected     string
	}{
		{
			name:         "existing message",
			messageID:    "test.message",
			templateData: nil,
			expected:     "This is a test message",
		},
		{
			name:      "message with template data",
			messageID: "test.with_data",
			templateData: map[string]interface{}{
				"name": "John",
			},
			expected: "Hello, John!",
		},
		{
			name:         "non-existent message",
			messageID:    "test.nonexistent",
			templateData: nil,
			expected:     "test.nonexistent", // Falls back to message ID
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := translator.Translate(tt.messageID, tt.templateData)

			if result != tt.expected {
				t.Errorf("Expected translation '%s', got '%s'", tt.expected, result)
			}

			// Check if warn log was created for non-existent messages
			if tt.name == "non-existent message" && len(mockLog.warnMessages) == 0 {
				t.Error("Expected warning log for non-existent message")
			}
		})
	}
}

func TestSetLanguage(t *testing.T) {
	// Create a temporary directory for test translation files
	tempDir, err := os.MkdirTemp("", "i18n_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a basic English translation file
	enContent := `{
		"test.message": "This is a test message"
	}`
	err = os.WriteFile(filepath.Join(tempDir, "en.json"), []byte(enContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test translation file: %v", err)
	}

	// Create a basic French translation file
	frContent := `{
		"test.message": "C'est un message de test"
	}`
	err = os.WriteFile(filepath.Join(tempDir, "fr.json"), []byte(frContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test translation file: %v", err)
	}

	mockLog := NewMockLogger()
	translator, err := NewTranslator(&config.I18nConfig{
		DefaultLanguage: "en",
		Language:        "en",
		LocalesDir:      tempDir,
	}, mockLog)

	if err != nil {
		t.Fatalf("Failed to create translator: %v", err)
	}

	// Test initial language
	result := translator.Translate("test.message", nil)
	if result != "This is a test message" {
		t.Errorf("Expected English message, got: %s", result)
	}

	// Change language to French
	translator.SetLanguage("fr")

	// Check config was updated
	if translator.config.Language != "fr" {
		t.Errorf("Expected language to be set to 'fr', got '%s'", translator.config.Language)
	}

	// Check log message was created
	if len(mockLog.infoMessages) == 0 {
		t.Error("Expected info log for language change")
	}

	// Test French translation
	result = translator.Translate("test.message", nil)
	if result != "C'est un message de test" {
		t.Errorf("Expected French message, got: %s", result)
	}
} 