package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
	Title    string   `json:"title"`
	Year     int      `json:"year"`
	IDs      MovieIDs `json:"ids"`
	Tagline  string   `json:"tagline,omitempty"`
	Overview string   `json:"overview,omitempty"`
	Released string   `json:"released,omitempty"`
	Runtime  int      `json:"runtime,omitempty"`
	Country  string   `json:"country,omitempty"`
	Updated  string   `json:"updated_at,omitempty"`
	Trailer  string   `json:"trailer,omitempty"`
	Homepage string   `json:"homepage,omitempty"`
	Rating   float64  `json:"rating,omitempty"`
	Votes    int      `json:"votes,omitempty"`
	Comment  int      `json:"comment_count,omitempty"`
	Genres   []string `json:"genres,omitempty"`
}

// Movie represents a watched movie with its metadata
type Movie struct {
	Movie         MovieInfo `json:"movie"`
	LastWatchedAt string    `json:"last_watched_at"`
	Plays         int       `json:"plays,omitempty"`
}

// CollectionMovie represents a movie in a collection
type CollectionMovie struct {
	Movie       MovieInfo `json:"movie"`
	CollectedAt string    `json:"collected_at"`
}

// ShowIDs represents the various IDs associated with a show
type ShowIDs struct {
	Trakt int    `json:"trakt"`
	TMDB  int    `json:"tmdb"`
	IMDB  string `json:"imdb"`
	Slug  string `json:"slug"`
	TVDB  int    `json:"tvdb"`
}

// ShowInfo represents the basic show information
type ShowInfo struct {
	Title     string   `json:"title"`
	Year      int      `json:"year"`
	IDs       ShowIDs  `json:"ids"`
	Overview  string   `json:"overview,omitempty"`
	FirstAired string  `json:"first_aired,omitempty"`
	Runtime   int      `json:"runtime,omitempty"`
	Network   string   `json:"network,omitempty"`
	Country   string   `json:"country,omitempty"`
	Updated   string   `json:"updated_at,omitempty"`
	Trailer   string   `json:"trailer,omitempty"`
	Homepage  string   `json:"homepage,omitempty"`
	Status    string   `json:"status,omitempty"`
	Rating    float64  `json:"rating,omitempty"`
	Votes     int      `json:"votes,omitempty"`
	Comment   int      `json:"comment_count,omitempty"`
	Genres    []string `json:"genres,omitempty"`
}

// EpisodeIDs represents the various IDs associated with an episode
type EpisodeIDs struct {
	Trakt int `json:"trakt"`
	TMDB  int `json:"tmdb"`
	TVDB  int `json:"tvdb"`
}

// EpisodeInfo represents the basic episode information
type EpisodeInfo struct {
	Season     int        `json:"season"`
	Number     int        `json:"number"`
	Title      string     `json:"title"`
	IDs        EpisodeIDs `json:"ids"`
	Overview   string     `json:"overview,omitempty"`
	FirstAired string     `json:"first_aired,omitempty"`
	Updated    string     `json:"updated_at,omitempty"`
	Rating     float64    `json:"rating,omitempty"`
	Votes      int        `json:"votes,omitempty"`
	Comment    int        `json:"comment_count,omitempty"`
}

// WatchedShow represents a watched show with its metadata
type WatchedShow struct {
	Show          ShowInfo     `json:"show"`
	Seasons       []ShowSeason `json:"seasons"`
	LastWatchedAt string       `json:"last_watched_at"`
	Plays         int          `json:"plays,omitempty"`
}

// ShowSeason represents a season of a show
type ShowSeason struct {
	Number   int            `json:"number"`
	Episodes []EpisodeInfo `json:"episodes"`
}

// Client represents a Trakt API client
type Client struct {
	config       *config.Config
	logger       logger.Logger
	httpClient   *http.Client
	tokenManager TokenManager
}

