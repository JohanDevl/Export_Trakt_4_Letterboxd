package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/api"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/auth"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/export"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/i18n"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/scheduler"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/security"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/security/keyring"
	"github.com/robfig/cron/v3"
)

func main() {
	// Add panic recovery to catch unhandled errors
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PANIC: %v\n", r)
			os.Exit(1)
		}
	}()

	// Parse command line flags
	configPath := flag.String("config", "config/config.toml", "Path to configuration file")
	exportType := flag.String("export", "watched", "Type of export (watched, collection, shows, ratings, watchlist, all)")
	exportMode := flag.String("mode", "normal", "Export mode (normal, initial, complete)")
	runOnce := flag.Bool("run", false, "Run the script immediately once then exit")
	scheduleFlag := flag.String("schedule", "", "Run the script according to cron schedule format (e.g., '0 */6 * * *' for every 6 hours)")
	validateSecurity := flag.Bool("validate-security", false, "Validate security configuration and exit")
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
	if cfg.Logging.File != "" && os.Getenv("DISABLE_LOG_FILE") == "" {
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

	// Handle security validation flag
	if *validateSecurity {
		log.Info("security.validation_starting", nil)
		exitCode := validateSecurityConfiguration(cfg, log)
		os.Exit(exitCode)
	}

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

	// Initialize security manager and keyring
	securityManager, err := security.NewManager(cfg.Security)
	if err != nil {
		log.Error("errors.security_manager_failed", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}
	defer securityManager.Close()

	var keyringMgr *keyring.Manager
	switch cfg.Security.KeyringBackend {
	case "system":
		keyringMgr, err = keyring.NewManager(keyring.SystemBackend)
	case "env":
		keyringMgr, err = keyring.NewManager(keyring.EnvBackend)
	case "file":
		// For file backend, we need to provide options
		var options []keyring.Option
		if cfg.Security.EncryptionEnabled {
			// Generate a simple encryption key for demo purposes
			key := make([]byte, 32) // AES-256 key
			for i := range key {
				key[i] = byte(i % 256) // Simple pattern for demo
			}
			options = append(options, keyring.WithEncryptionKey(key))
		}
		options = append(options, keyring.WithFilePath("./config/credentials.enc"))
		keyringMgr, err = keyring.NewManager(keyring.FileBackend, options...)
	case "memory":
		keyringMgr, err = keyring.NewManager(keyring.MemoryBackend)
	default:
		keyringMgr, err = keyring.NewManager(keyring.SystemBackend)
	}
	if err != nil {
		log.Error("errors.keyring_manager_failed", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}

	// Initialize token manager
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)

	// Initialize Trakt client with token management
	var traktClient *api.Client
	if cfg.Auth.UseOAuth {
		traktClient = api.NewClientWithTokenManager(cfg, log, tokenManager)
	} else {
		traktClient = api.NewClient(cfg, log)
	}

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
		log.Info("export.starting_data_retrieval", map[string]interface{}{
			"export_type": *exportType,
		})

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

	case "auth":
		// Interactive OAuth authentication
		if err := runInteractiveAuth(cfg, log, tokenManager); err != nil {
			log.Error("auth.interactive_failed", map[string]interface{}{"error": err.Error()})
			fmt.Printf("❌ Authentication failed: %s\n", err.Error())
			os.Exit(1)
		}

	case "token-status":
		// Check token status
		if err := showTokenStatus(tokenManager); err != nil {
			log.Error("auth.status_failed", map[string]interface{}{"error": err.Error()})
			fmt.Printf("❌ Failed to check token status: %s\n", err.Error())
			os.Exit(1)
		}

	case "token-refresh":
		// Manual token refresh
		if err := refreshToken(tokenManager, log); err != nil {
			log.Error("auth.refresh_failed", map[string]interface{}{"error": err.Error()})
			fmt.Printf("❌ Token refresh failed: %s\n", err.Error())
			os.Exit(1)
		}

	case "token-clear":
		// Clear stored tokens
		if err := clearTokens(tokenManager, log); err != nil {
			log.Error("auth.clear_failed", map[string]interface{}{"error": err.Error()})
			fmt.Printf("❌ Failed to clear tokens: %s\n", err.Error())
			os.Exit(1)
		}

	case "auth-url":
		// Generate and display authentication URL
		if err := showAuthURL(cfg, log); err != nil {
			log.Error("auth.url_generation_failed", map[string]interface{}{"error": err.Error()})
			fmt.Printf("❌ Failed to generate auth URL: %s\n", err.Error())
			os.Exit(1)
		}

	case "auth-code":
		// Manual authentication with authorization code
		if len(flag.Args()) < 2 {
			fmt.Println("❌ Missing authorization code")
			fmt.Println("Usage: auth-code <authorization_code>")
			fmt.Println("Example: auth-code e2aa6bad787b30fd725e59e16ca52473515fd7ab38d6a7a71ff57fb6083c680d")
			os.Exit(1)
		}
		authCode := flag.Args()[1]
		if err := authenticateWithCode(cfg, log, tokenManager, authCode); err != nil {
			log.Error("auth.code_authentication_failed", map[string]interface{}{"error": err.Error()})
			fmt.Printf("❌ Authentication with code failed: %s\n", err.Error())
			os.Exit(1)
		}

	case "server":
		// Start persistent server with callback and export endpoints
		if err := startPersistentServer(cfg, log, tokenManager); err != nil {
			log.Error("server.start_failed", map[string]interface{}{"error": err.Error()})
			fmt.Printf("❌ Failed to start server: %s\n", err.Error())
			os.Exit(1)
		}

	case "fix-permissions":
		// Fix file permissions for credentials storage
		if err := fixCredentialsPermissions(cfg, log); err != nil {
			log.Error("permissions.fix_failed", map[string]interface{}{"error": err.Error()})
			fmt.Printf("❌ Failed to fix permissions: %s\n", err.Error())
			os.Exit(1)
		}
	
	default:
		log.Error("errors.invalid_command", map[string]interface{}{"command": command})
		fmt.Printf("Invalid command: %s. Valid commands are 'export', 'schedule', 'setup', 'validate', 'auth', 'auth-url', 'auth-code', 'server', 'fix-permissions', 'token-status', 'token-refresh', 'token-clear'\n", command)
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

	// Initialize security manager and keyring
	securityManager, err := security.NewManager(cfg.Security)
	if err != nil {
		log.Error("errors.security_manager_failed", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}
	defer securityManager.Close()

	var keyringMgr *keyring.Manager
	switch cfg.Security.KeyringBackend {
	case "system":
		keyringMgr, err = keyring.NewManager(keyring.SystemBackend)
	case "env":
		keyringMgr, err = keyring.NewManager(keyring.EnvBackend)
	case "file":
		// For file backend, we need to provide options
		var options []keyring.Option
		if cfg.Security.EncryptionEnabled {
			// Generate a simple encryption key for demo purposes
			key := make([]byte, 32) // AES-256 key
			for i := range key {
				key[i] = byte(i % 256) // Simple pattern for demo
			}
			options = append(options, keyring.WithEncryptionKey(key))
		}
		options = append(options, keyring.WithFilePath("./config/credentials.enc"))
		keyringMgr, err = keyring.NewManager(keyring.FileBackend, options...)
	case "memory":
		keyringMgr, err = keyring.NewManager(keyring.MemoryBackend)
	default:
		keyringMgr, err = keyring.NewManager(keyring.SystemBackend)
	}
	if err != nil {
		log.Error("errors.keyring_manager_failed", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}

	// Initialize token manager
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)

	// Check if authentication is required before proceeding
	if cfg.Auth.UseOAuth {
		log.Info("export.checking_authentication", nil)
		status, err := tokenManager.GetTokenStatus()
		if err != nil || !status.HasToken {
			log.Error("auth.token_missing", map[string]interface{}{
				"error": "No valid OAuth token found",
			})
			
			// Generate auth URL for display
			oauthMgr := auth.NewOAuthManager(cfg, log)
			authURL, _, urlErr := oauthMgr.GenerateAuthURL()
			
			fmt.Println("\n🔐 AUTHENTICATION REQUIRED")
			fmt.Println("==========================================")
			fmt.Printf("📱 Client ID: %s\n", cfg.Trakt.ClientID)
			fmt.Printf("🔗 Redirect URI: %s\n", cfg.Auth.RedirectURI)
			
			if urlErr == nil {
				fmt.Println("\n🚀 QUICK AUTHENTICATION:")
				fmt.Println("1. Open this URL in your browser:")
				fmt.Printf("   %s\n", authURL)
				fmt.Println("\n2. Authorize the application on Trakt.tv")
				fmt.Printf("3. Run this command to complete authentication:\n")
				fmt.Printf("   docker run --rm -v \"$(pwd)/config:/app/config\" -p %d:%d trakt-exporter auth\n", cfg.Auth.CallbackPort, cfg.Auth.CallbackPort)
			} else {
				fmt.Println("\n📋 TO AUTHENTICATE:")
				fmt.Println("1. Run the following command in a separate terminal:")
				fmt.Printf("   docker run --rm -v \"$(pwd)/config:/app/config\" -p %d:%d trakt-exporter auth\n", cfg.Auth.CallbackPort, cfg.Auth.CallbackPort)
				fmt.Println("\n2. Follow the OAuth authentication flow in your browser")
			}
			
			fmt.Println("\n4. Once authenticated, re-run this export command")
			fmt.Println("\n💡 Authentication only needs to be done once.")
			fmt.Println("   Tokens will be automatically refreshed afterwards.")
			fmt.Println("==========================================")
			os.Exit(1)
		}

		if !status.IsValid {
			log.Warn("auth.token_expired_will_refresh", map[string]interface{}{
				"expires_at": status.ExpiresAt,
			})
		}
	}

	// Initialize Trakt client with token management
	log.Info("export.initializing_trakt_client", nil)
	var traktClient *api.Client
	if cfg.Auth.UseOAuth {
		traktClient = api.NewClientWithTokenManager(cfg, log, tokenManager)
	} else {
		traktClient = api.NewClient(cfg, log)
	}
	
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

// getConfiguredTimezone returns the configured timezone or UTC as fallback
func getConfiguredTimezone(cfg *config.Config, log logger.Logger) *time.Location {
	// Try environment variable first (Docker TZ)
	if tz := os.Getenv("TZ"); tz != "" {
		if loc, err := time.LoadLocation(tz); err == nil {
			log.Info("scheduler.using_env_timezone", map[string]interface{}{
				"timezone": tz,
			})
			return loc
		}
		log.Warn("scheduler.invalid_env_timezone", map[string]interface{}{
			"timezone": tz,
		})
	}
	
	// Try config timezone
	if cfg.Export.Timezone != "" {
		if loc, err := time.LoadLocation(cfg.Export.Timezone); err == nil {
			log.Info("scheduler.using_config_timezone", map[string]interface{}{
				"timezone": cfg.Export.Timezone,
			})
			return loc
		}
		log.Warn("scheduler.invalid_config_timezone", map[string]interface{}{
			"timezone": cfg.Export.Timezone,
		})
	}
	
	// Fallback to UTC
	log.Info("scheduler.using_default_timezone", map[string]interface{}{
		"timezone": "UTC",
	})
	return time.UTC
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

	// Get configured timezone for display
	configuredTZ := getConfiguredTimezone(cfg, log)

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
		
		// Get next run time for display
		entries := c.Entries()
		var nextRunDisplay string
		if len(entries) > 0 {
			nextRun := entries[0].Next.In(configuredTZ)
			nextRunDisplay = nextRun.Format("2006-01-02 15:04:05 MST")
		}
		
		log.Info("scheduler.export_execution_completed", map[string]interface{}{
			"export_type": exportType,
			"export_mode": exportMode,
			"duration":    duration.String(),
			"next_run":    nextRunDisplay,
		})
		
		// Display visual completion message with next run
		fmt.Printf("\n✅ === EXPORT COMPLETED ===\n")
		fmt.Printf("⏱️  Duration: %s\n", duration.String())
		fmt.Printf("▶️  Next run: %s\n", nextRunDisplay)
		fmt.Printf("============================\n\n")
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
	
	// Get the next run time and display in configured timezone
	entries := c.Entries()
	if len(entries) > 0 {
		nextRun := entries[0].Next
		nextRunInTZ := nextRun.In(configuredTZ)
		
		log.Info("scheduler.started", map[string]interface{}{
			"schedule":        schedule,
			"entry_id":        entryID,
			"next_run":        nextRun.Format(time.RFC3339),
			"next_run_local":  nextRunInTZ.Format("2006-01-02 15:04:05 MST"),
			"timezone":        configuredTZ.String(),
		})
		fmt.Printf("\n🎯 === EXPORT SCHEDULER STARTED ===\n")
		fmt.Printf("⏰ Schedule: %s\n", schedule)
		fmt.Printf("📺 Export Type: %s\n", exportType)
		fmt.Printf("🔧 Export Mode: %s\n", exportMode)
		fmt.Printf("🌍 Timezone: %s\n", configuredTZ.String())
		fmt.Printf("▶️  Next run: %s\n", nextRunInTZ.Format("2006-01-02 15:04:05 MST"))
		fmt.Printf("=====================================\n\n")
		
		// Log upcoming executions for the next hour in configured timezone
		now := time.Now()
		oneHourLater := now.Add(time.Hour)
		log.Info("scheduler.upcoming_executions_preview", map[string]interface{}{
			"next_hour_from": now.Format(time.RFC3339),
			"next_hour_to":   oneHourLater.Format(time.RFC3339),
			"timezone":       configuredTZ.String(),
		})
		
		count := 0
		if len(entries) > 0 {
			entry := entries[0]
			nextExec := entry.Next
			for nextExec.Before(oneHourLater) && count < 10 {
				nextExecInTZ := nextExec.In(configuredTZ)
				log.Info("scheduler.upcoming_execution", map[string]interface{}{
					"execution_time": nextExecInTZ.Format("2006-01-02 15:04:05 MST"),
					"in_minutes":     int(time.Until(nextExec).Minutes()),
					"timezone":       configuredTZ.String(),
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

// validateSecurityConfiguration performs comprehensive security validation
func validateSecurityConfiguration(cfg *config.Config, log logger.Logger) int {
	fmt.Println("🔒 Security Configuration Validation")
	fmt.Println("=====================================")
	
	var errors []string
	var warnings []string
	
	// 1. Validate security configuration
	if err := cfg.Security.Validate(); err != nil {
		errors = append(errors, fmt.Sprintf("Security config validation failed: %v", err))
	} else {
		fmt.Println("✅ Security configuration is valid")
	}
	
	// 2. Check security level
	securityLevel := cfg.Security.SecurityLevel()
	switch securityLevel {
	case "high":
		fmt.Println("✅ Security level: HIGH - All security features enabled")
	case "medium":
		fmt.Println("⚠️  Security level: MEDIUM - Some security features disabled")
		warnings = append(warnings, "Consider enabling all security features for production use")
	case "low":
		fmt.Println("❌ Security level: LOW - Critical security features disabled")
		errors = append(errors, "Security level is too low for production use")
	}
	
	// 3. Test security manager initialization
	fmt.Println("\n🔧 Testing Security Manager...")
	securityManager, err := security.NewManager(cfg.Security)
	if err != nil {
		errors = append(errors, fmt.Sprintf("Security manager initialization failed: %v", err))
	} else {
		fmt.Println("✅ Security manager initialized successfully")
		
		// Test encryption if enabled
		if cfg.Security.EncryptionEnabled {
			testData := "test-encryption-data"
			encrypted, err := securityManager.EncryptData(testData)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Encryption test failed: %v", err))
			} else {
				decrypted, err := securityManager.DecryptData(encrypted)
				if err != nil {
					errors = append(errors, fmt.Sprintf("Decryption test failed: %v", err))
				} else if decrypted != testData {
					errors = append(errors, "Encryption/decryption round-trip failed")
				} else {
					fmt.Println("✅ Encryption/decryption test passed")
				}
			}
		}
		
		// Test input validation
		testInput := "<script>alert('xss')</script>"
		sanitized := securityManager.SanitizeInput(testInput)
		if sanitized == testInput {
			warnings = append(warnings, "Input sanitization may not be working properly")
		} else {
			fmt.Println("✅ Input sanitization working")
		}
		
		// Test file path validation
		maliciousPath := "../../../etc/passwd"
		if err := securityManager.ValidateFilePath(maliciousPath); err == nil {
			errors = append(errors, "Path traversal protection not working")
		} else {
			fmt.Println("✅ Path traversal protection working")
		}
		
		// Clean up
		if err := securityManager.Close(); err != nil {
			warnings = append(warnings, fmt.Sprintf("Security manager cleanup warning: %v", err))
		}
	}
	
	// 4. Check file permissions
	fmt.Println("\n📁 Checking File Permissions...")
	configFile := "config/config.toml"
	if info, err := os.Stat(configFile); err == nil {
		mode := info.Mode()
		if mode&0077 != 0 {
			warnings = append(warnings, fmt.Sprintf("Config file %s has overly permissive permissions: %v", configFile, mode))
		} else {
			fmt.Println("✅ Config file permissions are secure")
		}
	} else {
		fmt.Printf("ℹ️  Config file %s not found (using defaults)\n", configFile)
	}
	
	// 5. Check credential storage
	fmt.Println("\n🔑 Checking Credential Storage...")
	switch cfg.Security.KeyringBackend {
	case "system":
		fmt.Println("✅ Using system keyring (most secure)")
	case "env":
		fmt.Println("⚠️  Using environment variables for credentials")
		warnings = append(warnings, "Environment variables are less secure than system keyring")
		
		// Check if credentials are in config file
		if cfg.Trakt.ClientID != "" || cfg.Trakt.ClientSecret != "" {
			errors = append(errors, "Credentials found in config file while using env backend")
		}
	case "file":
		fmt.Println("⚠️  Using encrypted file for credentials")
		warnings = append(warnings, "File-based credential storage is less secure than system keyring")
	default:
		errors = append(errors, fmt.Sprintf("Unknown keyring backend: %s", cfg.Security.KeyringBackend))
	}
	
	// 6. Check HTTPS enforcement
	fmt.Println("\n🌐 Checking HTTPS Configuration...")
	if cfg.Security.RequireHTTPS {
		fmt.Println("✅ HTTPS enforcement enabled")
		
		// Check if API URL uses HTTPS
		if !strings.HasPrefix(cfg.Trakt.APIBaseURL, "https://") {
			errors = append(errors, "API base URL must use HTTPS when HTTPS enforcement is enabled")
		}
	} else {
		warnings = append(warnings, "HTTPS enforcement is disabled")
	}
	
	// 7. Check audit logging
	fmt.Println("\n📝 Checking Audit Configuration...")
	if cfg.Security.AuditLogging {
		fmt.Println("✅ Audit logging enabled")
		
		if cfg.Security.Audit.IncludeSensitive {
			warnings = append(warnings, "Audit logging includes sensitive information (not recommended for production)")
		}
		
		if cfg.Security.Audit.RetentionDays < 30 {
			warnings = append(warnings, "Audit log retention period is less than 30 days")
		}
	} else {
		warnings = append(warnings, "Audit logging is disabled")
	}
	
	// 8. Check rate limiting
	fmt.Println("\n🚦 Checking Rate Limiting...")
	if cfg.Security.RateLimitEnabled {
		fmt.Println("✅ Rate limiting enabled")
	} else {
		warnings = append(warnings, "Rate limiting is disabled")
	}
	
	// 9. Display summary
	fmt.Println("\n📊 Security Validation Summary")
	fmt.Println("==============================")
	
	if len(errors) == 0 && len(warnings) == 0 {
		fmt.Println("🎉 All security checks passed!")
		log.Info("security.validation_success", nil)
		return 0
	}
	
	if len(warnings) > 0 {
		fmt.Printf("⚠️  %d Warning(s):\n", len(warnings))
		for i, warning := range warnings {
			fmt.Printf("   %d. %s\n", i+1, warning)
		}
		fmt.Println()
	}
	
	if len(errors) > 0 {
		fmt.Printf("❌ %d Error(s):\n", len(errors))
		for i, error := range errors {
			fmt.Printf("   %d. %s\n", i+1, error)
		}
		fmt.Println()
		
		log.Error("security.validation_failed", map[string]interface{}{
			"error_count": len(errors),
			"warning_count": len(warnings),
		})
		
		fmt.Println("🔒 Security validation failed. Please fix the errors above.")
		return 1
	}
	
	log.Info("security.validation_warning", map[string]interface{}{
		"warning_count": len(warnings),
	})
	
	fmt.Println("⚠️  Security validation completed with warnings. Review recommendations above.")
	return 0
}

// runInteractiveAuth performs interactive OAuth authentication
func runInteractiveAuth(cfg *config.Config, log logger.Logger, tokenManager *auth.TokenManager) error {
	oauthMgr := auth.NewOAuthManager(cfg, log)
	
	fmt.Println("🔑 Starting Interactive OAuth Authentication")
	fmt.Println("==========================================")
	
	// Check if credentials are configured
	if cfg.Trakt.ClientID == "" || cfg.Trakt.ClientSecret == "" {
		fmt.Println("❌ Missing Trakt.tv API credentials")
		fmt.Println("\nPlease configure your Trakt.tv API credentials:")
		fmt.Println("1. Go to https://trakt.tv/oauth/applications")
		fmt.Println("2. Create a new application or modify existing one")
		fmt.Println("3. Set client_id and client_secret in your config file")
		fmt.Printf("4. Set redirect_uri to: %s\n", cfg.Auth.RedirectURI)
		return fmt.Errorf("missing API credentials")
	}
	
	fmt.Printf("📱 Client ID: %s\n", cfg.Trakt.ClientID)
	fmt.Printf("🔗 Redirect URI: %s\n", cfg.Auth.RedirectURI)
	
	// Start local callback server
	callbackURL, codeChan, errChan, err := oauthMgr.StartLocalCallbackServer()
	if err != nil {
		return fmt.Errorf("failed to start callback server: %w", err)
	}
	
	fmt.Printf("🌐 Local callback server started at: %s\n", callbackURL)
	
	// Generate authorization URL
	authURL, state, err := oauthMgr.GenerateAuthURL()
	if err != nil {
		return fmt.Errorf("failed to generate auth URL: %w", err)
	}
	
	fmt.Println("\n📋 NEXT STEPS:")
	fmt.Println("1. Open the following URL in your browser:")
	fmt.Printf("   %s\n\n", authURL)
	fmt.Println("2. Authorize the application on Trakt.tv")
	fmt.Println("3. You will be redirected back automatically")
	fmt.Println("\nWaiting for authorization...")
	
	// Wait for authorization code or error
	select {
	case code := <-codeChan:
		fmt.Println("✅ Authorization code received!")
		
		// Exchange code for token
		token, err := oauthMgr.ExchangeCodeForToken(code, state, state)
		if err != nil {
			return fmt.Errorf("failed to exchange code for token: %w", err)
		}
		
		// Store token
		if err := tokenManager.StoreToken(token); err != nil {
			return fmt.Errorf("failed to store token: %w", err)
		}
		
		fmt.Println("🎉 Authentication successful!")
		fmt.Printf("📅 Token expires: %s\n", oauthMgr.GetTokenExpiryTime(token).Format("2006-01-02 15:04:05"))
		fmt.Println("🔄 Automatic refresh is enabled")
		
		return nil
		
	case err := <-errChan:
		return fmt.Errorf("authentication error: %w", err)
		
	case <-time.After(5 * time.Minute):
		return fmt.Errorf("authentication timeout after 5 minutes")
	}
}

// showTokenStatus displays the current token status
func showTokenStatus(tokenManager *auth.TokenManager) error {
	fmt.Println("🔍 Token Status Check")
	fmt.Println("=====================")
	
	status, err := tokenManager.GetTokenStatus()
	if err != nil {
		return fmt.Errorf("failed to get token status: %w", err)
	}
	
	fmt.Println(status.String())
	
	if status.Error != "" {
		fmt.Printf("\n❌ Error: %s\n", status.Error)
	}
	
	if status.Message != "" {
		fmt.Printf("\n💡 Info: %s\n", status.Message)
	}
	
	if !status.HasToken {
		fmt.Println("\n🆘 No token found. Run 'auth' command to authenticate:")
		fmt.Println("   docker exec -it <container> /app/export-trakt auth")
	}
	
	return nil
}

// refreshToken manually refreshes the access token
func refreshToken(tokenManager *auth.TokenManager, log logger.Logger) error {
	fmt.Println("🔄 Refreshing Access Token")
	fmt.Println("===========================")
	
	// Check current status first
	status, err := tokenManager.GetTokenStatus()
	if err != nil {
		return fmt.Errorf("failed to get token status: %w", err)
	}
	
	if !status.HasToken {
		fmt.Println("❌ No token to refresh. Run 'auth' command first.")
		return fmt.Errorf("no token available")
	}
	
	if !status.HasRefreshToken {
		fmt.Println("❌ No refresh token available. Re-authentication required.")
		fmt.Println("Run: auth")
		return fmt.Errorf("no refresh token available")
	}
	
	if err := tokenManager.RefreshToken(); err != nil {
		return fmt.Errorf("token refresh failed: %w", err)
	}
	
	// Show new status
	newStatus, err := tokenManager.GetTokenStatus()
	if err != nil {
		return fmt.Errorf("failed to get new token status: %w", err)
	}
	
	fmt.Println("✅ Token refreshed successfully!")
	fmt.Printf("📅 New expiry: %s\n", newStatus.ExpiresAt.Format("2006-01-02 15:04:05"))
	
	return nil
}

// clearTokens removes all stored tokens
func clearTokens(tokenManager *auth.TokenManager, log logger.Logger) error {
	fmt.Println("🗑️  Clearing Stored Tokens")
	fmt.Println("===========================")
	
	if err := tokenManager.ClearToken(); err != nil {
		return fmt.Errorf("failed to clear tokens: %w", err)
	}
	
	fmt.Println("✅ All tokens cleared successfully!")
	fmt.Println("💡 Run 'auth' command to re-authenticate when needed.")
	
	return nil
}

// showAuthURL generates and displays the OAuth authentication URL
func showAuthURL(cfg *config.Config, log logger.Logger) error {
	oauthMgr := auth.NewOAuthManager(cfg, log)
	
	fmt.Println("🔗 OAuth Authentication URL Generator")
	fmt.Println("=====================================")
	
	// Check if credentials are configured
	if cfg.Trakt.ClientID == "" || cfg.Trakt.ClientSecret == "" {
		fmt.Println("❌ Missing Trakt.tv API credentials")
		fmt.Println("\nPlease configure your Trakt.tv API credentials:")
		fmt.Println("1. Go to https://trakt.tv/oauth/applications")
		fmt.Println("2. Create a new application or modify existing one")
		fmt.Println("3. Set client_id and client_secret in your config file")
		fmt.Printf("4. Set redirect_uri to: %s\n", cfg.Auth.RedirectURI)
		return fmt.Errorf("missing API credentials")
	}
	
	fmt.Printf("📱 Client ID: %s\n", cfg.Trakt.ClientID)
	fmt.Printf("🔗 Redirect URI: %s\n", cfg.Auth.RedirectURI)
	
	// Generate authorization URL
	authURL, state, err := oauthMgr.GenerateAuthURL()
	if err != nil {
		return fmt.Errorf("failed to generate auth URL: %w", err)
	}
	
	fmt.Println("\n🚀 AUTHENTICATION STEPS:")
	fmt.Println("1. Copy and open this URL in your browser:")
	fmt.Printf("   %s\n\n", authURL)
	fmt.Println("2. Authorize the application on Trakt.tv")
	fmt.Println("3. You will be redirected to localhost - this is normal")
	fmt.Println("4. Copy the 'code' parameter from the URL")
	fmt.Println("5. Run the interactive auth command:")
	fmt.Printf("   docker run --rm -v \"$(pwd)/config:/app/config\" -p %d:%d trakt-exporter auth\n", cfg.Auth.CallbackPort, cfg.Auth.CallbackPort)
	fmt.Println("\n💾 State (for security):", state)
	fmt.Println("\n💡 This URL is valid for 10 minutes.")
	
	return nil
}

// authenticateWithCode performs OAuth authentication using a provided authorization code
func authenticateWithCode(cfg *config.Config, log logger.Logger, tokenManager *auth.TokenManager, authCode string) error {
	oauthMgr := auth.NewOAuthManager(cfg, log)
	
	fmt.Println("🔑 Manual OAuth Authentication with Code")
	fmt.Println("=========================================")
	
	// Check if credentials are configured
	if cfg.Trakt.ClientID == "" || cfg.Trakt.ClientSecret == "" {
		fmt.Println("❌ Missing Trakt.tv API credentials")
		fmt.Println("\nPlease configure your Trakt.tv API credentials in config.toml")
		return fmt.Errorf("missing API credentials")
	}
	
	fmt.Printf("📱 Client ID: %s\n", cfg.Trakt.ClientID)
	fmt.Printf("🔗 Redirect URI: %s\n", cfg.Auth.RedirectURI)
	fmt.Printf("🔐 Authorization Code: %s\n", authCode)
	
	// Generate a state for this manual authentication (not validated since we're not using callback)
	_, state, err := oauthMgr.GenerateAuthURL()
	if err != nil {
		return fmt.Errorf("failed to generate state: %w", err)
	}
	
	fmt.Println("\n🔄 Exchanging authorization code for tokens...")
	
	// Exchange code for token (we'll use the generated state)
	token, err := oauthMgr.ExchangeCodeForToken(authCode, state, state)
	if err != nil {
		fmt.Printf("❌ Token exchange failed: %s\n", err.Error())
		fmt.Println("\n💡 Possible reasons:")
		fmt.Println("   - Authorization code has expired (they expire quickly)")
		fmt.Println("   - Authorization code has already been used")
		fmt.Println("   - Redirect URI mismatch in Trakt.tv app settings")
		fmt.Printf("   - Expected redirect URI: %s\n", cfg.Auth.RedirectURI)
		return err
	}
	
	// Store the token
	if err := tokenManager.StoreToken(token); err != nil {
		fmt.Printf("❌ Failed to store token: %s\n", err.Error())
		return err
	}
	
	fmt.Println("✅ Authentication successful!")
	fmt.Printf("📅 Token expires: %s\n", oauthMgr.GetTokenExpiryTime(token).Format("2006-01-02 15:04:05"))
	fmt.Println("🔄 Automatic refresh is enabled")
	fmt.Println("\n💡 You can now run export commands normally.")
	
	return nil
}

// startPersistentServer starts a persistent HTTP server that handles OAuth callbacks and export requests
func startPersistentServer(cfg *config.Config, log logger.Logger, tokenManager *auth.TokenManager) error {
	oauthMgr := auth.NewOAuthManager(cfg, log)
	
	port := cfg.Auth.CallbackPort
	if port == 0 {
		port = 8080
	}
	
	fmt.Println("🚀 Starting Persistent Export & OAuth Server")
	fmt.Println("=============================================")
	fmt.Printf("📱 Client ID: %s\n", cfg.Trakt.ClientID)
	fmt.Printf("🔗 Redirect URI: %s\n", cfg.Auth.RedirectURI)
	fmt.Printf("🌐 Server running on: http://0.0.0.0:%d\n", port)
	
	// Global state management for OAuth
	var currentState string
	var currentAuthURL string
	
	// Function to generate fresh auth URL and state
	generateFreshAuth := func() error {
		authURL, state, err := oauthMgr.GenerateAuthURL()
		if err != nil {
			return err
		}
		currentAuthURL = authURL
		currentState = state
		return nil
	}
	
	// Generate initial auth URL
	if err := generateFreshAuth(); err != nil {
		return fmt.Errorf("failed to generate auth URL: %w", err)
	}
	
	fmt.Println("\n🔗 OAUTH AUTHENTICATION:")
	fmt.Println("Open this URL in your browser to authenticate:")
	fmt.Printf("   %s\n", currentAuthURL)
	fmt.Println("\nAfter authentication, you can:")
	fmt.Printf("   • Visit http://192.168.1.24:%d/status for token status\n", port)
	fmt.Printf("   • Visit http://192.168.1.24:%d/export/watched for exports\n", port)
	fmt.Printf("   • Visit http://192.168.1.24:%d/health for server health\n", port)
	
	// OAuth callback handler
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		log.Info("server.oauth_callback_received", nil)
		
		code := r.URL.Query().Get("code")
		errorParam := r.URL.Query().Get("error")
		receivedState := r.URL.Query().Get("state")
		
		if errorParam != "" {
			errDescription := r.URL.Query().Get("error_description")
			log.Error("server.oauth_error", map[string]interface{}{
				"error": errorParam,
				"description": errDescription,
			})
			
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head><title>Authentication Error</title></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px;">
<h2>❌ Authentication Error</h2>
<p><strong>Error:</strong> %s</p>
<p><strong>Description:</strong> %s</p>
<p>Please close this window and try again.</p>
<p><a href="/auth-url">Get new authentication URL</a></p>
</body>
</html>`, errorParam, errDescription)
			return
		}

		if code == "" {
			log.Error("server.no_auth_code", nil)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<head><title>Authentication Error</title></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px;">
<h2>❌ Authentication Error</h2>
<p>No authorization code received.</p>
<p>Please close this window and try again.</p>
<p><a href="/auth-url">Get new authentication URL</a></p>
</body>
</html>`)
			return
		}

		// Exchange code for token
		token, err := oauthMgr.ExchangeCodeForToken(code, currentState, receivedState)
		if err != nil {
			log.Error("server.token_exchange_failed", map[string]interface{}{
				"error": err.Error(),
			})
			
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head><title>Authentication Failed</title></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px;">
<h2>❌ Authentication Failed</h2>
<p>Failed to exchange authorization code for token.</p>
<p><strong>Error:</strong> %s</p>
<p><a href="/auth-url">Try again</a></p>
</body>
</html>`, err.Error())
			return
		}

		// Store the token
		if err := tokenManager.StoreToken(token); err != nil {
			log.Error("server.token_store_failed", map[string]interface{}{
				"error": err.Error(),
			})
			
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head><title>Token Storage Failed</title></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px;">
<h2>⚠️ Token Storage Failed</h2>
<p>Authentication succeeded but failed to store token.</p>
<p><strong>Error:</strong> %s</p>
</body>
</html>`, err.Error())
			return
		}

		log.Info("server.oauth_success", map[string]interface{}{
			"expires_at": oauthMgr.GetTokenExpiryTime(token),
		})

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head><title>Authentication Success</title></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px;">
<h2>✅ Authentication Successful!</h2>
<p>You have successfully authenticated with Trakt.tv.</p>
<p><strong>Token expires:</strong> %s</p>
<p><strong>Automatic refresh:</strong> Enabled</p>
<h3>🚀 Available Actions:</h3>
<ul>
<li><a href="/status">Check token status</a></li>
<li><a href="/export/watched">Export watched movies</a></li>
<li><a href="/export/all">Export all data</a></li>
</ul>
<script>
setTimeout(function() {
	if (confirm("Authentication successful! Close this window?")) {
		window.close();
	}
}, 3000);
</script>
</body>
</html>`, oauthMgr.GetTokenExpiryTime(token).Format("2006-01-02 15:04:05"))
	})

	// Status endpoint
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		status, err := tokenManager.GetTokenStatus()
		if err != nil {
			http.Error(w, fmt.Sprintf("Error checking status: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if status.HasToken && status.IsValid {
			fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head><title>Token Status</title></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px;">
<h2>✅ Token Status: Valid</h2>
<p><strong>Expires:</strong> %s</p>
<p><strong>Auto-refresh:</strong> %s</p>
<h3>🚀 Available Actions:</h3>
<ul>
<li><a href="/export/watched">Export watched movies</a></li>
<li><a href="/export/collection">Export collection</a></li>
<li><a href="/export/shows">Export TV shows</a></li>
<li><a href="/export/ratings">Export ratings</a></li>
<li><a href="/export/watchlist">Export watchlist</a></li>
<li><a href="/export/all">Export all data</a></li>
</ul>
</body>
</html>`, status.ExpiresAt.Format("2006-01-02 15:04:05"), fmt.Sprintf("%t", status.HasRefreshToken))
		} else {
			fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head><title>Token Status</title></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px;">
<h2>❌ Token Status: Invalid or Missing</h2>
<p>You need to authenticate first.</p>
<p><a href="/auth-url">Get authentication URL</a></p>
</body>
</html>`)
		}
	})

	// Auth URL endpoint
	http.HandleFunc("/auth-url", func(w http.ResponseWriter, r *http.Request) {
		// Generate fresh auth URL and update current state
		if err := generateFreshAuth(); err != nil {
			http.Error(w, fmt.Sprintf("Error generating auth URL: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head><title>OAuth Authentication</title></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px;">
<h2>🔗 OAuth Authentication</h2>
<p>Click the link below to authenticate with Trakt.tv:</p>
<p><a href="%s" target="_blank" style="background: #e74c3c; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Authenticate with Trakt.tv</a></p>
<p><small>After authentication, you'll be redirected back to this server automatically.</small></p>
</body>
</html>`, currentAuthURL)
	})

	// Export endpoints
	http.HandleFunc("/export/", func(w http.ResponseWriter, r *http.Request) {
		exportType := strings.TrimPrefix(r.URL.Path, "/export/")
		if exportType == "" {
			exportType = "watched"
		}

		// Check authentication
		status, err := tokenManager.GetTokenStatus()
		if err != nil || !status.HasToken || !status.IsValid {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<head><title>Authentication Required</title></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px;">
<h2>🔐 Authentication Required</h2>
<p>You need to authenticate before exporting data.</p>
<p><a href="/auth-url">Authenticate with Trakt.tv</a></p>
</body>
</html>`)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head><title>Export Started</title></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px;">
<h2>🚀 Export Started</h2>
<p>Export type: <strong>%s</strong></p>
<p>The export is running in the background. Check the Docker logs for progress.</p>
<p>Files will be available in the exports directory when complete.</p>
<p><a href="/status">Back to status</a></p>
</body>
</html>`, exportType)

		// Trigger export in background
		go func() {
			log.Info("server.export_triggered", map[string]interface{}{
				"export_type": exportType,
				"client_ip": r.RemoteAddr,
			})
			
			runExportOnce(cfg, log, exportType, "normal")
		}()
	})

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"healthy","service":"export-trakt","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`)
	})

	// Root endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head><title>Export Trakt 4 Letterboxd Server</title></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px;">
<h1>🎬 Export Trakt 4 Letterboxd</h1>
<h2>Server Status: Running</h2>
<h3>🔗 Available Endpoints:</h3>
<ul>
<li><a href="/auth-url">Get OAuth authentication URL</a></li>
<li><a href="/status">Check token status</a></li>
<li><a href="/export/watched">Export watched movies</a></li>
<li><a href="/export/collection">Export collection</a></li>
<li><a href="/export/shows">Export TV shows</a></li>
<li><a href="/export/ratings">Export ratings</a></li>
<li><a href="/export/watchlist">Export watchlist</a></li>
<li><a href="/export/all">Export all data</a></li>
<li><a href="/health">Health check (JSON)</a></li>
</ul>
<h3>📱 Quick Authentication:</h3>
<p>Open this URL to authenticate: <br><a href="%s">%s</a></p>
</body>
</html>`, currentAuthURL, currentAuthURL)
	})

	// Start server
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
		Handler: nil,
	}

	log.Info("server.starting", map[string]interface{}{
		"port": port,
		"auth_url": currentAuthURL,
	})

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Info("server.shutdown_signal_received", map[string]interface{}{
			"signal": sig.String(),
		})
		fmt.Printf("\nReceived signal %s, shutting down server...\n", sig)
		
		if err := server.Close(); err != nil {
			log.Error("server.shutdown_error", map[string]interface{}{
				"error": err.Error(),
			})
		}
		
		log.Info("server.shutdown_complete", nil)
		os.Exit(0)
	}()

	fmt.Printf("\n✅ Server started! Press Ctrl+C to stop.\n\n")
	
	// Start HTTP server
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

func fixCredentialsPermissions(cfg *config.Config, log logger.Logger) error {
	credentialsPath := "./config/credentials.enc"
	
	fmt.Printf("🔧 Fixing credentials file permissions...\n\n")
	
	// Check if file exists
	info, err := os.Stat(credentialsPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("✅ Credentials file doesn't exist yet - no action needed.\n")
			fmt.Printf("   File will be created with proper permissions when you authenticate.\n")
			return nil
		}
		return fmt.Errorf("failed to check credentials file: %w", err)
	}
	
	currentMode := info.Mode()
	fmt.Printf("📋 Current file permissions: %o\n", currentMode&os.ModePerm)
	
	// Check Docker environment
	isDocker := false
	if _, err := os.Stat("/.dockerenv"); err == nil {
		isDocker = true
		fmt.Printf("🐳 Detected Docker environment\n")
	}
	
	// Determine target permissions
	targetMode := os.FileMode(0600)
	if isDocker {
		// In Docker, we might need more relaxed permissions
		if currentMode&0077 != 0 && currentMode&0044 == 0 {
			fmt.Printf("✅ File permissions are acceptable for Docker environment.\n")
			return nil
		}
		// Try to set more restrictive permissions, but accept failure in Docker
		targetMode = os.FileMode(0644)
	}
	
	fmt.Printf("🎯 Target permissions: %o\n", targetMode)
	
	// Try to change permissions
	if err := os.Chmod(credentialsPath, targetMode); err != nil {
		if isDocker {
			fmt.Printf("⚠️  Warning: Could not change file permissions in Docker environment.\n")
			fmt.Printf("   This is normal - Docker handles file permissions differently.\n")
			fmt.Printf("   Your credentials should still work properly.\n")
			return nil
		}
		return fmt.Errorf("failed to change file permissions: %w", err)
	}
	
	// Verify the change
	newInfo, err := os.Stat(credentialsPath)
	if err != nil {
		return fmt.Errorf("failed to verify permissions change: %w", err)
	}
	
	newMode := newInfo.Mode()
	fmt.Printf("✅ Permissions updated successfully: %o\n", newMode&os.ModePerm)
	
	fmt.Printf("\n💡 Tips:\n")
	fmt.Printf("   - If you're still having issues, try using the 'env' keyring backend\n")
	fmt.Printf("   - Set TRAKT_CLIENT_ID and TRAKT_CLIENT_SECRET environment variables\n")
	fmt.Printf("   - Update config.toml: keyring_backend = \"env\"\n")
	
	return nil
} 