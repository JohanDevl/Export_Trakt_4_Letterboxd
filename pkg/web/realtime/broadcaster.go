package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/auth"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

// StatusBroadcaster manages automatic status updates
type StatusBroadcaster struct {
	hub          *Hub
	config       *config.Config
	logger       logger.Logger
	tokenManager *auth.TokenManager
	isRunning    bool
	stopChannel  chan struct{}
}

// NewStatusBroadcaster creates a new status broadcaster
func NewStatusBroadcaster(hub *Hub, cfg *config.Config, logger logger.Logger, tokenManager *auth.TokenManager) *StatusBroadcaster {
	return &StatusBroadcaster{
		hub:          hub,
		config:       cfg,
		logger:       logger,
		tokenManager: tokenManager,
		stopChannel:  make(chan struct{}),
	}
}

// Start begins the status broadcasting routine
func (sb *StatusBroadcaster) Start() {
	if sb.isRunning {
		return
	}
	
	sb.isRunning = true
	sb.logger.Info("realtime.status_broadcaster_starting", nil)
	
	go sb.broadcastLoop()
}

// Stop halts the status broadcasting routine
func (sb *StatusBroadcaster) Stop() {
	if !sb.isRunning {
		return
	}
	
	sb.isRunning = false
	close(sb.stopChannel)
	sb.logger.Info("realtime.status_broadcaster_stopped", nil)
}

// broadcastLoop runs the main broadcasting loop
func (sb *StatusBroadcaster) broadcastLoop() {
	// Initial broadcast
	sb.broadcastStatus()
	
	// Set up tickers
	statusTicker := time.NewTicker(30 * time.Second)      // Status updates every 30s
	healthTicker := time.NewTicker(60 * time.Second)      // Health check every 60s
	exportTicker := time.NewTicker(5 * time.Second)       // Export updates every 5s
	
	defer statusTicker.Stop()
	defer healthTicker.Stop()
	defer exportTicker.Stop()
	
	for {
		select {
		case <-sb.stopChannel:
			return
			
		case <-statusTicker.C:
			sb.broadcastStatus()
			
		case <-healthTicker.C:
			sb.broadcastHealth()
			
		case <-exportTicker.C:
			sb.checkAndBroadcastExportProgress()
		}
	}
}

// broadcastStatus sends current status information
func (sb *StatusBroadcaster) broadcastStatus() {
	if sb.hub.GetClientCount() == 0 {
		return // No clients to broadcast to
	}
	
	status := sb.gatherStatusData()
	sb.hub.BroadcastMessage(StatusUpdate, status)
	
	sb.logger.Debug("realtime.status_broadcast", map[string]interface{}{
		"clients": sb.hub.GetClientCount(),
	})
}

// broadcastHealth sends server health information
func (sb *StatusBroadcaster) broadcastHealth() {
	if sb.hub.GetClientCount() == 0 {
		return
	}
	
	health := sb.gatherHealthData()
	sb.hub.BroadcastMessage(ServerHealth, health)
}

// gatherStatusData collects current status information
func (sb *StatusBroadcaster) gatherStatusData() map[string]interface{} {
	// Get token status
	tokenStatus := map[string]interface{}{
		"isValid": false,
		"hasToken": false,
		"expiresAt": "",
	}
	
	if sb.tokenManager != nil {
		if status, err := sb.tokenManager.GetTokenStatus(); err == nil {
			tokenStatus["isValid"] = status.IsValid
			tokenStatus["hasToken"] = status.HasToken
			if !status.ExpiresAt.IsZero() {
				tokenStatus["expiresAt"] = status.ExpiresAt.Format(time.RFC3339)
			}
		}
	}
	
	// Get server status
	serverStatus := "healthy"
	apiStatus := "unknown"
	
	// Check if export directory exists and is writable
	exportDir := sb.config.Letterboxd.ExportDir
	if exportDir == "" {
		exportDir = "./exports"
	}
	
	if _, err := os.Stat(exportDir); os.IsNotExist(err) {
		serverStatus = "warning"
	}
	
	// Simple API status check based on token validity
	if tokenStatus["isValid"].(bool) {
		apiStatus = "healthy"
	} else {
		apiStatus = "error"
	}
	
	return map[string]interface{}{
		"serverStatus": serverStatus,
		"tokenStatus":  tokenStatus,
		"apiStatus":    apiStatus,
		"timestamp":    time.Now().Format(time.RFC3339),
		"uptime":       time.Since(time.Now().Add(-time.Hour)).String(), // Placeholder uptime
	}
}

