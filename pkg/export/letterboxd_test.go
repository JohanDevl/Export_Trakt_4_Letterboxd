package export

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/api"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/stretchr/testify/assert"
)

// MockLogger implements the logger.Logger interface for testing
type MockLogger struct {
	lastMessage string
	lastData    map[string]interface{}
}

func (m *MockLogger) Info(messageID string, data ...map[string]interface{}) {
	m.lastMessage = messageID
	if len(data) > 0 {
		m.lastData = data[0]
	}
}

func (m *MockLogger) Infof(messageID string, data map[string]interface{}) {
	m.lastMessage = messageID
	m.lastData = data
}

func (m *MockLogger) Error(messageID string, data ...map[string]interface{}) {
	m.lastMessage = messageID
	if len(data) > 0 {
		m.lastData = data[0]
	}
}

func (m *MockLogger) Errorf(messageID string, data map[string]interface{}) {
	m.lastMessage = messageID
	m.lastData = data
}

func (m *MockLogger) Debug(messageID string, data ...map[string]interface{}) {
	m.lastMessage = messageID
	if len(data) > 0 {
		m.lastData = data[0]
	}
}

func (m *MockLogger) Debugf(messageID string, data map[string]interface{}) {
	m.lastMessage = messageID
	m.lastData = data
}

func (m *MockLogger) Warn(messageID string, data ...map[string]interface{}) {
	m.lastMessage = messageID
	if len(data) > 0 {
		m.lastData = data[0]
	}
}

func (m *MockLogger) Warnf(messageID string, data map[string]interface{}) {
	m.lastMessage = messageID
	m.lastData = data
}

func (m *MockLogger) SetLogLevel(level string) {
	// No-op for testing
}

func (m *MockLogger) SetLogFile(filePath string) error {
	// No-op for testing
	return nil
}

func (m *MockLogger) SetTranslator(t logger.Translator) {
	// No-op for testing
}

// TestNewLetterboxdExporter tests the creation of a new Letterboxd exporter
func TestNewLetterboxdExporter(t *testing.T) {
	cfg := &config.Config{
		Letterboxd: config.LetterboxdConfig{
			ExportDir: "test_exports",
		},
		Export: config.ExportConfig{
			Format:     "csv",
			DateFormat: "2006-01-02",
		},
	}
	log := &MockLogger{}

	exporter := NewLetterboxdExporter(cfg, log)
	if exporter == nil {
		t.Error("Expected non-nil exporter")
	}
	if exporter.config != cfg {
		t.Error("Expected config to be set")
	}
	// Cannot directly compare interface values, just check it's not nil
	if exporter.log == nil {
		t.Error("Expected logger to be set")
	}
}

// TestExportMovies tests the export of movies to a CSV file
func TestExportMovies(t *testing.T) {
	// Create a temporary directory for test exports
	tmpDir, err := os.MkdirTemp("", "letterboxd_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test configuration
	cfg := &config.Config{
		Letterboxd: config.LetterboxdConfig{
			ExportDir: tmpDir,
		},
		Export: config.ExportConfig{
			Format:     "csv",
			DateFormat: "2006-01-02",
		},
	}
	log := &MockLogger{}

	// Create test movies
	testMovies := []api.Movie{
		{
			Movie: api.MovieInfo{
				Title: "Test Movie 1",
				Year:  2020,
				IDs: api.MovieIDs{
					IMDB: "tt1234567",
				},
			},
			LastWatchedAt: time.Now().Format(time.RFC3339),
		},
		{
			Movie: api.MovieInfo{
				Title: "Test Movie 2",
				Year:  2021,
				IDs: api.MovieIDs{
					IMDB: "tt2345678",
				},
			},
			LastWatchedAt: time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
		},
	}

	// Create exporter and export movies
	exporter := NewLetterboxdExporter(cfg, log)
	err = exporter.ExportMovies(testMovies)
	if err != nil {
		t.Fatalf("Failed to export movies: %v", err)
	}

	// Check if export file was created
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read export directory: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("Expected 1 export file, got %d", len(files))
	}

	// Check file content
	filePath := filepath.Join(tmpDir, files[0].Name())
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	// Verify file content
	fileContent := string(content)
	expectedHeaders := "Title,Year,WatchedDate,Rating,IMDb ID"
	if len(fileContent) == 0 || content[0] == 0 {
		t.Error("Export file is empty")
	}
	if fileContent[:len(expectedHeaders)] != expectedHeaders {
		t.Errorf("Expected headers '%s', got '%s'", expectedHeaders, fileContent[:len(expectedHeaders)])
	}
	for _, movie := range testMovies {
		if !strings.Contains(fileContent, movie.Movie.Title) {
			t.Errorf("Export file does not contain movie title '%s'", movie.Movie.Title)
		}
	}
}

// TestExportMoviesErrorHandling tests error handling in the export process
func TestExportMoviesErrorHandling(t *testing.T) {
	// Test with invalid export directory
	cfg := &config.Config{
		Letterboxd: config.LetterboxdConfig{
			ExportDir: "/nonexistent/directory/that/should/not/exist",
		},
		Export: config.ExportConfig{
			Format:     "csv",
			DateFormat: "2006-01-02",
		},
	}
	log := &MockLogger{}

	exporter := NewLetterboxdExporter(cfg, log)
	err := exporter.ExportMovies([]api.Movie{})
	if err == nil {
		t.Error("Expected error for invalid export directory, got nil")
	}
}

func TestExportCollectionMovies(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "export_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create mock config and logger
	mockConfig := &config.Config{
		Letterboxd: config.LetterboxdConfig{
			ExportDir: tempDir,
		},
		Export: config.ExportConfig{
			DateFormat: "2006-01-02",
		},
	}
	mockLogger := &MockLogger{}

	// Create test movies
	testMovies := []api.CollectionMovie{
		{
			Movie: api.MovieInfo{
				Title: "The Dark Knight",
				Year:  2008,
				IDs: api.MovieIDs{
					Trakt: 16,
					IMDB:  "tt0468569",
					TMDB:  155,
					Slug:  "the-dark-knight-2008",
				},
			},
			CollectedAt: "2023-01-15T23:40:30.000Z",
		},
		{
			Movie: api.MovieInfo{
				Title: "Inception",
				Year:  2010,
				IDs: api.MovieIDs{
					Trakt: 417,
					IMDB:  "tt1375666",
					TMDB:  27205,
					Slug:  "inception-2010",
				},
			},
			CollectedAt: "2023-03-20T18:25:43.000Z",
		},
	}

	// Create exporter and export movies
	exporter := NewLetterboxdExporter(mockConfig, mockLogger)
	err = exporter.ExportCollectionMovies(testMovies)

	// Assert no error
	assert.NoError(t, err)

	// Find the CSV file (it has a timestamp in the name)
	files, err := filepath.Glob(filepath.Join(tempDir, "collection-export-*.csv"))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(files), "Expected 1 export file")

	// Read the CSV file
	file, err := os.Open(files[0])
	assert.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	assert.NoError(t, err)

	// Check the header
	assert.Equal(t, []string{"Title", "Year", "CollectedDate", "IMDb ID"}, records[0])

	// Check movie records
	assert.Equal(t, "The Dark Knight", records[1][0])
	assert.Equal(t, "2008", records[1][1])
	assert.Equal(t, "2023-01-15", records[1][2])
	assert.Equal(t, "tt0468569", records[1][3])

	assert.Equal(t, "Inception", records[2][0])
	assert.Equal(t, "2010", records[2][1])
	assert.Equal(t, "2023-03-20", records[2][2])
	assert.Equal(t, "tt1375666", records[2][3])
} 