package metrics

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

// PrometheusMetrics implements the MetricsCollector interface
type PrometheusMetrics struct {
	logger   *logrus.Logger
	registry *prometheus.Registry
	
	// Application metrics
	exportsTotal        *prometheus.CounterVec
	exportDuration      *prometheus.HistogramVec
	exportErrors        *prometheus.CounterVec
	apiCallsTotal       *prometheus.CounterVec
	apiCallDuration     *prometheus.HistogramVec
	apiCallErrors       *prometheus.CounterVec
	
	// Business metrics
	moviesExported      *prometheus.CounterVec
	ratingsExported     *prometheus.CounterVec
	watchlistExported   *prometheus.CounterVec
	cacheHitRate        *prometheus.GaugeVec
	
	// System metrics
	goroutinesCount     prometheus.Gauge
	memoryUsage         *prometheus.GaugeVec
	cpuUsage            prometheus.Gauge
	startTime           prometheus.Gauge
	
	// Health metrics
	healthStatus        *prometheus.GaugeVec
	componentHealth     *prometheus.GaugeVec
}

// NewPrometheusMetrics creates a new Prometheus metrics collector
func NewPrometheusMetrics(logger *logrus.Logger) *PrometheusMetrics {
	registry := prometheus.NewRegistry()
	
	pm := &PrometheusMetrics{
		logger:   logger,
		registry: registry,
		
		// Application metrics
		exportsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "export_trakt_exports_total",
				Help: "Total number of exports performed",
			},
			[]string{"status", "type", "format"},
		),
		
		exportDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "export_trakt_export_duration_seconds",
				Help:    "Export operation duration in seconds",
				Buckets: []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0, 120.0, 300.0},
			},
			[]string{"type", "status"},
		),
		
		exportErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "export_trakt_export_errors_total",
				Help: "Total number of export errors",
			},
			[]string{"type", "error_code", "component"},
		),
		
		apiCallsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "export_trakt_api_calls_total",
				Help: "Total number of API calls made",
			},
			[]string{"service", "endpoint", "method", "status_code"},
		),
		
		apiCallDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "export_trakt_api_call_duration_seconds",
				Help:    "API call duration in seconds",
				Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
			},
			[]string{"service", "endpoint", "method"},
		),
		
		apiCallErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "export_trakt_api_call_errors_total",
				Help: "Total number of API call errors",
			},
			[]string{"service", "endpoint", "error_type"},
		),
		
		// Business metrics
		moviesExported: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "export_trakt_movies_exported_total",
				Help: "Total number of movies exported",
			},
			[]string{"category", "status"},
		),
		
		ratingsExported: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "export_trakt_ratings_exported_total",
				Help: "Total number of ratings exported",
			},
			[]string{"rating_value"},
		),
		
		watchlistExported: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "export_trakt_watchlist_exported_total",
				Help: "Total number of watchlist items exported",
			},
			[]string{"item_type"},
		),
		
		cacheHitRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "export_trakt_cache_hit_rate",
				Help: "Cache hit rate percentage",
			},
			[]string{"cache_type"},
		),
		
		// System metrics
		goroutinesCount: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "export_trakt_goroutines_count",
				Help: "Number of goroutines currently running",
			},
		),
		
		memoryUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "export_trakt_memory_usage_bytes",
				Help: "Memory usage in bytes",
			},
			[]string{"type"},
		),
		
		cpuUsage: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "export_trakt_cpu_usage_percent",
				Help: "CPU usage percentage",
			},
		),
		
		startTime: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "export_trakt_start_time_seconds",
				Help: "Start time of the application in Unix time",
			},
		),
		
		// Health metrics
		healthStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "export_trakt_health_status",
				Help: "Overall health status (1=healthy, 0.5=degraded, 0=unhealthy)",
			},
			[]string{"version"},
		),
		
		componentHealth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "export_trakt_component_health",
				Help: "Individual component health status (1=healthy, 0.5=degraded, 0=unhealthy)",
			},
			[]string{"component", "version"},
		),
	}
	
	// Set start time
	pm.startTime.Set(float64(time.Now().Unix()))
	
	return pm
}

