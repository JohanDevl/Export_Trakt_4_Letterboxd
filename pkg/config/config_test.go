package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test cases
	tests := []struct {
		name        string
		configData  string
		expectError bool
		validate    func(*testing.T, *Config)
	}{
		{
			name: "valid config",
			configData: `
[trakt]
client_id = "test_client_id"
client_secret = "test_client_secret"
access_token = "test_access_token"
api_base_url = "https://api.trakt.tv"

[letterboxd]
export_dir = "exports"

[export]
format = "csv"
date_format = "2006-01-02"

[logging]
level = "info"
file = "logs/export.log"

[i18n]
default_language = "en"
language = "en"
locales_dir = "locales"
`,
			expectError: false,
			validate: func(t *testing.T, cfg *Config) {
				// Validate Trakt config
				if cfg.Trakt.ClientID != "test_client_id" {
					t.Errorf("Expected ClientID 'test_client_id', got '%s'", cfg.Trakt.ClientID)
				}
				if cfg.Trakt.APIBaseURL != "https://api.trakt.tv" {
					t.Errorf("Expected APIBaseURL 'https://api.trakt.tv', got '%s'", cfg.Trakt.APIBaseURL)
				}

				// Validate Letterboxd config
				if cfg.Letterboxd.ExportDir != "exports" {
					t.Errorf("Expected ExportDir 'exports', got '%s'", cfg.Letterboxd.ExportDir)
				}

				// Validate Export config
				if cfg.Export.Format != "csv" {
					t.Errorf("Expected Format 'csv', got '%s'", cfg.Export.Format)
				}
				if cfg.Export.DateFormat != "2006-01-02" {
					t.Errorf("Expected DateFormat '2006-01-02', got '%s'", cfg.Export.DateFormat)
				}

				// Validate Logging config
				if cfg.Logging.Level != "info" {
					t.Errorf("Expected Level 'info', got '%s'", cfg.Logging.Level)
				}
				if cfg.Logging.File != "logs/export.log" {
					t.Errorf("Expected File 'logs/export.log', got '%s'", cfg.Logging.File)
				}

				// Validate I18n config
				if cfg.I18n.DefaultLanguage != "en" {
					t.Errorf("Expected DefaultLanguage 'en', got '%s'", cfg.I18n.DefaultLanguage)
				}
				if cfg.I18n.Language != "en" {
					t.Errorf("Expected Language 'en', got '%s'", cfg.I18n.Language)
				}
				if cfg.I18n.LocalesDir != "locales" {
					t.Errorf("Expected LocalesDir 'locales', got '%s'", cfg.I18n.LocalesDir)
				}
			},
		},
		{
			name: "missing required fields",
			configData: `
[trakt]
api_base_url = "https://api.trakt.tv"
`,
			expectError: true, // Now we expect an error due to missing required fields
		},
		{
			name:        "invalid TOML",
			configData:  "invalid = ] TOML",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary config file
			configPath := filepath.Join(tmpDir, "config.toml")
			if err := os.WriteFile(configPath, []byte(tt.configData), 0644); err != nil {
				t.Fatalf("Failed to write test config file: %v", err)
			}

			// Load the config
			cfg, err := LoadConfig(configPath)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Run validation if provided and no error occurred
			if !tt.expectError && err == nil && tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("nonexistent.toml")
	if err == nil {
		t.Error("Expected error when loading nonexistent file, got nil")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "missing trakt api base url",
			config: Config{
				Trakt: TraktConfig{},
				Letterboxd: LetterboxdConfig{
					ExportDir: "exports",
				},
				Export: ExportConfig{
					Format:     "csv",
					DateFormat: "2006-01-02",
				},
				Logging: LoggingConfig{
					Level: "info",
				},
				I18n: I18nConfig{
					DefaultLanguage: "en",
					Language:       "en",
					LocalesDir:    "locales",
				},
			},
			expectError: true,
			errorMsg:    "trakt config: api_base_url is required",
		},
		{
			name: "invalid log level",
			config: Config{
				Trakt: TraktConfig{
					APIBaseURL: "https://api.trakt.tv",
				},
				Letterboxd: LetterboxdConfig{
					ExportDir: "exports",
				},
				Export: ExportConfig{
					Format:     "csv",
					DateFormat: "2006-01-02",
				},
				Logging: LoggingConfig{
					Level: "invalid",
				},
				I18n: I18nConfig{
					DefaultLanguage: "en",
					Language:       "en",
					LocalesDir:    "locales",
				},
			},
			expectError: true,
			errorMsg:    "logging config: invalid log level: invalid",
		},
		{
			name: "missing i18n language",
			config: Config{
				Trakt: TraktConfig{
					APIBaseURL: "https://api.trakt.tv",
				},
				Letterboxd: LetterboxdConfig{
					ExportDir: "exports",
				},
				Export: ExportConfig{
					Format:     "csv",
					DateFormat: "2006-01-02",
				},
				Logging: LoggingConfig{
					Level: "info",
				},
				I18n: I18nConfig{
					DefaultLanguage: "en",
					LocalesDir:    "locales",
				},
			},
			expectError: true,
			errorMsg:    "i18n config: language is required",
		},
		{
			name: "valid config",
			config: Config{
				Trakt: TraktConfig{
					APIBaseURL: "https://api.trakt.tv",
				},
				Letterboxd: LetterboxdConfig{
					ExportDir: "exports",
				},
				Export: ExportConfig{
					Format:     "csv",
					DateFormat: "2006-01-02",
				},
				Logging: LoggingConfig{
					Level: "info",
				},
				I18n: I18nConfig{
					DefaultLanguage: "en",
					Language:       "en",
					LocalesDir:    "locales",
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
} 