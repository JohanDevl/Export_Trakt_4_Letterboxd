package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/performance/cache"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/performance/metrics"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/performance/pool"
)

// OptimizedClient represents an optimized Trakt API client with caching and concurrency
type OptimizedClient struct {
	config     *config.Config
	logger     logger.Logger
	httpClient *http.Client
	cache      *cache.APIResponseCache
	metrics    *metrics.PerformanceMetrics
	workerPool *pool.WorkerPool
	rateLimiter chan struct{}
	
	// Connection pooling
	transport *http.Transport
}

// OptimizedClientConfig holds configuration for the optimized client
type OptimizedClientConfig struct {
	Config           *config.Config
	Logger           logger.Logger
	CacheConfig      cache.CacheConfig
	WorkerPoolSize   int
	RateLimitPerSec  int
	ConnectionPool   int
	RequestTimeout   time.Duration
}

// NewOptimizedClient creates a new optimized Trakt API client
func NewOptimizedClient(cfg OptimizedClientConfig) *OptimizedClient {
	// Set defaults
	if cfg.WorkerPoolSize <= 0 {
		cfg.WorkerPoolSize = 10
	}
	if cfg.RateLimitPerSec <= 0 {
		cfg.RateLimitPerSec = 100
	}
	if cfg.ConnectionPool <= 0 {
		cfg.ConnectionPool = 20
	}
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 30 * time.Second
	}
	
	// Default cache config
	if cfg.CacheConfig.Capacity <= 0 {
		cfg.CacheConfig.Capacity = 1000
	}
	if cfg.CacheConfig.TTL <= 0 {
		cfg.CacheConfig.TTL = 24 * time.Hour
	}

	// Create HTTP transport with connection pooling
	transport := &http.Transport{
		MaxIdleConns:        cfg.ConnectionPool,
		MaxIdleConnsPerHost: cfg.ConnectionPool / 2,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
	}

	// Create HTTP client
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   cfg.RequestTimeout,
	}

	// Create cache
	apiCache := cache.NewAPIResponseCache(cfg.CacheConfig)

	// Create metrics
	performanceMetrics := metrics.NewPerformanceMetrics(cfg.Logger)

	// Create rate limiter
	rateLimiter := make(chan struct{}, cfg.RateLimitPerSec)

	// Create worker pool
	workerPoolConfig := pool.WorkerPoolConfig{
		Workers:    cfg.WorkerPoolSize,
		BufferSize: cfg.WorkerPoolSize * 2,
		Logger:     cfg.Logger,
		Metrics:    performanceMetrics,
	}
	workerPool := pool.NewWorkerPool(workerPoolConfig)

	client := &OptimizedClient{
		config:      cfg.Config,
		logger:      cfg.Logger,
		httpClient:  httpClient,
		cache:       apiCache,
		metrics:     performanceMetrics,
		workerPool:  workerPool,
		rateLimiter: rateLimiter,
		transport:   transport,
	}

	// Start worker pool
	workerPool.Start()

	// Start rate limiter refill goroutine
	go client.rateLimiterRefill(cfg.RateLimitPerSec)

	return client
}

// rateLimiterRefill refills the rate limiter at the specified rate
func (c *OptimizedClient) rateLimiterRefill(ratePerSec int) {
	interval := time.Second / time.Duration(ratePerSec)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case c.rateLimiter <- struct{}{}:
		default:
			// Channel is full, skip this tick
		}
	}
}

