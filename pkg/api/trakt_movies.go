package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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

// Rating represents a user rating for movies
type Rating struct {
	Movie      MovieInfo `json:"movie"`
	RatedAt    string    `json:"rated_at"`
	Rating     float64   `json:"rating"`
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
