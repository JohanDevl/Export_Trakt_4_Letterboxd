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
	"github.com/stretchr/testify/require"
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
		return
	}

	// Safely check if config is properly set
	if exporter.config == nil {
		t.Error("Expected config to be set, but got nil")
	} else if exporter.config != cfg {
		t.Error("Expected config to match the provided config")
	}

	// Safely check if logger is properly set
	if exporter.log == nil {
		t.Error("Expected logger to be set, but got nil")
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
	expectedHeaders := "Title,Year,WatchedDate,Rating10,IMDb ID,Rewatch"
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

func TestExportShows(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "letterboxd-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test config
	cfg := &config.Config{
		Letterboxd: config.LetterboxdConfig{
			ExportDir:     tempDir,
			ShowsFilename: "test-shows-export.csv",
		},
		Export: config.ExportConfig{
			DateFormat: "2006-01-02",
		},
	}

	// Create a test logger
	log := &MockLogger{}

	// Create an exporter with the test config
	exporter := NewLetterboxdExporter(cfg, log)

	// Create some test data
	testShow := api.WatchedShow{
		Show: api.ShowInfo{
			Title: "Game of Thrones",
			Year:  2011,
			IDs: api.ShowIDs{
				IMDB: "tt0944947",
			},
		},
		Seasons: []api.ShowSeason{
			{
				Number: 1,
				Episodes: []api.EpisodeInfo{
					{
						Number: 1,
						Title:  "Winter Is Coming",
						IDs: api.EpisodeIDs{
							Trakt: 73640,
							TVDB:  3254641,
						},
					},
					{
						Number: 2,
						Title:  "The Kingsroad",
						IDs: api.EpisodeIDs{
							Trakt: 73641,
							TVDB:  3254651,
						},
					},
				},
			},
			{
				Number: 2,
				Episodes: []api.EpisodeInfo{
					{
						Number: 1,
						Title:  "The North Remembers",
						IDs: api.EpisodeIDs{
							Trakt: 73642,
							TVDB:  4077553,
						},
					},
				},
			},
		},
		LastWatchedAt: "2022-01-01T12:00:00Z",
	}

	shows := []api.WatchedShow{testShow}

	// Test the export function
	err = exporter.ExportShows(shows)
	require.NoError(t, err)

	// Verify the file exists
	filePath := filepath.Join(tempDir, "test-shows-export.csv")
	_, err = os.Stat(filePath)
	require.NoError(t, err)

	// Read the file content
	file, err := os.Open(filePath)
	require.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	lines, err := reader.ReadAll()
	require.NoError(t, err)

	// Verify header
	require.Equal(t, []string{"Title", "Year", "Season", "Episode", "EpisodeTitle", "LastWatched", "Rating10", "IMDb ID"}, lines[0])

	// Verify content
	require.Len(t, lines, 4) // header + 3 episodes
	require.Equal(t, "Game of Thrones", lines[1][0])
	require.Equal(t, "2011", lines[1][1])
	require.Equal(t, "1", lines[1][2])
	require.Equal(t, "1", lines[1][3])
	require.Equal(t, "Winter Is Coming", lines[1][4])
	require.Equal(t, "2022-01-01", lines[1][5])

	require.Equal(t, "Game of Thrones", lines[2][0])
	require.Equal(t, "1", lines[2][2])
	require.Equal(t, "2", lines[2][3])
	require.Equal(t, "The Kingsroad", lines[2][4])

	require.Equal(t, "Game of Thrones", lines[3][0])
	require.Equal(t, "2", lines[3][2])
	require.Equal(t, "1", lines[3][3])
	require.Equal(t, "The North Remembers", lines[3][4])
} 