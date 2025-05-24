package metrics

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

// PerformanceMetrics tracks application performance metrics
type PerformanceMetrics struct {
	mu sync.RWMutex
	
	// API metrics
	apiCallsTotal    int64
	apiCallsSuccess  int64
	apiCallsError    int64
	apiResponseTimes []time.Duration
	
	// Processing metrics
	itemsProcessed   int64
	itemsError       int64
	processingTimes  []time.Duration
	
	// Cache metrics
	cacheHits        int64
	cacheMisses      int64
	
	// Job metrics
	jobsProcessed    int64
	jobsErrored      int64
	jobDurations     []time.Duration
	
	// Memory metrics
	maxMemoryUsage   uint64
	currentMemory    uint64
	
	// Timing
	startTime        time.Time
	logger           logger.Logger
}

// NewPerformanceMetrics creates a new performance metrics collector
func NewPerformanceMetrics(logger logger.Logger) *PerformanceMetrics {
	return &PerformanceMetrics{
		startTime:        time.Now(),
		logger:           logger,
		apiResponseTimes: make([]time.Duration, 0, 1000),
		processingTimes:  make([]time.Duration, 0, 1000),
		jobDurations:     make([]time.Duration, 0, 1000),
	}
}

// API Metrics

// IncrementAPICall increments the total API calls counter
func (pm *PerformanceMetrics) IncrementAPICall() {
	atomic.AddInt64(&pm.apiCallsTotal, 1)
}

// IncrementAPISuccess increments successful API calls
func (pm *PerformanceMetrics) IncrementAPISuccess() {
	atomic.AddInt64(&pm.apiCallsSuccess, 1)
}

// IncrementAPIError increments API error calls
func (pm *PerformanceMetrics) IncrementAPIError() {
	atomic.AddInt64(&pm.apiCallsError, 1)
}

// RecordAPIResponseTime records an API response time
func (pm *PerformanceMetrics) RecordAPIResponseTime(duration time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	// Keep only the last 1000 response times to prevent memory growth
	if len(pm.apiResponseTimes) >= 1000 {
		pm.apiResponseTimes = pm.apiResponseTimes[1:]
	}
	pm.apiResponseTimes = append(pm.apiResponseTimes, duration)
}

// Processing Metrics

// IncrementItemsProcessed increments processed items counter
func (pm *PerformanceMetrics) IncrementItemsProcessed() {
	atomic.AddInt64(&pm.itemsProcessed, 1)
}

// IncrementItemsError increments processing error counter
func (pm *PerformanceMetrics) IncrementItemsError() {
	atomic.AddInt64(&pm.itemsError, 1)
}

// RecordProcessingTime records processing time
func (pm *PerformanceMetrics) RecordProcessingTime(duration time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	if len(pm.processingTimes) >= 1000 {
		pm.processingTimes = pm.processingTimes[1:]
	}
	pm.processingTimes = append(pm.processingTimes, duration)
}

// Cache Metrics

// IncrementCacheHit increments cache hit counter
func (pm *PerformanceMetrics) IncrementCacheHit() {
	atomic.AddInt64(&pm.cacheHits, 1)
}

// IncrementCacheMiss increments cache miss counter
func (pm *PerformanceMetrics) IncrementCacheMiss() {
	atomic.AddInt64(&pm.cacheMisses, 1)
}

// Job Metrics

// IncrementJobsProcessed increments jobs processed counter
func (pm *PerformanceMetrics) IncrementJobsProcessed() {
	atomic.AddInt64(&pm.jobsProcessed, 1)
}

// IncrementJobsErrored increments jobs error counter
func (pm *PerformanceMetrics) IncrementJobsErrored() {
	atomic.AddInt64(&pm.jobsErrored, 1)
}

// RecordJobDuration records job execution duration
func (pm *PerformanceMetrics) RecordJobDuration(duration time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	if len(pm.jobDurations) >= 1000 {
		pm.jobDurations = pm.jobDurations[1:]
	}
	pm.jobDurations = append(pm.jobDurations, duration)
}

// Memory Metrics

// UpdateMemoryUsage updates current memory usage
func (pm *PerformanceMetrics) UpdateMemoryUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	current := m.HeapAlloc
	atomic.StoreUint64(&pm.currentMemory, current)
	
	// Update max if current is higher
	for {
		max := atomic.LoadUint64(&pm.maxMemoryUsage)
		if current <= max || atomic.CompareAndSwapUint64(&pm.maxMemoryUsage, max, current) {
			break
		}
	}
}

// Getters for Statistics

