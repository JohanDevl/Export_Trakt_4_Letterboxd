package performance

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/performance/cache"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/performance/metrics"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/performance/pool"
)

// BenchmarkWorkerPool tests worker pool performance
func BenchmarkWorkerPool(b *testing.B) {
	logger := &mockLogger{}
	metrics := metrics.NewPerformanceMetrics(logger)
	
	config := pool.WorkerPoolConfig{
		Workers:    runtime.NumCPU(),
		BufferSize: 1000,
		Logger:     logger,
		Metrics:    metrics,
	}
	
	workerPool := pool.NewWorkerPool(config)
	workerPool.Start()
	defer workerPool.Stop()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			job := &testJob{id: "test"}
			workerPool.Submit(job)
		}
	})
}

// BenchmarkLRUCache tests LRU cache performance
func BenchmarkLRUCache(b *testing.B) {
	config := cache.CacheConfig{
		Capacity: 10000,
		TTL:      time.Hour,
	}
	
	lruCache := cache.NewLRUCache(config)
	
	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		lruCache.Set(string(rune(i)), i)
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := string(rune(b.N % 1000))
			lruCache.Get(key)
		}
	})
}

// BenchmarkLRUCacheSet tests LRU cache set performance
func BenchmarkLRUCacheSet(b *testing.B) {
	config := cache.CacheConfig{
		Capacity: 10000,
		TTL:      time.Hour,
	}
	
	lruCache := cache.NewLRUCache(config)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := string(rune(i))
			lruCache.Set(key, i)
			i++
		}
	})
}

// BenchmarkMetricsCollection tests metrics collection performance
func BenchmarkMetricsCollection(b *testing.B) {
	logger := &mockLogger{}
	metrics := metrics.NewPerformanceMetrics(logger)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			metrics.IncrementAPICall()
			metrics.IncrementAPISuccess()
			metrics.RecordAPIResponseTime(time.Millisecond * 100)
			metrics.IncrementItemsProcessed()
		}
	})
}

// BenchmarkConcurrentAccess tests concurrent access patterns
func BenchmarkConcurrentAccess(b *testing.B) {
	logger := &mockLogger{}
	performanceMetrics := metrics.NewPerformanceMetrics(logger)
	
	config := cache.CacheConfig{
		Capacity: 1000,
		TTL:      time.Hour,
	}
	apiCache := cache.NewAPIResponseCache(config)
	
	// Pre-populate cache
	for i := 0; i < 100; i++ {
		key := string(rune(i))
		value := map[string]interface{}{"data": i}
		apiCache.SetJSON(key, value)
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := string(rune(b.N % 100))
			var result map[string]interface{}
			
			// Simulate cache access
			if apiCache.GetJSON(key, &result) {
				performanceMetrics.IncrementCacheHit()
			} else {
				performanceMetrics.IncrementCacheMiss()
			}
			
			// Simulate processing
			performanceMetrics.IncrementItemsProcessed()
		}
	})
}

// BenchmarkMemoryAllocation tests memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	b.Run("WithPool", func(b *testing.B) {
		pool := &sync.Pool{
			New: func() interface{} {
				return make([]byte, 1024)
			},
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data := pool.Get().([]byte)
			// Simulate work
			_ = data
			pool.Put(data)
		}
	})
	
	b.Run("WithoutPool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data := make([]byte, 1024)
			// Simulate work
			_ = data
		}
	})
}

// BenchmarkGarbageCollectionImpact tests GC impact
func BenchmarkGarbageCollectionImpact(b *testing.B) {
	// Force GC before benchmark
	runtime.GC()
	
	var memStatsBefore runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)
	
	b.ResetTimer()
	
	// Allocate and process data
	for i := 0; i < b.N; i++ {
		data := make([]interface{}, 1000)
		for j := range data {
			data[j] = map[string]interface{}{
				"id":    j,
				"value": "test_value",
				"time":  time.Now(),
			}
		}
		// Data goes out of scope here
	}
	
	b.StopTimer()
	
	var memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsAfter)
	
	b.Logf("GC runs: %d", memStatsAfter.NumGC-memStatsBefore.NumGC)
	b.Logf("Heap allocations: %d bytes", memStatsAfter.TotalAlloc-memStatsBefore.TotalAlloc)
}

// TestJob implements the Job interface for testing
type testJob struct {
	id string
}

func (job *testJob) Execute(ctx context.Context) error {
	// Simulate work
	time.Sleep(time.Microsecond)
	return nil
}

func (job *testJob) ID() string {
	return job.id
}

// MockLogger implements the Logger interface for testing
type mockLogger struct{}

func (m *mockLogger) Debug(key string, data ...map[string]interface{}) {}
func (m *mockLogger) Debugf(key string, data map[string]interface{})   {}
func (m *mockLogger) Info(key string, data ...map[string]interface{})  {}
func (m *mockLogger) Infof(key string, data map[string]interface{})    {}
func (m *mockLogger) Warn(key string, data ...map[string]interface{})  {}
func (m *mockLogger) Warnf(key string, data map[string]interface{})    {}
func (m *mockLogger) Error(key string, data ...map[string]interface{}) {}
func (m *mockLogger) Errorf(key string, data map[string]interface{})   {}
func (m *mockLogger) SetLogLevel(level string)                         {}
func (m *mockLogger) SetLogFile(path string) error                     { return nil }
func (m *mockLogger) SetTranslator(t logger.Translator)               {}

// Performance test helpers

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
	
	config := pool.WorkerPoolConfig{
		Workers:    runtime.NumCPU(),
		BufferSize: 1000,
		Logger:     logger,
		Metrics:    performanceMetrics,
	}
	
	workerPool := pool.NewWorkerPool(config)
	workerPool.Start()
	defer workerPool.Stop()
	
	totalJobs := 10000
	start := time.Now()
	
	// Submit jobs
	for i := 0; i < totalJobs; i++ {
		job := &testJob{id: string(rune(i))}
		if err := workerPool.Submit(job); err != nil {
			t.Fatalf("Failed to submit job: %v", err)
		}
	}
	
	// Wait for all jobs to complete
	for {
		stats := workerPool.Stats()
		if stats.ProcessedJobs >= int64(totalJobs) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	
	duration := time.Since(start)
	throughput := float64(totalJobs) / duration.Seconds()
	
	t.Logf("Processed %d jobs in %v", totalJobs, duration)
	t.Logf("Throughput: %.2f jobs/second", throughput)
	
	// Expect at least 1000 jobs per second
	if throughput < 1000 {
		t.Errorf("Throughput too low: %.2f jobs/second, expected at least 1000", throughput)
	}
} 