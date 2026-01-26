package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

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
