package api

import (
	"context"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/performance/cache"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/performance/metrics"
)

// TraktAPIClient defines the unified interface for all Trakt API clients
type TraktAPIClient interface {
	// Core data retrieval operations
	GetWatchedMovies() ([]Movie, error)
	GetCollectionMovies() ([]CollectionMovie, error)
	GetWatchedShows() ([]WatchedShow, error)
	GetRatings() ([]Rating, error)
	GetWatchlist() ([]WatchlistMovie, error)
	GetShowRatings() ([]ShowRating, error)
	GetEpisodeRatings() ([]EpisodeRating, error)
	GetMovieHistory() ([]HistoryItem, error)
	
	// Configuration and lifecycle
	GetConfig() *config.Config
	Close() error
}

// OptimizedTraktAPIClient extends TraktAPIClient with performance-oriented operations
type OptimizedTraktAPIClient interface {
	TraktAPIClient
	
	// Concurrent operations
	GetWatchedMoviesConcurrent(ctx context.Context) ([]Movie, error)
	GetCollectionMoviesConcurrent(ctx context.Context) ([]CollectionMovie, error)
	GetRatingsConcurrent(ctx context.Context) ([]Rating, error)
	GetWatchlistConcurrent(ctx context.Context) ([]WatchlistMovie, error)
	
	// Batch operations
	ProcessBatchRequests(ctx context.Context, requests []BatchRequest) ([]BatchResult, error)
	
	// Performance management
	GetCacheStats() cache.CacheStats
	GetPerformanceMetrics() metrics.OverallStats
	ClearCache()
}

// ClientFactory defines a factory interface for creating API clients
type ClientFactory interface {
	// Create basic client
	CreateBasicClient(cfg *config.Config) (TraktAPIClient, error)
	
	// Create optimized client
	CreateOptimizedClient(cfg OptimizedClientConfig) (OptimizedTraktAPIClient, error)
	
	// Create client with specific capabilities
	CreateClientWithCapabilities(cfg ClientCapabilitiesConfig) (TraktAPIClient, error)
}

// ClientCapabilitiesConfig defines what capabilities a client should have
type ClientCapabilitiesConfig struct {
	BaseConfig        *config.Config
	EnableCaching     bool
	EnableMetrics     bool
	EnableRetry       bool
	EnableRateLimit   bool
	EnableConcurrency bool
	WorkerPoolSize    int
	CacheConfig       *cache.CacheConfig
}

// APIOperation represents a generic API operation that can be executed
type APIOperation interface {
	Execute(ctx context.Context, client TraktAPIClient) (interface{}, error)
	GetOperationName() string
	GetMaxRetries() int
	IsRetryable(error) bool
}

// APIOperationResult represents the result of an API operation
type APIOperationResult struct {
	Data      interface{}
	Error     error
	Duration  time.Duration
	Attempts  int
	Operation string
}

// APIExecutor defines an interface for executing API operations with various strategies
type APIExecutor interface {
	// Execute single operation
	Execute(ctx context.Context, operation APIOperation) *APIOperationResult
	
	// Execute multiple operations concurrently
	ExecuteBatch(ctx context.Context, operations []APIOperation) []*APIOperationResult
	
	// Execute with specific client
	ExecuteWithClient(ctx context.Context, client TraktAPIClient, operation APIOperation) *APIOperationResult
}

// RetryableOperation defines operations that can be retried
type RetryableOperation interface {
	APIOperation
	GetRetryPolicy() RetryPolicy
}

// RetryPolicy defines how retries should be handled
type RetryPolicy struct {
	MaxAttempts   int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
	RetryOn       []error // Specific errors to retry on
}

// BatchOperation groups multiple API operations
type BatchOperation struct {
	Operations []APIOperation
	MaxConcurrency int
	FailFast   bool // Stop on first error
}

// ClientMetrics provides unified metrics across all client types
type ClientMetrics interface {
	GetRequestCount() int64
	GetErrorCount() int64
	GetAverageResponseTime() time.Duration
	GetCacheHitRatio() float64
	GetCircuitBreakerStatus() string
}

// ClientHealth provides health check capabilities
type ClientHealth interface {
	HealthCheck(ctx context.Context) error
	GetLastError() error
	GetUptime() time.Duration
	IsHealthy() bool
}

// UnifiedTraktClient combines all client capabilities
type UnifiedTraktClient interface {
	OptimizedTraktAPIClient
	ClientMetrics
	ClientHealth
	
	// Advanced features
	ExecuteOperation(ctx context.Context, operation APIOperation) *APIOperationResult
	ExecuteBatchOperations(ctx context.Context, batch BatchOperation) []*APIOperationResult
	
	// Client management
	GetClientType() string
	GetCapabilities() []string
	Reconfigure(cfg interface{}) error
}