package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

// Mock logger for testing
type mockLogger struct {
	logs []map[string]interface{}
}

func (m *mockLogger) Debug(msg string, fields ...map[string]interface{}) {
	var fieldsMap map[string]interface{}
	if len(fields) > 0 {
		fieldsMap = fields[0]
	}
	m.logs = append(m.logs, map[string]interface{}{"level": "debug", "msg": msg, "fields": fieldsMap})
}

func (m *mockLogger) Info(msg string, fields ...map[string]interface{}) {
	var fieldsMap map[string]interface{}
	if len(fields) > 0 {
		fieldsMap = fields[0]
	}
	m.logs = append(m.logs, map[string]interface{}{"level": "info", "msg": msg, "fields": fieldsMap})
}

func (m *mockLogger) Warn(msg string, fields ...map[string]interface{}) {
	var fieldsMap map[string]interface{}
	if len(fields) > 0 {
		fieldsMap = fields[0]
	}
	m.logs = append(m.logs, map[string]interface{}{"level": "warn", "msg": msg, "fields": fieldsMap})
}

func (m *mockLogger) Error(msg string, fields ...map[string]interface{}) {
	var fieldsMap map[string]interface{}
	if len(fields) > 0 {
		fieldsMap = fields[0]
	}
	m.logs = append(m.logs, map[string]interface{}{"level": "error", "msg": msg, "fields": fieldsMap})
}

func (m *mockLogger) Fatal(msg string, fields ...map[string]interface{}) {
	var fieldsMap map[string]interface{}
	if len(fields) > 0 {
		fieldsMap = fields[0]
	}
	m.logs = append(m.logs, map[string]interface{}{"level": "fatal", "msg": msg, "fields": fieldsMap})
}

func (m *mockLogger) Debugf(msg string, data map[string]interface{}) {
	m.logs = append(m.logs, map[string]interface{}{"level": "debug", "msg": msg, "fields": data})
}

func (m *mockLogger) Infof(msg string, data map[string]interface{}) {
	m.logs = append(m.logs, map[string]interface{}{"level": "info", "msg": msg, "fields": data})
}

func (m *mockLogger) Warnf(msg string, data map[string]interface{}) {
	m.logs = append(m.logs, map[string]interface{}{"level": "warn", "msg": msg, "fields": data})
}

func (m *mockLogger) Errorf(msg string, data map[string]interface{}) {
	m.logs = append(m.logs, map[string]interface{}{"level": "error", "msg": msg, "fields": data})
}

func (m *mockLogger) SetLogLevel(level string) {}

func (m *mockLogger) SetLogFile(path string) error {
	return nil
}

func (m *mockLogger) SetTranslator(t logger.Translator) {}

// Test handler that captures request
type testHandler struct {
	called bool
}

func (t *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.called = true
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("test response"))
}

func TestNewCSRFMiddleware(t *testing.T) {
	log := &mockLogger{}
	
	csrf := NewCSRFMiddleware(log, true)
	
	if csrf == nil {
		t.Fatal("Expected CSRF middleware to be created")
	}
	
	if csrf.logger != log {
		t.Error("Expected logger to be set")
	}
	
	if !csrf.secureCookie {
		t.Error("Expected secure cookie to be enabled")
	}
	
	if csrf.sameSite != http.SameSiteLaxMode {
		t.Error("Expected SameSite to be Lax")
	}
	
	if csrf.tokens == nil {
		t.Error("Expected tokens map to be initialized")
	}
}

func TestCSRFGenerateToken(t *testing.T) {
	log := &mockLogger{}
	csrf := NewCSRFMiddleware(log, false)
	
	token1, err := csrf.generateToken()
	if err != nil {
		t.Fatalf("Expected no error generating token, got: %v", err)
	}
	
	if len(token1) == 0 {
		t.Error("Expected non-empty token")
	}
	
	// Generate another token and ensure they're different
	token2, err := csrf.generateToken()
	if err != nil {
		t.Fatalf("Expected no error generating second token, got: %v", err)
	}
	
	if token1 == token2 {
		t.Error("Expected tokens to be different")
	}
}

func TestCSRFStoreAndValidateToken(t *testing.T) {
	log := &mockLogger{}
	csrf := NewCSRFMiddleware(log, false)
	
	token := "test-token"
	csrf.storeToken(token)
	
	// Valid token should validate
	if !csrf.validateToken(token) {
		t.Error("Expected valid token to validate")
	}
	
	// Empty token should not validate
	if csrf.validateToken("") {
		t.Error("Expected empty token to be invalid")
	}
	
	// Non-existent token should not validate
	if csrf.validateToken("non-existent") {
		t.Error("Expected non-existent token to be invalid")
	}
}

