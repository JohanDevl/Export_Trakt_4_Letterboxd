package api

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

const (
	maxRetries       = 3
	retryInterval    = time.Second
	APIVersion       = "2"
	AuthHeaderPrefix = "Bearer "
)

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
	req.Header.Set("trakt-api-version", APIVersion)
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

	req.Header.Set("Authorization", AuthHeaderPrefix+accessToken)
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


// GetConfig returns the client's configuration
func (c *Client) GetConfig() *config.Config {
	return c.config
} 