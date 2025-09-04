package cache

import (
	"container/list"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"
	"unsafe"
)

// CacheItem represents an item in the cache
type CacheItem struct {
	Key        string
	Value      interface{}
	ExpiresAt  time.Time
	AccessedAt time.Time
	element    *list.Element
	size       int64 // Estimated size in bytes
}

// LRUCache represents an LRU cache with TTL support
type LRUCache struct {
	mutex        sync.RWMutex
	capacity     int
	maxMemory    int64 // Maximum memory usage in bytes
	currentMemory int64 // Current memory usage in bytes
	items        map[string]*CacheItem
	evictList    *list.List
	ttl          time.Duration
	
	// Statistics
	hits         int64
	misses       int64
	sets         int64
	evicts       int64
	memoryEvicts int64 // Evictions due to memory pressure
}

// CacheConfig holds configuration for LRU cache
type CacheConfig struct {
	Capacity  int
	MaxMemory int64         // Maximum memory usage in bytes (0 = no limit)
	TTL       time.Duration
}

// NewLRUCache creates a new LRU cache
func NewLRUCache(config CacheConfig) *LRUCache {
	if config.Capacity <= 0 {
		config.Capacity = 1000
	}
	
	if config.TTL <= 0 {
		config.TTL = 24 * time.Hour
	}

	return &LRUCache{
		capacity:  config.Capacity,
		maxMemory: config.MaxMemory,
		items:     make(map[string]*CacheItem),
		evictList: list.New(),
		ttl:       config.TTL,
	}
}

// Get retrieves an item from the cache
func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	item, exists := c.items[key]
	if !exists {
		c.misses++
		return nil, false
	}

	// Check if item has expired
	if time.Now().After(item.ExpiresAt) {
		c.removeElement(item)
		c.misses++
		return nil, false
	}

	// Move to front (most recently used)
	c.evictList.MoveToFront(item.element)
	item.AccessedAt = time.Now()
	c.hits++
	
	return item.Value, true
}

// Set adds or updates an item in the cache
func (c *LRUCache) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	expiresAt := now.Add(c.ttl)

	// Estimate size of new item
	newSize := estimateSize(key, value)
	
	// Check if item already exists
	if item, exists := c.items[key]; exists {
		// Update existing item
		oldSize := item.size
		item.Value = value
		item.ExpiresAt = expiresAt
		item.AccessedAt = now
		item.size = newSize
		c.evictList.MoveToFront(item.element)
		c.sets++
		
		// Update memory usage
		c.currentMemory = c.currentMemory - oldSize + newSize
		
		// Check memory pressure
		if c.maxMemory > 0 && c.currentMemory > c.maxMemory {
			c.evictForMemory()
		}
		return
	}

	// Create new item
	item := &CacheItem{
		Key:        key,
		Value:      value,
		ExpiresAt:  expiresAt,
		AccessedAt: now,
		size:       newSize,
	}

	// Add to front of list
	item.element = c.evictList.PushFront(item)
	c.items[key] = item
	c.sets++
	c.currentMemory += newSize

	// Check if we exceed capacity or memory limit
	if c.evictList.Len() > c.capacity {
		c.evictOldest()
	} else if c.maxMemory > 0 && c.currentMemory > c.maxMemory {
		c.evictForMemory()
	}
}

// Delete removes an item from the cache
func (c *LRUCache) Delete(key string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	item, exists := c.items[key]
	if !exists {
		return false
	}

	c.removeElement(item)
	return true
}

// Clear removes all items from the cache
func (c *LRUCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]*CacheItem)
	c.evictList.Init()
	c.currentMemory = 0
}

// Size returns the current number of items in the cache
func (c *LRUCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.items)
}

// Stats returns cache statistics
func (c *LRUCache) Stats() CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	total := c.hits + c.misses
	hitRatio := float64(0)
	if total > 0 {
		hitRatio = float64(c.hits) / float64(total)
	}

	memoryRatio := float64(0)
	if c.maxMemory > 0 {
		memoryRatio = float64(c.currentMemory) / float64(c.maxMemory)
	}

	return CacheStats{
		Hits:          c.hits,
		Misses:        c.misses,
		Sets:          c.sets,
		Evicts:        c.evicts,
		MemoryEvicts:  c.memoryEvicts,
		HitRatio:      hitRatio,
		Size:          len(c.items),
		Capacity:      c.capacity,
		CurrentMemory: c.currentMemory,
		MaxMemory:     c.maxMemory,
		MemoryRatio:   memoryRatio,
	}
}

// CleanupExpired removes expired items from the cache
func (c *LRUCache) CleanupExpired() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	expired := make([]*CacheItem, 0)

	// Find expired items
	for _, item := range c.items {
		if now.After(item.ExpiresAt) {
			expired = append(expired, item)
		}
	}

	// Remove expired items
	for _, item := range expired {
		c.removeElement(item)
	}

	return len(expired)
}

// Keys returns all cache keys
func (c *LRUCache) Keys() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	keys := make([]string, 0, len(c.items))
	for key := range c.items {
		keys = append(keys, key)
	}
	return keys
}

// evictOldest removes the least recently used item
func (c *LRUCache) evictOldest() {
	element := c.evictList.Back()
	if element != nil {
		item := element.Value.(*CacheItem)
		c.removeElement(item)
		c.evicts++
	}
}

