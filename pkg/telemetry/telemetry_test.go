package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/monitoring"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/monitoring/tracing"
)

func TestNewTelemetryManager(t *testing.T) {
	tests := []struct {
		name        string
		config      TelemetryConfig
		version     string
		expectError bool
	}{
		{
			name: "basic config",
			config: TelemetryConfig{
				Monitoring: monitoring.MonitoringConfig{
					Enabled:             true,
					MetricsEnabled:      true,
					TracingEnabled:      false,
					HealthChecksEnabled: true,
					MetricsPort:         9090,
					MetricsPath:         "/metrics",
				},
				Logging: monitoring.LoggingConfig{
					Level:  "info",
					Format: "json",
					Output: "stdout",
				},
				Alerts: monitoring.AlertsConfig{},
			},
			version:     "1.0.0",
			expectError: false,
		},
		{
			name: "with tracing enabled",
			config: TelemetryConfig{
				Monitoring: monitoring.MonitoringConfig{
					Enabled:             true,
					MetricsEnabled:      true,
					TracingEnabled:      true,
					HealthChecksEnabled: true,
					MetricsPort:         9091,
					MetricsPath:         "/metrics",
				},
				Tracing: tracing.TracingConfig{
					ServiceName:    "export-trakt",
					ServiceVersion: "1.0.0",
					Enabled:        true,
					SamplingRate:   1.0,
				},
				Logging: monitoring.LoggingConfig{
					Level:  "debug",
					Format: "json",
					Output: "stdout",
				},
				Alerts: monitoring.AlertsConfig{},
			},
			version:     "1.0.0",
			expectError: false,
		},
		{
			name: "minimal config",
			config: TelemetryConfig{
				Monitoring: monitoring.MonitoringConfig{
					Enabled:             false,
					MetricsEnabled:      false,
					TracingEnabled:      false,
					HealthChecksEnabled: false,
				},
				Logging: monitoring.LoggingConfig{
					Level:  "info",
					Format: "text",
					Output: "stdout",
				},
				Alerts: monitoring.AlertsConfig{},
			},
			version:     "1.0.0",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm, err := NewTelemetryManager(tt.config, tt.version)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if tm == nil {
				t.Fatal("TelemetryManager should not be nil")
			}
			
			// Verify components are initialized based on config
			if tt.config.Monitoring.MetricsEnabled && tm.metrics == nil {
				t.Error("Metrics should be initialized when enabled")
			}
			
			if tt.config.Monitoring.TracingEnabled && tm.tracer == nil {
				t.Error("Tracer should be initialized when enabled")
			}
			
			if tt.config.Monitoring.HealthChecksEnabled && tm.healthChecker == nil {
				t.Error("Health checker should be initialized when enabled")
			}
			
			if tm.alertManager == nil {
				t.Error("Alert manager should always be initialized")
			}
			
			if tm.logger == nil {
				t.Error("Logger should always be initialized")
			}
		})
	}
}

