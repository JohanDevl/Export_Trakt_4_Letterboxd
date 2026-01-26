package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/auth"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/web/middleware"
)

const (
	ExportCacheTTL = 30 * time.Minute
)

type ExportsData struct {
	Title        string
	CurrentPage  string
	ServerStatus string
	LastUpdated  string
	CSRFToken    string
	TokenStatus  *TokenStatusData
	Exports      []ExportItem
	Alert        *AlertData
	Pagination   *PaginationData
}

type PaginationData struct {
	CurrentPage  int
	TotalPages   int
	TotalItems   int
	ItemsPerPage int
	HasPrevious  bool
	HasNext      bool
	ShowFirst    bool
	ShowLast     bool
	PageNumbers  []int
}

type ExportItem struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Date        time.Time `json:"date"`
	Status      string    `json:"status"`
	Duration    string    `json:"duration"`
	FileSize    string    `json:"fileSize"`
	RecordCount int       `json:"recordCount"`
	Files       []string  `json:"files"`
	Error       string    `json:"error"`
}

type ExportAPIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type ExportCache struct {
	mu              sync.RWMutex
	exports         []ExportItem
	lastScan        time.Time
	cacheTTL        time.Duration
	recentExports   []ExportItem // Cache for recent exports (7 days)
	recentLastScan  time.Time
	recentCacheTTL  time.Duration // Shorter TTL for recent exports
}

// invalidateCache clears the cache to force a refresh
func (h *ExportsHandler) invalidateCache() {
	h.cache.mu.Lock()
	defer h.cache.mu.Unlock()
	h.cache.exports = nil
	h.cache.lastScan = time.Time{}
	h.cache.recentExports = nil
	h.cache.recentLastScan = time.Time{}
}

type ExportsHandler struct {
	config         *config.Config
	logger         logger.Logger
	tokenManager   *auth.TokenManager
	templates      *template.Template
	exportsDir     string
	cache          *ExportCache
	csrfMiddleware *middleware.CSRFMiddleware
}

func NewExportsHandler(cfg *config.Config, log logger.Logger, tokenManager *auth.TokenManager, templates *template.Template, csrfMiddleware *middleware.CSRFMiddleware) *ExportsHandler {
	exportsDir := "./exports"
	if cfg.Letterboxd.ExportDir != "" {
		exportsDir = cfg.Letterboxd.ExportDir
	}

	return &ExportsHandler{
		config:         cfg,
		logger:         log,
		tokenManager:   tokenManager,
		templates:      templates,
		exportsDir:     exportsDir,
		csrfMiddleware: csrfMiddleware,
		cache: &ExportCache{
			cacheTTL:       ExportCacheTTL, // Cache pendant 30 minutes pour de meilleures performances
			recentCacheTTL: 1 * time.Minute,  // Cache des exports récents refresh plus souvent
		},
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
	// Check if this is an AJAX request for pagination
	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		h.handleGetExportsPaginated(w, r)
		return
	}

	data := h.prepareExportsData(r)

	// Debug log the data being passed to template
	h.logger.Info("web.template_data_debug", map[string]interface{}{
		"exports_count": len(data.Exports),
		"pagination_nil": data.Pagination == nil,
		"token_status_nil": data.TokenStatus == nil,
	})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "exports.html", data); err != nil {
		h.logger.Error("web.template_error", map[string]interface{}{
			"error":    err.Error(),
			"template": "exports.html",
			"exports_count": len(data.Exports),
			"data_exports_nil": data.Exports == nil,
			"pagination_nil": data.Pagination == nil,
		})
		// Don't call http.Error since headers might already be written
		w.Write([]byte("Template Error: " + err.Error()))
		return
	}
}

