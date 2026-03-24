package performance

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/performance/cache"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/performance/metrics"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/performance/pool"
)

// TestCacheHitRatio tests cache hit ratio under load
func TestCacheHitRatio(t *testing.T) {
	config := cache.CacheConfig{
		Capacity: 100,
		TTL:      time.Hour,
	}

	apiCache := cache.NewAPIResponseCache(config)

	// Pre-populate cache with 50 items
	for i := 0; i < 50; i++ {
		key := string(rune(i))
		value := map[string]interface{}{"data": i}
		apiCache.SetJSON(key, value)
	}

	hits := 0
	misses := 0
	iterations := 1000

	// Test with 80% cache hit ratio (access existing items 80% of the time)
	for i := 0; i < iterations; i++ {
		var result map[string]interface{}
		var key string

		if i%10 < 8 {
			// 80% chance to access existing item
			key = string(rune(i % 50))
		} else {
			// 20% chance to access new item
			key = string(rune(50 + i))
		}

		if apiCache.GetJSON(key, &result) {
			hits++
		} else {
			misses++
			// Cache miss - add the item
			value := map[string]interface{}{"data": i}
			apiCache.SetJSON(key, value)
		}
	}

	hitRatio := float64(hits) / float64(hits+misses) * 100
	t.Logf("Cache hit ratio: %.2f%% (hits: %d, misses: %d)", hitRatio, hits, misses)

	// Expect at least 60% hit ratio
	if hitRatio < 60 {
		t.Errorf("Cache hit ratio too low: %.2f%%, expected at least 60%%", hitRatio)
	}
}

// TestMemoryUsage tests memory usage patterns
func TestMemoryUsage(t *testing.T) {
	var m1, m2 runtime.MemStats

	// Get initial memory stats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Allocate data
	logger := &mockLogger{}
	metrics := metrics.NewPerformanceMetrics(logger)

	config := cache.CacheConfig{
		Capacity: 1000,
		TTL:      time.Hour,
	}
	apiCache := cache.NewAPIResponseCache(config)

	// Populate cache
	for i := 0; i < 1000; i++ {
		key := string(rune(i))
		value := map[string]interface{}{
			"id":    i,
			"data":  "test_data",
			"time":  time.Now(),
		}
		apiCache.SetJSON(key, value)
		metrics.IncrementCacheHit()
	}

	// Get final memory stats
	runtime.GC()
	runtime.ReadMemStats(&m2)

	allocatedMB := float64(m2.HeapAlloc-m1.HeapAlloc) / 1024 / 1024
	t.Logf("Memory allocated: %.2f MB", allocatedMB)
	t.Logf("Total allocations: %d", m2.TotalAlloc-m1.TotalAlloc)
	t.Logf("GC runs: %d", m2.NumGC-m1.NumGC)

	// Check that memory usage is reasonable (should be less than 50MB for this test)
	if allocatedMB > 50 {
		t.Errorf("Memory usage too high: %.2f MB, expected less than 50MB", allocatedMB)
	}
}

// TestWorkerPoolThroughput tests worker pool throughput
func TestWorkerPoolThroughput(t *testing.T) {
	logger := &mockLogger{}
	performanceMetrics := metrics.NewPerformanceMetrics(logger)

	totalJobs := 1000
	config := pool.WorkerPoolConfig{
		Workers:    runtime.NumCPU(),
		BufferSize: totalJobs,
		Logger:     logger,
		Metrics:    performanceMetrics,
	}

	workerPool := pool.NewWorkerPool(config)
	workerPool.Start()
	defer workerPool.Stop()

	// Start a goroutine to consume results to prevent blocking
	go func() {
		for range workerPool.Results() {
		}
	}()

	start := time.Now()
	jobsSubmitted := 0

	for i := 0; i < totalJobs; i++ {
		job := &testJob{id: fmt.Sprintf("job-%d", i)}
		if err := workerPool.Submit(job); err == nil {
			jobsSubmitted++
		}
	}

	// Wait for all submitted jobs to complete
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatalf("Timeout waiting for jobs to complete")
		case <-ticker.C:
			stats := workerPool.Stats()
			if stats.ProcessedJobs >= int64(jobsSubmitted) {
				goto completed
			}
		}
	}

completed:
	duration := time.Since(start)
	throughput := float64(jobsSubmitted) / duration.Seconds()

	t.Logf("Submitted %d/%d jobs, processed in %v (%.0f jobs/sec)", jobsSubmitted, totalJobs, duration, throughput)

	if float64(jobsSubmitted)/float64(totalJobs) < 0.90 {
		t.Errorf("Too few jobs submitted: %d/%d (%.1f%%)", jobsSubmitted, totalJobs, float64(jobsSubmitted)/float64(totalJobs)*100)
	}

	if throughput < 100 {
		t.Errorf("Throughput too low: %.2f jobs/second", throughput)
	}
}
