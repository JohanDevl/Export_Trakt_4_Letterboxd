package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type MovieStats struct {
	TotalMovies          int
	WatchedMovies        int
	RatedMovies          int
	WatchlistedMovies    int
	UniqueMovies         map[string]bool
	YearDistribution     map[string]int
	GenreDistribution    map[string]int
	RatingDistribution   map[int]int
	WatchedByMonth       map[string]int
	TopDirectors         map[string]int
	TopActors            map[string]int
	AverageRating        float64
	LongestMovie         string
	LongestMovieRuntime  int
	ShortestMovie        string
	ShortestMovieRuntime int
	OldestMovie          string
	OldestMovieYear      int
	NewestMovie          string
	NewestMovieYear      int
}

func main() {
	watchedPath := flag.String("watched", "", "Path to watched movies CSV")
	ratingsPath := flag.String("ratings", "", "Path to ratings CSV")
	watchlistPath := flag.String("watchlist", "", "Path to watchlist CSV")
	outputPath := flag.String("output", "movie_stats.md", "Path to output markdown file")
	exportDir := flag.String("dir", "", "Directory containing export files (alternative to specifying individual files)")
	flag.Parse()

	// If export directory is provided, look for files there
	if *exportDir != "" {
		files, err := filepath.Glob(filepath.Join(*exportDir, "*.csv"))
		if err != nil {
			fmt.Printf("Error finding CSV files: %v\n", err)
			os.Exit(1)
		}

		for _, file := range files {
			filename := filepath.Base(file)
			if strings.HasPrefix(filename, "watched_") && *watchedPath == "" {
				*watchedPath = file
			} else if strings.HasPrefix(filename, "ratings_") && *ratingsPath == "" {
				*ratingsPath = file
			} else if strings.HasPrefix(filename, "watchlist_") && *watchlistPath == "" {
				*watchlistPath = file
			}
		}
	}

	// Check if we have at least one file to process
	if *watchedPath == "" && *ratingsPath == "" && *watchlistPath == "" {
		fmt.Println("No input files specified. Use -watched, -ratings, -watchlist flags or -dir to specify inputs.")
		flag.Usage()
		os.Exit(1)
	}

	stats := &MovieStats{
		UniqueMovies:         make(map[string]bool),
		YearDistribution:     make(map[string]int),
		GenreDistribution:    make(map[string]int),
		RatingDistribution:   make(map[int]int),
		WatchedByMonth:       make(map[string]int),
		TopDirectors:         make(map[string]int),
		TopActors:            make(map[string]int),
		OldestMovieYear:      9999,
		LongestMovieRuntime:  0,
		ShortestMovieRuntime: 999999,
	}

	// Process watched movies
	if *watchedPath != "" {
		fmt.Printf("Processing watched movies from %s...\n", *watchedPath)
		err := processWatchedMovies(*watchedPath, stats)
		if err != nil {
			fmt.Printf("Error processing watched movies: %v\n", err)
		}
	}

	// Process ratings
	if *ratingsPath != "" {
		fmt.Printf("Processing ratings from %s...\n", *ratingsPath)
		err := processRatings(*ratingsPath, stats)
		if err != nil {
			fmt.Printf("Error processing ratings: %v\n", err)
		}
	}

	// Process watchlist
	if *watchlistPath != "" {
		fmt.Printf("Processing watchlist from %s...\n", *watchlistPath)
		err := processWatchlist(*watchlistPath, stats)
		if err != nil {
			fmt.Printf("Error processing watchlist: %v\n", err)
		}
	}

	// Calculate stats
	calculateDerivedStats(stats)

	// Generate and save report
	report := generateReport(stats)
	err := os.WriteFile(*outputPath, []byte(report), 0644)
	if err != nil {
		fmt.Printf("Error writing report: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Stats report generated at %s\n", *outputPath)
}

func processWatchedMovies(filePath string, stats *MovieStats) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return err
	}

	// Find column indices
	nameIdx := findColumnIndex(header, "Name")
	yearIdx := findColumnIndex(header, "Year")
	watchedDateIdx := findColumnIndex(header, "WatchedDate")
	runtimeIdx := findColumnIndex(header, "Runtime")
	genresIdx := findColumnIndex(header, "Genres")
	directorsIdx := findColumnIndex(header, "Directors")
	castIdx := findColumnIndex(header, "Cast")

	if nameIdx == -1 || yearIdx == -1 {
		return fmt.Errorf("required columns not found in CSV")
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Warning: error reading record: %v\n", err)
			continue
		}

		if len(record) <= nameIdx || len(record) <= yearIdx {
			fmt.Println("Warning: record has fewer columns than expected")
			continue
		}

		name := record[nameIdx]
		year := record[yearIdx]

		// Count unique movies
		movieKey := name + "_" + year
		stats.UniqueMovies[movieKey] = true

		// Year distribution
		stats.YearDistribution[year]++

		// Process watched date for monthly distribution
		if watchedDateIdx != -1 && len(record) > watchedDateIdx {
			watchedDate := record[watchedDateIdx]
			if watchedDate != "" {
				date, err := time.Parse("2006-01-02", watchedDate)
				if err == nil {
					monthKey := date.Format("2006-01")
					stats.WatchedByMonth[monthKey]++
				}
			}
		}

		// Process runtime
		if runtimeIdx != -1 && len(record) > runtimeIdx {
			runtimeStr := record[runtimeIdx]
			if runtimeStr != "" {
				runtime, err := strconv.Atoi(runtimeStr)
				if err == nil {
					if runtime > stats.LongestMovieRuntime {
						stats.LongestMovieRuntime = runtime
						stats.LongestMovie = name
					}
					if runtime < stats.ShortestMovieRuntime {
						stats.ShortestMovieRuntime = runtime
						stats.ShortestMovie = name
					}
				}
			}
		}

		// Process year for oldest/newest
		if year != "" {
			yearInt, err := strconv.Atoi(year)
			if err == nil {
				if yearInt < stats.OldestMovieYear {
					stats.OldestMovieYear = yearInt
					stats.OldestMovie = name
				}
				if yearInt > stats.NewestMovieYear {
					stats.NewestMovieYear = yearInt
					stats.NewestMovie = name
				}
			}
		}

		// Process genres
		if genresIdx != -1 && len(record) > genresIdx {
			genres := record[genresIdx]
			if genres != "" {
				for _, genre := range strings.Split(genres, ", ") {
					stats.GenreDistribution[genre]++
				}
			}
		}

		// Process directors
		if directorsIdx != -1 && len(record) > directorsIdx {
			directors := record[directorsIdx]
			if directors != "" {
				for _, director := range strings.Split(directors, ", ") {
					stats.TopDirectors[director]++
				}
			}
		}

		// Process cast
		if castIdx != -1 && len(record) > castIdx {
			cast := record[castIdx]
			if cast != "" {
				actors := strings.Split(cast, ", ")
				// Limit to first 5 actors to avoid overweighting ensemble casts
				maxActors := min(5, len(actors))
				for i := 0; i < maxActors; i++ {
					stats.TopActors[actors[i]]++
				}
			}
		}

		stats.WatchedMovies++
	}

	return nil
}

