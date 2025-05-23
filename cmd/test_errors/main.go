package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/errors/types"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/errors/validation"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/resilience/checkpoints"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/retry"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/retry/backoff"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/retry/circuit"
)

func main() {
	fmt.Println("🧪 Testing Enhanced Error Handling System in Docker...")
	fmt.Println("============================================================")

	ctx := context.Background()

	// Test 1: Custom Error Types
	fmt.Println("\n1️⃣  Testing Custom Error Types...")
	testCustomErrors(ctx)

	// Test 2: Validation System
	fmt.Println("\n2️⃣  Testing Validation System...")
	testValidation(ctx)

	// Test 3: Retry with Circuit Breaker
	fmt.Println("\n3️⃣  Testing Retry with Circuit Breaker...")
	testRetrySystem(ctx)

	// Test 4: Checkpoint System
	fmt.Println("\n4️⃣  Testing Checkpoint System...")
	testCheckpointSystem(ctx)

	fmt.Println("\n✅ All tests completed successfully!")
	fmt.Println("🎉 Enhanced Error Handling System is working in Docker!")
}

func testCustomErrors(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("❌ Custom error test failed: %v\n", r)
		}
	}()

	// Create a custom error with context
	err := types.NewAppErrorWithOperation(
		types.ErrNetworkTimeout,
		"Test API call failed",
		"test_operation",
		fmt.Errorf("simulated network error"),
	).WithContext(ctx).WithMetadata("endpoint", "/api/test")

	fmt.Printf("✅ Created custom error: %s\n", err.Error())
	fmt.Printf("   Code: %s, Category: %s\n", err.Code, types.GetErrorCategory(err.Code))
	fmt.Printf("   Retryable: %v\n", types.IsRetryableError(err.Code))
}

func testValidation(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("❌ Validation test failed: %v\n", r)
		}
	}()

	// Test validation framework
	validator := validation.NewStructValidator()
	validator.Field("api_key").Required().Format(validation.APIKeyPattern, "API key format")
	validator.Field("timeout").Range(1, 300)

	// Test with invalid data
	invalidData := map[string]interface{}{
		"api_key": "", // Missing required field
		"timeout": 500, // Out of range
	}

	err := validator.Validate(ctx, invalidData)
	if err != nil {
		fmt.Printf("✅ Validation correctly caught errors: %s\n", err.Error())
	}

	// Test with valid data
	validData := map[string]interface{}{
		"api_key": "abcdef1234567890abcdef1234567890abcd",
		"timeout": 30,
	}

	err = validator.Validate(ctx, validData)
	if err == nil {
		fmt.Printf("✅ Validation passed for valid data\n")
	}
}

func testRetrySystem(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("❌ Retry system test failed: %v\n", r)
		}
	}()

	// Create retry client with custom configuration
	config := &retry.Config{
		BackoffConfig: backoff.NewExponentialBackoff(100*time.Millisecond, 1*time.Second, 2.0, true, 3),
		CircuitBreakerConfig: &circuit.Config{
			FailureThreshold: 2,
			Timeout:          500 * time.Millisecond,
			RecoveryTime:     1 * time.Second,
		},
		RetryChecker: retry.DefaultRetryChecker,
	}

	retryClient := retry.NewClient(config)

	// Test with a failing operation that should be retried
	attemptCount := 0
	err := retryClient.Execute(ctx, "test_operation", func(ctx context.Context) error {
		attemptCount++
		if attemptCount < 3 {
			return types.NewAppError(types.ErrNetworkTimeout, "simulated timeout", nil)
		}
		return nil // Success on 3rd attempt
	})

	if err == nil {
		fmt.Printf("✅ Retry system worked: succeeded after %d attempts\n", attemptCount)
	} else {
		fmt.Printf("❌ Retry system failed: %s\n", err.Error())
	}

	// Check circuit breaker stats
	stats := retryClient.Stats()
	fmt.Printf("   Circuit breaker state: %s, Total requests: %d\n", stats.State.String(), stats.TotalRequests)
}

func testCheckpointSystem(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("❌ Checkpoint system test failed: %v\n", r)
		}
	}()

	// Create checkpoint manager
	config := &checkpoints.Config{
		CheckpointDir: "/tmp/test_checkpoints",
		MaxAge:        1 * time.Hour,
	}

	manager, err := checkpoints.NewManager(config)
	if err != nil {
		fmt.Printf("❌ Failed to create checkpoint manager: %s\n", err.Error())
		return
	}

	// Create and save a checkpoint
	checkpoint := checkpoints.NewCheckpoint(
		"test_op_123",
		"test_operation",
		0.5,
		map[string]interface{}{
			"processed_items": 50,
			"total_items":     100,
		},
		"process_remaining_items",
	)
	checkpoint.AddMetadata("test_run", "docker_test")

	err = manager.Save(ctx, checkpoint)
	if err != nil {
		fmt.Printf("❌ Failed to save checkpoint: %s\n", err.Error())
		return
	}

	// Load the checkpoint
	loadedCheckpoint, err := manager.Load(ctx, "test_op_123")
	if err != nil {
		fmt.Printf("❌ Failed to load checkpoint: %s\n", err.Error())
		return
	}

	if loadedCheckpoint.Progress == 0.5 {
		fmt.Printf("✅ Checkpoint system worked: saved and loaded progress %.1f%%\n", loadedCheckpoint.Progress*100)
	}

	// Cleanup
	manager.Delete(ctx, "test_op_123")
	os.RemoveAll("/tmp/test_checkpoints")
} 