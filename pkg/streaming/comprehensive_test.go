package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

// Mock logger for testing
type mockStreamingLogger struct {
	logs []map[string]interface{}
}

func (m *mockStreamingLogger) Debug(msg string, fields ...map[string]interface{}) {
	var fieldsMap map[string]interface{}
	if len(fields) > 0 {
		fieldsMap = fields[0]
	}
	m.logs = append(m.logs, map[string]interface{}{"level": "debug", "msg": msg, "fields": fieldsMap})
}

func (m *mockStreamingLogger) Info(msg string, fields ...map[string]interface{}) {
	var fieldsMap map[string]interface{}
	if len(fields) > 0 {
		fieldsMap = fields[0]
	}
	m.logs = append(m.logs, map[string]interface{}{"level": "info", "msg": msg, "fields": fieldsMap})
}

func (m *mockStreamingLogger) Warn(msg string, fields ...map[string]interface{}) {
	var fieldsMap map[string]interface{}
	if len(fields) > 0 {
		fieldsMap = fields[0]
	}
	m.logs = append(m.logs, map[string]interface{}{"level": "warn", "msg": msg, "fields": fieldsMap})
}

func (m *mockStreamingLogger) Error(msg string, fields ...map[string]interface{}) {
	var fieldsMap map[string]interface{}
	if len(fields) > 0 {
		fieldsMap = fields[0]
	}
	m.logs = append(m.logs, map[string]interface{}{"level": "error", "msg": msg, "fields": fieldsMap})
}

func (m *mockStreamingLogger) Fatal(msg string, fields ...map[string]interface{}) {
	var fieldsMap map[string]interface{}
	if len(fields) > 0 {
		fieldsMap = fields[0]
	}
	m.logs = append(m.logs, map[string]interface{}{"level": "fatal", "msg": msg, "fields": fieldsMap})
}

func (m *mockStreamingLogger) Debugf(msg string, data map[string]interface{}) {
	m.logs = append(m.logs, map[string]interface{}{"level": "debug", "msg": msg, "fields": data})
}

func (m *mockStreamingLogger) Infof(msg string, data map[string]interface{}) {
	m.logs = append(m.logs, map[string]interface{}{"level": "info", "msg": msg, "fields": data})
}

func (m *mockStreamingLogger) Warnf(msg string, data map[string]interface{}) {
	m.logs = append(m.logs, map[string]interface{}{"level": "warn", "msg": msg, "fields": data})
}

func (m *mockStreamingLogger) Errorf(msg string, data map[string]interface{}) {
	m.logs = append(m.logs, map[string]interface{}{"level": "error", "msg": msg, "fields": data})
}

func (m *mockStreamingLogger) SetLogLevel(level string) {}

func (m *mockStreamingLogger) SetLogFile(path string) error {
	return nil
}

func (m *mockStreamingLogger) SetTranslator(t logger.Translator) {}

// Mock batch handler for testing
type mockBatchHandler struct {
	name           string
	processBatches [][]interface{}
	shouldFail     bool
	processDelay   time.Duration
}

func (m *mockBatchHandler) ProcessBatch(ctx context.Context, batch []interface{}) error {
	if m.shouldFail {
		return fmt.Errorf("mock batch handler error")
	}
	
	if m.processDelay > 0 {
		time.Sleep(m.processDelay)
	}
	
	m.processBatches = append(m.processBatches, batch)
	return nil
}

func (m *mockBatchHandler) Name() string {
	return m.name
}

// Mock concurrent handler for testing
type mockConcurrentHandler struct {
	name         string
	processedItems []interface{}
	shouldFail   bool
	processDelay time.Duration
}

func (m *mockConcurrentHandler) ProcessItem(ctx context.Context, item interface{}) (interface{}, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("mock concurrent handler error")
	}
	
	if m.processDelay > 0 {
		time.Sleep(m.processDelay)
	}
	
	m.processedItems = append(m.processedItems, item)
	
	// Transform item (simple string transformation)
	if str, ok := item.(string); ok {
		return strings.ToUpper(str), nil
	}
	
	return item, nil
}

func (m *mockConcurrentHandler) Name() string {
	return m.name
}

