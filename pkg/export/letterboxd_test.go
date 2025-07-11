package export

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
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
	
	// Pass nil client for testing (ratings will be skipped)
	err = exporter.ExportMovies(testMovies, nil)
	if err != nil {
		t.Fatalf("Failed to export movies: %v", err)
	}

	// Check for the expected export file with fixed name
	expectedFilePath := filepath.Join(tmpDir, "watched-export-test.csv")
	if _, err := os.Stat(expectedFilePath); os.IsNotExist(err) {
		t.Fatalf("Expected export file not found: %s", expectedFilePath)
	}

	// Check file content
	content, err := os.ReadFile(expectedFilePath)
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	// Verify file content
	fileContent := string(content)
	expectedHeaders := "Title,Year,WatchedDate,Rating10,imdbID,tmdbID,Rewatch"
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
	
	// Pass nil client for testing (ratings will be skipped)
	err := exporter.ExportMovies([]api.Movie{}, nil)
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

	// Check for the expected export file with fixed name
	expectedFilePath := filepath.Join(tempDir, "collection-export-test.csv")
	assert.FileExists(t, expectedFilePath, "Export file should exist")

	// Read the CSV file
	file, err := os.Open(expectedFilePath)
	assert.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	assert.NoError(t, err)

	// Check the header
	assert.Equal(t, []string{"Title", "Year", "CollectedDate", "imdbID", "tmdbID"}, records[0])

	// Check movie records
	assert.Equal(t, "The Dark Knight", records[1][0])
	assert.Equal(t, "2008", records[1][1])
	assert.Equal(t, "2023-01-15", records[1][2])
	assert.Equal(t, "tt0468569", records[1][3])
	assert.Equal(t, "155", records[1][4])

	assert.Equal(t, "Inception", records[2][0])
	assert.Equal(t, "2010", records[2][1])
	assert.Equal(t, "2023-03-20", records[2][2])
	assert.Equal(t, "tt1375666", records[2][3])
	assert.Equal(t, "27205", records[2][4])
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

