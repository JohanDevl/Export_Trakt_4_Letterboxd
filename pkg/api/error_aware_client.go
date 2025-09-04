package api

import (
	"context"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/errors"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/errors/types"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/performance/cache"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/performance/metrics"
)

// ErrorAwareClient wraps any TraktAPIClient with unified error handling
type ErrorAwareClient struct {
	client       TraktAPIClient
	errorManager *errors.ErrorManager
}

// NewErrorAwareClient creates a new error-aware client wrapper
func NewErrorAwareClient(client TraktAPIClient, errorManager *errors.ErrorManager) TraktAPIClient {
	return &ErrorAwareClient{
		client:       client,
		errorManager: errorManager,
	}
}

// GetWatchedMovies implements TraktAPIClient with error handling
func (eac *ErrorAwareClient) GetWatchedMovies() ([]Movie, error) {
	movies, err := eac.client.GetWatchedMovies()
	if err != nil {
		ctx := context.Background()
		appErr := eac.errorManager.HandleError(ctx, err)
		return movies, appErr
	}
	return movies, nil
}

// GetCollectionMovies implements TraktAPIClient with error handling
func (eac *ErrorAwareClient) GetCollectionMovies() ([]CollectionMovie, error) {
	movies, err := eac.client.GetCollectionMovies()
	if err != nil {
		ctx := context.Background()
		appErr := eac.errorManager.HandleError(ctx, err)
		return movies, appErr
	}
	return movies, nil
}

// GetWatchedShows implements TraktAPIClient with error handling
func (eac *ErrorAwareClient) GetWatchedShows() ([]WatchedShow, error) {
	shows, err := eac.client.GetWatchedShows()
	if err != nil {
		ctx := context.Background()
		appErr := eac.errorManager.HandleError(ctx, err)
		return shows, appErr
	}
	return shows, nil
}

// GetRatings implements TraktAPIClient with error handling
func (eac *ErrorAwareClient) GetRatings() ([]Rating, error) {
	ratings, err := eac.client.GetRatings()
	if err != nil {
		ctx := context.Background()
		appErr := eac.errorManager.HandleError(ctx, err)
		return ratings, appErr
	}
	return ratings, nil
}

// GetWatchlist implements TraktAPIClient with error handling
func (eac *ErrorAwareClient) GetWatchlist() ([]WatchlistMovie, error) {
	watchlist, err := eac.client.GetWatchlist()
	if err != nil {
		ctx := context.Background()
		appErr := eac.errorManager.HandleError(ctx, err)
		return watchlist, appErr
	}
	return watchlist, nil
}

// GetShowRatings implements TraktAPIClient with error handling
func (eac *ErrorAwareClient) GetShowRatings() ([]ShowRating, error) {
	ratings, err := eac.client.GetShowRatings()
	if err != nil {
		ctx := context.Background()
		appErr := eac.errorManager.HandleError(ctx, err)
		return ratings, appErr
	}
	return ratings, nil
}

// GetEpisodeRatings implements TraktAPIClient with error handling
func (eac *ErrorAwareClient) GetEpisodeRatings() ([]EpisodeRating, error) {
	ratings, err := eac.client.GetEpisodeRatings()
	if err != nil {
		ctx := context.Background()
		appErr := eac.errorManager.HandleError(ctx, err)
		return ratings, appErr
	}
	return ratings, nil
}

// GetMovieHistory implements TraktAPIClient with error handling
func (eac *ErrorAwareClient) GetMovieHistory() ([]HistoryItem, error) {
	history, err := eac.client.GetMovieHistory()
	if err != nil {
		ctx := context.Background()
		appErr := eac.errorManager.HandleError(ctx, err)
		return history, appErr
	}
	return history, nil
}

// GetConfig implements TraktAPIClient
func (eac *ErrorAwareClient) GetConfig() *config.Config {
	return eac.client.GetConfig()
}

// Close implements TraktAPIClient with error handling
func (eac *ErrorAwareClient) Close() error {
	err := eac.client.Close()
	if err != nil {
		ctx := context.Background()
		appErr := eac.errorManager.HandleError(ctx, err)
		return appErr
	}
	return nil
}

