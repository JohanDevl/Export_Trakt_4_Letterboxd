package cache

import (
	"container/list"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CacheItem represents an item in the cache
type CacheItem struct {
	Key        string
	Value      interface{}
	ExpiresAt  time.Time
	AccessedAt time.Time
	element    *list.Element
}

// LRUCache represents an LRU cache with TTL support
type LRUCache struct {
	mutex     sync.RWMutex
	capacity  int
	items     map[string]*CacheItem
	evictList *list.List
	ttl       time.Duration
	
	// Statistics
	hits   int64
	misses int64
	sets   int64
	evicts int64
}

// CacheConfig holds configuration for LRU cache
type CacheConfig struct {
	Capacity int
	TTL      time.Duration
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

	// Check if item already exists
	if item, exists := c.items[key]; exists {
		// Update existing item
		item.Value = value
		item.ExpiresAt = expiresAt
		item.AccessedAt = now
		c.evictList.MoveToFront(item.element)
		c.sets++
		return
	}

	// Create new item
	item := &CacheItem{
		Key:        key,
		Value:      value,
		ExpiresAt:  expiresAt,
		AccessedAt: now,
	}

	// Add to front of list
	item.element = c.evictList.PushFront(item)
	c.items[key] = item
	c.sets++

	// Check if we exceed capacity
	if c.evictList.Len() > c.capacity {
		c.evictOldest()
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

	return CacheStats{
		Hits:      c.hits,
		Misses:    c.misses,
		Sets:      c.sets,
		Evicts:    c.evicts,
		HitRatio:  hitRatio,
		Size:      len(c.items),
		Capacity:  c.capacity,
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
}

// CacheStats represents cache statistics
type CacheStats struct {
	Hits     int64   `json:"hits"`
	Misses   int64   `json:"misses"`
	Sets     int64   `json:"sets"`
	Evicts   int64   `json:"evicts"`
	HitRatio float64 `json:"hit_ratio"`
	Size     int     `json:"size"`
	Capacity int     `json:"capacity"`
}

// String returns a string representation of cache stats
func (s CacheStats) String() string {
	return fmt.Sprintf("Cache Stats: Hits=%d, Misses=%d, Hit Ratio=%.2f%%, Size=%d/%d, Evicts=%d", 
		s.Hits, s.Misses, s.HitRatio*100, s.Size, s.Capacity, s.Evicts)
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