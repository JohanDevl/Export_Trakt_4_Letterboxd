package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/api"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/export"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/i18n"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config/config.toml", "Path to configuration file")
	flag.Parse()

	// Initialize logger
	log := logger.NewLogger()

	// Load configuration
	log.Info("startup.loading_config", map[string]interface{}{"path": *configPath})
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Errorf("errors.config_load_failed", map[string]interface{}{"error": err})
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

	// Initialize translator
	translator, err := i18n.NewTranslator(&cfg.I18n, log)
	if err != nil {
		log.Errorf("Failed to initialize translator: %v", err)
		os.Exit(1)
	}

	// Update logger to use translator
	log.SetTranslator(translator)

	log.Info("startup.starting")
	log.Info("startup.config_loaded")

	// Initialize Trakt client
	traktClient := api.NewTraktClient(&cfg.Trakt, log)

	// Get watched movies
	log.Info("export.retrieving_movies")
	movies, err := traktClient.GetWatchedMovies()
	if err != nil {
		log.Errorf("errors.api_request_failed", map[string]interface{}{"error": err})
		os.Exit(1)
	}

	log.Info("export.movies_retrieved", map[string]interface{}{"count": len(movies)})

	// Initialize Letterboxd exporter
	letterboxdExporter := export.NewLetterboxdExporter(cfg, log)

	// Export movies
	log.Info("export.exporting_movies")
	if err := letterboxdExporter.ExportMovies(movies); err != nil {
		log.Errorf("export.export_failed", map[string]interface{}{"error": err})
		os.Exit(1)
	}

	fmt.Println(translator.Translate("app.description", nil))
} 