func TestNewBatchProcessor(t *testing.T) {
	logger := &mockStreamingLogger{}
	handler := &mockBatchHandler{name: "test-batch-handler"}
	
	// Test with default config
	config := StreamConfig{
		Logger: logger,
	}
	
	processor := NewBatchProcessor(config, handler)
	
	if processor == nil {
		t.Fatal("Expected processor to be created")
	}
	
	if processor.logger != logger {
		t.Error("Expected logger to be set")
	}
	
	if processor.batchSize != 100 {
		t.Errorf("Expected default batch size 100, got %d", processor.batchSize)
	}
	
	if processor.bufferSize != 8192 {
		t.Errorf("Expected default buffer size 8192, got %d", processor.bufferSize)
	}
	
	if processor.processor != handler {
		t.Error("Expected handler to be set")
	}
}

func TestNewBatchProcessorWithCustomConfig(t *testing.T) {
	logger := &mockStreamingLogger{}
	handler := &mockBatchHandler{name: "test-handler"}
	
	config := StreamConfig{
		BatchSize:  50,
		BufferSize: 4096,
		Logger:     logger,
	}
	
	processor := NewBatchProcessor(config, handler)
	
	if processor.batchSize != 50 {
		t.Errorf("Expected batch size 50, got %d", processor.batchSize)
	}
	
	if processor.bufferSize != 4096 {
		t.Errorf("Expected buffer size 4096, got %d", processor.bufferSize)
	}
}

func TestBatchProcessorSetters(t *testing.T) {
	logger := &mockStreamingLogger{}
	handler := &mockBatchHandler{name: "test-handler"}
	config := StreamConfig{Logger: logger}
	
	processor := NewBatchProcessor(config, handler)
	
	// Test SetBatchSize
	processor.SetBatchSize(200)
	if processor.batchSize != 200 {
		t.Errorf("Expected batch size 200, got %d", processor.batchSize)
	}
	
	// Test invalid batch size (should be ignored)
	processor.SetBatchSize(-10)
	if processor.batchSize != 200 {
		t.Errorf("Expected batch size to remain 200, got %d", processor.batchSize)
	}
	
	// Test SetBufferSize
	processor.SetBufferSize(16384)
	if processor.bufferSize != 16384 {
		t.Errorf("Expected buffer size 16384, got %d", processor.bufferSize)
	}
	
	// Test invalid buffer size (should be ignored)
	processor.SetBufferSize(0)
	if processor.bufferSize != 16384 {
		t.Errorf("Expected buffer size to remain 16384, got %d", processor.bufferSize)
	}
}

func TestBatchProcessorProcess(t *testing.T) {
	logger := &mockStreamingLogger{}
	handler := &mockBatchHandler{name: "test-handler"}
	config := StreamConfig{
		BatchSize:  2, // Small batch size for testing
		BufferSize: 1024,
		Logger:     logger,
	}
	
	processor := NewBatchProcessor(config, handler)
	
	// Create test data
	testData := []map[string]string{
		{"id": "1", "name": "item1"},
		{"id": "2", "name": "item2"},
		{"id": "3", "name": "item3"},
	}
	
	// Create JSON input
	var jsonLines []string
	for _, item := range testData {
		jsonData, _ := json.Marshal(item)
		jsonLines = append(jsonLines, string(jsonData))
	}
	input := strings.NewReader(strings.Join(jsonLines, "\n"))
	
	var output strings.Builder
	ctx := context.Background()
	
	err := processor.Process(ctx, input, &output)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	// Check that batches were processed
	if len(handler.processBatches) != 2 {
		t.Errorf("Expected 2 batches to be processed, got %d", len(handler.processBatches))
	}
	
	// First batch should have 2 items
	if len(handler.processBatches[0]) != 2 {
		t.Errorf("Expected first batch to have 2 items, got %d", len(handler.processBatches[0]))
	}
	
	// Second batch should have 1 item (remainder)
	if len(handler.processBatches[1]) != 1 {
		t.Errorf("Expected second batch to have 1 item, got %d", len(handler.processBatches[1]))
	}
	
	// Check that logging occurred
	found := false
	for _, logEntry := range logger.logs {
		if logEntry["level"] == "info" && logEntry["msg"] == "stream.processing_started" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected processing start to be logged")
	}
}

