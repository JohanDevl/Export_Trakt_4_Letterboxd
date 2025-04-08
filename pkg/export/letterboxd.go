package export

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/api"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

// LetterboxdExporter handles the export of movies to Letterboxd format
type LetterboxdExporter struct {
	config *config.Config
	log    logger.Logger
}

// NewLetterboxdExporter creates a new Letterboxd exporter
func NewLetterboxdExporter(cfg *config.Config, log logger.Logger) *LetterboxdExporter {
	return &LetterboxdExporter{
		config: cfg,
		log:    log,
	}
}

// ExportMovies exports the given movies to a CSV file in Letterboxd format
func (e *LetterboxdExporter) ExportMovies(movies []api.Movie) error {
	if err := os.MkdirAll(e.config.Letterboxd.ExportDir, 0755); err != nil {
		e.log.Error("errors.export_dir_create_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create export directory: %w", err)
	}

	// Use configured filename, or generate one with timestamp if not specified
	var filename string
	if e.config.Letterboxd.WatchedFilename != "" {
		filename = e.config.Letterboxd.WatchedFilename
	} else {
		filename = fmt.Sprintf("letterboxd-export-%s.csv", time.Now().Format("2006-01-02"))
	}
	filePath := filepath.Join(e.config.Letterboxd.ExportDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		e.log.Error("errors.file_create_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create export file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"Title", "Year", "WatchedDate", "Rating10", "IMDb ID"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Get ratings for movies
	ratings, err := e.fetchRatings()
	if err != nil {
		e.log.Warn("export.ratings_fetch_failed", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Create a map of movie ratings for quick lookup
	movieRatings := make(map[string]string)
	for _, rating := range ratings {
		// Use IMDB ID as key for the ratings map
		if rating.Movie.IDs.IMDB != "" {
			// Convert to integer value (1-10)
			movieRatings[rating.Movie.IDs.IMDB] = strconv.Itoa(int(rating.Rating))
		}
	}

	// Write movies
	for _, movie := range movies {
		// Parse watched date
		watchedDate := ""
		if movie.LastWatchedAt != "" {
			if parsedTime, err := time.Parse(time.RFC3339, movie.LastWatchedAt); err == nil {
				watchedDate = parsedTime.Format(e.config.Export.DateFormat)
			}
		}

		// Get rating for this movie
		rating := ""
		if r, exists := movieRatings[movie.Movie.IDs.IMDB]; exists {
			rating = r
		}

		record := []string{
			movie.Movie.Title,
			strconv.Itoa(movie.Movie.Year),
			watchedDate,
			rating,
			movie.Movie.IDs.IMDB,
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write movie record: %w", err)
		}
	}

	e.log.Info("export.export_complete", map[string]interface{}{
		"count": len(movies),
		"path":  filePath,
	})
	return nil
}

// ExportCollectionMovies exports the user's movie collection to a CSV file in Letterboxd format
func (e *LetterboxdExporter) ExportCollectionMovies(movies []api.CollectionMovie) error {
	if err := os.MkdirAll(e.config.Letterboxd.ExportDir, 0755); err != nil {
		e.log.Error("errors.export_dir_create_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create export directory: %w", err)
	}

	// Use configured filename, or generate one with timestamp if not specified
	var filename string
	if e.config.Letterboxd.CollectionFilename != "" {
		filename = e.config.Letterboxd.CollectionFilename
	} else {
		filename = fmt.Sprintf("collection-export-%s.csv", time.Now().Format("2006-01-02"))
	}
	filePath := filepath.Join(e.config.Letterboxd.ExportDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		e.log.Error("errors.file_create_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create export file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"Title", "Year", "CollectedDate", "IMDb ID"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write movies
	for _, movie := range movies {
		// Parse collected date
		collectedDate := ""
		if movie.CollectedAt != "" {
			if parsedTime, err := time.Parse(time.RFC3339, movie.CollectedAt); err == nil {
				collectedDate = parsedTime.Format(e.config.Export.DateFormat)
			}
		}

		record := []string{
			movie.Movie.Title,
			strconv.Itoa(movie.Movie.Year),
			collectedDate,
			movie.Movie.IDs.IMDB,
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write movie record: %w", err)
		}
	}

	e.log.Info("export.collection_export_complete", map[string]interface{}{
		"count": len(movies),
		"path":  filePath,
	})
	return nil
}

// ExportShows exports the user's watched shows to a CSV file
func (e *LetterboxdExporter) ExportShows(shows []api.WatchedShow) error {
	if err := os.MkdirAll(e.config.Letterboxd.ExportDir, 0755); err != nil {
		e.log.Error("errors.export_dir_create_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create export directory: %w", err)
	}

	// Use configured filename, or generate one with timestamp if not specified
	var filename string
	if e.config.Letterboxd.ShowsFilename != "" {
		filename = e.config.Letterboxd.ShowsFilename
	} else {
		filename = fmt.Sprintf("shows-export-%s.csv", time.Now().Format("2006-01-02"))
	}
	filePath := filepath.Join(e.config.Letterboxd.ExportDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		e.log.Error("errors.file_create_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create export file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"Title", "Year", "Season", "Episode", "EpisodeTitle", "LastWatched", "IMDb ID"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Check if episode titles are available
	missingTitles := true
	checkLimit := 0
	outerLoop:
	for _, show := range shows {
		for _, season := range show.Seasons {
			for _, episode := range season.Episodes {
				if episode.Title != "" {
					missingTitles = false
					break outerLoop
				}
				// Check only a reasonable number of episodes
				checkLimit++
				if checkLimit > 20 {
					break outerLoop
				}
			}
		}
	}

	if missingTitles {
		e.log.Warn("export.episode_titles_missing", map[string]interface{}{
			"message": "Episode titles are missing. Check your Trakt API extended_info setting.",
		})
	}

	// Write episodes
	episodeCount := 0
	for _, show := range shows {
		for _, season := range show.Seasons {
			for _, episode := range season.Episodes {
				// Parse watched date
				watchedDate := ""
				if show.LastWatchedAt != "" {
					if parsedTime, err := time.Parse(time.RFC3339, show.LastWatchedAt); err == nil {
						watchedDate = parsedTime.Format(e.config.Export.DateFormat)
					}
				}

				record := []string{
					show.Show.Title,
					strconv.Itoa(show.Show.Year),
					strconv.Itoa(season.Number),
					strconv.Itoa(episode.Number),
					episode.Title,
					watchedDate,
					show.Show.IDs.IMDB,
				}

				if err := writer.Write(record); err != nil {
					return fmt.Errorf("failed to write episode record: %w", err)
				}
				episodeCount++
			}
		}
	}

	e.log.Info("export.shows_export_complete", map[string]interface{}{
		"shows":    len(shows),
		"episodes": episodeCount,
		"path":     filePath,
	})
	return nil
}

// ExportRatings exports the user's movie ratings to a CSV file in Letterboxd format
func (e *LetterboxdExporter) ExportRatings(ratings []api.Rating) error {
	if err := os.MkdirAll(e.config.Letterboxd.ExportDir, 0755); err != nil {
		e.log.Error("errors.export_dir_create_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create export directory: %w", err)
	}

	// Use configured filename, or generate one with timestamp if not specified
	filename := fmt.Sprintf("ratings-export-%s.csv", time.Now().Format("2006-01-02"))
	filePath := filepath.Join(e.config.Letterboxd.ExportDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		e.log.Error("errors.file_create_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create export file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header - Letterboxd format for ratings
	header := []string{"Title", "Year", "Rating10", "RatedDate", "IMDb ID"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write ratings
	for _, r := range ratings {
		// Parse rated date
		ratedDate := ""
		if r.RatedAt != "" {
			if parsedTime, err := time.Parse(time.RFC3339, r.RatedAt); err == nil {
				ratedDate = parsedTime.Format(e.config.Export.DateFormat)
			}
		}

		// Use integer rating directly (1-10)
		ratingStr := ""
		if r.Rating > 0 {
			ratingStr = strconv.Itoa(int(r.Rating))
		}

		record := []string{
			r.Movie.Title,
			strconv.Itoa(r.Movie.Year),
			ratingStr,
			ratedDate,
			r.Movie.IDs.IMDB,
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write rating record: %w", err)
		}
	}

	e.log.Info("export.ratings_export_complete", map[string]interface{}{
		"count": len(ratings),
		"path":  filePath,
	})
	return nil
}

// ExportWatchlist exports the user's movie watchlist to a CSV file in Letterboxd format
func (e *LetterboxdExporter) ExportWatchlist(watchlist []api.WatchlistMovie) error {
	if err := os.MkdirAll(e.config.Letterboxd.ExportDir, 0755); err != nil {
		e.log.Error("errors.export_dir_create_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create export directory: %w", err)
	}

	// Use configured filename, or generate one with timestamp if not specified
	filename := fmt.Sprintf("watchlist-export-%s.csv", time.Now().Format("2006-01-02"))
	filePath := filepath.Join(e.config.Letterboxd.ExportDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		e.log.Error("errors.file_create_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create export file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header - Letterboxd format for watchlist
	header := []string{"Title", "Year", "ListedDate", "Rating10", "IMDb ID"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write watchlist entries
	for _, wl := range watchlist {
		// Parse listed date
		listedDate := ""
		if wl.ListedAt != "" {
			if parsedTime, err := time.Parse(time.RFC3339, wl.ListedAt); err == nil {
				listedDate = parsedTime.Format(e.config.Export.DateFormat)
			}
		}

		record := []string{
			wl.Movie.Title,
			strconv.Itoa(wl.Movie.Year),
			listedDate,
			wl.Notes,
			wl.Movie.IDs.IMDB,
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write watchlist record: %w", err)
		}
	}

	e.log.Info("export.watchlist_export_complete", map[string]interface{}{
		"count": len(watchlist),
		"path":  filePath,
	})
	return nil
}

// ExportLetterboxdFormat exports the given movies to a CSV file in Letterboxd import format
// The format matches the official Letterboxd import format with columns:
// Title, Year, imdbID, tmdbID, WatchedDate, Rating10, Rewatch
func (e *LetterboxdExporter) ExportLetterboxdFormat(movies []api.Movie, ratings []api.Rating) error {
	if err := os.MkdirAll(e.config.Letterboxd.ExportDir, 0755); err != nil {
		e.log.Error("errors.export_dir_create_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create export directory: %w", err)
	}

	// Use configured filename, or generate one with timestamp if not specified
	filename := "letterboxd_import.csv"
	filePath := filepath.Join(e.config.Letterboxd.ExportDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		e.log.Error("errors.file_create_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create export file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"Title", "Year", "imdbID", "tmdbID", "WatchedDate", "Rating10", "Rewatch"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Create a map of movie ratings for quick lookup
	movieRatings := make(map[string]float64)
	for _, rating := range ratings {
		// Use IMDB ID as key for the ratings map
		if rating.Movie.IDs.IMDB != "" {
			movieRatings[rating.Movie.IDs.IMDB] = rating.Rating
		}
	}

	// Create a map to track plays for determining rewatches
	moviePlays := make(map[string]int)
	for _, movie := range movies {
		if movie.Movie.IDs.IMDB != "" {
			moviePlays[movie.Movie.IDs.IMDB] += movie.Plays
		}
	}

	// Write movies
	for _, movie := range movies {
		// Parse watched date
		watchedDate := ""
		if movie.LastWatchedAt != "" {
			if parsedTime, err := time.Parse(time.RFC3339, movie.LastWatchedAt); err == nil {
				watchedDate = parsedTime.Format(e.config.Export.DateFormat)
			}
		}

		// Get rating (scale is already 1-10 in Trakt)
		rating := ""
		if r, exists := movieRatings[movie.Movie.IDs.IMDB]; exists {
			rating = strconv.FormatFloat(r, 'f', 0, 64)
		}

		// Determine if this is a rewatch
		rewatch := "false"
		if movie.Plays > 1 {
			rewatch = "true"
		}

		// Convert TMDB ID to string
		tmdbID := strconv.Itoa(movie.Movie.IDs.TMDB)

		record := []string{
			movie.Movie.Title,
			strconv.Itoa(movie.Movie.Year),
			movie.Movie.IDs.IMDB,
			tmdbID,
			watchedDate,
			rating,
			rewatch,
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write movie record: %w", err)
		}
	}

	e.log.Info("export.letterboxd_export_complete", map[string]interface{}{
		"count": len(movies),
		"path":  filePath,
	})
	return nil
}

// fetchRatings is a helper function to get movie ratings
func (e *LetterboxdExporter) fetchRatings() ([]api.Rating, error) {
	// Create a new Trakt client with the same config
	client := api.NewClient(e.config, e.log)
	return client.GetRatings()
} 