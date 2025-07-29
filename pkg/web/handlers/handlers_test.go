package handlers

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/auth"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/security/keyring"
)

func TestDashboardHandler(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Auth: config.AuthConfig{
			CallbackPort: 8080,
		},
	}

	// Create test logger
	log := logger.NewLogger()

	// Create test token manager
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	if err != nil {
		t.Fatalf("Failed to create keyring manager: %v", err)
	}
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)

	// Create test templates
	templates := template.New("")

	// Create dashboard handler
	handler := NewDashboardHandler(cfg, log, tokenManager, templates)

	// Create test request
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	// Call handler
	handler.ServeHTTP(rec, req)

	// Check response (it will be an error due to missing template, but that's expected)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestStatusHandler(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Auth: config.AuthConfig{
			CallbackPort: 8080,
		},
	}

	// Create test logger
	log := logger.NewLogger()

	// Create test token manager
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	if err != nil {
		t.Fatalf("Failed to create keyring manager: %v", err)
	}
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)

	// Create test templates
	templates := template.New("")

	// Create status handler
	handler := NewStatusHandler(cfg, log, tokenManager, templates)

	// Test status page
	req := httptest.NewRequest("GET", "/status", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Check response (it will be an error due to missing template, but that's expected)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestExportsHandler(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Letterboxd: config.LetterboxdConfig{
			ExportDir: "./test_exports",
		},
	}

	// Create test logger
	log := logger.NewLogger()

	// Create test token manager
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	if err != nil {
		t.Fatalf("Failed to create keyring manager: %v", err)
	}
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)

	// Create test templates
	templates := template.New("")

	// Create exports handler
	handler := NewExportsHandler(cfg, log, tokenManager, templates)

	// Test GET request
	req := httptest.NewRequest("GET", "/exports", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Check response (it will be an error due to missing template, but that's expected)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestAuthHandler(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Trakt: config.TraktConfig{
			ClientID:     "test_client_id",
			ClientSecret: "test_client_secret",
		},
		Auth: config.AuthConfig{
			RedirectURI: "http://localhost:8080/callback",
		},
	}

	// Create test logger
	log := logger.NewLogger()

	// Create test token manager
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	if err != nil {
		t.Fatalf("Failed to create keyring manager: %v", err)
	}
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)

	// Create test templates
	templates := template.New("")

	// Create auth handler
	handler := NewAuthHandler(cfg, log, tokenManager, templates)

	// Test auth-url endpoint
	req := httptest.NewRequest("GET", "/auth-url", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Check response (it will be an error due to missing template, but that's expected)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestDownloadHandler(t *testing.T) {
	// Create temporary directory for test
	tempDir := "./test_downloads"
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test file
	testFile := tempDir + "/test.csv"
	err = os.WriteFile(testFile, []byte("test,data\n1,2"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create test logger
	log := logger.NewLogger()

	// Create download handler
	handler := NewDownloadHandler(tempDir, log)

	// Test file download
	req := httptest.NewRequest("GET", "/download/test.csv", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Check response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Check content type
	expectedContentType := "text/csv"
	if rec.Header().Get("Content-Type") != expectedContentType {
		t.Errorf("Expected content type %s, got %s", expectedContentType, rec.Header().Get("Content-Type"))
	}
}

func TestExportItemParsing(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Letterboxd: config.LetterboxdConfig{
			ExportDir: "./test_exports",
		},
	}

	// Create test logger
	log := logger.NewLogger()

	// Create test token manager
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	if err != nil {
		t.Fatalf("Failed to create keyring manager: %v", err)
	}
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)

	// Create test templates
	templates := template.New("")

	// Create exports handler
	handler := NewExportsHandler(cfg, log, tokenManager, templates)

	// Test export type parsing
	testCases := []struct {
		filename string
		expected string
	}{
		{"watched_movies.csv", "watched"},
		{"collection_2025.csv", "collection"},
		{"tv_shows.csv", "shows"},
		{"ratings_export.csv", "ratings"},
		{"watchlist.csv", "watchlist"},
		{"unknown_file.csv", ""},
	}

	for _, tc := range testCases {
		result := handler.parseExportType(tc.filename)
		if result != tc.expected {
			t.Errorf("parseExportType(%s) = %s, expected %s", tc.filename, result, tc.expected)
		}
	}
}

func TestFormatFileSize(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Letterboxd: config.LetterboxdConfig{
			ExportDir: "./test_exports",
		},
	}

	// Create test logger
	log := logger.NewLogger()

	// Create test token manager
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	if err != nil {
		t.Fatalf("Failed to create keyring manager: %v", err)
	}
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)

	// Create test templates
	templates := template.New("")

	// Create exports handler
	handler := NewExportsHandler(cfg, log, tokenManager, templates)

	// Test file size formatting
	testCases := []struct {
		size     int64
		expected string
	}{
		{100, "100 B"},
		{1024, "1.0 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tc := range testCases {
		result := handler.formatFileSize(tc.size)
		if result != tc.expected {
			t.Errorf("formatFileSize(%d) = %s, expected %s", tc.size, result, tc.expected)
		}
	}
}

func TestStatusFormattingFunctions(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Auth: config.AuthConfig{
			CallbackPort: 8080,
		},
	}

	// Create test logger
	log := logger.NewLogger()

	// Create test token manager
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	if err != nil {
		t.Fatalf("Failed to create keyring manager: %v", err)
	}
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)

	// Create test templates
	templates := template.New("")

	// Create status handler
	handler := NewStatusHandler(cfg, log, tokenManager, templates)

	// Test uptime formatting
	uptime := handler.formatUptime()
	if len(uptime) == 0 {
		t.Error("formatUptime() returned empty string")
	}

	// Test time remaining formatting
	futureTime := time.Now().Add(2 * time.Hour)
	remaining := handler.formatTimeRemaining(futureTime)
	if remaining == "Unknown" || remaining == "Expired" {
		t.Errorf("formatTimeRemaining() returned unexpected value: %s", remaining)
	}

	// Test expired time
	pastTime := time.Now().Add(-1 * time.Hour)
	expired := handler.formatTimeRemaining(pastTime)
	if expired != "Expired" {
		t.Errorf("formatTimeRemaining() should return 'Expired' for past time, got: %s", expired)
	}

	// Test zero time
	zeroTime := time.Time{}
	unknown := handler.formatTimeRemaining(zeroTime)
	if unknown != "Unknown" {
		t.Errorf("formatTimeRemaining() should return 'Unknown' for zero time, got: %s", unknown)
	}
}

