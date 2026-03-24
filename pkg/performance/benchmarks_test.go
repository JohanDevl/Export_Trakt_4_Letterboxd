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
				data := make([]byte, 1024)
				return &data
			},
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data := pool.Get().(*[]byte)
			// Simulate work
			_ = *data
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
	// Simulate very light work - just a simple calculation
	_ = len(job.id) * 42
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

 