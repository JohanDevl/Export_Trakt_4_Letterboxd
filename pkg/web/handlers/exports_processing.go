package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// scanExportFilesOptimized scanne les exports de manière optimisée
func (h *ExportsHandler) scanExportFilesOptimized() []ExportItem {
	var exports []ExportItem

	// Check if exports directory exists
	if _, err := os.Stat(h.exportsDir); os.IsNotExist(err) {
		h.logger.Info("web.exports_dir_not_found", map[string]interface{}{
			"dir": h.exportsDir,
		})
		return exports
	}

	// Scan récent en premier (30 derniers jours) pour des performances optimales
	recentExports := h.scanRecentExports(30)
	exports = append(exports, recentExports...)
	h.logger.Info("web.recent_exports_scanned", map[string]interface{}{
		"recent_count": len(recentExports),
	})

	// Toujours scanner les exports plus anciens pour avoir la liste complète
	olderExports := h.scanOlderExports(30)
	exports = append(exports, olderExports...)
	h.logger.Info("web.older_exports_scanned", map[string]interface{}{
		"older_count": len(olderExports),
		"total_before_sort": len(exports),
	})

	// Trier par date (plus récent en premier)
	sort.Slice(exports, func(i, j int) bool {
		return exports[i].Date.After(exports[j].Date)
	})

	h.logger.Info("web.exports_scanned_optimized", map[string]interface{}{
		"count": len(exports),
	})

	return exports
}

// scanRecentExports scanne seulement les exports récents
func (h *ExportsHandler) scanRecentExports(days int) []ExportItem {
	var exports []ExportItem
	cutoffTime := time.Now().AddDate(0, 0, -days)

	entries, err := os.ReadDir(h.exportsDir)
	if err != nil {
		h.logger.Error("web.scan_exports_dir_error", map[string]interface{}{
			"error": err.Error(),
		})
		return exports
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			// Handle individual CSV files
			if strings.HasSuffix(strings.ToLower(entry.Name()), ".csv") {
				info, _ := entry.Info()
				if info.ModTime().After(cutoffTime) {
					export := h.processCSVFileOptimized(filepath.Join(h.exportsDir, entry.Name()), entry.Name())
					if export != nil {
						exports = append(exports, *export)
					}
				}
			}
			continue
		}

		// Check timestamped directories
		dirName := entry.Name()
		if strings.HasPrefix(dirName, "export_") && len(dirName) >= 16 {
			// Parse date from directory name quickly
			if dirTime := h.parseDirTime(dirName); dirTime.After(cutoffTime) {
				export := h.processExportDirectoryOptimized(filepath.Join(h.exportsDir, dirName), dirName)
				if export != nil {
					exports = append(exports, *export)
				}
			}
		}
	}

	return exports
}

// scanOlderExports scanne les exports plus anciens si nécessaire
func (h *ExportsHandler) scanOlderExports(skipDays int) []ExportItem {
	var exports []ExportItem
	cutoffTime := time.Now().AddDate(0, 0, -skipDays)

	entries, err := os.ReadDir(h.exportsDir)
	if err != nil {
		return exports
	}

	// Limiter le scan aux 100 premiers dossiers les plus anciens pour améliorer les performances
	count := 0
	for _, entry := range entries {
		if count >= 100 {
			break
		}

		if !entry.IsDir() {
			if strings.HasSuffix(strings.ToLower(entry.Name()), ".csv") {
				info, _ := entry.Info()
				if info.ModTime().Before(cutoffTime) {
					export := h.processCSVFileOptimized(filepath.Join(h.exportsDir, entry.Name()), entry.Name())
					if export != nil {
						exports = append(exports, *export)
						count++
					}
				}
			}
			continue
		}

		dirName := entry.Name()
		if strings.HasPrefix(dirName, "export_") && len(dirName) >= 16 {
			if dirTime := h.parseDirTime(dirName); dirTime.Before(cutoffTime) {
				export := h.processExportDirectoryOptimized(filepath.Join(h.exportsDir, dirName), dirName)
				if export != nil {
					exports = append(exports, *export)
					count++
				}
			}
		}
	}

	return exports
}

// parseDirTime parse rapidement la date d'un nom de dossier
func (h *ExportsHandler) parseDirTime(dirName string) time.Time {
	parts := strings.Split(dirName, "_")
	if len(parts) < 3 {
		return time.Time{}
	}

	dateStr := parts[1] + " " + strings.ReplaceAll(parts[2], "-", ":")
	if exportDate, err := time.Parse("2006-01-02 15:04", dateStr); err == nil {
		return exportDate
	}
	return time.Time{}
}

