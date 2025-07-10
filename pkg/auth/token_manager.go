package auth

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/security/keyring"
)

const (
	accessTokenKey  = "trakt_access_token"
	refreshTokenKey = "trakt_refresh_token"
	tokenDataKey    = "trakt_token_data"
)

type TokenManager struct {
	config       *config.Config
	logger       logger.Logger
	oauthManager *OAuthManager
	keyringMgr   *keyring.Manager
	mutex        sync.RWMutex
	cachedToken  *TokenResponse
}

type StoredTokenData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	CreatedAt    int64  `json:"created_at"`
}

func NewTokenManager(cfg *config.Config, log logger.Logger, keyringMgr *keyring.Manager) *TokenManager {
	oauthMgr := NewOAuthManager(cfg, log)
	
	return &TokenManager{
		config:       cfg,
		logger:       log,
		oauthManager: oauthMgr,
		keyringMgr:   keyringMgr,
	}
}

func (tm *TokenManager) GetValidAccessToken() (string, error) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	token, err := tm.getCurrentToken()
	if err != nil {
		tm.logger.Error("token.get_current_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return "", fmt.Errorf("failed to get current token: %w", err)
	}

	if token == nil {
		tm.logger.Info("token.not_found", nil)
		return "", fmt.Errorf("no token available, authentication required")
	}

	if !tm.oauthManager.IsTokenExpired(token) {
		tm.logger.Debug("token.valid", map[string]interface{}{
			"expires_at": tm.oauthManager.GetTokenExpiryTime(token),
		})
		return token.AccessToken, nil
	}

	tm.logger.Info("token.expired_refreshing", map[string]interface{}{
		"expired_at": tm.oauthManager.GetTokenExpiryTime(token),
	})

	if token.RefreshToken == "" {
		tm.logger.Error("token.no_refresh_token", nil)
		return "", fmt.Errorf("token expired and no refresh token available, re-authentication required")
	}

	refreshedToken, err := tm.oauthManager.RefreshToken(token.RefreshToken)
	if err != nil {
		tm.logger.Error("token.refresh_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return "", fmt.Errorf("failed to refresh token: %w", err)
	}

	if err := tm.storeToken(refreshedToken); err != nil {
		tm.logger.Error("token.store_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return "", fmt.Errorf("failed to store refreshed token: %w", err)
	}

	tm.cachedToken = refreshedToken
	tm.logger.Info("token.refresh_success", map[string]interface{}{
		"new_expires_at": tm.oauthManager.GetTokenExpiryTime(refreshedToken),
	})

	return refreshedToken.AccessToken, nil
}

func (tm *TokenManager) StoreToken(token *TokenResponse) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if err := tm.storeToken(token); err != nil {
		return err
	}

	tm.cachedToken = token
	tm.logger.Info("token.stored", map[string]interface{}{
		"expires_at": tm.oauthManager.GetTokenExpiryTime(token),
	})

	return nil
}

func (tm *TokenManager) GetTokenStatus() (*TokenStatus, error) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	token, err := tm.getCurrentToken()
	if err != nil {
		return &TokenStatus{
			HasToken: false,
			Error:    err.Error(),
		}, nil
	}

	if token == nil {
		return &TokenStatus{
			HasToken: false,
			Message:  "No authentication token found",
		}, nil
	}

	expiryTime := tm.oauthManager.GetTokenExpiryTime(token)
	isExpired := tm.oauthManager.IsTokenExpired(token)
	hasRefreshToken := token.RefreshToken != ""

	status := &TokenStatus{
		HasToken:        true,
		IsValid:         !isExpired,
		ExpiresAt:       expiryTime,
		HasRefreshToken: hasRefreshToken,
		TokenType:       token.TokenType,
		Scope:           token.Scope,
	}

	if isExpired {
		if hasRefreshToken {
			status.Message = "Token expired but can be refreshed automatically"
		} else {
			status.Message = "Token expired - re-authentication required"
			status.Error = "No refresh token available"
		}
	} else {
		timeUntilExpiry := time.Until(expiryTime)
		if timeUntilExpiry < 24*time.Hour {
			status.Message = fmt.Sprintf("Token expires in %s", timeUntilExpiry.Round(time.Minute))
		} else {
			status.Message = "Token is valid"
		}
	}

	return status, nil
}

func (tm *TokenManager) RefreshToken() error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	token, err := tm.getCurrentToken()
	if err != nil {
		return fmt.Errorf("failed to get current token: %w", err)
	}

	if token == nil {
		return fmt.Errorf("no token to refresh")
	}

	if token.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	refreshedToken, err := tm.oauthManager.RefreshToken(token.RefreshToken)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	if err := tm.storeToken(refreshedToken); err != nil {
		return fmt.Errorf("failed to store refreshed token: %w", err)
	}

	tm.cachedToken = refreshedToken
	tm.logger.Info("token.manual_refresh_success", map[string]interface{}{
		"new_expires_at": tm.oauthManager.GetTokenExpiryTime(refreshedToken),
	})

	return nil
}

