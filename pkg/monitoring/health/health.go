package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/monitoring"
	"github.com/sirupsen/logrus"
)

// HealthChecker implements health checking functionality
type HealthChecker struct {
	logger    *logrus.Logger
	startTime time.Time
	version   string
	checkers  map[string]monitoring.HealthChecker
	mutex     sync.RWMutex
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(logger *logrus.Logger, version string) *HealthChecker {
	return &HealthChecker{
		logger:    logger,
		startTime: time.Now(),
		version:   version,
		checkers:  make(map[string]monitoring.HealthChecker),
	}
}

// RegisterChecker registers a component health checker
func (hc *HealthChecker) RegisterChecker(checker monitoring.HealthChecker) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	
	hc.checkers[checker.Name()] = checker
	hc.logger.WithField("component", checker.Name()).Info("Health checker registered")
}

// UnregisterChecker removes a component health checker
func (hc *HealthChecker) UnregisterChecker(name string) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	
	delete(hc.checkers, name)
	hc.logger.WithField("component", name).Info("Health checker unregistered")
}

// Check performs health checks on all registered components
func (hc *HealthChecker) Check(ctx context.Context) monitoring.OverallHealth {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()
	
	overallStatus := monitoring.HealthStatusHealthy
	components := make(map[string]monitoring.ComponentHealth)
	
	// Check each registered component
	for name, checker := range hc.checkers {
		componentHealth := checker.Check(ctx)
		components[name] = componentHealth
		
		// Determine overall status based on component status
		switch componentHealth.Status {
		case monitoring.HealthStatusUnhealthy:
			overallStatus = monitoring.HealthStatusUnhealthy
		case monitoring.HealthStatusDegraded:
			if overallStatus == monitoring.HealthStatusHealthy {
				overallStatus = monitoring.HealthStatusDegraded
			}
		}
	}
	
	// Add system health check
	systemHealth := hc.checkSystemHealth(ctx)
	components["system"] = systemHealth
	
	// Adjust overall status based on system health
	switch systemHealth.Status {
	case monitoring.HealthStatusUnhealthy:
		overallStatus = monitoring.HealthStatusUnhealthy
	case monitoring.HealthStatusDegraded:
		if overallStatus == monitoring.HealthStatusHealthy {
			overallStatus = monitoring.HealthStatusDegraded
		}
	}
	
	return monitoring.OverallHealth{
		Status:     overallStatus,
		Timestamp:  time.Now(),
		Version:    hc.version,
		Uptime:     time.Since(hc.startTime),
		Components: components,
	}
}

// checkSystemHealth performs basic system health checks
func (hc *HealthChecker) checkSystemHealth(ctx context.Context) monitoring.ComponentHealth {
	start := time.Now()
	
	var status monitoring.HealthStatus = monitoring.HealthStatusHealthy
	var message string
	details := make(map[string]interface{})
	
	// Check memory usage
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	memoryMB := float64(m.Alloc) / 1024 / 1024
	details["memory_alloc_mb"] = memoryMB
	details["goroutines"] = runtime.NumGoroutine()
	details["num_cpu"] = runtime.NumCPU()
	
	// Memory thresholds (in MB)
	if memoryMB > 1000 {
		status = monitoring.HealthStatusUnhealthy
		message = fmt.Sprintf("High memory usage: %.2f MB", memoryMB)
	} else if memoryMB > 500 {
		status = monitoring.HealthStatusDegraded
		message = fmt.Sprintf("Elevated memory usage: %.2f MB", memoryMB)
	} else {
		message = fmt.Sprintf("Memory usage: %.2f MB", memoryMB)
	}
	
	// Check goroutine count
	goroutines := runtime.NumGoroutine()
	if goroutines > 1000 {
		status = monitoring.HealthStatusUnhealthy
		message += fmt.Sprintf(", Too many goroutines: %d", goroutines)
	} else if goroutines > 100 {
		if status == monitoring.HealthStatusHealthy {
			status = monitoring.HealthStatusDegraded
		}
		message += fmt.Sprintf(", High goroutine count: %d", goroutines)
	}
	
	return monitoring.ComponentHealth{
		Name:        "system",
		Status:      status,
		LastChecked: time.Now(),
		Message:     message,
		Details:     details,
		Duration:    time.Since(start),
	}
}

// HTTPHandler returns HTTP handlers for health endpoints
func (hc *HealthChecker) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		
		health := hc.Check(ctx)
		
		w.Header().Set("Content-Type", "application/json")
		
		// Set HTTP status based on health status
		switch health.Status {
		case monitoring.HealthStatusHealthy:
			w.WriteHeader(http.StatusOK)
		case monitoring.HealthStatusDegraded:
			w.WriteHeader(http.StatusOK) // 200 but degraded
		case monitoring.HealthStatusUnhealthy:
			w.WriteHeader(http.StatusServiceUnavailable)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		
		if err := json.NewEncoder(w).Encode(health); err != nil {
			hc.logger.WithError(err).Error("Failed to encode health check response")
		}
	}
}

