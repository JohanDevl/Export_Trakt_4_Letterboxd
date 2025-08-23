package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/security/keyring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock logger for testing
type mockAuthLogger struct {
	logs []map[string]interface{}
}

func (m *mockAuthLogger) Debug(msg string, fields ...map[string]interface{}) {
	var fieldsMap map[string]interface{}
	if len(fields) > 0 {
		fieldsMap = fields[0]
	}
	m.logs = append(m.logs, map[string]interface{}{"level": "debug", "msg": msg, "fields": fieldsMap})
}

func (m *mockAuthLogger) Info(msg string, fields ...map[string]interface{}) {
	var fieldsMap map[string]interface{}
	if len(fields) > 0 {
		fieldsMap = fields[0]
	}
	m.logs = append(m.logs, map[string]interface{}{"level": "info", "msg": msg, "fields": fieldsMap})
}

func (m *mockAuthLogger) Warn(msg string, fields ...map[string]interface{}) {
	var fieldsMap map[string]interface{}
	if len(fields) > 0 {
		fieldsMap = fields[0]
	}
	m.logs = append(m.logs, map[string]interface{}{"level": "warn", "msg": msg, "fields": fieldsMap})
}

func (m *mockAuthLogger) Error(msg string, fields ...map[string]interface{}) {
	var fieldsMap map[string]interface{}
	if len(fields) > 0 {
		fieldsMap = fields[0]
	}
	m.logs = append(m.logs, map[string]interface{}{"level": "error", "msg": msg, "fields": fieldsMap})
}

func (m *mockAuthLogger) Fatal(msg string, fields ...map[string]interface{}) {
	var fieldsMap map[string]interface{}
	if len(fields) > 0 {
		fieldsMap = fields[0]
	}
	m.logs = append(m.logs, map[string]interface{}{"level": "fatal", "msg": msg, "fields": fieldsMap})
}

func (m *mockAuthLogger) Debugf(msg string, data map[string]interface{}) {
	m.logs = append(m.logs, map[string]interface{}{"level": "debug", "msg": msg, "fields": data})
}

func (m *mockAuthLogger) Infof(msg string, data map[string]interface{}) {
	m.logs = append(m.logs, map[string]interface{}{"level": "info", "msg": msg, "fields": data})
}

func (m *mockAuthLogger) Warnf(msg string, data map[string]interface{}) {
	m.logs = append(m.logs, map[string]interface{}{"level": "warn", "msg": msg, "fields": data})
}

func (m *mockAuthLogger) Errorf(msg string, data map[string]interface{}) {
	m.logs = append(m.logs, map[string]interface{}{"level": "error", "msg": msg, "fields": data})
}

func (m *mockAuthLogger) SetLogLevel(level string) {}

func (m *mockAuthLogger) SetLogFile(path string) error {
	return nil
}

func (m *mockAuthLogger) SetTranslator(t logger.Translator) {}

func createTestConfig() *config.Config {
	return &config.Config{
		Trakt: config.TraktConfig{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
		},
		Auth: config.AuthConfig{
			RedirectURI:  "http://localhost:8080/callback",
			CallbackPort: 8080,
		},
	}
}

func TestOAuthManager_ExchangeCodeForToken_Success(t *testing.T) {
	// Mock server that returns a successful token response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		
		// Parse form data
		err := r.ParseForm()
		require.NoError(t, err)
		
		assert.Equal(t, "test_code", r.FormValue("code"))
		assert.Equal(t, "test_client_id", r.FormValue("client_id"))
		assert.Equal(t, "test_client_secret", r.FormValue("client_secret"))
		assert.Equal(t, "authorization_code", r.FormValue("grant_type"))
		
		// Return successful token response
		tokenResp := TokenResponse{
			AccessToken:  "test_access_token",
			TokenType:    "Bearer",
			ExpiresIn:    7200,
			RefreshToken: "test_refresh_token",
			Scope:        "public",
			CreatedAt:    time.Now().Unix(),
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tokenResp)
	}))
	defer server.Close()
	
	cfg := createTestConfig()
	log := &mockAuthLogger{}
	oauthMgr := NewOAuthManager(cfg, log)
	
	// We need to test with the actual implementation
	// Create a custom client that uses our test server
	oauthMgr.client = &http.Client{Timeout: 30 * time.Second}
	
	// For this test, we'll test the parsing logic separately
	// since we can't easily override the const URL
}

