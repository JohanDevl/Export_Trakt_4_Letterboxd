package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/errors/types"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

// OperationErrorHandler handles operation-related errors
type OperationErrorHandler struct {
	BaseHandler
	Logger logger.Logger
	Config *CategoryRetryConfig
}

// NewOperationErrorHandler creates a new operation error handler
func NewOperationErrorHandler(logger logger.Logger, config *CategoryRetryConfig) *OperationErrorHandler {
	if config == nil {
		config = &CategoryRetryConfig{
			MaxAttempts:     3,
			InitialDelay:    time.Second,
			MaxDelay:        15 * time.Second,
			BackoffFactor:   2.0,
			EnableJitter:    true,
			CircuitBreaker:  true,
		}
	}

	return &OperationErrorHandler{
		BaseHandler: BaseHandler{category: types.CategoryOperation},
		Logger:      logger,
		Config:      config,
	}
}

// Handle processes operation errors with appropriate recovery strategies
func (h *OperationErrorHandler) Handle(ctx context.Context, err *types.AppError) *types.AppError {
	h.Logger.Warn("operation_error_handler.handling_error", map[string]interface{}{
		"error_code": err.Code,
		"message":    err.Message,
		"operation":  err.Context.Operation,
	})

	// Enhance error with operation-specific context
	enhancedErr := h.enhanceOperationError(err)

	// Apply operation-specific handling strategies
	switch err.Code {
	case types.ErrExportFailed:
		return h.handleExportFailed(ctx, enhancedErr)
	case types.ErrImportFailed:
		return h.handleImportFailed(ctx, enhancedErr)
	case types.ErrFileSystem:
		return h.handleFileSystemError(ctx, enhancedErr)
	case types.ErrProcessingFailed:
		return h.handleProcessingFailed(ctx, enhancedErr)
	case types.ErrOperationCanceled:
		return h.handleOperationCanceled(ctx, enhancedErr)
	case types.ErrOperationFailed:
		return h.handleOperationFailed(ctx, enhancedErr)
	default:
		return h.handleGenericOperationError(ctx, enhancedErr)
	}
}

// enhanceOperationError adds operation-specific metadata to the error
func (h *OperationErrorHandler) enhanceOperationError(err *types.AppError) *types.AppError {
	err.WithMetadata("handler", "operation")
	err.WithMetadata("category", string(types.CategoryOperation))
	err.WithMetadata("retryable", fmt.Sprintf("%t", types.IsRetryableError(err.Code)))
	
	if h.Config != nil {
		err.WithMetadata("max_attempts", fmt.Sprintf("%d", h.Config.MaxAttempts))
		err.WithMetadata("circuit_breaker", fmt.Sprintf("%t", h.Config.CircuitBreaker))
	}

	return err
}

// handleExportFailed handles export operation failures
func (h *OperationErrorHandler) handleExportFailed(ctx context.Context, err *types.AppError) *types.AppError {
	h.Logger.Error("operation_error_handler.export_failed", map[string]interface{}{
		"operation": err.Context.Operation,
		"details":   err.Details,
	})

	err.WithMetadata("handling_strategy", "export_failed")
	err.WithMetadata("recommended_action", "retry_export_operation")
	err.WithMetadata("checkpoint_recovery", "true")
	
	return err
}

// handleImportFailed handles import operation failures
func (h *OperationErrorHandler) handleImportFailed(ctx context.Context, err *types.AppError) *types.AppError {
	h.Logger.Error("operation_error_handler.import_failed", map[string]interface{}{
		"operation": err.Context.Operation,
		"details":   err.Details,
	})

	err.WithMetadata("handling_strategy", "import_failed")
	err.WithMetadata("recommended_action", "validate_input_and_retry")
	err.WithMetadata("data_validation", "required")
	
	return err
}

// handleFileSystemError handles file system operation errors
func (h *OperationErrorHandler) handleFileSystemError(ctx context.Context, err *types.AppError) *types.AppError {
	h.Logger.Error("operation_error_handler.filesystem_error", map[string]interface{}{
		"operation": err.Context.Operation,
		"details":   err.Details,
	})

	err.WithMetadata("handling_strategy", "filesystem_error")
	err.WithMetadata("recommended_action", "check_permissions_and_disk_space")
	err.WithMetadata("system_check", "required")
	
	return err
}

// handleProcessingFailed handles data processing failures
func (h *OperationErrorHandler) handleProcessingFailed(ctx context.Context, err *types.AppError) *types.AppError {
	h.Logger.Error("operation_error_handler.processing_failed", map[string]interface{}{
		"operation": err.Context.Operation,
		"details":   err.Details,
	})

	err.WithMetadata("handling_strategy", "processing_failed")
	err.WithMetadata("recommended_action", "validate_data_and_retry")
	err.WithMetadata("data_recovery", "possible")
	
	return err
}

// handleOperationCanceled handles canceled operations
func (h *OperationErrorHandler) handleOperationCanceled(ctx context.Context, err *types.AppError) *types.AppError {
	h.Logger.Info("operation_error_handler.operation_canceled", map[string]interface{}{
		"operation": err.Context.Operation,
	})

	err.WithMetadata("handling_strategy", "operation_canceled")
	err.WithMetadata("recommended_action", "cleanup_and_notify_user")
	err.WithMetadata("retry_allowed", "false")
	
	return err
}

// handleOperationFailed handles generic operation failures
func (h *OperationErrorHandler) handleOperationFailed(ctx context.Context, err *types.AppError) *types.AppError {
	h.Logger.Error("operation_error_handler.operation_failed", map[string]interface{}{
		"operation":  err.Context.Operation,
		"error_code": err.Code,
	})

	err.WithMetadata("handling_strategy", "operation_failed")
	err.WithMetadata("recommended_action", "retry_with_backoff")
	err.WithMetadata("diagnostic_required", "true")
	
	return err
}

// handleGenericOperationError handles other operation errors
func (h *OperationErrorHandler) handleGenericOperationError(ctx context.Context, err *types.AppError) *types.AppError {
	h.Logger.Warn("operation_error_handler.generic", map[string]interface{}{
		"operation":  err.Context.Operation,
		"error_code": err.Code,
	})

	err.WithMetadata("handling_strategy", "generic_operation")
	err.WithMetadata("recommended_action", "analyze_and_retry")
	
	return err
}