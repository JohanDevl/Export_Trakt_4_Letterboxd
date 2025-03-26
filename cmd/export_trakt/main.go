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
		log.Error("errors.config_load_failed", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}

	// Configure logger based on config
	log.SetLogLevel(cfg.Logging.Level)
	if cfg.Logging.File != "" {
		if err := log.SetLogFile(cfg.Logging.File); err != nil {
			log.Error("errors.log_file_failed", map[string]interface{}{"error": err.Error()})
			os.Exit(1)
		}
	}

	// Initialize translator
	translator, err := i18n.NewTranslator(&cfg.I18n, log)
	if err != nil {
		log.Error("errors.translator_failed", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}

	// Update logger to use translator
	log.SetTranslator(translator)

	log.Info("startup.starting", nil)
	log.Info("startup.config_loaded", nil)

	// Initialize Trakt client
	traktClient := api.NewClient(cfg, log)

	// Get watched movies
	log.Info("export.retrieving_movies", nil)
	movies, err := traktClient.GetWatchedMovies()
	if err != nil {
		log.Error("errors.api_request_failed", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}

	log.Info("export.movies_retrieved", map[string]interface{}{"count": len(movies)})

	// Initialize Letterboxd exporter
	letterboxdExporter := export.NewLetterboxdExporter(cfg, log)

	// Export movies
	log.Info("export.exporting_movies", nil)
	if err := letterboxdExporter.ExportMovies(movies); err != nil {
		log.Error("export.export_failed", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}

	fmt.Println(translator.Translate("app.description", nil))
} 