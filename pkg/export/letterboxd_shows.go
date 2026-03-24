package export

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/api"
)

// ExportShows exports the user's watched shows to a CSV file
func (e *LetterboxdExporter) ExportShows(shows []api.WatchedShow) error {
	// Get export directory
	exportDir, err := e.getExportDir()
	if err != nil {
		return err
	}

	// Use configured filename, or generate one with timestamp if not specified
	var filename string
	if e.config.Letterboxd.ShowsFilename != "" {
		filename = e.config.Letterboxd.ShowsFilename
	} else {
		// Use the configured timezone for filename timestamp
		now := e.getTimeInConfigTimezone()
		filename = fmt.Sprintf("shows-export_%s_%s.csv",
			now.Format(DateFormat),
			now.Format(TimeFormat))
	}
	filePath := filepath.Join(exportDir, filename)

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
	header := []string{"Title", "Year", "Season", "Episode", "EpisodeTitle", "LastWatched", "Rating10", "IMDb ID"}
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

	// Fetch episode ratings
	episodeRatings, err := e.fetchEpisodeRatings()
	if err != nil {
		e.log.Warn("export.episode_ratings_fetch_failed", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Create a map of episode ratings for quick lookup
	// Use a composite key of show_id:season:episode
	episodeRatingMap := make(map[string]int)
	for _, r := range episodeRatings {
		if r.Show.IDs.Trakt > 0 && r.Episode.Season > 0 && r.Episode.Number > 0 {
			key := fmt.Sprintf("%d:%d:%d", r.Show.IDs.Trakt, r.Episode.Season, r.Episode.Number)
			episodeRatingMap[key] = int(r.Rating)
		}
	}

	// Fetch show ratings too
	showRatings, err := e.fetchShowRatings()
	if err != nil {
		e.log.Warn("export.show_ratings_fetch_failed", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Create a map of show ratings for quick lookup
	showRatingMap := make(map[int]int)
	for _, r := range showRatings {
		if r.Show.IDs.Trakt > 0 {
			showRatingMap[r.Show.IDs.Trakt] = int(r.Rating)
		}
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

				// Get rating for this episode
				rating := ""
				key := fmt.Sprintf("%d:%d:%d", show.Show.IDs.Trakt, season.Number, episode.Number)
				if r, exists := episodeRatingMap[key]; exists {
					rating = strconv.Itoa(r)
				} else if r, exists := showRatingMap[show.Show.IDs.Trakt]; exists {
					// If no episode rating, use show rating
					rating = strconv.Itoa(r)
				}

				record := []string{
					show.Show.Title,
					strconv.Itoa(show.Show.Year),
					strconv.Itoa(season.Number),
					strconv.Itoa(episode.Number),
					episode.Title,
					watchedDate,
					rating,
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
