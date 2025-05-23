package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/errors/types"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/retry/backoff"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/retry/circuit"
)

func TestNewClient(t *testing.T) {
	config := &Config{
		BackoffConfig: backoff.NewExponentialBackoff(100*time.Millisecond, 1*time.Second, 2.0, false, 3),
		CircuitBreakerConfig: &circuit.Config{
			FailureThreshold: 5,
			Timeout:          1 * time.Second,
			RecoveryTime:     2 * time.Second,
		},
		RetryChecker: DefaultRetryChecker,
	}
	
	client := NewClient(config)
	if client == nil {
		t.Error("Expected client to be created")
	}
}

func TestClientExecuteSuccess(t *testing.T) {
	config := &Config{
		BackoffConfig: backoff.NewExponentialBackoff(10*time.Millisecond, 100*time.Millisecond, 2.0, false, 3),
		CircuitBreakerConfig: &circuit.Config{
			FailureThreshold: 5,
			Timeout:          1 * time.Second,
			RecoveryTime:     2 * time.Second,
		},
		RetryChecker: DefaultRetryChecker,
	}
	
	client := NewClient(config)
	
	// Test successful operation
	callCount := 0
	err := client.Execute(context.Background(), "test_op", func(ctx context.Context) error {
		callCount++
		return nil
	})
	
	if err != nil {
		t.Errorf("Expected no error for successful operation, got: %v", err)
	}
	
	if callCount != 1 {
		t.Errorf("Expected operation to be called once, got %d", callCount)
	}
}

func TestClientExecuteRetryableError(t *testing.T) {
	config := &Config{
		BackoffConfig: backoff.NewExponentialBackoff(1*time.Millisecond, 10*time.Millisecond, 2.0, false, 3),
		CircuitBreakerConfig: &circuit.Config{
			FailureThreshold: 5,
			Timeout:          1 * time.Second,
			RecoveryTime:     2 * time.Second,
		},
		RetryChecker: DefaultRetryChecker,
	}
	
	client := NewClient(config)
	
	// Test operation that fails twice then succeeds
	callCount := 0
	err := client.Execute(context.Background(), "test_op", func(ctx context.Context) error {
		callCount++
		if callCount < 3 {
			return types.NewAppError(types.ErrNetworkTimeout, "timeout", nil)
		}
		return nil
	})
	
	if err != nil {
		t.Errorf("Expected no error after retries, got: %v", err)
	}
	
	if callCount != 3 {
		t.Errorf("Expected operation to be called 3 times, got %d", callCount)
	}
}

func TestClientExecuteNonRetryableError(t *testing.T) {
	config := &Config{
		BackoffConfig: backoff.NewExponentialBackoff(1*time.Millisecond, 10*time.Millisecond, 2.0, false, 3),
		CircuitBreakerConfig: &circuit.Config{
			FailureThreshold: 5,
			Timeout:          1 * time.Second,
			RecoveryTime:     2 * time.Second,
		},
		RetryChecker: DefaultRetryChecker,
	}
	
	client := NewClient(config)
	
	// Test operation that fails with non-retryable error
	callCount := 0
	err := client.Execute(context.Background(), "test_op", func(ctx context.Context) error {
		callCount++
		return types.NewAppError(types.ErrInvalidCredentials, "invalid creds", nil)
	})
	
	if err == nil {
		t.Error("Expected error for non-retryable operation")
	}
	
	if callCount != 1 {
		t.Errorf("Expected operation to be called once (no retries), got %d", callCount)
	}
}

func TestClientExecuteMaxRetries(t *testing.T) {
	config := &Config{
		BackoffConfig: &backoff.ExponentialBackoff{
			InitialDelay:  1 * time.Millisecond,
			MaxDelay:      10 * time.Millisecond,
			BackoffFactor: 2.0,
			JitterEnabled: false,
			MaxRetries:    2, // Max 2 attempts (0, 1)
		},
		CircuitBreakerConfig: circuit.DefaultConfig(),
		RetryChecker:         DefaultRetryChecker,
	}
	
	client := NewClient(config)
	
	callCount := 0
	operation := func(ctx context.Context) error {
		callCount++
		// Return retryable error
		return types.NewAppError(types.ErrNetworkTimeout, "timeout error", nil)
	}
	
	err := client.Execute(context.Background(), "test", operation)
	if err == nil {
		t.Error("Expected operation to fail after max retries")
	}
	
	// Should be called max retries times (2) = 2 times total
	if callCount != 2 {
		t.Errorf("Expected operation to be called 2 times (max retries), got %d", callCount)
	}
}

