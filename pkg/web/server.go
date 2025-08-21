package web

import (
	"context"
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
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/web/handlers"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/web/middleware"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/web/realtime"
)

type Server struct {
	config             *config.Config
	logger             logger.Logger
	tokenManager       *auth.TokenManager
	templates          *template.Template
	server             *http.Server
	startTime          time.Time
	csrfMiddleware     *middleware.CSRFMiddleware
	securityMiddleware *middleware.SecurityHeaders
	
	// Real-time components
	realtimeHub        *realtime.Hub
	statusBroadcaster  *realtime.StatusBroadcaster
	websocketHandler   *realtime.SimpleWebSocketHandler
	sseHandler         *realtime.SSEHandler
}

type TemplateData struct {
	Title        string
	CurrentPage  string
	ServerStatus string
	LastUpdated  string
	CSRFToken    string
}

func NewServer(cfg *config.Config, log logger.Logger, tokenManager *auth.TokenManager) (*Server, error) {
	// Determine if we should use secure cookies (HTTPS)
	secureCookies := cfg.Security.RequireHTTPS
	
	// Initialize real-time components
	realtimeHub := realtime.NewHub(log)
	statusBroadcaster := realtime.NewStatusBroadcaster(realtimeHub, cfg, log, tokenManager)
	websocketHandler := realtime.NewSimpleWebSocketHandler(realtimeHub, log)
	sseHandler := realtime.NewSSEHandler(realtimeHub, log)
	
	s := &Server{
		config:             cfg,
		logger:             log,
		tokenManager:       tokenManager,
		startTime:          time.Now(),
		csrfMiddleware:     middleware.NewCSRFMiddleware(log, secureCookies),
		securityMiddleware: middleware.NewSecurityHeaders(log, secureCookies, true),
		realtimeHub:        realtimeHub,
		statusBroadcaster:  statusBroadcaster,
		websocketHandler:   websocketHandler,
		sseHandler:         sseHandler,
	}
	
	// Load templates
	if err := s.loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}
	
	// Setup HTTP server
	s.setupRoutes()
	
	return s, nil
}

func (s *Server) loadTemplates() error {
	// Define template functions
	funcMap := template.FuncMap{
		"title": func(str string) string {
			if len(str) == 0 {
				return str
			}
			return strings.ToUpper(str[:1]) + strings.ToLower(str[1:])
		},
		"upper":    strings.ToUpper,
		"lower":    strings.ToLower,
		"filename": filepath.Base,
		"mask": func(str string) string {
			if len(str) < 8 {
				return str
			}
			return str[:4] + "***" + str[len(str)-4:]
		},
		"eq": func(a, b interface{}) bool {
			return a == b
		},
		"contains": func(str, substr string) bool {
			return strings.Contains(str, substr)
		},
		"substr": func(str string, start int) string {
			if start >= len(str) {
				return ""
			}
			return str[start:]
		},
		"len": func(v interface{}) int {
			switch val := v.(type) {
			case []string:
				return len(val)
			case string:
				return len(val)
			default:
				return 0
			}
		},
		"gt": func(a, b int) bool {
			return a > b
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
	}
	
	// Find template directory
	templateDir := "./web/templates"
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		// Try alternative paths
		alternatives := []string{
			"../web/templates",
			"../../web/templates",
			"./templates",
		}
		
		for _, alt := range alternatives {
			if _, err := os.Stat(alt); err == nil {
				templateDir = alt
				break
			}
		}
	}
	
	// Load all template files
	templates := template.New("").Funcs(funcMap)
	
	err := filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && strings.HasSuffix(path, ".html") {
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read template %s: %w", path, err)
			}
			
			// Use relative path as template name
			relPath, err := filepath.Rel(templateDir, path)
			if err != nil {
				relPath = filepath.Base(path)
			}
			
			_, err = templates.New(relPath).Parse(string(content))
			if err != nil {
				return fmt.Errorf("failed to parse template %s: %w", relPath, err)
			}
		}
		
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}
	
	s.templates = templates
	s.logger.Info("web.templates_loaded", map[string]interface{}{
		"template_dir": templateDir,
	})
	
	return nil
}

