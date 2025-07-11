package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/auth"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/security/keyring"
)

func TestServerHealthEndpoint(t *testing.T) {
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

	// Test health endpoint by making a request to the server's handler
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	// Since the server handler is set up during setupRoutes, we need to call it
	server.handleHealth(rec, req)

	// Check response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Check content type
	expectedContentType := "application/json"
	if rec.Header().Get("Content-Type") != expectedContentType {
		t.Errorf("Expected content type %s, got %s", expectedContentType, rec.Header().Get("Content-Type"))
	}

	// Check that response contains expected JSON fields
	body := rec.Body.String()
	expectedFields := []string{"status", "service", "timestamp", "uptime", "version"}
	for _, field := range expectedFields {
		if !strings.Contains(body, field) {
			t.Errorf("Expected response to contain field '%s', got: %s", field, body)
		}
	}
}

func TestServerWebSocketEndpoints(t *testing.T) {
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

	// Test WebSocket status endpoint
	req := httptest.NewRequest("GET", "/ws/status", nil)
	rec := httptest.NewRecorder()
	server.handleWebSocket(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Errorf("Expected status %d for WebSocket status, got %d", http.StatusNotImplemented, rec.Code)
	}

	// Test WebSocket export endpoint
	req2 := httptest.NewRequest("GET", "/ws/export", nil)
	rec2 := httptest.NewRecorder()
	server.handleExportWebSocket(rec2, req2)

	if rec2.Code != http.StatusNotImplemented {
		t.Errorf("Expected status %d for WebSocket export, got %d", http.StatusNotImplemented, rec2.Code)
	}
}

func TestServerLegacyExport(t *testing.T) {
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

	// Test legacy export endpoint (should redirect due to no token)
	req := httptest.NewRequest("GET", "/export/watched", nil)
	rec := httptest.NewRecorder()
	server.handleLegacyExport(rec, req)

	// Should redirect to auth-url due to missing token
	if rec.Code != http.StatusSeeOther {
		t.Errorf("Expected status %d for legacy export without token, got %d", http.StatusSeeOther, rec.Code)
	}

	// Test with empty export type (should default to "watched")
	req2 := httptest.NewRequest("GET", "/export/", nil)
	rec2 := httptest.NewRecorder()
	server.handleLegacyExport(rec2, req2)

	if rec2.Code != http.StatusSeeOther {
		t.Errorf("Expected status %d for legacy export with empty type, got %d", http.StatusSeeOther, rec2.Code)
	}
}

func TestServerTemplateLoading(t *testing.T) {
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

	// Test that templates are loaded
	if server.templates == nil {
		t.Error("Templates should not be nil")
	}
}

func TestServerStop(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Auth: config.AuthConfig{
			CallbackPort: 0, // Use port 0 to get a random available port
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

	// Test stop function
	err = server.Stop(nil)
	if err != nil {
		t.Errorf("Stop should not return error, got: %v", err)
	}
}