func TestBatchProcessorProcessWithError(t *testing.T) {
	logger := &mockStreamingLogger{}
	handler := &mockBatchHandler{name: "failing-handler", shouldFail: true}
	config := StreamConfig{
		BatchSize:  2,
		Logger:     logger,
	}
	
	processor := NewBatchProcessor(config, handler)
	
	testData := `{"id": "1"}\n{"id": "2"}`
	input := strings.NewReader(testData)
	var output strings.Builder
	ctx := context.Background()
	
	err := processor.Process(ctx, input, &output)
	if err == nil {
		t.Error("Expected error when batch handler fails")
	}
	
	if !strings.Contains(err.Error(), "failed to process batch") {
		t.Errorf("Expected batch processing error, got: %v", err)
	}
}

func TestBatchProcessorProcessWithContextCancellation(t *testing.T) {
	logger := &mockStreamingLogger{}
	handler := &mockBatchHandler{name: "slow-handler", processDelay: 100 * time.Millisecond}
	config := StreamConfig{
		BatchSize:  1,
		Logger:     logger,
	}
	
	processor := NewBatchProcessor(config, handler)
	
	testData := `{"id": "1"}\n{"id": "2"}\n{"id": "3"}`
	input := strings.NewReader(testData)
	var output strings.Builder
	
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	
	err := processor.Process(ctx, input, &output)
	if err != context.DeadlineExceeded {
		t.Errorf("Expected context deadline exceeded, got: %v", err)
	}
}

func TestBatchProcessorProcessInvalidJSON(t *testing.T) {
	logger := &mockStreamingLogger{}
	handler := &mockBatchHandler{name: "test-handler"}
	config := StreamConfig{Logger: logger}
	
	processor := NewBatchProcessor(config, handler)
	
	// Invalid JSON
	input := strings.NewReader(`{"invalid": json}`)
	var output strings.Builder
	ctx := context.Background()
	
	err := processor.Process(ctx, input, &output)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
	
	if !strings.Contains(err.Error(), "failed to decode JSON") {
		t.Errorf("Expected JSON decode error, got: %v", err)
	}
}

func TestNewConcurrentStreamProcessor(t *testing.T) {
	logger := &mockStreamingLogger{}
	handler := &mockConcurrentHandler{name: "test-concurrent-handler"}
	
	config := StreamConfig{
		Logger: logger,
	}
	
	processor := NewConcurrentStreamProcessor(config, handler, 2)
	
	if processor == nil {
		t.Fatal("Expected processor to be created")
	}
	
	if processor.logger != logger {
		t.Error("Expected logger to be set")
	}
	
	if processor.workers != 2 {
		t.Errorf("Expected 2 workers, got %d", processor.workers)
	}
	
	if processor.processor != handler {
		t.Error("Expected handler to be set")
	}
}

func TestNewConcurrentStreamProcessorDefaults(t *testing.T) {
	logger := &mockStreamingLogger{}
	handler := &mockConcurrentHandler{name: "test-handler"}
	
	config := StreamConfig{Logger: logger}
	
	// Test with invalid workers count
	processor := NewConcurrentStreamProcessor(config, handler, -1)
	
	if processor.workers != 4 {
		t.Errorf("Expected default 4 workers, got %d", processor.workers)
	}
}

func TestConcurrentStreamProcessorProcess(t *testing.T) {
	logger := &mockStreamingLogger{}
	handler := &mockConcurrentHandler{name: "uppercase-handler"}
	config := StreamConfig{Logger: logger}
	
	processor := NewConcurrentStreamProcessor(config, handler, 2)
	
	// Create test data
	testItems := []string{"hello", "world", "test"}
	var jsonLines []string
	for _, item := range testItems {
		jsonData, _ := json.Marshal(item)
		jsonLines = append(jsonLines, string(jsonData))
	}
	input := strings.NewReader(strings.Join(jsonLines, "\n"))
	
	var output strings.Builder
	ctx := context.Background()
	
	err := processor.Process(ctx, input, &output)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	// Parse output to verify results
	decoder := json.NewDecoder(strings.NewReader(output.String()))
	var results []string
	
	for decoder.More() {
		var result string
		if err := decoder.Decode(&result); err != nil {
			break
		}
		results = append(results, result)
	}
	
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
	
	// Results should be uppercase (handler transforms strings to uppercase)
	for _, result := range results {
		if result != strings.ToUpper(result) {
			t.Errorf("Expected uppercase result, got: %s", result)
		}
	}
}

