package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"runtime"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/auth"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

type StatusData struct {
	Title           string
	CurrentPage     string
	ServerStatus    string
	LastUpdated     string
	OverallStatus   string
	OverallMessage  string
	TokenStatus     *TokenStatusData
	TokenTimeRemaining string
	APIStatus       string
	LastAPICheck    time.Time
	APIResponseTime string
	RateLimit       *RateLimitData
	Version         string
	Uptime          string
	Port            int
	BuildDate       string
	GoVersion       string
	ConfigStatus    string
	Config          *ConfigData
	Resources       *ResourcesData
	RecentLogs      []LogEntry
	Alert           *AlertData
}

type RateLimitData struct {
	Limit     int
	Remaining int
	ResetTime *time.Time
}

type ConfigData struct {
	ClientID        string
	RedirectURI     string
	UseOAuth        bool
	PerformanceMode string
	WorkerPoolSize  int
}

type ResourcesData struct {
	MemoryUsed    string
	MemoryTotal   string
	MemoryPercent int
	Goroutines    int
	DiskUsed      string
	DiskTotal     string
	DiskPercent   int
	CacheEntries  int
}

type LogEntry struct {
	Time    time.Time
	Level   string
	Message string
	Context string
}

type StatusHandler struct {
	config       *config.Config
	logger       logger.Logger
	tokenManager *auth.TokenManager
	templates    *template.Template
	startTime    time.Time
}

func NewStatusHandler(cfg *config.Config, log logger.Logger, tokenManager *auth.TokenManager, templates *template.Template) *StatusHandler {
	return &StatusHandler{
		config:       cfg,
		logger:       log,
		tokenManager: tokenManager,
		templates:    templates,
		startTime:    time.Now(),
	}
}

func (h *StatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/status" {
		h.handleAPIStatus(w, r)
		return
	}
	
	if r.URL.Path == "/api/test-connection" {
		h.handleTestConnection(w, r)
		return
	}
	
	if r.URL.Path == "/api/logs/recent" {
		h.handleRecentLogs(w, r)
		return
	}
	
	if r.URL.Path == "/api/logs/download" {
		h.handleDownloadLogs(w, r)
		return
	}
	
	h.handleStatusPage(w, r)
}

