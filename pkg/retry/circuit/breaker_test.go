package circuit

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewCircuitBreaker(t *testing.T) {
	config := &Config{
		FailureThreshold: 5,
		Timeout:          1 * time.Second,
		RecoveryTime:     2 * time.Second,
	}
	
	cb := NewCircuitBreaker(config)
	if cb == nil {
		t.Error("Expected circuit breaker to be created")
	}
	
	if cb.State() != StateClosed {
		t.Errorf("Expected initial state to be CLOSED, got %s", cb.State())
	}
}

func TestCircuitBreakerClosedState(t *testing.T) {
	config := &Config{
		FailureThreshold: 3,
		Timeout:          100 * time.Millisecond,
		RecoveryTime:     200 * time.Millisecond,
	}
	
	cb := NewCircuitBreaker(config)
	
	// Execute successful operations
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil // Success
	})
	
	if err != nil {
		t.Errorf("Expected successful execution, got error: %v", err)
	}
	
	// Should still be closed
	if cb.State() != StateClosed {
		t.Errorf("Expected state to remain CLOSED after success, got %s", cb.State())
	}
}

func TestCircuitBreakerOpenState(t *testing.T) {
	config := &Config{
		FailureThreshold: 2,
		Timeout:          100 * time.Millisecond,
		RecoveryTime:     200 * time.Millisecond,
	}
	
	cb := NewCircuitBreaker(config)
	
	// Execute failing operations to trigger open state
	testErr := errors.New("test error")
	
	for i := 0; i < 3; i++ {
		cb.Execute(context.Background(), func(ctx context.Context) error {
			return testErr
		})
	}
	
	// Should be open now
	if cb.State() != StateOpen {
		t.Errorf("Expected state to be OPEN after failures, got %s", cb.State())
	}
	
	// Should reject new calls
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})
	
	if err != ErrCircuitBreakerOpen {
		t.Errorf("Expected circuit breaker open error, got: %v", err)
	}
}

func TestCircuitBreakerRecovery(t *testing.T) {
	config := &Config{
		FailureThreshold: 2,
		Timeout:          50 * time.Millisecond,
		RecoveryTime:     100 * time.Millisecond,
	}
	
	cb := NewCircuitBreaker(config)
	
	// Trigger open state
	testErr := errors.New("test error")
	for i := 0; i < 3; i++ {
		cb.Execute(context.Background(), func(ctx context.Context) error {
			return testErr
		})
	}
	
	// Wait for recovery time
	time.Sleep(150 * time.Millisecond)
	
	// Execute successful operation
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil // Success
	})
	
	if err != nil {
		t.Errorf("Expected successful execution after recovery, got: %v", err)
	}
	
	// Should transition back to closed
	if cb.State() != StateClosed {
		t.Errorf("Expected state to be CLOSED after successful recovery, got %s", cb.State())
	}
}

func TestCircuitBreakerStats(t *testing.T) {
	config := &Config{
		FailureThreshold: 5,
		Timeout:          100 * time.Millisecond,
		RecoveryTime:     200 * time.Millisecond,
	}
	
	cb := NewCircuitBreaker(config)
	
	// Execute some operations
	cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil // Success
	})
	
	cb.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("failure")
	})
	
	stats := cb.Stats()
	if stats.TotalRequests < 2 {
		t.Errorf("Expected at least 2 total requests, got %d", stats.TotalRequests)
	}
	
	if stats.FailedRequests < 1 {
		t.Errorf("Expected at least 1 failure, got %d", stats.FailedRequests)
	}
}

func TestCircuitBreakerTimeout(t *testing.T) {
	config := &Config{
		FailureThreshold: 5,
		Timeout:          10 * time.Millisecond, // Very short timeout
		RecoveryTime:     100 * time.Millisecond,
	}
	
	cb := NewCircuitBreaker(config)
	
	// Execute operation that takes longer than timeout
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		time.Sleep(20 * time.Millisecond) // Sleep longer than timeout
		return nil
	})
	
	if err != ErrTimeout {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

func TestCircuitBreakerReset(t *testing.T) {
	config := &Config{
		FailureThreshold: 2,
		Timeout:          100 * time.Millisecond,
		RecoveryTime:     200 * time.Millisecond,
	}
	
	cb := NewCircuitBreaker(config)
	
	// Trigger open state
	testErr := errors.New("test error")
	for i := 0; i < 3; i++ {
		cb.Execute(context.Background(), func(ctx context.Context) error {
			return testErr
		})
	}
	
	// Reset the circuit breaker
	cb.Reset()
	
	// Should be closed again
	if cb.State() != StateClosed {
		t.Errorf("Expected state to be CLOSED after reset, got %s", cb.State())
	}
	
	// Should allow execution
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})
	
	if err != nil {
		t.Errorf("Expected successful execution after reset, got: %v", err)
	}
} 