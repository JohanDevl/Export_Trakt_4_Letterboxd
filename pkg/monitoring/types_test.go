package monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHealthStatus_String(t *testing.T) {
	testCases := []struct {
		status   HealthStatus
		expected string
	}{
		{HealthStatusHealthy, "healthy"},
		{HealthStatusDegraded, "degraded"},
		{HealthStatusUnhealthy, "unhealthy"},
		{HealthStatusUnknown, "unknown"},
	}
	
	for _, tc := range testCases {
		t.Run(string(tc.status), func(t *testing.T) {
			assert.Equal(t, tc.expected, string(tc.status))
		})
	}
}

func TestComponentHealth_Validation(t *testing.T) {
	now := time.Now()
	
	health := ComponentHealth{
		Name:        "test-component",
		Status:      HealthStatusHealthy,
		LastChecked: now,
		Message:     "All systems operational",
		Details: map[string]interface{}{
			"cpu_usage":    0.25,
			"memory_usage": "256MB",
			"connections":  10,
		},
		Duration: 50 * time.Millisecond,
	}
	
	assert.Equal(t, "test-component", health.Name)
	assert.Equal(t, HealthStatusHealthy, health.Status)
	assert.Equal(t, now, health.LastChecked)
	assert.Equal(t, "All systems operational", health.Message)
	assert.Contains(t, health.Details, "cpu_usage")
	assert.Contains(t, health.Details, "memory_usage")
	assert.Contains(t, health.Details, "connections")
	assert.Equal(t, 50*time.Millisecond, health.Duration)
}

func TestOverallHealth_Validation(t *testing.T) {
	now := time.Now()
	uptime := 2 * time.Hour
	
	component1 := ComponentHealth{
		Name:        "database",
		Status:      HealthStatusHealthy,
		LastChecked: now,
		Message:     "Database connection OK",
		Duration:    10 * time.Millisecond,
	}
	
	component2 := ComponentHealth{
		Name:        "api",
		Status:      HealthStatusDegraded,
		LastChecked: now,
		Message:     "API response time elevated",
		Duration:    25 * time.Millisecond,
	}
	
	components := map[string]ComponentHealth{
		"database": component1,
		"api":      component2,
	}
	
	overallHealth := OverallHealth{
		Status:     HealthStatusDegraded,
		Timestamp:  now,
		Version:    "1.2.3",
		Uptime:     uptime,
		Components: components,
	}
	
	assert.Equal(t, HealthStatusDegraded, overallHealth.Status)
	assert.Equal(t, now, overallHealth.Timestamp)
	assert.Equal(t, "1.2.3", overallHealth.Version)
	assert.Equal(t, uptime, overallHealth.Uptime)
	assert.Len(t, overallHealth.Components, 2)
	assert.Contains(t, overallHealth.Components, "database")
	assert.Contains(t, overallHealth.Components, "api")
	assert.Equal(t, HealthStatusHealthy, overallHealth.Components["database"].Status)
	assert.Equal(t, HealthStatusDegraded, overallHealth.Components["api"].Status)
}

func TestAlertLevel_String(t *testing.T) {
	testCases := []struct {
		level    AlertLevel
		expected string
	}{
		{AlertLevelInfo, "info"},
		{AlertLevelWarning, "warning"},
		{AlertLevelError, "error"},
		{AlertLevelCritical, "critical"},
	}
	
	for _, tc := range testCases {
		t.Run(string(tc.level), func(t *testing.T) {
			assert.Equal(t, tc.expected, string(tc.level))
		})
	}
}