// GetAPIStats returns API statistics
func (pm *PerformanceMetrics) GetAPIStats() APIStats {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	total := atomic.LoadInt64(&pm.apiCallsTotal)
	success := atomic.LoadInt64(&pm.apiCallsSuccess)
	errors := atomic.LoadInt64(&pm.apiCallsError)
	
	var avgResponseTime time.Duration
	if len(pm.apiResponseTimes) > 0 {
		var sum time.Duration
		for _, duration := range pm.apiResponseTimes {
			sum += duration
		}
		avgResponseTime = sum / time.Duration(len(pm.apiResponseTimes))
	}
	
	var successRate float64
	if total > 0 {
		successRate = float64(success) / float64(total) * 100
	}
	
	return APIStats{
		TotalCalls:       total,
		SuccessfulCalls:  success,
		ErrorCalls:       errors,
		SuccessRate:      successRate,
		AvgResponseTime:  avgResponseTime,
	}
}

// GetProcessingStats returns processing statistics
func (pm *PerformanceMetrics) GetProcessingStats() ProcessingStats {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	processed := atomic.LoadInt64(&pm.itemsProcessed)
	errors := atomic.LoadInt64(&pm.itemsError)
	
	var avgProcessingTime time.Duration
	if len(pm.processingTimes) > 0 {
		var sum time.Duration
		for _, duration := range pm.processingTimes {
			sum += duration
		}
		avgProcessingTime = sum / time.Duration(len(pm.processingTimes))
	}
	
	var successRate float64
	total := processed + errors
	if total > 0 {
		successRate = float64(processed) / float64(total) * 100
	}
	
	var throughput float64
	uptime := time.Since(pm.startTime)
	if uptime > 0 {
		throughput = float64(processed) / uptime.Seconds()
	}
	
	return ProcessingStats{
		ItemsProcessed:     processed,
		ItemsError:         errors,
		SuccessRate:        successRate,
		AvgProcessingTime:  avgProcessingTime,
		Throughput:         throughput,
	}
}

// GetCacheStats returns cache statistics
func (pm *PerformanceMetrics) GetCacheStats() CacheStats {
	hits := atomic.LoadInt64(&pm.cacheHits)
	misses := atomic.LoadInt64(&pm.cacheMisses)
	
	var hitRatio float64
	total := hits + misses
	if total > 0 {
		hitRatio = float64(hits) / float64(total) * 100
	}
	
	return CacheStats{
		Hits:     hits,
		Misses:   misses,
		HitRatio: hitRatio,
	}
}

// GetJobStats returns job statistics
func (pm *PerformanceMetrics) GetJobStats() JobStats {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	processed := atomic.LoadInt64(&pm.jobsProcessed)
	errors := atomic.LoadInt64(&pm.jobsErrored)
	
	var avgDuration time.Duration
	if len(pm.jobDurations) > 0 {
		var sum time.Duration
		for _, duration := range pm.jobDurations {
			sum += duration
		}
		avgDuration = sum / time.Duration(len(pm.jobDurations))
	}
	
	var successRate float64
	total := processed + errors
	if total > 0 {
		successRate = float64(processed) / float64(total) * 100
	}
	
	return JobStats{
		JobsProcessed: processed,
		JobsErrored:   errors,
		SuccessRate:   successRate,
		AvgDuration:   avgDuration,
	}
}

// GetMemoryStats returns memory statistics
func (pm *PerformanceMetrics) GetMemoryStats() MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return MemoryStats{
		CurrentAlloc:   atomic.LoadUint64(&pm.currentMemory),
		MaxAlloc:       atomic.LoadUint64(&pm.maxMemoryUsage),
		HeapAlloc:      m.HeapAlloc,
		HeapSys:        m.HeapSys,
		HeapObjects:    m.HeapObjects,
		StackInuse:     m.StackInuse,
		GCRuns:         m.NumGC,
		LastGC:         time.Unix(0, int64(m.LastGC)),
		NextGC:         m.NextGC,
		Goroutines:     runtime.NumGoroutine(),
	}
}

// GetOverallStats returns overall performance statistics
func (pm *PerformanceMetrics) GetOverallStats() OverallStats {
	uptime := time.Since(pm.startTime)
	
	return OverallStats{
		Uptime:          uptime,
		StartTime:       pm.startTime,
		API:             pm.GetAPIStats(),
		Processing:      pm.GetProcessingStats(),
		Cache:           pm.GetCacheStats(),
		Jobs:            pm.GetJobStats(),
		Memory:          pm.GetMemoryStats(),
	}
}

