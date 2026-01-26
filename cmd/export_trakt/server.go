package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/auth"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/web"
)

// Helper structures for export scanning
type ExportItem struct {
	ID          string
	Type        string
	Date        time.Time
	Status      string
	Duration    string
	FileSize    string
	RecordCount int
	Files       []string
	Error       string
}

// startPersistentServer starts a persistent HTTP server that handles OAuth callbacks and export requests
func startPersistentServer(cfg *config.Config, log logger.Logger, tokenManager *auth.TokenManager, scheduleFlag, exportType, exportMode string) error {
	// Use the real web package with pagination support
	webServer, err := web.NewServer(cfg, log, tokenManager)
	if err != nil {
		return fmt.Errorf("failed to create web server: %w", err)
	}

	port := cfg.Auth.CallbackPort
	if port == 0 {
		port = 8080
	}

	// Start scheduler if schedule flag is provided
	if scheduleFlag != "" {
		log.Info("server.starting_scheduler", map[string]interface{}{
			"schedule":    scheduleFlag,
			"export_type": exportType,
			"export_mode": exportMode,
		})

		go func() {
			runWithSchedule(cfg, log, scheduleFlag, exportType, exportMode)
		}()

		fmt.Println("🕒 Automatic Export Scheduler Started")
		fmt.Printf("📅 Schedule: %s\n", scheduleFlag)
		fmt.Printf("📦 Export Type: %s\n", exportType)
		fmt.Printf("🔧 Export Mode: %s\n", exportMode)
		fmt.Println()
	}

	fmt.Println("🚀 Starting Enhanced Web Interface Server with Pagination")
	fmt.Println("=========================================================")
	fmt.Printf("📱 Client ID: %s\n", cfg.Trakt.ClientID)
	fmt.Printf("🔗 Redirect URI: %s\n", cfg.Auth.RedirectURI)
	fmt.Printf("🌐 Server running on: http://0.0.0.0:%d\n", port)
	fmt.Printf("📊 Dashboard: http://0.0.0.0:%d/\n", port)
	fmt.Printf("📁 Exports: http://0.0.0.0:%d/exports\n", port)
	fmt.Printf("🔍 Status: http://0.0.0.0:%d/status\n", port)
	fmt.Println("📄 Features: Server-side pagination, lazy loading, configurable page sizes")
	fmt.Println()

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Info("server.shutdown_signal_received", map[string]interface{}{
			"signal": sig.String(),
		})
		fmt.Printf("\nReceived signal %s, shutting down server...\n", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := webServer.Stop(ctx); err != nil {
			log.Error("server.shutdown_error", map[string]interface{}{
				"error": err.Error(),
			})
		}

		log.Info("server.shutdown_complete", nil)
		os.Exit(0)
	}()

	fmt.Printf("\n✅ Enhanced Web Interface with Pagination started! Press Ctrl+C to stop.\n")
	fmt.Printf("🌐 Access your dashboard at: http://localhost:%d\n", port)
	fmt.Printf("📁 Exports page: http://localhost:%d/exports\n", port)
	fmt.Println()

	// Start the web server with pagination support
	return webServer.Start()
}

// scanExportFiles scans the exports directory for existing export files
func scanExportFiles(exportsDir string, log logger.Logger) []ExportItem {
	var exports []ExportItem

	// Check if exports directory exists
	if _, err := os.Stat(exportsDir); os.IsNotExist(err) {
		log.Info("web.exports_dir_not_found", map[string]interface{}{
			"dir": exportsDir,
		})
		return exports
	}

	// Scan for timestamped export directories and individual files
	entries, err := os.ReadDir(exportsDir)
	if err != nil {
		log.Error("web.scan_exports_dir_error", map[string]interface{}{
			"error": err.Error(),
		})
		return exports
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			// Handle individual CSV files in root exports directory
			if strings.HasSuffix(strings.ToLower(entry.Name()), ".csv") {
				export := processCSVFile(filepath.Join(exportsDir, entry.Name()), entry.Name())
				if export != nil {
					exports = append(exports, *export)
				}
			}
			continue
		}

		// Check if directory name matches export timestamp pattern
		dirName := entry.Name()
		if strings.HasPrefix(dirName, "export_") && len(dirName) >= 16 {
			export := processExportDirectory(filepath.Join(exportsDir, dirName), dirName)
			if export != nil {
				exports = append(exports, *export)
			}
		}
	}

	// Sort by date (newest first)
	for i := 0; i < len(exports)-1; i++ {
		for j := i + 1; j < len(exports); j++ {
			if exports[i].Date.Before(exports[j].Date) {
				exports[i], exports[j] = exports[j], exports[i]
			}
		}
	}

	log.Info("web.exports_scanned", map[string]interface{}{
		"count": len(exports),
	})

	return exports
}

func processExportDirectory(dirPath, dirName string) *ExportItem {
	// Parse timestamp from directory name (export_2025-07-11_15-43)
	parts := strings.Split(dirName, "_")
	if len(parts) < 3 {
		return nil
	}

	dateStr := parts[1] + " " + strings.ReplaceAll(parts[2], "-", ":")
	exportDate, err := time.Parse("2006-01-02 15:04", dateStr)
	if err != nil {
		// Fallback to directory modification time
		info, err := os.Stat(dirPath)
		if err != nil {
			return nil
		}
		exportDate = info.ModTime()
	}

	// Scan files in the export directory
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil
	}

	var csvFiles []string
	var totalSize int64
	var totalRecords int
	var exportTypes []string

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".csv") {
			csvFiles = append(csvFiles, file.Name())

			// Get file info
			filePath := filepath.Join(dirPath, file.Name())
			if info, err := file.Info(); err == nil {
				totalSize += info.Size()
			}

			// Count records
			if records := countCSVRecords(filePath); records > 0 {
				totalRecords += records
			}

			// Determine export type from filename
			if exportType := parseExportType(file.Name()); exportType != "" {
				exportTypes = append(exportTypes, exportType)
			}
		}
	}

	if len(csvFiles) == 0 {
		return nil
	}

	// Determine main export type
	mainType := "all"
	if len(exportTypes) == 1 {
		mainType = exportTypes[0]
	} else if len(exportTypes) > 1 {
		mainType = "all" // Multiple types = complete export
	}

	return &ExportItem{
		ID:          fmt.Sprintf("dir_%s", dirName),
		Type:        mainType,
		Date:        exportDate,
		Status:      "completed",
		Duration:    estimateExportDuration(totalRecords),
		FileSize:    formatFileSize(totalSize),
		RecordCount: totalRecords,
		Files:       csvFiles,
	}
}

