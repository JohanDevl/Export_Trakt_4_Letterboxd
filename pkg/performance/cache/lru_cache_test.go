package cache

import (
	"fmt"
	"testing"
	"time"
)

func TestNewLRUCache(t *testing.T) {
	config := CacheConfig{
		Capacity: 10,
		TTL:      5 * time.Minute,
	}
	cache := NewLRUCache(config)
	if cache == nil {
		t.Fatal("Expected cache to be created, got nil")
	}
	
	if cache.Size() != 0 {
		t.Errorf("Expected empty cache size 0, got %d", cache.Size())
	}
}

func TestCacheSetGet(t *testing.T) {
	config := CacheConfig{
		Capacity: 3,
		TTL:      time.Hour,
	}
	cache := NewLRUCache(config)
	
	// Test basic set/get
	cache.Set("key1", "value1")
	
	value, found := cache.Get("key1")
	if !found {
		t.Error("Expected to find key1")
	}
	if value != "value1" {
		t.Errorf("Expected value1, got %v", value)
	}
	
	// Test non-existent key
	_, found = cache.Get("nonexistent")
	if found {
		t.Error("Expected not to find nonexistent key")
	}
}

func TestCacheDelete(t *testing.T) {
	config := CacheConfig{
		Capacity: 10,
		TTL:      time.Hour,
	}
	cache := NewLRUCache(config)
	
	cache.Set("key1", "value1")
	cache.Delete("key1")
	
	_, found := cache.Get("key1")
	if found {
		t.Error("Expected key1 to be deleted")
	}
}

func TestCacheClear(t *testing.T) {
	config := CacheConfig{
		Capacity: 10,
		TTL:      time.Hour,
	}
	cache := NewLRUCache(config)
	
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	
	cache.Clear()
	
	if cache.Size() != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", cache.Size())
	}
}

func TestNewAPIResponseCache(t *testing.T) {
	config := CacheConfig{
		Capacity: 10,
		TTL:      time.Hour,
	}
	cache := NewAPIResponseCache(config)
	if cache == nil {
		t.Fatal("Expected API response cache to be created")
	}
}

func TestCacheKeys(t *testing.T) {
	config := CacheConfig{
		Capacity: 5,
		TTL:      time.Hour,
	}
	cache := NewLRUCache(config)
	
	// Test Keys method on empty cache
	keys := cache.Keys()
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys in empty cache, got %d", len(keys))
	}
	
	// Add some data and test keys again
	cache.Set("test1", "value1")
	cache.Set("test2", "value2")
	keys = cache.Keys()
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(keys))
	}
}

func TestCacheMemoryLimits(t *testing.T) {
	// Create cache with small memory limit (1KB)
	config := CacheConfig{
		Capacity:  1000, // Large capacity but limited by memory
		MaxMemory: 1024, // 1KB memory limit
		TTL:       time.Hour,
	}
	cache := NewLRUCache(config)
	
	// Add items that should exceed memory limit
	largeValue := make([]byte, 500) // 500 bytes
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("key%d", i)
		cache.Set(key, largeValue)
	}
	
	// Should have evicted some items due to memory pressure
	stats := cache.Stats()
	if stats.MemoryEvicts == 0 {
		t.Error("Expected memory evictions to occur")
	}
	
	if stats.CurrentMemory > stats.MaxMemory {
		t.Errorf("Memory usage %d exceeds limit %d", stats.CurrentMemory, stats.MaxMemory)
	}
	
	if stats.Size == 10 {
		t.Error("Expected some items to be evicted due to memory pressure")
	}
}

func TestCacheMemoryEstimation(t *testing.T) {
	// Test memory estimation for different value types
	tests := []struct {
		key   string
		value interface{}
		name  string
	}{
		{"string_key", "test_value", "string"},
		{"bytes_key", []byte("test_bytes"), "bytes"},
		{"int_key", 12345, "int"},
		{"bool_key", true, "bool"},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			size := estimateSize(test.key, test.value)
			if size <= 0 {
				t.Errorf("Expected positive size for %s, got %d", test.name, size)
			}
			// Size should at least include the key length
			if size < int64(len(test.key)) {
				t.Errorf("Size %d should be at least key length %d", size, len(test.key))
			}
		})
	}
}