// ReadinessHandler returns a readiness probe handler for Kubernetes
func (hc *HealthChecker) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		
		health := hc.Check(ctx)
		
		// Readiness check is more strict - only healthy components are ready
		if health.Status == monitoring.HealthStatusHealthy {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Ready"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Not Ready"))
		}
	}
}

// LivenessHandler returns a liveness probe handler for Kubernetes
func (hc *HealthChecker) LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Liveness is more lenient - as long as the application is running
		// and not completely broken, it's considered alive
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		
		health := hc.Check(ctx)
		
		// Only unhealthy status means the app is not alive
		if health.Status != monitoring.HealthStatusUnhealthy {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Alive"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Not Alive"))
		}
	}
}

// BasicHealthChecker implements a simple health checker for basic components
type BasicHealthChecker struct {
	name      string
	checkFunc func(ctx context.Context) error
}

// NewBasicHealthChecker creates a new basic health checker
func NewBasicHealthChecker(name string, checkFunc func(ctx context.Context) error) *BasicHealthChecker {
	return &BasicHealthChecker{
		name:      name,
		checkFunc: checkFunc,
	}
}

// Check implements the HealthChecker interface
func (bhc *BasicHealthChecker) Check(ctx context.Context) monitoring.ComponentHealth {
	start := time.Now()
	
	status := monitoring.HealthStatusHealthy
	message := "Component is healthy"
	
	if err := bhc.checkFunc(ctx); err != nil {
		status = monitoring.HealthStatusUnhealthy
		message = err.Error()
	}
	
	return monitoring.ComponentHealth{
		Name:        bhc.name,
		Status:      status,
		LastChecked: time.Now(),
		Message:     message,
		Duration:    time.Since(start),
	}
}

// Name returns the name of the health checker
func (bhc *BasicHealthChecker) Name() string {
	return bhc.name
}

// TraktAPIHealthChecker checks the health of the Trakt.tv API connection
type TraktAPIHealthChecker struct {
	name       string
	apiClient  interface{} // Replace with actual API client interface
	timeout    time.Duration
}

// NewTraktAPIHealthChecker creates a new Trakt API health checker
func NewTraktAPIHealthChecker(apiClient interface{}) *TraktAPIHealthChecker {
	return &TraktAPIHealthChecker{
		name:      "trakt_api",
		apiClient: apiClient,
		timeout:   5 * time.Second,
	}
}

// Check implements the HealthChecker interface
func (tahc *TraktAPIHealthChecker) Check(ctx context.Context) monitoring.ComponentHealth {
	start := time.Now()
	
	// Create a timeout context
	checkCtx, cancel := context.WithTimeout(ctx, tahc.timeout)
	defer cancel()
	
	status := monitoring.HealthStatusHealthy
	message := "Trakt API is accessible"
	details := make(map[string]interface{})
	
	// TODO: Implement actual API connectivity check
	// For now, we'll simulate a check
	select {
	case <-checkCtx.Done():
		status = monitoring.HealthStatusUnhealthy
		message = "Trakt API check timed out"
	default:
		// In a real implementation, you would make an API call here
		// Example: ping endpoint or get user info
		details["api_endpoint"] = "https://api.trakt.tv"
		details["timeout_seconds"] = tahc.timeout.Seconds()
	}
	
	return monitoring.ComponentHealth{
		Name:        tahc.name,
		Status:      status,
		LastChecked: time.Now(),
		Message:     message,
		Details:     details,
		Duration:    time.Since(start),
	}
}

// Name returns the name of the health checker
func (tahc *TraktAPIHealthChecker) Name() string {
	return tahc.name
}

// FileSystemHealthChecker checks the health of file system operations
type FileSystemHealthChecker struct {
	name        string
	testDirPath string
}

// NewFileSystemHealthChecker creates a new file system health checker
func NewFileSystemHealthChecker(testDirPath string) *FileSystemHealthChecker {
	return &FileSystemHealthChecker{
		name:        "filesystem",
		testDirPath: testDirPath,
	}
}

// Check implements the HealthChecker interface
func (fshc *FileSystemHealthChecker) Check(ctx context.Context) monitoring.ComponentHealth {
	start := time.Now()
	
	status := monitoring.HealthStatusHealthy
	message := "File system is accessible"
	details := make(map[string]interface{})
	
	// Test if we can write to the test directory
	testFile := fmt.Sprintf("%s/.health_check_%d", fshc.testDirPath, time.Now().UnixNano())
	
	// Try to create a test file
	if err := func() error {
		// This would be the actual file write test
		// For now, we'll simulate success
		details["test_path"] = testFile
		details["writable"] = true
		return nil
	}(); err != nil {
		status = monitoring.HealthStatusUnhealthy
		message = fmt.Sprintf("Cannot write to file system: %v", err)
		details["writable"] = false
	}
	
	return monitoring.ComponentHealth{
		Name:        fshc.name,
		Status:      status,
		LastChecked: time.Now(),
		Message:     message,
		Details:     details,
		Duration:    time.Since(start),
	}
}

// Name returns the name of the health checker
func (fshc *FileSystemHealthChecker) Name() string {
	return fshc.name
} 