func (h *ExportsHandler) handleGetExportsPaginated(w http.ResponseWriter, r *http.Request) {
	data := h.prepareExportsData(r)

	response := struct {
		Exports    []ExportItem    `json:"exports"`
		Pagination *PaginationData `json:"pagination"`
	}{
		Exports:    data.Exports,
		Pagination: data.Pagination,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("web.json_encode_error", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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

	// Start export in background
	exportID := fmt.Sprintf("export_%d", time.Now().Unix())
	go h.runExportAsync(exportID, exportType, historyMode)

	h.writeJSONResponse(w, ExportAPIResponse{
		Success: true,
		Data: map[string]interface{}{
			"export_id": exportID,
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

func (h *ExportsHandler) prepareExportsData(r *http.Request) *ExportsData {
	data := &ExportsData{
		Title:        "Export Management",
		CurrentPage:  "exports",
		ServerStatus: "healthy",
		LastUpdated:  h.formatTimeInConfigTimezone(time.Now(), "2006-01-02 15:04:05"),
		CSRFToken:    h.csrfMiddleware.GetToken(r),
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

	// Parse pagination parameters
	page := h.getIntParam(r, "page", 1)
	limit := h.getIntParam(r, "limit", 10)

	// Parse filter parameters
	typeFilter := r.URL.Query().Get("type")
	statusFilter := r.URL.Query().Get("status")

	// Validate parameters
	if page < 1 {
		page = 1
	}
	if limit < 5 {
		limit = 5
	} else if limit > 100 {
		limit = 100
	}

	// Get exports with caching and lazy loading
	allExports := h.getExportsWithCache(page, limit)
	h.logger.Info("web.all_exports_loaded", map[string]interface{}{
		"total_exports": len(allExports),
		"page": page,
		"limit": limit,
	})

	// Apply filters
	filteredExports := h.applyFilters(allExports, typeFilter, statusFilter)
	totalItems := len(filteredExports)
	totalPages := (totalItems + limit - 1) / limit
	h.logger.Info("web.exports_after_filter", map[string]interface{}{
		"filtered_count": len(filteredExports),
		"total_items": totalItems,
		"total_pages": totalPages,
		"type_filter": typeFilter,
		"status_filter": statusFilter,
	})

	if totalPages == 0 {
		totalPages = 1
	}

	if page > totalPages {
		page = totalPages
	}

	// Calculate pagination slice
	start := (page - 1) * limit
	end := start + limit
	if end > totalItems {
		end = totalItems
	}

	if start < totalItems {
		// Convert dates to configured timezone before displaying
		paginatedExports := filteredExports[start:end]
		for i := range paginatedExports {
			paginatedExports[i].Date = h.convertToConfigTimezone(paginatedExports[i].Date)
		}
		data.Exports = paginatedExports
	} else {
		data.Exports = []ExportItem{}
	}
	h.logger.Info("web.final_exports_prepared", map[string]interface{}{
		"final_exports_count": len(data.Exports),
		"start": start,
		"end": end,
		"page": page,
	})

	// Build pagination data
	data.Pagination = h.buildPaginationData(page, totalPages, totalItems, limit)

	h.logger.Info("web.pagination_debug", map[string]interface{}{
		"total_items": totalItems,
		"total_pages": totalPages,
		"current_page": page,
		"limit": limit,
		"pagination_nil": data.Pagination == nil,
	})

	return data
}

func (h *ExportsHandler) getIntParam(r *http.Request, key string, defaultValue int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}

	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}

	return defaultValue
}

func (h *ExportsHandler) applyFilters(exports []ExportItem, typeFilter, statusFilter string) []ExportItem {
	if typeFilter == "" && statusFilter == "" {
		return exports
	}

	var filtered []ExportItem
	for _, export := range exports {
		matchesType := typeFilter == "" || export.Type == typeFilter
		matchesStatus := statusFilter == "" || export.Status == statusFilter

		if matchesType && matchesStatus {
			filtered = append(filtered, export)
		}
	}

	return filtered
}

func (h *ExportsHandler) buildPaginationData(currentPage, totalPages, totalItems, itemsPerPage int) *PaginationData {
	pagination := &PaginationData{
		CurrentPage:  currentPage,
		TotalPages:   totalPages,
		TotalItems:   totalItems,
		ItemsPerPage: itemsPerPage,
		HasPrevious:  currentPage > 1,
		HasNext:      currentPage < totalPages,
		ShowFirst:    currentPage > 3,
		ShowLast:     currentPage < totalPages-2,
	}

	// Generate page numbers to show (max 5 pages around current)
	start := currentPage - 2
	end := currentPage + 2

	if start < 1 {
		start = 1
		end = 5
	}

	if end > totalPages {
		end = totalPages
		start = totalPages - 4
	}

	if start < 1 {
		start = 1
	}

	for i := start; i <= end; i++ {
		pagination.PageNumbers = append(pagination.PageNumbers, i)
	}

	return pagination
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

// runExportAsync executes an export command asynchronously
func (h *ExportsHandler) runExportAsync(exportID, exportType, historyMode string) {
	h.logger.Info("web.export_async_started", map[string]interface{}{
		"export_id":    exportID,
		"export_type":  exportType,
		"history_mode": historyMode,
	})

	// Find the current executable path
	execPath, err := os.Executable()
	if err != nil {
		h.logger.Error("web.export_async_failed", map[string]interface{}{
			"export_id": exportID,
			"error":     "Could not find executable path: " + err.Error(),
		})
		return
	}

	// Build command arguments
	args := []string{
		"--run",
		"--export", exportType,
		"--mode", "complete",
	}

	// Add history mode for watched exports
	if exportType == "watched" && historyMode != "" {
		args = append(args, "--history-mode", historyMode)
	}

	h.logger.Info("web.export_async_command", map[string]interface{}{
		"export_id": exportID,
		"command":   execPath,
		"args":      strings.Join(args, " "),
	})

	// Execute the command
	cmd := exec.Command(execPath, args...)
	cmd.Env = os.Environ() // Inherit environment variables

	// Capture both stdout and stderr for better debugging
	output, err := cmd.CombinedOutput()

	if err != nil {
		h.logger.Error("web.export_async_failed", map[string]interface{}{
			"export_id": exportID,
			"error":     err.Error(),
			"output":    string(output),
			"command":   execPath + " " + strings.Join(args, " "),
		})
	} else {
		h.logger.Info("web.export_async_completed", map[string]interface{}{
			"export_id": exportID,
			"output":    string(output),
		})
	}

	// Invalidate cache to ensure new export appears in the list
	h.invalidateCache()
	h.logger.Info("web.export_cache_invalidated", map[string]interface{}{
		"export_id": exportID,
	})
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

	// Check if file exists, if not try to find it in export subdirectories
	finalPath := absFilePath
	if _, err := os.Stat(absFilePath); os.IsNotExist(err) {
		// If direct path doesn't exist and it's a simple filename, search in export directories
		if !strings.Contains(urlPath, "/") {
			foundPath := h.findFileInExportDirs(urlPath)
			if foundPath != "" {
				finalPath = foundPath
				h.logger.Info("web.download_file_found_in_subdir", map[string]interface{}{
					"requested_file": urlPath,
					"found_path":     foundPath,
				})
			} else {
				h.logger.Warn("web.download_file_not_found", map[string]interface{}{
					"requested_path": urlPath,
					"full_path":      absFilePath,
					"client_ip":      r.RemoteAddr,
				})
				http.Error(w, "File not found", http.StatusNotFound)
				return
			}
		} else {
			h.logger.Warn("web.download_file_not_found", map[string]interface{}{
				"requested_path": urlPath,
				"full_path":      absFilePath,
				"client_ip":      r.RemoteAddr,
			})
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
	}

	h.logger.Info("web.file_download", map[string]interface{}{
		"requested_path": urlPath,
		"final_path":     finalPath,
		"client_ip":      r.RemoteAddr,
	})

	// Extract just the filename for the download
	filename := filepath.Base(finalPath)

	// Set headers for download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", h.getFileSize(finalPath)))

	// Serve the file
	http.ServeFile(w, r, finalPath)
}

func (h *DownloadHandler) getFileSize(filepath string) int64 {
	if info, err := os.Stat(filepath); err == nil {
		return info.Size()
	}
	return 0
}

// findFileInExportDirs searches for a file in export subdirectories
func (h *DownloadHandler) findFileInExportDirs(filename string) string {
	// Sanitize filename to prevent path traversal attacks
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") ||
	   strings.Contains(filename, "..") || strings.HasPrefix(filename, ".") {
		return ""
	}

	// Read the exports directory
	entries, err := os.ReadDir(h.exportsDir)
	if err != nil {
		return ""
	}

	// Look in export directories (export_YYYY-MM-DD_HH-MM format)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if directory name starts with "export_"
		dirName := entry.Name()
		if strings.HasPrefix(dirName, "export_") {
			// Check if the file exists in this directory
			possiblePath := filepath.Join(h.exportsDir, dirName, filename)
			if _, err := os.Stat(possiblePath); err == nil {
				// Verify the path is still within exports directory for security
				if absPath, err := filepath.Abs(possiblePath); err == nil {
					if absExportsDir, err := filepath.Abs(h.exportsDir); err == nil {
						if strings.HasPrefix(absPath, absExportsDir) {
							return absPath
						}
					}
				}
			}
		}
	}

	return ""
}
