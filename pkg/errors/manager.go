package errors

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/errors/handlers"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/errors/types"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

// ErrorManager provides centralized error handling and management
type ErrorManager struct {
	logger           logger.Logger
	handlers         map[types.ErrorCategory]handlers.ErrorHandler
	metrics          *ErrorMetrics
	recovery         *RecoveryManager
	notifications    NotificationService
	circuitBreakers  map[string]*CircuitBreakerState
	mutex            sync.RWMutex
	
	// Configuration
	config *ManagerConfig
}

// ManagerConfig defines configuration for the error manager
type ManagerConfig struct {
	EnableMetrics       bool
	EnableRecovery      bool
	EnableNotifications bool
	EnableCircuitBreaker bool
	MaxErrorsPerMinute  int
	AlertThreshold      int
	RetryConfig         *RetryConfig
}

// RetryConfig defines global retry configuration
type RetryConfig struct {
	DefaultMaxAttempts int
	DefaultDelay       time.Duration
	DefaultMaxDelay    time.Duration
	DefaultBackoffFactor float64
	PerCategoryConfig  map[types.ErrorCategory]*CategoryRetryConfig
}

// CategoryRetryConfig defines retry configuration per error category
type CategoryRetryConfig struct {
	MaxAttempts     int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	EnableJitter    bool
	CircuitBreaker  bool
}

// CircuitBreakerState tracks circuit breaker state for operations
type CircuitBreakerState struct {
	IsOpen          bool
	ErrorCount      int
	LastError       time.Time
	ConsecutiveFails int
	NextRetryAt     time.Time
}

// ErrorMetrics tracks error statistics
type ErrorMetrics struct {
	mutex                    sync.RWMutex
	TotalErrors              int64
	ErrorsByCategory         map[types.ErrorCategory]int64
	ErrorsByCode             map[string]int64
	ErrorsPerMinute         int64
	LastErrorTime           time.Time
	CircuitBreakerTrips     int64
	SuccessfulRecoveries    int64
	FailedRecoveries        int64
}

// RecoveryManager handles error recovery strategies
type RecoveryManager struct {
	strategies map[types.ErrorCategory]RecoveryStrategy
	maxRetries map[types.ErrorCategory]int
}

// RecoveryStrategy defines how to recover from specific error types
type RecoveryStrategy interface {
	CanRecover(error *types.AppError) bool
	Recover(ctx context.Context, error *types.AppError) error
	GetRecoveryTime() time.Duration
}

// NotificationService handles error notifications
type NotificationService interface {
	NotifyError(ctx context.Context, error *types.AppError) error
	NotifyRecovery(ctx context.Context, error *types.AppError) error
	NotifyCircuitBreakerTrip(ctx context.Context, operation string) error
}

// NewErrorManager creates a new error manager
func NewErrorManager(logger logger.Logger, config *ManagerConfig) *ErrorManager {
	if config == nil {
		config = DefaultManagerConfig()
	}

	em := &ErrorManager{
		logger:          logger,
		handlers:        make(map[types.ErrorCategory]handlers.ErrorHandler),
		metrics:         NewErrorMetrics(),
		recovery:        NewRecoveryManager(),
		circuitBreakers: make(map[string]*CircuitBreakerState),
		config:          config,
	}

	// Initialize default handlers
	em.initializeDefaultHandlers()

	return em
}

// DefaultManagerConfig returns default configuration
func DefaultManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		EnableMetrics:       true,
		EnableRecovery:      true,
		EnableNotifications: false,
		EnableCircuitBreaker: true,
		MaxErrorsPerMinute:  100,
		AlertThreshold:      10,
		RetryConfig:         DefaultRetryConfig(),
	}
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		DefaultMaxAttempts:   3,
		DefaultDelay:         time.Second,
		DefaultMaxDelay:      30 * time.Second,
		DefaultBackoffFactor: 2.0,
		PerCategoryConfig: map[types.ErrorCategory]*CategoryRetryConfig{
			types.CategoryNetwork: {
				MaxAttempts:     5,
				InitialDelay:    500 * time.Millisecond,
				MaxDelay:        10 * time.Second,
				BackoffFactor:   2.0,
				EnableJitter:    true,
				CircuitBreaker:  true,
			},
			types.CategoryAuthentication: {
				MaxAttempts:     2,
				InitialDelay:    time.Second,
				MaxDelay:        5 * time.Second,
				BackoffFactor:   1.5,
				EnableJitter:    false,
				CircuitBreaker:  false,
			},
			types.CategoryOperation: {
				MaxAttempts:     3,
				InitialDelay:    time.Second,
				MaxDelay:        15 * time.Second,
				BackoffFactor:   2.0,
				EnableJitter:    true,
				CircuitBreaker:  true,
			},
		},
	}
}

