package handlers

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/gorilla/mux"
)

// ExportHandler handles export-related requests
type ExportHandler struct {
	config *config.Config
	logger logger.Logger
}

// NewExportHandler creates a new export handler
func NewExportHandler(cfg *config.Config, log logger.Logger) *ExportHandler {
	return &ExportHandler{
		config: cfg,
		logger: log,
	}
}

// ExportInfo represents export information
type ExportInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Size        int64     `json:"size"`
	Format      string    `json:"format"`
	Path        string    `json:"path"`
	Error       string    `json:"error,omitempty"`
}

// ListExports returns a list of all exports
func (h *ExportHandler) ListExports(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	exportDir := h.config.Letterboxd.ExportDir
	if exportDir == "" {
		exportDir = "./exports"
	}

	var exports []ExportInfo

	// Read exports directory
	if _, err := os.Stat(exportDir); os.IsNotExist(err) {
		// Create exports directory if it doesn't exist
		os.MkdirAll(exportDir, 0755)
	}

	err := filepath.WalkDir(exportDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if d.IsDir() && path != exportDir {
			// This is an export directory
			info, err := d.Info()
			if err != nil {
				return nil
			}

			exportID := d.Name()
			exportInfo := ExportInfo{
				ID:        exportID,
				Name:      exportID,
				Status:    "completed", // Assume completed if directory exists
				CreatedAt: info.ModTime(),
				Size:      info.Size(),
				Format:    "csv",
				Path:      path,
			}

			// Check if there are files in the export directory
			files, err := os.ReadDir(path)
			if err == nil && len(files) > 0 {
				// Calculate total size
				totalSize := int64(0)
				for _, file := range files {
					if !file.IsDir() {
						if fileInfo, err := file.Info(); err == nil {
							totalSize += fileInfo.Size()
						}
					}
				}
				exportInfo.Size = totalSize
			}

			exports = append(exports, exportInfo)
		}

		return nil
	})

	if err != nil {
		h.logger.Error("export.list_failed", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "Failed to list exports", http.StatusInternalServerError)
		return
	}

	h.logger.Info("export.list_success", map[string]interface{}{
		"count": len(exports),
	})

	json.NewEncoder(w).Encode(exports)
}

// StartExport starts a new export
func (h *ExportHandler) StartExport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// TODO: Implement actual export logic
	// This would involve calling the existing export functionality

	exportID := fmt.Sprintf("export_%s", time.Now().Format("2006-01-02_15-04"))
	
	h.logger.Info("export.start_requested", map[string]interface{}{
		"export_id": exportID,
	})

	response := map[string]interface{}{
		"success":   true,
		"export_id": exportID,
		"message":   "Export started successfully",
		"status":    "running",
	}

	json.NewEncoder(w).Encode(response)
}

// GetExport returns information about a specific export
func (h *ExportHandler) GetExport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	exportID := vars["id"]

	if exportID == "" {
		http.Error(w, "Export ID is required", http.StatusBadRequest)
		return
	}

	exportDir := h.config.Letterboxd.ExportDir
	if exportDir == "" {
		exportDir = "./exports"
	}

	exportPath := filepath.Join(exportDir, exportID)
	
	info, err := os.Stat(exportPath)
	if os.IsNotExist(err) {
		http.Error(w, "Export not found", http.StatusNotFound)
		return
	}

	if err != nil {
		h.logger.Error("export.get_failed", map[string]interface{}{
			"export_id": exportID,
			"error":     err.Error(),
		})
		http.Error(w, "Failed to get export information", http.StatusInternalServerError)
		return
	}

	exportInfo := ExportInfo{
		ID:        exportID,
		Name:      exportID,
		Status:    "completed",
		CreatedAt: info.ModTime(),
		Size:      info.Size(),
		Format:    "csv",
		Path:      exportPath,
	}

	h.logger.Info("export.get_success", map[string]interface{}{
		"export_id": exportID,
	})

	json.NewEncoder(w).Encode(exportInfo)
}

// DeleteExport deletes a specific export
func (h *ExportHandler) DeleteExport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	exportID := vars["id"]

	if exportID == "" {
		http.Error(w, "Export ID is required", http.StatusBadRequest)
		return
	}

	exportDir := h.config.Letterboxd.ExportDir
	if exportDir == "" {
		exportDir = "./exports"
	}

	exportPath := filepath.Join(exportDir, exportID)
	
	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		http.Error(w, "Export not found", http.StatusNotFound)
		return
	}

	if err := os.RemoveAll(exportPath); err != nil {
		h.logger.Error("export.delete_failed", map[string]interface{}{
			"export_id": exportID,
			"error":     err.Error(),
		})
		http.Error(w, "Failed to delete export", http.StatusInternalServerError)
		return
	}

	h.logger.Info("export.delete_success", map[string]interface{}{
		"export_id": exportID,
	})

	response := map[string]interface{}{
		"success": true,
		"message": "Export deleted successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// DownloadExport allows downloading a specific export file
func (h *ExportHandler) DownloadExport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	exportID := vars["id"]

	if exportID == "" {
		http.Error(w, "Export ID is required", http.StatusBadRequest)
		return
	}

	exportDir := h.config.Letterboxd.ExportDir
	if exportDir == "" {
		exportDir = "./exports"
	}

	exportPath := filepath.Join(exportDir, exportID)
	
	// Check if export exists
	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		http.Error(w, "Export not found", http.StatusNotFound)
		return
	}

	// Get the file parameter (which file to download)
	filename := r.URL.Query().Get("file")
	if filename == "" {
		// Default to the main export file
		filename = "trakt_export.csv"
	}

	filePath := filepath.Join(exportPath, filename)
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	h.logger.Info("export.download_requested", map[string]interface{}{
		"export_id": exportID,
		"filename":  filename,
	})

	// Set headers for file download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/octet-stream")

	// Serve the file
	http.ServeFile(w, r, filePath)
} 