package auth

import (
	"testing"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestOAuthManager_ExchangeCodeForToken_EmptyCode(t *testing.T) {
	logger := logger.NewLogger()
	cfg := &config.Config{
		Trakt: config.TraktConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
		},
		Auth: config.AuthConfig{
			RedirectURI: "http://localhost:8080/callback",
		},
	}

	oauth := NewOAuthManager(cfg, logger)
	
	// Test with empty code
	token, err := oauth.ExchangeCodeForToken("", "state", "state")
	assert.Error(t, err)
	assert.Nil(t, token)
}

func TestOAuthManager_RefreshToken_EmptyToken(t *testing.T) {
	logger := logger.NewLogger()
	cfg := &config.Config{
		Trakt: config.TraktConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
		},
		Auth: config.AuthConfig{
			RedirectURI: "http://localhost:8080/callback",
		},
	}

	oauth := NewOAuthManager(cfg, logger)
	
	// Test with empty refresh token
	token, err := oauth.RefreshToken("")
	assert.Error(t, err)
	assert.Nil(t, token)
}

func TestOAuthManager_IsTokenExpired_EdgeCases(t *testing.T) {
	logger := logger.NewLogger()
	cfg := &config.Config{}
	oauth := NewOAuthManager(cfg, logger)

	// Test with token that expires in 4 minutes (should be considered expired due to 5-minute buffer)
	token := &TokenResponse{
		AccessToken: "test-token",
		ExpiresIn:   3600,
		CreatedAt:   time.Now().Unix() - 3300, // 55 minutes ago, expires in 5 minutes
	}
	expired := oauth.IsTokenExpired(token)
	assert.True(t, expired)

	// Test with fresh token
	token = &TokenResponse{
		AccessToken: "test-token",
		ExpiresIn:   3600,
		CreatedAt:   time.Now().Unix() - 60, // 1 minute ago
	}
	expired = oauth.IsTokenExpired(token)
	assert.False(t, expired)
}

func TestOAuthManager_generateState_ReturnsUniqueValues(t *testing.T) {
	logger := logger.NewLogger()
	cfg := &config.Config{}
	oauth := NewOAuthManager(cfg, logger)

	// Generate multiple states
	states := make(map[string]bool)
	for i := 0; i < 10; i++ {
		state, err := oauth.generateState()
		assert.NoError(t, err)
		assert.NotEmpty(t, state)
		
		// Check uniqueness
		assert.False(t, states[state], "State should be unique")
		states[state] = true
	}
}

func TestOAuthManager_ParseCallbackURL_ErrorConditions(t *testing.T) {
	logger := logger.NewLogger()
	cfg := &config.Config{}
	oauth := NewOAuthManager(cfg, logger)

	// Test with malformed URL
	code, state, err := oauth.ParseCallbackURL("://invalid-url")
	assert.Error(t, err)
	assert.Empty(t, code)
	assert.Empty(t, state)
	assert.Contains(t, err.Error(), "invalid callback URL")
}

func TestTokenResponse_JSONSerialization(t *testing.T) {
	token := &TokenResponse{
		AccessToken:  "test-token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: "refresh-token",
		Scope:        "public",
		CreatedAt:    1234567890,
	}

	// Test that all fields are properly accessible
	assert.Equal(t, "test-token", token.AccessToken)
	assert.Equal(t, "Bearer", token.TokenType)
	assert.Equal(t, 3600, token.ExpiresIn)
	assert.Equal(t, "refresh-token", token.RefreshToken)
	assert.Equal(t, "public", token.Scope)
	assert.Equal(t, int64(1234567890), token.CreatedAt)
}

func TestTokenError_ErrorHandling(t *testing.T) {
	tokenError := &TokenError{
		Error:            "invalid_grant",
		ErrorDescription: "The provided authorization grant is invalid",
	}

	// Test that error fields are accessible
	assert.Equal(t, "invalid_grant", tokenError.Error)
	assert.Equal(t, "The provided authorization grant is invalid", tokenError.ErrorDescription)
}

func TestOAuthManager_Configuration(t *testing.T) {
	logger := logger.NewLogger()
	cfg := &config.Config{
		Trakt: config.TraktConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
		},
		Auth: config.AuthConfig{
			RedirectURI: "http://localhost:8080/callback",
		},
	}

	oauth := NewOAuthManager(cfg, logger)

	// Test that configuration is properly stored
	assert.Equal(t, cfg, oauth.config)
	assert.Equal(t, logger, oauth.logger)
	assert.NotNil(t, oauth.client)
	assert.Equal(t, 30*time.Second, oauth.client.Timeout)
}

func TestOAuthManager_GetInteractiveAuthInstructions_Content(t *testing.T) {
	logger := logger.NewLogger()
	cfg := &config.Config{}
	oauth := NewOAuthManager(cfg, logger)

	instructions := oauth.GetInteractiveAuthInstructions()

	// Test that instructions contain essential information
	assert.Contains(t, instructions, "AUTHENTIFICATION")
	assert.Contains(t, instructions, "TRAKT.TV")
	assert.Contains(t, instructions, "token")
	assert.NotEmpty(t, instructions)
}