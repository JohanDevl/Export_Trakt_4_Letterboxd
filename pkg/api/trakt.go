package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

// TraktClient handles all interactions with the Trakt.tv API
type TraktClient struct {
	config *config.TraktConfig
	client *http.Client
	log    *logger.Logger
}

// NewTraktClient creates a new Trakt API client
func NewTraktClient(cfg *config.TraktConfig, log *logger.Logger) *TraktClient {
	return &TraktClient{
		config: cfg,
		client: &http.Client{
			Timeout: time.Second * 30,
		},
		log: log,
	}
}

// Movie represents a movie from Trakt.tv
type Movie struct {
	Title     string    `json:"title"`
	Year      int       `json:"year"`
	IDs       MovieIDs  `json:"ids"`
	WatchedAt time.Time `json:"watched_at,omitempty"`
	Rating    int       `json:"rating,omitempty"`
}

// MovieIDs contains various IDs for a movie
type MovieIDs struct {
	Trakt  int    `json:"trakt"`
	TMDB   int    `json:"tmdb"`
	IMDB   string `json:"imdb"`
	Slug   string `json:"slug"`
}

// GetWatchedMovies retrieves the user's watched movies
func (c *TraktClient) GetWatchedMovies() ([]Movie, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/sync/history/movies", c.config.APIBaseURL), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("trakt-api-version", "2")
	req.Header.Set("trakt-api-key", c.config.ClientID)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.AccessToken))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var movies []Movie
	if err := json.NewDecoder(resp.Body).Decode(&movies); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return movies, nil
} 