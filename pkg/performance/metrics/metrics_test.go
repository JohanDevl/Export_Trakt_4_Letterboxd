package metrics

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

// MockLogger implements the logger.Logger interface for testing
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(messageID string, data ...map[string]interface{}) {
	args := m.Called(messageID, data)
	_ = args
}

func (m *MockLogger) Infof(messageID string, data map[string]interface{}) {
	m.Called(messageID, data)
}

func (m *MockLogger) Error(messageID string, data ...map[string]interface{}) {
	m.Called(messageID, data)
}

func (m *MockLogger) Errorf(messageID string, data map[string]interface{}) {
	m.Called(messageID, data)
}

func (m *MockLogger) Warn(messageID string, data ...map[string]interface{}) {
	m.Called(messageID, data)
}

func (m *MockLogger) Warnf(messageID string, data map[string]interface{}) {
	m.Called(messageID, data)
}

func (m *MockLogger) Debug(messageID string, data ...map[string]interface{}) {
	m.Called(messageID, data)
}

func (m *MockLogger) Debugf(messageID string, data map[string]interface{}) {
	m.Called(messageID, data)
}

func (m *MockLogger) SetLogLevel(level string) {
	m.Called(level)
}

func (m *MockLogger) SetLogFile(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *MockLogger) SetTranslator(t logger.Translator) {
	m.Called(t)
}

// Test RingBuffer functionality
func TestNewRingBuffer(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
	}{
		{"small capacity", 10},
		{"medium capacity", 100},
		{"large capacity", 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb := NewRingBuffer(tt.capacity)
			assert.NotNil(t, rb)
			assert.Equal(t, tt.capacity, rb.Capacity())
			assert.Equal(t, 0, rb.Size())
			assert.Nil(t, rb.Values())
		})
	}
}

func TestRingBufferAdd(t *testing.T) {
	rb := NewRingBuffer(3)
	
	// Add first value
	rb.Add(100 * time.Millisecond)
	assert.Equal(t, 1, rb.Size())
	assert.Equal(t, []time.Duration{100 * time.Millisecond}, rb.Values())
	
	// Add second value
	rb.Add(200 * time.Millisecond)
	assert.Equal(t, 2, rb.Size())
	assert.Equal(t, []time.Duration{100 * time.Millisecond, 200 * time.Millisecond}, rb.Values())
	
	// Add third value (buffer full)
	rb.Add(300 * time.Millisecond)
	assert.Equal(t, 3, rb.Size())
	assert.Equal(t, []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 300 * time.Millisecond}, rb.Values())
	
	// Add fourth value (should overwrite first)
	rb.Add(400 * time.Millisecond)
	assert.Equal(t, 3, rb.Size())
	assert.Equal(t, []time.Duration{200 * time.Millisecond, 300 * time.Millisecond, 400 * time.Millisecond}, rb.Values())
}

func TestRingBufferClear(t *testing.T) {
	rb := NewRingBuffer(5)
	
	// Add some values
	for i := 0; i < 3; i++ {
		rb.Add(time.Duration(i) * time.Millisecond)
	}
	assert.Equal(t, 3, rb.Size())
	
	// Clear buffer
	rb.Clear()
	assert.Equal(t, 0, rb.Size())
	assert.Nil(t, rb.Values())
}

func TestRingBufferLatest(t *testing.T) {
	rb := NewRingBuffer(3)
	
	// Empty buffer
	assert.Equal(t, time.Duration(0), rb.Latest())
	
	// Add values and check latest
	rb.Add(100 * time.Millisecond)
	assert.Equal(t, 100*time.Millisecond, rb.Latest())
	
	rb.Add(200 * time.Millisecond)
	assert.Equal(t, 200*time.Millisecond, rb.Latest())
	
	rb.Add(300 * time.Millisecond)
	assert.Equal(t, 300*time.Millisecond, rb.Latest())
	
	// Wrap around
	rb.Add(400 * time.Millisecond)
	assert.Equal(t, 400*time.Millisecond, rb.Latest())
}

