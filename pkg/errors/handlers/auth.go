package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/errors/types"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

// AuthErrorHandler handles authentication-related errors
type AuthErrorHandler struct {
	BaseHandler
	Logger logger.Logger
	Config *CategoryRetryConfig
}

// NewAuthErrorHandler creates a new authentication error handler
func NewAuthErrorHandler(logger logger.Logger, config *CategoryRetryConfig) *AuthErrorHandler {
	if config == nil {
		config = &CategoryRetryConfig{
			MaxAttempts:     2,
			InitialDelay:    time.Second,
			MaxDelay:        5 * time.Second,
			BackoffFactor:   1.5,
			EnableJitter:    false,
			CircuitBreaker:  false,
		}
	}

	return &AuthErrorHandler{
		BaseHandler: BaseHandler{category: types.CategoryAuthentication},
		Logger:      logger,
		Config:      config,
	}
}

// Handle processes authentication errors with appropriate strategies
func (h *AuthErrorHandler) Handle(ctx context.Context, err *types.AppError) *types.AppError {
	h.Logger.Warn("auth_error_handler.handling_error", map[string]interface{}{
		"error_code": err.Code,
		"message":    err.Message,
		"operation":  err.Context.Operation,
	})

	// Enhance error with auth-specific context
	enhancedErr := h.enhanceAuthError(err)

	// Apply auth-specific handling strategies
	switch err.Code {
	case types.ErrInvalidCredentials:
		return h.handleInvalidCredentials(ctx, enhancedErr)
	case types.ErrTokenExpired:
		return h.handleTokenExpired(ctx, enhancedErr)
	case types.ErrUnauthorized:
		return h.handleUnauthorized(ctx, enhancedErr)
	case types.ErrForbidden:
		return h.handleForbidden(ctx, enhancedErr)
	case types.ErrAPIKeyMissing:
		return h.handleAPIKeyMissing(ctx, enhancedErr)
	default:
		return h.handleGenericAuthError(ctx, enhancedErr)
	}
}

// enhanceAuthError adds authentication-specific metadata to the error
func (h *AuthErrorHandler) enhanceAuthError(err *types.AppError) *types.AppError {
	err.WithMetadata("handler", "authentication")
	err.WithMetadata("category", string(types.CategoryAuthentication))
	err.WithMetadata("retryable", fmt.Sprintf("%t", types.IsRetryableError(err.Code)))
	err.WithMetadata("security_sensitive", "true")
	
	if h.Config != nil {
		err.WithMetadata("max_attempts", fmt.Sprintf("%d", h.Config.MaxAttempts))
	}

	return err
}

// handleInvalidCredentials handles invalid credential errors
func (h *AuthErrorHandler) handleInvalidCredentials(ctx context.Context, err *types.AppError) *types.AppError {
	h.Logger.Error("auth_error_handler.invalid_credentials", map[string]interface{}{
		"operation": err.Context.Operation,
		"user_id":   err.Context.UserID,
	})

	err.WithMetadata("handling_strategy", "invalid_credentials")
	err.WithMetadata("recommended_action", "prompt_for_new_credentials")
	err.WithMetadata("requires_user_intervention", "true")
	
	return err
}

// handleTokenExpired handles expired token errors
func (h *AuthErrorHandler) handleTokenExpired(ctx context.Context, err *types.AppError) *types.AppError {
	h.Logger.Info("auth_error_handler.token_expired", map[string]interface{}{
		"operation": err.Context.Operation,
		"user_id":   err.Context.UserID,
	})

	err.WithMetadata("handling_strategy", "token_expired")
	err.WithMetadata("recommended_action", "refresh_token")
	err.WithMetadata("auto_recoverable", "true")
	
	return err
}

// handleUnauthorized handles unauthorized access errors
func (h *AuthErrorHandler) handleUnauthorized(ctx context.Context, err *types.AppError) *types.AppError {
	h.Logger.Warn("auth_error_handler.unauthorized", map[string]interface{}{
		"operation": err.Context.Operation,
		"user_id":   err.Context.UserID,
	})

	err.WithMetadata("handling_strategy", "unauthorized")
	err.WithMetadata("recommended_action", "verify_authentication_status")
	err.WithMetadata("requires_user_intervention", "true")
	
	return err
}

// handleForbidden handles forbidden access errors
func (h *AuthErrorHandler) handleForbidden(ctx context.Context, err *types.AppError) *types.AppError {
	h.Logger.Warn("auth_error_handler.forbidden", map[string]interface{}{
		"operation": err.Context.Operation,
		"user_id":   err.Context.UserID,
	})

	err.WithMetadata("handling_strategy", "forbidden")
	err.WithMetadata("recommended_action", "check_permissions")
	err.WithMetadata("requires_admin_intervention", "true")
	
	return err
}

// handleAPIKeyMissing handles missing API key errors
func (h *AuthErrorHandler) handleAPIKeyMissing(ctx context.Context, err *types.AppError) *types.AppError {
	h.Logger.Error("auth_error_handler.api_key_missing", map[string]interface{}{
		"operation": err.Context.Operation,
	})

	err.WithMetadata("handling_strategy", "api_key_missing")
	err.WithMetadata("recommended_action", "configure_api_key")
	err.WithMetadata("requires_configuration", "true")
	
	return err
}

// handleGenericAuthError handles other authentication errors
func (h *AuthErrorHandler) handleGenericAuthError(ctx context.Context, err *types.AppError) *types.AppError {
	h.Logger.Warn("auth_error_handler.generic", map[string]interface{}{
		"operation":  err.Context.Operation,
		"error_code": err.Code,
	})

	err.WithMetadata("handling_strategy", "generic_auth")
	err.WithMetadata("recommended_action", "verify_authentication_configuration")
	
	return err
}