// TryRecoverFromError attempts to recover from an error using the error manager
func (eac *ErrorAwareClient) TryRecoverFromError(ctx context.Context, err error) error {
	if eac.errorManager == nil {
		return err
	}
	
	// Convert to AppError if needed
	var appErr *types.AppError
	if ae, ok := err.(*types.AppError); ok {
		appErr = ae
	} else {
		appErr = types.NewAppError(types.ErrOperationFailed, err.Error(), err)
	}
	
	// Try recovery
	return eac.errorManager.TryRecover(ctx, appErr)
}

// ErrorAwareOptimizedClient wraps OptimizedTraktAPIClient with unified error handling
type ErrorAwareOptimizedClient struct {
	client       OptimizedTraktAPIClient
	errorManager *errors.ErrorManager
}

// NewErrorAwareOptimizedClient creates a new error-aware optimized client wrapper
func NewErrorAwareOptimizedClient(client OptimizedTraktAPIClient, errorManager *errors.ErrorManager) OptimizedTraktAPIClient {
	return &ErrorAwareOptimizedClient{
		client:       client,
		errorManager: errorManager,
	}
}

// GetWatchedMovies implements TraktAPIClient with error handling
func (eaoc *ErrorAwareOptimizedClient) GetWatchedMovies() ([]Movie, error) {
	movies, err := eaoc.client.GetWatchedMovies()
	if err != nil {
		ctx := context.Background()
		appErr := eaoc.errorManager.HandleError(ctx, err)
		return movies, appErr
	}
	return movies, nil
}

// GetCollectionMovies implements TraktAPIClient with error handling
func (eaoc *ErrorAwareOptimizedClient) GetCollectionMovies() ([]CollectionMovie, error) {
	movies, err := eaoc.client.GetCollectionMovies()
	if err != nil {
		ctx := context.Background()
		appErr := eaoc.errorManager.HandleError(ctx, err)
		return movies, appErr
	}
	return movies, nil
}

// GetWatchedShows implements TraktAPIClient with error handling
func (eaoc *ErrorAwareOptimizedClient) GetWatchedShows() ([]WatchedShow, error) {
	shows, err := eaoc.client.GetWatchedShows()
	if err != nil {
		ctx := context.Background()
		appErr := eaoc.errorManager.HandleError(ctx, err)
		return shows, appErr
	}
	return shows, nil
}

// GetRatings implements TraktAPIClient with error handling
func (eaoc *ErrorAwareOptimizedClient) GetRatings() ([]Rating, error) {
	ratings, err := eaoc.client.GetRatings()
	if err != nil {
		ctx := context.Background()
		appErr := eaoc.errorManager.HandleError(ctx, err)
		return ratings, appErr
	}
	return ratings, nil
}

// GetWatchlist implements TraktAPIClient with error handling
func (eaoc *ErrorAwareOptimizedClient) GetWatchlist() ([]WatchlistMovie, error) {
	watchlist, err := eaoc.client.GetWatchlist()
	if err != nil {
		ctx := context.Background()
		appErr := eaoc.errorManager.HandleError(ctx, err)
		return watchlist, appErr
	}
	return watchlist, nil
}

// GetShowRatings implements TraktAPIClient with error handling
func (eaoc *ErrorAwareOptimizedClient) GetShowRatings() ([]ShowRating, error) {
	ratings, err := eaoc.client.GetShowRatings()
	if err != nil {
		ctx := context.Background()
		appErr := eaoc.errorManager.HandleError(ctx, err)
		return ratings, appErr
	}
	return ratings, nil
}

// GetEpisodeRatings implements TraktAPIClient with error handling
func (eaoc *ErrorAwareOptimizedClient) GetEpisodeRatings() ([]EpisodeRating, error) {
	ratings, err := eaoc.client.GetEpisodeRatings()
	if err != nil {
		ctx := context.Background()
		appErr := eaoc.errorManager.HandleError(ctx, err)
		return ratings, appErr
	}
	return ratings, nil
}