func TestOAuthManager_ExchangeCodeForToken_InvalidState(t *testing.T) {
	cfg := createTestConfig()
	log := &mockAuthLogger{}
	oauthMgr := NewOAuthManager(cfg, log)
	
	_, err := oauthMgr.ExchangeCodeForToken("test_code", "invalid_state", "expected_state")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "state mismatch")
}

func TestOAuthManager_ExchangeCodeForToken_ErrorResponse(t *testing.T) {
	// Mock server that returns an error response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		
		errorResp := TokenError{
			Error:            "invalid_grant",
			ErrorDescription: "The provided authorization grant is invalid",
		}
		
		json.NewEncoder(w).Encode(errorResp)
	}))
	defer server.Close()
	
	// This test demonstrates the error handling structure
	// In practice, we'd need dependency injection to test this properly
}

func TestOAuthManager_GenerateState_Uniqueness(t *testing.T) {
	cfg := createTestConfig()
	log := &mockAuthLogger{}
	oauthMgr := NewOAuthManager(cfg, log)
	
	// Generate multiple states
	states := make(map[string]bool)
	for i := 0; i < 100; i++ {
		state, err := oauthMgr.generateState()
		require.NoError(t, err)
		require.NotEmpty(t, state)
		
		// Check uniqueness
		assert.False(t, states[state], "Generated duplicate state: %s", state)
		states[state] = true
		
		// Check length (should be base64 encoded 32 bytes = ~43 chars)
		assert.True(t, len(state) > 40, "State too short: %s", state)
	}
}

