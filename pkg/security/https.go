package security

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/security/audit"
)

// HTTPSEnforcer provides HTTPS enforcement and secure HTTP client configuration
type HTTPSEnforcer struct {
	auditLog *audit.Logger
	config   HTTPSConfig
}

// HTTPSConfig holds configuration for HTTPS enforcement
type HTTPSConfig struct {
	RequireHTTPS       bool          `toml:"require_https"`
	AllowInsecure      bool          `toml:"allow_insecure"`      // For development only
	TLSMinVersion      uint16        `toml:"tls_min_version"`     // TLS 1.2 minimum
	Timeout            time.Duration `toml:"timeout"`             // Request timeout
	MaxRedirects       int           `toml:"max_redirects"`       // Maximum redirects
	AllowedHosts       []string      `toml:"allowed_hosts"`       // Whitelist of allowed hosts
	BlockedHosts       []string      `toml:"blocked_hosts"`       // Blacklist of blocked hosts
	UserAgent          string        `toml:"user_agent"`          // Custom user agent
	EnableHSTS         bool          `toml:"enable_hsts"`         // HTTP Strict Transport Security
}

// DefaultHTTPSConfig returns secure default HTTPS configuration
func DefaultHTTPSConfig() HTTPSConfig {
	return HTTPSConfig{
		RequireHTTPS:  true,
		AllowInsecure: false,
		TLSMinVersion: tls.VersionTLS12,
		Timeout:       time.Second * 30,
		MaxRedirects:  5,
		AllowedHosts: []string{
			"api.trakt.tv",
			"api.themoviedb.org",
			"www.omdbapi.com",
		},
		BlockedHosts: []string{
			"localhost",
			"127.0.0.1",
			"0.0.0.0",
		},
		UserAgent:  "Export_Trakt_4_Letterboxd/1.0 (+https://github.com/JohanDevl/Export_Trakt_4_Letterboxd)",
		EnableHSTS: true,
	}
}

// NewHTTPSEnforcer creates a new HTTPS enforcer
func NewHTTPSEnforcer(config HTTPSConfig, auditLog *audit.Logger) *HTTPSEnforcer {
	return &HTTPSEnforcer{
		auditLog: auditLog,
		config:   config,
	}
}

// CreateSecureClient creates a secure HTTP client with enforced HTTPS
func (h *HTTPSEnforcer) CreateSecureClient() *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion:         h.config.TLSMinVersion,
			InsecureSkipVerify: h.config.AllowInsecure, // Only for development
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			},
		},
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 2,
		IdleConnTimeout:     time.Minute * 2,
		DisableCompression:  false,
		ForceAttemptHTTP2:   true,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   h.config.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Limit redirects
			if len(via) >= h.config.MaxRedirects {
				return fmt.Errorf("stopped after %d redirects", h.config.MaxRedirects)
			}

			// Ensure redirects also use HTTPS
			if err := h.ValidateURL(req.URL.String()); err != nil {
				h.logSecurityViolation("insecure_redirect", req.URL.String(), err.Error())
				return fmt.Errorf("insecure redirect blocked: %w", err)
			}

			return nil
		},
	}

	return client
}

// ValidateURL validates that a URL meets security requirements
func (h *HTTPSEnforcer) ValidateURL(urlStr string) error {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Require HTTPS
	if h.config.RequireHTTPS && parsedURL.Scheme != "https" {
		h.logSecurityViolation("http_not_allowed", urlStr, "HTTP protocol not allowed")
		return fmt.Errorf("HTTPS required, got: %s", parsedURL.Scheme)
	}

	// Check against blocked hosts
	for _, blockedHost := range h.config.BlockedHosts {
		if strings.Contains(parsedURL.Host, blockedHost) {
			h.logSecurityViolation("blocked_host", urlStr, fmt.Sprintf("Host %s is blocked", blockedHost))
			return fmt.Errorf("access to host %s is blocked", blockedHost)
		}
	}

	// Check against allowed hosts (if whitelist is configured)
	if len(h.config.AllowedHosts) > 0 {
		allowed := false
		for _, allowedHost := range h.config.AllowedHosts {
			if strings.Contains(parsedURL.Host, allowedHost) {
				allowed = true
				break
			}
		}
		if !allowed {
			h.logSecurityViolation("host_not_allowed", urlStr, "Host not in allowed list")
			return fmt.Errorf("host %s is not allowed", parsedURL.Host)
		}
	}

	// Check for suspicious patterns
	if h.containsSuspiciousPatterns(urlStr) {
		h.logSecurityViolation("suspicious_url", urlStr, "URL contains suspicious patterns")
		return fmt.Errorf("URL contains suspicious patterns")
	}

	return nil
}

