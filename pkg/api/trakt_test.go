package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

// TestNewClient tests client initialization
func TestNewClient(t *testing.T) {
	cfg := &config.Config{
		Trakt: config.TraktConfig{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
			AccessToken:  "test_access_token",
			APIBaseURL:   "https://api.trakt.tv",
		},
	}
	log := &MockLogger{}

	client := NewClient(cfg, log)
	if client == nil {
		t.Error("Expected non-nil client")
	}
	if client.config != cfg {
		t.Error("Expected config to be set")
	}
	if client.logger != log {
		t.Error("Expected logger to be set")
	}
	if client.httpClient == nil {
		t.Error("Expected non-nil HTTP client")
	}
}

// TestGetWatchedMovies tests the GetWatchedMovies endpoint
func TestGetWatchedMovies(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/sync/watched/movies" {
			t.Errorf("Expected path '/sync/watched/movies', got '%s'", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("trakt-api-key") != "test_client_id" {
			t.Errorf("Expected client ID header, got '%s'", r.Header.Get("trakt-api-key"))
		}
		if r.Header.Get("Authorization") != "Bearer test_access_token" {
			t.Errorf("Expected auth header, got '%s'", r.Header.Get("Authorization"))
		}

		// Return test response
		movies := []Movie{
			{
				Movie: MovieInfo{
					Title: "Test Movie",
					Year:  2024,
					IDs: MovieIDs{
						Trakt:  12345,
						TMDB:   67890,
						IMDB:   "tt0123456",
						Slug:   "test-movie-2024",
					},
				},
				LastWatchedAt: time.Now().Format(time.RFC3339),
			},
		}
		json.NewEncoder(w).Encode(movies)
	}))
	defer server.Close()

	// Create client with test server URL
	cfg := &config.Config{
		Trakt: config.TraktConfig{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
			AccessToken:  "test_access_token",
			APIBaseURL:   server.URL,
		},
	}
	log := &MockLogger{}
	client := NewClient(cfg, log)

	// Test successful request
	movies, err := client.GetWatchedMovies()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(movies) != 1 {
		t.Errorf("Expected 1 movie, got %d", len(movies))
	}
	if movies[0].Movie.Title != "Test Movie" {
		t.Errorf("Expected movie title 'Test Movie', got '%s'", movies[0].Movie.Title)
	}
}

// TestGetWatchedMoviesError tests error handling in GetWatchedMovies
func TestGetWatchedMoviesError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid access token",
		})
	}))
	defer server.Close()

	// Create client with test server URL
	cfg := &config.Config{
		Trakt: config.TraktConfig{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
			AccessToken:  "invalid_token",
			APIBaseURL:   server.URL,
		},
	}
	log := &MockLogger{}
	client := NewClient(cfg, log)

	// Test error handling
	movies, err := client.GetWatchedMovies()
	if err == nil {
		t.Error("Expected error but got none")
	}
	if movies != nil {
		t.Error("Expected nil movies on error")
	}
	if log.lastMessage != "errors.api_request_failed" {
		t.Errorf("Expected error message logged, got '%s'", log.lastMessage)
	}
}

// TestRateLimiting tests the rate limiting functionality
func TestRateLimiting(t *testing.T) {
	// Create a test server that returns rate limit headers
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("X-Ratelimit-Limit", "1000")
		w.Header().Set("X-Ratelimit-Remaining", "999")
		w.Header().Set("X-Ratelimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Hour).Unix()))
		json.NewEncoder(w).Encode([]Movie{})
	}))
	defer server.Close()

	// Create client with test server URL
	cfg := &config.Config{
		Trakt: config.TraktConfig{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
			AccessToken:  "test_access_token",
			APIBaseURL:   server.URL,
		},
	}
	log := &MockLogger{}
	client := NewClient(cfg, log)

	// Make multiple requests in quick succession
	for i := 0; i < 3; i++ {
		_, err := client.GetWatchedMovies()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}

	// Verify rate limiting headers were processed
	if requestCount != 3 {
		t.Errorf("Expected 3 requests, got %d", requestCount)
	}
}