func processRatings(filePath string, stats *MovieStats) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return err
	}

	// Find column indices
	nameIdx := findColumnIndex(header, "Name")
	yearIdx := findColumnIndex(header, "Year")
	ratingIdx := findColumnIndex(header, "Rating")

	if nameIdx == -1 || yearIdx == -1 || ratingIdx == -1 {
		return fmt.Errorf("required columns not found in CSV")
	}

	totalRating := 0
	ratedCount := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Warning: error reading record: %v\n", err)
			continue
		}

		if len(record) <= nameIdx || len(record) <= yearIdx || len(record) <= ratingIdx {
			fmt.Println("Warning: record has fewer columns than expected")
			continue
		}

		name := record[nameIdx]
		year := record[yearIdx]
		ratingStr := record[ratingIdx]

		// Count unique movies
		movieKey := name + "_" + year
		stats.UniqueMovies[movieKey] = true

		// Process rating
		if ratingStr != "" {
			rating, err := strconv.Atoi(ratingStr)
			if err == nil {
				stats.RatingDistribution[rating]++
				totalRating += rating
				ratedCount++
			}
		}

		stats.RatedMovies++
	}

	// Calculate average rating
	if ratedCount > 0 {
		stats.AverageRating = float64(totalRating) / float64(ratedCount)
	}

	return nil
}

func processWatchlist(filePath string, stats *MovieStats) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return err
	}

	// Find column indices
	nameIdx := findColumnIndex(header, "Name")
	yearIdx := findColumnIndex(header, "Year")

	if nameIdx == -1 || yearIdx == -1 {
		return fmt.Errorf("required columns not found in CSV")
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Warning: error reading record: %v\n", err)
			continue
		}

		if len(record) <= nameIdx || len(record) <= yearIdx {
			fmt.Println("Warning: record has fewer columns than expected")
			continue
		}

		name := record[nameIdx]
		year := record[yearIdx]

		// Count unique movies
		movieKey := name + "_" + year
		stats.UniqueMovies[movieKey] = true

		// Year distribution for watchlist
		stats.YearDistribution[year]++

		stats.WatchlistedMovies++
	}

	return nil
}