func TestTelemetryManagerStartStop(t *testing.T) {
	config := TelemetryConfig{
		Monitoring: monitoring.MonitoringConfig{
			Enabled:             true,
			MetricsEnabled:      true,
			TracingEnabled:      false,
			HealthChecksEnabled: true,
			MetricsPort:         0, // Use random port for testing
			MetricsPath:         "/metrics",
		},
		Logging: monitoring.LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		Alerts: monitoring.AlertsConfig{},
	}

	tm, err := NewTelemetryManager(config, "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	// Test initial state
	if tm.IsRunning() {
		t.Error("TelemetryManager should not be running initially")
	}

	// Test Start
	ctx := context.Background()
	err = tm.Start(ctx)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if !tm.IsRunning() {
		t.Error("TelemetryManager should be running after Start")
	}

	// Test double start (should fail)
	err = tm.Start(ctx)
	if err == nil {
		t.Error("Expected error when starting already running manager")
	}

	// Test Stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err = tm.Stop(ctx)
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	if tm.IsRunning() {
		t.Error("TelemetryManager should not be running after Stop")
	}
}

func TestGetters(t *testing.T) {
	config := TelemetryConfig{
		Monitoring: monitoring.MonitoringConfig{
			Enabled:             true,
			MetricsEnabled:      true,
			TracingEnabled:      true,
			HealthChecksEnabled: true,
			MetricsPort:         9090,
			MetricsPath:         "/metrics",
		},
		Tracing: tracing.TracingConfig{
			ServiceName:    "export-trakt",
			ServiceVersion: "1.0.0",
			Enabled:        true,
			SamplingRate:   1.0,
		},
		Logging: monitoring.LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		Alerts: monitoring.AlertsConfig{},
	}

	tm, err := NewTelemetryManager(config, "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	// Test GetLogger
	logger := tm.GetLogger()
	if logger == nil {
		t.Error("GetLogger should not return nil")
	}

	// Test GetMetrics
	metrics := tm.GetMetrics()
	if metrics == nil {
		t.Error("GetMetrics should not return nil when metrics are enabled")
	}

	// Test GetTracer
	tracer := tm.GetTracer()
	if tracer == nil {
		t.Error("GetTracer should not return nil when tracing is enabled")
	}

	// Test GetHealthChecker
	healthChecker := tm.GetHealthChecker()
	if healthChecker == nil {
		t.Error("GetHealthChecker should not return nil when health checks are enabled")
	}

	// Test GetAlertManager
	alertManager := tm.GetAlertManager()
	if alertManager == nil {
		t.Error("GetAlertManager should not return nil")
	}
}

func TestGetTelemetryStatus(t *testing.T) {
	config := TelemetryConfig{
		Monitoring: monitoring.MonitoringConfig{
			Enabled:             true,
			MetricsEnabled:      true,
			TracingEnabled:      true,
			HealthChecksEnabled: true,
			MetricsPort:         9090,
			MetricsPath:         "/metrics",
		},
		Tracing: tracing.TracingConfig{
			ServiceName: "export-trakt",
			Enabled:     true,
		},
		Logging: monitoring.LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		Alerts: monitoring.AlertsConfig{},
	}

	tm, err := NewTelemetryManager(config, "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	status := tm.GetTelemetryStatus()
	
	if status == nil {
		t.Fatal("Status should not be nil")
	}

	// Check expected status fields
	expectedFields := []string{
		"running", "monitoring_enabled", "metrics_enabled", 
		"tracing_enabled", "health_checks_enabled", "alerts_enabled",
	}

	for _, field := range expectedFields {
		if _, exists := status[field]; !exists {
			t.Errorf("Status should contain field: %s", field)
		}
	}

	// Verify running status
	if status["running"] != tm.IsRunning() {
		t.Error("Status running field should match IsRunning()")
	}
}

func TestTraceExportOperation(t *testing.T) {
	config := TelemetryConfig{
		Monitoring: monitoring.MonitoringConfig{
			Enabled:        true,
			TracingEnabled: true,
		},
		Tracing: tracing.TracingConfig{
			ServiceName: "export-trakt",
			Enabled:     true,
		},
		Logging: monitoring.LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		Alerts: monitoring.AlertsConfig{},
	}

	tm, err := NewTelemetryManager(config, "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	exportType := "movies"

	// Test successful operation
	operationCalled := false
	err = tm.TraceExportOperation(ctx, exportType, func(ctx context.Context) error {
		operationCalled = true
		return nil
	})

	if err != nil {
		t.Errorf("TraceExportOperation should not return error: %v", err)
	}

	if !operationCalled {
		t.Error("Operation function should have been called")
	}

	// Test operation with error
	testError := fmt.Errorf("test error")
	err = tm.TraceExportOperation(ctx, exportType, func(ctx context.Context) error {
		return testError
	})

	if err != testError {
		t.Errorf("Expected error %v, got %v", testError, err)
	}
}

func TestTraceAPICall(t *testing.T) {
	config := TelemetryConfig{
		Monitoring: monitoring.MonitoringConfig{
			Enabled:        true,
			TracingEnabled: true,
		},
		Tracing: tracing.TracingConfig{
			ServiceName: "export-trakt",
			Enabled:     true,
		},
		Logging: monitoring.LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		Alerts: monitoring.AlertsConfig{},
	}

	tm, err := NewTelemetryManager(config, "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	service := "trakt"
	endpoint := "/movies"
	method := "GET"

	// Test successful API call
	apiCallMade := false
	err = tm.TraceAPICall(ctx, service, endpoint, method, func(ctx context.Context) error {
		apiCallMade = true
		return nil
	})

	if err != nil {
		t.Errorf("TraceAPICall should not return error: %v", err)
	}

	if !apiCallMade {
		t.Error("API call function should have been called")
	}

	// Test API call with error
	testError := fmt.Errorf("api error")
	err = tm.TraceAPICall(ctx, service, endpoint, method, func(ctx context.Context) error {
		return testError
	})

	if err != testError {
		t.Errorf("Expected error %v, got %v", testError, err)
	}
}

func TestHTTPEndpoints(t *testing.T) {
	config := TelemetryConfig{
		Monitoring: monitoring.MonitoringConfig{
			Enabled:             true,
			MetricsEnabled:      true,
			HealthChecksEnabled: true,
			MetricsPort:         0, // Use random port
			MetricsPath:         "/metrics",
		},
		Logging: monitoring.LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		Alerts: monitoring.AlertsConfig{},
	}

	tm, err := NewTelemetryManager(config, "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	if tm.httpServer == nil {
		t.Fatal("HTTP server should be initialized")
	}

	// Test health endpoint
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	tm.httpServer.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for /health, got %d", w.Code)
	}

	// Test metrics endpoint
	req = httptest.NewRequest("GET", "/metrics", nil)
	w = httptest.NewRecorder()
	tm.httpServer.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for /metrics, got %d", w.Code)
	}

	// Test alerts endpoint
	req = httptest.NewRequest("GET", "/alerts", nil)
	w = httptest.NewRecorder()
	tm.httpServer.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for /alerts, got %d", w.Code)
	}
}

func TestTelemetryManagerWithDisabledComponents(t *testing.T) {
	config := TelemetryConfig{
		Monitoring: monitoring.MonitoringConfig{
			Enabled:             false,
			MetricsEnabled:      false,
			TracingEnabled:      false,
			HealthChecksEnabled: false,
		},
		Logging: monitoring.LoggingConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		},
		Alerts: monitoring.AlertsConfig{},
	}

	tm, err := NewTelemetryManager(config, "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	// With all components disabled, only logger and alert manager should be present
	if tm.metrics != nil {
		t.Error("Metrics should be nil when disabled")
	}

	if tm.tracer != nil {
		t.Error("Tracer should be nil when disabled")
	}

	if tm.healthChecker != nil {
		t.Error("Health checker should be nil when disabled")
	}

	if tm.logger == nil {
		t.Error("Logger should always be initialized")
	}

	if tm.alertManager == nil {
		t.Error("Alert manager should always be initialized")
	}

	// HTTP server should be nil when monitoring is disabled
	if tm.httpServer != nil {
		t.Error("HTTP server should be nil when monitoring is disabled")
	}
}

func TestTelemetryConfigStructure(t *testing.T) {
	// Test that TelemetryConfig can be properly structured
	config := TelemetryConfig{
		Monitoring: monitoring.MonitoringConfig{
			Enabled:             true,
			MetricsEnabled:      true,
			TracingEnabled:      true,
			HealthChecksEnabled: true,
			MetricsPort:         9090,
			MetricsPath:         "/metrics",
		},
		Tracing: tracing.TracingConfig{
			ServiceName:    "test-service",
			ServiceVersion: "1.0.0",
			Enabled:        true,
			SamplingRate:   1.0,
		},
		Logging: monitoring.LoggingConfig{
			Level:  "debug",
			Format: "json",
			Output: "stdout",
		},
		Alerts: monitoring.AlertsConfig{},
	}

	// Verify all fields are accessible
	if config.Monitoring.Enabled != true {
		t.Error("Monitoring.Enabled should be accessible")
	}

	if config.Tracing.ServiceName != "test-service" {
		t.Error("Tracing.ServiceName should be accessible")
	}

	if config.Logging.Level != "debug" {
		t.Error("Logging.Level should be accessible")
	}
}

func TestConcurrentAccess(t *testing.T) {
	config := TelemetryConfig{
		Monitoring: monitoring.MonitoringConfig{
			Enabled:             true,
			MetricsEnabled:      true,
			HealthChecksEnabled: true,
			MetricsPort:         0,
			MetricsPath:         "/metrics",
		},
		Logging: monitoring.LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		Alerts: monitoring.AlertsConfig{},
	}

	tm, err := NewTelemetryManager(config, "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	// Test concurrent access to getters
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			
			// Call various getters concurrently
			_ = tm.GetLogger()
			_ = tm.GetMetrics()
			_ = tm.GetHealthChecker()
			_ = tm.GetAlertManager()
			_ = tm.IsRunning()
			_ = tm.GetTelemetryStatus()
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestAlertHistoryHandler(t *testing.T) {
	config := TelemetryConfig{
		Monitoring: monitoring.MonitoringConfig{
			Enabled:        true,
			MetricsEnabled: true,
			MetricsPort:    0,
			MetricsPath:    "/metrics",
		},
		Logging: monitoring.LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		Alerts: monitoring.AlertsConfig{},
	}

	tm, err := NewTelemetryManager(config, "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	// Test alert history endpoint with different methods
	methods := []string{"GET", "POST", "PUT", "DELETE"}
	
	for _, method := range methods {
		req := httptest.NewRequest(method, "/alerts", nil)
		w := httptest.NewRecorder()
		
		tm.alertHistoryHandler(w, req)
		
		if method == "GET" {
			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200 for GET /alerts, got %d", w.Code)
			}
		} else {
			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status 405 for %s /alerts, got %d", method, w.Code)
			}
		}
	}
} 