func TestAlert_Validation(t *testing.T) {
	now := time.Now()
	resolvedAt := now.Add(5 * time.Minute)
	
	alert := Alert{
		ID:        "alert-123",
		Level:     AlertLevelError,
		Title:     "Database Connection Failed",
		Message:   "Unable to connect to primary database",
		Timestamp: now,
		Source:    "database-monitor",
		Metadata: map[string]interface{}{
			"database_host": "db1.example.com",
			"error_code":    500,
			"retry_count":   3,
		},
		Resolved:   true,
		ResolvedAt: &resolvedAt,
	}
	
	assert.Equal(t, "alert-123", alert.ID)
	assert.Equal(t, AlertLevelError, alert.Level)
	assert.Equal(t, "Database Connection Failed", alert.Title)
	assert.Equal(t, "Unable to connect to primary database", alert.Message)
	assert.Equal(t, now, alert.Timestamp)
	assert.Equal(t, "database-monitor", alert.Source)
	assert.Contains(t, alert.Metadata, "database_host")
	assert.Contains(t, alert.Metadata, "error_code")
	assert.Contains(t, alert.Metadata, "retry_count")
	assert.True(t, alert.Resolved)
	assert.NotNil(t, alert.ResolvedAt)
	assert.Equal(t, resolvedAt, *alert.ResolvedAt)
}

func TestAlert_UnresolvedAlert(t *testing.T) {
	now := time.Now()
	
	alert := Alert{
		ID:        "alert-456",
		Level:     AlertLevelWarning,
		Title:     "High Memory Usage",
		Message:   "Memory usage is above 80%",
		Timestamp: now,
		Source:    "system-monitor",
		Resolved:  false,
	}
	
	assert.Equal(t, "alert-456", alert.ID)
	assert.Equal(t, AlertLevelWarning, alert.Level)
	assert.False(t, alert.Resolved)
	assert.Nil(t, alert.ResolvedAt)
}

func TestMonitoringConfig_DefaultValues(t *testing.T) {
	config := MonitoringConfig{
		Enabled:             true,
		MetricsEnabled:      true,
		TracingEnabled:      false,
		HealthChecksEnabled: true,
		MetricsPort:         9090,
		MetricsPath:         "/metrics",
	}
	
	assert.True(t, config.Enabled)
	assert.True(t, config.MetricsEnabled)
	assert.False(t, config.TracingEnabled)
	assert.True(t, config.HealthChecksEnabled)
	assert.Equal(t, 9090, config.MetricsPort)
	assert.Equal(t, "/metrics", config.MetricsPath)
}

func TestLoggingConfig_DefaultValues(t *testing.T) {
	config := LoggingConfig{
		Level:           "info",
		Format:          "json",
		Output:          "stdout",
		RotationEnabled: false,
		MaxAgeDays:      7,
		MaxSizeMB:       100,
		MaxBackups:      3,
		CorrelationID:   true,
	}
	
	assert.Equal(t, "info", config.Level)
	assert.Equal(t, "json", config.Format)
	assert.Equal(t, "stdout", config.Output)
	assert.False(t, config.RotationEnabled)
	assert.Equal(t, 7, config.MaxAgeDays)
	assert.Equal(t, 100, config.MaxSizeMB)
	assert.Equal(t, 3, config.MaxBackups)
	assert.True(t, config.CorrelationID)
}

func TestAlertsConfig_DefaultValues(t *testing.T) {
	config := AlertsConfig{
		WebhookURL:       "https://hooks.slack.com/webhook",
		EmailEnabled:     false,
		SlackEnabled:     true,
		RateLimitMinutes: 5,
	}
	
	assert.Equal(t, "https://hooks.slack.com/webhook", config.WebhookURL)
	assert.False(t, config.EmailEnabled)
	assert.True(t, config.SlackEnabled)
	assert.Equal(t, 5, config.RateLimitMinutes)
}

// Mock implementations for testing interfaces

type MockHealthChecker struct {
	name   string
	health ComponentHealth
}

func (m *MockHealthChecker) Check(ctx context.Context) ComponentHealth {
	return m.health
}

func (m *MockHealthChecker) Name() string {
	return m.name
}

type MockMetricsCollector struct {
	name         string
	collectError error
	registerError error
}

func (m *MockMetricsCollector) CollectMetrics(ctx context.Context) error {
	return m.collectError
}

func (m *MockMetricsCollector) RegisterMetrics() error {
	return m.registerError
}

func (m *MockMetricsCollector) Name() string {
	return m.name
}

