package telemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/monitoring"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/monitoring/tracing"
)

// InitializeTelemetryFromConfig initializes telemetry from application config
func InitializeTelemetryFromConfig(appConfig *config.Config, version string) (*TelemetryManager, error) {
	// Convert application config to telemetry config
	telemetryConfig := TelemetryConfig{
		Monitoring: monitoring.MonitoringConfig{
			Enabled:             true, // Enable by default
			MetricsEnabled:      true,
			TracingEnabled:      true,
			HealthChecksEnabled: true,
			MetricsPort:         9090,
			MetricsPath:         "/metrics",
		},
		Tracing: tracing.TracingConfig{
			Enabled:        true,
			ServiceName:    "export-trakt-letterboxd",
			ServiceVersion: version,
			Environment:    "production", // Could be read from config or env
			JaegerEndpoint: "",           // Optional
			SamplingRate:   0.1,          // 10% sampling
		},
		Logging: monitoring.LoggingConfig{
			Level:           appConfig.Logging.Level,
			Format:          "visual", // Default to visual for compatibility
			Output:          "stdout",
			RotationEnabled: false,
			MaxAgeDays:      30,
			MaxSizeMB:       100,
			MaxBackups:      3,
			CorrelationID:   true,
		},
		Alerts: monitoring.AlertsConfig{
			WebhookURL:       "",
			EmailEnabled:     false,
			SlackEnabled:     false,
			RateLimitMinutes: 5,
		},
	}

	// Override with environment variables if available
	if endpoint := getEnvOrDefault("JAEGER_ENDPOINT", ""); endpoint != "" {
		telemetryConfig.Tracing.JaegerEndpoint = endpoint
	}

	if webhookURL := getEnvOrDefault("ALERT_WEBHOOK_URL", ""); webhookURL != "" {
		telemetryConfig.Alerts.WebhookURL = webhookURL
	}

	return NewTelemetryManager(telemetryConfig, version)
}

// getEnvOrDefault gets environment variable or returns default
func getEnvOrDefault(key, defaultValue string) string {
	// In a real implementation, you'd use os.Getenv
	// For now, just return the default
	return defaultValue
}

// InstrumentedExportOperation wraps an export operation with full telemetry
func InstrumentedExportOperation(tm *TelemetryManager, ctx context.Context, operationType string, fn func(ctx context.Context) error) error {
	// Add correlation ID to context
	ctx = tm.GetLogger().WithCorrelationID(ctx)
	
	// Log operation start
	tm.GetLogger().InfoWithContext(ctx, "export.started", map[string]interface{}{
		"operation_type": operationType,
	})

	// Trace the operation with metrics and alerting
	err := tm.TraceExportOperation(ctx, operationType, fn)

	// Log operation completion
	if err != nil {
		tm.GetLogger().ErrorWithContext(ctx, "export.failed", map[string]interface{}{
			"operation_type": operationType,
			"error":          err.Error(),
		})
	} else {
		tm.GetLogger().InfoWithContext(ctx, "export.completed", map[string]interface{}{
			"operation_type": operationType,
		})
	}

	return err
}

// InstrumentedAPICall wraps an API call with full telemetry
func InstrumentedAPICall(tm *TelemetryManager, ctx context.Context, service, endpoint, method string, fn func(ctx context.Context) error) error {
	// Add correlation ID to context if not present
	if ctx.Value("correlation_id") == nil {
		ctx = tm.GetLogger().WithCorrelationID(ctx)
	}

	// Log API call start
	tm.GetLogger().DebugWithContext(ctx, "api.call.started", map[string]interface{}{
		"service":  service,
		"endpoint": endpoint,
		"method":   method,
	})

	// Trace the API call with metrics
	err := tm.TraceAPICall(ctx, service, endpoint, method, fn)

	// Log API call completion
	if err != nil {
		tm.GetLogger().WarnWithContext(ctx, "api.call.failed", map[string]interface{}{
			"service":  service,
			"endpoint": endpoint,
			"method":   method,
			"error":    err.Error(),
		})
	} else {
		tm.GetLogger().DebugWithContext(ctx, "api.call.completed", map[string]interface{}{
			"service":  service,
			"endpoint": endpoint,
			"method":   method,
		})
	}

	return err
}

// HealthCheckIntegration provides helpers for health check integration
type HealthCheckIntegration struct {
	tm *TelemetryManager
}

// NewHealthCheckIntegration creates a new health check integration helper
func NewHealthCheckIntegration(tm *TelemetryManager) *HealthCheckIntegration {
	return &HealthCheckIntegration{tm: tm}
}

