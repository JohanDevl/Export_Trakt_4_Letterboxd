# TestUtils Package

The `testutils` package provides centralized testing utilities for the Export Trakt 4 Letterboxd project, eliminating code duplication and providing consistent testing patterns across all packages.

## Overview

This package contains:
- **Mock implementations** for common interfaces
- **Test helpers** for file/directory operations, environment setup, and assertions
- **Job utilities** for testing concurrent operations and worker pools
- **Sample data fixtures** for consistent test data across the project

## Mock Implementations

### MockLogger

Provides multiple logger implementations for different testing needs:

```go
// Full mock with testify for expectations
logger := testutils.NewMockLogger()
logger.On("Info", "test.message", mock.Anything).Return()
logger.Info("test.message", map[string]interface{}{"key": "value"})
logger.AssertExpectations(t)

// No-op logger for tests that don't need logging verification
logger := testutils.NewNoOpLogger()

// Capturing logger for inspecting log messages
logger := testutils.NewCapturingLogger()
logger.Info("test", map[string]interface{}{"data": "value"})
messages := logger.GetMessagesByLevel("info")
```

### MockTokenManager

Mock implementation for OAuth token management:

```go
tokenMgr := testutils.NewMockTokenManager()
tokenMgr.SetToken("custom_token")
token, err := tokenMgr.GetValidAccessToken() // Returns "custom_token"

// Test error conditions
tokenMgr.SetError(errors.New("auth failed"))
_, err := tokenMgr.GetValidAccessToken() // Returns error
```

### MockTranslator

Mock implementation for i18n translation:

```go
translator := testutils.NewMockTranslator()
translator.SetTranslation("hello", "bonjour")
result := translator.Translate("hello", nil) // Returns "bonjour"
```

### MockMetricsRecorder

Mock implementation for performance metrics:

```go
metrics := testutils.NewMockMetricsRecorder()
metrics.IncrementJobsProcessed()
metrics.RecordJobDuration(100 * time.Millisecond)
assert.Equal(t, int64(1), metrics.GetJobsProcessed())
```

## Test Helpers

### File and Directory Operations

```go
// Create temporary directory with automatic cleanup
dir, cleanup := testutils.TempDir(t)
defer cleanup()

// Create temporary file with content
path, cleanup := testutils.TempFile(t, "test content")
defer cleanup()

// Create test CSV files
path := testutils.CreateTestCSV(t, dir, "test.csv", 
    []string{"Title", "Year"}, 
    [][]string{{"Inception", "2010"}})
```

### Environment Management

```go
// Set environment variable for test duration
cleanup := testutils.SetEnv(t, "TEST_VAR", "value")
defer cleanup()

// Temporarily unset environment variable
cleanup := testutils.UnsetEnv(t, "EXISTING_VAR")
defer cleanup()
```

### Configuration Helpers

```go
// Create test configuration
cfg := testutils.TestConfig()

// Create config with specific export directory
cfg := testutils.TestConfigWithExportDir("/custom/path")

// Create minimal configuration
cfg := testutils.TestConfigMinimal()
```

### Test Assertions

```go
// Assert condition becomes true within timeout
testutils.AssertEventually(t, func() bool {
    return someCondition()
}, time.Second, "condition should become true")

// Timeout wrapper for test functions
testutils.WithTestTimeout(t, 5*time.Second, func() {
    // Test code that should complete within 5 seconds
})
```

## Job Testing Utilities

### Test Jobs

```go
// Simple job that sleeps for specified duration
job := testutils.NewSimpleTestJob("job1", 100*time.Millisecond)

// Job that returns an error
job := testutils.NewErrorTestJob("error_job", "test error")

// Job that panics
job := testutils.NewPanicTestJob("panic_job", "test panic")

// Job that waits for signal
waitJob := testutils.NewWaitJob("wait_job")
go func() {
    waitJob.Execute(ctx)
}()
// Later...
waitJob.Signal() // Allow job to complete
```

### Job Groups and Batching

```go
// Create multiple test jobs
jobs := testutils.CreateTestJobs(5, 10*time.Millisecond)

// Create mix of successful and failing jobs
jobs := testutils.CreateMixedTestJobs(3, 2) // 3 success, 2 errors

// Execute jobs as a group
group := testutils.NewJobGroup(jobs...)
err := group.ExecuteAll(ctx)
errors := group.GetErrors()
```

### Counter and Batch Jobs

```go
// Job that increments a counter
counter := int64(0)
job := testutils.NewCounterJob("counter", &counter)

// Job that processes a batch of items
processor := func(item string) error { return nil }
job := testutils.NewBatchJob("batch", []string{"a", "b", "c"}, processor)
```

## Sample Data Fixtures

### Trakt API Data

```go
// Sample movie data
movie := testutils.SampleTraktMovie()
watched := testutils.SampleTraktWatchedMovie()
movies := testutils.SampleTraktMovies() // Multiple movies

// Sample history data
history := testutils.SampleHistoryEntries()
```

### CSV Data

```go
// Standard Letterboxd CSV structure
headers := testutils.SampleCSVHeaders()
rows := testutils.SampleCSVRows()
```

### Configuration and Auth Data

```go
// OAuth token response
tokenResp := testutils.SampleOAuthTokenResponse()

// API error response
errorResp := testutils.SampleAPIErrorResponse()

// HTTP headers and form data
headers := testutils.SampleHTTPHeaders()
formData := testutils.SampleFormData()
```

## Best Practices

1. **Use appropriate mock types**: Choose between MockLogger (with expectations), NoOpLogger (no verification), or CapturingLogger (message inspection) based on your test needs.

2. **Leverage test helpers**: Use TempDir/TempFile for file operations, SetEnv/UnsetEnv for environment variables, and the assertion helpers for complex conditions.

3. **Reuse fixtures**: Use the sample data fixtures to ensure consistent test data across different packages.

4. **Test concurrent operations**: Use the job utilities to test worker pools, concurrent processing, and error handling in multi-threaded scenarios.

5. **Clean up resources**: Always use the cleanup functions returned by helpers to ensure tests don't leave artifacts.

## Integration

To use testutils in your tests:

```go
import "github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/testutils"

func TestMyFunction(t *testing.T) {
    logger := testutils.NewNoOpLogger()
    cfg := testutils.TestConfig()
    
    // Your test code here...
}
```

This centralized approach ensures consistency across all test files and reduces maintenance overhead when interface changes occur.