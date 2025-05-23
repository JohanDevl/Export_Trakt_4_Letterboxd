package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPrometheusMetrics(t *testing.T) {
	logger := logrus.New()
	pm := NewPrometheusMetrics(logger)
	
	assert.NotNil(t, pm)
	assert.Equal(t, logger, pm.logger)
	assert.NotNil(t, pm.registry)
	assert.NotNil(t, pm.exportsTotal)
	assert.NotNil(t, pm.exportDuration)
	assert.NotNil(t, pm.exportErrors)
	assert.NotNil(t, pm.apiCallsTotal)
	assert.NotNil(t, pm.apiCallDuration)
	assert.NotNil(t, pm.apiCallErrors)
	assert.NotNil(t, pm.moviesExported)
	assert.NotNil(t, pm.ratingsExported)
	assert.NotNil(t, pm.watchlistExported)
	assert.NotNil(t, pm.cacheHitRate)
	assert.NotNil(t, pm.goroutinesCount)
	assert.NotNil(t, pm.memoryUsage)
	assert.NotNil(t, pm.cpuUsage)
	assert.NotNil(t, pm.startTime)
	assert.NotNil(t, pm.healthStatus)
	assert.NotNil(t, pm.componentHealth)
	
	// Check that start time was set
	assert.Greater(t, testutil.ToFloat64(pm.startTime), float64(0))
}

func TestPrometheusMetrics_RegisterMetrics(t *testing.T) {
	logger := logrus.New()
	pm := NewPrometheusMetrics(logger)
	
	err := pm.RegisterMetrics()
	assert.NoError(t, err)
	
	// Test registering again should not cause error (metrics should be already registered)
	err = pm.RegisterMetrics()
	assert.Error(t, err) // Should error because metrics are already registered
}

func TestPrometheusMetrics_Name(t *testing.T) {
	logger := logrus.New()
	pm := NewPrometheusMetrics(logger)
	
	assert.Equal(t, "prometheus_metrics", pm.Name())
}

func TestPrometheusMetrics_GetRegistry(t *testing.T) {
	logger := logrus.New()
	pm := NewPrometheusMetrics(logger)
	
	registry := pm.GetRegistry()
	assert.NotNil(t, registry)
	assert.Equal(t, pm.registry, registry)
}

func TestPrometheusMetrics_GetHTTPHandler(t *testing.T) {
	logger := logrus.New()
	pm := NewPrometheusMetrics(logger)
	
	handler := pm.GetHTTPHandler()
	assert.NotNil(t, handler)
}

func TestPrometheusMetrics_RecordExport(t *testing.T) {
	logger := logrus.New()
	pm := NewPrometheusMetrics(logger)
	err := pm.RegisterMetrics()
	require.NoError(t, err)
	
	// Record an export
	pm.RecordExport("movies", "success", "csv", 5*time.Second)
	
	// Check counter value
	metric := testutil.ToFloat64(pm.exportsTotal.WithLabelValues("success", "movies", "csv"))
	assert.Equal(t, float64(1), metric)
}

func TestPrometheusMetrics_RecordExportError(t *testing.T) {
	logger := logrus.New()
	pm := NewPrometheusMetrics(logger)
	err := pm.RegisterMetrics()
	require.NoError(t, err)
	
	pm.RecordExportError("movies", "api_error", "trakt")
	
	metric := testutil.ToFloat64(pm.exportErrors.WithLabelValues("movies", "api_error", "trakt"))
	assert.Equal(t, float64(1), metric)
}

func TestPrometheusMetrics_RecordAPICall(t *testing.T) {
	logger := logrus.New()
	pm := NewPrometheusMetrics(logger)
	err := pm.RegisterMetrics()
	require.NoError(t, err)
	
	pm.RecordAPICall("trakt", "/movies", "GET", "200", 500*time.Millisecond)
	
	// Check counter
	counter := testutil.ToFloat64(pm.apiCallsTotal.WithLabelValues("trakt", "/movies", "GET", "200"))
	assert.Equal(t, float64(1), counter)
}

func TestPrometheusMetrics_RecordAPIError(t *testing.T) {
	logger := logrus.New()
	pm := NewPrometheusMetrics(logger)
	err := pm.RegisterMetrics()
	require.NoError(t, err)
	
	pm.RecordAPIError("trakt", "/movies", "timeout")
	
	metric := testutil.ToFloat64(pm.apiCallErrors.WithLabelValues("trakt", "/movies", "timeout"))
	assert.Equal(t, float64(1), metric)
}

func TestPrometheusMetrics_RecordMoviesExported(t *testing.T) {
	logger := logrus.New()
	pm := NewPrometheusMetrics(logger)
	err := pm.RegisterMetrics()
	require.NoError(t, err)
	
	pm.RecordMoviesExported("watched", "success", 10)
	
	metric := testutil.ToFloat64(pm.moviesExported.WithLabelValues("watched", "success"))
	assert.Equal(t, float64(10), metric)
	
	// Record more movies
	pm.RecordMoviesExported("watched", "success", 5)
	metric = testutil.ToFloat64(pm.moviesExported.WithLabelValues("watched", "success"))
	assert.Equal(t, float64(15), metric)
}

