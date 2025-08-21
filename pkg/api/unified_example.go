package api

import (
	"context"
	"fmt"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/errors"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/performance/cache"
)

// UnifiedAPIExample demonstrates how to use the unified API architecture
func UnifiedAPIExample() error {
	// 1. Create configuration
	cfg := &config.Config{
		// Configuration details would be loaded here
	}

	// 2. Create logger (would typically be injected)
	log := createLogger() // Placeholder for actual logger creation

	// 3. Create error manager
	errorManager := errors.NewErrorManager(log, nil)

	// 4. Create client factory with error management
	factory := NewClientFactory(ClientFactoryConfig{
		ErrorManager: errorManager,
		Logger:       log,
	})

	// 5. Create a basic client
	basicClient, err := factory.CreateBasicClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create basic client: %w", err)
	}
	defer basicClient.Close()

	// 6. Create an optimized client
	optimizedConfig := OptimizedClientConfig{
		BaseConfig:       cfg,
		Logger:           log,
		WorkerPoolSize:   10,
		RateLimitPerSec:  100,
		ConnectionPool:   20,
		CacheConfig: cache.CacheConfig{
			Capacity: 1000,
			TTL:      24 * 60 * 60 * 1000, // 24 hours in milliseconds
		},
	}

	optimizedClient, err := factory.CreateOptimizedClient(optimizedConfig)
	if err != nil {
		return fmt.Errorf("failed to create optimized client: %w", err)
	}
	defer optimizedClient.Close()

	// 7. Use the clients - all errors are automatically handled by the ErrorManager
	ctx := context.Background()

	// Basic operations
	movies, err := basicClient.GetWatchedMovies()
	if err != nil {
		// Error has been processed by ErrorManager
		fmt.Printf("Error getting watched movies: %v\n", err)
		
		// Try to recover using the error manager if possible
		if errorAware, ok := basicClient.(*ErrorAwareClient); ok {
			if recoveryErr := errorAware.TryRecoverFromError(ctx, err); recoveryErr != nil {
				fmt.Printf("Recovery failed: %v\n", recoveryErr)
			} else {
				fmt.Println("Successfully recovered from error")
				// Retry the operation
				movies, err = basicClient.GetWatchedMovies()
			}
		}
	}
	
	if err == nil {
		fmt.Printf("Retrieved %d watched movies\n", len(movies))
	}

	// Optimized operations
	ratings, err := optimizedClient.GetRatingsConcurrent(ctx)
	if err != nil {
		fmt.Printf("Error getting ratings concurrently: %v\n", err)
	} else {
		fmt.Printf("Retrieved %d ratings concurrently\n", len(ratings))
	}

	// Get performance metrics
	metrics := optimizedClient.GetPerformanceMetrics()
	fmt.Printf("Performance metrics: %+v\n", metrics)

	// Get cache statistics
	cacheStats := optimizedClient.GetCacheStats()
	fmt.Printf("Cache statistics: %+v\n", cacheStats)

	// Get error manager metrics
	errorMetrics := errorManager.GetMetrics()
	fmt.Printf("Error metrics: Total errors: %d, By category: %+v\n", 
		errorMetrics.TotalErrors, errorMetrics.ErrorsByCategory)

	return nil
}

// createLogger is a placeholder function for logger creation
func createLogger() logger.Logger {
	// This would create an actual logger implementation
	// For example purposes, returning nil (would need actual implementation)
	return nil
}

// Demonstration of creating clients with specific capabilities
func CreateClientWithSpecificCapabilities() (TraktAPIClient, error) {
	cfg := &config.Config{
		// Configuration details
	}

	// Create factory
	factory := NewDefaultClientFactory()

	// Create client with specific capabilities
	capabilitiesConfig := ClientCapabilitiesConfig{
		BaseConfig:        cfg,
		EnableCaching:     true,
		EnableMetrics:     true,
		EnableRetry:       true,
		EnableRateLimit:   true,
		EnableConcurrency: true,
		WorkerPoolSize:    15,
		CacheConfig: &cache.CacheConfig{
			Capacity: 2000,
			TTL:      48 * 60 * 60 * 1000, // 48 hours
		},
	}

	return factory.CreateClientWithCapabilities(capabilitiesConfig)
}

// Demonstration of API operation execution
func ExecuteAPIOperationsExample() error {
	// Create client with error management
	cfg := &config.Config{}
	log := createLogger()
	errorManager := errors.NewErrorManager(log, nil)
	
	factory := NewClientFactory(ClientFactoryConfig{
		ErrorManager: errorManager,
		Logger:       log,
	})

	client, err := factory.CreateBasicClient(cfg)
	if err != nil {
		return err
	}
	defer client.Close()

	// Execute multiple operations
	operations := []func() error{
		func() error {
			_, err := client.GetWatchedMovies()
			return err
		},
		func() error {
			_, err := client.GetCollectionMovies()
			return err
		},
		func() error {
			_, err := client.GetRatings()
			return err
		},
	}

	for i, operation := range operations {
		if err := operation(); err != nil {
			fmt.Printf("Operation %d failed: %v\n", i+1, err)
			
			// Check if circuit breaker is open
			if errorManager.IsCircuitBreakerOpen(fmt.Sprintf("operation_%d", i+1)) {
				fmt.Printf("Circuit breaker is open for operation %d\n", i+1)
				continue
			}
		} else {
			fmt.Printf("Operation %d completed successfully\n", i+1)
		}
	}

	return nil
}