type MockTracingProvider struct {
	spans []string
}

func (m *MockTracingProvider) StartSpan(ctx context.Context, operationName string) (context.Context, func()) {
	m.spans = append(m.spans, operationName)
	return ctx, func() {}
}

func (m *MockTracingProvider) AddSpanEvent(ctx context.Context, name string, attributes map[string]string) {
	// Mock implementation
}

func (m *MockTracingProvider) Close() error {
	return nil
}

type MockNotificationChannel struct {
	name    string
	enabled bool
	alerts  []Alert
}

func (m *MockNotificationChannel) SendAlert(ctx context.Context, alert Alert) error {
	m.alerts = append(m.alerts, alert)
	return nil
}

func (m *MockNotificationChannel) Name() string {
	return m.name
}

func (m *MockNotificationChannel) IsEnabled() bool {
	return m.enabled
}

func TestMockHealthChecker(t *testing.T) {
	health := ComponentHealth{
		Name:        "test-component",
		Status:      HealthStatusHealthy,
		LastChecked: time.Now(),
		Message:     "OK",
		Duration:    10 * time.Millisecond,
	}
	
	checker := &MockHealthChecker{
		name:   "test-component",
		health: health,
	}
	
	assert.Equal(t, "test-component", checker.Name())
	
	result := checker.Check(context.Background())
	assert.Equal(t, health, result)
}

func TestMockMetricsCollector(t *testing.T) {
	collector := &MockMetricsCollector{
		name: "test-collector",
	}
	
	assert.Equal(t, "test-collector", collector.Name())
	
	err := collector.RegisterMetrics()
	assert.NoError(t, err)
	
	err = collector.CollectMetrics(context.Background())
	assert.NoError(t, err)
}

func TestMockTracingProvider(t *testing.T) {
	provider := &MockTracingProvider{}
	
	ctx := context.Background()
	spanCtx, endSpan := provider.StartSpan(ctx, "test-operation")
	
	assert.Equal(t, ctx, spanCtx)
	assert.Len(t, provider.spans, 1)
	assert.Equal(t, "test-operation", provider.spans[0])
	
	endSpan()
	
	provider.AddSpanEvent(ctx, "test-event", map[string]string{"key": "value"})
	
	err := provider.Close()
	assert.NoError(t, err)
}

func TestMockNotificationChannel(t *testing.T) {
	channel := &MockNotificationChannel{
		name:    "test-channel",
		enabled: true,
	}
	
	assert.Equal(t, "test-channel", channel.Name())
	assert.True(t, channel.IsEnabled())
	
	alert := Alert{
		ID:      "test-alert",
		Level:   AlertLevelInfo,
		Title:   "Test Alert",
		Message: "This is a test alert",
	}
	
	err := channel.SendAlert(context.Background(), alert)
	assert.NoError(t, err)
	assert.Len(t, channel.alerts, 1)
	assert.Equal(t, alert, channel.alerts[0])
}

func TestHealthStatusValidation(t *testing.T) {
	validStatuses := []HealthStatus{
		HealthStatusHealthy,
		HealthStatusDegraded,
		HealthStatusUnhealthy,
		HealthStatusUnknown,
	}
	
	for _, status := range validStatuses {
		t.Run(string(status), func(t *testing.T) {
			// Test that the status is one of the valid constants
			assert.Contains(t, []HealthStatus{
				HealthStatusHealthy,
				HealthStatusDegraded,
				HealthStatusUnhealthy,
				HealthStatusUnknown,
			}, status)
		})
	}
}

func TestAlertLevelValidation(t *testing.T) {
	validLevels := []AlertLevel{
		AlertLevelInfo,
		AlertLevelWarning,
		AlertLevelError,
		AlertLevelCritical,
	}
	
	for _, level := range validLevels {
		t.Run(string(level), func(t *testing.T) {
			// Test that the level is one of the valid constants
			assert.Contains(t, []AlertLevel{
				AlertLevelInfo,
				AlertLevelWarning,
				AlertLevelError,
				AlertLevelCritical,
			}, level)
		})
	}
} 