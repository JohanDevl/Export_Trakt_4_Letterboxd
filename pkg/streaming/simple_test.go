package streaming

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/stretchr/testify/assert"
)

// Simple test handler for testing
type TestBatchHandler struct {
	name string
}

func (h *TestBatchHandler) ProcessBatch(ctx context.Context, batch []interface{}) error {
	return nil
}

func (h *TestBatchHandler) Name() string {
	return h.name
}

func TestNewBatchProcessor_Basic(t *testing.T) {
	logger := logger.NewLogger()
	handler := &TestBatchHandler{name: "test-handler"}

	config := StreamConfig{
		Logger: logger,
	}

	processor := NewBatchProcessor(config, handler)

	assert.NotNil(t, processor)
	assert.Equal(t, 100, processor.batchSize)   // Default batch size
	assert.Equal(t, 8192, processor.bufferSize) // Default buffer size
	assert.Equal(t, logger, processor.logger)
	assert.Equal(t, handler, processor.processor)
}

func TestBatchProcessor_SetBatchSize(t *testing.T) {
	logger := logger.NewLogger()
	handler := &TestBatchHandler{name: "test-handler"}

	config := StreamConfig{
		Logger: logger,
	}

	processor := NewBatchProcessor(config, handler)

	// Test setting valid batch size
	processor.SetBatchSize(50)
	assert.Equal(t, 50, processor.batchSize)

	// Test setting invalid batch size (should not change)
	processor.SetBatchSize(0)
	assert.Equal(t, 50, processor.batchSize)

	processor.SetBatchSize(-1)
	assert.Equal(t, 50, processor.batchSize)
}

func TestBatchProcessor_SetBufferSize(t *testing.T) {
	logger := logger.NewLogger()
	handler := &TestBatchHandler{name: "test-handler"}

	config := StreamConfig{
		Logger: logger,
	}

	processor := NewBatchProcessor(config, handler)

	// Test setting valid buffer size
	processor.SetBufferSize(4096)
	assert.Equal(t, 4096, processor.bufferSize)

	// Test setting invalid buffer size (should not change)
	processor.SetBufferSize(0)
	assert.Equal(t, 4096, processor.bufferSize)

	processor.SetBufferSize(-1)
	assert.Equal(t, 4096, processor.bufferSize)
}

func TestBatchProcessor_Process_EmptyInput(t *testing.T) {
	logger := logger.NewLogger()
	handler := &TestBatchHandler{name: "test-handler"}

	config := StreamConfig{
		Logger: logger,
	}

	processor := NewBatchProcessor(config, handler)

	// Test with empty input
	reader := strings.NewReader("")
	var output strings.Builder

	ctx := context.Background()
	err := processor.Process(ctx, reader, &output)

	assert.NoError(t, err)
}

func TestBatchProcessor_Process_InvalidJSON(t *testing.T) {
	logger := logger.NewLogger()
	handler := &TestBatchHandler{name: "test-handler"}

	config := StreamConfig{
		Logger: logger,
	}

	processor := NewBatchProcessor(config, handler)

	// Test with invalid JSON
	reader := strings.NewReader("invalid json")
	var output strings.Builder

	ctx := context.Background()
	err := processor.Process(ctx, reader, &output)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode JSON")
}