// Ancienne méthode de scan complète - conservée pour référence
func (h *ExportsHandler) scanExportFiles() []ExportItem {
	var exports []ExportItem

	// Check if exports directory exists
	if _, err := os.Stat(h.exportsDir); os.IsNotExist(err) {
		h.logger.Info("web.exports_dir_not_found", map[string]interface{}{
			"dir": h.exportsDir,
		})
		return exports
	}

	// First scan for timestamped export directories (export_YYYY-MM-DD_HH-MM format)
	entries, err := os.ReadDir(h.exportsDir)
	if err != nil {
		h.logger.Error("web.scan_exports_dir_error", map[string]interface{}{
			"error": err.Error(),
		})
		return exports
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			// Handle individual CSV files in root exports directory
			if strings.HasSuffix(strings.ToLower(entry.Name()), ".csv") {
				export := h.processCSVFile(filepath.Join(h.exportsDir, entry.Name()), entry.Name())
				if export != nil {
					exports = append(exports, *export)
				}
			}
			continue
		}

		// Check if directory name matches export timestamp pattern
		dirName := entry.Name()
		if strings.HasPrefix(dirName, "export_") && len(dirName) >= 16 {
			export := h.processExportDirectory(filepath.Join(h.exportsDir, dirName), dirName)
			if export != nil {
				exports = append(exports, *export)
			}
		}
	}

	// Sort by date (newest first) - utiliser sort.Slice qui est plus efficace
	sort.Slice(exports, func(i, j int) bool {
		return exports[i].Date.After(exports[j].Date)
	})

	h.logger.Info("web.exports_scanned", map[string]interface{}{
		"count": len(exports),
	})

	return exports
}

// processExportDirectoryOptimized traite un dossier d'export de manière optimisée
func (h *ExportsHandler) processExportDirectoryOptimized(dirPath, dirName string) *ExportItem {
	// Parse timestamp from directory name
	exportDate := h.parseDirTime(dirName)
	if exportDate.IsZero() {
		// Fallback to directory modification time
		if info, err := os.Stat(dirPath); err == nil {
			exportDate = info.ModTime()
		} else {
			return nil
		}
	}

	// Scan files optimisé - ne pas lire tous les contenus immédiatement
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil
	}

	var csvFiles []string
	var totalSize int64
	var estimatedRecords int
	var exportTypes []string

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".csv") {
			csvFiles = append(csvFiles, file.Name())

			// Get file info
			if info, err := file.Info(); err == nil {
				totalSize += info.Size()
				// Estimation rapide: ~80 caractères par ligne en moyenne
				estimatedRecords += int(info.Size() / 80)
			}

			// Déterminer le type d'export
			if exportType := h.parseExportType(file.Name()); exportType != "" {
				exportTypes = append(exportTypes, exportType)
			}
		}
	}

	if len(csvFiles) == 0 {
		return nil
	}

	// Déterminer le type principal
	mainType := "all"
	if len(exportTypes) == 1 {
		mainType = exportTypes[0]
	} else if len(exportTypes) > 1 {
		mainType = "all"
	}

	// Estimation de durée
	duration := h.estimateExportDuration(estimatedRecords)

	return &ExportItem{
		ID:          fmt.Sprintf("dir_%s", dirName),
		Type:        mainType,
		Date:        exportDate,
		Status:      "completed",
		Duration:    duration,
		FileSize:    h.formatFileSize(totalSize),
		RecordCount: estimatedRecords,
		Files:       csvFiles,
	}
}

// processCSVFileOptimized traite un fichier CSV de manière optimisée
func (h *ExportsHandler) processCSVFileOptimized(filePath, fileName string) *ExportItem {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil
	}

	exportType := h.parseExportType(fileName)
	if exportType == "" {
		exportType = "unknown"
	}

	// Estimation rapide du nombre d'enregistrements
	estimatedRecords := int(info.Size() / 80) // ~80 caractères par ligne
	duration := h.estimateExportDuration(estimatedRecords)

	return &ExportItem{
		ID:          fmt.Sprintf("file_%s_%d", exportType, info.ModTime().Unix()),
		Type:        exportType,
		Date:        info.ModTime(),
		Status:      "completed",
		Duration:    duration,
		FileSize:    h.formatFileSize(info.Size()),
		RecordCount: estimatedRecords,
		Files:       []string{fileName},
	}
}

// Version originale conservée pour compatibilità
func (h *ExportsHandler) processExportDirectory(dirPath, dirName string) *ExportItem {
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
			if info, err := file.Info(); err == nil {
				totalSize += info.Size()
			}

			// Count records optimisé
			if records := h.countCSVRecordsOptimized(filepath.Join(dirPath, file.Name())); records > 0 {
				totalRecords += records
			}

			// Determine export type from filename
			if exportType := h.parseExportType(file.Name()); exportType != "" {
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

	// Calculate duration estimate (mock for now)
	duration := h.estimateExportDuration(totalRecords)

	return &ExportItem{
		ID:          fmt.Sprintf("dir_%s", dirName),
		Type:        mainType,
		Date:        exportDate,
		Status:      "completed",
		Duration:    duration,
		FileSize:    h.formatFileSize(totalSize),
		RecordCount: totalRecords,
		Files:       csvFiles,
	}
}

func (h *ExportsHandler) processCSVFile(filePath, fileName string) *ExportItem {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil
	}

	exportType := h.parseExportType(fileName)
	if exportType == "" {
		exportType = "unknown"
	}

	recordCount := h.countCSVRecordsOptimized(filePath)
	duration := h.estimateExportDuration(recordCount)

	return &ExportItem{
		ID:          fmt.Sprintf("file_%s_%d", exportType, info.ModTime().Unix()),
		Type:        exportType,
		Date:        info.ModTime(),
		Status:      "completed",
		Duration:    duration,
		FileSize:    h.formatFileSize(info.Size()),
		RecordCount: recordCount,
		Files:       []string{fileName},
	}
}

func (h *ExportsHandler) estimateExportDuration(recordCount int) string {
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

func (h *ExportsHandler) parseExportType(filename string) string {
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
