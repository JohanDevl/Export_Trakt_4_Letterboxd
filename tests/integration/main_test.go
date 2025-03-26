package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/api"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/export"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/i18n"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

// setupTestEnvironment creates a test environment with config files and directories
func setupTestEnvironment(t *testing.T) (string, *config.Config) {
	// Create temporary test directory
	testDir, err := os.MkdirTemp("", "integration_test")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create config directory
	configDir := filepath.Join(testDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Create export directory
	exportDir := filepath.Join(testDir, "exports")
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		t.Fatalf("Failed to create export directory: %v", err)
	}

	// Create logs directory
	logsDir := filepath.Join(testDir, "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		t.Fatalf("Failed to create logs directory: %v", err)
	}

	// Create locales directory
	localesDir := filepath.Join(testDir, "locales")
	if err := os.MkdirAll(localesDir, 0755); err != nil {
		t.Fatalf("Failed to create locales directory: %v", err)
	}

	// Create test config file
	configFile := filepath.Join(configDir, "config.toml")
	configContent := `
# Trakt.tv API Configuration
[trakt]
client_id = "test_client_id"
client_secret = "test_client_secret"
access_token = "test_access_token"
api_base_url = "fake_api_url"

# Letterboxd Export Configuration
[letterboxd]
export_dir = "` + exportDir + `"

# Export Settings
[export]
format = "csv"
date_format = "2006-01-02"

# Logging Configuration
[logging]
level = "info"
file = "` + filepath.Join(logsDir, "test.log") + `"

# Internationalization Settings
[i18n]
default_language = "en"
language = "en"
locales_dir = "` + localesDir + `"
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create a simple English locale file
	enLocaleFile := filepath.Join(localesDir, "en.json")
	enLocaleContent := `{
  "app": {
    "name": "Export Trakt 4 Letterboxd",
    "description": "Export your Trakt.tv history to Letterboxd format"
  },
  "export": {
    "export_complete": "Successfully exported {{.count}} movies to {{.path}}"
  },
  "errors": {
    "api_request_failed": "API request failed: {{.error}}"
  }
}`
	if err := os.WriteFile(enLocaleFile, []byte(enLocaleContent), 0644); err != nil {
		t.Fatalf("Failed to write locale file: %v", err)
	}

	// Load and parse the configuration
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	return testDir, cfg
}

// cleanup removes the test environment
func cleanup(testDir string) {
	os.RemoveAll(testDir)
}

// MockAPIClient implements a mock Trakt API client
type MockAPIClient struct {
	movies []api.Movie
	err    error
}

func NewMockAPIClient(movies []api.Movie, err error) *MockAPIClient {
	return &MockAPIClient{
		movies: movies,
		err:    err,
	}
}

func (m *MockAPIClient) GetWatchedMovies() ([]api.Movie, error) {
	return m.movies, m.err
}

// TestFullWorkflow tests the complete workflow from config loading to export
func TestFullWorkflow(t *testing.T) {
	// Set up test environment
	testDir, cfg := setupTestEnvironment(t)
	defer cleanup(testDir)

	// Initialize logger
	log := logger.NewLogger()
	log.SetLogLevel(cfg.Logging.Level)
	if err := log.SetLogFile(cfg.Logging.File); err != nil {
		t.Fatalf("Failed to set log file: %v", err)
	}

	// Initialize translator
	translator, err := i18n.NewTranslator(&cfg.I18n, log)
	if err != nil {
		t.Fatalf("Failed to initialize translator: %v", err)
	}
	log.SetTranslator(translator)

	// Create mock API client with test data
	mockMovies := []api.Movie{
		{
			Movie: api.MovieInfo{
				Title: "Integration Test Movie 1",
				Year:  2020,
				IDs: api.MovieIDs{
					IMDB: "tt1234567",
				},
			},
			LastWatchedAt: time.Now().Format(time.RFC3339),
		},
		{
			Movie: api.MovieInfo{
				Title: "Integration Test Movie 2",
				Year:  2021,
				IDs: api.MovieIDs{
					IMDB: "tt2345678",
				},
			},
			LastWatchedAt: time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
		},
	}
	mockClient := NewMockAPIClient(mockMovies, nil)

	// Create Letterboxd exporter
	exporter := export.NewLetterboxdExporter(cfg, log)

	// Run the workflow
	movies, err := mockClient.GetWatchedMovies()
	if err != nil {
		t.Fatalf("Failed to get watched movies: %v", err)
	}

	if len(movies) != 2 {
		t.Errorf("Expected 2 movies, got %d", len(movies))
	}

	// Export movies
	err = exporter.ExportMovies(movies)
	if err != nil {
		t.Fatalf("Failed to export movies: %v", err)
	}

	// Verify export file was created
	files, err := os.ReadDir(cfg.Letterboxd.ExportDir)
	if err != nil {
		t.Fatalf("Failed to read export directory: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("Expected 1 export file, got %d", len(files))
	}

	// Check file content
	exportFilePath := filepath.Join(cfg.Letterboxd.ExportDir, files[0].Name())
	content, err := os.ReadFile(exportFilePath)
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	fileContent := string(content)
	if len(fileContent) == 0 {
		t.Error("Export file is empty")
	}

	// Verify log file was created
	_, err = os.Stat(cfg.Logging.File)
	if os.IsNotExist(err) {
		t.Errorf("Log file was not created")
	}
}

// TestErrorHandling tests how the system handles API errors
func TestErrorHandling(t *testing.T) {
	// Set up test environment
	testDir, cfg := setupTestEnvironment(t)
	defer cleanup(testDir)

	// Initialize logger
	log := logger.NewLogger()
	log.SetLogLevel(cfg.Logging.Level)
	if err := log.SetLogFile(cfg.Logging.File); err != nil {
		t.Fatalf("Failed to set log file: %v", err)
	}

	// Initialize translator
	translator, err := i18n.NewTranslator(&cfg.I18n, log)
	if err != nil {
		t.Fatalf("Failed to initialize translator: %v", err)
	}
	log.SetTranslator(translator)

	// Create mock API client with error
	mockError := fmt.Errorf("API connection error")
	mockClient := NewMockAPIClient(nil, mockError)

	// Run the workflow and check error handling
	movies, err := mockClient.GetWatchedMovies()
	if err == nil {
		t.Error("Expected API error but got none")
	}
	if movies != nil {
		t.Errorf("Expected nil movies on error, got %d movies", len(movies))
	}
} 