package retry

import (
	"context"
	"fmt"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/errors/types"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/retry/backoff"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/retry/circuit"
)

// Client provides retry functionality with circuit breaker protection
type Client struct {
	backoff       backoff.Backoff
	circuitBreaker *circuit.CircuitBreaker
	retryChecker   RetryChecker
}

// RetryChecker determines if an error should be retried
type RetryChecker func(error) bool

// Config represents retry client configuration
type Config struct {
	BackoffConfig       *backoff.ExponentialBackoff
	CircuitBreakerConfig *circuit.Config
	RetryChecker        RetryChecker
}

// DefaultConfig returns default retry configuration
func DefaultConfig() *Config {
	return &Config{
		BackoffConfig:       backoff.DefaultExponentialBackoff(),
		CircuitBreakerConfig: circuit.DefaultConfig(),
		RetryChecker:        DefaultRetryChecker,
	}
}

// NewClient creates a new retry client
func NewClient(config *Config) *Client {
	if config == nil {
		config = DefaultConfig()
	}
	
	return &Client{
		backoff:       config.BackoffConfig,
		circuitBreaker: circuit.NewCircuitBreaker(config.CircuitBreakerConfig),
		retryChecker:   config.RetryChecker,
	}
}

// Execute executes a function with retry logic and circuit breaker protection
func (c *Client) Execute(ctx context.Context, operation string, fn func(context.Context) error) error {
	var lastErr error
	
	for attempt := 0; c.backoff.ShouldRetry(attempt); attempt++ {
		// Execute with circuit breaker protection
		err := c.circuitBreaker.Execute(ctx, fn)
		if err == nil {
			return nil // Success
		}
		
		lastErr = err
		
		// Check if error is retryable
		if !c.retryChecker(err) {
			return types.NewAppErrorWithOperation(
				types.ErrOperationFailed,
				"operation failed with non-retryable error",
				operation,
				err,
			)
		}
		
		// Check circuit breaker state
		if err == circuit.ErrCircuitBreakerOpen {
			return types.NewAppErrorWithOperation(
				types.ErrNetworkUnavailable,
				"circuit breaker is open, service unavailable",
				operation,
				err,
			)
		}
		
		// Calculate delay for next attempt
		if c.backoff.ShouldRetry(attempt + 1) {
			delay := c.backoff.CalculateDelay(attempt)
			select {
			case <-ctx.Done():
				return types.NewAppErrorWithOperation(
					types.ErrOperationCanceled,
					"operation canceled during retry delay",
					operation,
					ctx.Err(),
				)
			case <-time.After(delay):
				continue
			}
		}
	}
	
	// Max retries exceeded
	return types.NewAppErrorWithOperation(
		types.ErrOperationFailed,
		fmt.Sprintf("max retries exceeded after %d attempts", c.backoff.(*backoff.ExponentialBackoff).MaxRetries),
		operation,
		lastErr,
	)
}

// ExecuteWithTimeout executes a function with a specific timeout
func (c *Client) ExecuteWithTimeout(ctx context.Context, timeout time.Duration, operation string, fn func(context.Context) error) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	return c.Execute(timeoutCtx, operation, fn)
}

// DefaultRetryChecker is the default retry checker function
func DefaultRetryChecker(err error) bool {
	if err == nil {
		return false
	}
	
	// Check if it's an AppError with a retryable code
	if appErr, ok := err.(*types.AppError); ok {
		return types.IsRetryableError(appErr.Code)
	}
	
	// Check for common retryable errors
	switch err {
	case circuit.ErrTimeout:
		return true
	case context.DeadlineExceeded:
		return true
	case context.Canceled:
		return false // Don't retry canceled operations
	default:
		// Check error message for common patterns
		errMsg := err.Error()
		retryablePatterns := []string{
			"timeout",
			"connection refused",
			"no such host",
			"network unreachable",
			"temporary failure",
		}
		
		for _, pattern := range retryablePatterns {
			if contains(errMsg, pattern) {
				return true
			}
		}
		
		return false
	}
}

// CreateRetryCheckerForCodes creates a retry checker for specific error codes
func CreateRetryCheckerForCodes(retryCodes []string) RetryChecker {
	codeMap := make(map[string]bool)
	for _, code := range retryCodes {
		codeMap[code] = true
	}
	
	return func(err error) bool {
		if err == nil {
			return false
		}
		
		if appErr, ok := err.(*types.AppError); ok {
			return codeMap[appErr.Code]
		}
		
		return DefaultRetryChecker(err)
	}
}

// Stats returns retry client statistics
func (c *Client) Stats() circuit.Stats {
	return c.circuitBreaker.Stats()
}

// ResetCircuitBreaker resets the circuit breaker to closed state
func (c *Client) ResetCircuitBreaker() {
	c.circuitBreaker.Reset()
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
		 containsRecursive(s, substr, 0)))
}

func containsRecursive(s, substr string, index int) bool {
	if index > len(s)-len(substr) {
		return false
	}
	
	if s[index:index+len(substr)] == substr {
		return true
	}
	
	return containsRecursive(s, substr, index+1)
} 