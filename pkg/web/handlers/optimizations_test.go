package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/auth"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/security/keyring"
)

func TestImprovedCSVRecordCounting(t *testing.T) {
	// Create test handler
	cfg := &config.Config{}
	log := logger.NewLogger()
	keyringMgr, _ := keyring.NewManager(keyring.MemoryBackend)
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)
	handler := NewExportsHandler(cfg, log, tokenManager, nil, nil)
	
	// Create temp directory and CSV file
	tmpDir, err := os.MkdirTemp("", "csv_counting_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	testFile := filepath.Join(tmpDir, "test.csv")
	
	// Create a CSV with known record count (1000 records)
	var csvContent strings.Builder
	csvContent.WriteString("Title,Year,WatchedDate,Rating10,imdbID,tmdbID,Rewatch\n")
	for i := 0; i < 1000; i++ {
		csvContent.WriteString("Movie Title,2020,2023-01-01,8,tt1234567,12345,false\n")
	}
	
	err = os.WriteFile(testFile, []byte(csvContent.String()), 0644)
	if err != nil {
		t.Fatalf("Failed to write test CSV: %v", err)
	}
	
	// Test improved counting
	count := handler.countCSVRecordsOptimized(testFile)
	
	// Should be close to 1000 (allowing for small estimation error)
	if count < 950 || count > 1050 {
		t.Errorf("Expected count around 1000, got %d", count)
	}
	
	// Test with small file (should count exactly)
	smallFile := filepath.Join(tmpDir, "small.csv")
	smallContent := "Title,Year,WatchedDate,Rating10,imdbID,tmdbID,Rewatch\n"
	for i := 0; i < 10; i++ {
		smallContent += "Movie Title,2020,2023-01-01,8,tt1234567,12345,false\n"
	}
	
	err = os.WriteFile(smallFile, []byte(smallContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write small CSV: %v", err)
	}
	
	smallCount := handler.countCSVRecordsOptimized(smallFile)
	if smallCount != 10 {
		t.Errorf("Expected exact count of 10 for small file, got %d", smallCount)
	}
}

func TestTimezoneConversion(t *testing.T) {
	// Test with Europe/Paris timezone
	cfg := &config.Config{
		Export: config.ExportConfig{
			Timezone: "Europe/Paris",
		},
	}
	log := logger.NewLogger()
	keyringMgr, _ := keyring.NewManager(keyring.MemoryBackend)
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)
	handler := NewExportsHandler(cfg, log, tokenManager, nil, nil)
	
	// Test time conversion
	utcTime := time.Date(2023, 7, 15, 12, 0, 0, 0, time.UTC)
	convertedTime := handler.convertToConfigTimezone(utcTime)
	
	// In July, Paris is UTC+2
	expectedHour := 14
	if convertedTime.Hour() != expectedHour {
		t.Errorf("Expected hour %d for Paris timezone, got %d", expectedHour, convertedTime.Hour())
	}
	
	// Test with UTC timezone
	cfg.Export.Timezone = "UTC"
	handler.config = cfg
	
	utcConverted := handler.convertToConfigTimezone(utcTime)
	if !utcConverted.Equal(utcTime) {
		t.Errorf("UTC conversion should return same time")
	}
	
	// Test with invalid timezone (should fallback to UTC)
	cfg.Export.Timezone = "Invalid/Timezone"
	handler.config = cfg
	
	invalidConverted := handler.convertToConfigTimezone(utcTime)
	if !invalidConverted.Equal(utcTime.UTC()) {
		t.Errorf("Invalid timezone should fallback to UTC")
	}
}

func TestFormatTimeInConfigTimezone(t *testing.T) {
	cfg := &config.Config{
		Export: config.ExportConfig{
			Timezone: "Europe/Paris",
		},
	}
	log := logger.NewLogger()
	keyringMgr, _ := keyring.NewManager(keyring.MemoryBackend)
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)
	handler := NewExportsHandler(cfg, log, tokenManager, nil, nil)
	
	// Test formatting
	utcTime := time.Date(2023, 7, 15, 12, 0, 0, 0, time.UTC)
	formatted := handler.formatTimeInConfigTimezone(utcTime, "2006-01-02 15:04")
	
	expected := "2023-07-15 14:00" // UTC+2 for Paris in July
	if formatted != expected {
		t.Errorf("Expected formatted time %s, got %s", expected, formatted)
	}
}

func TestImprovedCachePerformance(t *testing.T) {
	cfg := &config.Config{}
	log := logger.NewLogger()
	keyringMgr, _ := keyring.NewManager(keyring.MemoryBackend)
	tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)
	handler := NewExportsHandler(cfg, log, tokenManager, nil, nil)
	
	// Verify cache TTL is increased for better performance
	if handler.cache.cacheTTL != 30*time.Minute {
		t.Errorf("Expected cache TTL of 30 minutes for better performance, got %v", handler.cache.cacheTTL)
	}
	
	// Verify recent cache TTL is shorter for fresh data
	if handler.cache.recentCacheTTL != 1*time.Minute {
		t.Errorf("Expected recent cache TTL of 1 minute, got %v", handler.cache.recentCacheTTL)
	}
}