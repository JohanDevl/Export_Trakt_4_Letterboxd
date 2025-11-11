package middleware

import (
	"net/http"
	"strings"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

// SecurityHeaders middleware adds security headers to all responses
type SecurityHeaders struct {
	logger    logger.Logger
	isHTTPS   bool
	enableCSP bool
}

// NewSecurityHeaders creates a new security headers middleware
func NewSecurityHeaders(logger logger.Logger, isHTTPS bool, enableCSP bool) *SecurityHeaders {
	return &SecurityHeaders{
		logger:    logger,
		isHTTPS:   isHTTPS,
		enableCSP: enableCSP,
	}
}

// Middleware returns the security headers middleware function
func (s *SecurityHeaders) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// X-Frame-Options: Prevent clickjacking attacks
		w.Header().Set("X-Frame-Options", "DENY")
		
		// X-Content-Type-Options: Prevent MIME sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")
		
		// X-XSS-Protection: Enable XSS filtering (legacy browsers)
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		
		// Referrer-Policy: Control referrer information
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// X-Permitted-Cross-Domain-Policies: Restrict cross-domain policies
		w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
		
		// X-Download-Options: Prevent file downloads from opening automatically
		w.Header().Set("X-Download-Options", "noopen")
		
		// Cache-Control for security-sensitive pages
		if s.isSecuritySensitivePath(r.URL.Path) {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		}
		
		// Strict-Transport-Security: Force HTTPS (only when serving over HTTPS)
		if s.isHTTPS {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}
		
		// Content-Security-Policy: Comprehensive CSP
		if s.enableCSP {
			csp := s.buildContentSecurityPolicy(r)
			w.Header().Set("Content-Security-Policy", csp)
		}
		
		s.logger.Debug("security.headers_applied", map[string]interface{}{
			"path":     r.URL.Path,
			"method":   r.Method,
			"is_https": s.isHTTPS,
			"csp_enabled": s.enableCSP,
		})
		
		next.ServeHTTP(w, r)
	})
}

// isSecuritySensitivePath checks if the path contains sensitive information
func (s *SecurityHeaders) isSecuritySensitivePath(path string) bool {
	sensitivePaths := []string{
		"/auth",
		"/callback",
		"/config",
		"/api/",
	}
	
	for _, sensitivePath := range sensitivePaths {
		if strings.HasPrefix(path, sensitivePath) {
			return true
		}
	}
	return false
}

// buildContentSecurityPolicy creates a comprehensive CSP policy
func (s *SecurityHeaders) buildContentSecurityPolicy(r *http.Request) string {
	var cspParts []string
	
	// Default source: only self
	cspParts = append(cspParts, "default-src 'self'")
	
	// Scripts: allow self and inline (needed for templates)
	// Note: 'unsafe-inline' is not ideal but needed for current inline scripts
	// TODO: Move to nonce-based CSP or external scripts
	cspParts = append(cspParts, "script-src 'self' 'unsafe-inline'")
	
	// Styles: allow self and inline styles
	cspParts = append(cspParts, "style-src 'self' 'unsafe-inline'")
	
	// Images: allow self and data URLs (for icons/emojis)
	cspParts = append(cspParts, "img-src 'self' data:")
	
	// Fonts: allow self
	cspParts = append(cspParts, "font-src 'self'")
	
	// Connections: allow self for AJAX requests
	cspParts = append(cspParts, "connect-src 'self'")
	
	// Media: allow self
	cspParts = append(cspParts, "media-src 'self'")
	
	// Objects: disallow plugins
	cspParts = append(cspParts, "object-src 'none'")
	
	// Base URI: only self
	cspParts = append(cspParts, "base-uri 'self'")
	
	// Forms: only submit to self
	cspParts = append(cspParts, "form-action 'self'")
	
	// Frame ancestors: none (same as X-Frame-Options)
	cspParts = append(cspParts, "frame-ancestors 'none'")
	
	// Frame source: none
	cspParts = append(cspParts, "frame-src 'none'")
	
	// Upgrade insecure requests if serving over HTTPS
	if s.isHTTPS {
		cspParts = append(cspParts, "upgrade-insecure-requests")
	}
	
	return strings.Join(cspParts, "; ")
}

// SecurityHeadersConfig contains configuration for security headers
type SecurityHeadersConfig struct {
	EnableCSP               bool
	EnableHSTS              bool
	HSTSMaxAge              int
	HSTSIncludeSubdomains   bool
	HSTSPreload             bool
	EnableClickjackProtection bool
	EnableContentTypeOptions  bool
	EnableXSSProtection       bool
	ReferrerPolicy            string
	CustomHeaders             map[string]string
}

// DefaultSecurityConfig returns a secure default configuration
func DefaultSecurityConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		EnableCSP:                 true,
		EnableHSTS:                true,
		HSTSMaxAge:                31536000, // 1 year
		HSTSIncludeSubdomains:     true,
		HSTSPreload:               false, // Set to true only when ready for preload list
		EnableClickjackProtection: true,
		EnableContentTypeOptions:  true,
		EnableXSSProtection:       true,
		ReferrerPolicy:            "strict-origin-when-cross-origin",
		CustomHeaders:             make(map[string]string),
	}
}

// NewSecurityHeadersWithConfig creates security headers middleware with custom config
func NewSecurityHeadersWithConfig(logger logger.Logger, isHTTPS bool, config SecurityHeadersConfig) *SecurityHeaders {
	return &SecurityHeaders{
		logger:    logger,
		isHTTPS:   isHTTPS,
		enableCSP: config.EnableCSP,
	}
}