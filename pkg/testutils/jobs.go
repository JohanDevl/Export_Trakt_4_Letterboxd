package testutils

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// TestJob provides a configurable job implementation for testing worker pools and job processors
type TestJob struct {
	id          string
	duration    time.Duration
	shouldError bool
	errorMsg    string
	shouldPanic bool
	panicMsg    string
	executed    int64
	onExecute   func(ctx context.Context) error
}

// NewTestJob creates a new test job with the given ID
func NewTestJob(id string) *TestJob {
	return &TestJob{
		id: id,
	}
}

// NewSimpleTestJob creates a simple test job that just sleeps for the given duration
func NewSimpleTestJob(id string, duration time.Duration) *TestJob {
	return &TestJob{
		id:       id,
		duration: duration,
	}
}

// NewErrorTestJob creates a test job that returns an error
func NewErrorTestJob(id, errorMsg string) *TestJob {
	return &TestJob{
		id:          id,
		shouldError: true,
		errorMsg:    errorMsg,
	}
}

// NewPanicTestJob creates a test job that panics
func NewPanicTestJob(id, panicMsg string) *TestJob {
	return &TestJob{
		id:          id,
		shouldPanic: true,
		panicMsg:    panicMsg,
	}
}

// ID returns the job ID
func (j *TestJob) ID() string {
	return j.id
}

