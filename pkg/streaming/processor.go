package streaming

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

// StreamProcessor interface for processing data streams
type StreamProcessor interface {
	Process(ctx context.Context, reader io.Reader, writer io.Writer) error
	SetBatchSize(size int)
	SetBufferSize(size int)
}

// BatchProcessor processes data in batches
type BatchProcessor struct {
	logger     logger.Logger
	batchSize  int
	bufferSize int
	processor  BatchHandler
}

// BatchHandler interface for handling batches of data
type BatchHandler interface {
	ProcessBatch(ctx context.Context, batch []interface{}) error
	Name() string
}

// StreamConfig holds configuration for stream processing
type StreamConfig struct {
	BatchSize    int
	BufferSize   int
	Logger       logger.Logger
	ProcessorFunc func(ctx context.Context, item interface{}) (interface{}, error)
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(config StreamConfig, handler BatchHandler) *BatchProcessor {
	if config.BatchSize <= 0 {
		config.BatchSize = 100
	}
	
	if config.BufferSize <= 0 {
		config.BufferSize = 8192
	}

	return &BatchProcessor{
		logger:     config.Logger,
		batchSize:  config.BatchSize,
		bufferSize: config.BufferSize,
		processor:  handler,
	}
}

// Process processes a stream of JSON objects
func (bp *BatchProcessor) Process(ctx context.Context, reader io.Reader, writer io.Writer) error {
	bufferedReader := bufio.NewReaderSize(reader, bp.bufferSize)
	decoder := json.NewDecoder(bufferedReader)
	encoder := json.NewEncoder(writer)
	
	batch := make([]interface{}, 0, bp.batchSize)
	processed := 0
	
	bp.logger.Info("stream.processing_started", map[string]interface{}{
		"batch_size":  bp.batchSize,
		"buffer_size": bp.bufferSize,
		"processor":   bp.processor.Name(),
	})
	
	start := time.Now()
	defer func() {
		bp.logger.Info("stream.processing_completed", map[string]interface{}{
			"processed_items": processed,
			"duration":        time.Since(start).String(),
			"processor":       bp.processor.Name(),
		})
	}()

	for decoder.More() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var item interface{}
		if err := decoder.Decode(&item); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode JSON: %w", err)
		}

		batch = append(batch, item)
		
		// Process batch when full
		if len(batch) >= bp.batchSize {
			if err := bp.processBatch(ctx, batch); err != nil {
				return fmt.Errorf("failed to process batch: %w", err)
			}
			
			// Write batch results (simplified - in real implementation, you'd handle results)
			for _, processedItem := range batch {
				if err := encoder.Encode(processedItem); err != nil {
					return fmt.Errorf("failed to encode result: %w", err)
				}
			}
			
			processed += len(batch)
			batch = batch[:0] // Reset batch
		}
	}

	// Process remaining items
	if len(batch) > 0 {
		if err := bp.processBatch(ctx, batch); err != nil {
			return fmt.Errorf("failed to process final batch: %w", err)
		}
		
		for _, processedItem := range batch {
			if err := encoder.Encode(processedItem); err != nil {
				return fmt.Errorf("failed to encode final result: %w", err)
			}
		}
		
		processed += len(batch)
	}

	return nil
}

// SetBatchSize sets the batch size
func (bp *BatchProcessor) SetBatchSize(size int) {
	if size > 0 {
		bp.batchSize = size
	}
}

// SetBufferSize sets the buffer size
func (bp *BatchProcessor) SetBufferSize(size int) {
	if size > 0 {
		bp.bufferSize = size
	}
}

// processBatch processes a batch of items
func (bp *BatchProcessor) processBatch(ctx context.Context, batch []interface{}) error {
	start := time.Now()
	
	err := bp.processor.ProcessBatch(ctx, batch)
	
	duration := time.Since(start)
	bp.logger.Debug("stream.batch_processed", map[string]interface{}{
		"batch_size": len(batch),
		"duration":   duration.String(),
		"processor":  bp.processor.Name(),
		"success":    err == nil,
	})
	
	return err
}