// TestRetryMechanism tests the retry mechanism for failed requests
func TestRetryMechanism(t *testing.T) {
	failCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if failCount < 2 {
			failCount++
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode([]Movie{})
	}))
	defer server.Close()

	cfg := &config.Config{
		Trakt: config.TraktConfig{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
			AccessToken:  "test_access_token",
			APIBaseURL:   server.URL,
		},
	}
	log := &MockLogger{}
	client := NewClient(cfg, log)

	// Test that request succeeds after retries
	movies, err := client.GetWatchedMovies()
	if err != nil {
		t.Errorf("Unexpected error after retries: %v", err)
	}
	if movies == nil {
		t.Error("Expected non-nil movies after retries")
	}
	if failCount != 2 {
		t.Errorf("Expected 2 failures before success, got %d", failCount)
	}
}

// TestResponseParsing tests parsing of various response formats
func TestResponseParsing(t *testing.T) {
	testCases := []struct {
		name     string
		response string
		validate func(*testing.T, []Movie)
	}{
		{
			name: "full movie details",
			response: `[{
				"movie": {
					"title": "Test Movie",
					"year": 2024,
					"ids": {
						"trakt": 12345,
						"tmdb": 67890,
						"imdb": "tt0123456",
						"slug": "test-movie-2024"
					}
				},
				"last_watched_at": "2024-03-26T12:00:00Z"
			}]`,
			validate: func(t *testing.T, movies []Movie) {
				if len(movies) != 1 {
					t.Fatalf("Expected 1 movie, got %d", len(movies))
				}
				m := movies[0]
				if m.Movie.Title != "Test Movie" {
					t.Errorf("Expected title 'Test Movie', got '%s'", m.Movie.Title)
				}
				if m.Movie.Year != 2024 {
					t.Errorf("Expected year 2024, got %d", m.Movie.Year)
				}
				if m.Movie.IDs.Trakt != 12345 {
					t.Errorf("Expected Trakt ID 12345, got %d", m.Movie.IDs.Trakt)
				}
			},
		},
		{
			name:     "empty response",
			response: "[]",
			validate: func(t *testing.T, movies []Movie) {
				if len(movies) != 0 {
					t.Errorf("Expected empty movie list, got %d movies", len(movies))
				}
			},
		},
		{
			name: "minimal movie details",
			response: `[{
				"movie": {
					"title": "Test Movie",
					"year": 2024
				},
				"last_watched_at": "2024-03-26T12:00:00Z"
			}]`,
			validate: func(t *testing.T, movies []Movie) {
				if len(movies) != 1 {
					t.Fatalf("Expected 1 movie, got %d", len(movies))
				}
				m := movies[0]
				if m.Movie.Title != "Test Movie" {
					t.Errorf("Expected title 'Test Movie', got '%s'", m.Movie.Title)
				}
				if m.Movie.Year != 2024 {
					t.Errorf("Expected year 2024, got %d", m.Movie.Year)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(tc.response))
			}))
			defer server.Close()

			cfg := &config.Config{
				Trakt: config.TraktConfig{
					ClientID:     "test_client_id",
					ClientSecret: "test_client_secret",
					AccessToken:  "test_access_token",
					APIBaseURL:   server.URL,
				},
			}
			log := &MockLogger{}
			client := NewClient(cfg, log)

			movies, err := client.GetWatchedMovies()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			tc.validate(t, movies)
		})
	}
}