// TokenManager interface for token management
type TokenManager interface {
	GetValidAccessToken() (string, error)
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

// NewClientWithTokenManager creates a new Trakt API client with token management
func NewClientWithTokenManager(cfg *config.Config, log logger.Logger, tokenMgr TokenManager) *Client {
	return &Client{
		config:       cfg,
		logger:       log,
		httpClient:   &http.Client{Timeout: time.Second * 30},
		tokenManager: tokenMgr,
	}
}

// makeRequest makes an HTTP request with retries and automatic token refresh
func (c *Client) makeRequest(req *http.Request) (*http.Response, error) {
	// Set authentication header
	if err := c.setAuthHeader(req); err != nil {
		return nil, fmt.Errorf("failed to set auth header: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			c.logger.Warn("api.retrying_request", map[string]interface{}{
				"attempt": attempt + 1,
				"max":     maxRetries,
			})
			time.Sleep(retryInterval * time.Duration(attempt))
		}

		// Clone the request for retry attempts
		reqClone := req.Clone(req.Context())
		if err := c.setAuthHeader(reqClone); err != nil {
			lastErr = fmt.Errorf("failed to set auth header on retry: %w", err)
			continue
		}

		resp, err := c.httpClient.Do(reqClone)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		// Handle authentication errors (401) with token refresh
		if resp.StatusCode == http.StatusUnauthorized && c.tokenManager != nil {
			resp.Body.Close()
			c.logger.Info("api.token_expired_refreshing", nil)
			
			// Try to refresh token and retry once
			if _, err := c.tokenManager.GetValidAccessToken(); err != nil {
				return nil, fmt.Errorf("token refresh failed: %w", err)
			}
			
			// Retry the request with new token
			reqRetry := req.Clone(req.Context())
			if err := c.setAuthHeader(reqRetry); err != nil {
				return nil, fmt.Errorf("failed to set refreshed auth header: %w", err)
			}
			
			retryResp, retryErr := c.httpClient.Do(reqRetry)
			if retryErr != nil {
				return nil, fmt.Errorf("retry after token refresh failed: %w", retryErr)
			}
			
			if retryResp.StatusCode == http.StatusUnauthorized {
				retryResp.Body.Close()
				return nil, fmt.Errorf("authentication failed even after token refresh")
			}
			
			c.logger.Info("api.token_refresh_success", nil)
			return retryResp, nil
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

// setAuthHeader sets the authentication header for the request
func (c *Client) setAuthHeader(req *http.Request) error {
	// Set basic headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("trakt-api-version", "2")
	req.Header.Set("trakt-api-key", c.config.Trakt.ClientID)

	// Get access token
	var accessToken string
	if c.tokenManager != nil {
		token, err := c.tokenManager.GetValidAccessToken()
		if err != nil {
			return fmt.Errorf("failed to get valid access token: %w", err)
		}
		accessToken = token
	} else {
		// Fallback to config token
		accessToken = c.config.Trakt.AccessToken
		if accessToken == "" {
			return fmt.Errorf("no access token available")
		}
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	return nil
}

// addExtendedInfo adds the extended parameter to the URL if it's configured
func (c *Client) addExtendedInfo(endpoint string) string {
	// Safety checks
	if c == nil || c.config == nil {
		return endpoint
	}
	
	if c.config.Trakt.ExtendedInfo == "" {
		return endpoint
	}

	baseURL, err := url.Parse(endpoint)
	if err != nil {
		if c.logger != nil {
			c.logger.Warn("api.url_parse_error", map[string]interface{}{
				"error": err.Error(),
			})
		}
		return endpoint
	}

	q := baseURL.Query()
	q.Set("extended", c.config.Trakt.ExtendedInfo)
	baseURL.RawQuery = q.Encode()
	return baseURL.String()
}

// GetWatchedMovies retrieves the list of watched movies from Trakt
func (c *Client) GetWatchedMovies() ([]Movie, error) {
	endpoint := c.addExtendedInfo(c.config.Trakt.APIBaseURL + "/sync/watched/movies")
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		c.logger.Error("errors.api_request_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Headers are now set by makeRequest via setAuthHeader

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
	endpoint := c.addExtendedInfo(c.config.Trakt.APIBaseURL + "/sync/collection/movies")
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		c.logger.Error("errors.api_request_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Headers are now set by makeRequest via setAuthHeader

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

// GetWatchedShows retrieves the list of watched shows from Trakt
func (c *Client) GetWatchedShows() ([]WatchedShow, error) {
	endpoint := c.addExtendedInfo(c.config.Trakt.APIBaseURL + "/sync/watched/shows")
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		c.logger.Error("errors.api_request_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Headers are now set by makeRequest via setAuthHeader

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
	var shows []WatchedShow
	if err := json.NewDecoder(resp.Body).Decode(&shows); err != nil {
		c.logger.Error("errors.api_response_parse_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	c.logger.Info("api.watched_shows_fetched", map[string]interface{}{
		"count": len(shows),
	})
	return shows, nil
}

// Rating represents a user rating for movies
type Rating struct {
	Movie      MovieInfo `json:"movie"`
	RatedAt    string    `json:"rated_at"`
	Rating     float64   `json:"rating"`
}

// ShowRating represents a user rating for shows
type ShowRating struct {
	Show      ShowInfo `json:"show"`
	RatedAt   string   `json:"rated_at"`
	Rating    float64  `json:"rating"`
}

// EpisodeRating represents a user rating for episodes
type EpisodeRating struct {
	Show     ShowInfo    `json:"show"`
	Episode  EpisodeInfo `json:"episode"`
	RatedAt  string      `json:"rated_at"`
	Rating   float64     `json:"rating"`
}

// WatchlistMovie represents a movie in the user's watchlist
type WatchlistMovie struct {
	Movie     MovieInfo `json:"movie"`
	ListedAt  string    `json:"listed_at"`
	Notes     string    `json:"notes,omitempty"`
}

// GetRatings retrieves the user's ratings from Trakt
func (c *Client) GetRatings() ([]Rating, error) {
	endpoint := c.addExtendedInfo(c.config.Trakt.APIBaseURL + "/sync/ratings/movies")
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		c.logger.Error("errors.api_request_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Headers are now set by makeRequest via setAuthHeader

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
	var ratings []Rating
	if err := json.NewDecoder(resp.Body).Decode(&ratings); err != nil {
		c.logger.Error("errors.api_response_parse_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	c.logger.Info("api.ratings_fetched", map[string]interface{}{
		"count": len(ratings),
	})
	return ratings, nil
}

// GetWatchlist retrieves the user's movie watchlist from Trakt
func (c *Client) GetWatchlist() ([]WatchlistMovie, error) {
	endpoint := c.addExtendedInfo(c.config.Trakt.APIBaseURL + "/sync/watchlist/movies")
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		c.logger.Error("errors.api_request_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Headers are now set by makeRequest via setAuthHeader

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
	var watchlist []WatchlistMovie
	if err := json.NewDecoder(resp.Body).Decode(&watchlist); err != nil {
		c.logger.Error("errors.api_response_parse_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	c.logger.Info("api.watchlist_fetched", map[string]interface{}{
		"count": len(watchlist),
	})
	return watchlist, nil
}

// GetShowRatings retrieves the user's TV show ratings from Trakt
func (c *Client) GetShowRatings() ([]ShowRating, error) {
	endpoint := c.addExtendedInfo(c.config.Trakt.APIBaseURL + "/sync/ratings/shows")
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		c.logger.Error("errors.api_request_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Headers are now set by makeRequest via setAuthHeader

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
	var ratings []ShowRating
	if err := json.NewDecoder(resp.Body).Decode(&ratings); err != nil {
		c.logger.Error("errors.api_response_parse_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	c.logger.Info("api.show_ratings_fetched", map[string]interface{}{
		"count": len(ratings),
	})
	return ratings, nil
}

// GetEpisodeRatings retrieves the user's TV episode ratings from Trakt
func (c *Client) GetEpisodeRatings() ([]EpisodeRating, error) {
	endpoint := c.addExtendedInfo(c.config.Trakt.APIBaseURL + "/sync/ratings/episodes")
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		c.logger.Error("errors.api_request_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Headers are now set by makeRequest via setAuthHeader

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
	var ratings []EpisodeRating
	if err := json.NewDecoder(resp.Body).Decode(&ratings); err != nil {
		c.logger.Error("errors.api_response_parse_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	c.logger.Info("api.episode_ratings_fetched", map[string]interface{}{
		"count": len(ratings),
	})
	return ratings, nil
}

// GetConfig returns the client's configuration
func (c *Client) GetConfig() *config.Config {
	return c.config
} 