// ConcurrentStreamProcessor processes data using multiple goroutines
type ConcurrentStreamProcessor struct {
	logger      logger.Logger
	workers     int
	batchSize   int
	bufferSize  int
	processor   ConcurrentHandler
}

// ConcurrentHandler interface for concurrent processing
type ConcurrentHandler interface {
	ProcessItem(ctx context.Context, item interface{}) (interface{}, error)
	Name() string
}

// NewConcurrentStreamProcessor creates a new concurrent stream processor
func NewConcurrentStreamProcessor(config StreamConfig, handler ConcurrentHandler, workers int) *ConcurrentStreamProcessor {
	if config.BatchSize <= 0 {
		config.BatchSize = 100
	}
	
	if config.BufferSize <= 0 {
		config.BufferSize = 8192
	}
	
	if workers <= 0 {
		workers = 4
	}

	return &ConcurrentStreamProcessor{
		logger:     config.Logger,
		workers:    workers,
		batchSize:  config.BatchSize,
		bufferSize: config.BufferSize,
		processor:  handler,
	}
}

// Process processes a stream with concurrent workers
func (csp *ConcurrentStreamProcessor) Process(ctx context.Context, reader io.Reader, writer io.Writer) error {
	bufferedReader := bufio.NewReaderSize(reader, csp.bufferSize)
	decoder := json.NewDecoder(bufferedReader)
	encoder := json.NewEncoder(writer)
	
	// Channels for worker communication
	jobs := make(chan interface{}, csp.workers*2)
	results := make(chan ProcessResult, csp.workers*2)
	
	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < csp.workers; i++ {
		wg.Add(1)
		go csp.worker(ctx, i, jobs, results, &wg)
	}
	
	// Start result collector
	var collectorWg sync.WaitGroup
	collectorWg.Add(1)
	go csp.resultCollector(ctx, results, encoder, &collectorWg)
	
	csp.logger.Info("concurrent_stream.processing_started", map[string]interface{}{
		"workers":     csp.workers,
		"batch_size":  csp.batchSize,
		"buffer_size": csp.bufferSize,
		"processor":   csp.processor.Name(),
	})
	
	start := time.Now()
	processed := 0
	
	// Send jobs to workers
	for decoder.More() {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			close(results)
			collectorWg.Wait()
			return ctx.Err()
		default:
		}

		var item interface{}
		if err := decoder.Decode(&item); err != nil {
			if err == io.EOF {
				break
			}
			close(jobs)
			wg.Wait()
			close(results)
			collectorWg.Wait()
			return fmt.Errorf("failed to decode JSON: %w", err)
		}

		jobs <- item
		processed++
	}
	
	// Signal completion and wait for workers
	close(jobs)
	wg.Wait()
	close(results)
	collectorWg.Wait()
	
	csp.logger.Info("concurrent_stream.processing_completed", map[string]interface{}{
		"processed_items": processed,
		"duration":        time.Since(start).String(),
		"workers":         csp.workers,
		"processor":       csp.processor.Name(),
	})
	
	return nil
}

// worker processes items from the jobs channel
func (csp *ConcurrentStreamProcessor) worker(ctx context.Context, id int, jobs <-chan interface{}, results chan<- ProcessResult, wg *sync.WaitGroup) {
	defer wg.Done()
	
	csp.logger.Debug("concurrent_worker.started", map[string]interface{}{
		"worker_id": id,
		"processor": csp.processor.Name(),
	})
	
	for job := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
		}
		
		start := time.Now()
		result, err := csp.processor.ProcessItem(ctx, job)
		duration := time.Since(start)
		
		processResult := ProcessResult{
			WorkerID: id,
			Result:   result,
			Error:    err,
			Duration: duration,
		}
		
		select {
		case results <- processResult:
		case <-ctx.Done():
			return
		}
	}
	
	csp.logger.Debug("concurrent_worker.stopped", map[string]interface{}{
		"worker_id": id,
		"processor": csp.processor.Name(),
	})
}

