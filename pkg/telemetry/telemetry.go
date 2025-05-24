package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/monitoring"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/monitoring/alerts"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/monitoring/health"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/monitoring/metrics"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/monitoring/tracing"
	"github.com/sirupsen/logrus"
)

// TelemetryConfig represents the complete telemetry configuration
type TelemetryConfig struct {
	Monitoring monitoring.MonitoringConfig   `toml:"monitoring"`
	Tracing    tracing.TracingConfig         `toml:"tracing"`
	Logging    monitoring.LoggingConfig      `toml:"logging"`
	Alerts     monitoring.AlertsConfig       `toml:"alerts"`
}

// TelemetryManager manages all telemetry components
type TelemetryManager struct {
	config         TelemetryConfig
	logger         *logger.StructuredLogger
	metrics        *metrics.PrometheusMetrics
	tracer         *tracing.OpenTelemetryTracer
	healthChecker  *health.HealthChecker
	alertManager   *alerts.AlertManager
	httpServer     *http.Server
	isRunning      bool
	mutex          sync.RWMutex
	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc
}

// NewTelemetryManager creates a new telemetry manager
func NewTelemetryManager(config TelemetryConfig, version string) (*TelemetryManager, error) {
	// Initialize structured logger
	structuredLogger, err := logger.NewStructuredLogger(config.Logging)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Create shutdown context
	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())

	tm := &TelemetryManager{
		config:         config,
		logger:         structuredLogger,
		shutdownCtx:    shutdownCtx,
		shutdownCancel: shutdownCancel,
	}

	// Initialize metrics if enabled
	if config.Monitoring.MetricsEnabled {
		tm.metrics = metrics.NewPrometheusMetrics(structuredLogger.Logger)
		if err := tm.metrics.RegisterMetrics(); err != nil {
			return nil, fmt.Errorf("failed to register metrics: %w", err)
		}
		structuredLogger.WithField("metrics_port", config.Monitoring.MetricsPort).Info("Metrics system initialized")
	}

	// Initialize tracing if enabled
	if config.Monitoring.TracingEnabled {
		tm.tracer, err = tracing.NewOpenTelemetryTracer(structuredLogger.Logger, config.Tracing)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize tracing: %w", err)
		}
		structuredLogger.Info("Tracing system initialized")
	}

	// Initialize health checker if enabled
	if config.Monitoring.HealthChecksEnabled {
		tm.healthChecker = health.NewHealthChecker(structuredLogger.Logger, version)
		
		// Register default health checkers
		tm.registerDefaultHealthCheckers()
		
		structuredLogger.Info("Health checking system initialized")
	}

	// Initialize alert manager
	tm.alertManager = alerts.NewAlertManager(structuredLogger.Logger, config.Alerts)
	structuredLogger.Info("Alert management system initialized")

	// Setup HTTP server for metrics and health endpoints
	if err := tm.setupHTTPServer(); err != nil {
		return nil, fmt.Errorf("failed to setup HTTP server: %w", err)
	}

	return tm, nil
}

// registerDefaultHealthCheckers registers default health checkers
func (tm *TelemetryManager) registerDefaultHealthCheckers() {
	if tm.healthChecker == nil {
		return
	}

	// Register Trakt API health checker
	traktAPIChecker := health.NewTraktAPIHealthChecker(nil) // TODO: Pass actual API client
	tm.healthChecker.RegisterChecker(traktAPIChecker)

	// Register file system health checker
	fsChecker := health.NewFileSystemHealthChecker("./exports")
	tm.healthChecker.RegisterChecker(fsChecker)

	// Register basic memory checker
	memoryChecker := health.NewBasicHealthChecker("memory", func(ctx context.Context) error {
		// Basic memory check - could be enhanced
		return nil
	})
	tm.healthChecker.RegisterChecker(memoryChecker)
}