// removeElement removes an item from both the map and list
func (c *LRUCache) removeElement(item *CacheItem) {
	delete(c.items, item.Key)
	c.evictList.Remove(item.element)
	c.currentMemory -= item.size
}

// CacheStats represents cache statistics
type CacheStats struct {
	Hits          int64   `json:"hits"`
	Misses        int64   `json:"misses"`
	Sets          int64   `json:"sets"`
	Evicts        int64   `json:"evicts"`
	MemoryEvicts  int64   `json:"memory_evicts"`
	HitRatio      float64 `json:"hit_ratio"`
	Size          int     `json:"size"`
	Capacity      int     `json:"capacity"`
	CurrentMemory int64   `json:"current_memory"`
	MaxMemory     int64   `json:"max_memory"`
	MemoryRatio   float64 `json:"memory_ratio"`
}

// String returns a string representation of cache stats
func (s CacheStats) String() string {
	memoryStr := ""
	if s.MaxMemory > 0 {
		memoryStr = fmt.Sprintf(", Memory=%d/%d (%.1f%%)", s.CurrentMemory, s.MaxMemory, s.MemoryRatio*100)
	}
	return fmt.Sprintf("Cache Stats: Hits=%d, Misses=%d, Hit Ratio=%.2f%%, Size=%d/%d, Evicts=%d (Memory: %d)%s", 
		s.Hits, s.Misses, s.HitRatio*100, s.Size, s.Capacity, s.Evicts, s.MemoryEvicts, memoryStr)
}

// estimateSize estimates the memory size of a cache item
func estimateSize(key string, value interface{}) int64 {
	size := int64(len(key)) // Key size
	
	// Estimate value size based on type
	switch v := value.(type) {
	case string:
		size += int64(len(v))
	case []byte:
		size += int64(len(v))
	case int, int32, int64, uint, uint32, uint64:
		size += 8
	case float32, float64:
		size += 8
	case bool:
		size += 1
	default:
		// For complex types, use reflection to estimate size
		size += estimateReflectSize(reflect.ValueOf(v))
	}
	
	// Add overhead for CacheItem struct (estimated)
	size += int64(unsafe.Sizeof(CacheItem{}))
	
	return size
}

// estimateReflectSize estimates size using reflection (less accurate but handles any type)
func estimateReflectSize(v reflect.Value) int64 {
	if !v.IsValid() {
		return 0
	}
	
	switch v.Kind() {
	case reflect.String:
		return int64(v.Len())
	case reflect.Slice, reflect.Array:
		size := int64(v.Len() * 8) // Estimate 8 bytes per element
		if v.Len() > 0 && v.Index(0).Kind() == reflect.Uint8 {
			// Special case for byte slices
			return int64(v.Len())
		}
		return size
	case reflect.Map:
		return int64(v.Len() * 16) // Estimate 16 bytes per map entry
	case reflect.Ptr:
		if v.IsNil() {
			return 8
		}
		return 8 + estimateReflectSize(v.Elem())
	case reflect.Struct:
		size := int64(0)
		for i := 0; i < v.NumField(); i++ {
			size += estimateReflectSize(v.Field(i))
		}
		return size
	default:
		// Default size estimate
		return 64
	}
}

// evictForMemory evicts items until memory usage is under limit
func (c *LRUCache) evictForMemory() {
	for c.maxMemory > 0 && c.currentMemory > c.maxMemory && c.evictList.Len() > 0 {
		element := c.evictList.Back()
		if element != nil {
			item := element.Value.(*CacheItem)
			c.removeElement(item)
			c.memoryEvicts++
		}
	}
}

// APIResponseCache wraps LRUCache for API response caching
type APIResponseCache struct {
	cache *LRUCache
}

// NewAPIResponseCache creates a new API response cache
func NewAPIResponseCache(config CacheConfig) *APIResponseCache {
	return &APIResponseCache{
		cache: NewLRUCache(config),
	}
}

// GetResponse retrieves a cached API response
func (c *APIResponseCache) GetResponse(endpoint string) ([]byte, bool) {
	value, exists := c.cache.Get(endpoint)
	if !exists {
		return nil, false
	}
	
	data, ok := value.([]byte)
	if !ok {
		return nil, false
	}
	
	return data, true
}

// SetResponse caches an API response
func (c *APIResponseCache) SetResponse(endpoint string, response []byte) {
	c.cache.Set(endpoint, response)
}

// GetJSON retrieves and unmarshals a cached JSON response
func (c *APIResponseCache) GetJSON(endpoint string, v interface{}) bool {
	data, exists := c.GetResponse(endpoint)
	if !exists {
		return false
	}
	
	err := json.Unmarshal(data, v)
	return err == nil
}

// SetJSON marshals and caches a JSON response
func (c *APIResponseCache) SetJSON(endpoint string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	c.SetResponse(endpoint, data)
	return nil
}

// Stats returns cache statistics
func (c *APIResponseCache) Stats() CacheStats {
	return c.cache.Stats()
}

// Clear clears the cache
func (c *APIResponseCache) Clear() {
	c.cache.Clear()
}

// CleanupExpired removes expired items
func (c *APIResponseCache) CleanupExpired() int {
	return c.cache.CleanupExpired()
} 