func TestBatchProcessor_Process_CancelledContext(t *testing.T) {
	logger := logger.NewLogger()
	handler := &TestBatchHandler{name: "test-handler"}

	config := StreamConfig{
		Logger: logger,
	}

	processor := NewBatchProcessor(config, handler)

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	reader := strings.NewReader(`{"test": "data"}`)
	var output strings.Builder

	err := processor.Process(ctx, reader, &output)

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

// Simple test handler for concurrent processing
type TestConcurrentHandler struct {
	name string
}

func (h *TestConcurrentHandler) ProcessItem(ctx context.Context, item interface{}) (interface{}, error) {
	return item, nil
}

func (h *TestConcurrentHandler) Name() string {
	return h.name
}

func TestNewConcurrentStreamProcessor_Basic(t *testing.T) {
	logger := logger.NewLogger()
	handler := &TestConcurrentHandler{name: "test-handler"}

	config := StreamConfig{
		Logger: logger,
	}

	processor := NewConcurrentStreamProcessor(config, handler, 2)

	assert.NotNil(t, processor)
	assert.Equal(t, 2, processor.workers)
	assert.Equal(t, 100, processor.batchSize)   // Default batch size
	assert.Equal(t, 8192, processor.bufferSize) // Default buffer size
	assert.Equal(t, logger, processor.logger)
	assert.Equal(t, handler, processor.processor)
}

func TestNewConcurrentStreamProcessor_DefaultWorkers(t *testing.T) {
	logger := logger.NewLogger()
	handler := &TestConcurrentHandler{name: "test-handler"}

	config := StreamConfig{
		Logger: logger,
	}

	processor := NewConcurrentStreamProcessor(config, handler, 0)

	assert.NotNil(t, processor)
	assert.Equal(t, 4, processor.workers) // Default workers
}

func TestNewMemoryEfficientProcessor_Basic(t *testing.T) {
	logger := logger.NewLogger()
	processorFunc := func(item interface{}) (interface{}, error) {
		return item, nil
	}

	processor := NewMemoryEfficientProcessor(logger, 256, processorFunc)

	assert.NotNil(t, processor)
	assert.Equal(t, 256, processor.maxMemoryMB)
	assert.Equal(t, logger, processor.logger)
	assert.NotNil(t, processor.processor)
}

func TestNewMemoryEfficientProcessor_DefaultMemory(t *testing.T) {
	logger := logger.NewLogger()
	processorFunc := func(item interface{}) (interface{}, error) {
		return item, nil
	}

	processor := NewMemoryEfficientProcessor(logger, 0, processorFunc)

	assert.NotNil(t, processor)
	assert.Equal(t, 256, processor.maxMemoryMB) // Default memory
}

func TestMemoryEfficientProcessor_shouldGC(t *testing.T) {
	logger := logger.NewLogger()
	processorFunc := func(item interface{}) (interface{}, error) {
		return item, nil
	}

	processor := NewMemoryEfficientProcessor(logger, 256, processorFunc)

	// Test shouldGC (currently always returns false)
	shouldGC := processor.shouldGC()
	assert.False(t, shouldGC)
}

func TestProcessResult_Structure(t *testing.T) {
	result := ProcessResult{
		WorkerID: 1,
		Result:   "test result",
		Error:    nil,
		Duration: time.Second,
	}

	assert.Equal(t, 1, result.WorkerID)
	assert.Equal(t, "test result", result.Result)
	assert.NoError(t, result.Error)
	assert.Equal(t, time.Second, result.Duration)
}

func TestStreamConfig_Structure(t *testing.T) {
	logger := logger.NewLogger()
	processorFunc := func(ctx context.Context, item interface{}) (interface{}, error) {
		return item, nil
	}

	config := StreamConfig{
		BatchSize:     50,
		BufferSize:    4096,
		Logger:        logger,
		ProcessorFunc: processorFunc,
	}

	assert.Equal(t, 50, config.BatchSize)
	assert.Equal(t, 4096, config.BufferSize)
	assert.Equal(t, logger, config.Logger)
	assert.NotNil(t, config.ProcessorFunc)
}

func TestBatchProcessor_ConfigValues(t *testing.T) {
	logger := logger.NewLogger()
	handler := &TestBatchHandler{name: "test-handler"}

	// Test with custom config values
	config := StreamConfig{
		BatchSize:  25,
		BufferSize: 2048,
		Logger:     logger,
	}

	processor := NewBatchProcessor(config, handler)

	assert.Equal(t, 25, processor.batchSize)
	assert.Equal(t, 2048, processor.bufferSize)
}

func TestConcurrentStreamProcessor_ConfigValues(t *testing.T) {
	logger := logger.NewLogger()
	handler := &TestConcurrentHandler{name: "test-handler"}

	// Test with custom config values
	config := StreamConfig{
		BatchSize:  25,
		BufferSize: 2048,
		Logger:     logger,
	}

	processor := NewConcurrentStreamProcessor(config, handler, 8)

	assert.Equal(t, 8, processor.workers)
	assert.Equal(t, 25, processor.batchSize)
	assert.Equal(t, 2048, processor.bufferSize)
}