func TestPrometheusMetrics_RecordRatingsExported(t *testing.T) {
	logger := logrus.New()
	pm := NewPrometheusMetrics(logger)
	err := pm.RegisterMetrics()
	require.NoError(t, err)
	
	pm.RecordRatingsExported("5", 3)
	
	metric := testutil.ToFloat64(pm.ratingsExported.WithLabelValues("5"))
	assert.Equal(t, float64(3), metric)
}

func TestPrometheusMetrics_RecordWatchlistExported(t *testing.T) {
	logger := logrus.New()
	pm := NewPrometheusMetrics(logger)
	err := pm.RegisterMetrics()
	require.NoError(t, err)
	
	pm.RecordWatchlistExported("movie", 7)
	
	metric := testutil.ToFloat64(pm.watchlistExported.WithLabelValues("movie"))
	assert.Equal(t, float64(7), metric)
}

func TestPrometheusMetrics_UpdateCacheHitRate(t *testing.T) {
	logger := logrus.New()
	pm := NewPrometheusMetrics(logger)
	err := pm.RegisterMetrics()
	require.NoError(t, err)
	
	pm.UpdateCacheHitRate("movie_cache", 0.85)
	
	metric := testutil.ToFloat64(pm.cacheHitRate.WithLabelValues("movie_cache"))
	assert.Equal(t, 0.85, metric)
	
	// Update with new value
	pm.UpdateCacheHitRate("movie_cache", 0.92)
	metric = testutil.ToFloat64(pm.cacheHitRate.WithLabelValues("movie_cache"))
	assert.Equal(t, 0.92, metric)
}

func TestPrometheusMetrics_UpdateHealthStatus(t *testing.T) {
	logger := logrus.New()
	pm := NewPrometheusMetrics(logger)
	err := pm.RegisterMetrics()
	require.NoError(t, err)
	
	// Test healthy status
	pm.UpdateHealthStatus("1.0.0", true)
	
	metric := testutil.ToFloat64(pm.healthStatus.WithLabelValues("1.0.0"))
	assert.Equal(t, float64(1), metric)
	
	// Test unhealthy status
	pm.UpdateHealthStatus("1.0.0", false)
	metric = testutil.ToFloat64(pm.healthStatus.WithLabelValues("1.0.0"))
	assert.Equal(t, float64(0), metric)
}

func TestPrometheusMetrics_UpdateComponentHealth(t *testing.T) {
	logger := logrus.New()
	pm := NewPrometheusMetrics(logger)
	err := pm.RegisterMetrics()
	require.NoError(t, err)
	
	// Test healthy component
	pm.UpdateComponentHealth("trakt_api", "1.0.0", "healthy")
	metric := testutil.ToFloat64(pm.componentHealth.WithLabelValues("trakt_api", "1.0.0"))
	assert.Equal(t, float64(1), metric)
	
	// Test degraded component
	pm.UpdateComponentHealth("filesystem", "1.0.0", "degraded")
	metric = testutil.ToFloat64(pm.componentHealth.WithLabelValues("filesystem", "1.0.0"))
	assert.Equal(t, 0.5, metric)
	
	// Test unhealthy component
	pm.UpdateComponentHealth("database", "1.0.0", "unhealthy")
	metric = testutil.ToFloat64(pm.componentHealth.WithLabelValues("database", "1.0.0"))
	assert.Equal(t, float64(0), metric)
	
	// Test unknown status (defaults to unhealthy)
	pm.UpdateComponentHealth("unknown", "1.0.0", "unknown_status")
	metric = testutil.ToFloat64(pm.componentHealth.WithLabelValues("unknown", "1.0.0"))
	assert.Equal(t, float64(0), metric)
}

func TestPrometheusMetrics_CollectMetrics(t *testing.T) {
	logger := logrus.New()
	pm := NewPrometheusMetrics(logger)
	err := pm.RegisterMetrics()
	require.NoError(t, err)
	
	// Test collecting metrics
	ctx := context.Background()
	err = pm.CollectMetrics(ctx)
	assert.NoError(t, err)
	
	// Check that system metrics were updated
	goroutinesMetric := testutil.ToFloat64(pm.goroutinesCount)
	assert.Greater(t, goroutinesMetric, float64(0))
	
	// Check memory metrics
	allocMetric := testutil.ToFloat64(pm.memoryUsage.WithLabelValues("alloc"))
	assert.Greater(t, allocMetric, float64(0))
	
	heapAllocMetric := testutil.ToFloat64(pm.memoryUsage.WithLabelValues("heap_alloc"))
	assert.Greater(t, heapAllocMetric, float64(0))
	
	heapSysMetric := testutil.ToFloat64(pm.memoryUsage.WithLabelValues("heap_sys"))
	assert.Greater(t, heapSysMetric, float64(0))
}