// Execute runs the job
func (j *TestJob) Execute(ctx context.Context) error {
	atomic.AddInt64(&j.executed, 1)
	
	if j.shouldPanic {
		panic(j.panicMsg)
	}
	
	if j.onExecute != nil {
		return j.onExecute(ctx)
	}
	
	if j.duration > 0 {
		select {
		case <-time.After(j.duration):
			// Duration completed
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	
	if j.shouldError {
		return fmt.Errorf(j.errorMsg)
	}
	
	return nil
}

// SetDuration sets the execution duration
func (j *TestJob) SetDuration(duration time.Duration) *TestJob {
	j.duration = duration
	return j
}

// SetError configures the job to return an error
func (j *TestJob) SetError(errorMsg string) *TestJob {
	j.shouldError = true
	j.errorMsg = errorMsg
	return j
}

// SetPanic configures the job to panic
func (j *TestJob) SetPanic(panicMsg string) *TestJob {
	j.shouldPanic = true
	j.panicMsg = panicMsg
	return j
}

// SetOnExecute sets a custom execution function
func (j *TestJob) SetOnExecute(fn func(ctx context.Context) error) *TestJob {
	j.onExecute = fn
	return j
}

// ExecutionCount returns how many times the job has been executed
func (j *TestJob) ExecutionCount() int64 {
	return atomic.LoadInt64(&j.executed)
}

// WaitJob provides a job that waits for a signal before completing
type WaitJob struct {
	*TestJob
	waitCh chan struct{}
}

// NewWaitJob creates a job that waits for a signal
func NewWaitJob(id string) *WaitJob {
	return &WaitJob{
		TestJob: NewTestJob(id),
		waitCh:  make(chan struct{}),
	}
}

// Execute waits for the signal or context cancellation
func (j *WaitJob) Execute(ctx context.Context) error {
	atomic.AddInt64(&j.executed, 1)
	
	if j.shouldPanic {
		panic(j.panicMsg)
	}
	
	select {
	case <-j.waitCh:
		if j.shouldError {
			return fmt.Errorf(j.errorMsg)
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Signal allows the job to complete
func (j *WaitJob) Signal() {
	close(j.waitCh)
}

// CounterJob provides a job that increments a shared counter
type CounterJob struct {
	id      string
	counter *int64
}

// NewCounterJob creates a job that increments the given counter
func NewCounterJob(id string, counter *int64) *CounterJob {
	return &CounterJob{
		id:      id,
		counter: counter,
	}
}

// ID returns the job ID
func (j *CounterJob) ID() string {
	return j.id
}

// Execute increments the counter
func (j *CounterJob) Execute(ctx context.Context) error {
	atomic.AddInt64(j.counter, 1)
	return nil
}

// BatchJob provides a job that can process multiple items
type BatchJob struct {
	id        string
	items     []string
	processor func(item string) error
	processed []string
	mutex     sync.Mutex
}

// NewBatchJob creates a job that processes a batch of items
func NewBatchJob(id string, items []string, processor func(item string) error) *BatchJob {
	return &BatchJob{
		id:        id,
		items:     items,
		processor: processor,
		processed: make([]string, 0),
	}
}

// ID returns the job ID
func (j *BatchJob) ID() string {
	return j.id
}

// Execute processes all items
func (j *BatchJob) Execute(ctx context.Context) error {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	
	for _, item := range j.items {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := j.processor(item); err != nil {
				return fmt.Errorf("failed to process item %s: %w", item, err)
			}
			j.processed = append(j.processed, item)
		}
	}
	
	return nil
}

// GetProcessedItems returns the items that have been processed
func (j *BatchJob) GetProcessedItems() []string {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	result := make([]string, len(j.processed))
	copy(result, j.processed)
	return result
}

// SlowJob provides a job that simulates slow processing
type SlowJob struct {
	*TestJob
	steps     int
	stepDelay time.Duration
	progress  int64
}

// NewSlowJob creates a job that processes in steps with delays
func NewSlowJob(id string, steps int, stepDelay time.Duration) *SlowJob {
	return &SlowJob{
		TestJob:   NewTestJob(id),
		steps:     steps,
		stepDelay: stepDelay,
	}
}

// Execute processes the job in steps
func (j *SlowJob) Execute(ctx context.Context) error {
	atomic.AddInt64(&j.executed, 1)
	
	for i := 0; i < j.steps; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(j.stepDelay):
			atomic.AddInt64(&j.progress, 1)
		}
	}
	
	if j.shouldError {
		return fmt.Errorf(j.errorMsg)
	}
	
	return nil
}

// GetProgress returns the current progress
func (j *SlowJob) GetProgress() int64 {
	return atomic.LoadInt64(&j.progress)
}

// JobGroup manages a group of jobs for testing
type JobGroup struct {
	jobs   []Job
	wg     sync.WaitGroup
	errors []error
	mutex  sync.Mutex
}

// Job interface for test jobs
type Job interface {
	ID() string
	Execute(ctx context.Context) error
}

// NewJobGroup creates a new job group
func NewJobGroup(jobs ...Job) *JobGroup {
	return &JobGroup{
		jobs:   jobs,
		errors: make([]error, 0),
	}
}

// Add adds jobs to the group
func (jg *JobGroup) Add(jobs ...Job) {
	jg.jobs = append(jg.jobs, jobs...)
}

// ExecuteAll executes all jobs concurrently
func (jg *JobGroup) ExecuteAll(ctx context.Context) error {
	jg.wg.Add(len(jg.jobs))
	
	for _, job := range jg.jobs {
		go func(j Job) {
			defer jg.wg.Done()
			if err := j.Execute(ctx); err != nil {
				jg.mutex.Lock()
				jg.errors = append(jg.errors, fmt.Errorf("job %s failed: %w", j.ID(), err))
				jg.mutex.Unlock()
			}
		}(job)
	}
	
	jg.wg.Wait()
	
	if len(jg.errors) > 0 {
		return fmt.Errorf("job group execution failed with %d errors: %v", len(jg.errors), jg.errors[0])
	}
	
	return nil
}

// GetErrors returns all errors that occurred during execution
func (jg *JobGroup) GetErrors() []error {
	jg.mutex.Lock()
	defer jg.mutex.Unlock()
	result := make([]error, len(jg.errors))
	copy(result, jg.errors)
	return result
}

// GetJobs returns all jobs in the group
func (jg *JobGroup) GetJobs() []Job {
	return jg.jobs
}

// CreateTestJobs creates a slice of simple test jobs
func CreateTestJobs(count int, duration time.Duration) []Job {
	jobs := make([]Job, count)
	for i := 0; i < count; i++ {
		jobs[i] = NewSimpleTestJob(fmt.Sprintf("job_%d", i), duration)
	}
	return jobs
}

// CreateMixedTestJobs creates a mix of successful and failing jobs
func CreateMixedTestJobs(successCount, errorCount int) []Job {
	jobs := make([]Job, 0, successCount+errorCount)
	
	// Add successful jobs
	for i := 0; i < successCount; i++ {
		jobs = append(jobs, NewSimpleTestJob(fmt.Sprintf("success_job_%d", i), 10*time.Millisecond))
	}
	
	// Add error jobs
	for i := 0; i < errorCount; i++ {
		jobs = append(jobs, NewErrorTestJob(fmt.Sprintf("error_job_%d", i), fmt.Sprintf("test error %d", i)))
	}
	
	return jobs
}