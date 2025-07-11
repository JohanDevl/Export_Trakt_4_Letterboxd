package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/auth"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

type ExportsData struct {
	Title        string
	CurrentPage  string
	ServerStatus string
	LastUpdated  string
	TokenStatus  *TokenStatusData
	Exports      []ExportItem
	Alert        *AlertData
}

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

type ExportAPIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type ExportsHandler struct {
	config       *config.Config
	logger       logger.Logger
	tokenManager *auth.TokenManager
	templates    *template.Template
	exportsDir   string
}

func NewExportsHandler(cfg *config.Config, log logger.Logger, tokenManager *auth.TokenManager, templates *template.Template) *ExportsHandler {
	exportsDir := "./exports"
	if cfg.Letterboxd.ExportDir != "" {
		exportsDir = cfg.Letterboxd.ExportDir
	}
	
	return &ExportsHandler{
		config:       cfg,
		logger:       log,
		tokenManager: tokenManager,
		templates:    templates,
		exportsDir:   exportsDir,
	}
}

func (h *ExportsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.handleGetExports(w, r)
	case "POST":
		h.handleStartExport(w, r)
	case "DELETE":
		h.handleDeleteExport(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ExportsHandler) handleGetExports(w http.ResponseWriter, r *http.Request) {
	data := h.prepareExportsData()
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "exports.html", data); err != nil {
		h.logger.Error("web.template_error", map[string]interface{}{
			"error":    err.Error(),
			"template": "exports.html",
		})
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *ExportsHandler) handleStartExport(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	tokenStatus, err := h.tokenManager.GetTokenStatus()
	if err != nil || !tokenStatus.IsValid {
		h.writeJSONResponse(w, ExportAPIResponse{
			Success: false,
			Error:   "Authentication required",
		})
		return
	}
	
	exportType := r.URL.Query().Get("type")
	if exportType == "" {
		exportType = "watched"
	}
	
	historyMode := r.URL.Query().Get("historyMode")
	if historyMode == "" {
		historyMode = "aggregated"
	}
	
	h.logger.Info("web.export_started", map[string]interface{}{
		"type":         exportType,
		"history_mode": historyMode,
		"client_ip":    r.RemoteAddr,
	})
	
	// In a real implementation, this would trigger the actual export
	// For now, we'll just log it and return success
	h.writeJSONResponse(w, ExportAPIResponse{
		Success: true,
		Data: map[string]interface{}{
			"export_id": fmt.Sprintf("export_%d", time.Now().Unix()),
			"type":      exportType,
			"status":    "started",
		},
	})
}

func (h *ExportsHandler) handleDeleteExport(w http.ResponseWriter, r *http.Request) {
	exportID := strings.TrimPrefix(r.URL.Path, "/api/export/")
	if exportID == "" {
		h.writeJSONResponse(w, ExportAPIResponse{
			Success: false,
			Error:   "Export ID required",
		})
		return
	}
	
	h.logger.Info("web.export_deleted", map[string]interface{}{
		"export_id": exportID,
		"client_ip": r.RemoteAddr,
	})
	
	// In a real implementation, this would delete the export files and metadata
	h.writeJSONResponse(w, ExportAPIResponse{
		Success: true,
		Data:    map[string]interface{}{"deleted": exportID},
	})
}

func (h *ExportsHandler) prepareExportsData() *ExportsData {
	data := &ExportsData{
		Title:        "Export Management",
		CurrentPage:  "exports",
		ServerStatus: "healthy",
		LastUpdated:  time.Now().Format("2006-01-02 15:04:05"),
	}
	
	// Get token status
	if tokenStatus, err := h.tokenManager.GetTokenStatus(); err == nil {
		data.TokenStatus = &TokenStatusData{
			IsValid:         tokenStatus.IsValid,
			HasToken:        tokenStatus.HasToken,
			ExpiresAt:       tokenStatus.ExpiresAt,
			HasRefreshToken: tokenStatus.HasRefreshToken,
		}
	} else {
		data.TokenStatus = &TokenStatusData{
			IsValid:  false,
			HasToken: false,
		}
	}
	
	// Scan for existing export files
	data.Exports = h.scanExportFiles()
	
	return data
}

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
	
	// Sort by date (newest first)
	for i := 0; i < len(exports)-1; i++ {
		for j := i + 1; j < len(exports); j++ {
			if exports[i].Date.Before(exports[j].Date) {
				exports[i], exports[j] = exports[j], exports[i]
			}
		}
	}
	
	h.logger.Info("web.exports_scanned", map[string]interface{}{
		"count": len(exports),
	})
	
	return exports
}

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
			filePath := filepath.Join(dirPath, file.Name())
			if info, err := file.Info(); err == nil {
				totalSize += info.Size()
			}
			
			// Count records
			if records := h.countCSVRecords(filePath); records > 0 {
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
	
	recordCount := h.countCSVRecords(filePath)
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

func (h *ExportsHandler) countCSVRecords(filename string) int {
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

func (h *ExportsHandler) formatFileSize(size int64) string {
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

func (h *ExportsHandler) writeJSONResponse(w http.ResponseWriter, response ExportAPIResponse) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("web.json_encode_error", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// DownloadHandler handles file downloads
type DownloadHandler struct {
	exportsDir string
	logger     logger.Logger
}

func NewDownloadHandler(exportsDir string, log logger.Logger) *DownloadHandler {
	return &DownloadHandler{
		exportsDir: exportsDir,
		logger:     log,
	}
}

func (h *DownloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Parse the URL path to extract the file path
	urlPath := strings.TrimPrefix(r.URL.Path, "/download/")
	if urlPath == "" {
		http.Error(w, "File path required", http.StatusBadRequest)
		return
	}
	
	// Handle both direct files and files in subdirectories
	var fullPath string
	
	// Check if it's a path with directory (e.g., "export_2025-07-11_15-43/watched.csv")
	if strings.Contains(urlPath, "/") {
		// For subdirectory files, use the full relative path
		fullPath = filepath.Join(h.exportsDir, urlPath)
	} else {
		// For direct files, just add to exports directory
		fullPath = filepath.Join(h.exportsDir, urlPath)
	}
	
	// Security check: ensure the file is within the exports directory
	absExportsDir, err := filepath.Abs(h.exportsDir)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	absFilePath, err := filepath.Abs(fullPath)
	if err != nil {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}
	
	if !strings.HasPrefix(absFilePath, absExportsDir) {
		h.logger.Warn("web.download_access_denied", map[string]interface{}{
			"requested_path": urlPath,
			"client_ip":      r.RemoteAddr,
		})
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}
	
	// Check if file exists
	if _, err := os.Stat(absFilePath); os.IsNotExist(err) {
		h.logger.Warn("web.download_file_not_found", map[string]interface{}{
			"requested_path": urlPath,
			"full_path":      absFilePath,
			"client_ip":      r.RemoteAddr,
		})
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	
	h.logger.Info("web.file_download", map[string]interface{}{
		"requested_path": urlPath,
		"full_path":      absFilePath,
		"client_ip":      r.RemoteAddr,
	})
	
	// Extract just the filename for the download
	filename := filepath.Base(absFilePath)
	
	// Set headers for download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", h.getFileSize(absFilePath)))
	
	// Serve the file
	http.ServeFile(w, r, absFilePath)
}

func (h *DownloadHandler) getFileSize(filepath string) int64 {
	if info, err := os.Stat(filepath); err == nil {
		return info.Size()
	}
	return 0
}