package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/monitoring"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/monitoring/tracing"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/telemetry"
)

// simulateExportMovies simulates a movie export operation
func simulateExportMovies(ctx context.Context) error {
	fmt.Println("📽️ Starting movie export...")
	
	// Simulate some work
	time.Sleep(500 * time.Millisecond)
	
	// Simulate processing different movie categories
	categories := []string{"watched", "collected", "watchlist"}
	for _, category := range categories {
		fmt.Printf("  Processing %s movies...\n", category)
		time.Sleep(200 * time.Millisecond)
	}
	
	fmt.Println("✅ Movie export completed successfully")
	return nil
}

// simulateTraktAPICall simulates an API call to Trakt.tv
func simulateTraktAPICall(ctx context.Context) error {
	fmt.Println("🔌 Making API call to Trakt.tv...")
	
	// Simulate API call latency
	time.Sleep(300 * time.Millisecond)
	
	fmt.Println("✅ API call completed successfully")
	return nil
}

// simulateFailingOperation simulates an operation that fails
func simulateFailingOperation(ctx context.Context) error {
	fmt.Println("❌ Simulating a failing operation...")
	
	time.Sleep(100 * time.Millisecond)
	
	return fmt.Errorf("simulated export failure: network timeout")
}

func main() {
	fmt.Println("🚀 Export Trakt 4 Letterboxd - Monitoring Example")
	fmt.Println(repeat("=", 60))
	
	// Initialize telemetry configuration
	config := telemetry.TelemetryConfig{
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
			JaegerEndpoint: "", // Leave empty for this example
			SamplingRate:   1.0, // 100% sampling for demo
		},
		Logging: monitoring.LoggingConfig{
			Level:         "info",
			Format:        "visual", // Use visual format for better demo output
			Output:        "stdout",
			CorrelationID: true,
		},
		Alerts: monitoring.AlertsConfig{
			RateLimitMinutes: 1, // Short rate limit for demo
		},
	}

	// Initialize telemetry manager
	tm, err := telemetry.NewTelemetryManager(config, "1.0.0-demo")
	if err != nil {
		log.Fatalf("Failed to initialize telemetry: %v", err)
	}

	// Start telemetry services
	ctx := context.Background()
	if err := tm.Start(ctx); err != nil {
		log.Fatalf("Failed to start telemetry: %v", err)
	}
	defer tm.Stop(ctx)

	fmt.Printf("📊 Monitoring endpoints available at:\n")
	fmt.Printf("  - Health: http://localhost:9090/health\n")
	fmt.Printf("  - Metrics: http://localhost:9090/metrics\n")
	fmt.Printf("  - Alerts: http://localhost:9090/alerts/history\n")
	fmt.Printf("  - Ready: http://localhost:9090/health/ready\n")
	fmt.Printf("  - Live: http://localhost:9090/health/live\n\n")

	// Register custom health checker
	healthIntegration := telemetry.NewHealthCheckIntegration(tm)
	healthIntegration.RegisterCustomHealthChecker("demo_service", func(ctx context.Context) error {
		// Always healthy for this demo
		return nil
	})

	// Example 1: Successful export operation with tracing
	fmt.Println("🎬 Example 1: Successful Movie Export")
	fmt.Println(repeat("-", 40))
	
	err = telemetry.InstrumentedExportOperation(tm, ctx, "movies", func(ctx context.Context) error {
		// Simulate API call within the export operation
		return telemetry.InstrumentedAPICall(tm, ctx, "trakt", "/movies/watched", "GET", simulateTraktAPICall)
	})
	
	if err != nil {
		fmt.Printf("❌ Export failed: %v\n", err)
	} else {
		fmt.Println("✅ Export completed successfully")
	}
	fmt.Println()

	// Example 2: Failed export operation
	fmt.Println("🎬 Example 2: Failed Export Operation")
	fmt.Println(repeat("-", 40))
	
	err = telemetry.InstrumentedExportOperation(tm, ctx, "ratings", simulateFailingOperation)
	if err != nil {
		fmt.Printf("❌ Export failed as expected: %v\n", err)
	}
	fmt.Println()

	// Example 3: Custom metrics recording
	fmt.Println("📊 Example 3: Recording Custom Metrics")
	fmt.Println(repeat("-", 40))
	
	if metrics := tm.GetMetrics(); metrics != nil {
		// Record some business metrics
		metrics.RecordMoviesExported("watched", "success", 150)
		metrics.RecordMoviesExported("collected", "success", 87)
		metrics.RecordRatingsExported("5", 45)
		metrics.RecordRatingsExported("4", 67)
		metrics.UpdateCacheHitRate("movie_cache", 0.85)
		
		fmt.Println("✅ Custom metrics recorded")
	}
	fmt.Println()

	// Example 4: Health check demonstration
	fmt.Println("❤️ Example 4: Health Check Status")
	fmt.Println(repeat("-", 40))
	
	if healthChecker := tm.GetHealthChecker(); healthChecker != nil {
		health := healthChecker.Check(ctx)
		fmt.Printf("Overall Status: %s\n", health.Status)
		fmt.Printf("Uptime: %v\n", health.Uptime)
		fmt.Printf("Components checked: %d\n", len(health.Components))
		
		for name, component := range health.Components {
			fmt.Printf("  - %s: %s (%v)\n", name, component.Status, component.Duration)
		}
	}
	fmt.Println()

	// Example 5: Custom alert
	fmt.Println("🚨 Example 5: Sending Custom Alert")
	fmt.Println(repeat("-", 40))
	
	if alertManager := tm.GetAlertManager(); alertManager != nil {
		alert := alertManager.CreateAlert(
			monitoring.AlertLevelWarning,
			"Demo Alert",
			"This is a demonstration alert from the monitoring example",
			"demo_app",
			map[string]interface{}{
				"component":    "monitoring_example",
				"demo_metric": 42,
			},
		)
		
		if err := alertManager.SendAlert(ctx, alert); err != nil {
			fmt.Printf("❌ Failed to send alert: %v\n", err)
		} else {
			fmt.Println("✅ Demo alert sent successfully")
		}
	}
	fmt.Println()

	// Example 6: Middleware demonstration
	fmt.Println("⚙️ Example 6: Telemetry Middleware")
	fmt.Println(repeat("-", 40))
	
	middleware := telemetry.NewTelemetryMiddleware(tm)
	wrappedOperation := middleware.WrapOperation("demo_operation", func(ctx context.Context) error {
		fmt.Println("  Executing wrapped operation...")
		time.Sleep(300 * time.Millisecond)
		return nil
	})
	
	if err := wrappedOperation(ctx); err != nil {
		fmt.Printf("❌ Wrapped operation failed: %v\n", err)
	} else {
		fmt.Println("✅ Wrapped operation completed successfully")
	}
	fmt.Println()

	// Show final telemetry status
	fmt.Println("📈 Telemetry System Status")
	fmt.Println(repeat("-", 40))
	status := tm.GetTelemetryStatus()
	for key, value := range status {
		fmt.Printf("  %s: %v\n", key, value)
	}
	fmt.Println()

	// Keep the server running for a bit to allow testing endpoints
	fmt.Println("🔄 Keeping monitoring endpoints active for 30 seconds...")
	fmt.Println("   You can test the endpoints now:")
	fmt.Println("   curl http://localhost:9090/health")
	fmt.Println("   curl http://localhost:9090/metrics | head -20")
	fmt.Println()
	
	// Run health monitoring a few times to generate data
	for i := 0; i < 5; i++ {
		time.Sleep(3 * time.Second)
		fmt.Printf("⏰ Health check #%d...\n", i+1)
	}

	fmt.Println("🏁 Monitoring example completed!")
	fmt.Println("Check the metrics and health endpoints before the application shuts down.")
}

// Helper function to repeat strings (Go doesn't have this built-in)
func repeat(s string, count int) string {
	if count <= 0 {
		return ""
	}
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

// Override the * operator for strings using a helper function
var _ = repeat("=", 60) // This will be computed but not used, just for the pattern 