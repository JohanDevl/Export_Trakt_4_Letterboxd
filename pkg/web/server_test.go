package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/auth"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/security/keyring"
)

func TestNewServer(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Auth: config.AuthConfig{
			CallbackPort: 8080,
		},
		Letterboxd: config.LetterboxdConfig{
			ExportDir: "./test_exports",
		},
	}

	// Create test logger
	log := logger.NewLogger()

	// Create test token manager
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	if err != nil {
		t.Fatalf("Failed to create keyring manager: %v", err)
	}
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)

	// Create server
	server, err := NewServer(cfg, log, tokenManager)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if server == nil {
		t.Fatal("Server should not be nil")
	}

	// Test that server has required fields
	if server.config == nil {
		t.Error("Server config should not be nil")
	}

	if server.logger == nil {
		t.Error("Server logger should not be nil")
	}

	if server.tokenManager == nil {
		t.Error("Server tokenManager should not be nil")
	}
}

func TestServerGetters(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Auth: config.AuthConfig{
			CallbackPort: 8080,
		},
		Letterboxd: config.LetterboxdConfig{
			ExportDir: "./test_exports",
		},
	}

	// Create test logger
	log := logger.NewLogger()

	// Create test token manager
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	if err != nil {
		t.Fatalf("Failed to create keyring manager: %v", err)
	}
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)

	// Create server
	server, err := NewServer(cfg, log, tokenManager)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test GetAddr
	addr := server.GetAddr()
	expectedAddr := ":8080"
	if addr != expectedAddr {
		t.Errorf("Expected address %s, got %s", expectedAddr, addr)
	}

	// Test GetStartTime
	startTime := server.GetStartTime()
	if startTime.IsZero() {
		t.Error("Start time should not be zero")
	}
}

func TestServerRoutes(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Auth: config.AuthConfig{
			CallbackPort: 8080,
		},
		Letterboxd: config.LetterboxdConfig{
			ExportDir: "./test_exports",
		},
	}

	// Create test logger
	log := logger.NewLogger()

	// Create test token manager
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	if err != nil {
		t.Fatalf("Failed to create keyring manager: %v", err)
	}
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)

	// Create server
	server, err := NewServer(cfg, log, tokenManager)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Test that server has HTTP server configured
	if server.server == nil {
		t.Error("HTTP server should not be nil")
	}

	if server.server.Handler == nil {
		t.Error("HTTP server handler should not be nil")
	}
}

func TestTemplateData(t *testing.T) {
	data := TemplateData{
		Title:        "Test Title",
		CurrentPage:  "test",
		ServerStatus: "healthy",
		LastUpdated:  "2025-07-11 12:00:00",
	}

	if data.Title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got '%s'", data.Title)
	}

	if data.CurrentPage != "test" {
		t.Errorf("Expected current page 'test', got '%s'", data.CurrentPage)
	}

	if data.ServerStatus != "healthy" {
		t.Errorf("Expected server status 'healthy', got '%s'", data.ServerStatus)
	}
}

func TestServerMiddleware(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Auth: config.AuthConfig{
			CallbackPort: 8080,
		},
		Letterboxd: config.LetterboxdConfig{
			ExportDir: "./test_exports",
		},
	}

	// Create test logger
	log := logger.NewLogger()

	// Create test token manager
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	if err != nil {
		t.Fatalf("Failed to create keyring manager: %v", err)
	}
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)

	// Create server
	server, err := NewServer(cfg, log, tokenManager)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Test logging middleware
	loggedHandler := server.withLogging(testHandler)
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	loggedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Test CORS middleware
	corsHandler := server.withCORS(testHandler)

	// Test API request with CORS
	apiReq := httptest.NewRequest("GET", "/api/test", nil)
	apiRec := httptest.NewRecorder()
	corsHandler.ServeHTTP(apiRec, apiReq)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Test OPTIONS request
	optionsReq := httptest.NewRequest("OPTIONS", "/api/test", nil)
	optionsRec := httptest.NewRecorder()
	corsHandler.ServeHTTP(optionsRec, optionsReq)

	if optionsRec.Code != http.StatusOK {
		t.Errorf("Expected status %d for OPTIONS, got %d", http.StatusOK, optionsRec.Code)
	}
}