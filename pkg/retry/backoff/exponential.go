package backoff

import (
	"math"
	"math/rand"
	"time"
)

// ExponentialBackoff implements exponential backoff with jitter
type ExponentialBackoff struct {
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
	JitterEnabled bool
	MaxRetries    int
}

// DefaultExponentialBackoff returns a default exponential backoff configuration
func DefaultExponentialBackoff() *ExponentialBackoff {
	return &ExponentialBackoff{
		InitialDelay:  1 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		JitterEnabled: true,
		MaxRetries:    3,
	}
}

// NewExponentialBackoff creates a new exponential backoff with custom settings
func NewExponentialBackoff(initialDelay, maxDelay time.Duration, factor float64, jitter bool, maxRetries int) *ExponentialBackoff {
	return &ExponentialBackoff{
		InitialDelay:  initialDelay,
		MaxDelay:      maxDelay,
		BackoffFactor: factor,
		JitterEnabled: jitter,
		MaxRetries:    maxRetries,
	}
}

// CalculateDelay calculates the delay for a given attempt
func (eb *ExponentialBackoff) CalculateDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return eb.InitialDelay
	}
	
	// Calculate exponential delay
	delay := float64(eb.InitialDelay) * math.Pow(eb.BackoffFactor, float64(attempt))
	
	// Apply maximum delay limit
	if time.Duration(delay) > eb.MaxDelay {
		delay = float64(eb.MaxDelay)
	}
	
	// Apply jitter if enabled
	if eb.JitterEnabled {
		jitter := rand.Float64() * 0.1 * delay // 10% jitter
		delay = delay + jitter
	}
	
	return time.Duration(delay)
}

// ShouldRetry determines if a retry should be attempted
func (eb *ExponentialBackoff) ShouldRetry(attempt int) bool {
	return attempt < eb.MaxRetries
}

// LinearBackoff implements linear backoff
type LinearBackoff struct {
	InitialDelay  time.Duration
	DelayStep     time.Duration
	MaxDelay      time.Duration
	JitterEnabled bool
	MaxRetries    int
}

// NewLinearBackoff creates a new linear backoff
func NewLinearBackoff(initialDelay, delayStep, maxDelay time.Duration, jitter bool, maxRetries int) *LinearBackoff {
	return &LinearBackoff{
		InitialDelay:  initialDelay,
		DelayStep:     delayStep,
		MaxDelay:      maxDelay,
		JitterEnabled: jitter,
		MaxRetries:    maxRetries,
	}
}

// CalculateDelay calculates the delay for a given attempt (linear)
func (lb *LinearBackoff) CalculateDelay(attempt int) time.Duration {
	delay := lb.InitialDelay + time.Duration(attempt)*lb.DelayStep
	
	// Apply maximum delay limit
	if delay > lb.MaxDelay {
		delay = lb.MaxDelay
	}
	
	// Apply jitter if enabled
	if lb.JitterEnabled {
		jitterAmount := float64(delay) * 0.1 * rand.Float64() // 10% jitter
		delay = delay + time.Duration(jitterAmount)
	}
	
	return delay
}

// ShouldRetry determines if a retry should be attempted
func (lb *LinearBackoff) ShouldRetry(attempt int) bool {
	return attempt < lb.MaxRetries
}

// Backoff interface for different backoff strategies
type Backoff interface {
	CalculateDelay(attempt int) time.Duration
	ShouldRetry(attempt int) bool
} 