// Reset resets all metrics
func (pm *PerformanceMetrics) Reset() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	// Reset counters
	atomic.StoreInt64(&pm.apiCallsTotal, 0)
	atomic.StoreInt64(&pm.apiCallsSuccess, 0)
	atomic.StoreInt64(&pm.apiCallsError, 0)
	atomic.StoreInt64(&pm.itemsProcessed, 0)
	atomic.StoreInt64(&pm.itemsError, 0)
	atomic.StoreInt64(&pm.cacheHits, 0)
	atomic.StoreInt64(&pm.cacheMisses, 0)
	atomic.StoreInt64(&pm.jobsProcessed, 0)
	atomic.StoreInt64(&pm.jobsErrored, 0)
	atomic.StoreUint64(&pm.maxMemoryUsage, 0)
	atomic.StoreUint64(&pm.currentMemory, 0)
	
	// Reset slices
	pm.apiResponseTimes = pm.apiResponseTimes[:0]
	pm.processingTimes = pm.processingTimes[:0]
	pm.jobDurations = pm.jobDurations[:0]
	
	// Reset start time
	pm.startTime = time.Now()
	
	pm.logger.Info("performance_metrics.reset", nil)
}

// LogStats logs current statistics
func (pm *PerformanceMetrics) LogStats() {
	stats := pm.GetOverallStats()
	
	pm.logger.Info("performance_metrics.stats", map[string]interface{}{
		"uptime":               stats.Uptime.String(),
		"api_calls_total":      stats.API.TotalCalls,
		"api_success_rate":     stats.API.SuccessRate,
		"api_avg_response":     stats.API.AvgResponseTime.String(),
		"items_processed":      stats.Processing.ItemsProcessed,
		"processing_throughput": stats.Processing.Throughput,
		"cache_hit_ratio":      stats.Cache.HitRatio,
		"memory_current_mb":    stats.Memory.CurrentAlloc / 1024 / 1024,
		"memory_max_mb":        stats.Memory.MaxAlloc / 1024 / 1024,
		"goroutines":           stats.Memory.Goroutines,
		"gc_runs":              stats.Memory.GCRuns,
	})
}

// Stats Types

// APIStats represents API performance statistics
type APIStats struct {
	TotalCalls      int64         `json:"total_calls"`
	SuccessfulCalls int64         `json:"successful_calls"`
	ErrorCalls      int64         `json:"error_calls"`
	SuccessRate     float64       `json:"success_rate"`
	AvgResponseTime time.Duration `json:"avg_response_time"`
}

// ProcessingStats represents processing performance statistics
type ProcessingStats struct {
	ItemsProcessed    int64         `json:"items_processed"`
	ItemsError        int64         `json:"items_error"`
	SuccessRate       float64       `json:"success_rate"`
	AvgProcessingTime time.Duration `json:"avg_processing_time"`
	Throughput        float64       `json:"throughput"` // items per second
}

// CacheStats represents cache performance statistics
type CacheStats struct {
	Hits     int64   `json:"hits"`
	Misses   int64   `json:"misses"`
	HitRatio float64 `json:"hit_ratio"`
}

// JobStats represents job execution statistics
type JobStats struct {
	JobsProcessed int64         `json:"jobs_processed"`
	JobsErrored   int64         `json:"jobs_errored"`
	SuccessRate   float64       `json:"success_rate"`
	AvgDuration   time.Duration `json:"avg_duration"`
}

// MemoryStats represents memory usage statistics
type MemoryStats struct {
	CurrentAlloc uint64    `json:"current_alloc"`
	MaxAlloc     uint64    `json:"max_alloc"`
	HeapAlloc    uint64    `json:"heap_alloc"`
	HeapSys      uint64    `json:"heap_sys"`
	HeapObjects  uint64    `json:"heap_objects"`
	StackInuse   uint64    `json:"stack_inuse"`
	GCRuns       uint32    `json:"gc_runs"`
	LastGC       time.Time `json:"last_gc"`
	NextGC       uint64    `json:"next_gc"`
	Goroutines   int       `json:"goroutines"`
}

// OverallStats represents overall application statistics
type OverallStats struct {
	Uptime     time.Duration   `json:"uptime"`
	StartTime  time.Time       `json:"start_time"`
	API        APIStats        `json:"api"`
	Processing ProcessingStats `json:"processing"`
	Cache      CacheStats      `json:"cache"`
	Jobs       JobStats        `json:"jobs"`
	Memory     MemoryStats     `json:"memory"`
} 