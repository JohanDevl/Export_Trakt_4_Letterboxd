# CLAUDE.md - Observabilité et Métriques

## Module Overview

Ce module fournit une stack complète d'observabilité avec métriques Prometheus, tracing OpenTelemetry, health checks, alerting et dashboards pour monitoring en temps réel de l'application.

## Architecture du Module

### 📊 Métriques Prometheus
```go
type PrometheusMetrics struct {
    registry        *prometheus.Registry
    apiRequests     *prometheus.CounterVec
    apiDuration     *prometheus.HistogramVec
    cacheHitRatio   prometheus.Gauge
    exportCount     *prometheus.CounterVec
    systemResources *prometheus.GaugeVec
}

func NewPrometheusMetrics() *PrometheusMetrics {
    return &PrometheusMetrics{
        apiRequests: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "trakt_api_requests_total",
                Help: "Total number of API requests",
            },
            []string{"endpoint", "status"},
        ),
        apiDuration: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:    "trakt_api_duration_seconds",
                Help:    "API request duration",
                Buckets: prometheus.DefBuckets,
            },
            []string{"endpoint"},
        ),
    }
}
```

#### Métriques Exposées
- **`trakt_api_requests_total`** : Compteur de requêtes API
- **`trakt_api_duration_seconds`** : Histogramme des latences
- **`trakt_cache_hit_ratio`** : Ratio de cache hits
- **`trakt_export_operations_total`** : Compteur d'exports
- **`trakt_system_memory_bytes`** : Utilisation mémoire
- **`trakt_worker_pool_active`** : Workers actifs

### 🔍 Tracing OpenTelemetry
```go
type TracingConfig struct {
    Enabled     bool   `toml:"enabled"`
    ServiceName string `toml:"service_name"`
    Endpoint    string `toml:"endpoint"`
    SampleRate  float64 `toml:"sample_rate"`
}

func InitTracing(cfg TracingConfig) (trace.TracerProvider, error) {
    exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(
        jaeger.WithEndpoint(cfg.Endpoint),
    ))
    
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithSampler(trace.TraceIDRatioBased(cfg.SampleRate)),
        trace.WithResource(resource.NewWithAttributes(
            semconv.ServiceNameKey.String(cfg.ServiceName),
        )),
    )
    
    return tp, nil
}
```

#### Spans Automatiques
```go
func TraceAPICall(ctx context.Context, endpoint string, fn func() error) error {
    tracer := otel.Tracer("trakt-api")
    ctx, span := tracer.Start(ctx, fmt.Sprintf("api.%s", endpoint))
    defer span.End()
    
    span.SetAttributes(
        attribute.String("api.endpoint", endpoint),
        attribute.String("api.method", "GET"),
    )
    
    err := fn()
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
    }
    
    return err
}
```

### 🏥 Health Checks
```go
type HealthChecker struct {
    checks map[string]HealthCheck
    mutex  sync.RWMutex
}

type HealthCheck interface {
    Name() string
    Check(ctx context.Context) HealthResult
}

type HealthResult struct {
    Status  HealthStatus `json:"status"`
    Message string       `json:"message,omitempty"`
    Details map[string]interface{} `json:"details,omitempty"`
    Latency time.Duration `json:"latency"`
}

type HealthStatus string

const (
    StatusHealthy   HealthStatus = "healthy"
    StatusUnhealthy HealthStatus = "unhealthy"
    StatusDegraded  HealthStatus = "degraded"
)
```

#### Health Checks Intégrés
- **API Connectivity** : Test de connectivité Trakt.tv
- **Auth Status** : Validation des tokens OAuth
- **Database** : État des connexions (si applicable)
- **Cache** : Performance et disponibilité du cache
- **Disk Space** : Espace disponible pour exports
- **Memory Usage** : Utilisation mémoire et GC

### 🚨 Système d'Alertes
```go
type AlertManager struct {
    rules     []AlertRule
    channels  []AlertChannel
    silences  map[string]time.Time
}

type AlertRule struct {
    Name        string
    Condition   string              // PromQL ou expression
    Threshold   float64
    Duration    time.Duration
    Severity    AlertSeverity
    Labels      map[string]string
}

type AlertSeverity string

const (
    SeverityInfo     AlertSeverity = "info"
    SeverityWarning  AlertSeverity = "warning"
    SeverityCritical AlertSeverity = "critical"
)
```

