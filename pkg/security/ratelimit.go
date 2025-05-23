package security

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/security/audit"
)

// RateLimiter provides rate limiting functionality for API calls and operations
type RateLimiter struct {
	limits     map[string]*bucketLimiter
	mu         sync.RWMutex
	auditLog   *audit.Logger
	config     RateLimitConfig
}

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	Enabled         bool          `toml:"enabled"`
	DefaultLimit    int           `toml:"default_limit"`     // requests per minute
	BurstLimit      int           `toml:"burst_limit"`       // burst capacity
	WindowDuration  time.Duration `toml:"window_duration"`   // rate limit window
	CleanupInterval time.Duration `toml:"cleanup_interval"`  // cleanup expired entries
	Limits          map[string]RateLimit `toml:"limits"` // per-service limits
}

// RateLimit defines rate limiting parameters for a specific service
type RateLimit struct {
	RequestsPerMinute int           `toml:"requests_per_minute"`
	BurstCapacity     int           `toml:"burst_capacity"`
	Window            time.Duration `toml:"window"`
}

// bucketLimiter implements a token bucket rate limiter
type bucketLimiter struct {
	tokens          float64
	capacity        float64
	refillRate      float64 // tokens per second
	lastRefill      time.Time
	mu              sync.Mutex
}

// DefaultRateLimitConfig returns secure default rate limiting configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Enabled:         true,
		DefaultLimit:    60,  // 60 requests per minute
		BurstLimit:      10,  // 10 burst capacity
		WindowDuration:  time.Minute,
		CleanupInterval: time.Minute * 5,
		Limits: map[string]RateLimit{
			"trakt_api": {
				RequestsPerMinute: 40,    // Conservative Trakt.tv limit
				BurstCapacity:     5,
				Window:           time.Minute,
			},
			"export": {
				RequestsPerMinute: 120,   // Higher limit for exports
				BurstCapacity:     20,
				Window:           time.Minute,
			},
			"auth": {
				RequestsPerMinute: 10,    // Strict limit for auth operations
				BurstCapacity:     3,
				Window:           time.Minute,
			},
		},
	}
}

// NewRateLimiter creates a new rate limiter with the given configuration
func NewRateLimiter(config RateLimitConfig, auditLog *audit.Logger) *RateLimiter {
	rl := &RateLimiter{
		limits:   make(map[string]*bucketLimiter),
		auditLog: auditLog,
		config:   config,
	}

	// Start cleanup goroutine if enabled
	if config.Enabled && config.CleanupInterval > 0 {
		go rl.cleanupRoutine(config.CleanupInterval)
	}

	return rl
}

// Allow checks if a request is allowed for the given service
func (rl *RateLimiter) Allow(service string) bool {
	if !rl.config.Enabled {
		return true
	}

	rl.mu.Lock()
	limiter := rl.getLimiter(service)
	rl.mu.Unlock()

	allowed := limiter.allow()
	
	if !allowed {
		rl.logRateLimitHit(service)
	}

	return allowed
}

// AllowN checks if N requests are allowed for the given service
func (rl *RateLimiter) AllowN(service string, n int) bool {
	if !rl.config.Enabled {
		return true
	}

	rl.mu.Lock()
	limiter := rl.getLimiter(service)
	rl.mu.Unlock()

	allowed := limiter.allowN(n)
	
	if !allowed {
		rl.logRateLimitHit(service)
	}

	return allowed
}

// Wait blocks until a request is allowed for the given service
func (rl *RateLimiter) Wait(ctx context.Context, service string) error {
	if !rl.config.Enabled {
		return nil
	}

	ticker := time.NewTicker(time.Millisecond * 100)
	defer ticker.Stop()

	for {
		if rl.Allow(service) {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Continue checking
		}
	}
}

// GetStats returns current rate limiting statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["enabled"] = rl.config.Enabled
	stats["total_limiters"] = len(rl.limits)
	
	servicStats := make(map[string]interface{})
	for service, limiter := range rl.limits {
		limiter.mu.Lock()
		servicStats[service] = map[string]interface{}{
			"tokens":       limiter.tokens,
			"capacity":     limiter.capacity,
			"refill_rate":  limiter.refillRate,
			"last_refill":  limiter.lastRefill,
		}
		limiter.mu.Unlock()
	}
	stats["services"] = servicStats

	return stats
}

// getLimiter gets or creates a rate limiter for the given service
func (rl *RateLimiter) getLimiter(service string) *bucketLimiter {
	limiter, exists := rl.limits[service]
	if !exists {
		limiter = rl.createLimiter(service)
		rl.limits[service] = limiter
	}
	return limiter
}

// createLimiter creates a new bucket limiter for a service
func (rl *RateLimiter) createLimiter(service string) *bucketLimiter {
	var limit RateLimit
	var exists bool
	
	if limit, exists = rl.config.Limits[service]; !exists {
		// Use default limits
		window := rl.config.WindowDuration
		if window == 0 {
			window = time.Minute // Default to 1 minute if not configured
		}
		
		limit = RateLimit{
			RequestsPerMinute: rl.config.DefaultLimit,
			BurstCapacity:     rl.config.BurstLimit,
			Window:           window,
		}
	}

	// Ensure Window is not zero to avoid division by zero
	if limit.Window == 0 {
		limit.Window = time.Minute
	}

	capacity := float64(limit.BurstCapacity)
	refillRate := float64(limit.RequestsPerMinute) / limit.Window.Seconds()

	return &bucketLimiter{
		tokens:     capacity,
		capacity:   capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// allow checks if one request is allowed
func (bl *bucketLimiter) allow() bool {
	return bl.allowN(1)
}

// allowN checks if N requests are allowed
func (bl *bucketLimiter) allowN(n int) bool {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	now := time.Now()
	bl.refill(now)

	if bl.tokens >= float64(n) {
		bl.tokens -= float64(n)
		return true
	}

	return false
}

// refill adds tokens to the bucket based on elapsed time
func (bl *bucketLimiter) refill(now time.Time) {
	elapsed := now.Sub(bl.lastRefill).Seconds()
	bl.tokens = min(bl.capacity, bl.tokens+(elapsed*bl.refillRate))
	bl.lastRefill = now
}

// logRateLimitHit logs when rate limit is hit
func (rl *RateLimiter) logRateLimitHit(service string) {
	if rl.auditLog != nil {
		rl.auditLog.LogEvent(audit.AuditEvent{
			EventType: audit.RateLimitHit,
			Severity:  audit.SeverityMedium,
			Source:    "rate_limiter",
			Action:    "limit_exceeded",
			Result:    "denied",
			Message:   fmt.Sprintf("Rate limit exceeded for service: %s", service),
			Details: map[string]interface{}{
				"service": service,
			},
		})
	}
}

// cleanupRoutine periodically cleans up unused rate limiters
func (rl *RateLimiter) cleanupRoutine(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

// cleanup removes unused rate limiters
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.config.CleanupInterval * 2) // Keep limiters that were used recently

	for service, limiter := range rl.limits {
		limiter.mu.Lock()
		lastUsed := limiter.lastRefill
		limiter.mu.Unlock()

		if lastUsed.Before(cutoff) {
			delete(rl.limits, service)
		}
	}
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
} 