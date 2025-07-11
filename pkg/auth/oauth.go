package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

const (
	traktAuthURL  = "https://trakt.tv/oauth/authorize"
	traktTokenURL = "https://api.trakt.tv/oauth/token"
)

type OAuthManager struct {
	config *config.Config
	logger logger.Logger
	client *http.Client
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	CreatedAt    int64  `json:"created_at"`
}

type TokenError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func NewOAuthManager(cfg *config.Config, log logger.Logger) *OAuthManager {
	return &OAuthManager{
		config: cfg,
		logger: log,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (o *OAuthManager) GenerateAuthURL() (string, string, error) {
	state, err := o.generateState()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate state: %w", err)
	}

	authURL := fmt.Sprintf("%s?response_type=code&client_id=%s&redirect_uri=%s&state=%s",
		traktAuthURL,
		url.QueryEscape(o.config.Trakt.ClientID),
		url.QueryEscape(o.config.Auth.RedirectURI),
		url.QueryEscape(state),
	)

	o.logger.Info("oauth.auth_url_generated", map[string]interface{}{
		"redirect_uri": o.config.Auth.RedirectURI,
		"state":        state,
	})

	return authURL, state, nil
}

func (o *OAuthManager) ExchangeCodeForToken(code, state, expectedState string) (*TokenResponse, error) {
	if state != expectedState {
		return nil, fmt.Errorf("state mismatch: expected %s, got %s", expectedState, state)
	}

	o.logger.Info("oauth.exchanging_code", map[string]interface{}{
		"code_length": len(code),
	})

	data := url.Values{
		"code":          {code},
		"client_id":     {o.config.Trakt.ClientID},
		"client_secret": {o.config.Trakt.ClientSecret},
		"redirect_uri":  {o.config.Auth.RedirectURI},
		"grant_type":    {"authorization_code"},
	}

	resp, err := o.client.PostForm(traktTokenURL, data)
	if err != nil {
		o.logger.Error("oauth.exchange_request_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("token exchange request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var tokenError TokenError
		if err := json.Unmarshal(body, &tokenError); err != nil {
			return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
		}
		o.logger.Error("oauth.exchange_failed", map[string]interface{}{
			"status":      resp.StatusCode,
			"error":       tokenError.Error,
			"description": tokenError.ErrorDescription,
		})
		return nil, fmt.Errorf("token exchange failed: %s - %s", tokenError.Error, tokenError.ErrorDescription)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	tokenResp.CreatedAt = time.Now().Unix()

	o.logger.Info("oauth.token_exchange_success", map[string]interface{}{
		"token_type": tokenResp.TokenType,
		"expires_in": tokenResp.ExpiresIn,
		"scope":      tokenResp.Scope,
	})

	return &tokenResp, nil
}

func (o *OAuthManager) RefreshToken(refreshToken string) (*TokenResponse, error) {
	o.logger.Info("oauth.refreshing_token", nil)

	data := url.Values{
		"refresh_token": {refreshToken},
		"client_id":     {o.config.Trakt.ClientID},
		"client_secret": {o.config.Trakt.ClientSecret},
		"grant_type":    {"refresh_token"},
	}

	resp, err := o.client.PostForm(traktTokenURL, data)
	if err != nil {
		o.logger.Error("oauth.refresh_request_failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("token refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var tokenError TokenError
		if err := json.Unmarshal(body, &tokenError); err != nil {
			return nil, fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, string(body))
		}
		o.logger.Error("oauth.refresh_failed", map[string]interface{}{
			"status":      resp.StatusCode,
			"error":       tokenError.Error,
			"description": tokenError.ErrorDescription,
		})
		return nil, fmt.Errorf("token refresh failed: %s - %s", tokenError.Error, tokenError.ErrorDescription)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	tokenResp.CreatedAt = time.Now().Unix()

	o.logger.Info("oauth.token_refresh_success", map[string]interface{}{
		"token_type": tokenResp.TokenType,
		"expires_in": tokenResp.ExpiresIn,
		"scope":      tokenResp.Scope,
	})

	return &tokenResp, nil
}

func (o *OAuthManager) ValidateToken(accessToken string) error {
	req, err := http.NewRequest("GET", "https://api.trakt.tv/users/settings", nil)
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("trakt-api-version", "2")
	req.Header.Set("trakt-api-key", o.config.Trakt.ClientID)

	resp, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("token validation request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("token is invalid or expired")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token validation failed with status %d", resp.StatusCode)
	}

	o.logger.Info("oauth.token_validation_success", nil)
	return nil
}

func (o *OAuthManager) IsTokenExpired(token *TokenResponse) bool {
	if token.ExpiresIn == 0 {
		return false
	}

	expiryTime := time.Unix(token.CreatedAt, 0).Add(time.Duration(token.ExpiresIn) * time.Second)
	
	bufferTime := 5 * time.Minute
	return time.Now().Add(bufferTime).After(expiryTime)
}

func (o *OAuthManager) GetTokenExpiryTime(token *TokenResponse) time.Time {
	if token.ExpiresIn == 0 {
		return time.Time{}
	}
	return time.Unix(token.CreatedAt, 0).Add(time.Duration(token.ExpiresIn) * time.Second)
}

func (o *OAuthManager) generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (o *OAuthManager) StartLocalCallbackServer() (string, chan string, chan error, error) {
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)
	
	port := o.config.Auth.CallbackPort
	if port == 0 {
		port = 8080
	}

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		errorParam := r.URL.Query().Get("error")
		
		if errorParam != "" {
			errDescription := r.URL.Query().Get("error_description")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head><title>Authentication Error</title></head>
<body>
<h2>❌ Authentication Error</h2>
<p><strong>Error:</strong> %s</p>
<p><strong>Description:</strong> %s</p>
<p>Please close this window and try again.</p>
</body>
</html>`, errorParam, errDescription)
			errChan <- fmt.Errorf("oauth error: %s - %s", errorParam, errDescription)
			return
		}

		if code == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<head><title>Authentication Error</title></head>
<body>
<h2>❌ Authentication Error</h2>
<p>No authorization code received.</p>
<p>Please close this window and try again.</p>
</body>
</html>`)
			errChan <- fmt.Errorf("no authorization code received")
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<head><title>Authentication Success</title></head>
<body>
<h2>✅ Authentication Successful!</h2>
<p>You have successfully authenticated with Trakt.tv.</p>
<p>You can now close this window and return to your application.</p>
<script>setTimeout(function(){window.close();}, 3000);</script>
</body>
</html>`)
		
		codeChan <- code
		
		go func() {
			time.Sleep(2 * time.Second)
			server.Close()
		}()
	})

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("callback server error: %w", err)
		}
	}()

	callbackURL := fmt.Sprintf("http://localhost:%d/callback", port)
	o.logger.Info("oauth.callback_server_started", map[string]interface{}{
		"callback_url": callbackURL,
		"port":         port,
	})

	return callbackURL, codeChan, errChan, nil
}

func (o *OAuthManager) ParseCallbackURL(callbackURL string) (string, string, error) {
	parsedURL, err := url.Parse(callbackURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse callback URL: %w", err)
	}

	query := parsedURL.Query()
	code := query.Get("code")
	state := query.Get("state")
	errorParam := query.Get("error")

	if errorParam != "" {
		errorDescription := query.Get("error_description")
		return "", "", fmt.Errorf("oauth error: %s - %s", errorParam, errorDescription)
	}

	if code == "" {
		return "", "", fmt.Errorf("no authorization code in callback URL")
	}

	return code, state, nil
}

func (o *OAuthManager) GetInteractiveAuthInstructions() string {
	return `
🔑 AUTHENTIFICATION TRAKT.TV REQUISE

Votre token d'accès a expiré ou n'est pas configuré.
Pour configurer l'authentification OAuth automatique :

1️⃣ ÉTAPE 1: Configurez votre application Trakt.tv
   - Allez sur https://trakt.tv/oauth/applications
   - Créez une nouvelle application ou modifiez votre application existante
   - Définissez le Redirect URI: http://localhost:8080/callback

2️⃣ ÉTAPE 2: Lancez l'authentification interactive
   Dans votre container Docker, exécutez :
   
   docker exec -it <container_name> /app/export-trakt auth

3️⃣ ÉTAPE 3: Suivez les instructions affichées
   - Une URL d'autorisation sera générée
   - Ouvrez cette URL dans votre navigateur
   - Autorisez l'application sur Trakt.tv
   - Le token sera automatiquement sauvegardé

4️⃣ ÉTAPE 4: Vérifiez la configuration
   docker exec -it <container_name> /app/export-trakt token-status

💡 CONSEILS:
- Le token sera automatiquement renouvelé à l'avenir
- Gardez votre client_secret sécurisé
- Le token est stocké de manière sécurisée via le système de keyring

🆘 EN CAS DE PROBLÈME:
- Vérifiez que le port 8080 est accessible
- Assurez-vous que votre Redirect URI est correctement configuré
- Consultez les logs avec: docker logs <container_name>
`
}