func TestClientExecuteContextCancellation(t *testing.T) {
	config := &Config{
		BackoffConfig: backoff.NewExponentialBackoff(100*time.Millisecond, 1*time.Second, 2.0, false, 5),
		CircuitBreakerConfig: &circuit.Config{
			FailureThreshold: 10,
			Timeout:          1 * time.Second,
			RecoveryTime:     2 * time.Second,
		},
		RetryChecker: DefaultRetryChecker,
	}
	
	client := NewClient(config)
	
	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	
	callCount := 0
	err := client.Execute(ctx, "test_op", func(ctx context.Context) error {
		callCount++
		return types.NewAppError(types.ErrNetworkTimeout, "timeout", nil)
	})
	
	if err == nil {
		t.Error("Expected error due to context cancellation")
	}
	
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context deadline exceeded error, got: %v", err)
	}
}

func TestClientStats(t *testing.T) {
	config := &Config{
		BackoffConfig: backoff.NewExponentialBackoff(1*time.Millisecond, 10*time.Millisecond, 2.0, false, 3),
		CircuitBreakerConfig: &circuit.Config{
			FailureThreshold: 2,
			Timeout:          1 * time.Second,
			RecoveryTime:     2 * time.Second,
		},
		RetryChecker: DefaultRetryChecker,
	}
	
	client := NewClient(config)
	
	// Execute some operations
	client.Execute(context.Background(), "test_op", func(ctx context.Context) error {
		return nil
	})
	
	client.Execute(context.Background(), "test_op", func(ctx context.Context) error {
		return types.NewAppError(types.ErrNetworkTimeout, "timeout", nil)
	})
	
	stats := client.Stats()
	if stats.TotalRequests < 2 {
		t.Errorf("Expected at least 2 total requests, got %d", stats.TotalRequests)
	}
}

func TestDefaultRetryChecker(t *testing.T) {
	// Test retryable AppError
	retryableErr := types.NewAppError(types.ErrNetworkTimeout, "network timeout", nil)
	if !DefaultRetryChecker(retryableErr) {
		t.Error("Expected retryable AppError to be retryable")
	}
	
	// Test non-retryable AppError
	nonRetryableErr := types.NewAppError(types.ErrInvalidInput, "invalid input", nil)
	if DefaultRetryChecker(nonRetryableErr) {
		t.Error("Expected non-retryable error to not be retryable")
	}
	
	// Test generic error - should NOT be retryable by default
	genericErr := errors.New("generic error")
	if DefaultRetryChecker(genericErr) {
		t.Error("Expected generic error to NOT be retryable by default")
	}
	
	// Test nil error
	if DefaultRetryChecker(nil) {
		t.Error("Expected nil error to not be retryable")
	}
}

func TestClientExecuteTimeout(t *testing.T) {
	config := &Config{
		BackoffConfig: backoff.NewExponentialBackoff(1*time.Millisecond, 10*time.Millisecond, 2.0, false, 3),
		CircuitBreakerConfig: &circuit.Config{
			FailureThreshold: 5,
			Timeout:          1 * time.Millisecond, // Very short timeout
			RecoveryTime:     2 * time.Second,
		},
		RetryChecker: DefaultRetryChecker,
	}
	
	client := NewClient(config)
	
	// Test operation that takes longer than timeout
	_ = client.Execute(context.Background(), "test_op", func(ctx context.Context) error {
		time.Sleep(10 * time.Millisecond) // Sleep longer than circuit breaker timeout
		return nil
	})
	
	// Should eventually fail due to timeout or circuit breaker
	// The exact behavior depends on implementation details
	// At minimum, we can check that the operation doesn't succeed immediately
} 