// RegisterMetrics registers all metrics with the Prometheus registry
func (pm *PrometheusMetrics) RegisterMetrics() error {
	metrics := []prometheus.Collector{
		pm.exportsTotal,
		pm.exportDuration,
		pm.exportErrors,
		pm.apiCallsTotal,
		pm.apiCallDuration,
		pm.apiCallErrors,
		pm.moviesExported,
		pm.ratingsExported,
		pm.watchlistExported,
		pm.cacheHitRate,
		pm.goroutinesCount,
		pm.memoryUsage,
		pm.cpuUsage,
		pm.startTime,
		pm.healthStatus,
		pm.componentHealth,
	}
	
	for _, metric := range metrics {
		if err := pm.registry.Register(metric); err != nil {
			pm.logger.WithError(err).Error("Failed to register metric")
			return err
		}
	}
	
	pm.logger.Info("Prometheus metrics registered successfully")
	return nil
}

// CollectMetrics updates system metrics
func (pm *PrometheusMetrics) CollectMetrics(ctx context.Context) error {
	// Update goroutines count
	pm.goroutinesCount.Set(float64(runtime.NumGoroutine()))
	
	// Update memory metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	pm.memoryUsage.WithLabelValues("alloc").Set(float64(m.Alloc))
	pm.memoryUsage.WithLabelValues("total_alloc").Set(float64(m.TotalAlloc))
	pm.memoryUsage.WithLabelValues("sys").Set(float64(m.Sys))
	pm.memoryUsage.WithLabelValues("heap_alloc").Set(float64(m.HeapAlloc))
	pm.memoryUsage.WithLabelValues("heap_sys").Set(float64(m.HeapSys))
	pm.memoryUsage.WithLabelValues("heap_inuse").Set(float64(m.HeapInuse))
	
	return nil
}

// Name returns the name of the metrics collector
func (pm *PrometheusMetrics) Name() string {
	return "prometheus_metrics"
}

// GetRegistry returns the Prometheus registry
func (pm *PrometheusMetrics) GetRegistry() *prometheus.Registry {
	return pm.registry
}

// GetHTTPHandler returns the HTTP handler for serving metrics
func (pm *PrometheusMetrics) GetHTTPHandler() http.Handler {
	return promhttp.HandlerFor(pm.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// Business metric recording methods

// RecordExport records an export operation
func (pm *PrometheusMetrics) RecordExport(exportType, status, format string, duration time.Duration) {
	pm.exportsTotal.WithLabelValues(status, exportType, format).Inc()
	pm.exportDuration.WithLabelValues(exportType, status).Observe(duration.Seconds())
}

// RecordExportError records an export error
func (pm *PrometheusMetrics) RecordExportError(exportType, errorCode, component string) {
	pm.exportErrors.WithLabelValues(exportType, errorCode, component).Inc()
}

// RecordAPICall records an API call
func (pm *PrometheusMetrics) RecordAPICall(service, endpoint, method, statusCode string, duration time.Duration) {
	pm.apiCallsTotal.WithLabelValues(service, endpoint, method, statusCode).Inc()
	pm.apiCallDuration.WithLabelValues(service, endpoint, method).Observe(duration.Seconds())
}

// RecordAPIError records an API call error
func (pm *PrometheusMetrics) RecordAPIError(service, endpoint, errorType string) {
	pm.apiCallErrors.WithLabelValues(service, endpoint, errorType).Inc()
}

// RecordMoviesExported records movies exported
func (pm *PrometheusMetrics) RecordMoviesExported(category, status string, count int) {
	pm.moviesExported.WithLabelValues(category, status).Add(float64(count))
}

// RecordRatingsExported records ratings exported
func (pm *PrometheusMetrics) RecordRatingsExported(ratingValue string, count int) {
	pm.ratingsExported.WithLabelValues(ratingValue).Add(float64(count))
}

// RecordWatchlistExported records watchlist items exported
func (pm *PrometheusMetrics) RecordWatchlistExported(itemType string, count int) {
	pm.watchlistExported.WithLabelValues(itemType).Add(float64(count))
}

// UpdateCacheHitRate updates the cache hit rate
func (pm *PrometheusMetrics) UpdateCacheHitRate(cacheType string, hitRate float64) {
	pm.cacheHitRate.WithLabelValues(cacheType).Set(hitRate)
}

// UpdateHealthStatus updates the overall health status
func (pm *PrometheusMetrics) UpdateHealthStatus(version string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	pm.healthStatus.WithLabelValues(version).Set(value)
}

// UpdateComponentHealth updates individual component health
func (pm *PrometheusMetrics) UpdateComponentHealth(component, version string, status string) {
	var value float64
	switch status {
	case "healthy":
		value = 1.0
	case "degraded":
		value = 0.5
	case "unhealthy":
		value = 0.0
	default:
		value = 0.0
	}
	pm.componentHealth.WithLabelValues(component, version).Set(value)
} 