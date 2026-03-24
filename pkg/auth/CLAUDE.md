# CLAUDE.md - Authentification OAuth 2.0

## Module Overview

Ce module implémente un système d'authentification OAuth 2.0 complet et sécurisé pour l'API Trakt.tv. Il gère le flux d'autorisation, le stockage sécurisé des tokens, le rafraîchissement automatique, et fournit une abstraction complète pour toutes les opérations liées à l'authentification.

## Architecture du Module

### 🔐 Composants Principaux

#### OAuthManager
- **Gestion du Flux OAuth** : Génération d'URLs d'autorisation, échange de codes contre tokens
- **Rafraîchissement Automatique** : Renouvellement transparent des access tokens
- **Validation de Sécurité** : Vérification des states et protection CSRF
- **Gestion des Erreurs** : Handling robuste des erreurs API et réseau

#### TokenManager
- **Gestion Centralisée des Tokens** : Cache en mémoire et stockage persistant
- **Accès Thread-Safe** : Synchronisation avec mutex pour concurrence
- **Intégration Keyring** : Stockage sécurisé via système de keyring
- **Statut et Diagnostics** : Monitoring de la validité et expiration des tokens

### 📋 Structures de Données

#### TokenResponse (OAuth Token)
```go
type TokenResponse struct {
    AccessToken  string `json:"access_token"`   // Token d'accès pour API
    TokenType    string `json:"token_type"`     // Type "Bearer"
    ExpiresIn    int    `json:"expires_in"`     // Durée de vie en secondes
    RefreshToken string `json:"refresh_token"`  // Token de rafraîchissement
    Scope        string `json:"scope"`          // Permissions accordées
    CreatedAt    int64  `json:"created_at"`     // Timestamp de création
}
```

#### StoredTokenData (Stockage Persistant)
```go
type StoredTokenData struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    TokenType    string `json:"token_type"`
    ExpiresIn    int    `json:"expires_in"`
    Scope        string `json:"scope"`
    CreatedAt    int64  `json:"created_at"`
}
```

#### TokenStatus (Diagnostics)
```go
type TokenStatus struct {
    HasToken        bool      // Token disponible
    IsValid         bool      // Token non expiré
    ExpiresAt       time.Time // Date d'expiration
    HasRefreshToken bool      // Refresh token disponible
    TokenType       string    // Type de token
    Scope           string    // Permissions
    Error           string    // Erreur éventuelle
    Message         string    // Message informatif
}
```

### 🚀 Flux d'Authentification OAuth 2.0

#### 1. Génération d'URL d'Autorisation
```go
func (o *OAuthManager) GenerateAuthURL() (string, string, error) {
    // Génère un state cryptographiquement sécurisé
    state, err := o.generateState()
    
    // Construit l'URL d'autorisation Trakt.tv
    authURL := fmt.Sprintf("%s?response_type=code&client_id=%s&redirect_uri=%s&state=%s",
        traktAuthURL,
        url.QueryEscape(o.config.Trakt.ClientID),
        url.QueryEscape(o.config.Auth.RedirectURI),
        url.QueryEscape(state),
    )
    
    return authURL, state, nil
}
```

#### 2. Échange Code → Token
```go
func (o *OAuthManager) ExchangeCodeForToken(code, state, expectedState string) (*TokenResponse, error) {
    // Validation du state pour prévenir CSRF
    if state != expectedState {
        return nil, fmt.Errorf("state mismatch")
    }
    
    // Construction de la requête POST
    data := url.Values{
        "code":          {code},
        "client_id":     {o.config.Trakt.ClientID},
        "client_secret": {o.config.Trakt.ClientSecret},
        "redirect_uri":  {o.config.Auth.RedirectURI},
        "grant_type":    {"authorization_code"},
    }
    
    // Échange avec l'API Trakt.tv
    resp, err := o.client.PostForm(traktTokenURL, data)
    
    // Parsing et validation de la réponse
    return parseTokenResponse(resp)
}
```

