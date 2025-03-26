package benchmarks

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/api"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/export"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/i18n"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

// setupTestEnvironment creates a temporary directory and configuration for benchmarks
func setupTestEnvironment(b *testing.B) (string, *config.Config, func()) {
	// Create temporary directory for test outputs
	tempDir, err := os.MkdirTemp("", "export_trakt_benchmark_*")
	if err != nil {
		b.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create subdirectories
	os.MkdirAll(filepath.Join(tempDir, "exports"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "logs"), 0755)
	os.MkdirAll(filepath.Join(tempDir, "config"), 0755)

	// Create test configuration
	cfg := &config.Config{
		Trakt: config.TraktConfig{
			ClientID:     os.Getenv("TRAKT_CLIENT_ID"),
			ClientSecret: os.Getenv("TRAKT_CLIENT_SECRET"),
			RedirectURI:  "urn:ietf:wg:oauth:2.0:oob",
			TokenFile:    filepath.Join(tempDir, "config", "token.json"),
			Timeout:      30,
			MaxRetries:   3,
			RateLimit:    60,
		},
		Export: config.ExportConfig{
			OutputDir:         filepath.Join(tempDir, "exports"),
			FileFormat:        "trakt_{{type}}_{{timestamp}}.csv",
			Mode:              "normal",
			IncludeWatchlist:  true,
			IncludeCollections: false,
			IncludeRatings:    true,
			MinRating:         0,
			ConvertRatings:    true,
			KeepTempFiles:     false,
		},
		Logging: config.LoggingConfig{
			Level:   "info",
			File:    filepath.Join(tempDir, "logs", "export.log"),
			MaxSize: 10,
			MaxFiles: 3,
			Console: false,
			Color:   false,
		},
		I18n: config.I18nConfig{
			Language:   "en",
			LocalesDir: "../../locales",
		},
	}

	// Create cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cfg, cleanup
}

// BenchmarkGetWatchedMovies benchmarks the API client's GetWatchedMovies method
func BenchmarkGetWatchedMovies(b *testing.B) {
	// Skip if no API credentials
	if os.Getenv("TRAKT_CLIENT_ID") == "" || os.Getenv("TRAKT_CLIENT_SECRET") == "" {
		b.Skip("Skipping benchmark: No Trakt API credentials provided")
	}

	tempDir, cfg, cleanup := setupTestEnvironment(b)
	defer cleanup()

	log, err := logger.NewLogger(&cfg.Logging)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	client, err := api.NewTraktClient(&cfg.Trakt, log)
	if err != nil {
		b.Fatalf("Failed to create API client: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		movies, err := client.GetWatchedMovies()
		if err != nil {
			b.Fatalf("Failed to get watched movies: %v", err)
		}
		b.Logf("Retrieved %d watched movies", len(movies))
	}
}

// BenchmarkGetRatedMovies benchmarks the API client's GetRatedMovies method
func BenchmarkGetRatedMovies(b *testing.B) {
	// Skip if no API credentials
	if os.Getenv("TRAKT_CLIENT_ID") == "" || os.Getenv("TRAKT_CLIENT_SECRET") == "" {
		b.Skip("Skipping benchmark: No Trakt API credentials provided")
	}

	tempDir, cfg, cleanup := setupTestEnvironment(b)
	defer cleanup()

	log, err := logger.NewLogger(&cfg.Logging)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	client, err := api.NewTraktClient(&cfg.Trakt, log)
	if err != nil {
		b.Fatalf("Failed to create API client: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		movies, err := client.GetRatedMovies()
		if err != nil {
			b.Fatalf("Failed to get rated movies: %v", err)
		}
		b.Logf("Retrieved %d rated movies", len(movies))
	}
}

// BenchmarkGetWatchlistMovies benchmarks the API client's GetWatchlistMovies method
func BenchmarkGetWatchlistMovies(b *testing.B) {
	// Skip if no API credentials
	if os.Getenv("TRAKT_CLIENT_ID") == "" || os.Getenv("TRAKT_CLIENT_SECRET") == "" {
		b.Skip("Skipping benchmark: No Trakt API credentials provided")
	}

	tempDir, cfg, cleanup := setupTestEnvironment(b)
	defer cleanup()

	log, err := logger.NewLogger(&cfg.Logging)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	client, err := api.NewTraktClient(&cfg.Trakt, log)
	if err != nil {
		b.Fatalf("Failed to create API client: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		movies, err := client.GetWatchlistMovies()
		if err != nil {
			b.Fatalf("Failed to get watchlist movies: %v", err)
		}
		b.Logf("Retrieved %d watchlist movies", len(movies))
	}
}

// BenchmarkExportWatchedMovies benchmarks the export package's ExportWatchedMovies function
func BenchmarkExportWatchedMovies(b *testing.B) {
	// Skip if no API credentials
	if os.Getenv("TRAKT_CLIENT_ID") == "" || os.Getenv("TRAKT_CLIENT_SECRET") == "" {
		b.Skip("Skipping benchmark: No Trakt API credentials provided")
	}

	tempDir, cfg, cleanup := setupTestEnvironment(b)
	defer cleanup()

	log, err := logger.NewLogger(&cfg.Logging)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	client, err := api.NewTraktClient(&cfg.Trakt, log)
	if err != nil {
		b.Fatalf("Failed to create API client: %v", err)
	}

	translator, err := i18n.NewTranslator(&cfg.I18n, log)
	if err != nil {
		b.Fatalf("Failed to create translator: %v", err)
	}

	log.SetTranslator(translator)

	// Define export options
	options := &export.ExportOptions{
		OutputDir:          cfg.Export.OutputDir,
		IncludeRatings:     cfg.Export.IncludeRatings,
		IncludeWatchlist:   cfg.Export.IncludeWatchlist,
		IncludeCollections: cfg.Export.IncludeCollections,
		MinRating:          cfg.Export.MinRating,
		ConvertRatings:     cfg.Export.ConvertRatings,
		KeepTempFiles:      cfg.Export.KeepTempFiles,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := export.ExportWatchedMovies(client, log, options)
		if err != nil {
			b.Fatalf("Failed to export watched movies: %v", err)
		}
	}
}

// BenchmarkCSVGeneration benchmarks the generation of CSV files from movie data
func BenchmarkCSVGeneration(b *testing.B) {
	tempDir, cfg, cleanup := setupTestEnvironment(b)
	defer cleanup()

	log, err := logger.NewLogger(&cfg.Logging)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	// Create sample movie data for benchmark
	var watchedMovies []api.WatchedMovie
	for i := 0; i < 1000; i++ {
		watchedMovies = append(watchedMovies, api.WatchedMovie{
			Movie: api.Movie{
				Title: fmt.Sprintf("Movie %d", i),
				Year:  2000 + (i % 23),
				IDs: api.MovieIDs{
					Trakt: i,
					TMDB:  i * 10,
				},
			},
			WatchedAt: time.Now().AddDate(0, 0, -i%365).Format(time.RFC3339),
		})
	}

	var ratedMovies []api.RatedMovie
	for i := 0; i < 500; i++ {
		ratedMovies = append(ratedMovies, api.RatedMovie{
			Movie: api.Movie{
				Title: fmt.Sprintf("Rated Movie %d", i),
				Year:  2000 + (i % 23),
				IDs: api.MovieIDs{
					Trakt: i + 1000,
					TMDB:  (i + 1000) * 10,
				},
			},
			Rating: float64(i%10 + 1),
			RatedAt: time.Now().AddDate(0, 0, -i%365).Format(time.RFC3339),
		})
	}

	// Export options
	options := &export.ExportOptions{
		OutputDir:      cfg.Export.OutputDir,
		IncludeRatings: true,
		ConvertRatings: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputFile := filepath.Join(cfg.Export.OutputDir, fmt.Sprintf("test_export_%d.csv", i))
		err := export.GenerateLetterboxdCSV(watchedMovies, ratedMovies, outputFile, options, log)
		if err != nil {
			b.Fatalf("Failed to generate CSV: %v", err)
		}
	}
}

// BenchmarkMovieDataProcessing benchmarks the processing of raw API data into export format
func BenchmarkMovieDataProcessing(b *testing.B) {
	// Create sample raw API data
	var watchedMovies []api.WatchedMovie
	for i := 0; i < 1000; i++ {
		watchedMovies = append(watchedMovies, api.WatchedMovie{
			Movie: api.Movie{
				Title: fmt.Sprintf("Movie %d", i),
				Year:  2000 + (i % 23),
				IDs: api.MovieIDs{
					Trakt: i,
					TMDB:  i * 10,
				},
			},
			WatchedAt: time.Now().AddDate(0, 0, -i%365).Format(time.RFC3339),
		})
	}

	var ratedMovies []api.RatedMovie
	for i := 0; i < 500; i++ {
		ratedMovies = append(ratedMovies, api.RatedMovie{
			Movie: api.Movie{
				Title: fmt.Sprintf("Rated Movie %d", i),
				Year:  2000 + (i % 23),
				IDs: api.MovieIDs{
					Trakt: i + 1000,
					TMDB:  (i + 1000) * 10,
				},
			},
			Rating: float64(i%10 + 1),
			RatedAt: time.Now().AddDate(0, 0, -i%365).Format(time.RFC3339),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processed := export.ProcessMovieData(watchedMovies, ratedMovies, true)
		if len(processed) == 0 {
			b.Fatal("Failed to process movie data")
		}
	}
}

// Run all benchmarks with:
// go test -bench=. ./tests/benchmarks -v

// For memory profiling:
// go test -bench=. -benchmem ./tests/benchmarks -v 