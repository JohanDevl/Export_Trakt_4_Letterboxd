package handlers

import "time"

// CategoryRetryConfig defines retry configuration per error category
type CategoryRetryConfig struct {
	MaxAttempts   int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
	EnableJitter  bool
	CircuitBreaker bool
}