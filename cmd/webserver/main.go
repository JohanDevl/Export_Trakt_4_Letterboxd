package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/webui"
)

var (
	configPath = flag.String("config", "./config/config.toml", "Path to configuration file")
	port       = flag.String("port", "8080", "Port to run the web server on")
	help       = flag.Bool("help", false, "Show help message")
)

func main() {
	flag.Parse()

	if *help {
		printHelp()
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	appLogger := logger.NewLogger()
	appLogger.SetLogLevel(cfg.Logging.Level)
	
	if cfg.Logging.File != "" {
		if err := appLogger.SetLogFile(cfg.Logging.File); err != nil {
			log.Printf("Warning: Failed to set log file: %v", err)
		}
	}

	appLogger.Info("webserver.starting", map[string]interface{}{
		"port":    *port,
		"config":  *configPath,
		"version": "2.0.0-dev",
	})

	// Create a simple monitoring manager placeholder
	// TODO: Replace with actual monitoring.NewManager() when available
	monitoringManager := &MockMonitoringManager{}

	// Create web server
	server, err := webui.NewServer(cfg, appLogger, monitoringManager)
	if err != nil {
		appLogger.Error("webserver.create_failed", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}

	// Channel to listen for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		appLogger.Info("webserver.listening", map[string]interface{}{
			"port": *port,
			"url":  fmt.Sprintf("http://localhost:%s", *port),
		})

		if err := server.Start(*port); err != nil {
			appLogger.Error("webserver.start_failed", map[string]interface{}{
				"error": err.Error(),
			})
			stop <- syscall.SIGTERM
		}
	}()

	// Print startup information
	printStartupInfo(*port)

	// Wait for interrupt signal
	<-stop

	appLogger.Info("webserver.shutting_down", nil)

	// Create a deadline for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Stop(ctx); err != nil {
		appLogger.Error("webserver.shutdown_failed", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}

	appLogger.Info("webserver.shutdown_complete", nil)
}

func printHelp() {
	fmt.Println("Export Trakt 4 Letterboxd - Web Interface")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Printf("  %s [flags]\n", os.Args[0])
	fmt.Println("")
	fmt.Println("Flags:")
	fmt.Println("  -config string")
	fmt.Println("        Path to configuration file (default \"./config/config.toml\")")
	fmt.Println("  -port string")
	fmt.Println("        Port to run the web server on (default \"8080\")")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  ./webserver")
	fmt.Println("  ./webserver -port 3000")
	fmt.Println("  ./webserver -config /path/to/config.toml -port 8080")
	fmt.Println("")
	fmt.Println("Environment Variables:")
	fmt.Println("  EXPORT_TRAKT_CONFIG_PATH - Path to configuration file")
	fmt.Println("  EXPORT_TRAKT_WEB_PORT    - Port for web server")
}

func printStartupInfo(port string) {
	fmt.Println("")
	fmt.Println("🎬 Export Trakt 4 Letterboxd - Web Interface")
	fmt.Println("===========================================")
	fmt.Printf("🌐 Web Interface: http://localhost:%s\n", port)
	fmt.Printf("📊 Metrics:       http://localhost:%s/api/v1/metrics\n", port)
	fmt.Printf("💓 Health:        http://localhost:%s/api/v1/health\n", port)
	fmt.Println("")
	fmt.Println("Press Ctrl+C to stop the server")
	fmt.Println("")
}

// MockMonitoringManager is a placeholder for the actual monitoring manager
type MockMonitoringManager struct{}

// Add any required methods for the monitoring interface here
// This is temporary until the actual monitoring.Manager is available 