func TestGetCollectionMovies(t *testing.T) {
	// Set up mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/sync/collection/movies", r.URL.Path)

		// Check required headers
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "2", r.Header.Get("trakt-api-version"))
		assert.Equal(t, "test-client-id", r.Header.Get("trakt-api-key"))
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Set rate limiting headers
		w.Header().Set("X-Ratelimit-Remaining", "150")

		// Return mock response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{
				"movie": {
					"title": "The Dark Knight",
					"year": 2008,
					"ids": {
						"trakt": 16,
						"slug": "the-dark-knight-2008",
						"imdb": "tt0468569",
						"tmdb": 155
					}
				},
				"collected_at": "2023-01-15T23:40:30.000Z"
			},
			{
				"movie": {
					"title": "Inception",
					"year": 2010,
					"ids": {
						"trakt": 417,
						"slug": "inception-2010",
						"imdb": "tt1375666",
						"tmdb": 27205
					}
				},
				"collected_at": "2023-03-20T18:25:43.000Z"
			}
		]`))
	}))
	defer mockServer.Close()

	// Create client with mock server URL
	mockConfig := &config.Config{
		Trakt: config.TraktConfig{
			ClientID:    "test-client-id",
			AccessToken: "test-token",
			APIBaseURL:  mockServer.URL,
		},
	}
	mockLogger := &MockLogger{}
	client := NewClient(mockConfig, mockLogger)

	// Call the method to test
	movies, err := client.GetCollectionMovies()

	// Assert no error
	assert.NoError(t, err)

	// Assert movies were correctly parsed
	assert.Equal(t, 2, len(movies))

	// Assert first movie details
	assert.Equal(t, "The Dark Knight", movies[0].Movie.Title)
	assert.Equal(t, 2008, movies[0].Movie.Year)
	assert.Equal(t, 16, movies[0].Movie.IDs.Trakt)
	assert.Equal(t, "tt0468569", movies[0].Movie.IDs.IMDB)
	assert.Equal(t, "2023-01-15T23:40:30.000Z", movies[0].CollectedAt)

	// Assert second movie details
	assert.Equal(t, "Inception", movies[1].Movie.Title)
	assert.Equal(t, 2010, movies[1].Movie.Year)
	assert.Equal(t, 417, movies[1].Movie.IDs.Trakt)
	assert.Equal(t, "tt1375666", movies[1].Movie.IDs.IMDB)
	assert.Equal(t, "2023-03-20T18:25:43.000Z", movies[1].CollectedAt)
}

func TestGetCollectionMoviesError(t *testing.T) {
	// Set up mock server that returns an error
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Invalid OAuth token"}`))
	}))
	defer mockServer.Close()

	// Create client with mock server URL
	mockConfig := &config.Config{
		Trakt: config.TraktConfig{
			ClientID:    "test-client-id",
			AccessToken: "invalid-token",
			APIBaseURL:  mockServer.URL,
		},
	}
	mockLogger := &MockLogger{}
	client := NewClient(mockConfig, mockLogger)

	// Call the method to test
	movies, err := client.GetCollectionMovies()

	// Assert error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed with status 401")
	assert.Nil(t, movies)
}

// TestGetWatchedShows tests the GetWatchedShows endpoint
func TestGetWatchedShows(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/sync/watched/shows" {
			t.Errorf("Expected path '/sync/watched/shows', got '%s'", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("trakt-api-key") != "test_client_id" {
			t.Errorf("Expected client ID header, got '%s'", r.Header.Get("trakt-api-key"))
		}
		if r.Header.Get("Authorization") != "Bearer test_access_token" {
			t.Errorf("Expected auth header, got '%s'", r.Header.Get("Authorization"))
		}

		// Return test response
		shows := []WatchedShow{
			{
				Show: ShowInfo{
					Title: "Test Show",
					Year:  2024,
					IDs: ShowIDs{
						Trakt:  12345,
						TMDB:   67890,
						IMDB:   "tt0123456",
						Slug:   "test-show-2024",
					},
				},
				Seasons: []ShowSeason{
					{
						Number: 1,
						Episodes: []EpisodeInfo{
							{
								Number: 1,
								IDs: EpisodeIDs{
									Trakt: 12345,
								},
							},
						},
					},
				},
				LastWatchedAt: time.Now().Format(time.RFC3339),
			},
		}
		json.NewEncoder(w).Encode(shows)
	}))
	defer server.Close()

	// Create client with test server URL
	cfg := &config.Config{
		Trakt: config.TraktConfig{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
			AccessToken:  "test_access_token",
			APIBaseURL:   server.URL,
		},
	}
	log := &MockLogger{}
	client := NewClient(cfg, log)

	// Test successful request
	shows, err := client.GetWatchedShows()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(shows) != 1 {
		t.Fatalf("Expected 1 show, got %d", len(shows))
	}
	if shows[0].Show.Title != "Test Show" {
		t.Errorf("Expected show title 'Test Show', got '%s'", shows[0].Show.Title)
	}
	if len(shows[0].Seasons) != 1 {
		t.Errorf("Expected 1 season, got %d", len(shows[0].Seasons))
	}

	// Verify the success message was logged
	assert.Equal(t, "api.watched_shows_fetched", log.lastMessage)
	assert.Equal(t, 1, log.lastData["count"])
}