// gatherHealthData collects server health information
func (sb *StatusBroadcaster) gatherHealthData() map[string]interface{} {
	hubStats := sb.hub.GetStats()
	
	return map[string]interface{}{
		"status":     "healthy",
		"timestamp":  time.Now().Format(time.RFC3339),
		"clients": map[string]interface{}{
			"total":     hubStats.TotalClients,
			"websocket": hubStats.WebSocketClients,
			"sse":       hubStats.SSEClients,
		},
		"messages": map[string]interface{}{
			"total":      hubStats.MessagesTotal,
			"bytes_sent": hubStats.BytesSent,
		},
		"uptime": time.Since(hubStats.StartTime).String(),
	}
}

// checkAndBroadcastExportProgress checks for ongoing exports and broadcasts progress
func (sb *StatusBroadcaster) checkAndBroadcastExportProgress() {
	if sb.hub.GetClientCount() == 0 {
		return
	}
	
	// Check for export progress files or status
	exportDir := sb.config.Letterboxd.ExportDir
	if exportDir == "" {
		exportDir = "./exports"
	}
	
	// Look for .progress files that might indicate ongoing exports
	progressFiles, err := filepath.Glob(filepath.Join(exportDir, "*.progress"))
	if err != nil {
		return
	}
	
	for _, progressFile := range progressFiles {
		progress := sb.readExportProgress(progressFile)
		if progress != nil {
			sb.hub.BroadcastMessage(ExportProgress, progress)
		}
	}
}

// readExportProgress reads export progress from a progress file
func (sb *StatusBroadcaster) readExportProgress(progressFile string) map[string]interface{} {
	data, err := os.ReadFile(progressFile)
	if err != nil {
		return nil
	}
	
	var progress map[string]interface{}
	if err := json.Unmarshal(data, &progress); err != nil {
		return nil
	}
	
	// Add timestamp if not present
	if _, exists := progress["timestamp"]; !exists {
		progress["timestamp"] = time.Now().Format(time.RFC3339)
	}
	
	return progress
}

// BroadcastExportStart notifies clients that an export has started
func (sb *StatusBroadcaster) BroadcastExportStart(exportType string) {
	sb.hub.BroadcastMessage(ExportProgress, map[string]interface{}{
		"status":      "started",
		"exportType":  exportType,
		"progress":    0,
		"message":     fmt.Sprintf("Starting %s export...", exportType),
		"timestamp":   time.Now().Format(time.RFC3339),
	})
}

// BroadcastExportProgress notifies clients of export progress
func (sb *StatusBroadcaster) BroadcastExportProgress(exportType string, progress int, message string) {
	sb.hub.BroadcastMessage(ExportProgress, map[string]interface{}{
		"status":      "progress",
		"exportType":  exportType,
		"progress":    progress,
		"message":     message,
		"timestamp":   time.Now().Format(time.RFC3339),
	})
}

// BroadcastExportComplete notifies clients that an export has completed
func (sb *StatusBroadcaster) BroadcastExportComplete(exportType string, success bool, message string, filePath string) {
	status := "completed"
	if !success {
		status = "failed"
	}
	
	payload := map[string]interface{}{
		"status":     status,
		"exportType": exportType,
		"progress":   100,
		"message":    message,
		"timestamp":  time.Now().Format(time.RFC3339),
	}
	
	if success && filePath != "" {
		payload["filePath"] = filePath
		payload["fileName"] = filepath.Base(filePath)
	}
	
	sb.hub.BroadcastMessage(ExportProgress, payload)
}

// BroadcastAlert sends an alert message to all clients
func (sb *StatusBroadcaster) BroadcastAlert(alertType, message string) {
	sb.hub.BroadcastMessage(Alert, map[string]interface{}{
		"type":      alertType,
		"message":   message,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// BroadcastLogEntry sends a log entry to clients (for real-time log viewing)
func (sb *StatusBroadcaster) BroadcastLogEntry(level, message string, context map[string]interface{}) {
	sb.hub.BroadcastMessage(LogEntry, map[string]interface{}{
		"level":     level,
		"message":   message,
		"context":   context,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// TokenUpdated notifies clients when token status changes
func (sb *StatusBroadcaster) TokenUpdated() {
	if sb.tokenManager != nil {
		if status, err := sb.tokenManager.GetTokenStatus(); err == nil {
			sb.hub.BroadcastMessage(TokenUpdate, map[string]interface{}{
				"isValid":   status.IsValid,
				"hasToken":  status.HasToken,
				"expiresAt": status.ExpiresAt.Format(time.RFC3339),
				"timestamp": time.Now().Format(time.RFC3339),
			})
		}
	}
}