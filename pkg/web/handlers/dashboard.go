package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/johandevl/Export_Trakt_4_Letterboxd/pkg/auth"
	"github.com/johandevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/johandevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

type DashboardData struct {
	Title           string
	CurrentPage     string
	ServerStatus    string
	LastUpdated     string
	Uptime          string
	Port            int
	TokenStatus     *TokenStatusData
	APIStatus       string
	LastAPICheck    time.Time
	APIResponseTime string
	LastExport      *ExportData
	RecentActivity  []ActivityItem
	Alert           *AlertData
}

type TokenStatusData struct {
	IsValid          bool
	HasToken         bool
	ExpiresAt        time.Time
	HasRefreshToken  bool
	TimeRemaining    string
}

type ExportData struct {
	Type      string
	Date      time.Time
	Status    string
	FileCount int
	Duration  string
	FileSize  string
}

type ActivityItem struct {
	Icon        string
	Title       string
	Description string
	Time        time.Time
}

type AlertData struct {
	Type    string
	Icon    string
	Message string
}

type DashboardHandler struct {
	config       *config.Config
	logger       logger.Logger
	tokenManager *auth.TokenManager
	templates    *template.Template
	startTime    time.Time
}

func NewDashboardHandler(cfg *config.Config, log logger.Logger, tokenManager *auth.TokenManager, templates *template.Template) *DashboardHandler {
	return &DashboardHandler{
		config:       cfg,
		logger:       log,
		tokenManager: tokenManager,
		templates:    templates,
		startTime:    time.Now(),
	}
}

func (h *DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data := h.prepareDashboardData()
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "dashboard.html", data); err != nil {
		h.logger.Error("web.template_error", map[string]interface{}{
			"error":    err.Error(),
			"template": "dashboard.html",
		})
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *DashboardHandler) prepareDashboardData() *DashboardData {
	data := &DashboardData{
		Title:        "Dashboard",
		CurrentPage:  "dashboard",
		ServerStatus: "healthy",
		LastUpdated:  time.Now().Format("2006-01-02 15:04:05"),
		Uptime:       h.formatUptime(),
		Port:         h.config.Auth.CallbackPort,
	}
	
	// Get token status
	if tokenStatus, err := h.tokenManager.GetTokenStatus(); err == nil {
		data.TokenStatus = &TokenStatusData{
			IsValid:         tokenStatus.IsValid,
			HasToken:        tokenStatus.HasToken,
			ExpiresAt:       tokenStatus.ExpiresAt,
			HasRefreshToken: tokenStatus.HasRefreshToken,
			TimeRemaining:   h.formatTimeRemaining(tokenStatus.ExpiresAt),
		}
	} else {
		data.TokenStatus = &TokenStatusData{
			IsValid:  false,
			HasToken: false,
		}
	}
	
	// Mock API status (would be replaced with actual API health check)
	data.APIStatus = "healthy"
	data.LastAPICheck = time.Now().Add(-5 * time.Minute)
	data.APIResponseTime = "142ms"
	
	// Mock last export data (would be replaced with actual export history)
	data.LastExport = &ExportData{
		Type:      "watched",
		Date:      time.Now().Add(-2 * time.Hour),
		Status:    "completed",
		FileCount: 3,
		Duration:  "2m 34s",
		FileSize:  "1.2 MB",
	}
	
	// Mock recent activity (would be replaced with actual activity log)
	data.RecentActivity = []ActivityItem{
		{
			Icon:        "🎬",
			Title:       "Watched Movies Export Completed",
			Description: "Exported 1,234 movies to Letterboxd format",
			Time:        time.Now().Add(-2 * time.Hour),
		},
		{
			Icon:        "🔐",
			Title:       "Token Refreshed",
			Description: "OAuth token automatically refreshed",
			Time:        time.Now().Add(-4 * time.Hour),
		},
		{
			Icon:        "📊",
			Title:       "System Health Check",
			Description: "All systems operational",
			Time:        time.Now().Add(-6 * time.Hour),
		},
	}
	
	// Add alert if token is expired or missing
	if !data.TokenStatus.IsValid {
		data.Alert = &AlertData{
			Type:    "warning",
			Icon:    "⚠️",
			Message: "Authentication required. Please authenticate with Trakt.tv to use export features.",
		}
	}
	
	return data
}

func (h *DashboardHandler) formatUptime() string {
	uptime := time.Since(h.startTime)
	hours := int(uptime.Hours())
	minutes := int(uptime.Minutes()) % 60
	
	if hours > 24 {
		days := hours / 24
		hours = hours % 24
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else {
		return fmt.Sprintf("%dm", minutes)
	}
}

func (h *DashboardHandler) formatTimeRemaining(expiry time.Time) string {
	if expiry.IsZero() {
		return "Unknown"
	}
	
	remaining := time.Until(expiry)
	if remaining <= 0 {
		return "Expired"
	}
	
	hours := int(remaining.Hours())
	minutes := int(remaining.Minutes()) % 60
	
	if hours > 24 {
		days := hours / 24
		return fmt.Sprintf("%d days", days)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else {
		return fmt.Sprintf("%dm", minutes)
	}
}