func TestCSRFTokenExpiration(t *testing.T) {
	log := &mockLogger{}
	csrf := NewCSRFMiddleware(log, false)
	
	// Manually add an expired token
	expiredToken := "expired-token"
	csrf.tokensMux.Lock()
	csrf.tokens[expiredToken] = time.Now().Add(-25 * time.Hour) // Older than CSRFTokenMaxAge
	csrf.tokensMux.Unlock()
	
	// Expired token should not validate and should be cleaned up
	if csrf.validateToken(expiredToken) {
		t.Error("Expected expired token to be invalid")
	}
	
	// Token should be removed from map
	csrf.tokensMux.RLock()
	_, exists := csrf.tokens[expiredToken]
	csrf.tokensMux.RUnlock()
	
	if exists {
		t.Error("Expected expired token to be removed from map")
	}
}

func TestCSRFGetTokenFromRequest(t *testing.T) {
	log := &mockLogger{}
	csrf := NewCSRFMiddleware(log, false)
	
	// Test token in header
	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set(CSRFTokenHeader, "header-token")
	
	token := csrf.getTokenFromRequest(req)
	if token != "header-token" {
		t.Errorf("Expected 'header-token', got '%s'", token)
	}
	
	// Test token in form field (when no header)
	formData := url.Values{}
	formData.Set(CSRFFormField, "form-token")
	req2 := httptest.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	token = csrf.getTokenFromRequest(req2)
	if token != "form-token" {
		t.Errorf("Expected 'form-token', got '%s'", token)
	}
	
	// Test no token
	req3 := httptest.NewRequest("POST", "/test", nil)
	token = csrf.getTokenFromRequest(req3)
	if token != "" {
		t.Errorf("Expected empty token, got '%s'", token)
	}
}

func TestCSRFGetTokenFromCookie(t *testing.T) {
	log := &mockLogger{}
	csrf := NewCSRFMiddleware(log, false)
	
	// Test with cookie
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  CSRFCookieName,
		Value: "cookie-token",
	})
	
	token := csrf.getTokenFromCookie(req)
	if token != "cookie-token" {
		t.Errorf("Expected 'cookie-token', got '%s'", token)
	}
	
	// Test without cookie
	req2 := httptest.NewRequest("GET", "/test", nil)
	token = csrf.getTokenFromCookie(req2)
	if token != "" {
		t.Errorf("Expected empty token, got '%s'", token)
	}
}

func TestCSRFSetTokenCookie(t *testing.T) {
	log := &mockLogger{}
	csrf := NewCSRFMiddleware(log, true) // secure = true
	
	w := httptest.NewRecorder()
	csrf.setTokenCookie(w, "test-cookie-token")
	
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Expected 1 cookie, got %d", len(cookies))
	}
	
	cookie := cookies[0]
	if cookie.Name != CSRFCookieName {
		t.Errorf("Expected cookie name '%s', got '%s'", CSRFCookieName, cookie.Name)
	}
	
	if cookie.Value != "test-cookie-token" {
		t.Errorf("Expected cookie value 'test-cookie-token', got '%s'", cookie.Value)
	}
	
	if cookie.Path != "/" {
		t.Errorf("Expected cookie path '/', got '%s'", cookie.Path)
	}
	
	if !cookie.Secure {
		t.Error("Expected cookie to be secure")
	}
	
	if cookie.HttpOnly {
		t.Error("Expected cookie to not be HttpOnly (JS needs access)")
	}
	
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Error("Expected cookie SameSite to be Lax")
	}
}

func TestCSRFIsSafeMethod(t *testing.T) {
	log := &mockLogger{}
	csrf := NewCSRFMiddleware(log, false)
	
	safeMethods := []string{"GET", "HEAD", "OPTIONS", "TRACE", "get", "head"}
	unsafeMethods := []string{"POST", "PUT", "DELETE", "PATCH", "post", "put"}
	
	for _, method := range safeMethods {
		if !csrf.isSafeMethod(method) {
			t.Errorf("Expected '%s' to be safe method", method)
		}
	}
	
	for _, method := range unsafeMethods {
		if csrf.isSafeMethod(method) {
			t.Errorf("Expected '%s' to be unsafe method", method)
		}
	}
}

func TestCSRFMiddlewareSafeMethod(t *testing.T) {
	log := &mockLogger{}
	csrf := NewCSRFMiddleware(log, false)
	handler := &testHandler{}
	
	middleware := csrf.Middleware(handler)
	
	// Test GET request (safe method)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	middleware.ServeHTTP(w, req)
	
	if !handler.called {
		t.Error("Expected handler to be called for safe method")
	}
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	// Check that cookie was set
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Error("Expected CSRF cookie to be set")
	}
}

