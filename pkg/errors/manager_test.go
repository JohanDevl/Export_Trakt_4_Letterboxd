package errors

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/errors/handlers"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/errors/types"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

// Mock logger for testing
type mockLogger struct {
	logs []map[string]interface{}
}

func (m *mockLogger) Debug(msg string, fields ...map[string]interface{}) {
	allFields := make(map[string]interface{})
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			allFields[k] = v
		}
	}
	m.logs = append(m.logs, map[string]interface{}{"level": "debug", "msg": msg, "fields": allFields})
}

func (m *mockLogger) Info(msg string, fields ...map[string]interface{}) {
	allFields := make(map[string]interface{})
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			allFields[k] = v
		}
	}
	m.logs = append(m.logs, map[string]interface{}{"level": "info", "msg": msg, "fields": allFields})
}

func (m *mockLogger) Warn(msg string, fields ...map[string]interface{}) {
	allFields := make(map[string]interface{})
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			allFields[k] = v
		}
	}
	m.logs = append(m.logs, map[string]interface{}{"level": "warn", "msg": msg, "fields": allFields})
}

func (m *mockLogger) Error(msg string, fields ...map[string]interface{}) {
	allFields := make(map[string]interface{})
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			allFields[k] = v
		}
	}
	m.logs = append(m.logs, map[string]interface{}{"level": "error", "msg": msg, "fields": allFields})
}

func (m *mockLogger) Fatal(msg string, fields ...map[string]interface{}) {
	allFields := make(map[string]interface{})
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			allFields[k] = v
		}
	}
	m.logs = append(m.logs, map[string]interface{}{"level": "fatal", "msg": msg, "fields": allFields})
}

func (m *mockLogger) Debugf(msg string, data map[string]interface{}) {
	m.logs = append(m.logs, map[string]interface{}{"level": "debug", "msg": msg, "fields": data})
}

func (m *mockLogger) Infof(msg string, data map[string]interface{}) {
	m.logs = append(m.logs, map[string]interface{}{"level": "info", "msg": msg, "fields": data})
}

func (m *mockLogger) Warnf(msg string, data map[string]interface{}) {
	m.logs = append(m.logs, map[string]interface{}{"level": "warn", "msg": msg, "fields": data})
}

func (m *mockLogger) Errorf(msg string, data map[string]interface{}) {
	m.logs = append(m.logs, map[string]interface{}{"level": "error", "msg": msg, "fields": data})
}

func (m *mockLogger) SetLogLevel(level string) {}

func (m *mockLogger) SetLogFile(path string) error { 
	return nil 
}

func (m *mockLogger) SetTranslator(t logger.Translator) {}

// Mock recovery strategy
type mockRecoveryStrategy struct {
	canRecover bool
	shouldFail bool
}

func (m *mockRecoveryStrategy) CanRecover(err *types.AppError) bool {
	return m.canRecover
}

func (m *mockRecoveryStrategy) Recover(ctx context.Context, err *types.AppError) error {
	if m.shouldFail {
		return errors.New("recovery failed")
	}
	return nil
}

func (m *mockRecoveryStrategy) GetRecoveryTime() time.Duration {
	return time.Second
}

// Mock notification service
type mockNotificationService struct {
	errorNotifications          []string
	recoveryNotifications       []string
	circuitBreakerNotifications []string
}

func (m *mockNotificationService) NotifyError(ctx context.Context, err *types.AppError) error {
	m.errorNotifications = append(m.errorNotifications, err.Code)
	return nil
}

func (m *mockNotificationService) NotifyRecovery(ctx context.Context, err *types.AppError) error {
	m.recoveryNotifications = append(m.recoveryNotifications, err.Code)
	return nil
}

func (m *mockNotificationService) NotifyCircuitBreakerTrip(ctx context.Context, operation string) error {
	m.circuitBreakerNotifications = append(m.circuitBreakerNotifications, operation)
	return nil
}

func TestNewErrorManager(t *testing.T) {
	log := &mockLogger{}
	
	// Test with default config
	em := NewErrorManager(log, nil)
	if em == nil {
		t.Fatal("Expected error manager to be created")
	}
	
	if em.logger != log {
		t.Error("Expected logger to be set")
	}
	
	if em.config == nil {
		t.Error("Expected default config to be set")
	}
	
	if em.handlers == nil {
		t.Error("Expected handlers map to be initialized")
	}
	
	if em.metrics == nil {
		t.Error("Expected metrics to be initialized")
	}
	
	if em.recovery == nil {
		t.Error("Expected recovery manager to be initialized")
	}
	
	if em.circuitBreakers == nil {
		t.Error("Expected circuit breakers map to be initialized")
	}
}