#### 3. Rafraîchissement de Token
```go
func (o *OAuthManager) RefreshToken(refreshToken string) (*TokenResponse, error) {
    data := url.Values{
        "refresh_token": {refreshToken},
        "client_id":     {o.config.Trakt.ClientID},
        "client_secret": {o.config.Trakt.ClientSecret},
        "grant_type":    {"refresh_token"},
    }
    
    resp, err := o.client.PostForm(traktTokenURL, data)
    return parseTokenResponse(resp)
}
```

### 🔒 Gestion Sécurisée des Tokens

#### TokenManager avec Thread Safety
```go
type TokenManager struct {
    config       *config.Config
    logger       logger.Logger
    oauthManager *OAuthManager
    keyringMgr   *keyring.Manager  // Stockage sécurisé
    mutex        sync.RWMutex      // Synchronisation
    cachedToken  *TokenResponse    // Cache en mémoire
}
```

#### Accès Token Automatique
```go
func (tm *TokenManager) GetValidAccessToken() (string, error) {
    tm.mutex.Lock()
    defer tm.mutex.Unlock()
    
    // 1. Récupération du token courant
    token, err := tm.getCurrentToken()
    if err != nil || token == nil {
        return "", fmt.Errorf("no token available")
    }
    
    // 2. Vérification de l'expiration
    if !tm.oauthManager.IsTokenExpired(token) {
        return token.AccessToken, nil
    }
    
    // 3. Rafraîchissement automatique si nécessaire
    if token.RefreshToken == "" {
        return "", fmt.Errorf("re-authentication required")
    }
    
    refreshedToken, err := tm.oauthManager.RefreshToken(token.RefreshToken)
    if err != nil {
        return "", fmt.Errorf("failed to refresh token: %w", err)
    }
    
    // 4. Stockage du nouveau token
    if err := tm.storeToken(refreshedToken); err != nil {
        return "", fmt.Errorf("failed to store refreshed token: %w", err)
    }
    
    tm.cachedToken = refreshedToken
    return refreshedToken.AccessToken, nil
}
```

### 🏪 Backends de Stockage (Keyring)

#### Support Multi-Backend
```go
const (
    accessTokenKey  = "trakt_access_token"
    refreshTokenKey = "trakt_refresh_token"
    tokenDataKey    = "trakt_token_data"
)

// Stockage via keyring système (macOS Keychain, Windows Credential Manager, etc.)
err := tm.keyringMgr.Set(accessTokenKey, token.AccessToken)
err := tm.keyringMgr.Set(refreshTokenKey, token.RefreshToken)

// Stockage des métadonnées complètes
tokenData := &StoredTokenData{...}
jsonData, _ := json.Marshal(tokenData)
err := tm.keyringMgr.Set(tokenDataKey, string(jsonData))
```

#### Backends Supportés
- **system** : Keyring système natif (recommandé)
  - macOS : Keychain
  - Windows : Credential Manager
  - Linux : Secret Service (GNOME/KDE)
- **env** : Variables d'environnement
- **file** : Fichier chiffré AES-256
- **memory** : Stockage en mémoire (non persistant)

### 🌐 Serveur de Callback Local

#### Serveur HTTP Temporaire
```go
func (o *OAuthManager) StartLocalCallbackServer() (string, chan string, chan error, error) {
    codeChan := make(chan string, 1)
    errChan := make(chan error, 1)
    
    // Serveur HTTP pour recevoir le callback OAuth
    mux := http.NewServeMux()
    mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
        code := r.URL.Query().Get("code")
        state := r.URL.Query().Get("state")
        errorParam := r.URL.Query().Get("error")
        
        if errorParam != "" {
            errChan <- fmt.Errorf("oauth error: %s", errorParam)
            return
        }
        
        if code == "" {
            errChan <- fmt.Errorf("no authorization code received")
            return
        }
        
        codeChan <- code
        
        // Page de succès pour l'utilisateur
        fmt.Fprintf(w, `
            <html><body>
                <h2>✅ Authentication Successful!</h2>
                <p>You can now close this window and return to the terminal.</p>
            </body></html>
        `)
    })
    
    server := &http.Server{
        Addr:    fmt.Sprintf(":%d", o.config.Auth.CallbackPort),
        Handler: mux,
    }
    
    go server.ListenAndServe()
    
    callbackURL := fmt.Sprintf("http://localhost:%d/callback", o.config.Auth.CallbackPort)
    return callbackURL, codeChan, errChan, nil
}
```