func TestGetWatchedShowsError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid access token",
		})
	}))
	defer server.Close()

	// Create client with test server URL
	cfg := &config.Config{
		Trakt: config.TraktConfig{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
			AccessToken:  "invalid_token",
			APIBaseURL:   server.URL,
		},
	}
	log := &MockLogger{}
	client := NewClient(cfg, log)

	// Test error handling
	shows, err := client.GetWatchedShows()
	if err == nil {
		t.Error("Expected error but got none")
	}
	if shows != nil {
		t.Error("Expected nil shows on error")
	}
	if log.lastMessage != "errors.api_request_failed" {
		t.Errorf("Expected error message logged, got '%s'", log.lastMessage)
	}
}

// TestGetRatings tests the GetRatings endpoint
func TestGetRatings(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/sync/ratings/movies" {
			t.Errorf("Expected path '/sync/ratings/movies', got '%s'", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("trakt-api-key") != "test_client_id" {
			t.Errorf("Expected client ID header, got '%s'", r.Header.Get("trakt-api-key"))
		}
		if r.Header.Get("Authorization") != "Bearer test_access_token" {
			t.Errorf("Expected auth header, got '%s'", r.Header.Get("Authorization"))
		}

		// Return test response
		ratings := []Rating{
			{
				Movie: MovieInfo{
					Title: "Test Movie",
					Year:  2024,
					IDs: MovieIDs{
						Trakt:  12345,
						TMDB:   67890,
						IMDB:   "tt0123456",
						Slug:   "test-movie-2024",
					},
				},
				RatedAt: time.Now().Format(time.RFC3339),
				Rating:  8.5,
			},
		}
		json.NewEncoder(w).Encode(ratings)
	}))
	defer server.Close()

	// Create client with test server URL
	cfg := &config.Config{
		Trakt: config.TraktConfig{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
			AccessToken:  "test_access_token",
			APIBaseURL:   server.URL,
		},
	}
	log := &MockLogger{}
	client := NewClient(cfg, log)

	// Test successful request
	ratings, err := client.GetRatings()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(ratings) != 1 {
		t.Fatalf("Expected 1 rating, got %d", len(ratings))
	}
	if ratings[0].Movie.Title != "Test Movie" {
		t.Errorf("Expected movie title 'Test Movie', got '%s'", ratings[0].Movie.Title)
	}
	if ratings[0].Rating != 8.5 {
		t.Errorf("Expected rating 8.5, got %f", ratings[0].Rating)
	}
}

// TestGetRatingsError tests error handling in GetRatings
func TestGetRatingsError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid access token",
		})
	}))
	defer server.Close()

	// Create client with test server URL
	cfg := &config.Config{
		Trakt: config.TraktConfig{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
			AccessToken:  "invalid_token",
			APIBaseURL:   server.URL,
		},
	}
	log := &MockLogger{}
	client := NewClient(cfg, log)

	// Test error handling
	ratings, err := client.GetRatings()
	if err == nil {
		t.Error("Expected error but got none")
	}
	if ratings != nil {
		t.Error("Expected nil ratings on error")
	}
	if log.lastMessage != "errors.api_request_failed" {
		t.Errorf("Expected error message logged, got '%s'", log.lastMessage)
	}
}

// TestGetWatchlist tests the GetWatchlist endpoint
func TestGetWatchlist(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/sync/watchlist/movies" {
			t.Errorf("Expected path '/sync/watchlist/movies', got '%s'", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("trakt-api-key") != "test_client_id" {
			t.Errorf("Expected client ID header, got '%s'", r.Header.Get("trakt-api-key"))
		}
		if r.Header.Get("Authorization") != "Bearer test_access_token" {
			t.Errorf("Expected auth header, got '%s'", r.Header.Get("Authorization"))
		}

		// Return test response
		watchlist := []WatchlistMovie{
			{
				Movie: MovieInfo{
					Title: "Test Movie",
					Year:  2024,
					IDs: MovieIDs{
						Trakt:  12345,
						TMDB:   67890,
						IMDB:   "tt0123456",
						Slug:   "test-movie-2024",
					},
				},
				ListedAt: time.Now().Format(time.RFC3339),
				Notes:    "Test notes",
			},
		}
		json.NewEncoder(w).Encode(watchlist)
	}))
	defer server.Close()

	// Create client with test server URL
	cfg := &config.Config{
		Trakt: config.TraktConfig{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
			AccessToken:  "test_access_token",
			APIBaseURL:   server.URL,
		},
	}
	log := &MockLogger{}
	client := NewClient(cfg, log)

	// Test successful request
	watchlist, err := client.GetWatchlist()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(watchlist) != 1 {
		t.Fatalf("Expected 1 watchlist item, got %d", len(watchlist))
	}
	if watchlist[0].Movie.Title != "Test Movie" {
		t.Errorf("Expected movie title 'Test Movie', got '%s'", watchlist[0].Movie.Title)
	}
	if watchlist[0].Notes != "Test notes" {
		t.Errorf("Expected notes 'Test notes', got '%s'", watchlist[0].Notes)
	}
}

