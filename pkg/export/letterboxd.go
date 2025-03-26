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
	log    *logger.Logger
}

// NewLetterboxdExporter creates a new Letterboxd exporter
func NewLetterboxdExporter(cfg *config.Config, log *logger.Logger) *LetterboxdExporter {
	return &LetterboxdExporter{
		config: cfg,
		log:    log,
	}
}

// ExportMovies exports the given movies to a CSV file in Letterboxd format
func (e *LetterboxdExporter) ExportMovies(movies []api.Movie) error {
	if err := os.MkdirAll(e.config.Letterboxd.ExportDir, 0755); err != nil {
		return fmt.Errorf("failed to create export directory: %w", err)
	}

	filename := fmt.Sprintf("letterboxd-export-%s.csv", time.Now().Format("2006-01-02"))
	filepath := filepath.Join(e.config.Letterboxd.ExportDir, filename)

	file, err := os.Create(filepath)
	if err != nil {
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
		rating := ""
		if movie.Rating > 0 {
			rating = strconv.Itoa(movie.Rating)
		}

		watchedDate := ""
		if !movie.WatchedAt.IsZero() {
			watchedDate = movie.WatchedAt.Format(e.config.Export.DateFormat)
		}

		record := []string{
			movie.Title,
			strconv.Itoa(movie.Year),
			watchedDate,
			rating,
			movie.IDs.IMDB,
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write movie record: %w", err)
		}
	}

	e.log.Infof("Successfully exported %d movies to %s", len(movies), filepath)
	return nil
} 