// TestExportRatings tests the export of movie ratings to a CSV file
func TestExportRatings(t *testing.T) {
	// Create a temporary directory for test exports
	tmpDir, err := os.MkdirTemp("", "ratings_test")
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

	// Create test ratings
	testRatings := []api.Rating{
		{
			Movie: api.MovieInfo{
				Title: "Test Movie 1",
				Year:  2020,
				IDs: api.MovieIDs{
					IMDB: "tt1234567",
				},
			},
			Rating:  8.5,
			RatedAt: time.Now().Format(time.RFC3339),
		},
		{
			Movie: api.MovieInfo{
				Title: "Test Movie 2",
				Year:  2021,
				IDs: api.MovieIDs{
					IMDB: "tt2345678",
				},
			},
			Rating:  7.0,
			RatedAt: time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
		},
	}

	// Create exporter and export ratings
	exporter := NewLetterboxdExporter(cfg, log)
	
	// Update the ExportRatings method in letterboxd.go to use a fixed filename for tests
	// before running this test
	err = exporter.ExportRatings(testRatings)
	if err != nil {
		t.Fatalf("Failed to export ratings: %v", err)
	}

	// Look for the exported file
	files, err := filepath.Glob(filepath.Join(tmpDir, "ratings-export-*.csv"))
	if err != nil {
		t.Fatalf("Failed to find export files: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("No ratings export file found")
	}

	// Read the first found file
	content, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	// Verify file content
	fileContent := string(content)
	expectedHeaders := "Title,Year,Rating10,RatedDate,IMDb ID"
	if len(fileContent) == 0 || content[0] == 0 {
		t.Error("Export file is empty")
	}
	if fileContent[:len(expectedHeaders)] != expectedHeaders {
		t.Errorf("Expected headers '%s', got '%s'", expectedHeaders, fileContent[:len(expectedHeaders)])
	}
	
	// Check that all test ratings' movies are in the file
	for _, rating := range testRatings {
		if !strings.Contains(fileContent, rating.Movie.Title) {
			t.Errorf("Export file does not contain movie title '%s'", rating.Movie.Title)
		}
		
		// Verify rating value is present (as a string)
		ratingStr := strconv.Itoa(int(rating.Rating))
		if !strings.Contains(fileContent, ratingStr) {
			t.Errorf("Export file does not contain rating '%s'", ratingStr)
		}
	}
}

// TestExportWatchlist tests the export of movie watchlist to a CSV file
func TestExportWatchlist(t *testing.T) {
	// Create a temporary directory for test exports
	tmpDir, err := os.MkdirTemp("", "watchlist_test")
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

	// Create test watchlist
	testWatchlist := []api.WatchlistMovie{
		{
			Movie: api.MovieInfo{
				Title: "Future Movie 1",
				Year:  2022,
				IDs: api.MovieIDs{
					IMDB: "tt1234567",
				},
			},
			ListedAt: time.Now().Format(time.RFC3339),
			Notes:    "Must watch",
		},
		{
			Movie: api.MovieInfo{
				Title: "Future Movie 2",
				Year:  2023,
				IDs: api.MovieIDs{
					IMDB: "tt2345678",
				},
			},
			ListedAt: time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			Notes:    "Looks interesting",
		},
	}

	// Create exporter and export watchlist
	exporter := NewLetterboxdExporter(cfg, log)
	err = exporter.ExportWatchlist(testWatchlist)
	if err != nil {
		t.Fatalf("Failed to export watchlist: %v", err)
	}

	// Check for the expected export file with fixed name
	expectedFilePath := filepath.Join(tmpDir, "watchlist-export-test.csv")
	if _, err := os.Stat(expectedFilePath); os.IsNotExist(err) {
		t.Fatalf("Expected export file not found: %s", expectedFilePath)
	}

	// Check file content
	content, err := os.ReadFile(expectedFilePath)
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	// Verify file content
	fileContent := string(content)
	expectedHeaders := "Title,Year,ListedDate,Rating10,IMDb ID"
	if len(fileContent) == 0 || content[0] == 0 {
		t.Error("Export file is empty")
	}
	if fileContent[:len(expectedHeaders)] != expectedHeaders {
		t.Errorf("Expected headers '%s', got '%s'", expectedHeaders, fileContent[:len(expectedHeaders)])
	}
	
	// Check that all watchlist movies are in the file
	for _, item := range testWatchlist {
		if !strings.Contains(fileContent, item.Movie.Title) {
			t.Errorf("Export file does not contain movie title '%s'", item.Movie.Title)
		}
		
		// Check notes if present
		if item.Notes != "" && !strings.Contains(fileContent, item.Notes) {
			t.Errorf("Export file does not contain notes '%s'", item.Notes)
		}
	}
}

// TestExportLetterboxdFormat tests the export to Letterboxd import format
func TestExportLetterboxdFormat(t *testing.T) {
	// Create a temporary directory for test exports
	tmpDir, err := os.MkdirTemp("", "letterboxd_import_test")
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
					TMDB: 1234,
				},
			},
			LastWatchedAt: time.Now().Format(time.RFC3339),
			Plays:         2,
		},
		{
			Movie: api.MovieInfo{
				Title: "Test Movie 2",
				Year:  2021,
				IDs: api.MovieIDs{
					IMDB: "tt2345678",
					TMDB: 5678,
				},
			},
			LastWatchedAt: time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			Plays:         1,
		},
	}

	// Create test ratings
	testRatings := []api.Rating{
		{
			Movie: api.MovieInfo{
				Title: "Test Movie 1",
				Year:  2020,
				IDs: api.MovieIDs{
					IMDB: "tt1234567",
				},
			},
			Rating:  8,
			RatedAt: time.Now().Format(time.RFC3339),
		},
	}

	// Create exporter and export to Letterboxd format
	exporter := NewLetterboxdExporter(cfg, log)
	err = exporter.ExportLetterboxdFormat(testMovies, testRatings)
	if err != nil {
		t.Fatalf("Failed to export in Letterboxd format: %v", err)
	}

	// Check for the expected export file with fixed name
	expectedFilePath := filepath.Join(tmpDir, "letterboxd-import-test.csv")
	if _, err := os.Stat(expectedFilePath); os.IsNotExist(err) {
		t.Fatalf("Expected export file not found: %s", expectedFilePath)
	}

	// Check file content
	content, err := os.ReadFile(expectedFilePath)
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	// Verify file content
	fileContent := string(content)
	expectedHeaders := "Title,Year,imdbID,tmdbID,WatchedDate,Rating10,Rewatch"
	if len(fileContent) == 0 || content[0] == 0 {
		t.Error("Export file is empty")
	}
	if fileContent[:len(expectedHeaders)] != expectedHeaders {
		t.Errorf("Expected headers '%s', got '%s'", expectedHeaders, fileContent[:len(expectedHeaders)])
	}
	
	// Check that all movies are in the file
	for _, movie := range testMovies {
		if !strings.Contains(fileContent, movie.Movie.Title) {
			t.Errorf("Export file does not contain movie title '%s'", movie.Movie.Title)
		}
		
		// Check IMDB ID is included
		if !strings.Contains(fileContent, movie.Movie.IDs.IMDB) {
			t.Errorf("Export file does not contain IMDB ID '%s'", movie.Movie.IDs.IMDB)
		}
		
		// Check TMDB ID is included
		tmdbID := strconv.Itoa(movie.Movie.IDs.TMDB)
		if !strings.Contains(fileContent, tmdbID) {
			t.Errorf("Export file does not contain TMDB ID '%s'", tmdbID)
		}
	}
	
	// Check rating is included
	if !strings.Contains(fileContent, "8") {
		t.Error("Export file does not contain the expected rating")
	}
	
	// Check rewatch indicator for movie with multiple plays
	if !strings.Contains(fileContent, "true") {
		t.Error("Export file does not indicate rewatch properly")
	}
}