// makeOptimizedRequest makes an HTTP request with caching, rate limiting, and metrics
func (c *OptimizedClient) makeOptimizedRequest(ctx context.Context, endpoint string, result interface{}) error {
	start := time.Now()
	defer func() {
		c.metrics.RecordAPIResponseTime(time.Since(start))
	}()

	c.metrics.IncrementAPICall()

	// Add extended info if configured
	fullEndpoint := c.addExtendedInfo(endpoint)

	// Check cache first
	if c.cache.GetJSON(fullEndpoint, result) {
		c.metrics.IncrementCacheHit()
		c.metrics.IncrementAPISuccess()
		c.logger.Debug("api.cache_hit", map[string]interface{}{
			"endpoint": fullEndpoint,
		})
		return nil
	}

	c.metrics.IncrementCacheMiss()

	// Wait for rate limiter
	select {
	case <-c.rateLimiter:
	case <-ctx.Done():
		return ctx.Err()
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", fullEndpoint, nil)
	if err != nil {
		c.metrics.IncrementAPIError()
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("User-Agent", "Export_Trakt_4_Letterboxd/1.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	
	if c.config.Trakt.ClientID != "" {
		req.Header.Set("trakt-api-key", c.config.Trakt.ClientID)
	}

	// Make request with retries
	resp, err := c.makeRequestWithRetries(ctx, req)
	if err != nil {
		c.metrics.IncrementAPIError()
		return err
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		c.metrics.IncrementAPIError()
		return fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	// Parse response
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(result); err != nil {
		c.metrics.IncrementAPIError()
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Cache the result
	if err := c.cache.SetJSON(fullEndpoint, result); err != nil {
		c.logger.Warn("api.cache_set_failed", map[string]interface{}{
			"endpoint": fullEndpoint,
			"error":    err.Error(),
		})
	}

	c.metrics.IncrementAPISuccess()
	c.logger.Debug("api.request_success", map[string]interface{}{
		"endpoint": fullEndpoint,
		"duration": time.Since(start).String(),
	})

	return nil
}

// makeRequestWithRetries makes an HTTP request with exponential backoff retries
func (c *OptimizedClient) makeRequestWithRetries(ctx context.Context, req *http.Request) (*http.Response, error) {
	maxRetries := 3
	baseDelay := time.Second

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			delay := baseDelay * time.Duration(1<<uint(attempt-1))
			c.logger.Warn("api.retrying_request", map[string]interface{}{
				"attempt": attempt + 1,
				"delay":   delay.String(),
			})
			
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		// Only retry on server errors (5xx) or rate limiting (429)
		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// GetWatchedMoviesConcurrent retrieves watched movies using concurrent processing
func (c *OptimizedClient) GetWatchedMoviesConcurrent(ctx context.Context) ([]Movie, error) {
	var movies []Movie
	err := c.makeOptimizedRequest(ctx, "https://api.trakt.tv/users/me/watched/movies", &movies)
	if err != nil {
		return nil, fmt.Errorf("failed to get watched movies: %w", err)
	}

	c.logger.Info("api.movies_retrieved", map[string]interface{}{
		"count": len(movies),
	})

	return movies, nil
}

// GetCollectionMoviesConcurrent retrieves collection movies using concurrent processing
func (c *OptimizedClient) GetCollectionMoviesConcurrent(ctx context.Context) ([]CollectionMovie, error) {
	var movies []CollectionMovie
	err := c.makeOptimizedRequest(ctx, "https://api.trakt.tv/users/me/collection/movies", &movies)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection movies: %w", err)
	}

	c.logger.Info("api.collection_movies_retrieved", map[string]interface{}{
		"count": len(movies),
	})

	return movies, nil
}

// GetRatingsConcurrent retrieves ratings using concurrent processing
func (c *OptimizedClient) GetRatingsConcurrent(ctx context.Context) ([]Rating, error) {
	var ratings []Rating
	err := c.makeOptimizedRequest(ctx, "https://api.trakt.tv/users/me/ratings/movies", &ratings)
	if err != nil {
		return nil, fmt.Errorf("failed to get ratings: %w", err)
	}

	c.logger.Info("api.ratings_retrieved", map[string]interface{}{
		"count": len(ratings),
	})

	return ratings, nil
}

// GetWatchlistConcurrent retrieves watchlist using concurrent processing
func (c *OptimizedClient) GetWatchlistConcurrent(ctx context.Context) ([]WatchlistMovie, error) {
	var watchlist []WatchlistMovie
	err := c.makeOptimizedRequest(ctx, "https://api.trakt.tv/users/me/watchlist/movies", &watchlist)
	if err != nil {
		return nil, fmt.Errorf("failed to get watchlist: %w", err)
	}

	c.logger.Info("api.watchlist_retrieved", map[string]interface{}{
		"count": len(watchlist),
	})

	return watchlist, nil
}

// ProcessBatchRequests processes multiple API requests concurrently
func (c *OptimizedClient) ProcessBatchRequests(ctx context.Context, requests []BatchRequest) ([]BatchResult, error) {
	if len(requests) == 0 {
		return nil, nil
	}

	results := make([]BatchResult, len(requests))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, c.workerPool.Stats().Workers)

	for i, req := range requests {
		wg.Add(1)
		go func(index int, request BatchRequest) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			start := time.Now()
			err := c.makeOptimizedRequest(ctx, request.Endpoint, request.Result)
			duration := time.Since(start)

			results[index] = BatchResult{
				Index:    index,
				Request:  request,
				Error:    err,
				Duration: duration,
			}
		}(i, req)
	}

	wg.Wait()

	// Count successes and failures
	var successes, failures int
	for _, result := range results {
		if result.Error == nil {
			successes++
		} else {
			failures++
		}
	}

	c.logger.Info("api.batch_completed", map[string]interface{}{
		"total":     len(requests),
		"successes": successes,
		"failures":  failures,
	})

	return results, nil
}

// GetCacheStats returns cache statistics
func (c *OptimizedClient) GetCacheStats() cache.CacheStats {
	return c.cache.Stats()
}

// GetPerformanceMetrics returns performance metrics
func (c *OptimizedClient) GetPerformanceMetrics() metrics.OverallStats {
	return c.metrics.GetOverallStats()
}

// ClearCache clears the API response cache
func (c *OptimizedClient) ClearCache() {
	c.cache.Clear()
	c.logger.Info("api.cache_cleared", nil)
}

// Close closes the optimized client and cleans up resources
func (c *OptimizedClient) Close() error {
	// Stop worker pool
	c.workerPool.Stop()
	
	// Close transport connections
	c.transport.CloseIdleConnections()
	
	// Log final performance metrics
	c.metrics.LogStats()
	
	c.logger.Info("api.client_closed", nil)
	return nil
}

// addExtendedInfo adds the extended parameter to the URL if configured
func (c *OptimizedClient) addExtendedInfo(endpoint string) string {
	if c.config.Trakt.ExtendedInfo == "" {
		return endpoint
	}

	baseURL, err := url.Parse(endpoint)
	if err != nil {
		c.logger.Warn("api.url_parse_error", map[string]interface{}{
			"error": err.Error(),
		})
		return endpoint
	}

	q := baseURL.Query()
	q.Set("extended", c.config.Trakt.ExtendedInfo)
	baseURL.RawQuery = q.Encode()
	return baseURL.String()
}

// Batch processing types

// BatchRequest represents a batch API request
type BatchRequest struct {
	Endpoint string
	Result   interface{}
}

// BatchResult represents the result of a batch request
type BatchResult struct {
	Index    int
	Request  BatchRequest
	Error    error
	Duration time.Duration
}

// ClientStats represents client statistics
type ClientStats struct {
	Cache       cache.CacheStats      `json:"cache"`
	Performance metrics.OverallStats `json:"performance"`
	WorkerPool  pool.PoolStats        `json:"worker_pool"`
} 