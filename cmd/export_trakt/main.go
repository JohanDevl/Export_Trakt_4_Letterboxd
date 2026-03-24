package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/api"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/auth"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/export"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/i18n"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/scheduler"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/security"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/security/keyring"
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
	historyMode := flag.String("history-mode", "", "History mode for watched export (aggregated, individual) - overrides config")
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
		runExportOnce(cfg, log, *exportType, *exportMode, *historyMode)
		return
	}

	// Handle --schedule flag (cron scheduling) only if not in server mode
	if *scheduleFlag != "" && (len(flag.Args()) == 0 || flag.Args()[0] != "server") {
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
		"mode":    *exportMode,
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
			exportWatchedMovies(traktClient, letterboxdExporter, log, *historyMode)
		case "collection":
			exportCollection(traktClient, letterboxdExporter, log)
		case "shows":
			exportShows(traktClient, letterboxdExporter, log)
		case "ratings":
			exportRatings(traktClient, letterboxdExporter, log)
		case "watchlist":
			exportWatchlist(traktClient, letterboxdExporter, log)
		case "all":
			exportWatchedMovies(traktClient, letterboxdExporter, log, *historyMode)
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
		if err := startPersistentServer(cfg, log, tokenManager, *scheduleFlag, *exportType, *exportMode); err != nil {
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