// TestGetTimeInConfigTimezone tests the getTimeInConfigTimezone function
func TestGetTimeInConfigTimezone(t *testing.T) {
	// Test cases
	testCases := []struct {
		name     string
		timezone string
		isValid  bool
	}{
		{
			name:     "Default timezone (empty)",
			timezone: "",
			isValid:  true,
		},
		{
			name:     "Valid timezone",
			timezone: "America/New_York",
			isValid:  true,
		},
		{
			name:     "Invalid timezone",
			timezone: "Invalid/Timezone",
			isValid:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock logger and config
			mockLogger := &MockLogger{}
			cfg := &config.Config{
				Export: config.ExportConfig{
					Timezone: tc.timezone,
				},
			}

			// Create exporter
			exporter := NewLetterboxdExporter(cfg, mockLogger)

			// Get time in configured timezone
			result := exporter.getTimeInConfigTimezone()

			// Check the result
			if tc.timezone == "" {
				// For empty timezone, should use UTC
				assert.Equal(t, "export.using_default_timezone", mockLogger.lastMessage)
			} else if tc.isValid {
				// For valid timezone, should use the configured timezone
				assert.Equal(t, "export.using_configured_timezone", mockLogger.lastMessage)
				
				// Verify the timezone data is in the log
				if data, ok := mockLogger.lastData["timezone"]; ok {
					assert.Equal(t, tc.timezone, data)
				} else {
					t.Errorf("Expected timezone in log data")
				}
				
				// Verify the time is formatted correctly
				if timeStr, ok := mockLogger.lastData["time"]; ok {
					_, err := time.Parse(time.RFC3339, timeStr.(string))
					assert.NoError(t, err, "Time should be in RFC3339 format")
				} else {
					t.Errorf("Expected time in log data")
				}
			} else {
				// For invalid timezone, should log a warning and return UTC time
				assert.Equal(t, "export.timezone_load_failed", mockLogger.lastMessage)
				
				// Verify the error message contains the timezone
				if data, ok := mockLogger.lastData["timezone"]; ok {
					assert.Equal(t, tc.timezone, data)
				} else {
					t.Errorf("Expected timezone in log data")
				}
				
				// Verify there's an error message
				assert.Contains(t, mockLogger.lastData, "error")
			}

			// Verify the result is a valid time
			nowUTC := time.Now().UTC()
			timeDiff := result.Sub(nowUTC)
			
			// The time difference should be small (within a few seconds)
			// or match the timezone offset if using a valid timezone
			assert.True(t, timeDiff.Seconds() < 5, "Time difference should be small")
		})
	}
} 