// HandleError handles an error through the unified system
func (em *ErrorManager) HandleError(ctx context.Context, err error) *types.AppError {
	// Convert to AppError if necessary
	appErr := em.convertToAppError(err)
	
	// Update metrics
	if em.config.EnableMetrics {
		em.updateMetrics(appErr)
	}
	
	// Check circuit breaker
	if em.config.EnableCircuitBreaker {
		em.updateCircuitBreaker(appErr)
	}
	
	// Get category-specific handler
	category := types.GetErrorCategory(appErr.Code)
	handler, exists := em.handlers[category]
	
	if exists {
		// Handle with specific handler
		return handler.Handle(ctx, appErr)
	}
	
	// Default handling
	return em.defaultHandle(ctx, appErr)
}

// TryRecover attempts to recover from an error
func (em *ErrorManager) TryRecover(ctx context.Context, err *types.AppError) error {
	if !em.config.EnableRecovery {
		return fmt.Errorf("recovery disabled")
	}
	
	category := types.GetErrorCategory(err.Code)
	strategy, exists := em.recovery.strategies[category]
	
	if !exists || !strategy.CanRecover(err) {
		return fmt.Errorf("no recovery strategy available for error: %s", err.Code)
	}
	
	em.logger.Info("error_manager.attempting_recovery", map[string]interface{}{
		"error_code": err.Code,
		"operation":  err.Context.Operation,
	})
	
	if recoveryErr := strategy.Recover(ctx, err); recoveryErr != nil {
		em.metrics.FailedRecoveries++
		return fmt.Errorf("recovery failed: %w", recoveryErr)
	}
	
	em.metrics.SuccessfulRecoveries++
	em.logger.Info("error_manager.recovery_successful", map[string]interface{}{
		"error_code": err.Code,
		"operation":  err.Context.Operation,
	})
	
	if em.notifications != nil {
		em.notifications.NotifyRecovery(ctx, err)
	}
	
	return nil
}

// RegisterHandler registers a handler for a specific error category
func (em *ErrorManager) RegisterHandler(category types.ErrorCategory, handler handlers.ErrorHandler) {
	em.mutex.Lock()
	defer em.mutex.Unlock()
	
	em.handlers[category] = handler
	em.logger.Info("error_manager.handler_registered", map[string]interface{}{
		"category": string(category),
	})
}

// RegisterRecoveryStrategy registers a recovery strategy for a specific category
func (em *ErrorManager) RegisterRecoveryStrategy(category types.ErrorCategory, strategy RecoveryStrategy) {
	em.recovery.strategies[category] = strategy
	em.logger.Info("error_manager.recovery_strategy_registered", map[string]interface{}{
		"category": string(category),
	})
}

// SetNotificationService sets the notification service
func (em *ErrorManager) SetNotificationService(service NotificationService) {
	em.notifications = service
}

// GetMetrics returns current error metrics
func (em *ErrorManager) GetMetrics() *ErrorMetrics {
	em.metrics.mutex.RLock()
	defer em.metrics.mutex.RUnlock()
	
	// Return a copy to avoid race conditions
	metrics := &ErrorMetrics{
		TotalErrors:             em.metrics.TotalErrors,
		ErrorsByCategory:        make(map[types.ErrorCategory]int64),
		ErrorsByCode:            make(map[string]int64),
		ErrorsPerMinute:         em.metrics.ErrorsPerMinute,
		LastErrorTime:           em.metrics.LastErrorTime,
		CircuitBreakerTrips:     em.metrics.CircuitBreakerTrips,
		SuccessfulRecoveries:    em.metrics.SuccessfulRecoveries,
		FailedRecoveries:        em.metrics.FailedRecoveries,
	}
	
	for k, v := range em.metrics.ErrorsByCategory {
		metrics.ErrorsByCategory[k] = v
	}
	for k, v := range em.metrics.ErrorsByCode {
		metrics.ErrorsByCode[k] = v
	}
	
	return metrics
}

// IsCircuitBreakerOpen checks if circuit breaker is open for an operation
func (em *ErrorManager) IsCircuitBreakerOpen(operation string) bool {
	em.mutex.RLock()
	defer em.mutex.RUnlock()
	
	state, exists := em.circuitBreakers[operation]
	if !exists {
		return false
	}
	
	// Check if circuit breaker should reset
	if state.IsOpen && time.Now().After(state.NextRetryAt) {
		state.IsOpen = false
		state.ConsecutiveFails = 0
		em.logger.Info("error_manager.circuit_breaker_reset", map[string]interface{}{
			"operation": operation,
		})
	}
	
	return state.IsOpen
}

// convertToAppError converts any error to AppError
func (em *ErrorManager) convertToAppError(err error) *types.AppError {
	if appErr, ok := err.(*types.AppError); ok {
		return appErr
	}
	
	// Try to determine error category from error message
	code := types.ErrOperationFailed // default
	message := err.Error()
	
	// Simple heuristics to classify errors
	switch {
	case contains(message, "timeout", "deadline"):
		code = types.ErrNetworkTimeout
	case contains(message, "connection", "network"):
		code = types.ErrNetworkUnavailable
	case contains(message, "unauthorized", "auth"):
		code = types.ErrUnauthorized
	case contains(message, "forbidden"):
		code = types.ErrForbidden
	case contains(message, "rate limit"):
		code = types.ErrRateLimited
	case contains(message, "file", "directory"):
		code = types.ErrFileSystem
	}
	
	return types.NewAppError(code, message, err)
}

