package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

const (
	maxRetries    = 3
	retryInterval = time.Second
)

// MovieIDs represents the various IDs associated with a movie
type MovieIDs struct {
	Trakt int    `json:"trakt"`
	TMDB  int    `json:"tmdb"`
	IMDB  string `json:"imdb"`
	Slug  string `json:"slug"`
}

// MovieInfo represents the basic movie information
type MovieInfo struct {
	Title string   `json:"title"`
	Year  int     `json:"year"`
	IDs   MovieIDs `json:"ids"`
}

// Movie represents a watched movie with its metadata
type Movie struct {
	Movie         MovieInfo `json:"movie"`
	LastWatchedAt string    `json:"last_watched_at"`
}

// CollectionMovie represents a movie in a collection
type CollectionMovie struct {
	Movie       MovieInfo `json:"movie"`
	CollectedAt string    `json:"collected_at"`
}

// Client represents a Trakt API client
type Client struct {
	config     *config.Config
	logger     logger.Logger
	httpClient *http.Client
}

// NewClient creates a new Trakt API client
func NewClient(cfg *config.Config, log logger.Logger) *Client {
	return &Client{
		config: cfg,
		logger: log,
		httpClient: &http.Client{
			Timeout: time.Second * 30,
		},
	}
}

// makeRequest makes an HTTP request with retries
func (c *Client) makeRequest(req *http.Request) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			c.logger.Warn("api.retrying_request", map[string]interface{}{
				"attempt": attempt + 1,
				"max":     maxRetries,
			})
			time.Sleep(retryInterval * time.Duration(attempt))
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		// Only retry on server errors (5xx)
		if resp.StatusCode >= 500 {
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// GetWatchedMovies retrieves the list of watched movies from Trakt
func (c *Client) GetWatchedMovies() ([]Movie, error) {
	req, err := http.NewRequest("GET", c.config.Trakt.APIBaseURL+"/sync/watched/movies", nil)
	if err != nil {
		c.logger.Error("errors.api_request_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("trakt-api-version", "2")
	req.Header.Set("trakt-api-key", c.config.Trakt.ClientID)
	req.Header.Set("Authorization", "Bearer "+c.config.Trakt.AccessToken)

	resp, err := c.makeRequest(req)
	if err != nil {
		c.logger.Error("errors.api_request_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Handle rate limiting
	if limit := resp.Header.Get("X-Ratelimit-Remaining"); limit != "" {
		remaining, _ := strconv.Atoi(limit)
		if remaining < 100 {
			c.logger.Warn("api.rate_limit_warning", map[string]interface{}{
				"remaining": remaining,
			})
		}
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			errorResp = map[string]string{"error": "unknown error"}
		}
		c.logger.Error("errors.api_request_failed", map[string]interface{}{
			"status": resp.StatusCode,
			"error":  errorResp["error"],
		})
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, errorResp["error"])
	}

	// Parse response
	var movies []Movie
	if err := json.NewDecoder(resp.Body).Decode(&movies); err != nil {
		c.logger.Error("errors.api_response_parse_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return movies, nil
}

// GetCollectionMovies retrieves the list of movies in the user's collection from Trakt
func (c *Client) GetCollectionMovies() ([]CollectionMovie, error) {
	req, err := http.NewRequest("GET", c.config.Trakt.APIBaseURL+"/sync/collection/movies", nil)
	if err != nil {
		c.logger.Error("errors.api_request_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("trakt-api-version", "2")
	req.Header.Set("trakt-api-key", c.config.Trakt.ClientID)
	req.Header.Set("Authorization", "Bearer "+c.config.Trakt.AccessToken)

	resp, err := c.makeRequest(req)
	if err != nil {
		c.logger.Error("errors.api_request_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Handle rate limiting
	if limit := resp.Header.Get("X-Ratelimit-Remaining"); limit != "" {
		remaining, _ := strconv.Atoi(limit)
		if remaining < 100 {
			c.logger.Warn("api.rate_limit_warning", map[string]interface{}{
				"remaining": remaining,
			})
		}
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			errorResp = map[string]string{"error": "unknown error"}
		}
		c.logger.Error("errors.api_request_failed", map[string]interface{}{
			"status": resp.StatusCode,
			"error":  errorResp["error"],
		})
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, errorResp["error"])
	}

	// Parse response
	var movies []CollectionMovie
	if err := json.NewDecoder(resp.Body).Decode(&movies); err != nil {
		c.logger.Error("errors.api_response_parse_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	c.logger.Info("api.collection_movies_fetched", map[string]interface{}{
		"count": len(movies),
	})
	return movies, nil
} 