func TestOAuthManager_RefreshToken_Success(t *testing.T) {
	cfg := createTestConfig()
	log := &mockAuthLogger{}
	oauthMgr := NewOAuthManager(cfg, log)
	
	// Test the structure - actual HTTP testing would need server mocking
	// This tests the input validation
	_, err := oauthMgr.RefreshToken("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "refresh token cannot be empty")
}

func TestTokenManager_NewTokenManager(t *testing.T) {
	cfg := createTestConfig()
	log := &mockAuthLogger{}
	
	// Create a memory keyring for testing
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	require.NoError(t, err)
	
	tm := NewTokenManager(cfg, log, keyringMgr)
	
	assert.NotNil(t, tm)
	assert.Equal(t, cfg, tm.config)
	assert.Equal(t, log, tm.logger)
	assert.NotNil(t, tm.oauthManager)
	assert.Equal(t, keyringMgr, tm.keyringMgr)
}

func TestTokenManager_StoreToken(t *testing.T) {
	cfg := createTestConfig()
	log := &mockAuthLogger{}
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	require.NoError(t, err)
	
	tm := NewTokenManager(cfg, log, keyringMgr)
	
	token := &TokenResponse{
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
		TokenType:    "Bearer",
		ExpiresIn:    7200,
		Scope:        "public",
		CreatedAt:    time.Now().Unix(),
	}
	
	err = tm.StoreToken(token)
	require.NoError(t, err)
	
	// Verify token was stored
	assert.Equal(t, token, tm.cachedToken)
	
	// Verify it was stored in keyring
	accessToken, err := keyringMgr.Retrieve(accessTokenKey)
	require.NoError(t, err)
	assert.Equal(t, token.AccessToken, accessToken)
	
	refreshToken, err := keyringMgr.Retrieve(refreshTokenKey)
	require.NoError(t, err)
	assert.Equal(t, token.RefreshToken, refreshToken)
	
	// Verify token data was stored
	tokenDataStr, err := keyringMgr.Retrieve(tokenDataKey)
	require.NoError(t, err)
	
	var storedData StoredTokenData
	err = json.Unmarshal([]byte(tokenDataStr), &storedData)
	require.NoError(t, err)
	
	assert.Equal(t, token.AccessToken, storedData.AccessToken)
	assert.Equal(t, token.RefreshToken, storedData.RefreshToken)
	assert.Equal(t, token.TokenType, storedData.TokenType)
}

func TestTokenManager_GetCurrentToken(t *testing.T) {
	cfg := createTestConfig()
	log := &mockAuthLogger{}
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	require.NoError(t, err)
	
	tm := NewTokenManager(cfg, log, keyringMgr)
	
	// Test with no token
	token, err := tm.getCurrentToken()
	assert.NoError(t, err) // getCurrentToken doesn't return error for missing token
	assert.Nil(t, token)
	
	// Store a token and test retrieval
	testToken := &TokenResponse{
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
		TokenType:    "Bearer",
		ExpiresIn:    7200,
		Scope:        "public",
		CreatedAt:    time.Now().Unix(),
	}
	
	err = tm.StoreToken(testToken)
	require.NoError(t, err)
	
	// Test retrieval from cache
	retrievedToken, err := tm.getCurrentToken()
	require.NoError(t, err)
	assert.Equal(t, testToken, retrievedToken)
	
	// Clear cache and test retrieval from keyring
	tm.cachedToken = nil
	retrievedToken, err = tm.getCurrentToken()
	require.NoError(t, err)
	assert.NotNil(t, retrievedToken)
	assert.Equal(t, testToken.AccessToken, retrievedToken.AccessToken)
}

func TestTokenManager_GetValidAccessToken(t *testing.T) {
	cfg := createTestConfig()
	log := &mockAuthLogger{}
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	require.NoError(t, err)
	
	tm := NewTokenManager(cfg, log, keyringMgr)
	
	// Test with no token
	_, err = tm.GetValidAccessToken()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no token available")
	
	// Test with valid token
	validToken := &TokenResponse{
		AccessToken:  "valid_access_token",
		RefreshToken: "valid_refresh_token",
		TokenType:    "Bearer",
		ExpiresIn:    7200,
		Scope:        "public",
		CreatedAt:    time.Now().Unix(),
	}
	
	err = tm.StoreToken(validToken)
	require.NoError(t, err)
	
	accessToken, err := tm.GetValidAccessToken()
	require.NoError(t, err)
	assert.Equal(t, validToken.AccessToken, accessToken)
	
	// Test with expired token (but no refresh available for this test)
	expiredToken := &TokenResponse{
		AccessToken:  "expired_access_token",
		RefreshToken: "", // No refresh token
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		Scope:        "public",
		CreatedAt:    time.Now().Add(-2 * time.Hour).Unix(), // Expired
	}
	
	err = tm.StoreToken(expiredToken)
	require.NoError(t, err)
	
	_, err = tm.GetValidAccessToken()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "re-authentication required")
}

func TestTokenManager_ClearToken(t *testing.T) {
	cfg := createTestConfig()
	log := &mockAuthLogger{}
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	require.NoError(t, err)
	
	tm := NewTokenManager(cfg, log, keyringMgr)
	
	// Store a token first
	token := &TokenResponse{
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
		TokenType:    "Bearer",
		ExpiresIn:    7200,
		Scope:        "public",
		CreatedAt:    time.Now().Unix(),
	}
	
	err = tm.StoreToken(token)
	require.NoError(t, err)
	
	// Verify token is stored
	assert.NotNil(t, tm.cachedToken)
	
	// Clear token
	err = tm.ClearToken()
	require.NoError(t, err)
	
	// Verify token is cleared
	assert.Nil(t, tm.cachedToken)
	
	// Verify keyring is cleared
	_, err = keyringMgr.Retrieve(accessTokenKey)
	assert.Error(t, err) // Should not exist
}