func TestRingBufferAverage(t *testing.T) {
	rb := NewRingBuffer(3)
	
	// Empty buffer
	assert.Equal(t, time.Duration(0), rb.Average())
	
	// Add values
	rb.Add(100 * time.Millisecond)
	assert.Equal(t, 100*time.Millisecond, rb.Average())
	
	rb.Add(200 * time.Millisecond)
	assert.Equal(t, 150*time.Millisecond, rb.Average())
	
	rb.Add(300 * time.Millisecond)
	assert.Equal(t, 200*time.Millisecond, rb.Average())
}

func TestRingBufferMinMax(t *testing.T) {
	rb := NewRingBuffer(5)
	
	// Empty buffer
	assert.Equal(t, time.Duration(0), rb.Min())
	assert.Equal(t, time.Duration(0), rb.Max())
	
	// Add values
	values := []time.Duration{300 * time.Millisecond, 100 * time.Millisecond, 500 * time.Millisecond, 200 * time.Millisecond}
	for _, v := range values {
		rb.Add(v)
	}
	
	assert.Equal(t, 100*time.Millisecond, rb.Min())
	assert.Equal(t, 500*time.Millisecond, rb.Max())
}

func TestRingBufferPercentile(t *testing.T) {
	rb := NewRingBuffer(5)
	
	// Empty buffer
	assert.Equal(t, time.Duration(0), rb.Percentile(50))
	
	// Add values
	values := []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 300 * time.Millisecond}
	for _, v := range values {
		rb.Add(v)
	}
	
	// Test edge cases
	assert.Equal(t, rb.Min(), rb.Percentile(0))
	assert.Equal(t, rb.Max(), rb.Percentile(100))
	assert.Equal(t, time.Duration(0), rb.Percentile(-10)) // Invalid percentile
	assert.Equal(t, time.Duration(0), rb.Percentile(110)) // Invalid percentile
}

// Test PerformanceMetrics functionality
func TestNewPerformanceMetrics(t *testing.T) {
	mockLogger := &MockLogger{}
	
	pm := NewPerformanceMetrics(mockLogger)
	
	assert.NotNil(t, pm)
	assert.Equal(t, mockLogger, pm.logger)
	assert.NotNil(t, pm.apiResponseTimes)
	assert.NotNil(t, pm.processingTimes)
	assert.NotNil(t, pm.jobDurations)
	assert.Equal(t, 1000, pm.apiResponseTimes.Capacity())
	assert.False(t, pm.startTime.IsZero())
}

func TestPerformanceMetricsAPIMetrics(t *testing.T) {
	mockLogger := &MockLogger{}
	pm := NewPerformanceMetrics(mockLogger)
	
	// Initial state
	stats := pm.GetAPIStats()
	assert.Equal(t, int64(0), stats.TotalCalls)
	assert.Equal(t, int64(0), stats.SuccessfulCalls)
	assert.Equal(t, int64(0), stats.ErrorCalls)
	assert.Equal(t, float64(0), stats.SuccessRate)
	
	// Increment counters
	pm.IncrementAPICall()
	pm.IncrementAPISuccess()
	pm.RecordAPIResponseTime(100 * time.Millisecond)
	
	pm.IncrementAPICall()
	pm.IncrementAPIError()
	pm.RecordAPIResponseTime(200 * time.Millisecond)
	
	stats = pm.GetAPIStats()
	assert.Equal(t, int64(2), stats.TotalCalls)
	assert.Equal(t, int64(1), stats.SuccessfulCalls)
	assert.Equal(t, int64(1), stats.ErrorCalls)
	assert.Equal(t, float64(50), stats.SuccessRate)
	assert.Equal(t, 150*time.Millisecond, stats.AvgResponseTime)
}

func TestPerformanceMetricsProcessingMetrics(t *testing.T) {
	mockLogger := &MockLogger{}
	pm := NewPerformanceMetrics(mockLogger)
	
	// Process some items
	pm.IncrementItemsProcessed()
	pm.IncrementItemsProcessed()
	pm.IncrementItemsError()
	pm.RecordProcessingTime(50 * time.Millisecond)
	pm.RecordProcessingTime(150 * time.Millisecond)
	
	// Small delay to ensure measurable uptime
	time.Sleep(time.Millisecond)
	
	stats := pm.GetProcessingStats()
	assert.Equal(t, int64(2), stats.ItemsProcessed)
	assert.Equal(t, int64(1), stats.ItemsError)
	assert.InDelta(t, 66.67, stats.SuccessRate, 0.01)
	assert.Equal(t, 100*time.Millisecond, stats.AvgProcessingTime)
	assert.Greater(t, stats.Throughput, float64(0))
}

