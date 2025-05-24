package webui

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/webui/handlers"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/webui/middleware"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

//go:embed static/css/* static/js/* templates/*
var staticFiles embed.FS

// Server represents the web UI server
type Server struct {
	config     *config.Config
	logger     logger.Logger
	router     *mux.Router
	server     *http.Server
	upgrader   websocket.Upgrader
	monitoring interface{}
	templates  *template.Template
}

// NewServer creates a new web UI server
func NewServer(cfg *config.Config, log logger.Logger, mon interface{}) (*Server, error) {
	// Parse templates
	templates, err := template.ParseFS(staticFiles, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	s := &Server{
		config:     cfg,
		logger:     log,
		monitoring: mon,
		templates:  templates,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
	}

	s.setupRoutes()
	return s, nil
}

// setupRoutes configures all the routes for the web server
func (s *Server) setupRoutes() {
	s.router = mux.NewRouter()
	
	// Add middleware
	s.router.Use(middleware.Logger(s.logger))
	s.router.Use(middleware.CORS())
	s.router.Use(middleware.Security())

	// API routes
	api := s.router.PathPrefix("/api/v1").Subrouter()
	
	// Configuration endpoints
	configHandler := handlers.NewConfigHandler(s.config, s.logger)
	api.HandleFunc("/config", configHandler.GetConfig).Methods("GET")
	api.HandleFunc("/config", configHandler.UpdateConfig).Methods("POST")
	api.HandleFunc("/config/trakt/auth", configHandler.TraktAuth).Methods("POST")
	api.HandleFunc("/config/test", configHandler.TestConnection).Methods("POST")

	// Export endpoints
	exportHandler := handlers.NewExportHandler(s.config, s.logger)
	api.HandleFunc("/exports", exportHandler.ListExports).Methods("GET")
	api.HandleFunc("/exports", exportHandler.StartExport).Methods("POST")
	api.HandleFunc("/exports/{id}", exportHandler.GetExport).Methods("GET")
	api.HandleFunc("/exports/{id}", exportHandler.DeleteExport).Methods("DELETE")
	api.HandleFunc("/exports/{id}/download", exportHandler.DownloadExport).Methods("GET")

	// Monitoring endpoints
	monitoringHandler := handlers.NewMonitoringHandler(s.monitoring, s.logger)
	api.HandleFunc("/health", monitoringHandler.Health).Methods("GET")
	api.HandleFunc("/metrics", monitoringHandler.Metrics).Methods("GET")
	api.HandleFunc("/stats", monitoringHandler.Stats).Methods("GET")
	api.HandleFunc("/logs", monitoringHandler.Logs).Methods("GET")

	// WebSocket endpoint for real-time updates
	api.HandleFunc("/ws", s.handleWebSocket)

	// Static files
	s.router.PathPrefix("/static/").Handler(http.FileServer(http.FS(staticFiles)))

	// Frontend routes (SPA)
	uiHandler := handlers.NewUIHandler(s.templates, s.logger)
	s.router.HandleFunc("/", uiHandler.Dashboard).Methods("GET")
	s.router.HandleFunc("/config", uiHandler.Config).Methods("GET")
	s.router.HandleFunc("/exports", uiHandler.Exports).Methods("GET")
	s.router.HandleFunc("/monitoring", uiHandler.Monitoring).Methods("GET")
	s.router.HandleFunc("/logs", uiHandler.Logs).Methods("GET")
}

// Start starts the web server
func (s *Server) Start(port string) error {
	s.server = &http.Server{
		Addr:         ":" + port,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.logger.Info("webui.server_starting", map[string]interface{}{
		"port": port,
		"addr": s.server.Addr,
	})

	return s.server.ListenAndServe()
}

// Stop gracefully stops the web server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("webui.server_stopping", nil)
	return s.server.Shutdown(ctx)
}

// handleWebSocket handles WebSocket connections for real-time updates
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("webui.websocket_upgrade_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	defer conn.Close()

	s.logger.Info("webui.websocket_connected", map[string]interface{}{
		"remote_addr": r.RemoteAddr,
	})

	// Handle WebSocket messages
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.Error("webui.websocket_error", map[string]interface{}{
					"error": err.Error(),
				})
			}
			break
		}

		// Echo message back (for now)
		if err := conn.WriteMessage(messageType, message); err != nil {
			s.logger.Error("webui.websocket_write_error", map[string]interface{}{
				"error": err.Error(),
			})
			break
		}
	}
} 