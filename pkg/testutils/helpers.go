package testutils

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
)

// TempDir creates a temporary directory for tests and returns a cleanup function
func TempDir(t *testing.T) (string, func()) {
	t.Helper()
	
	dir, err := os.MkdirTemp("", "export_trakt_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	
	cleanup := func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("Failed to remove temp dir %s: %v", dir, err)
		}
	}
	
	return dir, cleanup
}

// TempFile creates a temporary file for tests and returns the path and cleanup function
func TempFile(t *testing.T, content string) (string, func()) {
	t.Helper()
	
	file, err := os.CreateTemp("", "export_trakt_test_*.tmp")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	
	if content != "" {
		if _, err := file.WriteString(content); err != nil {
			file.Close()
			os.Remove(file.Name())
			t.Fatalf("Failed to write temp file content: %v", err)
		}
	}
	
	file.Close()
	
	cleanup := func() {
		if err := os.Remove(file.Name()); err != nil {
			t.Errorf("Failed to remove temp file %s: %v", file.Name(), err)
		}
	}
	
	return file.Name(), cleanup
}

// SetEnv sets an environment variable for the duration of a test
func SetEnv(t *testing.T, key, value string) func() {
	t.Helper()
	
	oldValue, exists := os.LookupEnv(key)
	
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("Failed to set env var %s: %v", key, err)
	}
	
	return func() {
		if exists {
			os.Setenv(key, oldValue)
		} else {
			os.Unsetenv(key)
		}
	}
}

// UnsetEnv temporarily unsets an environment variable for the duration of a test
func UnsetEnv(t *testing.T, key string) func() {
	t.Helper()
	
	oldValue, exists := os.LookupEnv(key)
	os.Unsetenv(key)
	
	return func() {
		if exists {
			os.Setenv(key, oldValue)
		}
	}
}

// TestConfig creates a basic configuration for testing
func TestConfig() *config.Config {
	return &config.Config{
		Trakt: config.TraktConfig{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
			APIBaseURL:   "https://api.trakt.tv",
			ExtendedInfo: "full",
		},
		Letterboxd: config.LetterboxdConfig{
			ExportDir:                "./test_exports",
			WatchedFilename:          "watched.csv",
			CollectionFilename:       "collection.csv",
			ShowsFilename:            "shows.csv",
			RatingsFilename:          "ratings.csv",
			WatchlistFilename:        "watchlist.csv",
			LetterboxdImportFilename: "letterboxd_import.csv",
		},
		Export: config.ExportConfig{
			Format:      "csv",
			DateFormat:  "2006-01-02",
			Timezone:    "UTC",
			HistoryMode: "aggregated",
		},
		Logging: config.LoggingConfig{
			Level: "info",
			File:  "",
		},
		I18n: config.I18nConfig{
			DefaultLanguage: "en",
			Language:        "en",
			LocalesDir:     "./locales",
		},
		Auth: config.AuthConfig{
			RedirectURI:    "http://localhost:8080/callback",
			CallbackPort:   8080,
			UseOAuth:       true,
			AutoRefresh:    true,
		},
	}
}

// TestConfigWithExportDir creates a test configuration with a specific export directory
func TestConfigWithExportDir(exportDir string) *config.Config {
	cfg := TestConfig()
	cfg.Letterboxd.ExportDir = exportDir
	return cfg
}

// TestConfigMinimal creates a minimal configuration for testing
func TestConfigMinimal() *config.Config {
	return &config.Config{
		Trakt: config.TraktConfig{
			ClientID:   "test_id",
			APIBaseURL: "https://api.trakt.tv",
		},
		Letterboxd: config.LetterboxdConfig{
			ExportDir: "./exports",
		},
	}
}

// WithTestTimeout wraps a test function with a timeout
func WithTestTimeout(t *testing.T, timeout time.Duration, testFunc func()) {
	t.Helper()
	
	done := make(chan bool)
	
	go func() {
		testFunc()
		done <- true
	}()
	
	select {
	case <-done:
		// Test completed successfully
	case <-time.After(timeout):
		t.Fatalf("Test timed out after %v", timeout)
	}
}

// AssertEventually asserts that a condition becomes true within a timeout
func AssertEventually(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	t.Helper()
	
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	
	timeoutCh := time.After(timeout)
	
	for {
		select {
		case <-ticker.C:
			if condition() {
				return
			}
		case <-timeoutCh:
			t.Fatalf("Condition never became true within %v: %s", timeout, message)
		}
	}
}

// CreateTestCSV creates a test CSV file with specified content
func CreateTestCSV(t *testing.T, dir, filename string, headers []string, rows [][]string) string {
	t.Helper()
	
	path := filepath.Join(dir, filename)
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create test CSV file: %v", err)
	}
	defer file.Close()
	
	// Write headers
	if len(headers) > 0 {
		headerLine := ""
		for i, header := range headers {
			if i > 0 {
				headerLine += ","
			}
			headerLine += header
		}
		file.WriteString(headerLine + "\n")
	}
	
	// Write rows
	for _, row := range rows {
		rowLine := ""
		for i, cell := range row {
			if i > 0 {
				rowLine += ","
			}
			rowLine += cell
		}
		file.WriteString(rowLine + "\n")
	}
	
	return path
}

// SkipIfShort skips a test if running in short mode
func SkipIfShort(t *testing.T, reason string) {
	t.Helper()
	if testing.Short() {
		t.Skipf("Skipping test in short mode: %s", reason)
	}
}

// RequireEnv requires an environment variable to be set for a test
func RequireEnv(t *testing.T, key string) string {
	t.Helper()
	value := os.Getenv(key)
	if value == "" {
		t.Skipf("Environment variable %s is required for this test", key)
	}
	return value
}

// Parallel marks a test as safe to run in parallel
func Parallel(t *testing.T) {
	t.Helper()
	if !testing.Short() {
		t.Parallel()
	}
}

// Eventually runs a function repeatedly until it succeeds or timeout
func Eventually(condition func() error, timeout time.Duration, interval time.Duration) error {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		if err := condition(); err == nil {
			return nil
		}
		time.Sleep(interval)
	}
	
	// Final attempt
	return condition()
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// DirExists checks if a directory exists
func DirExists(path string) bool {
	info, err := os.Stat(path)
	return !os.IsNotExist(err) && info.IsDir()
}

// ReadFile reads a file and returns its content as string
func ReadFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}
	return string(content)
}

// WriteFile writes content to a file
func WriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file %s: %v", path, err)
	}
}