func TestNewErrorManagerWithConfig(t *testing.T) {
	log := &mockLogger{}
	config := &ManagerConfig{
		EnableMetrics:       true,
		EnableRecovery:      true,
		EnableNotifications: true,
		EnableCircuitBreaker: true,
		MaxErrorsPerMinute:  50,
		AlertThreshold:      5,
		RetryConfig:         DefaultRetryConfig(),
	}
	
	em := NewErrorManager(log, config)
	if em == nil {
		t.Fatal("Expected error manager to be created")
	}
	
	if em.config != config {
		t.Error("Expected custom config to be set")
	}
	
	if em.config.MaxErrorsPerMinute != 50 {
		t.Error("Expected custom max errors per minute")
	}
}

func TestDefaultManagerConfig(t *testing.T) {
	config := DefaultManagerConfig()
	
	if !config.EnableMetrics {
		t.Error("Expected metrics to be enabled by default")
	}
	
	if !config.EnableRecovery {
		t.Error("Expected recovery to be enabled by default")
	}
	
	if config.EnableNotifications {
		t.Error("Expected notifications to be disabled by default")
	}
	
	if !config.EnableCircuitBreaker {
		t.Error("Expected circuit breaker to be enabled by default")
	}
	
	if config.MaxErrorsPerMinute != 100 {
		t.Errorf("Expected max errors per minute to be 100, got %d", config.MaxErrorsPerMinute)
	}
	
	if config.AlertThreshold != 10 {
		t.Errorf("Expected alert threshold to be 10, got %d", config.AlertThreshold)
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()
	
	if config.DefaultMaxAttempts != 3 {
		t.Errorf("Expected default max attempts to be 3, got %d", config.DefaultMaxAttempts)
	}
	
	if config.DefaultDelay != time.Second {
		t.Errorf("Expected default delay to be 1s, got %v", config.DefaultDelay)
	}
	
	if config.DefaultMaxDelay != 30*time.Second {
		t.Errorf("Expected default max delay to be 30s, got %v", config.DefaultMaxDelay)
	}
	
	if config.DefaultBackoffFactor != 2.0 {
		t.Errorf("Expected default backoff factor to be 2.0, got %f", config.DefaultBackoffFactor)
	}
	
	// Check per-category configs
	networkConfig := config.PerCategoryConfig[types.CategoryNetwork]
	if networkConfig == nil {
		t.Error("Expected network category config to be set")
	} else {
		if networkConfig.MaxAttempts != 5 {
			t.Errorf("Expected network max attempts to be 5, got %d", networkConfig.MaxAttempts)
		}
		if !networkConfig.EnableJitter {
			t.Error("Expected jitter to be enabled for network errors")
		}
		if !networkConfig.CircuitBreaker {
			t.Error("Expected circuit breaker to be enabled for network errors")
		}
	}
}

func TestHandleError(t *testing.T) {
	log := &mockLogger{}
	em := NewErrorManager(log, nil)
	ctx := context.Background()
	
	// Test with regular error
	regularErr := errors.New("test error")
	appErr := em.HandleError(ctx, regularErr)
	
	if appErr == nil {
		t.Fatal("Expected app error to be returned")
	}
	
	if appErr.Message != "test error" {
		t.Errorf("Expected message 'test error', got '%s'", appErr.Message)
	}
	
	// Test with app error
	originalAppErr := types.NewAppError(types.ErrNetworkTimeout, "timeout error", regularErr)
	processedErr := em.HandleError(ctx, originalAppErr)
	
	if processedErr.Code != types.ErrNetworkTimeout {
		t.Errorf("Expected error code %s, got %s", types.ErrNetworkTimeout, processedErr.Code)
	}
	
	// Check that metrics were updated
	metrics := em.GetMetrics()
	if metrics.TotalErrors < 2 {
		t.Errorf("Expected at least 2 errors to be recorded, got %d", metrics.TotalErrors)
	}
}

func TestTryRecover(t *testing.T) {
	log := &mockLogger{}
	config := DefaultManagerConfig()
	config.EnableRecovery = true
	em := NewErrorManager(log, config)
	ctx := context.Background()
	
	// Test recovery with no strategy
	appErr := types.NewAppError(types.ErrOperationFailed, "test error", nil)
	err := em.TryRecover(ctx, appErr)
	if err == nil {
		t.Error("Expected recovery to fail when no strategy is available")
	}
	
	// Test with recovery disabled
	config.EnableRecovery = false
	em2 := NewErrorManager(log, config)
	err = em2.TryRecover(ctx, appErr)
	if err == nil {
		t.Error("Expected recovery to fail when disabled")
	}
	
	// Test successful recovery
	config.EnableRecovery = true
	em3 := NewErrorManager(log, config)
	strategy := &mockRecoveryStrategy{canRecover: true, shouldFail: false}
	em3.RegisterRecoveryStrategy(types.CategoryOperation, strategy)
	
	err = em3.TryRecover(ctx, appErr)
	if err != nil {
		t.Errorf("Expected recovery to succeed, got error: %v", err)
	}
	
	metrics := em3.GetMetrics()
	if metrics.SuccessfulRecoveries == 0 {
		t.Error("Expected successful recovery to be recorded")
	}
	
	// Test failed recovery
	failStrategy := &mockRecoveryStrategy{canRecover: true, shouldFail: true}
	em4 := NewErrorManager(log, config)
	em4.RegisterRecoveryStrategy(types.CategoryOperation, failStrategy)
	
	err = em4.TryRecover(ctx, appErr)
	if err == nil {
		t.Error("Expected recovery to fail")
	}
	
	metrics = em4.GetMetrics()
	if metrics.FailedRecoveries == 0 {
		t.Error("Expected failed recovery to be recorded")
	}
}

func TestRegisterHandler(t *testing.T) {
	log := &mockLogger{}
	em := NewErrorManager(log, nil)
	
	// Mock handler for testing
	var mockHandler handlers.ErrorHandler
	
	em.RegisterHandler(types.CategoryNetwork, mockHandler)
	
	// Check that the handler was registered
	if len(em.handlers) == 0 {
		t.Error("Expected handler to be registered")
	}
	
	// Check that logging occurred
	found := false
	for _, logEntry := range log.logs {
		if logEntry["level"] == "info" && logEntry["msg"] == "error_manager.handler_registered" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected handler registration to be logged")
	}
}

func TestRegisterRecoveryStrategy(t *testing.T) {
	log := &mockLogger{}
	em := NewErrorManager(log, nil)
	strategy := &mockRecoveryStrategy{}
	
	em.RegisterRecoveryStrategy(types.CategoryNetwork, strategy)
	
	// Check that logging occurred
	found := false
	for _, logEntry := range log.logs {
		if logEntry["level"] == "info" && logEntry["msg"] == "error_manager.recovery_strategy_registered" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected recovery strategy registration to be logged")
	}
}

func TestSetNotificationService(t *testing.T) {
	log := &mockLogger{}
	em := NewErrorManager(log, nil)
	notificationService := &mockNotificationService{}
	
	em.SetNotificationService(notificationService)
	
	if em.notifications != notificationService {
		t.Error("Expected notification service to be set")
	}
}

func TestGetMetrics(t *testing.T) {
	log := &mockLogger{}
	em := NewErrorManager(log, nil)
	ctx := context.Background()
	
	// Generate some errors to create metrics
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	
	em.HandleError(ctx, err1)
	em.HandleError(ctx, err2)
	
	metrics := em.GetMetrics()
	if metrics == nil {
		t.Fatal("Expected metrics to be returned")
	}
	
	if metrics.TotalErrors < 2 {
		t.Errorf("Expected at least 2 errors, got %d", metrics.TotalErrors)
	}
	
	if len(metrics.ErrorsByCategory) == 0 {
		t.Error("Expected errors by category to be populated")
	}
	
	if len(metrics.ErrorsByCode) == 0 {
		t.Error("Expected errors by code to be populated")
	}
}

func TestCircuitBreakerFunctionality(t *testing.T) {
	log := &mockLogger{}
	config := DefaultManagerConfig()
	config.EnableCircuitBreaker = true
	em := NewErrorManager(log, config)
	ctx := context.Background()
	
	operation := "test_operation"
	
	// Test that circuit breaker is initially closed
	if em.IsCircuitBreakerOpen(operation) {
		t.Error("Expected circuit breaker to be closed initially")
	}
	
	// Simulate multiple failures to trip circuit breaker
	for i := 0; i < 6; i++ {
		appErr := types.NewAppError(types.ErrOperationFailed, "test error", nil)
		appErr.Context.Operation = operation
		em.HandleError(ctx, appErr)
	}
	
	// Check that circuit breaker is now open
	if !em.IsCircuitBreakerOpen(operation) {
		t.Error("Expected circuit breaker to be open after multiple failures")
	}
	
	metrics := em.GetMetrics()
	if metrics.CircuitBreakerTrips == 0 {
		t.Error("Expected circuit breaker trip to be recorded")
	}
}

func TestConvertToAppError(t *testing.T) {
	log := &mockLogger{}
	em := NewErrorManager(log, nil)
	
	// Test with regular error containing timeout
	timeoutErr := errors.New("connection timeout occurred")
	appErr := em.convertToAppError(timeoutErr)
	
	if appErr.Code != types.ErrNetworkTimeout {
		t.Errorf("Expected timeout error code, got %s", appErr.Code)
	}
	
	// Test with network connection error
	networkErr := errors.New("network connection failed")
	appErr = em.convertToAppError(networkErr)
	
	if appErr.Code != types.ErrNetworkUnavailable {
		t.Errorf("Expected network error code, got %s", appErr.Code)
	}
	
	// Test with unauthorized error
	authErr := errors.New("unauthorized access")
	appErr = em.convertToAppError(authErr)
	
	if appErr.Code != types.ErrUnauthorized {
		t.Errorf("Expected unauthorized error code, got %s", appErr.Code)
	}
	
	// Test with existing AppError
	existingAppErr := types.NewAppError(types.ErrRateLimited, "rate limited", nil)
	result := em.convertToAppError(existingAppErr)
	
	if result != existingAppErr {
		t.Error("Expected existing AppError to be returned unchanged")
	}
}

func TestNewErrorMetrics(t *testing.T) {
	metrics := NewErrorMetrics()
	
	if metrics == nil {
		t.Fatal("Expected metrics to be created")
	}
	
	if metrics.ErrorsByCategory == nil {
		t.Error("Expected ErrorsByCategory map to be initialized")
	}
	
	if metrics.ErrorsByCode == nil {
		t.Error("Expected ErrorsByCode map to be initialized")
	}
	
	if metrics.TotalErrors != 0 {
		t.Error("Expected TotalErrors to be initialized to 0")
	}
}

func TestNewRecoveryManager(t *testing.T) {
	recovery := NewRecoveryManager()
	
	if recovery == nil {
		t.Fatal("Expected recovery manager to be created")
	}
	
	if recovery.strategies == nil {
		t.Error("Expected strategies map to be initialized")
	}
	
	if recovery.maxRetries == nil {
		t.Error("Expected maxRetries map to be initialized")
	}
}

func TestContainsHelper(t *testing.T) {
	// Test contains function
	if !contains("timeout error", "timeout") {
		t.Error("Expected 'timeout' to be found in 'timeout error'")
	}
	
	if contains("normal error", "timeout") {
		t.Error("Expected 'timeout' not to be found in 'normal error'")
	}
	
	if !contains("network connection failed", "network", "connection") {
		t.Error("Expected multiple substrings to be found")
	}
	
	if contains("test", "longer_than_test") {
		t.Error("Expected longer substring not to be found in shorter string")
	}
}

func TestConvertToHandlerConfig(t *testing.T) {
	// Test with nil config
	result := convertToHandlerConfig(nil)
	if result != nil {
		t.Error("Expected nil result for nil input")
	}
	
	// Test with valid config
	config := &CategoryRetryConfig{
		MaxAttempts:     5,
		InitialDelay:    time.Second,
		MaxDelay:        10 * time.Second,
		BackoffFactor:   2.0,
		EnableJitter:    true,
		CircuitBreaker:  true,
	}
	
	result = convertToHandlerConfig(config)
	if result == nil {
		t.Fatal("Expected non-nil result for valid input")
	}
	
	if result.MaxAttempts != config.MaxAttempts {
		t.Errorf("Expected MaxAttempts %d, got %d", config.MaxAttempts, result.MaxAttempts)
	}
	
	if result.InitialDelay != config.InitialDelay {
		t.Errorf("Expected InitialDelay %v, got %v", config.InitialDelay, result.InitialDelay)
	}
	
	if result.MaxDelay != config.MaxDelay {
		t.Errorf("Expected MaxDelay %v, got %v", config.MaxDelay, result.MaxDelay)
	}
	
	if result.BackoffFactor != config.BackoffFactor {
		t.Errorf("Expected BackoffFactor %f, got %f", config.BackoffFactor, result.BackoffFactor)
	}
	
	if result.EnableJitter != config.EnableJitter {
		t.Errorf("Expected EnableJitter %t, got %t", config.EnableJitter, result.EnableJitter)
	}
	
	if result.CircuitBreaker != config.CircuitBreaker {
		t.Errorf("Expected CircuitBreaker %t, got %t", config.CircuitBreaker, result.CircuitBreaker)
	}
}