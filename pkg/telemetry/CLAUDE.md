# CLAUDE.md - Télémétrie et Observabilité

## Module Overview

Ce module intègre la télémétrie complète avec OpenTelemetry pour tracing distribué, métriques personnalisées, correlation des logs et observabilité end-to-end de l'application.

## Architecture du Module

### 📡 Télémétrie Intégrée
```go
type TelemetryManager struct {
    tracer         trace.Tracer
    meter          metric.Meter
    logger         logger.Logger
    config         TelemetryConfig
    correlationIDs map[string]string
}

type TelemetryConfig struct {
    ServiceName     string  `toml:"service_name"`
    ServiceVersion  string  `toml:"service_version"`
    Environment     string  `toml:"environment"`
    SampleRate      float64 `toml:"sample_rate"`
    JaegerEndpoint  string  `toml:"jaeger_endpoint"`
    PrometheusPort  int     `toml:"prometheus_port"`
}
```

### 🔍 Tracing Distribué

#### Spans Automatiques
```go
func (tm *TelemetryManager) TraceOperation(ctx context.Context, name string, fn func(context.Context) error) error {
    ctx, span := tm.tracer.Start(ctx, name)
    defer span.End()
    
    // Enrichissement automatique
    span.SetAttributes(
        attribute.String("service.name", tm.config.ServiceName),
        attribute.String("service.version", tm.config.ServiceVersion),
        attribute.String("environment", tm.config.Environment),
    )
    
    // Correlation ID
    if corrID := tm.getCorrelationID(ctx); corrID != "" {
        span.SetAttributes(attribute.String("correlation.id", corrID))
    }
    
    err := fn(ctx)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
    }
    
    return err
}
```

#### Instrumentation API
```go
func (tm *TelemetryManager) InstrumentAPICall(endpoint string, method string) func(context.Context) error {
    return func(ctx context.Context) error {
        return tm.TraceOperation(ctx, fmt.Sprintf("api.%s", endpoint), func(ctx context.Context) error {
            span := trace.SpanFromContext(ctx)
            span.SetAttributes(
                attribute.String("http.method", method),
                attribute.String("http.url", endpoint),
                attribute.String("component", "api_client"),
            )
            
            // Appel API réel ici
            return nil
        })
    }
}
```

### 📊 Métriques Custom

#### Instruments Telemetry
```go
type TelemetryInstruments struct {
    exportCounter    metric.Int64Counter
    apiDuration      metric.Float64Histogram  
    cacheHitRatio    metric.Float64Gauge
    activeOperations metric.Int64UpDownCounter
}

func (tm *TelemetryManager) InitInstruments() error {
    var err error
    
    tm.instruments.exportCounter, err = tm.meter.Int64Counter(
        "trakt_exports_total",
        metric.WithDescription("Total number of exports performed"),
    )
    
    tm.instruments.apiDuration, err = tm.meter.Float64Histogram(
        "trakt_api_duration_seconds",
        metric.WithDescription("API call duration in seconds"),
        metric.WithUnit("s"),
    )
    
    return err
}
```

### 🔗 Correlation des Logs

#### Context Enrichment
```go
type CorrelatedLogger struct {
    base   logger.Logger
    tracer trace.Tracer
}

func (cl *CorrelatedLogger) Info(key string, data map[string]interface{}) {
    // Enrichissement automatique avec span context
    if span := trace.SpanFromContext(context.Background()); span.SpanContext().IsValid() {
        if data == nil {
            data = make(map[string]interface{})
        }
        
        spanCtx := span.SpanContext()
        data["trace_id"] = spanCtx.TraceID().String()
        data["span_id"] = spanCtx.SpanID().String()
    }
    
    cl.base.Info(key, data)
}
```

### 🚀 Configuration

#### Télémétrie Config
```toml
[telemetry]
enabled = true
service_name = "export-trakt-4-letterboxd"
service_version = "1.0.0"
environment = "production"
sample_rate = 0.1

[telemetry.tracing]
jaeger_endpoint = "http://jaeger:14268/api/traces"
enabled = true

[telemetry.metrics]
prometheus_enabled = true
prometheus_port = 8081
custom_metrics = true

[telemetry.logging]
correlation_enabled = true
structured_format = true
```

### 📈 Dashboards

#### Grafana Observability
- **Traces** : Vue end-to-end des opérations
- **Métriques** : Performance et usage en temps réel
- **Logs** : Corrélation automatique avec traces
- **Alertes** : Notifications basées sur SLIs/SLOs

### 🚀 Usage

#### Initialisation Complète
```go
telemetryConfig := telemetry.Config{
    ServiceName:    "export-trakt-4-letterboxd",
    ServiceVersion: "1.0.0", 
    Environment:    "production",
    SampleRate:     0.1,
}

tm, err := telemetry.NewManager(telemetryConfig)
if err != nil {
    log.Fatal("Failed to initialize telemetry:", err)
}
defer tm.Shutdown()

// Traçage d'opération
ctx := context.Background()
err = tm.TraceOperation(ctx, "export.movies", func(ctx context.Context) error {
    return performMovieExport(ctx)
})
```

Ce module fournit une observabilité complète avec corrélation automatique entre traces, métriques et logs pour un monitoring et debugging efficaces.