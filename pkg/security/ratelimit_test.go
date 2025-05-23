package security

import (
	"context"
	"testing"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/security/audit"
)

func TestRateLimiter_Allow(t *testing.T) {
	config := RateLimitConfig{
		Enabled:      true,
		DefaultLimit: 10, // 10 requests per minute
		BurstLimit:   5,  // 5 burst capacity
		Limits: map[string]RateLimit{
			"test_service": {
				RequestsPerMinute: 10,
				BurstCapacity:     5,
				Window:           time.Minute,
			},
		},
	}

	rl := NewRateLimiter(config, nil)

	// Test initial burst capacity
	for i := 0; i < 5; i++ {
		if !rl.Allow("test_service") {
			t.Errorf("Request %d should be allowed within burst capacity", i+1)
		}
	}

	// Next request should be denied (burst exhausted)
	if rl.Allow("test_service") {
		t.Error("Request should be denied after burst capacity exhausted")
	}
}

func TestRateLimiter_AllowN(t *testing.T) {
	config := RateLimitConfig{
		Enabled:      true,
		DefaultLimit: 60,
		BurstLimit:   10,
	}

	rl := NewRateLimiter(config, nil)

	// Test allowing multiple requests at once
	if !rl.AllowN("test_service", 5) {
		t.Error("Should allow 5 requests within burst capacity")
	}

	if !rl.AllowN("test_service", 5) {
		t.Error("Should allow another 5 requests within burst capacity")
	}

	// Should not allow more requests (burst exhausted)
	if rl.AllowN("test_service", 1) {
		t.Error("Should not allow request when burst capacity exhausted")
	}
}

func TestRateLimiter_Wait(t *testing.T) {
	config := RateLimitConfig{
		Enabled:      true,
		DefaultLimit: 60, // 1 request per second
		BurstLimit:   1,  // 1 burst capacity
	}

	rl := NewRateLimiter(config, nil)

	// First request should be immediate
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	if err := rl.Wait(ctx, "test_service"); err != nil {
		t.Errorf("First request should be immediate: %v", err)
	}

	// Second request should timeout (no tokens available)
	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel2()

	if err := rl.Wait(ctx2, "test_service"); err == nil {
		t.Error("Second request should timeout")
	}
}

func TestRateLimiter_DisabledConfig(t *testing.T) {
	config := RateLimitConfig{
		Enabled: false,
	}

	rl := NewRateLimiter(config, nil)

	// All requests should be allowed when disabled
	for i := 0; i < 100; i++ {
		if !rl.Allow("test_service") {
			t.Errorf("Request %d should be allowed when rate limiting is disabled", i+1)
		}
	}
}

func TestRateLimiter_DefaultLimits(t *testing.T) {
	config := RateLimitConfig{
		Enabled:      true,
		DefaultLimit: 5,
		BurstLimit:   2,
	}

	rl := NewRateLimiter(config, nil)

	// Use a service not defined in specific limits
	for i := 0; i < 2; i++ {
		if !rl.Allow("unknown_service") {
			t.Errorf("Request %d should be allowed within default burst capacity", i+1)
		}
	}

	// Next request should be denied
	if rl.Allow("unknown_service") {
		t.Error("Request should be denied after default burst capacity exhausted")
	}
}

func TestRateLimiter_WithAuditLogging(t *testing.T) {
	// Create audit logger for testing
	auditConfig := audit.Config{
		LogLevel:     "info",
		OutputFormat: "json",
	}
	auditLogger, err := audit.NewLogger(auditConfig)
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer auditLogger.Close()

	config := RateLimitConfig{
		Enabled:      true,
		DefaultLimit: 1,
		BurstLimit:   1,
	}

	rl := NewRateLimiter(config, auditLogger)

	// Exhaust the limit
	rl.Allow("test_service")

	// This should trigger audit logging
	if rl.Allow("test_service") {
		t.Error("Request should be denied and logged")
	}

	// Note: In a real test, you'd capture and verify the audit log output
}

func TestRateLimiter_GetStats(t *testing.T) {
	config := RateLimitConfig{
		Enabled:      true,
		DefaultLimit: 10,
		BurstLimit:   5,
	}

	rl := NewRateLimiter(config, nil)

	// Make some requests to populate stats
	rl.Allow("service1")
	rl.Allow("service2")

	stats := rl.GetStats()

	if !stats["enabled"].(bool) {
		t.Error("Stats should show rate limiting as enabled")
	}

	if stats["total_limiters"].(int) != 2 {
		t.Errorf("Expected 2 limiters, got %d", stats["total_limiters"].(int))
	}

	services := stats["services"].(map[string]interface{})
	if len(services) != 2 {
		t.Errorf("Expected 2 service stats, got %d", len(services))
	}
}

func TestBucketLimiter_Refill(t *testing.T) {
	limiter := &bucketLimiter{
		tokens:     0,
		capacity:   10,
		refillRate: 1, // 1 token per second
		lastRefill: time.Now().Add(-time.Second * 5), // 5 seconds ago
	}

	now := time.Now()
	limiter.refill(now)

	// Should have refilled approximately 5 tokens
	if limiter.tokens < 4.9 || limiter.tokens > 5.1 {
		t.Errorf("Expected approximately 5 tokens after refill, got %f", limiter.tokens)
	}

	if limiter.lastRefill != now {
		t.Error("Last refill time should be updated")
	}
}

func TestRateLimiter_SpecificServiceLimits(t *testing.T) {
	config := RateLimitConfig{
		Enabled:      true,
		DefaultLimit: 10,
		BurstLimit:   5,
		Limits: map[string]RateLimit{
			"strict_service": {
				RequestsPerMinute: 2,
				BurstCapacity:     1,
				Window:           time.Minute,
			},
		},
	}

	rl := NewRateLimiter(config, nil)

	// Strict service should only allow 1 request
	if !rl.Allow("strict_service") {
		t.Error("First request to strict service should be allowed")
	}

	if rl.Allow("strict_service") {
		t.Error("Second request to strict service should be denied")
	}

	// Default service should allow more
	for i := 0; i < 5; i++ {
		if !rl.Allow("default_service") {
			t.Errorf("Request %d to default service should be allowed", i+1)
		}
	}
} 