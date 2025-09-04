package testutils

import (
	"time"
)

// Sample API response data for testing

// SampleTraktMovie provides a sample movie from Trakt API
func SampleTraktMovie() map[string]interface{} {
	return map[string]interface{}{
		"title": "Inception",
		"year":  2010,
		"ids": map[string]interface{}{
			"trakt": 1390,
			"slug":  "inception-2010",
			"imdb":  "tt1375666",
			"tmdb":  27205,
		},
	}
}

// SampleTraktWatchedMovie provides a sample watched movie entry
func SampleTraktWatchedMovie() map[string]interface{} {
	return map[string]interface{}{
		"plays": 1,
		"last_watched_at": "2023-07-15T20:30:00.000Z",
		"last_updated_at": "2023-07-15T20:30:00.000Z",
		"movie": SampleTraktMovie(),
	}
}

// SampleTraktRatedMovie provides a sample rated movie entry
func SampleTraktRatedMovie() map[string]interface{} {
	return map[string]interface{}{
		"rating":   8,
		"rated_at": "2023-07-15T20:30:00.000Z",
		"movie":    SampleTraktMovie(),
	}
}

// SampleTraktWatchlistMovie provides a sample watchlist movie entry
func SampleTraktWatchlistMovie() map[string]interface{} {
	return map[string]interface{}{
		"listed_at": "2023-07-15T20:30:00.000Z",
		"movie":     SampleTraktMovie(),
	}
}

// SampleTraktHistoryEntry provides a sample history entry
func SampleTraktHistoryEntry() map[string]interface{} {
	return map[string]interface{}{
		"id":         12345,
		"watched_at": "2023-07-15T20:30:00.000Z",
		"action":     "watch",
		"type":       "movie",
		"movie":      SampleTraktMovie(),
	}
}

// SampleTraktMovies provides multiple sample movies
func SampleTraktMovies() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"title": "Inception",
			"year":  2010,
			"ids": map[string]interface{}{
				"trakt": 1390,
				"imdb":  "tt1375666",
				"tmdb":  27205,
			},
		},
		{
			"title": "The Dark Knight",
			"year":  2008,
			"ids": map[string]interface{}{
				"trakt": 1,
				"imdb":  "tt0468569",
				"tmdb":  155,
			},
		},
		{
			"title": "Interstellar",
			"year":  2014,
			"ids": map[string]interface{}{
				"trakt": 79672,
				"imdb":  "tt0816692",
				"tmdb":  157336,
			},
		},
	}
}

// SampleWatchedMovies provides multiple sample watched movies
func SampleWatchedMovies() []map[string]interface{} {
	movies := SampleTraktMovies()
	watched := make([]map[string]interface{}, len(movies))
	
	for i, movie := range movies {
		watched[i] = map[string]interface{}{
			"plays":            i + 1, // Varying play counts
			"last_watched_at":  time.Now().Add(-time.Duration(i)*24*time.Hour).Format(time.RFC3339),
			"last_updated_at":  time.Now().Add(-time.Duration(i)*24*time.Hour).Format(time.RFC3339),
			"movie":            movie,
		}
	}
	
	return watched
}

// SampleHistoryEntries provides multiple sample history entries
func SampleHistoryEntries() []map[string]interface{} {
	movies := SampleTraktMovies()
	history := make([]map[string]interface{}, 0)
	
	for i, movie := range movies {
		// Add multiple watch events for same movie to test rewatch logic
		for j := 0; j < i+1; j++ {
			history = append(history, map[string]interface{}{
				"id":         (i+1)*1000 + j,
				"watched_at": time.Now().Add(-time.Duration(i*7+j)*24*time.Hour).Format(time.RFC3339),
				"action":     "watch",
				"type":       "movie",
				"movie":      movie,
			})
		}
	}
	
	return history
}

// CSV Sample Data

// SampleCSVHeaders provides standard Letterboxd CSV headers
func SampleCSVHeaders() []string {
	return []string{"Title", "Year", "WatchedDate", "Rating10", "imdbID", "tmdbID", "Rewatch"}
}

// SampleCSVRow provides a sample CSV row
func SampleCSVRow() []string {
	return []string{"Inception", "2010", "2023-07-15", "9", "tt1375666", "27205", "false"}
}