// setupHTTPServer configures the HTTP server for telemetry endpoints
func (tm *TelemetryManager) setupHTTPServer() error {
	if !tm.config.Monitoring.Enabled {
		return nil
	}

	mux := http.NewServeMux()

	// Metrics endpoint
	if tm.metrics != nil {
		mux.Handle(tm.config.Monitoring.MetricsPath, tm.metrics.GetHTTPHandler())
		tm.logger.WithField("path", tm.config.Monitoring.MetricsPath).Info("Metrics endpoint registered")
	}

	// Health endpoints
	if tm.healthChecker != nil {
		mux.HandleFunc("/health", tm.healthChecker.HTTPHandler())
		mux.HandleFunc("/health/ready", tm.healthChecker.ReadinessHandler())
		mux.HandleFunc("/health/live", tm.healthChecker.LivenessHandler())
		tm.logger.Info("Health endpoints registered")
	}

	// Alert history endpoint
	mux.HandleFunc("/alerts", tm.alertHistoryHandler)
	mux.HandleFunc("/alerts/history", tm.alertHistoryHandler)

	// Create HTTP server
	tm.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", tm.config.Monitoring.MetricsPort),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return nil
}

// Start starts all telemetry components
func (tm *TelemetryManager) Start(ctx context.Context) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if tm.isRunning {
		return fmt.Errorf("telemetry manager is already running")
	}

	// Start HTTP server for metrics and health endpoints
	if tm.httpServer != nil {
		go func() {
			tm.logger.WithField("addr", tm.httpServer.Addr).Info("Starting telemetry HTTP server")
			if err := tm.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				tm.logger.WithError(err).Error("Telemetry HTTP server error")
			}
		}()
	}

	// Start metrics collection
	if tm.metrics != nil {
		go tm.startMetricsCollection(ctx)
	}

	// Start health monitoring
	if tm.healthChecker != nil {
		go tm.startHealthMonitoring(ctx)
	}

	tm.isRunning = true
	tm.logger.Info("Telemetry manager started successfully")

	return nil
}

// startMetricsCollection starts periodic metrics collection
func (tm *TelemetryManager) startMetricsCollection(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Collect metrics every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			tm.logger.Info("Stopping metrics collection")
			return
		case <-tm.shutdownCtx.Done():
			tm.logger.Info("Stopping metrics collection due to shutdown")
			return
		case <-ticker.C:
			if err := tm.metrics.CollectMetrics(ctx); err != nil {
				tm.logger.WithError(err).Error("Failed to collect metrics")
			}
		}
	}
}

// startHealthMonitoring starts periodic health monitoring
func (tm *TelemetryManager) startHealthMonitoring(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute) // Check health every minute
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			tm.logger.Info("Stopping health monitoring")
			return
		case <-tm.shutdownCtx.Done():
			tm.logger.Info("Stopping health monitoring due to shutdown")
			return
		case <-ticker.C:
			tm.performHealthCheck(ctx)
		}
	}
}

// performHealthCheck performs a health check and sends alerts if needed
func (tm *TelemetryManager) performHealthCheck(ctx context.Context) {
	overallHealth := tm.healthChecker.Check(ctx)

	// Update metrics
	if tm.metrics != nil {
		tm.metrics.UpdateHealthStatus("1.0", overallHealth.Status == monitoring.HealthStatusHealthy)
		
		for name, component := range overallHealth.Components {
			tm.metrics.UpdateComponentHealth(name, "1.0", string(component.Status))
		}
	}

	// Send alerts for unhealthy components
	for name, component := range overallHealth.Components {
		if component.Status != monitoring.HealthStatusHealthy {
			tm.alertManager.SendHealthAlert(ctx, name, component.Status, component.Message)
		}
	}

	// Log health status
	tm.logger.WithFields(logrus.Fields{
		"status":     overallHealth.Status,
		"uptime":     overallHealth.Uptime.String(),
		"components": len(overallHealth.Components),
	}).Info("Health check completed")
}

// Stop stops all telemetry components
func (tm *TelemetryManager) Stop(ctx context.Context) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if !tm.isRunning {
		return nil
	}

	tm.logger.Info("Stopping telemetry manager")

	// Signal shutdown to background goroutines
	tm.shutdownCancel()

	// Stop HTTP server
	if tm.httpServer != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		
		if err := tm.httpServer.Shutdown(shutdownCtx); err != nil {
			tm.logger.WithError(err).Error("Failed to shutdown HTTP server gracefully")
		} else {
			tm.logger.Info("HTTP server stopped")
		}
	}

	// Close tracing
	if tm.tracer != nil {
		if err := tm.tracer.Close(); err != nil {
			tm.logger.WithError(err).Error("Failed to close tracer")
		} else {
			tm.logger.Info("Tracer closed")
		}
	}

	tm.isRunning = false
	tm.logger.Info("Telemetry manager stopped")

	return nil
}