func TestConcurrentStreamProcessorWithErrors(t *testing.T) {
	logger := &mockStreamingLogger{}
	handler := &mockConcurrentHandler{name: "failing-handler", shouldFail: true}
	config := StreamConfig{Logger: logger}
	
	processor := NewConcurrentStreamProcessor(config, handler, 1)
	
	testData := `"test1"\n"test2"`
	input := strings.NewReader(testData)
	var output strings.Builder
	ctx := context.Background()
	
	err := processor.Process(ctx, input, &output)
	if err != nil {
		t.Fatalf("Process should not return error even with handler failures, got: %v", err)
	}
	
	// Check that errors were logged
	errorLogged := false
	for _, logEntry := range logger.logs {
		if logEntry["level"] == "error" {
			errorLogged = true
			break
		}
	}
	
	if !errorLogged {
		t.Error("Expected processing errors to be logged")
	}
}

func TestNewMemoryEfficientProcessor(t *testing.T) {
	logger := &mockStreamingLogger{}
	processorFunc := func(item interface{}) (interface{}, error) {
		return item, nil
	}
	
	processor := NewMemoryEfficientProcessor(logger, 128, processorFunc)
	
	if processor == nil {
		t.Fatal("Expected processor to be created")
	}
	
	if processor.logger != logger {
		t.Error("Expected logger to be set")
	}
	
	if processor.maxMemoryMB != 128 {
		t.Errorf("Expected max memory 128MB, got %d", processor.maxMemoryMB)
	}
	
	if processor.processor == nil {
		t.Error("Expected processor function to be set")
	}
}

func TestNewMemoryEfficientProcessorDefaults(t *testing.T) {
	logger := &mockStreamingLogger{}
	processorFunc := func(item interface{}) (interface{}, error) {
		return item, nil
	}
	
	// Test with invalid memory limit
	processor := NewMemoryEfficientProcessor(logger, -10, processorFunc)
	
	if processor.maxMemoryMB != 256 {
		t.Errorf("Expected default max memory 256MB, got %d", processor.maxMemoryMB)
	}
}

func TestMemoryEfficientProcessorProcess(t *testing.T) {
	logger := &mockStreamingLogger{}
	processorFunc := func(item interface{}) (interface{}, error) {
		// Simple transformation: add "processed_" prefix to strings
		if str, ok := item.(string); ok {
			return "processed_" + str, nil
		}
		return item, nil
	}
	
	processor := NewMemoryEfficientProcessor(logger, 128, processorFunc)
	
	// Create test data with JSON objects on separate lines
	testData := `{"name": "item1"}
{"name": "item2"}
{"name": "item3"}`
	
	input := strings.NewReader(testData)
	var output strings.Builder
	ctx := context.Background()
	
	err := processor.Process(ctx, input, &output)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	// Verify output contains processed items
	outputStr := output.String()
	if outputStr == "" {
		t.Error("Expected output to contain processed items")
	}
	
	// Check that processing completion was logged
	found := false
	for _, logEntry := range logger.logs {
		if logEntry["level"] == "info" && logEntry["msg"] == "memory_efficient.processing_completed" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected processing completion to be logged")
	}
}

func TestMemoryEfficientProcessorProcessWithErrors(t *testing.T) {
	logger := &mockStreamingLogger{}
	processorFunc := func(item interface{}) (interface{}, error) {
		return nil, fmt.Errorf("processor error")
	}
	
	processor := NewMemoryEfficientProcessor(logger, 128, processorFunc)
	
	testData := `{"name": "item1"}
{"name": "item2"}`
	
	input := strings.NewReader(testData)
	var output strings.Builder
	ctx := context.Background()
	
	err := processor.Process(ctx, input, &output)
	if err != nil {
		t.Fatalf("Process should not return error for processor failures, got: %v", err)
	}
	
	// Check that processing errors were logged
	errorLogged := false
	for _, logEntry := range logger.logs {
		if logEntry["level"] == "error" && logEntry["msg"] == "memory_efficient.processing_error" {
			errorLogged = true
			break
		}
	}
	
	if !errorLogged {
		t.Error("Expected processing errors to be logged")
	}
}

