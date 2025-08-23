package api

import (
	"fmt"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/errors"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/performance/cache"
)

// DefaultClientFactory implements the ClientFactory interface
type DefaultClientFactory struct {
	errorManager *errors.ErrorManager
}

// ClientFactoryConfig configures the client factory
type ClientFactoryConfig struct {
	ErrorManager *errors.ErrorManager
	Logger       logger.Logger
}

// NewDefaultClientFactory creates a new default client factory
func NewDefaultClientFactory() ClientFactory {
	return &DefaultClientFactory{}
}

// NewClientFactory creates a new client factory with configuration
func NewClientFactory(config ClientFactoryConfig) ClientFactory {
	return &DefaultClientFactory{
		errorManager: config.ErrorManager,
	}
}

// CreateBasicClient creates a basic Trakt API client
func (f *DefaultClientFactory) CreateBasicClient(cfg *config.Config) (TraktAPIClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration cannot be nil")
	}

	// Note: Logger needs to be provided externally in a real implementation
	// For now, we'll assume a logger is available or create a no-op logger
	// This would typically be injected via dependency injection
	
	// Create basic client
	client := NewClient(cfg, nil) // Logger will need to be injected properly
	
	// Wrap in adapter
	adaptedClient := NewClientAdapter(client)
	
	// Optionally wrap with error management
	if f.errorManager != nil {
		return NewErrorAwareClient(adaptedClient, f.errorManager), nil
	}
	
	return adaptedClient, nil
}

// CreateOptimizedClient creates an optimized Trakt API client
func (f *DefaultClientFactory) CreateOptimizedClient(cfg OptimizedClientConfig) (OptimizedTraktAPIClient, error) {
	// Validate configuration
	if cfg.Config == nil {
		return nil, fmt.Errorf("base configuration cannot be nil")
	}

	// Configuration is already complete - just create the client
	optimizedConfig := cfg

	// Create optimized client
	client := NewOptimizedClient(optimizedConfig)

	// Wrap in adapter
	adaptedClient := NewOptimizedClientAdapter(client)
	
	// Optionally wrap with error management
	if f.errorManager != nil {
		return NewErrorAwareOptimizedClient(adaptedClient, f.errorManager), nil
	}
	
	return adaptedClient, nil
}

// CreateClientWithCapabilities creates a client with specific capabilities
func (f *DefaultClientFactory) CreateClientWithCapabilities(cfg ClientCapabilitiesConfig) (TraktAPIClient, error) {
	if cfg.BaseConfig == nil {
		return nil, fmt.Errorf("base configuration cannot be nil")
	}

	// If advanced capabilities are requested, create optimized client
	if cfg.EnableCaching || cfg.EnableMetrics || cfg.EnableConcurrency || cfg.WorkerPoolSize > 0 {
		var cacheConfig cache.CacheConfig
		if cfg.CacheConfig != nil {
			cacheConfig = *cfg.CacheConfig
		} else {
			cacheConfig = cache.CacheConfig{
				Capacity: 1000,
				TTL:      24 * time.Hour,
			}
		}

		optimizedConfig := OptimizedClientConfig{
			Config:           cfg.BaseConfig,
			Logger:           nil, // Would need to be provided
			WorkerPoolSize:   cfg.WorkerPoolSize,
			CacheConfig:      cacheConfig,
		}

		if cfg.WorkerPoolSize <= 0 {
			optimizedConfig.WorkerPoolSize = 10 // default
		}

		return f.CreateOptimizedClient(optimizedConfig)
	}

	// Otherwise, create basic client
	return f.CreateBasicClient(cfg.BaseConfig)
}