func TestExportEstimation(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Letterboxd: config.LetterboxdConfig{
			ExportDir: "./test_exports",
		},
	}

	// Create test logger
	log := logger.NewLogger()

	// Create test token manager
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	if err != nil {
		t.Fatalf("Failed to create keyring manager: %v", err)
	}
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)

	// Create test templates
	templates := template.New("")

	// Create exports handler
	handler := NewExportsHandler(cfg, log, tokenManager, templates)

	// Test duration estimation
	testCases := []struct {
		recordCount int
		expected    string
	}{
		{0, "< 1s"},
		{50, "< 1s"},
		{150, "1s"},
		{6000, "1m"},
		{12000, "2m"},
		{360000, "1h 0m"},
	}

	for _, tc := range testCases {
		result := handler.estimateExportDuration(tc.recordCount)
		if result != tc.expected {
			t.Errorf("estimateExportDuration(%d) = %s, expected %s", tc.recordCount, result, tc.expected)
		}
	}
}

func TestHandlerMethods(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Letterboxd: config.LetterboxdConfig{
			ExportDir: "./test_exports",
		},
	}

	// Create test logger
	log := logger.NewLogger()

	// Create test token manager
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	if err != nil {
		t.Fatalf("Failed to create keyring manager: %v", err)
	}
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)

	// Create test templates
	templates := template.New("")

	// Create exports handler
	handler := NewExportsHandler(cfg, log, tokenManager, templates)

	// Test unsupported method
	req := httptest.NewRequest("PUT", "/exports", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d for PUT method, got %d", http.StatusMethodNotAllowed, rec.Code)
	}

	// Test POST method (export start)
	req2 := httptest.NewRequest("POST", "/exports?type=watched", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	// Should return JSON response indicating authentication required
	if rec2.Code != http.StatusOK {
		t.Errorf("Expected status %d for POST method, got %d", http.StatusOK, rec2.Code)
	}

	// Test DELETE method
	req3 := httptest.NewRequest("DELETE", "/api/export/test123", nil)
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)

	// Should succeed (mock delete)
	if rec3.Code != http.StatusOK {
		t.Errorf("Expected status %d for DELETE, got %d", http.StatusOK, rec3.Code)
	}
}