func processCSVFile(filePath, fileName string) *ExportItem {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil
	}

	exportType := parseExportType(fileName)
	if exportType == "" {
		exportType = "unknown"
	}

	recordCount := countCSVRecords(filePath)

	return &ExportItem{
		ID:          fmt.Sprintf("file_%s_%d", exportType, info.ModTime().Unix()),
		Type:        exportType,
		Date:        info.ModTime(),
		Status:      "completed",
		Duration:    estimateExportDuration(recordCount),
		FileSize:    formatFileSize(info.Size()),
		RecordCount: recordCount,
		Files:       []string{fileName},
	}
}

func parseExportType(filename string) string {
	filename = strings.ToLower(filename)

	if strings.Contains(filename, "watched") {
		return "watched"
	} else if strings.Contains(filename, "collection") {
		return "collection"
	} else if strings.Contains(filename, "shows") || strings.Contains(filename, "tv") {
		return "shows"
	} else if strings.Contains(filename, "ratings") {
		return "ratings"
	} else if strings.Contains(filename, "watchlist") {
		return "watchlist"
	}

	return ""
}

func countCSVRecords(filename string) int {
	content, err := os.ReadFile(filename)
	if err != nil {
		return 0
	}

	lines := strings.Split(string(content), "\n")
	// Subtract 1 for header row, and filter out empty lines
	count := 0
	for i, line := range lines {
		if i == 0 { // Skip header
			continue
		}
		if strings.TrimSpace(line) != "" {
			count++
		}
	}

	return count
}

func estimateExportDuration(recordCount int) string {
	if recordCount == 0 {
		return "< 1s"
	}

	// Rough estimate: 100 records per second
	seconds := recordCount / 100
	if seconds < 1 {
		return "< 1s"
	} else if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	} else if seconds < 3600 {
		minutes := seconds / 60
		remainingSeconds := seconds % 60
		if remainingSeconds > 0 {
			return fmt.Sprintf("%dm %ds", minutes, remainingSeconds)
		}
		return fmt.Sprintf("%dm", minutes)
	} else {
		hours := seconds / 3600
		minutes := (seconds % 3600) / 60
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
}

func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func getAuthSection(isValid bool) string {
	if !isValid {
		return `<div class="auth-required">
			<p>🔐 Authentication required to perform exports</p>
			<a href="/auth-url" class="btn btn-primary">Authenticate with Trakt.tv</a>
		</div>`
	}
	return ""
}

func getExportsSection(exports []ExportItem) string {
	if len(exports) == 0 {
		return `<div class="no-exports">
			<div style="font-size: 3rem; margin-bottom: 1rem;">📭</div>
			<h3>No exports found</h3>
			<p>Your export history will appear here once you start your first export.</p>
			<p>Use the export buttons above to get started!</p>
		</div>`
	}

	html := `<div class="export-list">`

	for _, export := range exports {
		// Determine export type icon and name
		typeIcon := "📄"
		typeName := export.Type
		switch export.Type {
		case "all":
			typeIcon = "📦"
			typeName = "Complete Export"
		case "watched":
			typeIcon = "🎬"
			typeName = "Watched Movies"
		case "collection":
			typeIcon = "📚"
			typeName = "Collection"
		case "shows":
			typeIcon = "📺"
			typeName = "TV Shows"
		case "ratings":
			typeIcon = "⭐"
			typeName = "Ratings"
		case "watchlist":
			typeIcon = "📝"
			typeName = "Watchlist"
		}

		// Build download links
		downloadLinks := ""
		for _, file := range export.Files {
			var downloadPath string
			if strings.HasPrefix(export.ID, "dir_") {
				// For directory exports, include the directory name in the path
				dirName := strings.TrimPrefix(export.ID, "dir_")
				downloadPath = fmt.Sprintf("/download/%s/%s", dirName, file)
			} else {
				// For individual files, use direct path
				downloadPath = fmt.Sprintf("/download/%s", file)
			}

			downloadLinks += fmt.Sprintf(`
				<a href="%s" class="btn btn-secondary" title="Download %s">
					📥 %s
				</a>`, downloadPath, file, file)
		}

		html += fmt.Sprintf(`
			<div class="export-item">
				<div class="export-info">
					<h4>%s %s</h4>
					<div class="export-details">
						<span>📅 %s</span>
						<span>⏱️ %s</span>
						<span>💾 %s</span>
						<span>📊 %d records</span>
						<span>📁 %d files</span>
						<span class="status-indicator completed">Completed</span>
					</div>
				</div>
				<div class="export-actions">
					%s
				</div>
			</div>`,
			typeIcon, typeName,
			export.Date.Format("2006-01-02 15:04"),
			export.Duration,
			export.FileSize,
			export.RecordCount,
			len(export.Files),
			downloadLinks)
	}

	html += `</div>`
	return html
}
