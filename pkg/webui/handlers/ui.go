package handlers

import (
	"html/template"
	"net/http"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

// UIHandler handles UI-related requests
type UIHandler struct {
	templates *template.Template
	logger    logger.Logger
}

// NewUIHandler creates a new UI handler
func NewUIHandler(templates *template.Template, log logger.Logger) *UIHandler {
	return &UIHandler{
		templates: templates,
		logger:    log,
	}
}

// PageData represents data passed to templates
type PageData struct {
	Title       string
	Section     string
	Version     string
	CurrentPath string
}

// Dashboard renders the dashboard page
func (h *UIHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:       "Dashboard - Export Trakt 4 Letterboxd",
		Section:     "dashboard",
		Version:     "2.0.0-dev",
		CurrentPath: r.URL.Path,
	}

	h.logger.Info("ui.dashboard_requested", nil)

	if err := h.templates.ExecuteTemplate(w, "dashboard.html", data); err != nil {
		h.logger.Error("ui.template_error", map[string]interface{}{
			"template": "dashboard.html",
			"error":    err.Error(),
		})
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}

// Config renders the configuration page
func (h *UIHandler) Config(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:       "Configuration - Export Trakt 4 Letterboxd",
		Section:     "config",
		Version:     "2.0.0-dev",
		CurrentPath: r.URL.Path,
	}

	h.logger.Info("ui.config_requested", nil)

	if err := h.templates.ExecuteTemplate(w, "config.html", data); err != nil {
		h.logger.Error("ui.template_error", map[string]interface{}{
			"template": "config.html",
			"error":    err.Error(),
		})
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}

// Exports renders the exports page
func (h *UIHandler) Exports(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:       "Exports - Export Trakt 4 Letterboxd",
		Section:     "exports",
		Version:     "2.0.0-dev",
		CurrentPath: r.URL.Path,
	}

	h.logger.Info("ui.exports_requested", nil)

	if err := h.templates.ExecuteTemplate(w, "exports.html", data); err != nil {
		h.logger.Error("ui.template_error", map[string]interface{}{
			"template": "exports.html",
			"error":    err.Error(),
		})
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}

// Monitoring renders the monitoring page
func (h *UIHandler) Monitoring(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:       "Monitoring - Export Trakt 4 Letterboxd",
		Section:     "monitoring",
		Version:     "2.0.0-dev",
		CurrentPath: r.URL.Path,
	}

	h.logger.Info("ui.monitoring_requested", nil)

	if err := h.templates.ExecuteTemplate(w, "monitoring.html", data); err != nil {
		h.logger.Error("ui.template_error", map[string]interface{}{
			"template": "monitoring.html",
			"error":    err.Error(),
		})
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}

// Logs renders the logs page
func (h *UIHandler) Logs(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:       "Logs - Export Trakt 4 Letterboxd",
		Section:     "logs",
		Version:     "2.0.0-dev",
		CurrentPath: r.URL.Path,
	}

	h.logger.Info("ui.logs_requested", nil)

	if err := h.templates.ExecuteTemplate(w, "logs.html", data); err != nil {
		h.logger.Error("ui.template_error", map[string]interface{}{
			"template": "logs.html",
			"error":    err.Error(),
		})
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
} 