// resultCollector collects and writes results
func (csp *ConcurrentStreamProcessor) resultCollector(ctx context.Context, results <-chan ProcessResult, encoder *json.Encoder, wg *sync.WaitGroup) {
	defer wg.Done()
	
	processed := 0
	errors := 0
	
	for result := range results {
		select {
		case <-ctx.Done():
			return
		default:
		}
		
		if result.Error != nil {
			errors++
			csp.logger.Error("concurrent_stream.processing_error", map[string]interface{}{
				"worker_id": result.WorkerID,
				"error":     result.Error.Error(),
				"duration":  result.Duration.String(),
			})
			continue
		}
		
		if err := encoder.Encode(result.Result); err != nil {
			csp.logger.Error("concurrent_stream.encoding_error", map[string]interface{}{
				"worker_id": result.WorkerID,
				"error":     err.Error(),
			})
			continue
		}
		
		processed++
		
		if processed%1000 == 0 {
			csp.logger.Info("concurrent_stream.progress", map[string]interface{}{
				"processed": processed,
				"errors":    errors,
			})
		}
	}
	
	csp.logger.Info("concurrent_stream.collection_completed", map[string]interface{}{
		"processed": processed,
		"errors":    errors,
	})
}

// ProcessResult represents the result of processing an item
type ProcessResult struct {
	WorkerID int
	Result   interface{}
	Error    error
	Duration time.Duration
}

// MemoryEfficientProcessor processes data with memory optimization
type MemoryEfficientProcessor struct {
	logger        logger.Logger
	maxMemoryMB   int
	currentMemory int64
	processor     func(interface{}) (interface{}, error)
}

// NewMemoryEfficientProcessor creates a memory-efficient processor
func NewMemoryEfficientProcessor(logger logger.Logger, maxMemoryMB int, processor func(interface{}) (interface{}, error)) *MemoryEfficientProcessor {
	if maxMemoryMB <= 0 {
		maxMemoryMB = 256 // Default 256MB
	}

	return &MemoryEfficientProcessor{
		logger:      logger,
		maxMemoryMB: maxMemoryMB,
		processor:   processor,
	}
}

// Process processes data with memory monitoring
func (mep *MemoryEfficientProcessor) Process(ctx context.Context, reader io.Reader, writer io.Writer) error {
	scanner := bufio.NewScanner(reader)
	encoder := json.NewEncoder(writer)
	
	processed := 0
	start := time.Now()
	
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		var item interface{}
		if err := json.Unmarshal(scanner.Bytes(), &item); err != nil {
			mep.logger.Warn("memory_efficient.decode_error", map[string]interface{}{
				"error": err.Error(),
			})
			continue
		}
		
		result, err := mep.processor(item)
		if err != nil {
			mep.logger.Error("memory_efficient.processing_error", map[string]interface{}{
				"error": err.Error(),
			})
			continue
		}
		
		if err := encoder.Encode(result); err != nil {
			mep.logger.Error("memory_efficient.encoding_error", map[string]interface{}{
				"error": err.Error(),
			})
			continue
		}
		
		processed++
		
		// Memory check every 100 items
		if processed%100 == 0 {
			// In a real implementation, you would check actual memory usage
			// For demonstration, we'll simulate it
			if mep.shouldGC() {
				mep.logger.Debug("memory_efficient.gc_triggered", map[string]interface{}{
					"processed": processed,
				})
			}
		}
	}
	
	mep.logger.Info("memory_efficient.processing_completed", map[string]interface{}{
		"processed_items": processed,
		"duration":        time.Since(start).String(),
	})
	
	return scanner.Err()
}

// shouldGC determines if garbage collection should be triggered
func (mep *MemoryEfficientProcessor) shouldGC() bool {
	// Simplified memory check - in real implementation, use runtime.MemStats
	return false
} 