func TestPrometheusMetrics_MetricsOutput(t *testing.T) {
	logger := logrus.New()
	pm := NewPrometheusMetrics(logger)
	err := pm.RegisterMetrics()
	require.NoError(t, err)
	
	// Record some sample data
	pm.RecordExport("movies", "success", "csv", 2*time.Second)
	pm.RecordMoviesExported("watched", "success", 5)
	pm.UpdateHealthStatus("1.0.0", true)
	
	// Collect metrics to ensure system metrics are populated
	err = pm.CollectMetrics(context.Background())
	require.NoError(t, err)
	
	// Get metrics from registry
	metricFamilies, err := pm.registry.Gather()
	require.NoError(t, err)
	
	// Check that we have metrics
	assert.Greater(t, len(metricFamilies), 0)
	
	// Check for specific metrics by name
	metricNames := make(map[string]bool)
	for _, mf := range metricFamilies {
		metricNames[mf.GetName()] = true
	}
	
	expectedMetrics := []string{
		"export_trakt_exports_total",
		"export_trakt_movies_exported_total",
		"export_trakt_health_status",
		"export_trakt_goroutines_count",
		"export_trakt_memory_usage_bytes",
		"export_trakt_start_time_seconds",
	}
	
	for _, expectedMetric := range expectedMetrics {
		assert.True(t, metricNames[expectedMetric], "Expected metric %s not found", expectedMetric)
	}
}

func TestPrometheusMetrics_ConcurrentAccess(t *testing.T) {
	logger := logrus.New()
	pm := NewPrometheusMetrics(logger)
	err := pm.RegisterMetrics()
	require.NoError(t, err)
	
	// Test concurrent metric recording
	done := make(chan bool, 100)
	
	// Start multiple goroutines recording different metrics
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				pm.RecordExport("movies", "success", "csv", time.Second)
				pm.RecordMoviesExported("watched", "success", 1)
				pm.UpdateCacheHitRate("movie_cache", 0.8)
				pm.UpdateHealthStatus("1.0.0", true)
			}
			done <- true
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Verify metrics were recorded correctly
	exportsCount := testutil.ToFloat64(pm.exportsTotal.WithLabelValues("success", "movies", "csv"))
	assert.Equal(t, float64(100), exportsCount)
	
	moviesCount := testutil.ToFloat64(pm.moviesExported.WithLabelValues("watched", "success"))
	assert.Equal(t, float64(100), moviesCount)
}

func TestPrometheusMetrics_MetricLabels(t *testing.T) {
	logger := logrus.New()
	pm := NewPrometheusMetrics(logger)
	err := pm.RegisterMetrics()
	require.NoError(t, err)
	
	// Test different label combinations
	pm.RecordExport("movies", "success", "csv", time.Second)
	pm.RecordExport("movies", "error", "json", 2*time.Second)
	pm.RecordExport("ratings", "success", "csv", 3*time.Second)
	
	// Check that different label combinations create separate metrics
	successMoviesCsv := testutil.ToFloat64(pm.exportsTotal.WithLabelValues("success", "movies", "csv"))
	errorMoviesJson := testutil.ToFloat64(pm.exportsTotal.WithLabelValues("error", "movies", "json"))
	successRatingsCsv := testutil.ToFloat64(pm.exportsTotal.WithLabelValues("success", "ratings", "csv"))
	
	assert.Equal(t, float64(1), successMoviesCsv)
	assert.Equal(t, float64(1), errorMoviesJson)
	assert.Equal(t, float64(1), successRatingsCsv)
}

func TestPrometheusMetrics_RegistryIsolation(t *testing.T) {
	logger := logrus.New()
	pm1 := NewPrometheusMetrics(logger)
	pm2 := NewPrometheusMetrics(logger)
	
	// Register metrics for both
	err1 := pm1.RegisterMetrics()
	err2 := pm2.RegisterMetrics()
	
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	
	// Record data in one instance
	pm1.RecordExport("movies", "success", "csv", time.Second)
	
	// The other instance should not see this data
	pm1Metric := testutil.ToFloat64(pm1.exportsTotal.WithLabelValues("success", "movies", "csv"))
	pm2Metric := testutil.ToFloat64(pm2.exportsTotal.WithLabelValues("success", "movies", "csv"))
	
	assert.Equal(t, float64(1), pm1Metric)
	assert.Equal(t, float64(0), pm2Metric)
	
	// Verify registries are different by checking they can both register metrics without conflict
	assert.NotNil(t, pm1.GetRegistry())
	assert.NotNil(t, pm2.GetRegistry())
} 