func (s *Server) setupRoutes() {
	mux := http.NewServeMux()
	
	// Static files
	staticDir := "./web/static"
	if _, err := os.Stat(staticDir); err == nil {
		mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	}
	
	// Create handlers
	dashboardHandler := handlers.NewDashboardHandler(s.config, s.logger, s.tokenManager, s.templates)
	exportsHandler := handlers.NewExportsHandler(s.config, s.logger, s.tokenManager, s.templates, s.csrfMiddleware)
	statusHandler := handlers.NewStatusHandler(s.config, s.logger, s.tokenManager, s.templates)
	authHandler := handlers.NewAuthHandler(s.config, s.logger, s.tokenManager, s.templates)
	
	// Download handler for export files
	exportsDir := "./exports"
	if s.config.Letterboxd.ExportDir != "" {
		exportsDir = s.config.Letterboxd.ExportDir
	}
	downloadHandler := handlers.NewDownloadHandler(exportsDir, s.logger)
	
	// Register routes
	mux.Handle("/", dashboardHandler)
	mux.Handle("/exports", exportsHandler)
	mux.Handle("/api/export", exportsHandler)
	mux.Handle("/api/export/", exportsHandler)
	mux.Handle("/status", statusHandler)
	mux.Handle("/api/status", statusHandler)
	mux.Handle("/api/test-connection", statusHandler)
	mux.Handle("/api/logs/recent", statusHandler)
	mux.Handle("/api/logs/download", statusHandler)
	mux.Handle("/auth-url", authHandler)
	mux.Handle("/callback", authHandler)
	mux.Handle("/download/", downloadHandler)
	
	// Config page
	mux.HandleFunc("/config", s.handleConfig)
	
	// Legacy export endpoints for compatibility
	mux.HandleFunc("/export/", s.handleLegacyExport)
	
	// Health check endpoint
	mux.HandleFunc("/health", s.handleHealth)
	
	// Real-time endpoints
	mux.HandleFunc("/ws/status", s.websocketHandler.HandleWebSocket)
	mux.HandleFunc("/ws/export", s.websocketHandler.HandleWebSocket)
	mux.HandleFunc("/sse/status", s.sseHandler.HandleSSEStatus)
	mux.HandleFunc("/sse/export", s.sseHandler.HandleSSEExports)
	mux.HandleFunc("/sse/all", s.sseHandler.HandleSSE)
	
	// Add middleware (order matters: security headers first, then CSRF, then CORS, then logging)
	handler := s.withLogging(s.withCORS(s.csrfMiddleware.Middleware(s.securityMiddleware.Middleware(mux))))
	
	port := s.config.Auth.CallbackPort
	if port == 0 {
		port = 8080
	}
	
	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

func (s *Server) handleLegacyExport(w http.ResponseWriter, r *http.Request) {
	exportType := strings.TrimPrefix(r.URL.Path, "/export/")
	if exportType == "" {
		exportType = "watched"
	}
	
	// Check authentication
	status, err := s.tokenManager.GetTokenStatus()
	if err != nil || !status.HasToken || !status.IsValid {
		http.Redirect(w, r, "/auth-url", http.StatusSeeOther)
		return
	}
	
	s.logger.Info("web.legacy_export_triggered", map[string]interface{}{
		"export_type": exportType,
		"client_ip":   r.RemoteAddr,
	})
	
	// Redirect to modern exports page with success message
	http.Redirect(w, r, fmt.Sprintf("/exports?triggered=%s", exportType), http.StatusSeeOther)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"service":   "export-trakt-web",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    time.Since(s.startTime).String(),
		"version":   "1.0.0",
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	// Simple JSON encoding without external dependencies
	fmt.Fprintf(w, `{
		"status": "%s",
		"service": "%s",
		"timestamp": "%s",
		"uptime": "%s",
		"version": "%s"
	}`, health["status"], health["service"], health["timestamp"], health["uptime"], health["version"])
}

// GetStatusBroadcaster returns the status broadcaster for external use
func (s *Server) GetStatusBroadcaster() *realtime.StatusBroadcaster {
	return s.statusBroadcaster
}

// GetRealtimeHub returns the realtime hub for external use
func (s *Server) GetRealtimeHub() *realtime.Hub {
	return s.realtimeHub
}

func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Skip logging for static files and health checks to reduce noise
		if !strings.HasPrefix(r.URL.Path, "/static/") && r.URL.Path != "/health" {
			defer func() {
				s.logger.Info("web.request", map[string]interface{}{
					"method":      r.Method,
					"path":        r.URL.Path,
					"remote_addr": r.RemoteAddr,
					"user_agent":  r.UserAgent(),
					"duration":    time.Since(start).String(),
				})
			}()
		}
		
		next.ServeHTTP(w, r)
	})
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers for API endpoints
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/ws/") {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

func (s *Server) Start() error {
	s.logger.Info("web.server_starting", map[string]interface{}{
		"addr":       s.server.Addr,
		"start_time": s.startTime.Format(time.RFC3339),
	})
	
	// Start real-time components
	go s.realtimeHub.Start()
	s.statusBroadcaster.Start()
	
	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("web.server_stopping", nil)
	
	// Stop real-time components
	s.statusBroadcaster.Stop()
	
	return s.server.Shutdown(ctx)
}

func (s *Server) GetAddr() string {
	return s.server.Addr
}

func (s *Server) GetStartTime() time.Time {
	return s.startTime
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title             string
		CurrentPage       string
		ServerStatus      string
		LastUpdated       string
		ExportDir         string
		LogLevel          string
		ServerPort        int
		UseOAuth          bool
		AutoRefresh       bool
		RedirectURI       string
		ClientID          string
		DateFormat        string
		Timezone          string
		HistoryMode       string
		ExtendedInfo      string
		EncryptionEnabled bool
		KeyringBackend    string
		AuditLogging      bool
		RateLimitEnabled  bool
	}{
		Title:             "Configuration",
		CurrentPage:       "config",
		ServerStatus:      "healthy",
		LastUpdated:       time.Now().Format("2006-01-02 15:04:05"),
		ExportDir:         s.config.Letterboxd.ExportDir,
		LogLevel:          s.config.Logging.Level,
		ServerPort:        s.config.Auth.CallbackPort,
		UseOAuth:          s.config.Auth.UseOAuth,
		AutoRefresh:       s.config.Auth.AutoRefresh,
		RedirectURI:       s.config.Auth.RedirectURI,
		ClientID:          s.config.Trakt.ClientID,
		DateFormat:        s.config.Export.DateFormat,
		Timezone:          s.config.Export.Timezone,
		HistoryMode:       s.config.Export.HistoryMode,
		ExtendedInfo:      s.config.Trakt.ExtendedInfo,
		EncryptionEnabled: s.config.Security.EncryptionEnabled,
		KeyringBackend:    s.config.Security.KeyringBackend,
		AuditLogging:      s.config.Security.AuditLogging,
		RateLimitEnabled:  s.config.Security.RateLimitEnabled,
	}
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "config.html", data); err != nil {
		s.logger.Error("web.template_error", map[string]interface{}{
			"error":    err.Error(),
			"template": "config.html",
		})
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}