func (tm *TokenManager) ClearToken() error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	tm.cachedToken = nil

	if err := tm.keyringMgr.Delete(accessTokenKey); err != nil && err != keyring.ErrCredentialNotFound {
		tm.logger.Warn("token.clear_access_token_failed", map[string]interface{}{
			"error": err.Error(),
		})
	}

	if err := tm.keyringMgr.Delete(refreshTokenKey); err != nil && err != keyring.ErrCredentialNotFound {
		tm.logger.Warn("token.clear_refresh_token_failed", map[string]interface{}{
			"error": err.Error(),
		})
	}

	if err := tm.keyringMgr.Delete(tokenDataKey); err != nil && err != keyring.ErrCredentialNotFound {
		tm.logger.Warn("token.clear_token_data_failed", map[string]interface{}{
			"error": err.Error(),
		})
	}

	tm.logger.Info("token.cleared", nil)
	return nil
}

func (tm *TokenManager) ValidateStoredToken() error {
	token, err := tm.GetValidAccessToken()
	if err != nil {
		return err
	}

	return tm.oauthManager.ValidateToken(token)
}

func (tm *TokenManager) getCurrentToken() (*TokenResponse, error) {
	if tm.cachedToken != nil {
		return tm.cachedToken, nil
	}

	tokenData, err := tm.keyringMgr.Retrieve(tokenDataKey)
	if err != nil {
		if err == keyring.ErrCredentialNotFound {
			accessToken, err := tm.keyringMgr.Retrieve(accessTokenKey)
			if err != nil {
				if err == keyring.ErrCredentialNotFound {
					if tm.config.Trakt.AccessToken != "" {
						legacyToken := &TokenResponse{
							AccessToken: tm.config.Trakt.AccessToken,
							TokenType:   "Bearer",
							ExpiresIn:   0, // Legacy tokens don't expire
							CreatedAt:   time.Now().Unix(),
						}
						tm.logger.Info("token.using_legacy_config", nil)
						return legacyToken, nil
					}
					return nil, nil
				}
				return nil, fmt.Errorf("failed to get access token: %w", err)
			}

			refreshToken, _ := tm.keyringMgr.Retrieve(refreshTokenKey)

			legacyToken := &TokenResponse{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
				TokenType:    "Bearer",
				ExpiresIn:    0, // Legacy tokens don't expire
				CreatedAt:    time.Now().Unix(),
			}
			return legacyToken, nil
		}
		return nil, fmt.Errorf("failed to get token data: %w", err)
	}

	var storedData StoredTokenData
	if err := json.Unmarshal([]byte(tokenData), &storedData); err != nil {
		return nil, fmt.Errorf("failed to parse stored token data: %w", err)
	}

	token := &TokenResponse{
		AccessToken:  storedData.AccessToken,
		RefreshToken: storedData.RefreshToken,
		TokenType:    storedData.TokenType,
		ExpiresIn:    storedData.ExpiresIn,
		Scope:        storedData.Scope,
		CreatedAt:    storedData.CreatedAt,
	}

	tm.cachedToken = token
	return token, nil
}

func (tm *TokenManager) storeToken(token *TokenResponse) error {
	storedData := StoredTokenData{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresIn:    token.ExpiresIn,
		Scope:        token.Scope,
		CreatedAt:    token.CreatedAt,
	}

	tokenDataJSON, err := json.Marshal(storedData)
	if err != nil {
		return fmt.Errorf("failed to marshal token data: %w", err)
	}

	if err := tm.keyringMgr.Store(tokenDataKey, string(tokenDataJSON)); err != nil {
		return fmt.Errorf("failed to store token data: %w", err)
	}

	if err := tm.keyringMgr.Store(accessTokenKey, token.AccessToken); err != nil {
		return fmt.Errorf("failed to store access token: %w", err)
	}

	if token.RefreshToken != "" {
		if err := tm.keyringMgr.Store(refreshTokenKey, token.RefreshToken); err != nil {
			return fmt.Errorf("failed to store refresh token: %w", err)
		}
	}

	return nil
}

type TokenStatus struct {
	HasToken        bool      `json:"has_token"`
	IsValid         bool      `json:"is_valid"`
	ExpiresAt       time.Time `json:"expires_at,omitempty"`
	HasRefreshToken bool      `json:"has_refresh_token"`
	TokenType       string    `json:"token_type,omitempty"`
	Scope           string    `json:"scope,omitempty"`
	Message         string    `json:"message,omitempty"`
	Error           string    `json:"error,omitempty"`
}

func (ts *TokenStatus) String() string {
	if !ts.HasToken {
		return "❌ No authentication token found"
	}

	status := "✅ Token is valid"
	if !ts.IsValid {
		status = "⚠️ Token is expired"
	}

	refreshStatus := ""
	if ts.HasRefreshToken {
		refreshStatus = " (auto-refresh available)"
	} else {
		refreshStatus = " (manual re-auth required)"
	}

	expiryInfo := ""
	if !ts.ExpiresAt.IsZero() {
		if ts.IsValid {
			timeUntilExpiry := time.Until(ts.ExpiresAt)
			expiryInfo = fmt.Sprintf("\n   Expires: %s (in %s)", 
				ts.ExpiresAt.Format("2006-01-02 15:04:05"), 
				timeUntilExpiry.Round(time.Minute))
		} else {
			expiryInfo = fmt.Sprintf("\n   Expired: %s", ts.ExpiresAt.Format("2006-01-02 15:04:05"))
		}
	}

	scopeInfo := ""
	if ts.Scope != "" {
		scopeInfo = fmt.Sprintf("\n   Scope: %s", ts.Scope)
	}

	return fmt.Sprintf("%s%s%s%s", status, refreshStatus, expiryInfo, scopeInfo)
}