// GetMovieHistory implements TraktAPIClient with error handling
func (eaoc *ErrorAwareOptimizedClient) GetMovieHistory() ([]HistoryItem, error) {
	history, err := eaoc.client.GetMovieHistory()
	if err != nil {
		ctx := context.Background()
		appErr := eaoc.errorManager.HandleError(ctx, err)
		return history, appErr
	}
	return history, nil
}

// GetConfig implements TraktAPIClient
func (eaoc *ErrorAwareOptimizedClient) GetConfig() *config.Config {
	return eaoc.client.GetConfig()
}

// Close implements TraktAPIClient with error handling
func (eaoc *ErrorAwareOptimizedClient) Close() error {
	err := eaoc.client.Close()
	if err != nil {
		ctx := context.Background()
		appErr := eaoc.errorManager.HandleError(ctx, err)
		return appErr
	}
	return nil
}

// GetWatchedMoviesConcurrent implements OptimizedTraktAPIClient with error handling
func (eaoc *ErrorAwareOptimizedClient) GetWatchedMoviesConcurrent(ctx context.Context) ([]Movie, error) {
	movies, err := eaoc.client.GetWatchedMoviesConcurrent(ctx)
	if err != nil {
		appErr := eaoc.errorManager.HandleError(ctx, err)
		return movies, appErr
	}
	return movies, nil
}

// GetCollectionMoviesConcurrent implements OptimizedTraktAPIClient with error handling
func (eaoc *ErrorAwareOptimizedClient) GetCollectionMoviesConcurrent(ctx context.Context) ([]CollectionMovie, error) {
	movies, err := eaoc.client.GetCollectionMoviesConcurrent(ctx)
	if err != nil {
		appErr := eaoc.errorManager.HandleError(ctx, err)
		return movies, appErr
	}
	return movies, nil
}

// GetRatingsConcurrent implements OptimizedTraktAPIClient with error handling
func (eaoc *ErrorAwareOptimizedClient) GetRatingsConcurrent(ctx context.Context) ([]Rating, error) {
	ratings, err := eaoc.client.GetRatingsConcurrent(ctx)
	if err != nil {
		appErr := eaoc.errorManager.HandleError(ctx, err)
		return ratings, appErr
	}
	return ratings, nil
}

// GetWatchlistConcurrent implements OptimizedTraktAPIClient with error handling
func (eaoc *ErrorAwareOptimizedClient) GetWatchlistConcurrent(ctx context.Context) ([]WatchlistMovie, error) {
	watchlist, err := eaoc.client.GetWatchlistConcurrent(ctx)
	if err != nil {
		appErr := eaoc.errorManager.HandleError(ctx, err)
		return watchlist, appErr
	}
	return watchlist, nil
}

// ProcessBatchRequests implements OptimizedTraktAPIClient with error handling
func (eaoc *ErrorAwareOptimizedClient) ProcessBatchRequests(ctx context.Context, requests []BatchRequest) ([]BatchResult, error) {
	results, err := eaoc.client.ProcessBatchRequests(ctx, requests)
	if err != nil {
		appErr := eaoc.errorManager.HandleError(ctx, err)
		return results, appErr
	}
	return results, nil
}

// GetCacheStats implements OptimizedTraktAPIClient
func (eaoc *ErrorAwareOptimizedClient) GetCacheStats() cache.CacheStats {
	return eaoc.client.GetCacheStats()
}

// GetPerformanceMetrics implements OptimizedTraktAPIClient
func (eaoc *ErrorAwareOptimizedClient) GetPerformanceMetrics() metrics.OverallStats {
	return eaoc.client.GetPerformanceMetrics()
}

// ClearCache implements OptimizedTraktAPIClient
func (eaoc *ErrorAwareOptimizedClient) ClearCache() {
	eaoc.client.ClearCache()
}

// TryRecoverFromError attempts to recover from an error using the error manager
func (eaoc *ErrorAwareOptimizedClient) TryRecoverFromError(ctx context.Context, err error) error {
	if appErr, ok := err.(*types.AppError); ok {
		return eaoc.errorManager.TryRecover(ctx, appErr)
	}
	
	// Convert to AppError first
	appErr := eaoc.errorManager.HandleError(ctx, err)
	return eaoc.errorManager.TryRecover(ctx, appErr)
}