// GetLogger returns the structured logger
func (tm *TelemetryManager) GetLogger() *logger.StructuredLogger {
	return tm.logger
}

// GetMetrics returns the metrics collector
func (tm *TelemetryManager) GetMetrics() *metrics.PrometheusMetrics {
	return tm.metrics
}

// GetTracer returns the tracer
func (tm *TelemetryManager) GetTracer() *tracing.OpenTelemetryTracer {
	return tm.tracer
}

// GetHealthChecker returns the health checker
func (tm *TelemetryManager) GetHealthChecker() *health.HealthChecker {
	return tm.healthChecker
}

// GetAlertManager returns the alert manager
func (tm *TelemetryManager) GetAlertManager() *alerts.AlertManager {
	return tm.alertManager
}

// TraceExportOperation is a convenience method for tracing export operations
func (tm *TelemetryManager) TraceExportOperation(ctx context.Context, exportType string, fn func(ctx context.Context) error) error {
	if tm.tracer == nil {
		return fn(ctx)
	}
	
	start := time.Now()
	err := tm.tracer.TraceExportOperation(ctx, exportType, fn)
	duration := time.Since(start)
	
	// Record metrics
	if tm.metrics != nil {
		status := "success"
		if err != nil {
			status = "error"
		}
		tm.metrics.RecordExport(exportType, status, "csv", duration)
	}
	
	// Send alert if needed
	if tm.alertManager != nil {
		statusForAlert := "success"
		if err != nil {
			statusForAlert = "error"
		}
		tm.alertManager.SendExportAlert(ctx, exportType, statusForAlert, duration, err)
	}
	
	return err
}

// TraceAPICall is a convenience method for tracing API calls
func (tm *TelemetryManager) TraceAPICall(ctx context.Context, service, endpoint, method string, fn func(ctx context.Context) error) error {
	if tm.tracer == nil {
		return fn(ctx)
	}
	
	start := time.Now()
	err := tm.tracer.TraceAPICall(ctx, service, endpoint, method, fn)
	duration := time.Since(start)
	
	// Record metrics
	if tm.metrics != nil {
		statusCode := "200"
		if err != nil {
			statusCode = "500"
		}
		tm.metrics.RecordAPICall(service, endpoint, method, statusCode, duration)
	}
	
	return err
}

// alertHistoryHandler handles alert history HTTP requests
func (tm *TelemetryManager) alertHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if tm.alertManager == nil {
		http.Error(w, "Alert manager not available", http.StatusServiceUnavailable)
		return
	}

	history := tm.alertManager.GetAlertHistory()
	
	w.Header().Set("Content-Type", "application/json")
	
	// Simple JSON response - in production you'd use a proper JSON library
	response := `{"alerts": [`
	for i, alert := range history {
		if i > 0 {
			response += ","
		}
		response += fmt.Sprintf(`{
			"id": "%s",
			"level": "%s",
			"title": "%s",
			"message": "%s",
			"timestamp": "%s",
			"source": "%s",
			"resolved": %t
		}`, alert.ID, alert.Level, alert.Title, alert.Message, 
			alert.Timestamp.Format(time.RFC3339), alert.Source, alert.Resolved)
	}
	response += `]}`
	
	w.Write([]byte(response))
}

// IsRunning returns whether the telemetry manager is running
func (tm *TelemetryManager) IsRunning() bool {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	return tm.isRunning
}

// GetTelemetryStatus returns the current status of all telemetry components
func (tm *TelemetryManager) GetTelemetryStatus() map[string]interface{} {
	status := make(map[string]interface{})
	
	status["running"] = tm.IsRunning()
	status["monitoring_enabled"] = tm.config.Monitoring.Enabled
	status["metrics_enabled"] = tm.metrics != nil
	status["tracing_enabled"] = tm.tracer != nil
	status["health_checks_enabled"] = tm.healthChecker != nil
	status["alerts_enabled"] = tm.alertManager != nil
	
	if tm.httpServer != nil {
		status["http_server_addr"] = tm.httpServer.Addr
	}
	
	return status
} 