// RegisterCustomHealthChecker registers a custom health checker
func (hci *HealthCheckIntegration) RegisterCustomHealthChecker(name string, checkFunc func(ctx context.Context) error) {
	if hci.tm.GetHealthChecker() == nil {
		return
	}

	checker := &CustomHealthChecker{
		name:      name,
		checkFunc: checkFunc,
	}

	hci.tm.GetHealthChecker().RegisterChecker(checker)
}

// CustomHealthChecker implements a customizable health checker
type CustomHealthChecker struct {
	name      string
	checkFunc func(ctx context.Context) error
}

// Check implements the HealthChecker interface
func (chc *CustomHealthChecker) Check(ctx context.Context) monitoring.ComponentHealth {
	start := time.Now()
	
	status := monitoring.HealthStatusHealthy
	message := fmt.Sprintf("%s is healthy", chc.name)
	
	if err := chc.checkFunc(ctx); err != nil {
		status = monitoring.HealthStatusUnhealthy
		message = err.Error()
	}
	
	return monitoring.ComponentHealth{
		Name:        chc.name,
		Status:      status,
		LastChecked: time.Now(),
		Message:     message,
		Duration:    time.Since(start),
	}
}

// Name returns the name of the health checker
func (chc *CustomHealthChecker) Name() string {
	return chc.name
}

// ExampleUsage demonstrates how to use the telemetry system
func ExampleUsage() {
	// Initialize telemetry manager
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
			Enabled:        true,
			ServiceName:    "export-trakt-letterboxd",
			ServiceVersion: "1.0.0",
			Environment:    "development",
			SamplingRate:   1.0, // 100% for development
		},
		Logging: monitoring.LoggingConfig{
			Level:         "debug",
			Format:        "json",
			Output:        "stdout",
			CorrelationID: true,
		},
		Alerts: monitoring.AlertsConfig{
			RateLimitMinutes: 5,
		},
	}

	tm, err := NewTelemetryManager(config, "1.0.0")
	if err != nil {
		panic(err)
	}

	// Start telemetry
	ctx := context.Background()
	if err := tm.Start(ctx); err != nil {
		panic(err)
	}

	// Example: Instrumented export operation
	err = InstrumentedExportOperation(tm, ctx, "movies", func(ctx context.Context) error {
		// Simulate export work
		time.Sleep(100 * time.Millisecond)
		
		// Example: API call within the export
		return InstrumentedAPICall(tm, ctx, "trakt", "/movies/watched", "GET", func(ctx context.Context) error {
			// Simulate API call
			time.Sleep(50 * time.Millisecond)
			return nil
		})
	})

	if err != nil {
		fmt.Printf("Export failed: %v\n", err)
	}

	// Example: Custom health checker
	healthIntegration := NewHealthCheckIntegration(tm)
	healthIntegration.RegisterCustomHealthChecker("database", func(ctx context.Context) error {
		// Check database connectivity
		return nil // Healthy
	})

	// Stop telemetry gracefully
	defer tm.Stop(ctx)
}

// TelemetryMiddleware provides middleware for HTTP handlers or other operations
type TelemetryMiddleware struct {
	tm *TelemetryManager
}

// NewTelemetryMiddleware creates a new telemetry middleware
func NewTelemetryMiddleware(tm *TelemetryManager) *TelemetryMiddleware {
	return &TelemetryMiddleware{tm: tm}
}

// WrapOperation wraps any operation with telemetry
func (tmw *TelemetryMiddleware) WrapOperation(operationName string, fn func(ctx context.Context) error) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		// Add correlation ID
		ctx = tmw.tm.GetLogger().WithCorrelationID(ctx)
		
		// Start span
		spanCtx, endSpan := tmw.tm.GetTracer().StartSpan(ctx, operationName)
		defer endSpan()
		
		// Log operation start
		tmw.tm.GetLogger().WithContext(spanCtx).WithField("operation", operationName).Info("Operation started")
		
		start := time.Now()
		err := fn(spanCtx)
		duration := time.Since(start)
		
		// Record metrics (placeholder for custom operation metrics)
		if tmw.tm.GetMetrics() != nil {
			// You could add a custom metric for operations here
			// Example: tmw.tm.GetMetrics().RecordOperation(operationName, status, duration)
			_ = duration // Placeholder to avoid unused variable warning
		}
		
		// Log completion
		if err != nil {
			tmw.tm.GetLogger().WithContext(spanCtx).WithField("operation", operationName).WithField("duration", duration).WithError(err).Error("Operation failed")
		} else {
			tmw.tm.GetLogger().WithContext(spanCtx).WithField("operation", operationName).WithField("duration", duration).Info("Operation completed")
		}
		
		return err
	}
} 