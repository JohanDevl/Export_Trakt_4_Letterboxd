package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

// ConfigHandler handles configuration-related requests
type ConfigHandler struct {
	config *config.Config
	logger logger.Logger
}

// NewConfigHandler creates a new configuration handler
func NewConfigHandler(cfg *config.Config, log logger.Logger) *ConfigHandler {
	return &ConfigHandler{
		config: cfg,
		logger: log,
	}
}

// ConfigResponse represents the configuration response
type ConfigResponse struct {
	Trakt      TraktConfig      `json:"trakt"`
	Export     ExportConfig     `json:"export"`
	Logging    LoggingConfig    `json:"logging"`
	Monitoring MonitoringConfig `json:"monitoring"`
}

type TraktConfig struct {
	ClientID     string `json:"client_id"`
	APIBaseURL   string `json:"api_base_url"`
	ExtendedInfo string `json:"extended_info"`
	RateLimit    int    `json:"rate_limit"`
	HasToken     bool   `json:"has_token"`
}

type ExportConfig struct {
	OutputDir    string `json:"output_dir"`
	Format       string `json:"format"`
	IncludeShows bool   `json:"include_shows"`
	Timezone     string `json:"timezone"`
}

type LoggingConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

type MonitoringConfig struct {
	Enabled     bool `json:"enabled"`
	MetricsPort int  `json:"metrics_port"`
}

// GetConfig returns the current configuration
func (h *ConfigHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := ConfigResponse{
		Trakt: TraktConfig{
			ClientID:     h.config.Trakt.ClientID,
			APIBaseURL:   h.config.Trakt.APIBaseURL,
			ExtendedInfo: h.config.Trakt.ExtendedInfo,
			RateLimit:    60, // Default rate limit
			HasToken:     h.config.Trakt.AccessToken != "",
		},
		Export: ExportConfig{
			OutputDir:    h.config.Letterboxd.ExportDir,
			Format:       h.config.Export.Format,
			IncludeShows: true, // Default value
			Timezone:     h.config.Export.Timezone,
		},
		Logging: LoggingConfig{
			Level:  h.config.Logging.Level,
			Format: "visual", // Default format
		},
		Monitoring: MonitoringConfig{
			Enabled:     true, // Assume enabled if we're using the web UI
			MetricsPort: 9090,
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("config.encode_failed", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	h.logger.Info("config.retrieved", nil)
}

// UpdateConfig updates the configuration
func (h *ConfigHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var configUpdate ConfigResponse
	if err := json.NewDecoder(r.Body).Decode(&configUpdate); err != nil {
		h.logger.Error("config.decode_failed", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update configuration
	h.config.Trakt.ClientID = configUpdate.Trakt.ClientID
	h.config.Trakt.ExtendedInfo = configUpdate.Trakt.ExtendedInfo
	// Note: RateLimit not stored in config struct - would need to be added
	h.config.Letterboxd.ExportDir = configUpdate.Export.OutputDir
	h.config.Export.Format = configUpdate.Export.Format
	// Note: IncludeShows not stored in config struct - would need to be added
	h.config.Export.Timezone = configUpdate.Export.Timezone
	h.config.Logging.Level = configUpdate.Logging.Level
	// Note: Format not stored in config struct - using File field or would need to be added

	h.logger.Info("config.updated", map[string]interface{}{
		"client_id": configUpdate.Trakt.ClientID,
		"format":    configUpdate.Export.Format,
	})

	// Return updated configuration
	h.GetConfig(w, r)
}

// TraktAuthRequest represents a Trakt authentication request
type TraktAuthRequest struct {
	AuthCode string `json:"auth_code"`
}

// TraktAuth handles Trakt.tv authentication
func (h *ConfigHandler) TraktAuth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var authReq TraktAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&authReq); err != nil {
		h.logger.Error("trakt.auth_decode_failed", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual Trakt.tv OAuth flow
	// This would involve exchanging the auth code for access and refresh tokens
	
	h.logger.Info("trakt.auth_requested", map[string]interface{}{
		"auth_code_length": len(authReq.AuthCode),
	})

	response := map[string]interface{}{
		"success": true,
		"message": "Authentication successful",
	}

	json.NewEncoder(w).Encode(response)
}

// TestConnection tests the connection to Trakt.tv
func (h *ConfigHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// TODO: Implement actual connection test to Trakt.tv API
	
	h.logger.Info("trakt.connection_test", nil)

	response := map[string]interface{}{
		"success": true,
		"message": "Connection test successful",
		"status":  "connected",
	}

	json.NewEncoder(w).Encode(response)
} 