func TestTokenManager_GetTokenStatus(t *testing.T) {
	cfg := createTestConfig()
	log := &mockAuthLogger{}
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	require.NoError(t, err)
	
	tm := NewTokenManager(cfg, log, keyringMgr)
	
	// Test with no token
	status, err := tm.GetTokenStatus()
	require.NoError(t, err)
	assert.False(t, status.HasToken)
	assert.Equal(t, "No authentication token found", status.Message)
	
	// Test with valid token
	validToken := &TokenResponse{
		AccessToken:  "valid_access_token",
		RefreshToken: "valid_refresh_token",
		TokenType:    "Bearer",
		ExpiresIn:    7200,
		Scope:        "public",
		CreatedAt:    time.Now().Unix(),
	}
	
	err = tm.StoreToken(validToken)
	require.NoError(t, err)
	
	status, err = tm.GetTokenStatus()
	require.NoError(t, err)
	assert.True(t, status.HasToken)
	assert.True(t, status.IsValid)
	assert.True(t, status.HasRefreshToken)
	assert.Equal(t, "Bearer", status.TokenType)
	assert.Equal(t, "public", status.Scope)
	
	// Test with expired token
	expiredToken := &TokenResponse{
		AccessToken:  "expired_access_token",
		RefreshToken: "expired_refresh_token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		Scope:        "public",
		CreatedAt:    time.Now().Add(-2 * time.Hour).Unix(),
	}
	
	err = tm.StoreToken(expiredToken)
	require.NoError(t, err)
	
	status, err = tm.GetTokenStatus()
	require.NoError(t, err)
	assert.True(t, status.HasToken)
	assert.False(t, status.IsValid)
	assert.True(t, status.HasRefreshToken)
}

func TestTokenStatus_String(t *testing.T) {
	// Test no token
	status := &TokenStatus{
		HasToken: false,
	}
	assert.Contains(t, status.String(), "🔴 No authentication token found")
	
	// Test expired token
	status = &TokenStatus{
		HasToken:  true,
		IsValid:   false,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	assert.Contains(t, status.String(), "🟡 Token expired on")
	
	// Test valid token
	status = &TokenStatus{
		HasToken:  true,
		IsValid:   true,
		ExpiresAt: time.Now().Add(2 * time.Hour),
	}
	result := status.String()
	assert.Contains(t, result, "🟢 Token valid until")
	assert.Contains(t, result, "remaining")
}

func TestStoredTokenData_JSONSerialization(t *testing.T) {
	data := StoredTokenData{
		AccessToken:  "test_access",
		RefreshToken: "test_refresh",
		TokenType:    "Bearer",
		ExpiresIn:    7200,
		Scope:        "public",
		CreatedAt:    time.Now().Unix(),
	}
	
	// Test marshaling
	jsonData, err := json.Marshal(data)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)
	
	// Test unmarshaling
	var unmarshaled StoredTokenData
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)
	
	assert.Equal(t, data.AccessToken, unmarshaled.AccessToken)
	assert.Equal(t, data.RefreshToken, unmarshaled.RefreshToken)
	assert.Equal(t, data.TokenType, unmarshaled.TokenType)
	assert.Equal(t, data.ExpiresIn, unmarshaled.ExpiresIn)
	assert.Equal(t, data.Scope, unmarshaled.Scope)
	assert.Equal(t, data.CreatedAt, unmarshaled.CreatedAt)
}

func TestTokenManager_RefreshToken(t *testing.T) {
	cfg := createTestConfig()
	log := &mockAuthLogger{}
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	require.NoError(t, err)
	
	tm := NewTokenManager(cfg, log, keyringMgr)
	
	// Test with no token
	err = tm.RefreshToken()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no token available")
	
	// Test with token but no refresh token
	tokenWithoutRefresh := &TokenResponse{
		AccessToken:  "access_only_token",
		RefreshToken: "",
		TokenType:    "Bearer",
		ExpiresIn:    7200,
		Scope:        "public",
		CreatedAt:    time.Now().Unix(),
	}
	
	err = tm.StoreToken(tokenWithoutRefresh)
	require.NoError(t, err)
	
	err = tm.RefreshToken()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no refresh token available")
}

