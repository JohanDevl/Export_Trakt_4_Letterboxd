package main

import (
	"fmt"
	"os"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/api"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/auth"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/export"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/security"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/security/keyring"
)

func exportWatchedMovies(client *api.Client, exporter *export.LetterboxdExporter, log logger.Logger, historyMode string) {
	// Determine which history mode to use
	effectiveHistoryMode := historyMode
	if effectiveHistoryMode == "" {
		effectiveHistoryMode = client.GetConfig().Export.HistoryMode
	}

	// Export based on history mode
	if effectiveHistoryMode == "individual" {
		// Get complete movie history (individual watch events)
		log.Info("export.retrieving_movie_history", nil)
		history, err := client.GetMovieHistory()
		if err != nil {
			log.Error("errors.api_request_failed", map[string]interface{}{"error": err.Error()})
			os.Exit(1)
		}

		log.Info("export.history_retrieved", map[string]interface{}{
			"count": len(history),
			"mode":  "individual",
		})

		// Export individual watch history
		log.Info("export.exporting_movie_history", nil)
		if err := exporter.ExportMovieHistory(history, client); err != nil {
			log.Error("export.export_failed", map[string]interface{}{"error": err.Error()})
			os.Exit(1)
		}
		return
	}

	// Default: aggregated mode (original behavior)
	log.Info("export.retrieving_watched_movies", map[string]interface{}{
		"mode": "aggregated",
	})
	movies, err := client.GetWatchedMovies()
	if err != nil {
		log.Error("errors.api_request_failed", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}

	log.Info("export.movies_retrieved", map[string]interface{}{
		"count": len(movies),
		"mode":  "aggregated",
	})

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
	if err := exporter.ExportMovies(movies, client); err != nil {
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
func runExportOnce(cfg *config.Config, log logger.Logger, exportType, exportMode, historyMode string) {
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
		exportWatchedMovies(traktClient, letterboxdExporter, log, historyMode)
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
		exportWatchedMovies(traktClient, letterboxdExporter, log, historyMode)
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