func (h *StatusHandler) handleStatusPage(w http.ResponseWriter, r *http.Request) {
	data := h.prepareStatusData()
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "status.html", data); err != nil {
		h.logger.Error("web.template_error", map[string]interface{}{
			"error":    err.Error(),
			"template": "status.html",
		})
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *StatusHandler) handleAPIStatus(w http.ResponseWriter, r *http.Request) {
	data := h.prepareStatusData()
	
	response := ExportAPIResponse{
		Success: true,
		Data: map[string]interface{}{
			"serverStatus": data.ServerStatus,
			"tokenStatus":  data.TokenStatus,
			"apiStatus":    data.APIStatus,
			"uptime":       data.Uptime,
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *StatusHandler) handleTestConnection(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	// Mock API connection test
	// In a real implementation, this would test the actual Trakt.tv API
	time.Sleep(100 * time.Millisecond) // Simulate API call
	
	responseTime := time.Since(start)
	
	response := ExportAPIResponse{
		Success: true,
		Data: map[string]interface{}{
			"status":       "connected",
			"responseTime": fmt.Sprintf("%.0fms", float64(responseTime.Nanoseconds())/1000000),
		},
	}
	
	h.logger.Info("web.api_connection_test", map[string]interface{}{
		"response_time": responseTime.String(),
		"client_ip":     r.RemoteAddr,
	})
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *StatusHandler) handleRecentLogs(w http.ResponseWriter, r *http.Request) {
	// Mock recent logs
	logs := []LogEntry{
		{
			Time:    time.Now().Add(-1 * time.Minute),
			Level:   "info",
			Message: "Export completed successfully",
			Context: "export_type=watched, records=1234",
		},
		{
			Time:    time.Now().Add(-5 * time.Minute),
			Level:   "info",
			Message: "Token refreshed automatically",
			Context: "expires_at=2025-07-12T10:30:00Z",
		},
		{
			Time:    time.Now().Add(-10 * time.Minute),
			Level:   "warning",
			Message: "API rate limit approaching",
			Context: "remaining=15, limit=1000",
		},
	}
	
	response := ExportAPIResponse{
		Success: true,
		Data: map[string]interface{}{
			"logs": logs,
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *StatusHandler) handleDownloadLogs(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would generate and serve actual log files
	logContent := fmt.Sprintf(`# Export Trakt 4 Letterboxd - Log Export
# Generated: %s
# 
[%s] INFO: Server started on port %d
[%s] INFO: OAuth token is valid
[%s] INFO: Export completed successfully
`, 
		time.Now().Format("2006-01-02 15:04:05"),
		time.Now().Add(-1*time.Hour).Format("15:04:05"),
		h.config.Auth.CallbackPort,
		time.Now().Add(-30*time.Minute).Format("15:04:05"),
		time.Now().Add(-5*time.Minute).Format("15:04:05"),
	)
	
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=export-trakt-logs-%s.txt", time.Now().Format("2006-01-02")))
	w.Header().Set("Content-Type", "text/plain")
	
	w.Write([]byte(logContent))
}

func (h *StatusHandler) prepareStatusData() *StatusData {
	data := &StatusData{
		Title:        "System Status",
		CurrentPage:  "status",
		ServerStatus: "healthy",
		LastUpdated:  time.Now().Format("2006-01-02 15:04:05"),
		Version:      "1.0.0",
		Uptime:       h.formatUptime(),
		Port:         h.config.Auth.CallbackPort,
		BuildDate:    "2025-07-11",
		GoVersion:    runtime.Version(),
		ConfigStatus: "healthy",
	}
	
	// Get token status
	if tokenStatus, err := h.tokenManager.GetTokenStatus(); err == nil {
		data.TokenStatus = &TokenStatusData{
			IsValid:         tokenStatus.IsValid,
			HasToken:        tokenStatus.HasToken,
			ExpiresAt:       tokenStatus.ExpiresAt,
			HasRefreshToken: tokenStatus.HasRefreshToken,
		}
		data.TokenTimeRemaining = h.formatTimeRemaining(tokenStatus.ExpiresAt)
	} else {
		data.TokenStatus = &TokenStatusData{
			IsValid:  false,
			HasToken: false,
		}
		data.TokenTimeRemaining = "N/A"
	}
	
	// Mock API status
	data.APIStatus = "healthy"
	data.LastAPICheck = time.Now().Add(-2 * time.Minute)
	data.APIResponseTime = "158ms"
	
	// Mock rate limit data
	data.RateLimit = &RateLimitData{
		Limit:     1000,
		Remaining: 856,
		ResetTime: &[]time.Time{time.Now().Add(45 * time.Minute)}[0],
	}
	
	// Configuration data
	data.Config = &ConfigData{
		ClientID:        h.maskClientID(h.config.Trakt.ClientID),
		RedirectURI:     h.config.Auth.RedirectURI,
		UseOAuth:        h.config.Auth.UseOAuth,
		PerformanceMode: "optimized",
		WorkerPoolSize:  10,
	}
	
	// System resources
	data.Resources = h.getSystemResources()
	
	// Recent logs (mock data)
	data.RecentLogs = []LogEntry{
		{
			Time:    time.Now().Add(-1 * time.Minute),
			Level:   "info",
			Message: "Export completed successfully",
			Context: "export_type=watched, records=1234",
		},
		{
			Time:    time.Now().Add(-5 * time.Minute),
			Level:   "info",
			Message: "Token refreshed automatically",
		},
		{
			Time:    time.Now().Add(-10 * time.Minute),
			Level:   "warning",
			Message: "API rate limit approaching",
			Context: "remaining=15, limit=1000",
		},
		{
			Time:    time.Now().Add(-15 * time.Minute),
			Level:   "info",
			Message: "Server health check passed",
		},
	}
	
	// Determine overall status
	data.OverallStatus, data.OverallMessage = h.calculateOverallStatus(data)
	
	return data
}

func (h *StatusHandler) formatUptime() string {
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

func (h *StatusHandler) formatTimeRemaining(expiry time.Time) string {
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

func (h *StatusHandler) maskClientID(clientID string) string {
	if len(clientID) < 8 {
		return clientID
	}
	return clientID[:4] + "***" + clientID[len(clientID)-4:]
}

func (h *StatusHandler) getSystemResources() *ResourcesData {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// Convert bytes to MB
	memoryUsed := m.Alloc / 1024 / 1024
	memoryTotal := m.Sys / 1024 / 1024
	memoryPercent := int((memoryUsed * 100) / memoryTotal)
	
	return &ResourcesData{
		MemoryUsed:    fmt.Sprintf("%d MB", memoryUsed),
		MemoryTotal:   fmt.Sprintf("%d MB", memoryTotal),
		MemoryPercent: memoryPercent,
		Goroutines:    runtime.NumGoroutine(),
		DiskUsed:      "2.3 GB",  // Mock data
		DiskTotal:     "10 GB",   // Mock data
		DiskPercent:   23,        // Mock data
		CacheEntries:  142,       // Mock data
	}
}

func (h *StatusHandler) calculateOverallStatus(data *StatusData) (string, string) {
	if !data.TokenStatus.IsValid {
		return "warning", "Authentication required for full functionality"
	}
	
	if data.APIStatus != "healthy" {
		return "error", "API connection issues detected"
	}
	
	if data.ServerStatus != "healthy" {
		return "error", "Server health issues detected"
	}
	
	return "healthy", "All systems operational"
}