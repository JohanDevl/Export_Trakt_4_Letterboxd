package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/api"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/export"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/i18n"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/scheduler"
	"github.com/robfig/cron/v3"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config/config.toml", "Path to configuration file")
	exportType := flag.String("export", "watched", "Type of export (watched, collection, shows, ratings, watchlist, all)")
	exportMode := flag.String("mode", "normal", "Export mode (normal, initial, complete)")
	runOnce := flag.Bool("run", false, "Run the script immediately once then exit")
	scheduleFlag := flag.String("schedule", "", "Run the script according to cron schedule format (e.g., '0 */6 * * *' for every 6 hours)")
	flag.Parse()

	// Get command from args
	command := "export" // Default command
	if len(flag.Args()) > 0 {
		command = flag.Args()[0]
	}

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

	// Handle --run flag (immediate execution)
	if *runOnce {
		log.Info("startup.run_once_mode", map[string]interface{}{
			"export_type": *exportType,
			"export_mode": *exportMode,
		})
		runExportOnce(cfg, log, *exportType, *exportMode)
		return
	}

	// Handle --schedule flag (cron scheduling)
	if *scheduleFlag != "" {
		log.Info("startup.schedule_mode", map[string]interface{}{
			"schedule":    *scheduleFlag,
			"export_type": *exportType,
			"export_mode": *exportMode,
		})
		runWithSchedule(cfg, log, *scheduleFlag, *exportType, *exportMode)
		return
	}

	log.Info("startup.starting", map[string]interface{}{
		"command": command,
		"mode": *exportMode, // Log the export mode
	})
	log.Info("startup.config_loaded", nil)

	// Initialize Trakt client
	traktClient := api.NewClient(cfg, log)

	// Process command
	switch strings.ToLower(command) {
	case "export":
		// Initialize Letterboxd exporter
		letterboxdExporter := export.NewLetterboxdExporter(cfg, log)

		// Log export mode
		log.Info("export.mode", map[string]interface{}{
			"mode": *exportMode,
		})

		// Perform the export based on type
		switch *exportType {
		case "watched":
			exportWatchedMovies(traktClient, letterboxdExporter, log)
		case "collection":
			exportCollection(traktClient, letterboxdExporter, log)
		case "shows":
			exportShows(traktClient, letterboxdExporter, log)
		case "ratings":
			exportRatings(traktClient, letterboxdExporter, log)
		case "watchlist":
			exportWatchlist(traktClient, letterboxdExporter, log)
		case "all":
			exportWatchedMovies(traktClient, letterboxdExporter, log)
			exportCollection(traktClient, letterboxdExporter, log)
			exportShows(traktClient, letterboxdExporter, log)
			exportRatings(traktClient, letterboxdExporter, log)
			exportWatchlist(traktClient, letterboxdExporter, log)
		default:
			log.Error("errors.invalid_export_type", map[string]interface{}{"type": *exportType})
			fmt.Printf("Invalid export type: %s. Valid types are 'watched', 'collection', 'shows', 'ratings', 'watchlist', or 'all'\n", *exportType)
			os.Exit(1)
		}

		fmt.Println(translator.Translate("app.description", nil))
	
	case "schedule":
		// Initialize scheduler
		sched := scheduler.NewScheduler(cfg, log)
		
		// Set export mode and type to environment variables for the scheduler
		os.Setenv("EXPORT_MODE", *exportMode)
		os.Setenv("EXPORT_TYPE", *exportType)
		
		// Start scheduler (this will block until the program is terminated)
		if err := sched.Start(); err != nil {
			log.Error("scheduler.start_failed", map[string]interface{}{"error": err.Error()})
			os.Exit(1)
		}
		
		// Block forever (or until SIGINT/SIGTERM)
		select {}
	
	case "setup":
		// Handle setup command - just inform for now
		fmt.Println(translator.Translate("setup.instructions", nil))
	
	case "validate":
		// Validate the configuration
		fmt.Println(translator.Translate("validate.success", nil))
	
	default:
		log.Error("errors.invalid_command", map[string]interface{}{"command": command})
		fmt.Printf("Invalid command: %s. Valid commands are 'export', 'schedule', 'setup', 'validate'\n", command)
		os.Exit(1)
	}
}