func TestStatusCalculation(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Auth: config.AuthConfig{
			CallbackPort: 8080,
		},
	}

	// Create test logger
	log := logger.NewLogger()

	// Create test token manager
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	if err != nil {
		t.Fatalf("Failed to create keyring manager: %v", err)
	}
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)

	// Create test templates
	templates := template.New("")

	// Create status handler
	handler := NewStatusHandler(cfg, log, tokenManager, templates)

	// Test overall status calculation with mock data
	data := &StatusData{
		TokenStatus: &TokenStatusData{
			IsValid:  false,
			HasToken: false,
		},
		APIStatus:    "healthy",
		ServerStatus: "healthy",
	}

	status, message := handler.calculateOverallStatus(data)
	if status != "warning" {
		t.Errorf("Expected status 'warning' for invalid token, got '%s'", status)
	}
	if message == "" {
		t.Error("Expected non-empty message for warning status")
	}

	// Test with valid token but unhealthy API
	data.TokenStatus.IsValid = true
	data.APIStatus = "unhealthy"
	status, message = handler.calculateOverallStatus(data)
	if status != "error" {
		t.Errorf("Expected status 'error' for unhealthy API, got '%s'", status)
	}

	// Test with all healthy
	data.APIStatus = "healthy"
	status, message = handler.calculateOverallStatus(data)
	if status != "healthy" {
		t.Errorf("Expected status 'healthy' for all healthy, got '%s'", status)
	}
}

func TestResourcesData(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Auth: config.AuthConfig{
			CallbackPort: 8080,
		},
	}

	// Create test logger
	log := logger.NewLogger()

	// Create test token manager
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	if err != nil {
		t.Fatalf("Failed to create keyring manager: %v", err)
	}
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)

	// Create test templates
	templates := template.New("")

	// Create status handler
	handler := NewStatusHandler(cfg, log, tokenManager, templates)

	// Test system resources gathering
	resources := handler.getSystemResources()
	if resources == nil {
		t.Error("Resources should not be nil")
	}

	if resources.Goroutines <= 0 {
		t.Error("Goroutines count should be positive")
	}

	if resources.MemoryUsed == "" {
		t.Error("MemoryUsed should not be empty")
	}

	if resources.MemoryTotal == "" {
		t.Error("MemoryTotal should not be empty")
	}
}