func TestPerformanceMetricsCacheMetrics(t *testing.T) {
	mockLogger := &MockLogger{}
	pm := NewPerformanceMetrics(mockLogger)
	
	// Record cache operations
	pm.IncrementCacheHit()
	pm.IncrementCacheHit()
	pm.IncrementCacheHit()
	pm.IncrementCacheMiss()
	
	stats := pm.GetCacheStats()
	assert.Equal(t, int64(3), stats.Hits)
	assert.Equal(t, int64(1), stats.Misses)
	assert.Equal(t, float64(75), stats.HitRatio)
}

func TestPerformanceMetricsJobMetrics(t *testing.T) {
	mockLogger := &MockLogger{}
	pm := NewPerformanceMetrics(mockLogger)
	
	// Record job operations
	pm.IncrementJobsProcessed()
	pm.IncrementJobsProcessed()
	pm.IncrementJobsErrored()
	pm.RecordJobDuration(100 * time.Millisecond)
	pm.RecordJobDuration(200 * time.Millisecond)
	
	stats := pm.GetJobStats()
	assert.Equal(t, int64(2), stats.JobsProcessed)
	assert.Equal(t, int64(1), stats.JobsErrored)
	assert.InDelta(t, 66.67, stats.SuccessRate, 0.01)
	assert.Equal(t, 150*time.Millisecond, stats.AvgDuration)
}

func TestPerformanceMetricsMemoryMetrics(t *testing.T) {
	mockLogger := &MockLogger{}
	pm := NewPerformanceMetrics(mockLogger)
	
	// Update memory usage
	pm.UpdateMemoryUsage()
	
	stats := pm.GetMemoryStats()
	assert.Greater(t, stats.CurrentAlloc, uint64(0))
	assert.Greater(t, stats.HeapAlloc, uint64(0))
	assert.Greater(t, stats.Goroutines, 0)
	
	// Test that max memory is updated
	initialMax := atomic.LoadUint64(&pm.maxMemoryUsage)
	pm.UpdateMemoryUsage()
	updatedMax := atomic.LoadUint64(&pm.maxMemoryUsage)
	assert.GreaterOrEqual(t, updatedMax, initialMax)
}

func TestPerformanceMetricsOverallStats(t *testing.T) {
	mockLogger := &MockLogger{}
	pm := NewPerformanceMetrics(mockLogger)
	
	// Add some metrics
	pm.IncrementAPICall()
	pm.IncrementItemsProcessed()
	pm.IncrementCacheHit()
	pm.IncrementJobsProcessed()
	pm.UpdateMemoryUsage()
	
	stats := pm.GetOverallStats()
	assert.Greater(t, stats.Uptime, time.Duration(0))
	assert.False(t, stats.StartTime.IsZero())
	assert.Equal(t, int64(1), stats.API.TotalCalls)
	assert.Equal(t, int64(1), stats.Processing.ItemsProcessed)
	assert.Equal(t, int64(1), stats.Cache.Hits)
	assert.Equal(t, int64(1), stats.Jobs.JobsProcessed)
	assert.Greater(t, stats.Memory.CurrentAlloc, uint64(0))
}

func TestPerformanceMetricsReset(t *testing.T) {
	mockLogger := &MockLogger{}
	mockLogger.On("Info", "performance_metrics.reset", mock.Anything).Return()
	
	pm := NewPerformanceMetrics(mockLogger)
	
	// Add some metrics
	pm.IncrementAPICall()
	pm.IncrementItemsProcessed()
	pm.IncrementCacheHit()
	pm.IncrementJobsProcessed()
	pm.RecordAPIResponseTime(100 * time.Millisecond)
	
	// Verify metrics are set
	assert.Equal(t, int64(1), atomic.LoadInt64(&pm.apiCallsTotal))
	assert.Equal(t, 1, pm.apiResponseTimes.Size())
	
	// Reset metrics
	pm.Reset()
	
	// Verify all metrics are reset
	assert.Equal(t, int64(0), atomic.LoadInt64(&pm.apiCallsTotal))
	assert.Equal(t, int64(0), atomic.LoadInt64(&pm.apiCallsSuccess))
	assert.Equal(t, int64(0), atomic.LoadInt64(&pm.itemsProcessed))
	assert.Equal(t, int64(0), atomic.LoadInt64(&pm.cacheHits))
	assert.Equal(t, int64(0), atomic.LoadInt64(&pm.jobsProcessed))
	assert.Equal(t, uint64(0), atomic.LoadUint64(&pm.maxMemoryUsage))
	assert.Equal(t, 0, pm.apiResponseTimes.Size())
	
	mockLogger.AssertCalled(t, "Info", "performance_metrics.reset", mock.Anything)
}

