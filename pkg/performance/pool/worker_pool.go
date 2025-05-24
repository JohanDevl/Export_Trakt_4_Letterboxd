package pool

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

// Job represents a unit of work to be processed
type Job interface {
	Execute(ctx context.Context) error
	ID() string
}

// MetricsRecorder interface for recording worker pool metrics
type MetricsRecorder interface {
	IncrementJobsProcessed()
	RecordJobDuration(duration time.Duration)
	IncrementJobsErrored()
}

// Result represents the result of a job execution
type Result struct {
	JobID    string
	Error    error
	Duration time.Duration
	Data     interface{}
}

// WorkerPool represents a pool of workers for concurrent processing
type WorkerPool struct {
	workers    int
	jobs       chan Job
	results    chan Result
	ctx        context.Context
	cancel     context.CancelFunc
	wg         *sync.WaitGroup
	logger     logger.Logger
	metrics    MetricsRecorder
	started    int64
	stopped    int64
	
	// Worker pool statistics
	processedJobs int64
	failedJobs    int64
	totalDuration int64
}

// WorkerPoolConfig holds configuration for worker pool
type WorkerPoolConfig struct {
	Workers      int
	BufferSize   int
	Logger       logger.Logger
	Metrics      MetricsRecorder
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(config WorkerPoolConfig) *WorkerPool {
	if config.Workers <= 0 {
		config.Workers = runtime.NumCPU()
	}
	
	if config.BufferSize <= 0 {
		config.BufferSize = config.Workers * 2
	}

	ctx, cancel := context.WithCancel(context.Background())
	
	return &WorkerPool{
		workers: config.Workers,
		jobs:    make(chan Job, config.BufferSize),
		results: make(chan Result, config.BufferSize),
		ctx:     ctx,
		cancel:  cancel,
		wg:      &sync.WaitGroup{},
		logger:  config.Logger,
		metrics: config.Metrics,
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start() {
	if !atomic.CompareAndSwapInt64(&wp.started, 0, 1) {
		return // Already started
	}

	wp.logger.Info("worker_pool.starting", map[string]interface{}{
		"workers": wp.workers,
	})

	// Start workers
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}

	wp.logger.Info("worker_pool.started", map[string]interface{}{
		"workers": wp.workers,
	})
}

// Stop stops the worker pool gracefully
func (wp *WorkerPool) Stop() {
	if !atomic.CompareAndSwapInt64(&wp.stopped, 0, 1) {
		return // Already stopped
	}

	wp.logger.Info("worker_pool.stopping", nil)
	
	// Close jobs channel to signal workers to stop
	close(wp.jobs)
	
	// Wait for all workers to finish
	wp.wg.Wait()
	
	// Close results channel
	close(wp.results)
	
	// Cancel context
	wp.cancel()
	
	wp.logger.Info("worker_pool.stopped", map[string]interface{}{
		"processed_jobs": atomic.LoadInt64(&wp.processedJobs),
		"failed_jobs":    atomic.LoadInt64(&wp.failedJobs),
		"avg_duration":   wp.getAverageDuration(),
	})
}

// Submit submits a job to the worker pool
func (wp *WorkerPool) Submit(job Job) error {
	select {
	case wp.jobs <- job:
		return nil
	case <-wp.ctx.Done():
		return wp.ctx.Err()
	default:
		return ErrPoolFull
	}
}

// Results returns the results channel
func (wp *WorkerPool) Results() <-chan Result {
	return wp.results
}

// Stats returns current worker pool statistics
func (wp *WorkerPool) Stats() PoolStats {
	return PoolStats{
		Workers:       wp.workers,
		ProcessedJobs: atomic.LoadInt64(&wp.processedJobs),
		FailedJobs:    atomic.LoadInt64(&wp.failedJobs),
		AvgDuration:   wp.getAverageDuration(),
		IsRunning:     atomic.LoadInt64(&wp.started) == 1 && atomic.LoadInt64(&wp.stopped) == 0,
	}
}

// worker is the main worker function
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	
	wp.logger.Debug("worker.started", map[string]interface{}{
		"worker_id": id,
	})

	for job := range wp.jobs {
		start := time.Now()
		
		wp.logger.Debug("worker.processing_job", map[string]interface{}{
			"worker_id": id,
			"job_id":    job.ID(),
		})

		// Execute job with timeout context
		ctx, cancel := context.WithTimeout(wp.ctx, 30*time.Second)
		err := job.Execute(ctx)
		cancel()
		
		duration := time.Since(start)
		
		// Update statistics
		atomic.AddInt64(&wp.processedJobs, 1)
		atomic.AddInt64(&wp.totalDuration, int64(duration))
		
		if err != nil {
			atomic.AddInt64(&wp.failedJobs, 1)
			wp.logger.Error("worker.job_failed", map[string]interface{}{
				"worker_id": id,
				"job_id":    job.ID(),
				"error":     err.Error(),
				"duration":  duration.String(),
			})
		} else {
			wp.logger.Debug("worker.job_completed", map[string]interface{}{
				"worker_id": id,
				"job_id":    job.ID(),
				"duration":  duration.String(),
			})
		}

		// Send result
		result := Result{
			JobID:    job.ID(),
			Error:    err,
			Duration: duration,
		}

		select {
		case wp.results <- result:
		case <-wp.ctx.Done():
			return
		}

		// Update metrics if available
		if wp.metrics != nil {
			wp.metrics.IncrementJobsProcessed()
			wp.metrics.RecordJobDuration(duration)
			if err != nil {
				wp.metrics.IncrementJobsErrored()
			}
		}
	}
	
	wp.logger.Debug("worker.stopped", map[string]interface{}{
		"worker_id": id,
	})
}

// getAverageDuration calculates the average job duration
func (wp *WorkerPool) getAverageDuration() time.Duration {
	totalJobs := atomic.LoadInt64(&wp.processedJobs)
	if totalJobs == 0 {
		return 0
	}
	
	totalDuration := atomic.LoadInt64(&wp.totalDuration)
	return time.Duration(totalDuration / totalJobs)
}

// PoolStats represents worker pool statistics
type PoolStats struct {
	Workers       int
	ProcessedJobs int64
	FailedJobs    int64
	AvgDuration   time.Duration
	IsRunning     bool
}

// Errors
var (
	ErrPoolFull = fmt.Errorf("worker pool is full")
) 