func calculateDerivedStats(stats *MovieStats) {
	stats.TotalMovies = len(stats.UniqueMovies)
}

func generateReport(stats *MovieStats) string {
	var sb strings.Builder

	sb.WriteString("# Movie Collection Statistics\n\n")
	sb.WriteString(fmt.Sprintf("*Generated on %s*\n\n", time.Now().Format("January 2, 2006")))

	// Overall numbers
	sb.WriteString("## Overview\n\n")
	sb.WriteString(fmt.Sprintf("Total unique movies: **%d**\n\n", stats.TotalMovies))
	sb.WriteString(fmt.Sprintf("- Watched movies: **%d**\n", stats.WatchedMovies))
	sb.WriteString(fmt.Sprintf("- Rated movies: **%d**\n", stats.RatedMovies))
	sb.WriteString(fmt.Sprintf("- Watchlisted movies: **%d**\n\n", stats.WatchlistedMovies))

	if stats.AverageRating > 0 {
		sb.WriteString(fmt.Sprintf("Average rating: **%.1f/10**\n\n", stats.AverageRating))
	}

	// Interesting facts
	sb.WriteString("## Interesting Facts\n\n")
	
	if stats.OldestMovie != "" && stats.OldestMovieYear < 9999 {
		sb.WriteString(fmt.Sprintf("- Oldest movie: **%s (%d)**\n", stats.OldestMovie, stats.OldestMovieYear))
	}
	
	if stats.NewestMovie != "" && stats.NewestMovieYear > 0 {
		sb.WriteString(fmt.Sprintf("- Newest movie: **%s (%d)**\n", stats.NewestMovie, stats.NewestMovieYear))
	}
	
	if stats.LongestMovie != "" && stats.LongestMovieRuntime > 0 {
		sb.WriteString(fmt.Sprintf("- Longest movie: **%s (%d minutes)**\n", stats.LongestMovie, stats.LongestMovieRuntime))
	}
	
	if stats.ShortestMovie != "" && stats.ShortestMovieRuntime < 999999 {
		sb.WriteString(fmt.Sprintf("- Shortest movie: **%s (%d minutes)**\n", stats.ShortestMovie, stats.ShortestMovieRuntime))
	}
	
	sb.WriteString("\n")

	// Rating distribution
	if len(stats.RatingDistribution) > 0 {
		sb.WriteString("## Rating Distribution\n\n")
		sb.WriteString("| Rating | Count | Percentage |\n")
		sb.WriteString("|--------|-------|------------|\n")
		
		ratings := make([]int, 0, len(stats.RatingDistribution))
		for rating := range stats.RatingDistribution {
			ratings = append(ratings, rating)
		}
		sort.Ints(ratings)
		
		for _, rating := range ratings {
			count := stats.RatingDistribution[rating]
			percentage := float64(count) / float64(stats.RatedMovies) * 100
			sb.WriteString(fmt.Sprintf("| %d/10 | %d | %.1f%% |\n", rating, count, percentage))
		}
		sb.WriteString("\n")
	}

	// Year distribution (top 10)
	if len(stats.YearDistribution) > 0 {
		sb.WriteString("## Movies by Decade\n\n")
		
		// Group by decade
		decades := make(map[string]int)
		for year, count := range stats.YearDistribution {
			if year == "" {
				continue
			}
			yearInt, err := strconv.Atoi(year)
			if err != nil {
				continue
			}
			decade := (yearInt / 10) * 10
			decadeStr := fmt.Sprintf("%ds", decade)
			decades[decadeStr] += count
		}
		
		// Convert to slice for sorting
		type decadeCount struct {
			decade string
			count  int
		}
		
		decadeCounts := make([]decadeCount, 0, len(decades))
		for decade, count := range decades {
			decadeCounts = append(decadeCounts, decadeCount{decade, count})
		}
		
		// Sort by decade
		sort.Slice(decadeCounts, func(i, j int) bool {
			return decadeCounts[i].decade < decadeCounts[j].decade
		})
		
		sb.WriteString("| Decade | Count | Percentage |\n")
		sb.WriteString("|--------|-------|------------|\n")
		
		for _, dc := range decadeCounts {
			percentage := float64(dc.count) / float64(stats.TotalMovies) * 100
			sb.WriteString(fmt.Sprintf("| %s | %d | %.1f%% |\n", dc.decade, dc.count, percentage))
		}
		sb.WriteString("\n")
	}

	// Genre distribution (top 10)
	if len(stats.GenreDistribution) > 0 {
		sb.WriteString("## Top Genres\n\n")
		
		// Convert to slice for sorting
		type genreCount struct {
			genre string
			count int
		}
		
		genreCounts := make([]genreCount, 0, len(stats.GenreDistribution))
		for genre, count := range stats.GenreDistribution {
			genreCounts = append(genreCounts, genreCount{genre, count})
		}
		
		// Sort by count descending
		sort.Slice(genreCounts, func(i, j int) bool {
			return genreCounts[i].count > genreCounts[j].count
		})
		
		sb.WriteString("| Genre | Count | Percentage |\n")
		sb.WriteString("|-------|-------|------------|\n")
		
		// Display top 10 genres
		topN := min(10, len(genreCounts))
		for i := 0; i < topN; i++ {
			percentage := float64(genreCounts[i].count) / float64(stats.WatchedMovies) * 100
			sb.WriteString(fmt.Sprintf("| %s | %d | %.1f%% |\n", genreCounts[i].genre, genreCounts[i].count, percentage))
		}
		sb.WriteString("\n")
	}

	// Top directors (top 10)
	if len(stats.TopDirectors) > 0 {
		sb.WriteString("## Top Directors\n\n")
		
		// Convert to slice for sorting
		type directorCount struct {
			director string
			count    int
		}
		
		directorCounts := make([]directorCount, 0, len(stats.TopDirectors))
		for director, count := range stats.TopDirectors {
			directorCounts = append(directorCounts, directorCount{director, count})
		}
		
		// Sort by count descending
		sort.Slice(directorCounts, func(i, j int) bool {
			return directorCounts[i].count > directorCounts[j].count
		})
		
		sb.WriteString("| Director | Movies |\n")
		sb.WriteString("|----------|--------|\n")
		
		// Display top 10 directors
		topN := min(10, len(directorCounts))
		for i := 0; i < topN; i++ {
			sb.WriteString(fmt.Sprintf("| %s | %d |\n", directorCounts[i].director, directorCounts[i].count))
		}
		sb.WriteString("\n")
	}

	// Top actors (top 10)
	if len(stats.TopActors) > 0 {
		sb.WriteString("## Top Actors\n\n")
		
		// Convert to slice for sorting
		type actorCount struct {
			actor string
			count int
		}
		
		actorCounts := make([]actorCount, 0, len(stats.TopActors))
		for actor, count := range stats.TopActors {
			actorCounts = append(actorCounts, actorCount{actor, count})
		}
		
		// Sort by count descending
		sort.Slice(actorCounts, func(i, j int) bool {
			return actorCounts[i].count > actorCounts[j].count
		})
		
		sb.WriteString("| Actor | Movies |\n")
		sb.WriteString("|-------|--------|\n")
		
		// Display top 10 actors
		topN := min(10, len(actorCounts))
		for i := 0; i < topN; i++ {
			sb.WriteString(fmt.Sprintf("| %s | %d |\n", actorCounts[i].actor, actorCounts[i].count))
		}
		sb.WriteString("\n")
	}

	// Monthly watching activity
	if len(stats.WatchedByMonth) > 0 {
		sb.WriteString("## Watching Activity by Month\n\n")
		
		// Convert to slice for sorting
		type monthCount struct {
			month string
			count int
		}
		
		monthCounts := make([]monthCount, 0, len(stats.WatchedByMonth))
		for month, count := range stats.WatchedByMonth {
			monthCounts = append(monthCounts, monthCount{month, count})
		}
		
		// Sort by month
		sort.Slice(monthCounts, func(i, j int) bool {
			return monthCounts[i].month > monthCounts[j].month
		})
		
		sb.WriteString("| Month | Movies Watched |\n")
		sb.WriteString("|-------|----------------|\n")
		
		// Display most recent 12 months
		topN := min(12, len(monthCounts))
		for i := 0; i < topN; i++ {
			// Format month for display (2006-01 -> January 2006)
			t, err := time.Parse("2006-01", monthCounts[i].month)
			if err == nil {
				formattedMonth := t.Format("January 2006")
				sb.WriteString(fmt.Sprintf("| %s | %d |\n", formattedMonth, monthCounts[i].count))
			} else {
				sb.WriteString(fmt.Sprintf("| %s | %d |\n", monthCounts[i].month, monthCounts[i].count))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func findColumnIndex(header []string, columnName string) int {
	for i, name := range header {
		if name == columnName {
			return i
		}
	}
	return -1
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
} 