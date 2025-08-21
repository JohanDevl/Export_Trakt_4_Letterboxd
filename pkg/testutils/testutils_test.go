package testutils

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockLogger(t *testing.T) {
	logger := NewMockLogger()
	
	// Test basic logging
	logger.On("Info", "test.message", []map[string]interface{}{{"key": "value"}}).Return()
	logger.Info("test.message", map[string]interface{}{"key": "value"})
	
	assert.Equal(t, "test.message", logger.LastMessage)
	assert.Equal(t, "value", logger.LastData["key"])
	
	logger.AssertExpectations(t)
}

func TestNoOpLogger(t *testing.T) {
	logger := NewNoOpLogger()
	
	// Should not panic or error
	logger.Info("test")
	logger.Error("test")
	logger.Debug("test")
	logger.Warn("test")
	logger.SetLogLevel("debug")
	err := logger.SetLogFile("/tmp/test.log")
	assert.NoError(t, err)
}

func TestCapturingLogger(t *testing.T) {
	logger := NewCapturingLogger()
	
	// Log some messages
	logger.Info("info.message", map[string]interface{}{"level": "info"})
	logger.Error("error.message", map[string]interface{}{"level": "error"})
	logger.Debug("debug.message")
	
	// Check captured messages
	messages := logger.GetMessages()
	assert.Len(t, messages, 3)
	
	// Check by level
	infoMessages := logger.GetMessagesByLevel("info")
	assert.Len(t, infoMessages, 1)
	assert.Equal(t, "info.message", infoMessages[0].MessageID)
	
	errorMessages := logger.GetMessagesByLevel("error")
	assert.Len(t, errorMessages, 1)
	assert.Equal(t, "error.message", errorMessages[0].MessageID)
	
	// Test clear
	logger.Clear()
	assert.Len(t, logger.GetMessages(), 0)
}

func TestMockTokenManager(t *testing.T) {
	tokenMgr := NewMockTokenManager()
	
	// Test default token
	token, err := tokenMgr.GetValidAccessToken()
	assert.NoError(t, err)
	assert.Equal(t, "test_token", token)
	
	// Test custom token
	tokenMgr.SetToken("custom_token")
	token, err = tokenMgr.GetValidAccessToken()
	assert.NoError(t, err)
	assert.Equal(t, "custom_token", token)
	
	// Test error
	testErr := assert.AnError
	tokenMgr.SetError(testErr)
	_, err = tokenMgr.GetValidAccessToken()
	assert.Error(t, err)
	assert.Equal(t, testErr, err)
}

func TestMockTranslator(t *testing.T) {
	translator := NewMockTranslator()
	
	// Test default behavior (return messageID)
	result := translator.Translate("test.message", nil)
	assert.Equal(t, "test.message", result)
	
	// Test custom translation
	translator.SetTranslation("test.message", "translated message")
	result = translator.Translate("test.message", nil)
	assert.Equal(t, "translated message", result)
}

func TestMockMetricsRecorder(t *testing.T) {
	metrics := NewMockMetricsRecorder()
	
	// Test initial state
	assert.Equal(t, int64(0), metrics.GetJobsProcessed())
	assert.Equal(t, int64(0), metrics.GetJobsErrored())
	assert.Len(t, metrics.GetJobDurations(), 0)
	
	// Test recording
	metrics.On("IncrementJobsProcessed").Return()
	metrics.On("IncrementJobsErrored").Return()
	metrics.On("RecordJobDuration", 100*time.Millisecond).Return()
	
	metrics.IncrementJobsProcessed()
	metrics.IncrementJobsErrored()
	metrics.RecordJobDuration(100 * time.Millisecond)
	
	assert.Equal(t, int64(1), metrics.GetJobsProcessed())
	assert.Equal(t, int64(1), metrics.GetJobsErrored())
	assert.Len(t, metrics.GetJobDurations(), 1)
	assert.Equal(t, 100*time.Millisecond, metrics.GetJobDurations()[0])
	
	// Test reset
	metrics.Reset()
	assert.Equal(t, int64(0), metrics.GetJobsProcessed())
	assert.Equal(t, int64(0), metrics.GetJobsErrored())
	assert.Len(t, metrics.GetJobDurations(), 0)
	
	metrics.AssertExpectations(t)
}

