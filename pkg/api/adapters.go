package api

import (
	"context"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/performance/cache"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/performance/metrics"
)

// ClientAdapter adapts the basic Client to implement TraktAPIClient
type ClientAdapter struct {
	client *Client
}

// NewClientAdapter creates a new adapter for the basic Client
func NewClientAdapter(client *Client) TraktAPIClient {
	return &ClientAdapter{client: client}
}

// GetWatchedMovies implements TraktAPIClient
func (ca *ClientAdapter) GetWatchedMovies() ([]Movie, error) {
	return ca.client.GetWatchedMovies()
}

// GetCollectionMovies implements TraktAPIClient
func (ca *ClientAdapter) GetCollectionMovies() ([]CollectionMovie, error) {
	return ca.client.GetCollectionMovies()
}

// GetWatchedShows implements TraktAPIClient
func (ca *ClientAdapter) GetWatchedShows() ([]WatchedShow, error) {
	return ca.client.GetWatchedShows()
}

// GetRatings implements TraktAPIClient
func (ca *ClientAdapter) GetRatings() ([]Rating, error) {
	return ca.client.GetRatings()
}

// GetWatchlist implements TraktAPIClient
func (ca *ClientAdapter) GetWatchlist() ([]WatchlistMovie, error) {
	return ca.client.GetWatchlist()
}

// GetShowRatings implements TraktAPIClient
func (ca *ClientAdapter) GetShowRatings() ([]ShowRating, error) {
	return ca.client.GetShowRatings()
}

// GetEpisodeRatings implements TraktAPIClient
func (ca *ClientAdapter) GetEpisodeRatings() ([]EpisodeRating, error) {
	return ca.client.GetEpisodeRatings()
}

// GetMovieHistory implements TraktAPIClient
func (ca *ClientAdapter) GetMovieHistory() ([]HistoryItem, error) {
	return ca.client.GetMovieHistory()
}

// GetConfig implements TraktAPIClient
func (ca *ClientAdapter) GetConfig() *config.Config {
	return ca.client.GetConfig()
}

// Close implements TraktAPIClient - basic client doesn't need explicit cleanup
func (ca *ClientAdapter) Close() error {
	// Basic client doesn't require explicit cleanup
	return nil
}

// OptimizedClientAdapter adapts the OptimizedClient to implement both TraktAPIClient and OptimizedTraktAPIClient
type OptimizedClientAdapter struct {
	client *OptimizedClient
}

// NewOptimizedClientAdapter creates a new adapter for the OptimizedClient
func NewOptimizedClientAdapter(client *OptimizedClient) OptimizedTraktAPIClient {
	return &OptimizedClientAdapter{client: client}
}

// GetWatchedMovies implements TraktAPIClient - uses concurrent version for better performance
func (oca *OptimizedClientAdapter) GetWatchedMovies() ([]Movie, error) {
	ctx := context.Background()
	return oca.client.GetWatchedMoviesConcurrent(ctx)
}

// GetCollectionMovies implements TraktAPIClient - uses concurrent version for better performance
func (oca *OptimizedClientAdapter) GetCollectionMovies() ([]CollectionMovie, error) {
	ctx := context.Background()
	return oca.client.GetCollectionMoviesConcurrent(ctx)
}

// GetWatchedShows implements TraktAPIClient - fallback implementation
func (oca *OptimizedClientAdapter) GetWatchedShows() ([]WatchedShow, error) {
	// OptimizedClient doesn't have GetWatchedShows, would need to be implemented
	// For now, return empty slice - this should be implemented in the actual OptimizedClient
	return []WatchedShow{}, nil
}

// GetRatings implements TraktAPIClient - uses concurrent version for better performance
func (oca *OptimizedClientAdapter) GetRatings() ([]Rating, error) {
	ctx := context.Background()
	return oca.client.GetRatingsConcurrent(ctx)
}

// GetWatchlist implements TraktAPIClient - uses concurrent version for better performance
func (oca *OptimizedClientAdapter) GetWatchlist() ([]WatchlistMovie, error) {
	ctx := context.Background()
	return oca.client.GetWatchlistConcurrent(ctx)
}

// GetShowRatings implements TraktAPIClient - fallback implementation
func (oca *OptimizedClientAdapter) GetShowRatings() ([]ShowRating, error) {
	// OptimizedClient doesn't have GetShowRatings, would need to be implemented
	// For now, return empty slice - this should be implemented in the actual OptimizedClient
	return []ShowRating{}, nil
}

// GetEpisodeRatings implements TraktAPIClient - fallback implementation
func (oca *OptimizedClientAdapter) GetEpisodeRatings() ([]EpisodeRating, error) {
	// OptimizedClient doesn't have GetEpisodeRatings, would need to be implemented
	// For now, return empty slice - this should be implemented in the actual OptimizedClient
	return []EpisodeRating{}, nil
}

// GetMovieHistory implements TraktAPIClient - fallback implementation
func (oca *OptimizedClientAdapter) GetMovieHistory() ([]HistoryItem, error) {
	// OptimizedClient doesn't have GetMovieHistory, would need to be implemented
	// For now, return empty slice - this should be implemented in the actual OptimizedClient
	return []HistoryItem{}, nil
}

// GetConfig implements TraktAPIClient
func (oca *OptimizedClientAdapter) GetConfig() *config.Config {
	return oca.client.config
}

// Close implements TraktAPIClient
func (oca *OptimizedClientAdapter) Close() error {
	return oca.client.Close()
}

// GetWatchedMoviesConcurrent implements OptimizedTraktAPIClient
func (oca *OptimizedClientAdapter) GetWatchedMoviesConcurrent(ctx context.Context) ([]Movie, error) {
	return oca.client.GetWatchedMoviesConcurrent(ctx)
}

// GetCollectionMoviesConcurrent implements OptimizedTraktAPIClient
func (oca *OptimizedClientAdapter) GetCollectionMoviesConcurrent(ctx context.Context) ([]CollectionMovie, error) {
	return oca.client.GetCollectionMoviesConcurrent(ctx)
}

// GetRatingsConcurrent implements OptimizedTraktAPIClient
func (oca *OptimizedClientAdapter) GetRatingsConcurrent(ctx context.Context) ([]Rating, error) {
	return oca.client.GetRatingsConcurrent(ctx)
}

// GetWatchlistConcurrent implements OptimizedTraktAPIClient
func (oca *OptimizedClientAdapter) GetWatchlistConcurrent(ctx context.Context) ([]WatchlistMovie, error) {
	return oca.client.GetWatchlistConcurrent(ctx)
}

// ProcessBatchRequests implements OptimizedTraktAPIClient
func (oca *OptimizedClientAdapter) ProcessBatchRequests(ctx context.Context, requests []BatchRequest) ([]BatchResult, error) {
	return oca.client.ProcessBatchRequests(ctx, requests)
}

// GetCacheStats implements OptimizedTraktAPIClient
func (oca *OptimizedClientAdapter) GetCacheStats() cache.CacheStats {
	return oca.client.GetCacheStats()
}

// GetPerformanceMetrics implements OptimizedTraktAPIClient
func (oca *OptimizedClientAdapter) GetPerformanceMetrics() metrics.OverallStats {
	return oca.client.GetPerformanceMetrics()
}

// ClearCache implements OptimizedTraktAPIClient
func (oca *OptimizedClientAdapter) ClearCache() {
	oca.client.ClearCache()
}