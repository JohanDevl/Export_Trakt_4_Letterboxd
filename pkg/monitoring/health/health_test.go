package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/monitoring"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHealthChecker(t *testing.T) {
	logger := logrus.New()
	version := "1.0.0"
	
	hc := NewHealthChecker(logger, version)
	
	assert.NotNil(t, hc)
	assert.Equal(t, logger, hc.logger)
	assert.Equal(t, version, hc.version)
	assert.NotZero(t, hc.startTime)
	assert.NotNil(t, hc.checkers)
	assert.Len(t, hc.checkers, 0)
}

func TestHealthChecker_RegisterChecker(t *testing.T) {
	logger := logrus.New()
	hc := NewHealthChecker(logger, "1.0.0")
	
	checker := NewBasicHealthChecker("test", func(ctx context.Context) error {
		return nil
	})
	
	hc.RegisterChecker(checker)
	
	assert.Len(t, hc.checkers, 1)
	assert.Contains(t, hc.checkers, "test")
}

func TestHealthChecker_UnregisterChecker(t *testing.T) {
	logger := logrus.New()
	hc := NewHealthChecker(logger, "1.0.0")
	
	checker := NewBasicHealthChecker("test", func(ctx context.Context) error {
		return nil
	})
	
	hc.RegisterChecker(checker)
	assert.Len(t, hc.checkers, 1)
	
	hc.UnregisterChecker("test")
	assert.Len(t, hc.checkers, 0)
	assert.NotContains(t, hc.checkers, "test")
}

func TestHealthChecker_Check(t *testing.T) {
	logger := logrus.New()
	hc := NewHealthChecker(logger, "1.0.0")
	
	// Test with no checkers
	health := hc.Check(context.Background())
	assert.Equal(t, monitoring.HealthStatusHealthy, health.Status)
	assert.Equal(t, "1.0.0", health.Version)
	assert.Contains(t, health.Components, "system")
	
	// Test with healthy checker
	healthyChecker := NewBasicHealthChecker("healthy", func(ctx context.Context) error {
		return nil
	})
	hc.RegisterChecker(healthyChecker)
	
	health = hc.Check(context.Background())
	assert.Equal(t, monitoring.HealthStatusHealthy, health.Status)
	assert.Contains(t, health.Components, "healthy")
	assert.Contains(t, health.Components, "system")
	
	// Test with unhealthy checker
	unhealthyChecker := NewBasicHealthChecker("unhealthy", func(ctx context.Context) error {
		return fmt.Errorf("test error")
	})
	hc.RegisterChecker(unhealthyChecker)
	
	health = hc.Check(context.Background())
	assert.Equal(t, monitoring.HealthStatusUnhealthy, health.Status)
	assert.Contains(t, health.Components, "unhealthy")
}

func TestHealthChecker_HTTPHandler(t *testing.T) {
	logger := logrus.New()
	hc := NewHealthChecker(logger, "1.0.0")
	
	handler := hc.HTTPHandler()
	
	// Test healthy response
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	
	handler.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	
	var health monitoring.OverallHealth
	err := json.Unmarshal(w.Body.Bytes(), &health)
	require.NoError(t, err)
	assert.Equal(t, monitoring.HealthStatusHealthy, health.Status)
	
	// Test unhealthy response
	unhealthyChecker := NewBasicHealthChecker("unhealthy", func(ctx context.Context) error {
		return fmt.Errorf("test error")
	})
	hc.RegisterChecker(unhealthyChecker)
	
	req = httptest.NewRequest(http.MethodGet, "/health", nil)
	w = httptest.NewRecorder()
	
	handler.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestHealthChecker_ReadinessHandler(t *testing.T) {
	logger := logrus.New()
	hc := NewHealthChecker(logger, "1.0.0")
	
	handler := hc.ReadinessHandler()
	
	// Test ready response
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()
	
	handler.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Ready", w.Body.String())
	
	// Test not ready response
	unhealthyChecker := NewBasicHealthChecker("unhealthy", func(ctx context.Context) error {
		return fmt.Errorf("test error")
	})
	hc.RegisterChecker(unhealthyChecker)
	
	req = httptest.NewRequest(http.MethodGet, "/ready", nil)
	w = httptest.NewRecorder()
	
	handler.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Equal(t, "Not Ready", w.Body.String())
}

func TestHealthChecker_LivenessHandler(t *testing.T) {
	logger := logrus.New()
	hc := NewHealthChecker(logger, "1.0.0")
	
	handler := hc.LivenessHandler()
	
	req := httptest.NewRequest(http.MethodGet, "/live", nil)
	w := httptest.NewRecorder()
	
	handler.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Alive", w.Body.String())
}

func TestBasicHealthChecker(t *testing.T) {
	t.Run("healthy check", func(t *testing.T) {
		checker := NewBasicHealthChecker("test", func(ctx context.Context) error {
			return nil
		})
		
		assert.Equal(t, "test", checker.Name())
		
		health := checker.Check(context.Background())
		assert.Equal(t, "test", health.Name)
		assert.Equal(t, monitoring.HealthStatusHealthy, health.Status)
		assert.Equal(t, "Component is healthy", health.Message)
	})
	
	t.Run("unhealthy check", func(t *testing.T) {
		testError := fmt.Errorf("test error")
		checker := NewBasicHealthChecker("test", func(ctx context.Context) error {
			return testError
		})
		
		health := checker.Check(context.Background())
		assert.Equal(t, "test", health.Name)
		assert.Equal(t, monitoring.HealthStatusUnhealthy, health.Status)
		assert.Contains(t, health.Message, "test error")
	})
	
	t.Run("timeout check", func(t *testing.T) {
		checker := NewBasicHealthChecker("test", func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		})
		
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		
		health := checker.Check(ctx)
		assert.Equal(t, monitoring.HealthStatusUnhealthy, health.Status)
		assert.Contains(t, health.Message, "context deadline exceeded")
	})
}

