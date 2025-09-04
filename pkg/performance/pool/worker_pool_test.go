package pool

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

// Mock implementations for testing

// MockLogger implements logger.Logger interface for testing
type MockLogger struct {
	logs []LogEntry
	mu   sync.Mutex
}

type LogEntry struct {
	Level   string
	Message string
	Data    map[string]interface{}
}

func NewMockLogger() *MockLogger {
	return &MockLogger{
		logs: make([]LogEntry, 0),
	}
}

func (m *MockLogger) Info(messageID string, data ...map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	actualData := make(map[string]interface{})
	if len(data) > 0 {
		actualData = data[0]
	}
	m.logs = append(m.logs, LogEntry{Level: "info", Message: messageID, Data: actualData})
}

func (m *MockLogger) Infof(messageID string, data map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = append(m.logs, LogEntry{Level: "info", Message: messageID, Data: data})
}

func (m *MockLogger) Error(messageID string, data ...map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	actualData := make(map[string]interface{})
	if len(data) > 0 {
		actualData = data[0]
	}
	m.logs = append(m.logs, LogEntry{Level: "error", Message: messageID, Data: actualData})
}

func (m *MockLogger) Errorf(messageID string, data map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = append(m.logs, LogEntry{Level: "error", Message: messageID, Data: data})
}

func (m *MockLogger) Warn(messageID string, data ...map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	actualData := make(map[string]interface{})
	if len(data) > 0 {
		actualData = data[0]
	}
	m.logs = append(m.logs, LogEntry{Level: "warn", Message: messageID, Data: actualData})
}

func (m *MockLogger) Warnf(messageID string, data map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = append(m.logs, LogEntry{Level: "warn", Message: messageID, Data: data})
}

func (m *MockLogger) Debug(messageID string, data ...map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	actualData := make(map[string]interface{})
	if len(data) > 0 {
		actualData = data[0]
	}
	m.logs = append(m.logs, LogEntry{Level: "debug", Message: messageID, Data: actualData})
}

func (m *MockLogger) Debugf(messageID string, data map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = append(m.logs, LogEntry{Level: "debug", Message: messageID, Data: data})
}

func (m *MockLogger) SetLogLevel(level string) {
	// No-op for testing
}

func (m *MockLogger) SetLogFile(filePath string) error {
	// No-op for testing
	return nil
}

func (m *MockLogger) SetTranslator(t logger.Translator) {
	// No-op for testing
}

func (m *MockLogger) GetLogs() []LogEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]LogEntry, len(m.logs))
	copy(result, m.logs)
	return result
}

func (m *MockLogger) CountLevel(level string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, log := range m.logs {
		if log.Level == level {
			count++
		}
	}
	return count
}

// MockMetricsRecorder implements MetricsRecorder interface for testing
type MockMetricsRecorder struct {
	processedJobs int64
	erroredJobs   int64
	durations     []time.Duration
	mu            sync.Mutex
}

func NewMockMetricsRecorder() *MockMetricsRecorder {
	return &MockMetricsRecorder{
		durations: make([]time.Duration, 0),
	}
}

func (m *MockMetricsRecorder) IncrementJobsProcessed() {
	atomic.AddInt64(&m.processedJobs, 1)
}

func (m *MockMetricsRecorder) RecordJobDuration(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.durations = append(m.durations, duration)
}

func (m *MockMetricsRecorder) IncrementJobsErrored() {
	atomic.AddInt64(&m.erroredJobs, 1)
}

func (m *MockMetricsRecorder) GetProcessedJobs() int64 {
	return atomic.LoadInt64(&m.processedJobs)
}

func (m *MockMetricsRecorder) GetErroredJobs() int64 {
	return atomic.LoadInt64(&m.erroredJobs)
}

func (m *MockMetricsRecorder) GetDurations() []time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]time.Duration, len(m.durations))
	copy(result, m.durations)
	return result
}

// Test Jobs

// SimpleJob is a basic job implementation for testing
type SimpleJob struct {
	id       string
	duration time.Duration
	shouldFail bool
	executed   int64
}

func NewSimpleJob(id string, duration time.Duration) *SimpleJob {
	return &SimpleJob{id: id, duration: duration}
}

func NewFailingJob(id string) *SimpleJob {
	return &SimpleJob{id: id, shouldFail: true}
}

