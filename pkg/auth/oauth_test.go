package auth

import (
	"testing"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuthManager_GenerateAuthURL(t *testing.T) {
	cfg := &config.Config{
		Trakt: config.TraktConfig{
			ClientID: "test_client_id",
		},
		Auth: config.AuthConfig{
			RedirectURI: "http://localhost:8080/callback",
		},
	}
	
	log := logger.NewLogger()
	oauthMgr := NewOAuthManager(cfg, log)
	
	authURL, state, err := oauthMgr.GenerateAuthURL()
	
	require.NoError(t, err)
	assert.NotEmpty(t, authURL)
	assert.NotEmpty(t, state)
	assert.Contains(t, authURL, "trakt.tv/oauth/authorize")
	assert.Contains(t, authURL, "client_id=test_client_id")
	assert.Contains(t, authURL, "redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fcallback")
	assert.Contains(t, authURL, "state=")
}

func TestOAuthManager_IsTokenExpired(t *testing.T) {
	cfg := &config.Config{}
	log := logger.NewLogger()
	oauthMgr := NewOAuthManager(cfg, log)
	
	tests := []struct {
		name     string
		token    *TokenResponse
		expected bool
	}{
		{
			name: "token with no expiry",
			token: &TokenResponse{
				AccessToken: "test_token",
				ExpiresIn:   0,
				CreatedAt:   time.Now().Unix(),
			},
			expected: false,
		},
		{
			name: "fresh token",
			token: &TokenResponse{
				AccessToken: "test_token",
				ExpiresIn:   3600, // 1 hour
				CreatedAt:   time.Now().Unix(),
			},
			expected: false,
		},
		{
			name: "expired token",
			token: &TokenResponse{
				AccessToken: "test_token",
				ExpiresIn:   3600, // 1 hour
				CreatedAt:   time.Now().Add(-2 * time.Hour).Unix(),
			},
			expected: true,
		},
		{
			name: "token expiring soon",
			token: &TokenResponse{
				AccessToken: "test_token",
				ExpiresIn:   300, // 5 minutes
				CreatedAt:   time.Now().Unix(),
			},
			expected: true, // Should be considered expired due to buffer time
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := oauthMgr.IsTokenExpired(tt.token)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOAuthManager_GetTokenExpiryTime(t *testing.T) {
	cfg := &config.Config{}
	log := logger.NewLogger()
	oauthMgr := NewOAuthManager(cfg, log)
	
	createdAt := time.Now()
	token := &TokenResponse{
		AccessToken: "test_token",
		ExpiresIn:   3600, // 1 hour
		CreatedAt:   createdAt.Unix(),
	}
	
	expiryTime := oauthMgr.GetTokenExpiryTime(token)
	expected := createdAt.Add(time.Hour)
	
	assert.WithinDuration(t, expected, expiryTime, time.Second)
}

func TestOAuthManager_ParseCallbackURL(t *testing.T) {
	cfg := &config.Config{}
	log := logger.NewLogger()
	oauthMgr := NewOAuthManager(cfg, log)
	
	tests := []struct {
		name          string
		callbackURL   string
		expectedCode  string
		expectedState string
		expectError   bool
	}{
		{
			name:          "valid callback with code and state",
			callbackURL:   "http://localhost:8080/callback?code=test_code&state=test_state",
			expectedCode:  "test_code",
			expectedState: "test_state",
			expectError:   false,
		},
		{
			name:        "callback with error",
			callbackURL: "http://localhost:8080/callback?error=access_denied&error_description=User+denied+access",
			expectError: true,
		},
		{
			name:        "callback without code",
			callbackURL: "http://localhost:8080/callback?state=test_state",
			expectError: true,
		},
		{
			name:        "invalid URL",
			callbackURL: "not-a-url",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, state, err := oauthMgr.ParseCallbackURL(tt.callbackURL)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCode, code)
				assert.Equal(t, tt.expectedState, state)
			}
		})
	}
}

func TestOAuthManager_GetInteractiveAuthInstructions(t *testing.T) {
	cfg := &config.Config{}
	log := logger.NewLogger()
	oauthMgr := NewOAuthManager(cfg, log)
	
	instructions := oauthMgr.GetInteractiveAuthInstructions()
	
	assert.NotEmpty(t, instructions)
	assert.Contains(t, instructions, "AUTHENTIFICATION TRAKT.TV")
	assert.Contains(t, instructions, "docker exec")
	assert.Contains(t, instructions, "http://localhost:8080/callback")
}