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

	filename := fmt.Sprintf("letterboxd-export-%s.csv", time.Now().Format("2006-01-02"))
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
	header := []string{"Title", "Year", "WatchedDate", "Rating", "IMDb ID"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
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

		record := []string{
			movie.Movie.Title,
			strconv.Itoa(movie.Movie.Year),
			watchedDate,
			"", // Rating not available in the current API response
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