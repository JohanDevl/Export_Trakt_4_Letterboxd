package export

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/api"
)

// ExportMovies exports the given movies to a CSV file in Letterboxd format
func (e *LetterboxdExporter) ExportMovies(movies []api.Movie, client *api.Client) error {
	// Get export directory
	exportDir, err := e.getExportDir()
	if err != nil {
		return err
	}

	// Check if we're in a test environment
	isTestEnv := containsAny(exportDir, []string{"test", "tmp", "temp"})

	// Use configured filename, or generate one with timestamp if not specified
	var filename string
	if e.config.Letterboxd.WatchedFilename != "" {
		filename = e.config.Letterboxd.WatchedFilename
	} else if isTestEnv {
		// Use a fixed filename for tests to make it easier to locate
		filename = "watched-export-test.csv"
	} else {
		// Use the configured timezone for filename timestamp
		now := e.getTimeInConfigTimezone()
		filename = fmt.Sprintf("letterboxd-export_%s_%s.csv",
			now.Format(DateFormat),
			now.Format(TimeFormat))
	}
	filePath := filepath.Join(exportDir, filename)

	// Create export file
	file, err := os.Create(filePath)
	if err != nil {
		e.log.Error("errors.file_create_failed", map[string]interface{}{
			"error": err.Error(),
			"path":  filePath,
		})
		return fmt.Errorf("failed to create export file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"Title", "Year", "WatchedDate", "Rating10", "imdbID", "tmdbID", "Rewatch"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Get ratings for movies
	var ratings []api.Rating
	if client != nil {
		ratingsData, err := client.GetRatings()
		if err != nil {
			e.log.Warn("export.ratings_fetch_failed", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			ratings = ratingsData
		}
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

	// Sort movies by watched date (most recent first)
	sortedMovies := make([]api.Movie, len(movies))
	copy(sortedMovies, movies)

	// Sort the movies slice by LastWatchedAt (newest to oldest)
	sort.Slice(sortedMovies, func(i, j int) bool {
		timeI, errI := time.Parse(time.RFC3339, sortedMovies[i].LastWatchedAt)
		timeJ, errJ := time.Parse(time.RFC3339, sortedMovies[j].LastWatchedAt)

		// Handle parsing errors or empty dates
		if errI != nil && errJ != nil {
			return false // Both invalid, order doesn't matter
		}
		if errI != nil {
			return false // i has invalid date, put at end
		}
		if errJ != nil {
			return true // j has invalid date, i comes first
		}

		// Return true if timeI is after timeJ (reverse chronological order)
		return timeI.After(timeJ)
	})

	// Write movies
	for _, movie := range sortedMovies {
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
			watchedDate,
			rating,
			movie.Movie.IDs.IMDB,
			tmdbID,
			rewatch,
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

// ExportMovieHistory exports the user's complete movie watch history to a CSV file with individual watch events
func (e *LetterboxdExporter) ExportMovieHistory(history []api.HistoryItem, apiClient *api.Client) error {
	// Get export directory
	exportDir, err := e.getExportDir()
	if err != nil {
		return err
	}

	// Check if we're in a test environment
	isTestEnv := containsAny(exportDir, []string{"test", "tmp", "temp"})

	// Use configured filename, or generate one with timestamp if not specified
	var filename string
	if e.config.Letterboxd.WatchedFilename != "" {
		filename = e.config.Letterboxd.WatchedFilename
	} else if isTestEnv {
		filename = "watched-history-test.csv"
	} else {
		now := e.getTimeInConfigTimezone()
		filename = fmt.Sprintf("watched-history_%s_%s.csv",
			now.Format(DateFormat),
			now.Format(TimeFormat))
	}
	filePath := filepath.Join(exportDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		e.log.Error("errors.file_create_failed", map[string]interface{}{
			"error": err.Error(),
			"path":  filePath,
		})
		return fmt.Errorf("failed to create export file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"Title", "Year", "WatchedDate", "Rating10", "imdbID", "tmdbID", "Rewatch"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Get ratings if available
	movieRatings := make(map[string]string)
	if apiClient != nil {
		if ratings, err := apiClient.GetRatings(); err == nil {
			for _, rating := range ratings {
				if rating.Movie.IDs.IMDB != "" {
					movieRatings[rating.Movie.IDs.IMDB] = strconv.Itoa(int(rating.Rating))
				}
			}
		}
	}

	// Sort history by watched date (most recent first)
	sortedHistory := make([]api.HistoryItem, len(history))
	copy(sortedHistory, history)

	sort.Slice(sortedHistory, func(i, j int) bool {
		timeI, errI := time.Parse(time.RFC3339, sortedHistory[i].WatchedAt)
		timeJ, errJ := time.Parse(time.RFC3339, sortedHistory[j].WatchedAt)

		if errI != nil && errJ != nil {
			return false
		}
		if errI != nil {
			return false
		}
		if errJ != nil {
			return true
		}

		return timeI.After(timeJ)
	})

	// Track first occurrence of each movie to determine rewatch status
	movieFirstWatch := make(map[string]bool)

	// Process in reverse order to identify first watches
	for i := len(sortedHistory) - 1; i >= 0; i-- {
		item := sortedHistory[i]
		if item.Movie.IDs.IMDB != "" {
			if !movieFirstWatch[item.Movie.IDs.IMDB] {
				movieFirstWatch[item.Movie.IDs.IMDB] = true
			}
		}
	}

	// Track which movies we've seen to determine rewatch
	// Process from oldest to newest to correctly identify rewatches
	seenMovies := make(map[string]bool)
	rewatchMap := make(map[int]bool) // Map index to rewatch status

	// First pass: Process in reverse order (oldest first) to determine rewatch status
	for i := len(sortedHistory) - 1; i >= 0; i-- {
		item := sortedHistory[i]
		if seenMovies[item.Movie.IDs.IMDB] {
			rewatchMap[i] = true // This is a rewatch
		} else {
			rewatchMap[i] = false // First time watching this movie
			seenMovies[item.Movie.IDs.IMDB] = true
		}
	}

	// Write history entries (in newest to oldest order)
	for i, item := range sortedHistory {
		// Parse watched date
		watchedDate := ""
		if item.WatchedAt != "" {
			if parsedTime, err := time.Parse(time.RFC3339, item.WatchedAt); err == nil {
				watchedDate = parsedTime.Format(e.config.Export.DateFormat)
			}
		}

		// Get rating for this movie
		rating := ""
		if r, exists := movieRatings[item.Movie.IDs.IMDB]; exists {
			rating = r
		}

		// Determine if this is a rewatch using pre-calculated map
		rewatch := "false"
		if rewatchMap[i] {
			rewatch = "true"
		}

		// Convert TMDB ID to string
		tmdbID := strconv.Itoa(item.Movie.IDs.TMDB)

		record := []string{
			item.Movie.Title,
			strconv.Itoa(item.Movie.Year),
			watchedDate,
			rating,
			item.Movie.IDs.IMDB,
			tmdbID,
			rewatch,
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write history record: %w", err)
		}
	}

	e.log.Info("export.history_export_complete", map[string]interface{}{
		"count": len(history),
		"path":  filePath,
	})
	return nil
}

// ExportCollectionMovies exports the user's movie collection to a CSV file in Letterboxd format
func (e *LetterboxdExporter) ExportCollectionMovies(movies []api.CollectionMovie) error {
	// Get export directory
	exportDir, err := e.getExportDir()
	if err != nil {
		return err
	}

	// Check if we're in a test environment
	isTestEnv := containsAny(exportDir, []string{"test", "tmp", "temp"})

	// Use configured filename, or generate one with timestamp if not specified
	var filename string
	if e.config.Letterboxd.CollectionFilename != "" {
		filename = e.config.Letterboxd.CollectionFilename
	} else if isTestEnv {
		// Use a fixed filename for tests to make it easier to locate
		filename = "collection-export-test.csv"
	} else {
		// Use the configured timezone for filename timestamp
		now := e.getTimeInConfigTimezone()
		filename = fmt.Sprintf("collection-export_%s_%s.csv",
			now.Format(DateFormat),
			now.Format(TimeFormat))
	}
	filePath := filepath.Join(exportDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		e.log.Error("errors.file_create_failed", map[string]interface{}{
			"error": err.Error(),
			"path":  filePath,
		})
		return fmt.Errorf("failed to create export file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"Title", "Year", "CollectedDate", "imdbID", "tmdbID"}
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

		// Convert TMDB ID to string
		tmdbID := strconv.Itoa(movie.Movie.IDs.TMDB)

		record := []string{
			movie.Movie.Title,
			strconv.Itoa(movie.Movie.Year),
			collectedDate,
			movie.Movie.IDs.IMDB,
			tmdbID,
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

// ExportLetterboxdFormat exports the given movies to a CSV file in Letterboxd import format
// The format matches the official Letterboxd import format with columns:
// Title, Year, imdbID, tmdbID, WatchedDate, Rating10, Rewatch
func (e *LetterboxdExporter) ExportLetterboxdFormat(movies []api.Movie, ratings []api.Rating) error {
	// Get export directory
	exportDir, err := e.getExportDir()
	if err != nil {
		return err
	}

	// Check if we're in a test environment
	isTestEnv := containsAny(exportDir, []string{"test", "tmp", "temp"})

	// Use configured filename, or standard name
	var filename string
	if e.config.Letterboxd.LetterboxdImportFilename != "" {
		filename = e.config.Letterboxd.LetterboxdImportFilename
	} else if isTestEnv {
		// Use a fixed filename for tests to make it easier to locate
		filename = "letterboxd-import-test.csv"
	} else {
		filename = "letterboxd_import.csv"
	}
	filePath := filepath.Join(exportDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		e.log.Error("errors.file_create_failed", map[string]interface{}{
			"error": err.Error(),
			"path":  filePath,
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
