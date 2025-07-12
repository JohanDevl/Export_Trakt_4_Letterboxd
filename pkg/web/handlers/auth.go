package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/auth"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

type AuthData struct {
	Title        string
	CurrentPage  string
	ServerStatus string
	LastUpdated  string
	AuthURL      string
	ClientID     string
	RedirectURI  string
	Alert        *AlertData
}

type AuthHandler struct {
	config       *config.Config
	logger       logger.Logger
	tokenManager *auth.TokenManager
	oauthManager *auth.OAuthManager
	templates    *template.Template
	currentState string
}

func NewAuthHandler(cfg *config.Config, log logger.Logger, tokenManager *auth.TokenManager, templates *template.Template) *AuthHandler {
	return &AuthHandler{
		config:       cfg,
		logger:       log,
		tokenManager: tokenManager,
		oauthManager: auth.NewOAuthManager(cfg, log),
		templates:    templates,
	}
}

func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/auth-url":
		h.handleAuthURL(w, r)
	case "/callback":
		h.handleCallback(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *AuthHandler) handleAuthURL(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("web.auth_url_request_started", map[string]interface{}{
		"client_ip": r.RemoteAddr,
		"user_agent": r.UserAgent(),
	})
	
	// Generate fresh auth URL and state
	authURL, state, err := h.oauthManager.GenerateAuthURL()
	if err != nil {
		h.logger.Error("web.auth_url_generation_failed", map[string]interface{}{
			"error": err.Error(),
		})
		
		data := &AuthData{
			Title:        "Authentication Error",
			CurrentPage:  "auth",
			ServerStatus: "error",
			Alert: &AlertData{
				Type:    "error",
				Icon:    "❌",
				Message: "Failed to generate authentication URL: " + err.Error(),
			},
		}
		
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		if h.templates == nil {
			h.logger.Error("web.templates_nil", map[string]interface{}{
				"template": "auth-error.html",
			})
			w.Write([]byte("Template system not initialized"))
			return
		}
		if err := h.templates.ExecuteTemplate(w, "auth-error.html", data); err != nil {
			h.logger.Error("web.template_error", map[string]interface{}{
				"error":    err.Error(),
				"template": "auth-error.html",
			})
			w.Write([]byte("Internal Server Error"))
		}
		return
	}
	
	h.currentState = state
	
	data := &AuthData{
		Title:        "OAuth Authentication",
		CurrentPage:  "auth",
		ServerStatus: "healthy",
		LastUpdated:  time.Now().Format("2006-01-02 15:04:05"),
		AuthURL:      authURL,
		ClientID:     h.config.Trakt.ClientID,
		RedirectURI:  h.config.Auth.RedirectURI,
	}
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	// Execute standalone template directly (thread-safe)
	if err := h.templates.ExecuteTemplate(w, "auth-url.html", data); err != nil {
		h.logger.Error("web.template_error", map[string]interface{}{
			"error":    err.Error(),
			"template": "auth-url.html",
		})
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	h.logger.Info("web.auth_url_request_completed", map[string]interface{}{
		"client_ip": r.RemoteAddr,
		"auth_url_length": len(authURL),
	})
}

func (h *AuthHandler) handleCallback(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("web.oauth_callback_received", map[string]interface{}{
		"client_ip": r.RemoteAddr,
	})
	
	code := r.URL.Query().Get("code")
	errorParam := r.URL.Query().Get("error")
	receivedState := r.URL.Query().Get("state")
	
	// Handle OAuth error
	if errorParam != "" {
		errDescription := r.URL.Query().Get("error_description")
		h.logger.Error("web.oauth_error", map[string]interface{}{
			"error":       errorParam,
			"description": errDescription,
		})
		
		data := &AuthData{
			Title:        "Authentication Error",
			CurrentPage:  "auth",
			ServerStatus: "error",
			LastUpdated:  time.Now().Format("2006-01-02 15:04:05"),
			Alert: &AlertData{
				Type:    "error",
				Icon:    "❌",
				Message: fmt.Sprintf("OAuth Error: %s - %s", errorParam, errDescription),
			},
		}
		
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		if h.templates == nil {
			h.logger.Error("web.templates_nil", map[string]interface{}{
				"template": "auth-error.html",
			})
			w.Write([]byte("Template system not initialized"))
			return
		}
		if err := h.templates.ExecuteTemplate(w, "auth-error.html", data); err != nil {
			h.logger.Error("web.template_error", map[string]interface{}{
				"error":    err.Error(),
				"template": "auth-error.html",
			})
			w.Write([]byte("Internal Server Error"))
		}
		return
	}
	
	// Check for authorization code
	if code == "" {
		h.logger.Error("web.no_auth_code", nil)
		
		data := &AuthData{
			Title:        "Authentication Error",
			CurrentPage:  "auth",
			ServerStatus: "error",
			LastUpdated:  time.Now().Format("2006-01-02 15:04:05"),
			Alert: &AlertData{
				Type:    "error",
				Icon:    "❌",
				Message: "No authorization code received from Trakt.tv",
			},
		}
		
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		if h.templates == nil {
			h.logger.Error("web.templates_nil", map[string]interface{}{
				"template": "auth-error.html",
			})
			w.Write([]byte("Template system not initialized"))
			return
		}
		if err := h.templates.ExecuteTemplate(w, "auth-error.html", data); err != nil {
			h.logger.Error("web.template_error", map[string]interface{}{
				"error":    err.Error(),
				"template": "auth-error.html",
			})
			w.Write([]byte("Internal Server Error"))
		}
		return
	}
	
	// Exchange code for token
	token, err := h.oauthManager.ExchangeCodeForToken(code, h.currentState, receivedState)
	if err != nil {
		h.logger.Error("web.token_exchange_failed", map[string]interface{}{
			"error": err.Error(),
		})
		
		data := &AuthData{
			Title:        "Authentication Failed",
			CurrentPage:  "auth",
			ServerStatus: "error",
			LastUpdated:  time.Now().Format("2006-01-02 15:04:05"),
			Alert: &AlertData{
				Type:    "error",
				Icon:    "❌",
				Message: "Failed to exchange authorization code for token: " + err.Error(),
			},
		}
		
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		if h.templates == nil {
			h.logger.Error("web.templates_nil", map[string]interface{}{
				"template": "auth-error.html",
			})
			w.Write([]byte("Template system not initialized"))
			return
		}
		if err := h.templates.ExecuteTemplate(w, "auth-error.html", data); err != nil {
			h.logger.Error("web.template_error", map[string]interface{}{
				"error":    err.Error(),
				"template": "auth-error.html",
			})
			w.Write([]byte("Internal Server Error"))
		}
		return
	}
	
	// Store the token
	if err := h.tokenManager.StoreToken(token); err != nil {
		h.logger.Error("web.token_store_failed", map[string]interface{}{
			"error": err.Error(),
		})
		
		data := &AuthData{
			Title:        "Token Storage Failed",
			CurrentPage:  "auth",
			ServerStatus: "warning",
			LastUpdated:  time.Now().Format("2006-01-02 15:04:05"),
			Alert: &AlertData{
				Type:    "warning",
				Icon:    "⚠️",
				Message: "Authentication succeeded but failed to store token: " + err.Error(),
			},
		}
		
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		if h.templates == nil {
			h.logger.Error("web.templates_nil", map[string]interface{}{
				"template": "auth-error.html",
			})
			w.Write([]byte("Template system not initialized"))
			return
		}
		if err := h.templates.ExecuteTemplate(w, "auth-error.html", data); err != nil {
			h.logger.Error("web.template_error", map[string]interface{}{
				"error":    err.Error(),
				"template": "auth-error.html",
			})
			w.Write([]byte("Internal Server Error"))
		}
		return
	}
	
	h.logger.Info("web.oauth_success", map[string]interface{}{
		"expires_at": h.oauthManager.GetTokenExpiryTime(token).Format("2006-01-02 15:04:05"),
	})
	
	data := &AuthData{
		Title:        "Authentication Successful",
		CurrentPage:  "auth",
		ServerStatus: "healthy",
		LastUpdated:  time.Now().Format("2006-01-02 15:04:05"),
		Alert: &AlertData{
			Type:    "success",
			Icon:    "✅",
			Message: "Successfully authenticated with Trakt.tv!",
		},
	}
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "auth-success.html", data); err != nil {
		h.logger.Error("web.template_error", map[string]interface{}{
			"error":    err.Error(),
			"template": "auth-success.html",
		})
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}