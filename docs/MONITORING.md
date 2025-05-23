# Monitoring and Observability

This document describes the comprehensive monitoring and observability system implemented for Export Trakt 4 Letterboxd.

## 🔍 Overview

The monitoring system provides complete visibility into application behavior through:

- **📊 Metrics Collection** - Prometheus metrics for performance and business metrics
- **🔍 Distributed Tracing** - OpenTelemetry tracing for request flow visibility
- **❤️ Health Monitoring** - Comprehensive health checks for all components
- **🚨 Alerting** - Smart alerting via webhooks, email, and Slack
- **📝 Structured Logging** - Enhanced logging with correlation IDs and context

## 🚀 Quick Start

### Basic Setup

```go
import (
    "context"
    "github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/telemetry"
)

// Initialize telemetry from application config
tm, err := telemetry.InitializeTelemetryFromConfig(appConfig, "1.0.0")
if err != nil {
    log.Fatal(err)
}

// Start telemetry services
ctx := context.Background()
if err := tm.Start(ctx); err != nil {
    log.Fatal(err)
}
defer tm.Stop(ctx)
```

### Instrumenting Operations

```go
// Trace an export operation with full telemetry
err := telemetry.InstrumentedExportOperation(tm, ctx, "movies", func(ctx context.Context) error {
    // Your export logic here
    return exportMovies(ctx)
})

// Trace API calls
err = telemetry.InstrumentedAPICall(tm, ctx, "trakt", "/movies/watched", "GET", func(ctx context.Context) error {
    // Your API call logic
    return callTraktAPI(ctx)
})
```

## 📊 Metrics

### Available Metrics

The system automatically collects the following metrics:

#### Application Metrics

- `export_trakt_exports_total` - Total number of exports performed
- `export_trakt_export_duration_seconds` - Export operation duration
- `export_trakt_export_errors_total` - Total number of export errors
- `export_trakt_api_calls_total` - Total number of API calls made
- `export_trakt_api_call_duration_seconds` - API call duration
- `export_trakt_api_call_errors_total` - Total number of API call errors

#### Business Metrics

- `export_trakt_movies_exported_total` - Total number of movies exported
- `export_trakt_ratings_exported_total` - Total number of ratings exported
- `export_trakt_watchlist_exported_total` - Total number of watchlist items exported
- `export_trakt_cache_hit_rate` - Cache hit rate percentage

#### System Metrics

- `export_trakt_goroutines_count` - Number of goroutines currently running
- `export_trakt_memory_usage_bytes` - Memory usage in bytes
- `export_trakt_cpu_usage_percent` - CPU usage percentage
- `export_trakt_start_time_seconds` - Start time of the application

#### Health Metrics

- `export_trakt_health_status` - Overall health status
- `export_trakt_component_health` - Individual component health status

### Accessing Metrics

Metrics are exposed at `http://localhost:9090/metrics` by default and can be scraped by Prometheus.

### Recording Custom Metrics

```go
// Record an export operation
tm.GetMetrics().RecordExport("movies", "success", "csv", duration)

// Record API call
tm.GetMetrics().RecordAPICall("trakt", "/movies", "GET", "200", duration)

// Record business metrics
tm.GetMetrics().RecordMoviesExported("watched", "success", 150)
```

## 🔍 Distributed Tracing

### OpenTelemetry Integration

The system uses OpenTelemetry for distributed tracing with optional Jaeger export.

### Configuration

```toml
[tracing]
enabled = true
service_name = "export-trakt-letterboxd"
service_version = "1.0.0"
environment = "production"
jaeger_endpoint = "http://localhost:14268/api/traces"
sampling_rate = 0.1  # 10% of traces
```

### Using Tracing

```go
// Manual span creation
ctx, endSpan := tm.GetTracer().StartSpan(ctx, "custom_operation")
defer endSpan()

// Add span events
tm.GetTracer().AddSpanEvent(ctx, "operation.step.completed", map[string]string{
    "step": "data_processing",
})

// Record errors
tm.GetTracer().AddSpanError(ctx, err)
```

### Trace Context

Traces automatically include:

