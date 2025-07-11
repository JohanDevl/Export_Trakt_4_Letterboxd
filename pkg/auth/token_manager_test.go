package auth

import (
	"testing"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/security/keyring"
)

func TestNewTokenManager(t *testing.T) {
	cfg := &config.Config{
		Trakt: config.TraktConfig{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
		},
		Auth: config.AuthConfig{
			RedirectURI: "http://localhost:8080/callback",
		},
	}
	
	log := logger.NewLogger()
	keyringMgr, _ := keyring.NewManager(keyring.MemoryBackend)
	
	tm := NewTokenManager(cfg, log, keyringMgr)
	if tm == nil {
		t.Fatal("Expected token manager to be created, got nil")
	}
	
	if tm.config != cfg {
		t.Error("Expected config to be set")
	}
	if tm.logger != log {
		t.Error("Expected logger to be set")
	}
	if tm.keyringMgr != keyringMgr {
		t.Error("Expected keyring manager to be set")
	}
}

// Test basic functionality without testing private types