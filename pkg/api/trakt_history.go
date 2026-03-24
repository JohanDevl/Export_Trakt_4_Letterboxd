package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// HistoryItem represents a single watch event from the user's watch history
type HistoryItem struct {
	ID        int       `json:"id"`
	WatchedAt string    `json:"watched_at"`
	Action    string    `json:"action"`
	Type      string    `json:"type"`
	Movie     MovieInfo `json:"movie,omitempty"`
	Show      ShowInfo  `json:"show,omitempty"`
}

// MovieHistoryResponse represents the paginated response from the history API
type MovieHistoryResponse struct {
	Items []HistoryItem `json:"items,omitempty"`
}

// GetMovieHistory retrieves the user's complete movie watch history from Trakt
func (c *Client) GetMovieHistory() ([]HistoryItem, error) {
	var allHistory []HistoryItem
	page := 1
	limit := 100

	for {
		endpoint := fmt.Sprintf("%s/sync/history/movies?page=%d&limit=%d",
			c.config.Trakt.APIBaseURL, page, limit)
		endpoint = c.addExtendedInfo(endpoint)

		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			c.logger.Error("errors.api_request_failed", map[string]interface{}{
				"error": err.Error(),
			})
			return nil, fmt.Errorf("failed to create history request: %w", err)
		}

		resp, err := c.makeRequest(req)
		if err != nil {
			c.logger.Error("errors.api_request_failed", map[string]interface{}{
				"error": err.Error(),
			})
			return nil, fmt.Errorf("failed to execute history request: %w", err)
		}
		defer resp.Body.Close()

		// Handle rate limiting
		if limitHeader := resp.Header.Get("X-Ratelimit-Remaining"); limitHeader != "" {
			remaining, _ := strconv.Atoi(limitHeader)
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

		// Parse response - Trakt history API returns array directly, not wrapped
		var pageHistory []HistoryItem
		if err := json.NewDecoder(resp.Body).Decode(&pageHistory); err != nil {
			c.logger.Error("errors.api_response_parse_failed", map[string]interface{}{
				"error": err.Error(),
			})
			return nil, fmt.Errorf("failed to parse history response: %w", err)
		}

		// Filter for completed viewing actions (watch, scrobble) - exclude checkin
		var watchHistory []HistoryItem
		for _, item := range pageHistory {
			if item.Action == "watch" || item.Action == "scrobble" {
				watchHistory = append(watchHistory, item)
			}
		}

		allHistory = append(allHistory, watchHistory...)

		// Check if we have more pages
		if len(pageHistory) < limit {
			break // No more pages
		}

		page++

		// Safety check to prevent infinite loops
		if page > 1000 {
			c.logger.Warn("api.history_pagination_limit_reached", map[string]interface{}{
				"page": page,
			})
			break
		}
	}

	c.logger.Info("api.movie_history_fetched", map[string]interface{}{
		"count": len(allHistory),
		"pages": page,
	})
	return allHistory, nil
}