func TestCSRFMiddlewareUnsafeMethodValid(t *testing.T) {
	log := &mockLogger{}
	csrf := NewCSRFMiddleware(log, false)
	handler := &testHandler{}
	
	// Generate and store a valid token
	token, _ := csrf.generateToken()
	csrf.storeToken(token)
	
	middleware := csrf.Middleware(handler)
	
	// Test POST request with valid CSRF token
	formData := url.Values{}
	formData.Set(CSRFFormField, token)
	req := httptest.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{
		Name:  CSRFCookieName,
		Value: token,
	})
	
	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)
	
	if !handler.called {
		t.Error("Expected handler to be called for valid CSRF token")
	}
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestCSRFMiddlewareUnsafeMethodMissingCookie(t *testing.T) {
	log := &mockLogger{}
	csrf := NewCSRFMiddleware(log, false)
	handler := &testHandler{}
	
	middleware := csrf.Middleware(handler)
	
	// Test POST request without cookie token
	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set(CSRFTokenHeader, "some-token")
	
	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)
	
	if handler.called {
		t.Error("Expected handler not to be called when cookie token is missing")
	}
	
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestCSRFMiddlewareUnsafeMethodMissingRequest(t *testing.T) {
	log := &mockLogger{}
	csrf := NewCSRFMiddleware(log, false)
	handler := &testHandler{}
	
	middleware := csrf.Middleware(handler)
	
	// Test POST request without request token
	req := httptest.NewRequest("POST", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  CSRFCookieName,
		Value: "cookie-token",
	})
	
	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)
	
	if handler.called {
		t.Error("Expected handler not to be called when request token is missing")
	}
	
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestCSRFMiddlewareTokenMismatch(t *testing.T) {
	log := &mockLogger{}
	csrf := NewCSRFMiddleware(log, false)
	handler := &testHandler{}
	
	// Store a valid token
	validToken, _ := csrf.generateToken()
	csrf.storeToken(validToken)
	
	middleware := csrf.Middleware(handler)
	
	// Test POST request with mismatched tokens
	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set(CSRFTokenHeader, "wrong-token")
	req.AddCookie(&http.Cookie{
		Name:  CSRFCookieName,
		Value: validToken,
	})
	
	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)
	
	if handler.called {
		t.Error("Expected handler not to be called when tokens don't match")
	}
	
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestCSRFGetToken(t *testing.T) {
	log := &mockLogger{}
	csrf := NewCSRFMiddleware(log, false)
	
	// Test with cookie
	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  CSRFCookieName,
		Value: "test-token",
	})
	
	token := csrf.GetToken(req)
	if token != "test-token" {
		t.Errorf("Expected 'test-token', got '%s'", token)
	}
}

// Security Headers Tests

func TestNewSecurityHeaders(t *testing.T) {
	log := &mockLogger{}
	
	security := NewSecurityHeaders(log, true, true)
	
	if security == nil {
		t.Fatal("Expected security headers middleware to be created")
	}
	
	if security.logger != log {
		t.Error("Expected logger to be set")
	}
	
	if !security.isHTTPS {
		t.Error("Expected HTTPS to be enabled")
	}
	
	if !security.enableCSP {
		t.Error("Expected CSP to be enabled")
	}
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	log := &mockLogger{}
	security := NewSecurityHeaders(log, true, true)
	handler := &testHandler{}
	
	middleware := security.Middleware(handler)
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	middleware.ServeHTTP(w, req)
	
	if !handler.called {
		t.Error("Expected handler to be called")
	}
	
	// Check security headers
	headers := w.Header()
	
	if headers.Get("X-Frame-Options") != "DENY" {
		t.Error("Expected X-Frame-Options to be DENY")
	}
	
	if headers.Get("X-Content-Type-Options") != "nosniff" {
		t.Error("Expected X-Content-Type-Options to be nosniff")
	}
	
	if headers.Get("X-XSS-Protection") != "1; mode=block" {
		t.Error("Expected X-XSS-Protection to be '1; mode=block'")
	}
	
	if headers.Get("Referrer-Policy") != "strict-origin-when-cross-origin" {
		t.Error("Expected Referrer-Policy to be 'strict-origin-when-cross-origin'")
	}
	
	if headers.Get("X-Permitted-Cross-Domain-Policies") != "none" {
		t.Error("Expected X-Permitted-Cross-Domain-Policies to be none")
	}
	
	if headers.Get("X-Download-Options") != "noopen" {
		t.Error("Expected X-Download-Options to be noopen")
	}
	
	// Check HTTPS-specific headers
	if headers.Get("Strict-Transport-Security") == "" {
		t.Error("Expected Strict-Transport-Security header to be set for HTTPS")
	}
	
	// Check CSP header
	if headers.Get("Content-Security-Policy") == "" {
		t.Error("Expected Content-Security-Policy header to be set")
	}
}