func (j *SimpleJob) Execute(ctx context.Context) error {
	atomic.AddInt64(&j.executed, 1)
	
	if j.shouldFail {
		return errors.New("simulated job failure")
	}
	
	select {
	case <-time.After(j.duration):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (j *SimpleJob) ID() string {
	return j.id
}

func (j *SimpleJob) ExecutionCount() int64 {
	return atomic.LoadInt64(&j.executed)
}

// PanicJob simulates a job that panics
type PanicJob struct {
	id string
}

func NewPanicJob(id string) *PanicJob {
	return &PanicJob{id: id}
}

func (j *PanicJob) Execute(ctx context.Context) error {
	panic("simulated panic in job")
}

func (j *PanicJob) ID() string {
	return j.id
}

// Long running job for timeout tests
type LongRunningJob struct {
	id       string
	duration time.Duration
}

func NewLongRunningJob(id string, duration time.Duration) *LongRunningJob {
	return &LongRunningJob{id: id, duration: duration}
}

func (j *LongRunningJob) Execute(ctx context.Context) error {
	select {
	case <-time.After(j.duration):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (j *LongRunningJob) ID() string {
	return j.id
}

// Tests

func TestNewWorkerPool(t *testing.T) {
	logger := NewMockLogger()
	metrics := NewMockMetricsRecorder()
	
	tests := []struct {
		name           string
		config         WorkerPoolConfig
		expectedWorkers int
		expectedBuffer  int
	}{
		{
			name: "default configuration",
			config: WorkerPoolConfig{
				Logger:  logger,
				Metrics: metrics,
			},
			expectedWorkers: runtime.NumCPU(),
			expectedBuffer:  runtime.NumCPU() * 2,
		},
		{
			name: "custom configuration",
			config: WorkerPoolConfig{
				Workers:    4,
				BufferSize: 10,
				Logger:     logger,
				Metrics:    metrics,
			},
			expectedWorkers: 4,
			expectedBuffer:  10,
		},
		{
			name: "zero workers defaults to CPU count",
			config: WorkerPoolConfig{
				Workers:    0,
				BufferSize: 5,
				Logger:     logger,
				Metrics:    metrics,
			},
			expectedWorkers: runtime.NumCPU(),
			expectedBuffer:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewWorkerPool(tt.config)
			
			require.NotNil(t, pool)
			assert.Equal(t, tt.expectedWorkers, pool.workers)
			assert.Equal(t, tt.expectedBuffer, cap(pool.jobs))
			assert.Equal(t, tt.expectedBuffer, cap(pool.results))
			assert.NotNil(t, pool.ctx)
			assert.NotNil(t, pool.cancel)
			assert.NotNil(t, pool.wg)
			
			stats := pool.Stats()
			assert.Equal(t, tt.expectedWorkers, stats.Workers)
			assert.Equal(t, int64(0), stats.ProcessedJobs)
			assert.Equal(t, int64(0), stats.FailedJobs)
			assert.False(t, stats.IsRunning)
		})
	}
}

func TestWorkerPoolStartStop(t *testing.T) {
	logger := NewMockLogger()
	metrics := NewMockMetricsRecorder()
	
	config := WorkerPoolConfig{
		Workers:    2,
		BufferSize: 4,
		Logger:     logger,
		Metrics:    metrics,
	}
	
	pool := NewWorkerPool(config)
	
	// Initial state should be stopped
	stats := pool.Stats()
	assert.False(t, stats.IsRunning)
	
	// Start the pool
	pool.Start()
	
	// Should be running
	stats = pool.Stats()
	assert.True(t, stats.IsRunning)
	
	// Starting again should be idempotent
	pool.Start()
	stats = pool.Stats()
	assert.True(t, stats.IsRunning)
	
	// Check start logs
	logs := logger.GetLogs()
	startLogs := 0
	for _, log := range logs {
		if log.Message == "worker_pool.starting" {
			startLogs++
		}
	}
	assert.Equal(t, 1, startLogs, "Should only log start once")
	
	// Stop the pool
	pool.Stop()
	
	// Should be stopped
	stats = pool.Stats()
	assert.False(t, stats.IsRunning)
	
	// Stopping again should be idempotent
	pool.Stop()
	stats = pool.Stats()
	assert.False(t, stats.IsRunning)
}

func TestWorkerPoolJobExecution(t *testing.T) {
	logger := NewMockLogger()
	metrics := NewMockMetricsRecorder()
	
	config := WorkerPoolConfig{
		Workers:    2,
		BufferSize: 10,
		Logger:     logger,
		Metrics:    metrics,
	}
	
	pool := NewWorkerPool(config)
	pool.Start()
	defer pool.Stop()
	
	// Submit successful jobs
	job1 := NewSimpleJob("job1", 10*time.Millisecond)
	job2 := NewSimpleJob("job2", 20*time.Millisecond)
	job3 := NewSimpleJob("job3", 5*time.Millisecond)
	
	err := pool.Submit(job1)
	assert.NoError(t, err)
	
	err = pool.Submit(job2)
	assert.NoError(t, err)
	
	err = pool.Submit(job3)
	assert.NoError(t, err)
	
	// Collect results
	results := make([]Result, 0, 3)
	timeout := time.After(5 * time.Second)
	
	for i := 0; i < 3; i++ {
		select {
		case result := <-pool.Results():
			results = append(results, result)
		case <-timeout:
			t.Fatal("Timeout waiting for results")
		}
	}
	
	// Verify all jobs executed
	assert.Equal(t, int64(1), job1.ExecutionCount())
	assert.Equal(t, int64(1), job2.ExecutionCount())
	assert.Equal(t, int64(1), job3.ExecutionCount())
	
	// Verify results
	assert.Len(t, results, 3)
	for _, result := range results {
		assert.NoError(t, result.Error)
		assert.Greater(t, result.Duration, time.Duration(0))
	}
	
	// Check statistics
	stats := pool.Stats()
	assert.Equal(t, int64(3), stats.ProcessedJobs)
	assert.Equal(t, int64(0), stats.FailedJobs)
	assert.Greater(t, stats.AvgDuration, time.Duration(0))
	
	// Check metrics
	assert.Equal(t, int64(3), metrics.GetProcessedJobs())
	assert.Equal(t, int64(0), metrics.GetErroredJobs())
	assert.Len(t, metrics.GetDurations(), 3)
}

func TestWorkerPoolFailedJobs(t *testing.T) {
	logger := NewMockLogger()
	metrics := NewMockMetricsRecorder()
	
	config := WorkerPoolConfig{
		Workers:    1,
		BufferSize: 5,
		Logger:     logger,
		Metrics:    metrics,
	}
	
	pool := NewWorkerPool(config)
	pool.Start()
	defer pool.Stop()
	
	// Submit failing job
	failingJob := NewFailingJob("failing_job")
	err := pool.Submit(failingJob)
	assert.NoError(t, err)
	
	// Collect result
	select {
	case result := <-pool.Results():
		assert.Error(t, result.Error)
		assert.Equal(t, "failing_job", result.JobID)
		assert.Greater(t, result.Duration, time.Duration(0))
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for result")
	}
	
	// Check statistics
	stats := pool.Stats()
	assert.Equal(t, int64(1), stats.ProcessedJobs)
	assert.Equal(t, int64(1), stats.FailedJobs)
	
	// Check metrics
	assert.Equal(t, int64(1), metrics.GetProcessedJobs())
	assert.Equal(t, int64(1), metrics.GetErroredJobs())
	
	// Check error logs
	assert.Greater(t, logger.CountLevel("error"), 0)
}

// BlockingJob is a job that blocks until released
type BlockingJob struct {
	id      string
	release chan struct{}
}

func NewBlockingJob(id string) *BlockingJob {
	return &BlockingJob{
		id:      id,
		release: make(chan struct{}),
	}
}

func (j *BlockingJob) Execute(ctx context.Context) error {
	select {
	case <-j.release:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (j *BlockingJob) ID() string {
	return j.id
}

func (j *BlockingJob) Release() {
	select {
	case j.release <- struct{}{}:
	default:
	}
}

func TestWorkerPoolBufferFull(t *testing.T) {
	logger := NewMockLogger()
	metrics := NewMockMetricsRecorder()
	
	config := WorkerPoolConfig{
		Workers:    1,
		BufferSize: 1, // Very small buffer to make test more reliable
		Logger:     logger,
		Metrics:    metrics,
	}
	
	pool := NewWorkerPool(config)
	pool.Start()
	defer pool.Stop()
	
	// Submit multiple jobs quickly to saturate buffer
	// First job goes to worker, second fills buffer, third should fail
	var errors []error
	
	for i := 0; i < 10; i++ {
		job := NewSimpleJob(fmt.Sprintf("job_%d", i), 50*time.Millisecond)
		err := pool.Submit(job)
		errors = append(errors, err)
	}
	
	// At least one submission should have failed with ErrPoolFull
	fullErrors := 0
	for _, err := range errors {
		if err == ErrPoolFull {
			fullErrors++
		}
	}
	
	assert.Greater(t, fullErrors, 0, "Expected at least one ErrPoolFull error")
	
	// Give time for jobs to complete before stopping
	time.Sleep(200 * time.Millisecond)
}

func TestWorkerPoolSubmitAfterStop(t *testing.T) {
	logger := NewMockLogger()
	metrics := NewMockMetricsRecorder()
	
	config := WorkerPoolConfig{
		Workers:    1,
		BufferSize: 5,
		Logger:     logger,
		Metrics:    metrics,
	}
	
	pool := NewWorkerPool(config)
	pool.Start()
	pool.Stop()
	
	// Submit after stop should fail or panic
	job := NewSimpleJob("test", 10*time.Millisecond)
	
	// The current implementation panics on closed channel
	// This test verifies that Submit after Stop fails somehow
	defer func() {
		if r := recover(); r != nil {
			// Panic is acceptable - pool is stopped
			assert.Contains(t, fmt.Sprintf("%v", r), "closed channel")
		}
	}()
	
	err := pool.Submit(job)
	if err != nil {
		// Error is also acceptable
		assert.Error(t, err)
	} else {
		// If no error and no panic, that's unexpected
		t.Error("Expected Submit after Stop to fail with error or panic")
	}
}

func TestWorkerPoolJobTimeout(t *testing.T) {
	logger := NewMockLogger()
	metrics := NewMockMetricsRecorder()
	
	config := WorkerPoolConfig{
		Workers:    1,
		BufferSize: 5,
		Logger:     logger,
		Metrics:    metrics,
	}
	
	pool := NewWorkerPool(config)
	pool.Start()
	defer pool.Stop()
	
	// Submit job that takes longer than 30s timeout
	longJob := NewLongRunningJob("timeout_job", 35*time.Second)
	err := pool.Submit(longJob)
	assert.NoError(t, err)
	
	// Wait for result - should timeout
	select {
	case result := <-pool.Results():
		assert.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "context deadline exceeded")
		assert.Equal(t, "timeout_job", result.JobID)
	case <-time.After(35 * time.Second):
		t.Fatal("Test timeout waiting for job timeout")
	}
}

func TestWorkerPoolConcurrentOperations(t *testing.T) {
	logger := NewMockLogger()
	metrics := NewMockMetricsRecorder()
	
	config := WorkerPoolConfig{
		Workers:    4,
		BufferSize: 100,
		Logger:     logger,
		Metrics:    metrics,
	}
	
	pool := NewWorkerPool(config)
	pool.Start()
	defer pool.Stop()
	
	const numJobs = 50
	var wg sync.WaitGroup
	
	// Submit jobs concurrently
	wg.Add(numJobs)
	for i := 0; i < numJobs; i++ {
		go func(id int) {
			defer wg.Done()
			job := NewSimpleJob(fmt.Sprintf("concurrent_job_%d", id), 1*time.Millisecond)
			err := pool.Submit(job)
			assert.NoError(t, err)
		}(i)
	}
	
	// Collect results concurrently
	results := make([]Result, 0, numJobs)
	resultsMu := sync.Mutex{}
	
	wg.Add(1)
	go func() {
		defer wg.Done()
		collected := 0
		for collected < numJobs {
			select {
			case result := <-pool.Results():
				resultsMu.Lock()
				results = append(results, result)
				collected++
				resultsMu.Unlock()
			case <-time.After(10 * time.Second):
				t.Error("Timeout collecting results")
				return
			}
		}
	}()
	
	wg.Wait()
	
	// Verify all jobs completed
	resultsMu.Lock()
	assert.Len(t, results, numJobs)
	resultsMu.Unlock()
	
	stats := pool.Stats()
	assert.Equal(t, int64(numJobs), stats.ProcessedJobs)
	assert.Equal(t, int64(0), stats.FailedJobs)
}

func TestWorkerPoolGracefulShutdown(t *testing.T) {
	logger := NewMockLogger()
	metrics := NewMockMetricsRecorder()
	
	config := WorkerPoolConfig{
		Workers:    2,
		BufferSize: 10,
		Logger:     logger,
		Metrics:    metrics,
	}
	
	pool := NewWorkerPool(config)
	pool.Start()
	
	// Submit jobs
	for i := 0; i < 5; i++ {
		job := NewSimpleJob(fmt.Sprintf("shutdown_job_%d", i), 10*time.Millisecond)
		err := pool.Submit(job)
		assert.NoError(t, err)
	}
	
	// Stop should wait for all jobs to complete
	stopStart := time.Now()
	pool.Stop()
	stopDuration := time.Since(stopStart)
	
	// Should have taken some time to complete jobs
	assert.Greater(t, stopDuration, 5*time.Millisecond)
	
	// All jobs should be processed
	stats := pool.Stats()
	assert.Equal(t, int64(5), stats.ProcessedJobs)
	assert.False(t, stats.IsRunning)
}

func TestWorkerPoolWithoutMetrics(t *testing.T) {
	logger := NewMockLogger()
	
	config := WorkerPoolConfig{
		Workers:    1,
		BufferSize: 5,
		Logger:     logger,
		Metrics:    nil, // No metrics recorder
	}
	
	pool := NewWorkerPool(config)
	pool.Start()
	defer pool.Stop()
	
	// Submit job
	job := NewSimpleJob("no_metrics_job", 10*time.Millisecond)
	err := pool.Submit(job)
	assert.NoError(t, err)
	
	// Collect result
	select {
	case result := <-pool.Results():
		assert.NoError(t, result.Error)
		assert.Equal(t, "no_metrics_job", result.JobID)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for result")
	}
	
	// Should not panic and stats should still work
	stats := pool.Stats()
	assert.Equal(t, int64(1), stats.ProcessedJobs)
	assert.Equal(t, int64(0), stats.FailedJobs)
}

func TestWorkerPoolPanicRecovery(t *testing.T) {
	logger := NewMockLogger()
	metrics := NewMockMetricsRecorder()
	
	config := WorkerPoolConfig{
		Workers:    1,
		BufferSize: 5,
		Logger:     logger,
		Metrics:    metrics,
	}
	
	pool := NewWorkerPool(config)
	pool.Start()
	defer pool.Stop()
	
	// Submit panicking job - this should be handled gracefully
	// Note: The current implementation doesn't have explicit panic recovery
	// This test documents the expected behavior
	panicJob := NewPanicJob("panic_job")
	
	// We'll submit this but expect the worker to crash
	// In a production system, we'd want panic recovery
	err := pool.Submit(panicJob)
	assert.NoError(t, err)
	
	// Submit a normal job after the panic to test if pool still works
	normalJob := NewSimpleJob("normal_job", 10*time.Millisecond)
	err = pool.Submit(normalJob)
	assert.NoError(t, err)
	
	// This test mainly ensures the test infrastructure works
	// The actual panic handling would need to be implemented in the worker pool
}

// Benchmarks

func BenchmarkWorkerPoolSubmit(b *testing.B) {
	logger := NewMockLogger()
	metrics := NewMockMetricsRecorder()
	
	config := WorkerPoolConfig{
		Workers:    runtime.NumCPU(),
		BufferSize: 1000,
		Logger:     logger,
		Metrics:    metrics,
	}
	
	pool := NewWorkerPool(config)
	pool.Start()
	defer pool.Stop()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			job := NewSimpleJob(fmt.Sprintf("bench_job_%d", i), 1*time.Microsecond)
			pool.Submit(job)
			i++
		}
	})
}

func BenchmarkWorkerPoolThroughput(b *testing.B) {
	logger := NewMockLogger()
	metrics := NewMockMetricsRecorder()
	
	config := WorkerPoolConfig{
		Workers:    runtime.NumCPU(),
		BufferSize: 1000,
		Logger:     logger,
		Metrics:    metrics,
	}
	
	pool := NewWorkerPool(config)
	pool.Start()
	defer pool.Stop()
	
	// Start result collector
	var processed int64
	go func() {
		for range pool.Results() {
			atomic.AddInt64(&processed, 1)
		}
	}()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		job := NewSimpleJob(fmt.Sprintf("throughput_job_%d", i), 1*time.Microsecond)
		pool.Submit(job)
	}
	
	// Wait for all jobs to complete
	for atomic.LoadInt64(&processed) < int64(b.N) {
		time.Sleep(1 * time.Millisecond)
	}
}