// SecureRequest modifies an HTTP request to add security headers
func (h *HTTPSEnforcer) SecureRequest(req *http.Request) error {
	// Validate the URL first
	if err := h.ValidateURL(req.URL.String()); err != nil {
		return err
	}

	// Set secure headers
	if h.config.UserAgent != "" {
		req.Header.Set("User-Agent", h.config.UserAgent)
	}

	// Add security headers
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")

	// Remove potentially sensitive headers that could leak information
	req.Header.Del("X-Forwarded-For")
	req.Header.Del("X-Real-IP")
	req.Header.Del("X-Forwarded-Proto")

	// Log the request for audit purposes
	if h.auditLog != nil {
		h.auditLog.LogEvent(audit.AuditEvent{
			EventType: audit.DataAccess,
			Severity:  audit.SeverityLow,
			Source:    "https_enforcer",
			Action:    "secure_request",
			Target:    req.URL.Host,
			Result:    "success",
			Message:   fmt.Sprintf("Secure HTTP request to %s", req.URL.String()),
			Details: map[string]interface{}{
				"method":     req.Method,
				"url":        req.URL.String(),
				"host":       req.URL.Host,
				"user_agent": req.Header.Get("User-Agent"),
			},
		})
	}

	return nil
}

// ValidateResponse checks the response for security issues
func (h *HTTPSEnforcer) ValidateResponse(resp *http.Response) error {
	if resp == nil {
		return fmt.Errorf("response is nil")
	}

	// Check for HSTS header if enabled
	if h.config.EnableHSTS && resp.Request.URL.Scheme == "https" {
		hsts := resp.Header.Get("Strict-Transport-Security")
		if hsts == "" {
			h.logSecurityViolation("missing_hsts", resp.Request.URL.String(), "Missing HSTS header")
			// Don't fail the request, just log the warning
		}
	}

	// Check for suspicious response headers
	if h.hasSuspiciousHeaders(resp) {
		h.logSecurityViolation("suspicious_headers", resp.Request.URL.String(), "Response contains suspicious headers")
		// Log but don't fail - might be false positive
	}

	// Validate content type if specified
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" && h.isUnsafeContentType(contentType) {
		h.logSecurityViolation("unsafe_content_type", resp.Request.URL.String(), 
			fmt.Sprintf("Unsafe content type: %s", contentType))
		return fmt.Errorf("unsafe content type: %s", contentType)
	}

	return nil
}

// GetSecurityStats returns HTTPS enforcement statistics
func (h *HTTPSEnforcer) GetSecurityStats() map[string]interface{} {
	return map[string]interface{}{
		"https_required":  h.config.RequireHTTPS,
		"tls_min_version": h.getTLSVersionName(h.config.TLSMinVersion),
		"timeout":         h.config.Timeout.String(),
		"max_redirects":   h.config.MaxRedirects,
		"allowed_hosts":   len(h.config.AllowedHosts),
		"blocked_hosts":   len(h.config.BlockedHosts),
		"hsts_enabled":    h.config.EnableHSTS,
	}
}

// containsSuspiciousPatterns checks for suspicious URL patterns
func (h *HTTPSEnforcer) containsSuspiciousPatterns(urlStr string) bool {
	suspiciousPatterns := []string{
		"../",           // Path traversal
		"..\\",          // Windows path traversal
		"javascript:",   // XSS
		"data:",         // Data URLs
		"file://",       // File protocol
		"ftp://",        // FTP protocol
		"gopher://",     // Gopher protocol
		"<script",       // Script injection
		"eval(",         // Code evaluation
		"alert(",        // Alert boxes
	}

	lowerURL := strings.ToLower(urlStr)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerURL, pattern) {
			return true
		}
	}

	return false
}

// hasSuspiciousHeaders checks for suspicious response headers
func (h *HTTPSEnforcer) hasSuspiciousHeaders(resp *http.Response) bool {
	suspiciousHeaders := map[string][]string{
		"Server": {"nginx/0.", "apache/1.", "iis/4."}, // Very old versions
		"X-Powered-By": {"PHP/4.", "ASP.NET/1."},     // Old tech
	}

	for header, suspiciousValues := range suspiciousHeaders {
		headerValue := resp.Header.Get(header)
		if headerValue != "" {
			for _, suspicious := range suspiciousValues {
				if strings.Contains(strings.ToLower(headerValue), suspicious) {
					return true
				}
			}
		}
	}

	return false
}

// isUnsafeContentType checks if a content type is potentially unsafe
func (h *HTTPSEnforcer) isUnsafeContentType(contentType string) bool {
	unsafeTypes := []string{
		"text/html",           // Could contain scripts
		"application/x-javascript",
		"application/javascript",
		"text/javascript",
		"application/x-shockwave-flash",
		"application/java-archive",
		"application/x-java-archive",
	}

	lowerType := strings.ToLower(strings.Split(contentType, ";")[0])
	for _, unsafe := range unsafeTypes {
		if lowerType == unsafe {
			return true
		}
	}

	return false
}

// getTLSVersionName returns a human-readable TLS version name
func (h *HTTPSEnforcer) getTLSVersionName(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("Unknown (%d)", version)
	}
}

// logSecurityViolation logs security violations
func (h *HTTPSEnforcer) logSecurityViolation(violationType, url, description string) {
	if h.auditLog != nil {
		h.auditLog.LogEvent(audit.AuditEvent{
			EventType: audit.SecurityViolation,
			Severity:  audit.SeverityHigh,
			Source:    "https_enforcer",
			Action:    violationType,
			Target:    url,
			Result:    "blocked",
			Message:   fmt.Sprintf("HTTPS security violation: %s", description),
			Details: map[string]interface{}{
				"violation_type": violationType,
				"url":           url,
				"description":   description,
			},
		})
	}
} 