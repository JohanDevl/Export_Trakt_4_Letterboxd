package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/errors/types"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

// NetworkErrorHandler handles network-related errors with retry logic
type NetworkErrorHandler struct {
	BaseHandler
	Logger logger.Logger
	Config *CategoryRetryConfig
}


// NewNetworkErrorHandler creates a new network error handler
func NewNetworkErrorHandler(logger logger.Logger, config *CategoryRetryConfig) *NetworkErrorHandler {
	if config == nil {
		config = &CategoryRetryConfig{
			MaxAttempts:     5,
			InitialDelay:    500 * time.Millisecond,
			MaxDelay:        10 * time.Second,
			BackoffFactor:   2.0,
			EnableJitter:    true,
			CircuitBreaker:  true,
		}
	}

	return &NetworkErrorHandler{
		BaseHandler: BaseHandler{category: types.CategoryNetwork},
		Logger:      logger,
		Config:      config,
	}
}

// Handle processes network errors with appropriate retry strategies
func (h *NetworkErrorHandler) Handle(ctx context.Context, err *types.AppError) *types.AppError {
	h.Logger.Warn("network_error_handler.handling_error", map[string]interface{}{
		"error_code": err.Code,
		"message":    err.Message,
		"operation":  err.Context.Operation,
	})

	// Enhance error with network-specific context
	enhancedErr := h.enhanceNetworkError(err)

	// Apply network-specific handling strategies
	switch err.Code {
	case types.ErrNetworkTimeout:
		return h.handleTimeout(ctx, enhancedErr)
	case types.ErrNetworkUnavailable:
		return h.handleUnavailable(ctx, enhancedErr)
	case types.ErrRateLimited:
		return h.handleRateLimit(ctx, enhancedErr)
	case types.ErrConnectionRefused:
		return h.handleConnectionRefused(ctx, enhancedErr)
	case types.ErrDNSResolution:
		return h.handleDNSResolution(ctx, enhancedErr)
	default:
		return h.handleGenericNetworkError(ctx, enhancedErr)
	}
}

// enhanceNetworkError adds network-specific metadata to the error
func (h *NetworkErrorHandler) enhanceNetworkError(err *types.AppError) *types.AppError {
	err.WithMetadata("handler", "network")
	err.WithMetadata("category", string(types.CategoryNetwork))
	err.WithMetadata("retryable", fmt.Sprintf("%t", types.IsRetryableError(err.Code)))
	err.WithMetadata("temporary", fmt.Sprintf("%t", types.IsTemporaryError(err.Code)))
	
	if h.Config != nil {
		err.WithMetadata("max_attempts", fmt.Sprintf("%d", h.Config.MaxAttempts))
		err.WithMetadata("initial_delay", h.Config.InitialDelay.String())
	}

	return err
}

// handleTimeout handles network timeout errors
func (h *NetworkErrorHandler) handleTimeout(ctx context.Context, err *types.AppError) *types.AppError {
	h.Logger.Info("network_error_handler.handling_timeout", map[string]interface{}{
		"operation": err.Context.Operation,
		"config":    h.Config,
	})

	err.WithMetadata("handling_strategy", "timeout_retry")
	err.WithMetadata("recommended_action", "retry_with_backoff")
	
	return err
}

// handleUnavailable handles service unavailable errors
func (h *NetworkErrorHandler) handleUnavailable(ctx context.Context, err *types.AppError) *types.AppError {
	h.Logger.Info("network_error_handler.handling_unavailable", map[string]interface{}{
		"operation": err.Context.Operation,
		"config":    h.Config,
	})

	err.WithMetadata("handling_strategy", "service_unavailable")
	err.WithMetadata("recommended_action", "retry_with_exponential_backoff")
	
	return err
}

// handleRateLimit handles rate limiting errors
func (h *NetworkErrorHandler) handleRateLimit(ctx context.Context, err *types.AppError) *types.AppError {
	h.Logger.Info("network_error_handler.handling_rate_limit", map[string]interface{}{
		"operation": err.Context.Operation,
		"config":    h.Config,
	})

	err.WithMetadata("handling_strategy", "rate_limit")
	err.WithMetadata("recommended_action", "delay_and_retry")
	
	// Add longer delay for rate limit errors
	if h.Config != nil {
		backoffDelay := time.Duration(float64(h.Config.MaxDelay) * 1.5)
		err.WithMetadata("recommended_delay", backoffDelay.String())
	}
	
	return err
}

// handleConnectionRefused handles connection refused errors
func (h *NetworkErrorHandler) handleConnectionRefused(ctx context.Context, err *types.AppError) *types.AppError {
	h.Logger.Info("network_error_handler.handling_connection_refused", map[string]interface{}{
		"operation": err.Context.Operation,
		"config":    h.Config,
	})

	err.WithMetadata("handling_strategy", "connection_refused")
	err.WithMetadata("recommended_action", "retry_with_backoff")
	
	return err
}

// handleDNSResolution handles DNS resolution errors
func (h *NetworkErrorHandler) handleDNSResolution(ctx context.Context, err *types.AppError) *types.AppError {
	h.Logger.Info("network_error_handler.handling_dns_resolution", map[string]interface{}{
		"operation": err.Context.Operation,
		"config":    h.Config,
	})

	err.WithMetadata("handling_strategy", "dns_resolution")
	err.WithMetadata("recommended_action", "retry_with_short_delay")
	
	return err
}

// handleGenericNetworkError handles other network errors
func (h *NetworkErrorHandler) handleGenericNetworkError(ctx context.Context, err *types.AppError) *types.AppError {
	h.Logger.Info("network_error_handler.handling_generic", map[string]interface{}{
		"operation":  err.Context.Operation,
		"error_code": err.Code,
		"config":     h.Config,
	})

	err.WithMetadata("handling_strategy", "generic_network")
	err.WithMetadata("recommended_action", "retry_with_standard_backoff")
	
	return err
}