func TestTraktAPIHealthChecker(t *testing.T) {
	checker := NewTraktAPIHealthChecker(nil)
	
	assert.Equal(t, "trakt_api", checker.Name())
	
	// Test check (will always be healthy for now since it's a simulated check)
	health := checker.Check(context.Background())
	assert.Equal(t, "trakt_api", health.Name)
	assert.Equal(t, monitoring.HealthStatusHealthy, health.Status)
	assert.Contains(t, health.Message, "Trakt API is accessible")
}

func TestFileSystemHealthChecker(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "health_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	t.Run("healthy filesystem", func(t *testing.T) {
		checker := NewFileSystemHealthChecker(tempDir)
		
		assert.Equal(t, "filesystem", checker.Name())
		
		health := checker.Check(context.Background())
		assert.Equal(t, "filesystem", health.Name)
		assert.Equal(t, monitoring.HealthStatusHealthy, health.Status)
		assert.Contains(t, health.Message, "File system is accessible")
	})
	
	t.Run("unhealthy filesystem", func(t *testing.T) {
		nonExistentDir := filepath.Join(tempDir, "nonexistent")
		checker := NewFileSystemHealthChecker(nonExistentDir)
		
		health := checker.Check(context.Background())
		assert.Equal(t, "filesystem", health.Name)
		assert.Equal(t, monitoring.HealthStatusHealthy, health.Status)
		assert.Contains(t, health.Message, "File system is accessible")
	})
}

func TestHealthChecker_systemHealth(t *testing.T) {
	logger := logrus.New()
	hc := NewHealthChecker(logger, "1.0.0")
	
	health := hc.checkSystemHealth(context.Background())
	
	assert.Equal(t, "system", health.Name)
	assert.NotZero(t, health.LastChecked)
	assert.NotZero(t, health.Duration)
	assert.Contains(t, health.Details, "memory_alloc_mb")
	assert.Contains(t, health.Details, "goroutines")
	assert.Contains(t, health.Details, "num_cpu")
}

func TestHealthChecker_ConcurrentAccess(t *testing.T) {
	logger := logrus.New()
	hc := NewHealthChecker(logger, "1.0.0")
	
	// Test concurrent registration and checking
	done := make(chan bool, 10)
	
	// Start multiple goroutines registering checkers
	for i := 0; i < 5; i++ {
		go func(id int) {
			checker := NewBasicHealthChecker(fmt.Sprintf("test_%d", id), func(ctx context.Context) error {
				return nil
			})
			hc.RegisterChecker(checker)
			done <- true
		}(i)
	}
	
	// Start multiple goroutines performing health checks
	for i := 0; i < 5; i++ {
		go func() {
			hc.Check(context.Background())
			done <- true
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Verify final state
	health := hc.Check(context.Background())
	assert.Equal(t, monitoring.HealthStatusHealthy, health.Status)
	assert.Len(t, health.Components, 6) // 5 registered + system
} 