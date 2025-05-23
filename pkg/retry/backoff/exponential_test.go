package backoff

import (
	"testing"
	"time"
)

func TestNewExponentialBackoff(t *testing.T) {
	initialDelay := 100 * time.Millisecond
	maxDelay := 5 * time.Second
	backoffFactor := 2.0
	jitter := true
	maxRetries := 5
	
	eb := NewExponentialBackoff(initialDelay, maxDelay, backoffFactor, jitter, maxRetries)
	
	if eb.InitialDelay != initialDelay {
		t.Errorf("Expected initial delay %v, got %v", initialDelay, eb.InitialDelay)
	}
	
	if eb.MaxDelay != maxDelay {
		t.Errorf("Expected max delay %v, got %v", maxDelay, eb.MaxDelay)
	}
	
	if eb.BackoffFactor != backoffFactor {
		t.Errorf("Expected backoff factor %f, got %f", backoffFactor, eb.BackoffFactor)
	}
	
	if eb.MaxRetries != maxRetries {
		t.Errorf("Expected max retries %d, got %d", maxRetries, eb.MaxRetries)
	}
}

func TestCalculateDelay(t *testing.T) {
	eb := NewExponentialBackoff(100*time.Millisecond, 5*time.Second, 2.0, false, 5)
	
	// Test first delay (attempt 0)
	delay := eb.CalculateDelay(0)
	if delay != 100*time.Millisecond {
		t.Errorf("Expected first delay to be 100ms, got %v", delay)
	}
	
	// Test second delay (attempt 1)
	delay = eb.CalculateDelay(1)
	if delay != 200*time.Millisecond {
		t.Errorf("Expected second delay to be 200ms, got %v", delay)
	}
	
	// Test third delay (attempt 2)
	delay = eb.CalculateDelay(2)
	if delay != 400*time.Millisecond {
		t.Errorf("Expected third delay to be 400ms, got %v", delay)
	}
}

func TestCalculateDelayMaxCap(t *testing.T) {
	eb := NewExponentialBackoff(1*time.Second, 2*time.Second, 2.0, false, 10)
	
	// Test that delay is capped at maxDelay
	delay := eb.CalculateDelay(5) // Should exceed maxDelay
	if delay != 2*time.Second {
		t.Errorf("Expected delay to be capped at 2s, got %v", delay)
	}
}

func TestCalculateDelayWithJitter(t *testing.T) {
	eb := NewExponentialBackoff(100*time.Millisecond, 5*time.Second, 2.0, true, 5)
	
	// Test that jitter adds randomness (delay should vary)
	delay1 := eb.CalculateDelay(1)
	delay2 := eb.CalculateDelay(1)
	
	// Both should be around 200ms but may differ due to jitter
	expectedBase := 200 * time.Millisecond
	
	// Check that delays are within reasonable bounds (±50% for jitter)
	if delay1 < expectedBase/2 || delay1 > expectedBase*3/2 {
		t.Errorf("Delay1 %v is outside expected jitter range around %v", delay1, expectedBase)
	}
	
	if delay2 < expectedBase/2 || delay2 > expectedBase*3/2 {
		t.Errorf("Delay2 %v is outside expected jitter range around %v", delay2, expectedBase)
	}
}

func TestShouldRetry(t *testing.T) {
	eb := NewExponentialBackoff(100*time.Millisecond, 5*time.Second, 2.0, false, 3)
	
	// Should retry for attempts within limit
	if !eb.ShouldRetry(0) {
		t.Error("Expected to retry on attempt 0")
	}
	
	if !eb.ShouldRetry(1) {
		t.Error("Expected to retry on attempt 1")
	}
	
	if !eb.ShouldRetry(2) {
		t.Error("Expected to retry on attempt 2")
	}
	
	// Should not retry when max retries exceeded
	if eb.ShouldRetry(3) {
		t.Error("Expected not to retry on attempt 3 (exceeds max)")
	}
	
	if eb.ShouldRetry(5) {
		t.Error("Expected not to retry on attempt 5 (exceeds max)")
	}
}

func TestZeroValues(t *testing.T) {
	eb := NewExponentialBackoff(0, 0, 0, false, 0)
	
	// Should handle zero values gracefully
	delay := eb.CalculateDelay(0)
	if delay < 0 {
		t.Errorf("Expected non-negative delay, got %v", delay)
	}
	
	// Should not retry with max retries = 0
	if eb.ShouldRetry(0) {
		t.Error("Expected not to retry with max retries = 0")
	}
} 