### 📊 Diagnostics et Monitoring

#### Statut Détaillé des Tokens
```go
func (tm *TokenManager) GetTokenStatus() (*TokenStatus, error) {
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
    
    return &TokenStatus{
        HasToken:        true,
        IsValid:         !isExpired,
        ExpiresAt:       expiryTime,
        HasRefreshToken: token.RefreshToken != "",
        TokenType:       token.TokenType,
        Scope:           token.Scope,
    }, nil
}
```

#### Affichage Formaté du Statut
```go
func (ts *TokenStatus) String() string {
    if !ts.HasToken {
        return "🔴 No authentication token found"
    }
    
    if !ts.IsValid {
        return fmt.Sprintf("🟡 Token expired on %s", ts.ExpiresAt.Format("2006-01-02 15:04:05"))
    }
    
    timeUntilExpiry := time.Until(ts.ExpiresAt)
    return fmt.Sprintf("🟢 Token valid until %s (%s remaining)", 
        ts.ExpiresAt.Format("2006-01-02 15:04:05"),
        timeUntilExpiry.Round(time.Minute))
}
```

### 🔧 Gestion des Erreurs

#### Types d'Erreurs OAuth
```go
type TokenError struct {
    Error            string `json:"error"`
    ErrorDescription string `json:"error_description"`
}

// Erreurs communes :
// - invalid_grant : Code expiré ou invalide
// - invalid_client : Credentials incorrects
// - invalid_request : Paramètres manquants
// - unsupported_grant_type : Type de grant non supporté
```

#### Logs Structurés
```go
// Succès
tm.logger.Info("oauth.token_exchange_success", map[string]interface{}{
    "token_type": tokenResp.TokenType,
    "expires_in": tokenResp.ExpiresIn,
    "scope":      tokenResp.Scope,
})

// Erreurs
tm.logger.Error("oauth.exchange_failed", map[string]interface{}{
    "status":      resp.StatusCode,
    "error":       tokenError.Error,
    "description": tokenError.ErrorDescription,
})
```

### 🛡️ Sécurité

#### Protection CSRF avec State
```go
func (o *OAuthManager) generateState() (string, error) {
    // Génération cryptographiquement sécurisée
    b := make([]byte, 32)
    _, err := rand.Read(b)
    if err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(b), nil
}
```

#### Validation de Token
```go
func (o *OAuthManager) IsTokenExpired(token *TokenResponse) bool {
    if token.ExpiresIn <= 0 {
        return false // Token sans expiration
    }
    
    expiryTime := time.Unix(token.CreatedAt, 0).Add(time.Duration(token.ExpiresIn) * time.Second)
    
    // Marge de sécurité de 5 minutes
    return time.Now().Add(5 * time.Minute).After(expiryTime)
}

func (o *OAuthManager) GetTokenExpiryTime(token *TokenResponse) time.Time {
    return time.Unix(token.CreatedAt, 0).Add(time.Duration(token.ExpiresIn) * time.Second)
}
```

### 🔄 Opérations de Gestion

#### Nettoyage des Tokens
```go
func (tm *TokenManager) ClearToken() error {
    tm.mutex.Lock()
    defer tm.mutex.Unlock()
    
    // Suppression du keyring
    tm.keyringMgr.Delete(accessTokenKey)
    tm.keyringMgr.Delete(refreshTokenKey)
    tm.keyringMgr.Delete(tokenDataKey)
    
    // Nettoyage du cache
    tm.cachedToken = nil
    
    tm.logger.Info("token.cleared", nil)
    return nil
}
```

