package security

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/security/audit"
)

func TestNewHTTPSEnforcer(t *testing.T) {
	config := DefaultHTTPSConfig()
	enforcer := NewHTTPSEnforcer(config, nil)

	if enforcer == nil {
		t.Fatal("NewHTTPSEnforcer returned nil")
	}

	if enforcer.config.RequireHTTPS != config.RequireHTTPS {
		t.Errorf("Expected RequireHTTPS %v, got %v", config.RequireHTTPS, enforcer.config.RequireHTTPS)
	}
}

func TestValidateURL(t *testing.T) {
	config := DefaultHTTPSConfig()
	enforcer := NewHTTPSEnforcer(config, nil)

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid HTTPS URL",
			url:     "https://api.trakt.tv/users/me",
			wantErr: false,
		},
		{
			name:    "valid HTTPS URL with port",
			url:     "https://api.trakt.tv:443/users/me",
			wantErr: false,
		},
		{
			name:    "HTTP URL when HTTPS required",
			url:     "http://api.trakt.tv/users/me",
			wantErr: true,
		},
		{
			name:    "blocked localhost",
			url:     "https://localhost:8080/api",
			wantErr: true,
		},
		{
			name:    "blocked 127.0.0.1",
			url:     "https://127.0.0.1:8080/api",
			wantErr: true,
		},
		{
			name:    "invalid URL format",
			url:     "not-a-url",
			wantErr: true,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "disallowed host",
			url:     "https://malicious.com/api",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := enforcer.ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateSecureClient(t *testing.T) {
	config := DefaultHTTPSConfig()
	enforcer := NewHTTPSEnforcer(config, nil)

	client := enforcer.CreateSecureClient()

	if client == nil {
		t.Fatal("CreateSecureClient returned nil")
	}

	// Check timeout
	expectedTimeout := time.Duration(30) * time.Second
	if client.Timeout != expectedTimeout {
		t.Errorf("Expected timeout %v, got %v", expectedTimeout, client.Timeout)
	}

	// Check that transport is configured
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected http.Transport")
	}

	if transport.TLSClientConfig == nil {
		t.Fatal("Expected TLS config to be set")
	}

	// Check TLS version
	if transport.TLSClientConfig.MinVersion != config.TLSMinVersion {
		t.Errorf("Expected TLS min version %d, got %d", config.TLSMinVersion, transport.TLSClientConfig.MinVersion)
	}

	// Check that insecure skip verify is false in production
	if transport.TLSClientConfig.InsecureSkipVerify && !config.AllowInsecure {
		t.Error("InsecureSkipVerify should be false in production")
	}
}