func TestMemoryEfficientProcessorProcessInvalidJSON(t *testing.T) {
	logger := &mockStreamingLogger{}
	processorFunc := func(item interface{}) (interface{}, error) {
		return item, nil
	}
	
	processor := NewMemoryEfficientProcessor(logger, 128, processorFunc)
	
	// Invalid JSON lines
	testData := `{"valid": "json"}
invalid json line
{"another": "valid"}`
	
	input := strings.NewReader(testData)
	var output strings.Builder
	ctx := context.Background()
	
	err := processor.Process(ctx, input, &output)
	if err != nil {
		t.Fatalf("Process should not return error for decode failures, got: %v", err)
	}
	
	// Check that decode errors were logged
	warningLogged := false
	for _, logEntry := range logger.logs {
		if logEntry["level"] == "warn" && logEntry["msg"] == "memory_efficient.decode_error" {
			warningLogged = true
			break
		}
	}
	
	if !warningLogged {
		t.Error("Expected decode errors to be logged as warnings")
	}
}

func TestMemoryEfficientProcessorContextCancellation(t *testing.T) {
	logger := &mockStreamingLogger{}
	processorFunc := func(item interface{}) (interface{}, error) {
		time.Sleep(100 * time.Millisecond) // Simulate slow processing
		return item, nil
	}
	
	processor := NewMemoryEfficientProcessor(logger, 128, processorFunc)
	
	testData := `{"item": "1"}
{"item": "2"}`
	
	input := strings.NewReader(testData)
	var output strings.Builder
	
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	
	err := processor.Process(ctx, input, &output)
	if err != context.DeadlineExceeded {
		t.Errorf("Expected context deadline exceeded, got: %v", err)
	}
}

func TestMemoryEfficientProcessorShouldGC(t *testing.T) {
	logger := &mockStreamingLogger{}
	processorFunc := func(item interface{}) (interface{}, error) {
		return item, nil
	}
	
	processor := NewMemoryEfficientProcessor(logger, 128, processorFunc)
	
	// Test shouldGC (currently always returns false)
	result := processor.shouldGC()
	if result {
		t.Error("Expected shouldGC to return false in current implementation")
	}
}

func TestProcessResultStruct(t *testing.T) {
	// Test ProcessResult struct
	result := ProcessResult{
		WorkerID: 1,
		Result:   "test result",
		Error:    fmt.Errorf("test error"),
		Duration: time.Second,
	}
	
	if result.WorkerID != 1 {
		t.Errorf("Expected WorkerID 1, got %d", result.WorkerID)
	}
	
	if result.Result != "test result" {
		t.Errorf("Expected result 'test result', got %v", result.Result)
	}
	
	if result.Error.Error() != "test error" {
		t.Errorf("Expected error 'test error', got %v", result.Error)
	}
	
	if result.Duration != time.Second {
		t.Errorf("Expected duration 1s, got %v", result.Duration)
	}
}

// Integration test for end-to-end processing
func TestStreamingIntegration(t *testing.T) {
	logger := &mockStreamingLogger{}
	
	// Test batch processor
	batchHandler := &mockBatchHandler{name: "integration-batch"}
	batchConfig := StreamConfig{
		BatchSize:  3,
		BufferSize: 1024,
		Logger:     logger,
	}
	batchProcessor := NewBatchProcessor(batchConfig, batchHandler)
	
	// Create test data
	testItems := []string{"item1", "item2", "item3", "item4", "item5"}
	var jsonLines []string
	for _, item := range testItems {
		jsonData, _ := json.Marshal(item)
		jsonLines = append(jsonLines, string(jsonData))
	}
	
	input := strings.NewReader(strings.Join(jsonLines, "\n"))
	var output strings.Builder
	ctx := context.Background()
	
	err := batchProcessor.Process(ctx, input, &output)
	if err != nil {
		t.Fatalf("Batch processing failed: %v", err)
	}
	
	// Verify batches
	if len(batchHandler.processBatches) != 2 {
		t.Errorf("Expected 2 batches, got %d", len(batchHandler.processBatches))
	}
	
	// Test concurrent processor
	concurrentHandler := &mockConcurrentHandler{name: "integration-concurrent"}
	concurrentProcessor := NewConcurrentStreamProcessor(batchConfig, concurrentHandler, 2)
	
	input2 := strings.NewReader(strings.Join(jsonLines, "\n"))
	var output2 strings.Builder
	
	err = concurrentProcessor.Process(ctx, input2, &output2)
	if err != nil {
		t.Fatalf("Concurrent processing failed: %v", err)
	}
	
	// Verify all items were processed concurrently
	if len(concurrentHandler.processedItems) != 5 {
		t.Errorf("Expected 5 processed items, got %d", len(concurrentHandler.processedItems))
	}
}