// TestGetWatchlistError tests error handling in GetWatchlist
func TestGetWatchlistError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid access token",
		})
	}))
	defer server.Close()

	// Create client with test server URL
	cfg := &config.Config{
		Trakt: config.TraktConfig{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
			AccessToken:  "invalid_token",
			APIBaseURL:   server.URL,
		},
	}
	log := &MockLogger{}
	client := NewClient(cfg, log)

	// Test error handling
	watchlist, err := client.GetWatchlist()
	if err == nil {
		t.Error("Expected error but got none")
	}
	if watchlist != nil {
		t.Error("Expected nil watchlist on error")
	}
	if log.lastMessage != "errors.api_request_failed" {
		t.Errorf("Expected error message logged, got '%s'", log.lastMessage)
	}
}

func TestAddExtendedInfo(t *testing.T) {
	// Test with extended info
	cfg := &config.Config{
		Trakt: config.TraktConfig{
			APIBaseURL:   "https://api.trakt.tv",
			ExtendedInfo: "full",
		},
	}
	log := &MockLogger{}
	client := NewClient(cfg, log)

	url := client.addExtendedInfo("https://api.trakt.tv/movies/popular")
	assert.Equal(t, "https://api.trakt.tv/movies/popular?extended=full", url)

	// Test without extended info
	cfg = &config.Config{
		Trakt: config.TraktConfig{
			APIBaseURL:   "https://api.trakt.tv",
			ExtendedInfo: "",
		},
	}
	client = NewClient(cfg, log)

	url = client.addExtendedInfo("https://api.trakt.tv/movies/popular")
	assert.Equal(t, "https://api.trakt.tv/movies/popular", url)

	// Test with invalid URL
	cfg = &config.Config{
		Trakt: config.TraktConfig{
			APIBaseURL:   "https://api.trakt.tv",
			ExtendedInfo: "full",
		},
	}
	client = NewClient(cfg, log)

	url = client.addExtendedInfo("://invalid")
	assert.Equal(t, "://invalid", url)
}

// TestGetShowRatings tests the GetShowRatings endpoint
func TestGetShowRatings(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/sync/ratings/shows" {
			t.Errorf("Expected path '/sync/ratings/shows', got '%s'", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("trakt-api-key") != "test_client_id" {
			t.Errorf("Expected client ID header, got '%s'", r.Header.Get("trakt-api-key"))
		}
		if r.Header.Get("Authorization") != "Bearer test_access_token" {
			t.Errorf("Expected auth header, got '%s'", r.Header.Get("Authorization"))
		}

		// Set rate-limiting header for coverage
		w.Header().Set("X-Ratelimit-Remaining", "50")

		// Return test response
		ratings := []ShowRating{
			{
				Show: ShowInfo{
					Title: "Test Show",
					Year:  2020,
					IDs: ShowIDs{
						Trakt: 12345,
						TMDB:  67890,
						IMDB:  "tt0123456",
						TVDB:  98765,
						Slug:  "test-show-2020",
					},
				},
				RatedAt: time.Now().Format(time.RFC3339),
				Rating:  8.5,
			},
		}
		json.NewEncoder(w).Encode(ratings)
	}))
	defer server.Close()

	// Create client with test server URL
	cfg := &config.Config{
		Trakt: config.TraktConfig{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
			AccessToken:  "test_access_token",
			APIBaseURL:   server.URL,
		},
	}
	log := &MockLogger{}
	client := NewClient(cfg, log)

	// Test successful request
	ratings, err := client.GetShowRatings()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(ratings) != 1 {
		t.Errorf("Expected 1 show rating, got %d", len(ratings))
	}
	if ratings[0].Show.Title != "Test Show" {
		t.Errorf("Expected show title 'Test Show', got '%s'", ratings[0].Show.Title)
	}
	if ratings[0].Rating != 8.5 {
		t.Errorf("Expected rating 8.5, got %f", ratings[0].Rating)
	}
}