#### Règles d'Alerte Prédéfinies
```yaml
alerts:
  - name: "API Error Rate High"
    condition: "rate(trakt_api_requests_total{status!=\"200\"}[5m]) > 0.1"
    threshold: 0.1
    duration: "5m"
    severity: "warning"
    
  - name: "Cache Hit Ratio Low"
    condition: "trakt_cache_hit_ratio < 0.5"
    threshold: 0.5
    duration: "10m"
    severity: "info"
    
  - name: "Export Failures"
    condition: "increase(trakt_export_operations_total{status=\"error\"}[15m]) > 0"
    severity: "critical"
```

### 📈 Dashboard Configuration

#### Grafana Dashboard
```json
{
  "dashboard": {
    "title": "Export Trakt 4 Letterboxd",
    "panels": [
      {
        "title": "API Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(trakt_api_requests_total[5m])"
          }
        ]
      },
      {
        "title": "Cache Performance", 
        "type": "stat",
        "targets": [
          {
            "expr": "trakt_cache_hit_ratio"
          }
        ]
      },
      {
        "title": "Export Operations",
        "type": "table",
        "targets": [
          {
            "expr": "increase(trakt_export_operations_total[1h])"
          }
        ]
      }
    ]
  }
}
```

### 🔧 Configuration

#### Configuration Monitoring
```toml
[monitoring]
enabled = true
metrics_port = 8081
health_port = 8082

[monitoring.prometheus]
enabled = true
endpoint = "/metrics"
namespace = "trakt"
subsystem = "exporter"

[monitoring.tracing]
enabled = true
service_name = "export-trakt-4-letterboxd"
endpoint = "http://jaeger:14268/api/traces"
sample_rate = 0.1

[monitoring.health]
enabled = true
endpoint = "/health"
checks_interval = "30s"

[monitoring.alerts]
enabled = true
smtp_host = "smtp.gmail.com"
smtp_port = 587
webhook_url = "https://hooks.slack.com/services/..."
```

### 📊 Endpoints Monitoring

#### Métriques Endpoint
```
GET /metrics
Content-Type: text/plain

# HELP trakt_api_requests_total Total API requests
# TYPE trakt_api_requests_total counter
trakt_api_requests_total{endpoint="watched_movies",status="200"} 1247
trakt_api_requests_total{endpoint="watched_movies",status="429"} 3

# HELP trakt_api_duration_seconds API request duration
# TYPE trakt_api_duration_seconds histogram
trakt_api_duration_seconds_bucket{endpoint="watched_movies",le="0.1"} 156
trakt_api_duration_seconds_bucket{endpoint="watched_movies",le="0.5"} 891
```

#### Health Check Endpoint
```
GET /health
Content-Type: application/json

{
  "status": "healthy",
  "timestamp": "2025-07-11T15:43:22Z",
  "uptime": "2h34m12s",
  "checks": {
    "api_connectivity": {
      "status": "healthy",
      "latency": "89ms",
      "message": "Trakt API responding normally"
    },
    "auth_status": {
      "status": "healthy", 
      "details": {
        "token_valid": true,
        "expires_in": "2h15m"
      }
    },
    "cache": {
      "status": "healthy",
      "details": {
        "hit_ratio": 0.847,
        "size": "1.2MB",
        "entries": 456
      }
    }
  }
}
```

### 🚀 Utilisation

#### Activation Monitoring
```go
// Configuration monitoring
monitoringCfg := monitoring.Config{
    Enabled:     true,
    MetricsPort: 8081,
    HealthPort:  8082,
}

// Initialisation
monitor, err := monitoring.NewMonitor(monitoringCfg, log)
if err != nil {
    log.Fatal("Failed to initialize monitoring:", err)
}

// Démarrage
go monitor.Start()

// Enregistrement métriques custom
monitor.RecordAPICall("watched_movies", 200, 145*time.Millisecond)
monitor.UpdateCacheStats(0.85, 1024*1024)
monitor.RecordExport("movies", "success", 150)
```

#### Intégration avec Docker Compose
```yaml
version: '3.8'
services:
  export-trakt:
    image: johandevl/export-trakt-4-letterboxd:latest
    ports:
      - "8080:8080"    # App
      - "8081:8081"    # Metrics
      - "8082:8082"    # Health
    environment:
      - MONITORING_ENABLED=true
      
  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      
  grafana:
    image: grafana/grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
```

Ce module fournit une observabilité complète permettant un monitoring proactif, debugging efficace et optimisation continue des performances.