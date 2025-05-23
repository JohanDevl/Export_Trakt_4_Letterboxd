package circuit

import (
	"context"
	"errors"
	"sync"
	"time"
)

// State represents the circuit breaker state
type State int

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateHalfOpen:
		return "HALF_OPEN"
	case StateOpen:
		return "OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu               sync.RWMutex
	state            State
	failureThreshold int
	timeout          time.Duration
	recoveryTime     time.Duration
	lastFailureTime  time.Time
	consecutiveFailures int
	totalRequests    int64
	successfulRequests int64
	failedRequests   int64
}

// Config represents circuit breaker configuration
type Config struct {
	FailureThreshold int
	Timeout          time.Duration
	RecoveryTime     time.Duration
}

// DefaultConfig returns default circuit breaker configuration
func DefaultConfig() *Config {
	return &Config{
		FailureThreshold: 5,
		Timeout:          30 * time.Second,
		RecoveryTime:     60 * time.Second,
	}
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config *Config) *CircuitBreaker {
	if config == nil {
		config = DefaultConfig()
	}
	
	return &CircuitBreaker{
		state:            StateClosed,
		failureThreshold: config.FailureThreshold,
		timeout:          config.Timeout,
		recoveryTime:     config.RecoveryTime,
	}
}

var (
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
	ErrTimeout            = errors.New("operation timeout")
)

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(context.Context) error) error {
	// Check if circuit breaker allows the request
	if !cb.allowRequest() {
		return ErrCircuitBreakerOpen
	}
	
	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, cb.timeout)
	defer cancel()
	
	// Execute the function
	done := make(chan error, 1)
	go func() {
		done <- fn(timeoutCtx)
	}()
	
	select {
	case err := <-done:
		cb.recordResult(err == nil)
		return err
	case <-timeoutCtx.Done():
		cb.recordResult(false)
		return ErrTimeout
	}
}

// allowRequest checks if the circuit breaker allows the request
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	now := time.Now()
	
	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if recovery time has passed
		if now.Sub(cb.lastFailureTime) >= cb.recoveryTime {
			cb.state = StateHalfOpen
			cb.consecutiveFailures = 0
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// recordResult records the result of an operation
func (cb *CircuitBreaker) recordResult(success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	cb.totalRequests++
	
	if success {
		cb.successfulRequests++
		cb.consecutiveFailures = 0
		
		// If in half-open state and success, move to closed
		if cb.state == StateHalfOpen {
			cb.state = StateClosed
		}
	} else {
		cb.failedRequests++
		cb.consecutiveFailures++
		cb.lastFailureTime = time.Now()
		
		// Check if failure threshold is reached
		if cb.consecutiveFailures >= cb.failureThreshold {
			cb.state = StateOpen
		}
	}
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Stats returns the current statistics
func (cb *CircuitBreaker) Stats() Stats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	
	return Stats{
		State:               cb.state,
		TotalRequests:       cb.totalRequests,
		SuccessfulRequests:  cb.successfulRequests,
		FailedRequests:      cb.failedRequests,
		ConsecutiveFailures: cb.consecutiveFailures,
		LastFailureTime:     cb.lastFailureTime,
	}
}

// Stats represents circuit breaker statistics
type Stats struct {
	State               State
	TotalRequests       int64
	SuccessfulRequests  int64
	FailedRequests      int64
	ConsecutiveFailures int
	LastFailureTime     time.Time
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	cb.state = StateClosed
	cb.consecutiveFailures = 0
	cb.totalRequests = 0
	cb.successfulRequests = 0
	cb.failedRequests = 0
	cb.lastFailureTime = time.Time{}
} 