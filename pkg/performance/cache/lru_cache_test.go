package cache

import (
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