func TestSecurityHeadersSecuritySensitivePath(t *testing.T) {
	log := &mockLogger{}
	security := NewSecurityHeaders(log, false, false)
	handler := &testHandler{}
	
	middleware := security.Middleware(handler)
	
	// Test security-sensitive path
	req := httptest.NewRequest("GET", "/auth/login", nil)
	w := httptest.NewRecorder()
	
	middleware.ServeHTTP(w, req)
	
	headers := w.Header()
	
	// Check cache control headers for sensitive paths
	if headers.Get("Cache-Control") != "no-cache, no-store, must-revalidate" {
		t.Error("Expected Cache-Control to be set for security-sensitive path")
	}
	
	if headers.Get("Pragma") != "no-cache" {
		t.Error("Expected Pragma to be set for security-sensitive path")
	}
	
	if headers.Get("Expires") != "0" {
		t.Error("Expected Expires to be set for security-sensitive path")
	}
}

func TestIsSecuritySensitivePath(t *testing.T) {
	log := &mockLogger{}
	security := NewSecurityHeaders(log, false, false)
	
	sensitivePaths := []string{
		"/auth",
		"/auth/login",
		"/callback",
		"/config",
		"/api/test",
		"/api/data/sensitive",
	}
	
	normalPaths := []string{
		"/",
		"/home",
		"/public",
		"/static/css/style.css",
	}
	
	for _, path := range sensitivePaths {
		if !security.isSecuritySensitivePath(path) {
			t.Errorf("Expected '%s' to be security sensitive", path)
		}
	}
	
	for _, path := range normalPaths {
		if security.isSecuritySensitivePath(path) {
			t.Errorf("Expected '%s' not to be security sensitive", path)
		}
	}
}

func TestBuildContentSecurityPolicy(t *testing.T) {
	log := &mockLogger{}
	security := NewSecurityHeaders(log, true, true)
	
	req := httptest.NewRequest("GET", "/test", nil)
	csp := security.buildContentSecurityPolicy(req)
	
	if csp == "" {
		t.Error("Expected CSP to be non-empty")
	}
	
	// Check for key CSP directives
	expectedDirectives := []string{
		"default-src 'self'",
		"script-src 'self' 'unsafe-inline'",
		"style-src 'self' 'unsafe-inline'",
		"img-src 'self' data:",
		"object-src 'none'",
		"frame-ancestors 'none'",
		"upgrade-insecure-requests",
	}
	
	for _, directive := range expectedDirectives {
		if !strings.Contains(csp, directive) {
			t.Errorf("Expected CSP to contain '%s'", directive)
		}
	}
}

func TestDefaultSecurityConfig(t *testing.T) {
	config := DefaultSecurityConfig()
	
	if !config.EnableCSP {
		t.Error("Expected CSP to be enabled by default")
	}
	
	if !config.EnableHSTS {
		t.Error("Expected HSTS to be enabled by default")
	}
	
	if config.HSTSMaxAge != 31536000 {
		t.Errorf("Expected HSTS max age to be 31536000, got %d", config.HSTSMaxAge)
	}
	
	if !config.HSTSIncludeSubdomains {
		t.Error("Expected HSTS to include subdomains by default")
	}
	
	if config.HSTSPreload {
		t.Error("Expected HSTS preload to be disabled by default")
	}
	
	if !config.EnableClickjackProtection {
		t.Error("Expected clickjack protection to be enabled by default")
	}
	
	if !config.EnableContentTypeOptions {
		t.Error("Expected content type options to be enabled by default")
	}
	
	if !config.EnableXSSProtection {
		t.Error("Expected XSS protection to be enabled by default")
	}
	
	if config.ReferrerPolicy != "strict-origin-when-cross-origin" {
		t.Errorf("Expected referrer policy to be 'strict-origin-when-cross-origin', got '%s'", config.ReferrerPolicy)
	}
	
	if config.CustomHeaders == nil {
		t.Error("Expected custom headers map to be initialized")
	}
}

func TestNewSecurityHeadersWithConfig(t *testing.T) {
	log := &mockLogger{}
	config := SecurityHeadersConfig{
		EnableCSP: false,
		EnableHSTS: true,
	}
	
	security := NewSecurityHeadersWithConfig(log, true, config)
	
	if security == nil {
		t.Fatal("Expected security headers middleware to be created")
	}
	
	if security.logger != log {
		t.Error("Expected logger to be set")
	}
	
	if !security.isHTTPS {
		t.Error("Expected HTTPS to be enabled")
	}
	
	if security.enableCSP {
		t.Error("Expected CSP to be disabled based on config")
	}
}