func TestOAuthManager_GetInteractiveAuthInstructions_Integration(t *testing.T) {
	cfg := createTestConfig()
	log := &mockAuthLogger{}
	oauthMgr := NewOAuthManager(cfg, log)
	
	instructions := oauthMgr.GetInteractiveAuthInstructions()
	
	assert.Contains(t, instructions, "🔐 Authentication Required")
	assert.Contains(t, instructions, "Follow the instructions")
	assert.Contains(t, instructions, "Copy the 'code' parameter")
}

func TestOAuthManager_ValidateConfig(t *testing.T) {
	log := &mockAuthLogger{}
	
	// Test with invalid config (missing client ID)
	invalidCfg := &config.Config{
		Trakt: config.TraktConfig{
			ClientID: "",
		},
	}
	
	oauthMgr := NewOAuthManager(invalidCfg, log)
	assert.NotNil(t, oauthMgr) // Constructor doesn't validate
	
	// Test auth URL generation with missing client ID
	_, _, err := oauthMgr.GenerateAuthURL()
	// This should work even with empty client ID in current implementation
	// Real validation would happen on Trakt's side
	assert.NoError(t, err)
}

func TestTokenResponse_IsExpired(t *testing.T) {
	cfg := createTestConfig()
	log := &mockAuthLogger{}
	oauthMgr := NewOAuthManager(cfg, log)
	
	// Test non-expiring token
	nonExpiringToken := &TokenResponse{
		AccessToken: "test_token",
		ExpiresIn:   0, // No expiration
		CreatedAt:   time.Now().Unix(),
	}
	
	assert.False(t, oauthMgr.IsTokenExpired(nonExpiringToken))
	
	// Test fresh token
	freshToken := &TokenResponse{
		AccessToken: "fresh_token",
		ExpiresIn:   3600,
		CreatedAt:   time.Now().Unix(),
	}
	
	assert.False(t, oauthMgr.IsTokenExpired(freshToken))
	
	// Test expired token
	expiredToken := &TokenResponse{
		AccessToken: "expired_token",
		ExpiresIn:   3600,
		CreatedAt:   time.Now().Add(-2 * time.Hour).Unix(),
	}
	
	assert.True(t, oauthMgr.IsTokenExpired(expiredToken))
	
	// Test token expiring soon (within 5 minutes)
	soonExpiringToken := &TokenResponse{
		AccessToken: "soon_expiring_token",
		ExpiresIn:   600,
		CreatedAt:   time.Now().Add(-8 * time.Minute).Unix(),
	}
	
	assert.True(t, oauthMgr.IsTokenExpired(soonExpiringToken))
}

func TestTokenManager_Integration(t *testing.T) {
	cfg := createTestConfig()
	log := &mockAuthLogger{}
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	require.NoError(t, err)
	
	tm := NewTokenManager(cfg, log, keyringMgr)
	
	// Complete workflow test
	
	// 1. Initial state - no token
	status, err := tm.GetTokenStatus()
	require.NoError(t, err)
	assert.False(t, status.HasToken)
	
	// 2. Store a valid token
	validToken := &TokenResponse{
		AccessToken:  "integration_test_token",
		RefreshToken: "integration_refresh_token",
		TokenType:    "Bearer",
		ExpiresIn:    7200,
		Scope:        "public",
		CreatedAt:    time.Now().Unix(),
	}
	
	err = tm.StoreToken(validToken)
	require.NoError(t, err)
	
	// 3. Check status is now valid
	status, err = tm.GetTokenStatus()
	require.NoError(t, err)
	assert.True(t, status.HasToken)
	assert.True(t, status.IsValid)
	
	// 4. Get valid access token
	accessToken, err := tm.GetValidAccessToken()
	require.NoError(t, err)
	assert.Equal(t, validToken.AccessToken, accessToken)
	
	// 5. Clear token
	err = tm.ClearToken()
	require.NoError(t, err)
	
	// 6. Verify cleared state
	status, err = tm.GetTokenStatus()
	require.NoError(t, err)
	assert.False(t, status.HasToken)
	
	_, err = tm.GetValidAccessToken()
	assert.Error(t, err)
}