func TestPerformanceMetricsLogStats(t *testing.T) {
	mockLogger := &MockLogger{}
	mockLogger.On("Info", "performance_metrics.stats", mock.Anything).Return()
	
	pm := NewPerformanceMetrics(mockLogger)
	
	// Add some metrics
	pm.IncrementAPICall()
	pm.IncrementAPISuccess()
	pm.RecordAPIResponseTime(100 * time.Millisecond)
	
	pm.LogStats()
	
	mockLogger.AssertCalled(t, "Info", "performance_metrics.stats", mock.Anything)
}

func TestPerformanceMetricsConcurrency(t *testing.T) {
	mockLogger := &MockLogger{}
	pm := NewPerformanceMetrics(mockLogger)
	
	// Run concurrent operations
	var wg sync.WaitGroup
	numGoroutines := 100
	operationsPerGoroutine := 100
	
	// API operations
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				pm.IncrementAPICall()
				pm.IncrementAPISuccess()
				pm.RecordAPIResponseTime(time.Duration(j) * time.Millisecond)
			}
		}()
	}
	
	// Cache operations
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				pm.IncrementCacheHit()
				pm.IncrementCacheMiss()
			}
		}()
	}
	
	wg.Wait()
	
	// Verify final counts
	expectedAPICalls := int64(numGoroutines * operationsPerGoroutine)
	expectedCacheHits := int64(numGoroutines * operationsPerGoroutine)
	expectedCacheMisses := int64(numGoroutines * operationsPerGoroutine)
	
	stats := pm.GetOverallStats()
	assert.Equal(t, expectedAPICalls, stats.API.TotalCalls)
	assert.Equal(t, expectedAPICalls, stats.API.SuccessfulCalls)
	assert.Equal(t, expectedCacheHits, stats.Cache.Hits)
	assert.Equal(t, expectedCacheMisses, stats.Cache.Misses)
}

func TestPerformanceMetricsEdgeCases(t *testing.T) {
	mockLogger := &MockLogger{}
	pm := NewPerformanceMetrics(mockLogger)
	
	// Test with no operations
	apiStats := pm.GetAPIStats()
	assert.Equal(t, float64(0), apiStats.SuccessRate)
	assert.Equal(t, time.Duration(0), apiStats.AvgResponseTime)
	
	processingStats := pm.GetProcessingStats()
	assert.Equal(t, float64(0), processingStats.SuccessRate)
	assert.Equal(t, time.Duration(0), processingStats.AvgProcessingTime)
	
	cacheStats := pm.GetCacheStats()
	assert.Equal(t, float64(0), cacheStats.HitRatio)
	
	jobStats := pm.GetJobStats()
	assert.Equal(t, float64(0), jobStats.SuccessRate)
	assert.Equal(t, time.Duration(0), jobStats.AvgDuration)
}

// Benchmark tests
func BenchmarkPerformanceMetricsIncrements(b *testing.B) {
	mockLogger := &MockLogger{}
	pm := NewPerformanceMetrics(mockLogger)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pm.IncrementAPICall()
		}
	})
}

func BenchmarkRingBufferAdd(b *testing.B) {
	rb := NewRingBuffer(1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Add(time.Duration(i) * time.Nanosecond)
	}
}

func BenchmarkRingBufferAverage(b *testing.B) {
	rb := NewRingBuffer(1000)
	
	// Fill buffer
	for i := 0; i < 1000; i++ {
		rb.Add(time.Duration(i) * time.Nanosecond)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Average()
	}
}