// updateMetrics updates error metrics
func (em *ErrorManager) updateMetrics(err *types.AppError) {
	em.metrics.mutex.Lock()
	defer em.metrics.mutex.Unlock()
	
	em.metrics.TotalErrors++
	em.metrics.LastErrorTime = time.Now()
	
	category := types.GetErrorCategory(err.Code)
	if em.metrics.ErrorsByCategory == nil {
		em.metrics.ErrorsByCategory = make(map[types.ErrorCategory]int64)
	}
	em.metrics.ErrorsByCategory[category]++
	
	if em.metrics.ErrorsByCode == nil {
		em.metrics.ErrorsByCode = make(map[string]int64)
	}
	em.metrics.ErrorsByCode[err.Code]++
	
	// Calculate errors per minute (simplified)
	em.metrics.ErrorsPerMinute++
}

// updateCircuitBreaker updates circuit breaker state
func (em *ErrorManager) updateCircuitBreaker(err *types.AppError) {
	operation := err.Context.Operation
	if operation == "" {
		return
	}
	
	em.mutex.Lock()
	defer em.mutex.Unlock()
	
	state, exists := em.circuitBreakers[operation]
	if !exists {
		state = &CircuitBreakerState{}
		em.circuitBreakers[operation] = state
	}
	
	state.ErrorCount++
	state.LastError = time.Now()
	state.ConsecutiveFails++
	
	// Check if should trip circuit breaker
	if state.ConsecutiveFails >= 5 && !state.IsOpen {
		state.IsOpen = true
		state.NextRetryAt = time.Now().Add(30 * time.Second)
		em.metrics.CircuitBreakerTrips++
		
		em.logger.Warn("error_manager.circuit_breaker_tripped", map[string]interface{}{
			"operation":         operation,
			"consecutive_fails": state.ConsecutiveFails,
		})
		
		if em.notifications != nil {
			em.notifications.NotifyCircuitBreakerTrip(context.Background(), operation)
		}
	}
}

// defaultHandle provides default error handling
func (em *ErrorManager) defaultHandle(ctx context.Context, err *types.AppError) *types.AppError {
	em.logger.Error("error_manager.unhandled_error", map[string]interface{}{
		"error_code": err.Code,
		"message":    err.Message,
		"operation":  err.Context.Operation,
	})
	
	if em.notifications != nil {
		em.notifications.NotifyError(ctx, err)
	}
	
	return err
}

// initializeDefaultHandlers sets up default error handlers
func (em *ErrorManager) initializeDefaultHandlers() {
	// Network errors handler
	networkConfig := em.config.RetryConfig.PerCategoryConfig[types.CategoryNetwork]
	em.handlers[types.CategoryNetwork] = &handlers.NetworkErrorHandler{
		Logger: em.logger,
		Config: convertToHandlerConfig(networkConfig),
	}
	
	// Authentication errors handler
	authConfig := em.config.RetryConfig.PerCategoryConfig[types.CategoryAuthentication]
	em.handlers[types.CategoryAuthentication] = &handlers.AuthErrorHandler{
		Logger: em.logger,
		Config: convertToHandlerConfig(authConfig),
	}
	
	// Operation errors handler
	opConfig := em.config.RetryConfig.PerCategoryConfig[types.CategoryOperation]
	em.handlers[types.CategoryOperation] = &handlers.OperationErrorHandler{
		Logger: em.logger,
		Config: convertToHandlerConfig(opConfig),
	}
}

// NewErrorMetrics creates new error metrics
func NewErrorMetrics() *ErrorMetrics {
	return &ErrorMetrics{
		ErrorsByCategory: make(map[types.ErrorCategory]int64),
		ErrorsByCode:     make(map[string]int64),
	}
}

// NewRecoveryManager creates new recovery manager
func NewRecoveryManager() *RecoveryManager {
	return &RecoveryManager{
		strategies: make(map[types.ErrorCategory]RecoveryStrategy),
		maxRetries: make(map[types.ErrorCategory]int),
	}
}

// Helper function to check if string contains any of the substrings
func contains(s string, substrings ...string) bool {
	for _, substr := range substrings {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

// convertToHandlerConfig converts CategoryRetryConfig to handlers.CategoryRetryConfig
func convertToHandlerConfig(config *CategoryRetryConfig) *handlers.CategoryRetryConfig {
	if config == nil {
		return nil
	}
	
	return &handlers.CategoryRetryConfig{
		MaxAttempts:    config.MaxAttempts,
		InitialDelay:   config.InitialDelay,
		MaxDelay:       config.MaxDelay,
		BackoffFactor:  config.BackoffFactor,
		EnableJitter:   config.EnableJitter,
		CircuitBreaker: config.CircuitBreaker,
	}
}