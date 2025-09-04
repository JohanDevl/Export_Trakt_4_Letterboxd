package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

const (
	// CSRFTokenLength is the length of CSRF tokens in bytes
	CSRFTokenLength = 32
	// CSRFTokenHeader is the header name for CSRF tokens
	CSRFTokenHeader = "X-CSRF-Token"
	// CSRFCookieName is the name of the CSRF cookie
	CSRFCookieName = "csrf_token"
	// CSRFFormField is the form field name for CSRF tokens
	CSRFFormField = "csrf_token"
	// CSRFTokenMaxAge is the maximum age of CSRF tokens
	CSRFTokenMaxAge = 24 * time.Hour
)

// CSRFMiddleware provides CSRF protection
type CSRFMiddleware struct {
	logger    logger.Logger
	tokens    map[string]time.Time
	tokensMux sync.RWMutex
	
	// Configuration
	secureCookie bool
	sameSite     http.SameSite
}

// NewCSRFMiddleware creates a new CSRF middleware
func NewCSRFMiddleware(logger logger.Logger, secureCookie bool) *CSRFMiddleware {
	csrf := &CSRFMiddleware{
		logger:       logger,
		tokens:       make(map[string]time.Time),
		secureCookie: secureCookie,
		sameSite:     http.SameSiteLaxMode,
	}
	
	// Start cleanup goroutine
	go csrf.cleanupExpiredTokens()
	
	return csrf
}

// generateToken generates a cryptographically secure random token
func (c *CSRFMiddleware) generateToken() (string, error) {
	bytes := make([]byte, CSRFTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate CSRF token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// storeToken stores a token with its creation time
func (c *CSRFMiddleware) storeToken(token string) {
	c.tokensMux.Lock()
	defer c.tokensMux.Unlock()
	c.tokens[token] = time.Now()
}

// validateToken validates a CSRF token
func (c *CSRFMiddleware) validateToken(token string) bool {
	if token == "" {
		return false
	}
	
	c.tokensMux.RLock()
	defer c.tokensMux.RUnlock()
	
	createdAt, exists := c.tokens[token]
	if !exists {
		return false
	}
	
	// Check if token is expired
	if time.Since(createdAt) > CSRFTokenMaxAge {
		delete(c.tokens, token)
		return false
	}
	
	return true
}

// getTokenFromRequest extracts CSRF token from request
func (c *CSRFMiddleware) getTokenFromRequest(r *http.Request) string {
	// Try header first
	if token := r.Header.Get(CSRFTokenHeader); token != "" {
		return token
	}
	
	// Try form field
	if err := r.ParseForm(); err == nil {
		if token := r.FormValue(CSRFFormField); token != "" {
			return token
		}
	}
	
	return ""
}

// getTokenFromCookie extracts CSRF token from cookie
func (c *CSRFMiddleware) getTokenFromCookie(r *http.Request) string {
	cookie, err := r.Cookie(CSRFCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// setTokenCookie sets the CSRF token cookie
func (c *CSRFMiddleware) setTokenCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     CSRFCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(CSRFTokenMaxAge.Seconds()),
		HttpOnly: false, // JavaScript needs access to read for AJAX requests
		Secure:   c.secureCookie,
		SameSite: c.sameSite,
	}
	http.SetCookie(w, cookie)
}

// cleanupExpiredTokens removes expired tokens periodically
func (c *CSRFMiddleware) cleanupExpiredTokens() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for range ticker.C {
		c.tokensMux.Lock()
		now := time.Now()
		for token, createdAt := range c.tokens {
			if now.Sub(createdAt) > CSRFTokenMaxAge {
				delete(c.tokens, token)
			}
		}
		c.tokensMux.Unlock()
	}
}

// isSafeMethod checks if HTTP method is safe (doesn't modify state)
func (c *CSRFMiddleware) isSafeMethod(method string) bool {
	switch strings.ToUpper(method) {
	case "GET", "HEAD", "OPTIONS", "TRACE":
		return true
	default:
		return false
	}
}

// Middleware returns the CSRF middleware function
func (c *CSRFMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For safe methods, just ensure we have a token cookie
		if c.isSafeMethod(r.Method) {
			// Check if we already have a valid token
			cookieToken := c.getTokenFromCookie(r)
			if cookieToken == "" || !c.validateToken(cookieToken) {
				// Generate new token and set cookie
				token, err := c.generateToken()
				if err != nil {
					c.logger.Error("csrf.token_generation_failed", map[string]interface{}{
						"error": err.Error(),
					})
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				c.storeToken(token)
				c.setTokenCookie(w, token)
			}
			next.ServeHTTP(w, r)
			return
		}
		
		// For unsafe methods (POST, PUT, DELETE, etc.), validate CSRF token
		cookieToken := c.getTokenFromCookie(r)
		requestToken := c.getTokenFromRequest(r)
		
		// Validate both tokens exist and match
		if cookieToken == "" {
			c.logger.Warn("csrf.missing_cookie_token", map[string]interface{}{
				"method": r.Method,
				"path":   r.URL.Path,
				"remote": r.RemoteAddr,
			})
			http.Error(w, "CSRF token missing from cookie", http.StatusForbidden)
			return
		}
		
		if requestToken == "" {
			c.logger.Warn("csrf.missing_request_token", map[string]interface{}{
				"method": r.Method,
				"path":   r.URL.Path,
				"remote": r.RemoteAddr,
			})
			http.Error(w, "CSRF token missing from request", http.StatusForbidden)
			return
		}
		
		// Validate cookie token
		if !c.validateToken(cookieToken) {
			c.logger.Warn("csrf.invalid_cookie_token", map[string]interface{}{
				"method": r.Method,
				"path":   r.URL.Path,
				"remote": r.RemoteAddr,
			})
			http.Error(w, "Invalid CSRF token in cookie", http.StatusForbidden)
			return
		}
		
		// Compare tokens using constant-time comparison
		if subtle.ConstantTimeCompare([]byte(cookieToken), []byte(requestToken)) != 1 {
			c.logger.Warn("csrf.token_mismatch", map[string]interface{}{
				"method": r.Method,
				"path":   r.URL.Path,
				"remote": r.RemoteAddr,
			})
			http.Error(w, "CSRF token mismatch", http.StatusForbidden)
			return
		}
		
		c.logger.Debug("csrf.token_validated", map[string]interface{}{
			"method": r.Method,
			"path":   r.URL.Path,
		})
		
		next.ServeHTTP(w, r)
	})
}

// GetToken returns the current CSRF token for the request
func (c *CSRFMiddleware) GetToken(r *http.Request) string {
	return c.getTokenFromCookie(r)
}