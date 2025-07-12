package handlers

import (
	"fmt"
	"html/template"
	"net/http"

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
		tmpl, cloneErr := h.templates.Clone()
		if cloneErr != nil {
			h.logger.Error("web.template_clone_error", map[string]interface{}{
				"error": cloneErr.Error(),
			})
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if _, parseErr := tmpl.ParseFiles("web/templates/auth-error.html"); parseErr != nil {
			h.logger.Error("web.template_parse_error", map[string]interface{}{
				"error":    parseErr.Error(),
				"template": "auth-error.html",
			})
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		tmpl.ExecuteTemplate(w, "base.html", data)
		return
	}
	
	h.currentState = state
	
	data := &AuthData{
		Title:        "OAuth Authentication",
		CurrentPage:  "auth",
		ServerStatus: "healthy",
		AuthURL:      authURL,
		ClientID:     h.config.Trakt.ClientID,
		RedirectURI:  h.config.Auth.RedirectURI,
	}
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	// Parse and execute base template with auth-url content
	tmpl, err := h.templates.Clone()
	if err != nil {
		h.logger.Error("web.template_clone_error", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	// Parse the auth-url template to associate it with base template
	if _, err := tmpl.ParseFiles("web/templates/auth-url.html"); err != nil {
		h.logger.Error("web.template_parse_error", map[string]interface{}{
			"error":    err.Error(),
			"template": "auth-url.html",
		})
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
		h.logger.Error("web.template_error", map[string]interface{}{
			"error":    err.Error(),
			"template": "base.html",
		})
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
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
			Alert: &AlertData{
				Type:    "error",
				Icon:    "❌",
				Message: fmt.Sprintf("OAuth Error: %s - %s", errorParam, errDescription),
			},
		}
		
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		tmpl, _ := h.templates.Clone()
		tmpl.ParseFiles("web/templates/auth-error.html")
		tmpl.ExecuteTemplate(w, "base.html", data)
		return
	}
	
	// Check for authorization code
	if code == "" {
		h.logger.Error("web.no_auth_code", nil)
		
		data := &AuthData{
			Title:        "Authentication Error",
			CurrentPage:  "auth",
			ServerStatus: "error",
			Alert: &AlertData{
				Type:    "error",
				Icon:    "❌",
				Message: "No authorization code received from Trakt.tv",
			},
		}
		
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		tmpl, _ := h.templates.Clone()
		tmpl.ParseFiles("web/templates/auth-error.html")
		tmpl.ExecuteTemplate(w, "base.html", data)
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
			Alert: &AlertData{
				Type:    "error",
				Icon:    "❌",
				Message: "Failed to exchange authorization code for token: " + err.Error(),
			},
		}
		
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		tmpl, _ := h.templates.Clone()
		tmpl.ParseFiles("web/templates/auth-error.html")
		tmpl.ExecuteTemplate(w, "base.html", data)
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
			Alert: &AlertData{
				Type:    "warning",
				Icon:    "⚠️",
				Message: "Authentication succeeded but failed to store token: " + err.Error(),
			},
		}
		
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		tmpl, _ := h.templates.Clone()
		tmpl.ParseFiles("web/templates/auth-error.html")
		tmpl.ExecuteTemplate(w, "base.html", data)
		return
	}
	
	h.logger.Info("web.oauth_success", map[string]interface{}{
		"expires_at": h.oauthManager.GetTokenExpiryTime(token).Format("2006-01-02 15:04:05"),
	})
	
	data := &AuthData{
		Title:        "Authentication Successful",
		CurrentPage:  "auth",
		ServerStatus: "healthy",
		Alert: &AlertData{
			Type:    "success",
			Icon:    "✅",
			Message: "Successfully authenticated with Trakt.tv!",
		},
	}
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl, _ := h.templates.Clone()
	tmpl.ParseFiles("web/templates/auth-success.html")
	if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
		h.logger.Error("web.template_error", map[string]interface{}{
			"error":    err.Error(),
			"template": "base.html",
		})
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}