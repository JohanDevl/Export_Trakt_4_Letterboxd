package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/api"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/export"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config/config.toml", "Path to configuration file")
	flag.Parse()

	// Initialize logger
	log := logger.NewLogger()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Errorf("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	// Configure logger based on config
	log.SetLogLevel(cfg.Logging.Level)
	if cfg.Logging.File != "" {
		if err := log.SetLogFile(cfg.Logging.File); err != nil {
			log.Errorf("Failed to set log file: %v", err)
			os.Exit(1)
		}
	}

	log.Info("Starting Export Trakt 4 Letterboxd")

	// Initialize Trakt client
	traktClient := api.NewTraktClient(&cfg.Trakt, log)

	// Get watched movies
	movies, err := traktClient.GetWatchedMovies()
	if err != nil {
		log.Errorf("Failed to get watched movies: %v", err)
		os.Exit(1)
	}

	log.Infof("Retrieved %d movies from Trakt.tv", len(movies))

	// Initialize Letterboxd exporter
	letterboxdExporter := export.NewLetterboxdExporter(cfg, log)

	// Export movies
	if err := letterboxdExporter.ExportMovies(movies); err != nil {
		log.Errorf("Failed to export movies: %v", err)
		os.Exit(1)
	}

	fmt.Println("Export completed successfully!")
} 