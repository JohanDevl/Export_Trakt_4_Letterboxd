package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// WatchlistMovie represents a movie in the user's watchlist
type WatchlistMovie struct {
	Movie     MovieInfo `json:"movie"`
	ListedAt  string    `json:"listed_at"`
	Notes     string    `json:"notes,omitempty"`
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