func exportWatchedMovies(client *api.Client, exporter *export.LetterboxdExporter, log logger.Logger) {
	// Get watched movies
	log.Info("export.retrieving_watched_movies", nil)
	movies, err := client.GetWatchedMovies()
	if err != nil {
		log.Error("errors.api_request_failed", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}

	log.Info("export.movies_retrieved", map[string]interface{}{"count": len(movies)})

	// If extended_info is set to "letterboxd", export in Letterboxd format
	if client.GetConfig().Trakt.ExtendedInfo == "letterboxd" {
		// Get ratings for Letterboxd format
		log.Info("export.retrieving_ratings", nil)
		ratings, err := client.GetRatings()
		if err != nil {
			log.Error("errors.api_request_failed", map[string]interface{}{"error": err.Error()})
			os.Exit(1)
		}
		
		log.Info("export.ratings_retrieved", map[string]interface{}{"count": len(ratings)})
		
		// Export in Letterboxd format
		log.Info("export.exporting_letterboxd_format", nil)
		if err := exporter.ExportLetterboxdFormat(movies, ratings); err != nil {
			log.Error("export.export_failed", map[string]interface{}{"error": err.Error()})
			os.Exit(1)
		}
		return
	}

	// Export movies in standard format
	log.Info("export.exporting_watched_movies", nil)
	if err := exporter.ExportMovies(movies); err != nil {
		log.Error("export.export_failed", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}
}

func exportCollection(client *api.Client, exporter *export.LetterboxdExporter, log logger.Logger) {
	// Get collection movies
	log.Info("export.retrieving_collection", nil)
	movies, err := client.GetCollectionMovies()
	if err != nil {
		log.Error("errors.api_request_failed", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}

	log.Info("export.collection_retrieved", map[string]interface{}{"count": len(movies)})

	// Export collection
	log.Info("export.exporting_collection", nil)
	if err := exporter.ExportCollectionMovies(movies); err != nil {
		log.Error("export.export_failed", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}
}

func exportShows(client *api.Client, exporter *export.LetterboxdExporter, log logger.Logger) {
	// Get watched shows
	log.Info("export.retrieving_watched_shows", nil)
	shows, err := client.GetWatchedShows()
	if err != nil {
		log.Error("errors.api_request_failed", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}

	// Count total episodes
	episodeCount := 0
	for _, show := range shows {
		for _, season := range show.Seasons {
			episodeCount += len(season.Episodes)
		}
	}

	log.Info("export.shows_retrieved", map[string]interface{}{
		"shows":    len(shows),
		"episodes": episodeCount,
	})

	// Export shows
	log.Info("export.exporting_shows", nil)
	if err := exporter.ExportShows(shows); err != nil {
		log.Error("export.export_failed", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}
}

func exportRatings(client *api.Client, exporter *export.LetterboxdExporter, log logger.Logger) {
	// Get ratings
	log.Info("export.retrieving_ratings", nil)
	ratings, err := client.GetRatings()
	if err != nil {
		log.Error("errors.api_request_failed", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}

	log.Info("export.ratings_retrieved", map[string]interface{}{"count": len(ratings)})

	// Export ratings
	log.Info("export.exporting_ratings", nil)
	if err := exporter.ExportRatings(ratings); err != nil {
		log.Error("export.export_failed", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}
}

func exportWatchlist(client *api.Client, exporter *export.LetterboxdExporter, log logger.Logger) {
	// Get watchlist
	log.Info("export.retrieving_watchlist", nil)
	watchlist, err := client.GetWatchlist()
	if err != nil {
		log.Error("errors.api_request_failed", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}

	log.Info("export.watchlist_retrieved", map[string]interface{}{"count": len(watchlist)})

	// Export watchlist
	log.Info("export.exporting_watchlist", nil)
	if err := exporter.ExportWatchlist(watchlist); err != nil {
		log.Error("export.export_failed", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}
}

// runExportOnce executes the export once and then exits
func runExportOnce(cfg *config.Config, log logger.Logger, exportType, exportMode string) {
	log.Info("export.starting_execution", map[string]interface{}{
		"export_type": exportType,
		"export_mode": exportMode,
		"timestamp":   time.Now().Format(time.RFC3339),
	})

	// Initialize Trakt client
	log.Info("export.initializing_trakt_client", nil)
	traktClient := api.NewClient(cfg, log)
	
	// Initialize Letterboxd exporter
	log.Info("export.initializing_letterboxd_exporter", nil)
	letterboxdExporter := export.NewLetterboxdExporter(cfg, log)
	
	// Log export mode
	log.Info("export.mode", map[string]interface{}{
		"mode": exportMode,
	})
	
	// Perform the export based on type
	log.Info("export.starting_data_retrieval", map[string]interface{}{
		"export_type": exportType,
	})

	switch exportType {
	case "watched":
		log.Info("export.executing_watched_movies", nil)
		exportWatchedMovies(traktClient, letterboxdExporter, log)
	case "collection":
		log.Info("export.executing_collection", nil)
		exportCollection(traktClient, letterboxdExporter, log)
	case "shows":
		log.Info("export.executing_shows", nil)
		exportShows(traktClient, letterboxdExporter, log)
	case "ratings":
		log.Info("export.executing_ratings", nil)
		exportRatings(traktClient, letterboxdExporter, log)
	case "watchlist":
		log.Info("export.executing_watchlist", nil)
		exportWatchlist(traktClient, letterboxdExporter, log)
	case "all":
		log.Info("export.executing_all_types", nil)
		exportWatchedMovies(traktClient, letterboxdExporter, log)
		exportCollection(traktClient, letterboxdExporter, log)
		exportShows(traktClient, letterboxdExporter, log)
		exportRatings(traktClient, letterboxdExporter, log)
		exportWatchlist(traktClient, letterboxdExporter, log)
	default:
		log.Error("errors.invalid_export_type", map[string]interface{}{"type": exportType})
		fmt.Printf("Invalid export type: %s. Valid types are 'watched', 'collection', 'shows', 'ratings', 'watchlist', or 'all'\n", exportType)
		os.Exit(1)
	}
	
	log.Info("export.completed_successfully", map[string]interface{}{
		"export_type": exportType,
		"export_mode": exportMode,
		"timestamp":   time.Now().Format(time.RFC3339),
	})
}

// runWithSchedule sets up a cron scheduler and runs the export according to the schedule
func runWithSchedule(cfg *config.Config, log logger.Logger, schedule, exportType, exportMode string) {
	log.Info("scheduler.initializing", map[string]interface{}{
		"schedule":    schedule,
		"export_type": exportType,
		"export_mode": exportMode,
	})

	// Check for verbose logging environment variable
	if os.Getenv("EXPORT_VERBOSE") == "true" {
		log.SetLogLevel("debug")
		log.Info("scheduler.verbose_mode_enabled", nil)
	}

	// Override log level if LOG_LEVEL environment variable is set
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		log.SetLogLevel(logLevel)
		log.Info("scheduler.log_level_set", map[string]interface{}{
			"level": logLevel,
		})
	}

	// Validate cron expression
	cronParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := cronParser.Parse(schedule)
	if err != nil {
		log.Error("errors.invalid_cron_schedule", map[string]interface{}{
			"schedule": schedule,
			"error":    err.Error(),
		})
		fmt.Printf("Invalid cron schedule format: %s\nError: %s\n", schedule, err.Error())
		fmt.Println("Example formats:")
		fmt.Println("  '0 */6 * * *'   - Every 6 hours")
		fmt.Println("  '0 9 * * 1'     - Every Monday at 9:00 AM")
		fmt.Println("  '30 14 * * *'   - Every day at 2:30 PM")
		os.Exit(1)
	}

	log.Info("scheduler.cron_validation_successful", map[string]interface{}{
		"schedule": schedule,
	})

	// Create a new cron scheduler
	c := cron.New()
	
	// Add the export job to the scheduler
	entryID, err := c.AddFunc(schedule, func() {
		log.Info("scheduler.job_triggered", map[string]interface{}{
			"schedule":    schedule,
			"export_type": exportType,
			"export_mode": exportMode,
			"timestamp":   time.Now().Format(time.RFC3339),
		})
		
		// Run the export with additional logging
		log.Info("scheduler.starting_export_execution", map[string]interface{}{
			"export_type": exportType,
			"export_mode": exportMode,
		})
		
		startTime := time.Now()
		runExportOnce(cfg, log, exportType, exportMode)
		duration := time.Since(startTime)
		
		log.Info("scheduler.export_execution_completed", map[string]interface{}{
			"export_type": exportType,
			"export_mode": exportMode,
			"duration":    duration.String(),
			"next_run":    c.Entries()[0].Next.Format(time.RFC3339),
		})
	})
	
	if err != nil {
		log.Error("errors.scheduler_add_failed", map[string]interface{}{
			"schedule": schedule,
			"error":    err.Error(),
		})
		fmt.Printf("Failed to add scheduled job: %s\n", err.Error())
		os.Exit(1)
	}
	
	log.Info("scheduler.job_added_successfully", map[string]interface{}{
		"entry_id": entryID,
		"schedule": schedule,
	})
	
	// Start the cron scheduler
	c.Start()
	log.Info("scheduler.cron_started", nil)
	
	// Get the next run time
	entries := c.Entries()
	if len(entries) > 0 {
		nextRun := entries[0].Next
		log.Info("scheduler.started", map[string]interface{}{
			"schedule":  schedule,
			"entry_id":  entryID,
			"next_run":  nextRun.Format(time.RFC3339),
			"next_run_local": nextRun.Format("2006-01-02 15:04:05 MST"),
		})
		fmt.Printf("Scheduler started successfully!\n")
		fmt.Printf("Schedule: %s\n", schedule)
		fmt.Printf("Export Type: %s\n", exportType)
		fmt.Printf("Export Mode: %s\n", exportMode)
		fmt.Printf("Next run: %s\n", nextRun.Format("2006-01-02 15:04:05 MST"))
		
		// Log upcoming executions for the next hour
		now := time.Now()
		oneHourLater := now.Add(time.Hour)
		log.Info("scheduler.upcoming_executions_preview", map[string]interface{}{
			"next_hour_from": now.Format(time.RFC3339),
			"next_hour_to":   oneHourLater.Format(time.RFC3339),
		})
		
		count := 0
		if len(entries) > 0 {
			entry := entries[0]
			nextExec := entry.Next
			for nextExec.Before(oneHourLater) && count < 10 {
				log.Info("scheduler.upcoming_execution", map[string]interface{}{
					"execution_time": nextExec.Format("2006-01-02 15:04:05 MST"),
					"in_minutes":     int(time.Until(nextExec).Minutes()),
				})
				// Calculate next execution after this one
				schedule, _ := cronParser.Parse(schedule)
				nextExec = schedule.Next(nextExec)
				count++
			}
		}
	}
	
	// Keep the program running until interrupted
	log.Info("scheduler.waiting", map[string]interface{}{
		"message": "Scheduler is running. Press Ctrl+C to stop.",
		"pid":     os.Getpid(),
	})
	fmt.Println("Scheduler is running. Press Ctrl+C to stop...")
	
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		sig := <-sigChan
		log.Info("scheduler.shutdown_signal_received", map[string]interface{}{
			"signal": sig.String(),
		})
		fmt.Printf("\nReceived signal %s, shutting down gracefully...\n", sig)
		c.Stop()
		log.Info("scheduler.shutdown_complete", nil)
		os.Exit(0)
	}()
	
	// Block forever (or until SIGINT/SIGTERM)
	select {}
} 