func TestTempDir(t *testing.T) {
	dir, cleanup := TempDir(t)
	defer cleanup()
	
	// Verify directory exists
	stat, err := os.Stat(dir)
	assert.NoError(t, err)
	assert.True(t, stat.IsDir())
	
	// Cleanup should remove directory
	cleanup()
	_, err = os.Stat(dir)
	assert.True(t, os.IsNotExist(err))
}

func TestTempFile(t *testing.T) {
	content := "test content"
	path, cleanup := TempFile(t, content)
	
	// Verify file exists and has correct content
	data, err := os.ReadFile(path)
	assert.NoError(t, err)
	assert.Equal(t, content, string(data))
	
	// Test that file exists before cleanup
	assert.True(t, FileExists(path))
	
	// Cleanup should remove file
	cleanup()
	assert.False(t, FileExists(path))
}

func TestSetEnv(t *testing.T) {
	key := "TEST_ENV_VAR"
	value := "test_value"
	
	// Ensure env var is not set initially
	_, exists := os.LookupEnv(key)
	require.False(t, exists)
	
	// Set env var
	cleanup := SetEnv(t, key, value)
	defer cleanup()
	
	// Verify env var is set
	actual := os.Getenv(key)
	assert.Equal(t, value, actual)
	
	// Cleanup should unset env var
	cleanup()
	_, exists = os.LookupEnv(key)
	assert.False(t, exists)
}

func TestUnsetEnv(t *testing.T) {
	key := "TEST_UNSET_VAR"
	value := "original_value"
	
	// Set env var initially
	os.Setenv(key, value)
	defer os.Unsetenv(key)
	
	// Unset env var
	cleanup := UnsetEnv(t, key)
	
	// Verify env var is unset
	_, exists := os.LookupEnv(key)
	assert.False(t, exists)
	
	// Cleanup should restore env var
	cleanup()
	actual := os.Getenv(key)
	assert.Equal(t, value, actual)
}

func TestTestConfig(t *testing.T) {
	cfg := TestConfig()
	
	assert.NotNil(t, cfg)
	assert.Equal(t, "test_client_id", cfg.Trakt.ClientID)
	assert.Equal(t, "https://api.trakt.tv", cfg.Trakt.APIBaseURL)
	assert.Equal(t, "./test_exports", cfg.Letterboxd.ExportDir)
	assert.Equal(t, "aggregated", cfg.Export.HistoryMode)
}

func TestTestConfigWithExportDir(t *testing.T) {
	dir := "/custom/export/dir"
	cfg := TestConfigWithExportDir(dir)
	
	assert.Equal(t, dir, cfg.Letterboxd.ExportDir)
}

func TestWithTestTimeout(t *testing.T) {
	// Test successful completion
	WithTestTimeout(t, time.Second, func() {
		time.Sleep(10 * time.Millisecond)
	})
	
	// Test would timeout (but we can't test this without failing the test)
}

func TestAssertEventually(t *testing.T) {
	counter := 0
	
	AssertEventually(t, func() bool {
		counter++
		return counter >= 3
	}, time.Second, "counter should reach 3")
	
	assert.GreaterOrEqual(t, counter, 3)
}

func TestCreateTestCSV(t *testing.T) {
	dir, cleanup := TempDir(t)
	defer cleanup()
	
	headers := []string{"Title", "Year", "Rating"}
	rows := [][]string{
		{"Inception", "2010", "9"},
		{"The Dark Knight", "2008", "10"},
	}
	
	path := CreateTestCSV(t, dir, "test.csv", headers, rows)
	
	// Verify file exists
	assert.True(t, FileExists(path))
	
	// Read and verify content
	content := ReadFile(t, path)
	expected := "Title,Year,Rating\nInception,2010,9\nThe Dark Knight,2008,10\n"
	assert.Equal(t, expected, content)
}