func TestSecureRequest(t *testing.T) {
	config := DefaultHTTPSConfig()
	enforcer := NewHTTPSEnforcer(config, nil)

	// Create test request
	req, err := http.NewRequest("GET", "https://api.trakt.tv/users/me", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Apply security headers
	err = enforcer.SecureRequest(req)
	if err != nil {
		t.Fatalf("SecureRequest failed: %v", err)
	}

	// Check that security headers were added
	expectedHeaders := map[string]string{
		"Cache-Control":   "no-cache",
		"X-Requested-With": "XMLHttpRequest",
	}

	for header, expectedValue := range expectedHeaders {
		if req.Header.Get(header) != expectedValue {
			t.Errorf("Expected header %s: %s, got: %s", header, expectedValue, req.Header.Get(header))
		}
	}

	// Check User-Agent is set
	userAgent := req.Header.Get("User-Agent")
	if userAgent == "" {
		t.Error("Expected User-Agent header to be set")
	}
}

func TestValidateResponse(t *testing.T) {
	config := DefaultHTTPSConfig()
	enforcer := NewHTTPSEnforcer(config, nil)

	// Create test server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	// Create request to test server
	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Make request
	client := server.Client()
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Validate response
	err = enforcer.ValidateResponse(resp)
	if err != nil {
		t.Errorf("ValidateResponse failed: %v", err)
	}
}

func TestValidateResponseSuspiciousContent(t *testing.T) {
	config := DefaultHTTPSConfig()
	enforcer := NewHTTPSEnforcer(config, nil)

	// Create test server with suspicious content
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<script>alert('xss')</script>`))
	}))
	defer server.Close()

	// Create request to test server
	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Make request
	client := server.Client()
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Validate response - should detect suspicious content
	err = enforcer.ValidateResponse(resp)
	if err == nil {
		t.Error("Expected ValidateResponse to detect suspicious content")
	}
}

func TestHTTPSConfigDefaults(t *testing.T) {
	config := DefaultHTTPSConfig()

	// Test default values
	if !config.RequireHTTPS {
		t.Error("Expected RequireHTTPS to be true by default")
	}

	if config.AllowInsecure {
		t.Error("Expected AllowInsecure to be false by default")
	}

	if config.TLSMinVersion != 771 { // TLS 1.2
		t.Errorf("Expected TLS min version 771 (TLS 1.2), got %d", config.TLSMinVersion)
	}

	expectedTimeout := 30 * time.Second
	if config.Timeout != expectedTimeout {
		t.Errorf("Expected timeout %v, got %v", expectedTimeout, config.Timeout)
	}

	if config.MaxRedirects != 5 {
		t.Errorf("Expected max redirects 5, got %d", config.MaxRedirects)
	}

	if !config.EnableHSTS {
		t.Error("Expected EnableHSTS to be true by default")
	}

	// Test allowed hosts
	expectedHosts := []string{"api.trakt.tv", "api.themoviedb.org"}
	for _, expected := range expectedHosts {
		found := false
		for _, allowed := range config.AllowedHosts {
			if allowed == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected allowed host %s not found", expected)
		}
	}

	// Test blocked hosts
	expectedBlocked := []string{"localhost", "127.0.0.1"}
	for _, expected := range expectedBlocked {
		found := false
		for _, blocked := range config.BlockedHosts {
			if blocked == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected blocked host %s not found", expected)
		}
	}
}

func TestHTTPSWithAuditLogging(t *testing.T) {
	// Create audit logger
	auditConfig := audit.Config{
		LogLevel:     "debug",
		OutputFormat: "json",
		LogFile:      "/tmp/https_test_audit.log",
	}
	auditLogger, err := audit.NewLogger(auditConfig)
	if err != nil {
		t.Fatal(err)
	}

	config := DefaultHTTPSConfig()
	enforcer := NewHTTPSEnforcer(config, auditLogger)

	// Test URL validation with audit logging
	err = enforcer.ValidateURL("http://malicious.com")
	if err == nil {
		t.Error("Expected URL validation to fail")
	}

	// Test blocked host with audit logging
	err = enforcer.ValidateURL("https://localhost:8080")
	if err == nil {
		t.Error("Expected blocked host validation to fail")
	}
}

func TestRedirectValidation(t *testing.T) {
	config := DefaultHTTPSConfig()
	config.MaxRedirects = 2
	enforcer := NewHTTPSEnforcer(config, nil)

	// Create test server with redirects
	redirectCount := 0
	var testServer *httptest.Server
	testServer = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if redirectCount < 3 {
			redirectCount++
			http.Redirect(w, r, testServer.URL+"/redirect", http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("final destination"))
	}))
	defer testServer.Close()

	client := enforcer.CreateSecureClient()

	// This should fail due to too many redirects
	_, err := client.Get(testServer.URL)
	if err == nil {
		t.Error("Expected request to fail due to too many redirects")
	}
}

func TestContextTimeout(t *testing.T) {
	config := DefaultHTTPSConfig()
	config.Timeout = 1 * time.Millisecond // Very short timeout
	enforcer := NewHTTPSEnforcer(config, nil)

	// Create test server with delay
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Longer than timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	client := enforcer.CreateSecureClient()

	// Create request with context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", testServer.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	// This should timeout
	_, err = client.Do(req)
	if err == nil {
		t.Error("Expected request to timeout")
	}
}

func TestCipherSuiteValidation(t *testing.T) {
	config := DefaultHTTPSConfig()
	enforcer := NewHTTPSEnforcer(config, nil)

	client := enforcer.CreateSecureClient()
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected http.Transport")
	}

	// Check that cipher suites are configured
	if len(transport.TLSClientConfig.CipherSuites) == 0 {
		t.Error("Expected cipher suites to be configured")
	}

	// Check for strong cipher suites
	strongCiphers := []uint16{
		0x1301, // TLS_AES_128_GCM_SHA256
		0x1302, // TLS_AES_256_GCM_SHA384
		0x1303, // TLS_CHACHA20_POLY1305_SHA256
		0xc02f, // TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
		0xc030, // TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
	}

	foundStrong := false
	for _, configured := range transport.TLSClientConfig.CipherSuites {
		for _, strong := range strongCiphers {
			if configured == strong {
				foundStrong = true
				break
			}
		}
		if foundStrong {
			break
		}
	}

	if !foundStrong {
		t.Error("Expected at least one strong cipher suite to be configured")
	}
} 