// SampleCSVRows provides multiple sample CSV rows
func SampleCSVRows() [][]string {
	return [][]string{
		{"Inception", "2010", "2023-07-15", "9", "tt1375666", "27205", "false"},
		{"The Dark Knight", "2008", "2023-07-14", "10", "tt0468569", "155", "false"},
		{"Interstellar", "2014", "2023-07-13", "8", "tt0816692", "157336", "true"},
	}
}

// Configuration Sample Data

// SampleOAuthTokenResponse provides a sample OAuth token response
func SampleOAuthTokenResponse() map[string]interface{} {
	return map[string]interface{}{
		"access_token":  "sample_access_token_12345",
		"token_type":    "Bearer",
		"expires_in":    7776000,
		"refresh_token": "sample_refresh_token_12345",
		"scope":         "public",
		"created_at":    time.Now().Unix(),
	}
}

// SampleAPIErrorResponse provides a sample API error response
func SampleAPIErrorResponse() map[string]interface{} {
	return map[string]interface{}{
		"error":             "invalid_grant",
		"error_description": "The provided authorization grant is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client.",
	}
}

// Performance Test Data

// SamplePerformanceMetrics provides sample performance metrics
func SamplePerformanceMetrics() map[string]interface{} {
	return map[string]interface{}{
		"api_calls_total":       100,
		"api_calls_success":     95,
		"api_calls_error":       5,
		"api_success_rate":      95.0,
		"avg_response_time_ms":  250,
		"items_processed":       500,
		"processing_throughput": 10.5,
		"cache_hits":           75,
		"cache_misses":         25,
		"cache_hit_ratio":      75.0,
		"memory_usage_mb":      128,
		"goroutines":           12,
	}
}

// Time Sample Data

// SampleTimeRange provides a sample time range for testing
func SampleTimeRange() (time.Time, time.Time) {
	end := time.Now()
	start := end.Add(-30 * 24 * time.Hour) // 30 days ago
	return start, end
}

// SampleDurations provides sample durations for testing
func SampleDurations() []time.Duration {
	return []time.Duration{
		50 * time.Millisecond,
		100 * time.Millisecond,
		200 * time.Millisecond,
		500 * time.Millisecond,
		1 * time.Second,
	}
}

// Error Sample Data

// SampleErrors provides sample errors for testing
func SampleErrors() []error {
	return []error{
		NewTestError("API_ERROR", "API request failed"),
		NewTestError("NETWORK_ERROR", "Network connection failed"),
		NewTestError("AUTH_ERROR", "Authentication failed"),
		NewTestError("RATE_LIMIT_ERROR", "Rate limit exceeded"),
	}
}

// TestError provides a simple error implementation for testing
type TestError struct {
	Code    string
	Message string
}

func NewTestError(code, message string) *TestError {
	return &TestError{
		Code:    code,
		Message: message,
	}
}

func (e *TestError) Error() string {
	return e.Message
}

func (e *TestError) GetCode() string {
	return e.Code
}

// Web Test Data

// SampleHTTPHeaders provides sample HTTP headers
func SampleHTTPHeaders() map[string]string {
	return map[string]string{
		"Content-Type":    "application/json",
		"Authorization":   "Bearer sample_token",
		"User-Agent":      "Export_Trakt_4_Letterboxd/1.0",
		"Accept":          "application/json",
		"Accept-Encoding": "gzip, deflate",
	}
}

// SampleFormData provides sample form data
func SampleFormData() map[string]string {
	return map[string]string{
		"client_id":     "test_client_id",
		"client_secret": "test_client_secret",
		"grant_type":    "authorization_code",
		"code":          "test_auth_code",
		"redirect_uri":  "http://localhost:8080/callback",
	}
}

// Cache Test Data

// SampleCacheData provides sample data for cache testing
func SampleCacheData() map[string]interface{} {
	return map[string]interface{}{
		"api_response_1": `{"title":"Inception","year":2010}`,
		"api_response_2": `{"title":"The Dark Knight","year":2008}`,
		"api_response_3": `{"title":"Interstellar","year":2014}`,
		"user_profile":   `{"username":"testuser","private":false}`,
		"movie_details":  `{"tagline":"Dreams feel real while we're in them."}`,
	}
}

// Security Test Data

// SampleEncryptionKey provides a sample encryption key for testing
func SampleEncryptionKey() string {
	return "test_encryption_key_32_characters"
}

// SampleSecretData provides sample sensitive data
func SampleSecretData() map[string]string {
	return map[string]string{
		"password":      "super_secret_password",
		"api_key":       "sk_test_1234567890abcdef",
		"private_token": "private_token_abcdef123456",
	}
}