// TestGetShowRatingsError tests error handling in GetShowRatings
func TestGetShowRatingsError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid access token",
		})
	}))
	defer server.Close()

	// Create client with test server URL
	cfg := &config.Config{
		Trakt: config.TraktConfig{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
			AccessToken:  "invalid_token",
			APIBaseURL:   server.URL,
		},
	}
	log := &MockLogger{}
	client := NewClient(cfg, log)

	// Test error handling
	ratings, err := client.GetShowRatings()
	if err == nil {
		t.Error("Expected error but got none")
	}
	if ratings != nil {
		t.Errorf("Expected nil ratings, got %v", ratings)
	}
}

// TestGetEpisodeRatings tests the GetEpisodeRatings endpoint
func TestGetEpisodeRatings(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/sync/ratings/episodes" {
			t.Errorf("Expected path '/sync/ratings/episodes', got '%s'", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("trakt-api-key") != "test_client_id" {
			t.Errorf("Expected client ID header, got '%s'", r.Header.Get("trakt-api-key"))
		}
		if r.Header.Get("Authorization") != "Bearer test_access_token" {
			t.Errorf("Expected auth header, got '%s'", r.Header.Get("Authorization"))
		}

		// Set rate-limiting header for coverage
		w.Header().Set("X-Ratelimit-Remaining", "50")

		// Return test response
		ratings := []EpisodeRating{
			{
				Show: ShowInfo{
					Title: "Test Show",
					Year:  2020,
					IDs: ShowIDs{
						Trakt: 12345,
						TMDB:  67890,
						IMDB:  "tt0123456",
						TVDB:  98765,
						Slug:  "test-show-2020",
					},
				},
				Episode: EpisodeInfo{
					Season: 1,
					Number: 2,
					Title:  "Test Episode",
					IDs: EpisodeIDs{
						Trakt: 54321,
						TMDB:  9876,
						TVDB:  1234,
					},
				},
				RatedAt: time.Now().Format(time.RFC3339),
				Rating:  9.0,
			},
		}
		json.NewEncoder(w).Encode(ratings)
	}))
	defer server.Close()

	// Create client with test server URL
	cfg := &config.Config{
		Trakt: config.TraktConfig{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
			AccessToken:  "test_access_token",
			APIBaseURL:   server.URL,
		},
	}
	log := &MockLogger{}
	client := NewClient(cfg, log)

	// Test successful request
	ratings, err := client.GetEpisodeRatings()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(ratings) != 1 {
		t.Errorf("Expected 1 episode rating, got %d", len(ratings))
	}
	if ratings[0].Show.Title != "Test Show" {
		t.Errorf("Expected show title 'Test Show', got '%s'", ratings[0].Show.Title)
	}
	if ratings[0].Episode.Title != "Test Episode" {
		t.Errorf("Expected episode title 'Test Episode', got '%s'", ratings[0].Episode.Title)
	}
	if ratings[0].Rating != 9.0 {
		t.Errorf("Expected rating 9.0, got %f", ratings[0].Rating)
	}
}

// TestGetEpisodeRatingsError tests error handling in GetEpisodeRatings
func TestGetEpisodeRatingsError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid access token",
		})
	}))
	defer server.Close()

	// Create client with test server URL
	cfg := &config.Config{
		Trakt: config.TraktConfig{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
			AccessToken:  "invalid_token",
			APIBaseURL:   server.URL,
		},
	}
	log := &MockLogger{}
	client := NewClient(cfg, log)

	// Test error handling
	ratings, err := client.GetEpisodeRatings()
	if err == nil {
		t.Error("Expected error but got none")
	}
	if ratings != nil {
		t.Errorf("Expected nil ratings, got %v", ratings)
	}
}

// TestGetConfig tests the GetConfig method
func TestGetConfig(t *testing.T) {
	// Create client with test config
	expectedConfig := &config.Config{
		Trakt: config.TraktConfig{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
			AccessToken:  "test_access_token",
			APIBaseURL:   "https://api.trakt.tv",
		},
	}
	log := &MockLogger{}
	client := NewClient(expectedConfig, log)

	// Test GetConfig
	returnedConfig := client.GetConfig()
	if returnedConfig != expectedConfig {
		t.Errorf("Expected config to be the same as what was passed in")
	}
} 