package monitoring

import (
	"context"
	"time"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// ComponentHealth represents the health of a single component
type ComponentHealth struct {
	Name        string                 `json:"name"`
	Status      HealthStatus           `json:"status"`
	LastChecked time.Time             `json:"last_checked"`
	Message     string                `json:"message,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Duration    time.Duration         `json:"duration"`
}

// OverallHealth represents the overall application health
type OverallHealth struct {
	Status     HealthStatus                `json:"status"`
	Timestamp  time.Time                  `json:"timestamp"`
	Version    string                     `json:"version"`
	Uptime     time.Duration              `json:"uptime"`
	Components map[string]ComponentHealth `json:"components"`
}

// HealthChecker interface for components that can report their health
type HealthChecker interface {
	Check(ctx context.Context) ComponentHealth
	Name() string
}

// MetricsCollector interface for collecting application metrics
type MetricsCollector interface {
	CollectMetrics(ctx context.Context) error
	RegisterMetrics() error
	Name() string
}

// TracingProvider interface for distributed tracing
type TracingProvider interface {
	StartSpan(ctx context.Context, operationName string) (context.Context, func())
	AddSpanEvent(ctx context.Context, name string, attributes map[string]string)
	Close() error
}

// AlertLevel represents the severity of an alert
type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "info"
	AlertLevelWarning  AlertLevel = "warning"
	AlertLevelError    AlertLevel = "error"
	AlertLevelCritical AlertLevel = "critical"
)

// Alert represents a monitoring alert
type Alert struct {
	ID          string                 `json:"id"`
	Level       AlertLevel             `json:"level"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Timestamp   time.Time             `json:"timestamp"`
	Source      string                `json:"source"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Resolved    bool                  `json:"resolved"`
	ResolvedAt  *time.Time            `json:"resolved_at,omitempty"`
}

// NotificationChannel interface for sending alerts
type NotificationChannel interface {
	SendAlert(ctx context.Context, alert Alert) error
	Name() string
	IsEnabled() bool
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	Enabled           bool   `toml:"enabled"`
	MetricsEnabled    bool   `toml:"metrics_enabled"`
	TracingEnabled    bool   `toml:"tracing_enabled"`
	HealthChecksEnabled bool `toml:"health_checks_enabled"`
	MetricsPort       int    `toml:"metrics_port"`
	MetricsPath       string `toml:"metrics_path"`
}

// LoggingConfig represents enhanced logging configuration
type LoggingConfig struct {
	Level           string `toml:"level"`
	Format          string `toml:"format"`
	Output          string `toml:"output"`
	RotationEnabled bool   `toml:"rotation_enabled"`
	MaxAgeDays      int    `toml:"max_age_days"`
	MaxSizeMB       int    `toml:"max_size_mb"`
	MaxBackups      int    `toml:"max_backups"`
	CorrelationID   bool   `toml:"correlation_id"`
}

// AlertsConfig represents alerting configuration
type AlertsConfig struct {
	WebhookURL      string `toml:"webhook_url"`
	EmailEnabled    bool   `toml:"email_enabled"`
	SlackEnabled    bool   `toml:"slack_enabled"`
	RateLimitMinutes int   `toml:"rate_limit_minutes"`
} 