func TestClientIDMasking(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Auth: config.AuthConfig{
			CallbackPort: 8080,
		},
	}

	// Create test logger
	log := logger.NewLogger()

	// Create test token manager
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	if err != nil {
		t.Fatalf("Failed to create keyring manager: %v", err)
	}
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)

	// Create test templates
	templates := template.New("")

	// Create status handler
	handler := NewStatusHandler(cfg, log, tokenManager, templates)

	// Test client ID masking
	testCases := []struct {
		input    string
		expected string
	}{
		{"12345678", "1234***5678"},
		{"abc", "abc"},
		{"", ""},
		{"abcdefghijklmnop", "abcd***mnop"},
	}

	for _, tc := range testCases {
		result := handler.maskClientID(tc.input)
		if result != tc.expected {
			t.Errorf("maskClientID(%s) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}

func TestCountCSVRecords(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Letterboxd: config.LetterboxdConfig{
			ExportDir: "./test_exports",
		},
	}

	// Create test logger
	log := logger.NewLogger()

	// Create test token manager
	keyringMgr, err := keyring.NewManager(keyring.MemoryBackend)
	if err != nil {
		t.Fatalf("Failed to create keyring manager: %v", err)
	}
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)

	// Create test templates
	templates := template.New("")

	// Create exports handler
	handler := NewExportsHandler(cfg, log, tokenManager, templates)

	// Create temporary CSV file for testing
	testDir := "./test_csv_dir"
	err = os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	testFile := testDir + "/test.csv"
	csvContent := "Title,Year,Rating\nMovie1,2023,8\nMovie2,2024,9\n\nMovie3,2025,7\n"
	err = os.WriteFile(testFile, []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test CSV file: %v", err)
	}

	// Test CSV record counting
	count := handler.countCSVRecords(testFile)
	expectedCount := 3 // Should count 3 non-empty data rows (excluding header and empty line)
	if count != expectedCount {
		t.Errorf("countCSVRecords() = %d, expected %d", count, expectedCount)
	}

	// Test with non-existent file
	nonExistentFile := testDir + "/nonexistent.csv"
	count = handler.countCSVRecords(nonExistentFile)
	if count != 0 {
		t.Errorf("countCSVRecords() for non-existent file = %d, expected 0", count)
	}

	// Test optimized CSV record counting
	countOpt := handler.countCSVRecordsOptimized(testFile)
	if countOpt != expectedCount {
		t.Errorf("countCSVRecordsOptimized() = %d, expected %d", countOpt, expectedCount)
	}

	// Test optimized counting with non-existent file
	countOptNonExistent := handler.countCSVRecordsOptimized(nonExistentFile)
	if countOptNonExistent != 0 {
		t.Errorf("countCSVRecordsOptimized() for non-existent file = %d, expected 0", countOptNonExistent)
	}
}

func TestCacheSystem(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{}
	log := logger.NewLogger()
	keyringMgr, _ := keyring.NewManager(keyring.MemoryBackend)
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)
	templates := template.New("")

	// Create exports handler
	handler := NewExportsHandler(cfg, log, tokenManager, templates)

	// Test cache initialization
	if handler.cache == nil {
		t.Error("Cache should be initialized")
	}

	if handler.cache.cacheTTL != 5*time.Minute {
		t.Errorf("Expected cache TTL of 5 minutes, got %v", handler.cache.cacheTTL)
	}

	// Test cache with empty exports dir
	exports := handler.getExportsWithCache(1, 10)
	if len(exports) != 0 {
		t.Errorf("Expected 0 exports for non-existent dir, got %d", len(exports))
	}
}

func TestLazyLoading(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{}
	log := logger.NewLogger()
	keyringMgr, _ := keyring.NewManager(keyring.MemoryBackend)
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)
	templates := template.New("")

	// Create temporary exports directory
	tempDir, err := ioutil.TempDir("", "test_exports_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set exports directory
	cfg.Letterboxd.ExportDir = tempDir
	handler := NewExportsHandler(cfg, log, tokenManager, templates)

	// Create test export directories with different dates
	recentDir := filepath.Join(tempDir, "export_2025-07-29_10-00")
	olderDir := filepath.Join(tempDir, "export_2024-01-01_10-00")

	if err := os.MkdirAll(recentDir, 0755); err != nil {
		t.Fatalf("Failed to create recent dir: %v", err)
	}
	if err := os.MkdirAll(olderDir, 0755); err != nil {
		t.Fatalf("Failed to create older dir: %v", err)
	}

	// Create test CSV files
	recentCSV := filepath.Join(recentDir, "watched.csv")

	if err := ioutil.WriteFile(recentCSV, []byte("Title,Year\nMovie1,2025\n"), 0644); err != nil {
		t.Fatalf("Failed to write recent CSV: %v", err)
	}

	// Test date parsing
	parsedTime := handler.parseDirTime("export_2025-07-29_10-00")
	if parsedTime.IsZero() {
		t.Error("Failed to parse directory time")
	}

	// Test optimized directory processing
	exportItem := handler.processExportDirectoryOptimized(recentDir, "export_2025-07-29_10-00")
	if exportItem == nil {
		t.Error("Expected export item from optimized processing")
	} else {
		if exportItem.Type != "watched" {
			t.Errorf("Expected type 'watched', got '%s'", exportItem.Type)
		}
		if exportItem.Status != "completed" {
			t.Errorf("Expected status 'completed', got '%s'", exportItem.Status)
		}
	}

	// Test optimized CSV file processing
	csvItem := handler.processCSVFileOptimized(recentCSV, "watched.csv")
	if csvItem == nil {
		t.Error("Expected CSV item from optimized processing")
	} else {
		if csvItem.Type != "watched" {
			t.Errorf("Expected type 'watched', got '%s'", csvItem.Type)
		}
	}
}