func TestFileExists(t *testing.T) {
	// Test with existing file
	path, cleanup := TempFile(t, "test")
	defer cleanup()
	
	assert.True(t, FileExists(path))
	
	// Test with non-existing file
	assert.False(t, FileExists("/non/existing/file"))
}

func TestDirExists(t *testing.T) {
	// Test with existing directory
	dir, cleanup := TempDir(t)
	defer cleanup()
	
	assert.True(t, DirExists(dir))
	
	// Test with non-existing directory
	assert.False(t, DirExists("/non/existing/dir"))
}

func TestTestJob(t *testing.T) {
	// Test simple job
	job := NewSimpleTestJob("test_job", 10*time.Millisecond)
	assert.Equal(t, "test_job", job.ID())
	
	start := time.Now()
	err := job.Execute(context.Background())
	duration := time.Since(start)
	
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, duration, 10*time.Millisecond)
	assert.Equal(t, int64(1), job.ExecutionCount())
}

func TestErrorTestJob(t *testing.T) {
	job := NewErrorTestJob("error_job", "test error")
	
	err := job.Execute(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test error")
}

func TestPanicTestJob(t *testing.T) {
	job := NewPanicTestJob("panic_job", "test panic")
	
	assert.Panics(t, func() {
		job.Execute(context.Background())
	})
}

func TestWaitJob(t *testing.T) {
	job := NewWaitJob("wait_job")
	
	// Test that job waits
	done := make(chan bool)
	go func() {
		job.Execute(context.Background())
		done <- true
	}()
	
	// Should not complete immediately
	select {
	case <-done:
		t.Fatal("Job completed too early")
	case <-time.After(10 * time.Millisecond):
		// Expected
	}
	
	// Signal completion
	job.Signal()
	
	// Should complete now
	select {
	case <-done:
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Job did not complete after signal")
	}
}

func TestCounterJob(t *testing.T) {
	counter := int64(0)
	job := NewCounterJob("counter_job", &counter)
	
	assert.Equal(t, int64(0), counter)
	
	err := job.Execute(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(1), counter)
}

func TestBatchJob(t *testing.T) {
	items := []string{"item1", "item2", "item3"}
	processed := make([]string, 0)
	
	processor := func(item string) error {
		processed = append(processed, item)
		return nil
	}
	
	job := NewBatchJob("batch_job", items, processor)
	
	err := job.Execute(context.Background())
	assert.NoError(t, err)
	
	processedItems := job.GetProcessedItems()
	assert.Equal(t, items, processedItems)
}

func TestJobGroup(t *testing.T) {
	jobs := CreateTestJobs(3, time.Millisecond)
	group := NewJobGroup(jobs...)
	
	err := group.ExecuteAll(context.Background())
	assert.NoError(t, err)
	assert.Empty(t, group.GetErrors())
}

func TestCreateMixedTestJobs(t *testing.T) {
	jobs := CreateMixedTestJobs(2, 1)
	assert.Len(t, jobs, 3)
	
	group := NewJobGroup(jobs...)
	err := group.ExecuteAll(context.Background())
	
	// Should have 1 error
	assert.Error(t, err)
	errors := group.GetErrors()
	assert.Len(t, errors, 1)
}

func TestSampleData(t *testing.T) {
	// Test fixtures
	movie := SampleTraktMovie()
	assert.Equal(t, "Inception", movie["title"])
	assert.Equal(t, 2010, movie["year"])
	
	watched := SampleTraktWatchedMovie()
	assert.Contains(t, watched, "movie")
	assert.Equal(t, 1, watched["plays"])
	
	movies := SampleTraktMovies()
	assert.Len(t, movies, 3)
	
	headers := SampleCSVHeaders()
	assert.Contains(t, headers, "Title")
	assert.Contains(t, headers, "Year")
	
	rows := SampleCSVRows()
	assert.Len(t, rows, 3)
	assert.Equal(t, "Inception", rows[0][0])
}

func TestSampleErrors(t *testing.T) {
	errors := SampleErrors()
	assert.NotEmpty(t, errors)
	
	testErr := errors[0].(*TestError)
	assert.Equal(t, "API_ERROR", testErr.GetCode())
	assert.Contains(t, testErr.Error(), "API request failed")
}