#### Rafraîchissement Manuel
```go
func (tm *TokenManager) RefreshToken() error {
    tm.mutex.Lock()
    defer tm.mutex.Unlock()
    
    token, err := tm.getCurrentToken()
    if err != nil || token == nil {
        return fmt.Errorf("no token available")
    }
    
    if token.RefreshToken == "" {
        return fmt.Errorf("no refresh token available")
    }
    
    refreshedToken, err := tm.oauthManager.RefreshToken(token.RefreshToken)
    if err != nil {
        return fmt.Errorf("token refresh failed: %w", err)
    }
    
    return tm.storeToken(refreshedToken)
}
```

### 📚 Exemples d'Usage

#### Authentification Interactive Complète
```go
// 1. Configuration
cfg := &config.Config{...}
log := logger.NewLogger()
keyringMgr, _ := keyring.NewManager(keyring.SystemBackend)
tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)

// 2. Vérification du statut
status, _ := tokenManager.GetTokenStatus()
if !status.HasToken {
    // 3. Authentification nécessaire
    oauthMgr := auth.NewOAuthManager(cfg, log)
    
    // 4. Démarrage serveur callback
    callbackURL, codeChan, errChan, _ := oauthMgr.StartLocalCallbackServer()
    
    // 5. Génération URL d'autorisation
    authURL, state, _ := oauthMgr.GenerateAuthURL()
    fmt.Printf("Open: %s\n", authURL)
    
    // 6. Attente du code d'autorisation
    select {
    case code := <-codeChan:
        // 7. Échange code contre token
        token, _ := oauthMgr.ExchangeCodeForToken(code, state, state)
        
        // 8. Stockage sécurisé
        tokenManager.StoreToken(token)
        
    case err := <-errChan:
        log.Error("Authentication failed", err)
    }
}

// 9. Utilisation normale
accessToken, _ := tokenManager.GetValidAccessToken()
```

#### Authentification avec Code Manuel
```go
// Pour environnements sans browser ou headless
authURL, state, _ := oauthMgr.GenerateAuthURL()
fmt.Printf("1. Open: %s\n", authURL)
fmt.Printf("2. Authorize and copy the code parameter\n")
fmt.Printf("3. Paste the code: ")

var authCode string
fmt.Scanln(&authCode)

token, err := oauthMgr.ExchangeCodeForToken(authCode, state, state)
if err != nil {
    log.Fatal("Authentication failed", err)
}

tokenManager.StoreToken(token)
```

### ⚙️ Configuration

#### Configuration OAuth
```toml
[auth]
use_oauth = true
redirect_uri = "http://localhost:8080/callback"
callback_port = 8080
token_refresh_threshold = 300  # 5 minutes avant expiration

[trakt]
client_id = "your_client_id"
client_secret = "your_client_secret"

[security]
keyring_backend = "system"  # system, env, file, memory
encryption_enabled = true
```

### 🚨 Gestion des Cas d'Erreur

#### Scénarios Courants
1. **Token Expiré** : Rafraîchissement automatique transparent
2. **Refresh Token Invalide** : Demande de ré-authentification
3. **Credentials Incorrects** : Erreur explicite avec instructions
4. **Réseau Indisponible** : Retry avec backoff exponentiel
5. **Keyring Indisponible** : Fallback vers variables d'environnement

#### Recovery et Fallback
- **Auto-Recovery** : Tentative de rafraîchissement automatique
- **Graceful Degradation** : Fonctionnement dégradé si possible
- **User Guidance** : Instructions claires pour résoudre les problèmes
- **Logging Complet** : Traçabilité pour debugging

Ce module fournit une fondation robuste et sécurisée pour toutes les opérations d'authentification OAuth 2.0 avec Trakt.tv, avec une expérience utilisateur fluide et une gestion d'erreurs complète.