- Correlation IDs
- Service information
- Operation metadata
- Error information
- Performance metrics

## ❤️ Health Monitoring

### Built-in Health Checks

The system includes several built-in health checks:

- **System Health** - Memory usage, goroutine count, CPU usage
- **Trakt API Health** - API connectivity and response time
- **File System Health** - Export directory accessibility
- **Component Health** - Individual component status

### Health Endpoints

- `GET /health` - Overall application health
- `GET /health/ready` - Kubernetes readiness probe
- `GET /health/live` - Kubernetes liveness probe

### Custom Health Checks

```go
// Register a custom health checker
healthIntegration := telemetry.NewHealthCheckIntegration(tm)
healthIntegration.RegisterCustomHealthChecker("database", func(ctx context.Context) error {
    // Check database connectivity
    return checkDatabaseConnection()
})
```

### Health Status Response

```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0",
  "uptime": "2h30m15s",
  "components": {
    "system": {
      "name": "system",
      "status": "healthy",
      "last_checked": "2024-01-15T10:30:00Z",
      "message": "Memory usage: 245.32 MB",
      "duration": "5ms",
      "details": {
        "memory_alloc_mb": 245.32,
        "goroutines": 25,
        "num_cpu": 8
      }
    }
  }
}
```

## 🚨 Alerting

### Alert Channels

The system supports multiple alert channels:

- **Webhook** - HTTP POST to custom endpoints
- **Email** - SMTP email alerts
- **Slack** - Slack webhook integration

### Alert Types

- **Export Alerts** - Success/failure notifications for export operations
- **Health Alerts** - Component health status changes
- **Custom Alerts** - Application-specific alerts

### Configuration

```toml
[alerts]
webhook_url = "https://hooks.example.com/alerts"
email_enabled = true
slack_enabled = true
rate_limit_minutes = 5  # Prevent alert spam
```

### Creating Custom Alerts

```go
alert := tm.GetAlertManager().CreateAlert(
    monitoring.AlertLevelWarning,
    "High Memory Usage",
    "Memory usage has exceeded 500MB",
    "system_monitor",
    map[string]interface{}{
        "memory_mb": 542.1,
        "threshold": 500,
    },
)

tm.GetAlertManager().SendAlert(ctx, alert)
```

### Alert History

View alert history at `http://localhost:9090/alerts/history`

## 📝 Enhanced Logging

### Structured Logging

The system provides structured logging with:

- **Correlation IDs** - Track requests across components
- **Trace Integration** - Automatic trace ID inclusion
- **Context Awareness** - Rich contextual information
- **Security** - Automatic PII sanitization

### Configuration

```toml
[logging]
level = "info"              # debug, info, warn, error, fatal
format = "json"             # visual, json, text
output = "stdout"           # stdout, stderr, file
correlation_id = true       # Enable correlation IDs
rotation_enabled = true     # Enable log rotation
max_age_days = 30          # Log retention
max_size_mb = 100          # Max log file size
max_backups = 3            # Number of backup files
```

### Using Structured Logging

```go
// Basic logging with context
tm.GetLogger().InfoWithContext(ctx, "export.started", map[string]interface{}{
    "operation_type": "movies",
    "user_id": "123",
})

// With structured fields
tm.GetLogger().WithFields(logrus.Fields{
    "component": "exporter",
    "operation": "movies",
}).Info("Export operation started")

// With correlation ID
ctx = tm.GetLogger().WithCorrelationID(ctx)
```

### Log Format Examples

#### JSON Format

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "info",
  "message": "Export operation started",
  "correlation_id": "550e8400-e29b-41d4-a716-446655440000",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7",
  "component": "exporter",
  "operation_type": "movies"
}
```

#### Visual Format

```
[2024-01-15 10:30:00] INFO Export operation started
  ├─ correlation_id: 550e8400-e29b-41d4-a716-446655440000
  ├─ trace_id: 4bf92f3577b34da6a3ce929d0e0e4736
  ├─ component: exporter
  └─ operation_type: movies