func TestExportUtilityFunctions(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{}
	log := logger.NewLogger()
	keyringMgr, _ := keyring.NewManager(keyring.MemoryBackend)
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)
	templates := template.New("")
	handler := NewExportsHandler(cfg, log, tokenManager, templates)

	// Test parse export type
	tests := []struct {
		filename string
		expected string
	}{
		{"watched.csv", "watched"},
		{"collection.csv", "collection"},
		{"shows.csv", "shows"},
		{"ratings.csv", "ratings"},
		{"watchlist.csv", "watchlist"},
		{"unknown.csv", ""},
	}

	for _, test := range tests {
		result := handler.parseExportType(test.filename)
		if result != test.expected {
			t.Errorf("parseExportType(%s): expected %s, got %s", test.filename, test.expected, result)
		}
	}

	// Test file size formatting
	testSizes := []struct {
		size     int64
		expected string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
	}

	for _, test := range testSizes {
		result := handler.formatFileSize(test.size)
		if result != test.expected {
			t.Errorf("formatFileSize(%d): expected %s, got %s", test.size, test.expected, result)
		}
	}

	// Test getIntParam
	req := httptest.NewRequest("GET", "/?page=5&limit=20&invalid=abc", nil)

	page := handler.getIntParam(req, "page", 1)
	if page != 5 {
		t.Errorf("getIntParam(page): expected 5, got %d", page)
	}

	limit := handler.getIntParam(req, "limit", 10)
	if limit != 20 {
		t.Errorf("getIntParam(limit): expected 20, got %d", limit)
	}

	defaultVal := handler.getIntParam(req, "missing", 42)
	if defaultVal != 42 {
		t.Errorf("getIntParam(missing): expected 42, got %d", defaultVal)
	}

	invalidVal := handler.getIntParam(req, "invalid", 99)
	if invalidVal != 99 {
		t.Errorf("getIntParam(invalid): expected 99, got %d", invalidVal)
	}

	// Test apply filters
	exports := []ExportItem{
		{Type: "watched", Status: "completed"},
		{Type: "ratings", Status: "completed"},
		{Type: "watched", Status: "failed"},
	}

	// Filter by type
	watchedOnly := handler.applyFilters(exports, "watched", "")
	if len(watchedOnly) != 2 {
		t.Errorf("applyFilters type watched: expected 2, got %d", len(watchedOnly))
	}

	// Filter by status
	completedOnly := handler.applyFilters(exports, "", "completed")
	if len(completedOnly) != 2 {
		t.Errorf("applyFilters status completed: expected 2, got %d", len(completedOnly))
	}

	// Filter by both
	watchedCompleted := handler.applyFilters(exports, "watched", "completed")
	if len(watchedCompleted) != 1 {
		t.Errorf("applyFilters watched+completed: expected 1, got %d", len(watchedCompleted))
	}

	// No filters
	allExports := handler.applyFilters(exports, "", "")
	if len(allExports) != 3 {
		t.Errorf("applyFilters no filter: expected 3, got %d", len(allExports))
	}
}