```

## 📈 Dashboards and Visualization

### Prometheus + Grafana

1. **Setup Prometheus** to scrape metrics from `:9090/metrics`
2. **Import Grafana dashboards** for visualization
3. **Configure alerts** in Grafana based on metrics

### Example Prometheus Configuration

```yaml
scrape_configs:
  - job_name: "export-trakt"
    static_configs:
      - targets: ["localhost:9090"]
    scrape_interval: 30s
    metrics_path: /metrics
```

### Key Metrics to Monitor

- Export success rate
- API call latency
- Memory usage trends
- Error rates
- Component health status

## 🔧 Configuration Reference

### Complete Configuration Example

```toml
[monitoring]
enabled = true
metrics_enabled = true
tracing_enabled = true
health_checks_enabled = true
metrics_port = 9090
metrics_path = "/metrics"

[tracing]
enabled = true
service_name = "export-trakt-letterboxd"
service_version = "1.0.0"
environment = "production"
jaeger_endpoint = "http://localhost:14268/api/traces"
sampling_rate = 0.1

[logging]
level = "info"
format = "json"
output = "stdout"
rotation_enabled = true
max_age_days = 30
max_size_mb = 100
max_backups = 3
correlation_id = true

[alerts]
webhook_url = "https://hooks.example.com/alerts"
email_enabled = false
slack_enabled = false
rate_limit_minutes = 5
```

### Environment Variables

- `JAEGER_ENDPOINT` - Override Jaeger endpoint
- `ALERT_WEBHOOK_URL` - Override alert webhook URL
- `MONITORING_PORT` - Override monitoring port
- `LOG_LEVEL` - Override log level

## 🐳 Docker and Kubernetes

### Docker Compose Example

```yaml
version: "3.8"
services:
  export-trakt:
    image: export-trakt:latest
    ports:
      - "9090:9090" # Monitoring endpoints
    environment:
      - JAEGER_ENDPOINT=http://jaeger:14268/api/traces
    volumes:
      - ./exports:/app/exports
      - ./logs:/app/logs

  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686" # Jaeger UI
      - "14268:14268" # Collector

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9091:9090" # Prometheus UI
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
```

### Kubernetes Health Checks

```yaml
apiVersion: v1
kind: Pod
spec:
  containers:
    - name: export-trakt
      image: export-trakt:latest
      ports:
        - containerPort: 9090
      livenessProbe:
        httpGet:
          path: /health/live
          port: 9090
        initialDelaySeconds: 30
        periodSeconds: 10
      readinessProbe:
        httpGet:
          path: /health/ready
          port: 9090
        initialDelaySeconds: 5
        periodSeconds: 5
```

## 🛠️ Troubleshooting

### Common Issues

1. **Metrics not appearing**

   - Check if monitoring is enabled in config
   - Verify port 9090 is accessible
   - Check Prometheus scrape configuration

2. **Traces not showing in Jaeger**

   - Verify Jaeger endpoint configuration
   - Check sampling rate (increase for testing)
   - Ensure tracing is enabled

3. **Health checks failing**

   - Check component dependencies
   - Verify file system permissions
   - Review health check thresholds

4. **Alerts not being sent**
   - Verify webhook/email configuration
   - Check rate limiting settings
   - Review alert manager logs

### Debug Commands

```bash
# Check health endpoint
curl http://localhost:9090/health

# Check metrics endpoint
curl http://localhost:9090/metrics

# Check alert history
curl http://localhost:9090/alerts/history

# View telemetry status
curl http://localhost:9090/telemetry/status
```

## 🔮 Future Enhancements

- **Advanced Dashboards** - Pre-built Grafana dashboards
- **Log Aggregation** - ELK stack integration
- **Advanced Alerting** - PagerDuty integration
- **Performance Profiling** - pprof integration
- **Custom Metrics** - Business-specific metrics
- **Real-time Monitoring** - WebSocket-based real-time updates

## 📚 Additional Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [OpenTelemetry Go Documentation](https://opentelemetry.io/docs/instrumentation/go/)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Logrus Documentation](https://github.com/sirupsen/logrus)

---

For more information or support, please check our [GitHub repository](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd) or create an issue.
