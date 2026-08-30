package cache

import (
	"crypto/md5"
	"fmt"
	"sync"
	"time"
)

// CacheEntry represents a cached value
type CacheEntry struct {
	Value      interface{}
	CreatedAt  time.Time
	ExpiresAt  time.Time
	AccessCount int
}

// Cache provides in-memory caching with TTL and LRU eviction
type Cache struct {
	mu       sync.RWMutex
	data     map[string]*CacheEntry
	maxSize  int
	ttl      time.Duration
	cleanupInterval time.Duration
	stopCleanup chan bool
}

// NewCache creates a new cache with the specified size and TTL
func NewCache(maxSize int, ttl time.Duration) *Cache {
	c := &Cache{
		data:    make(map[string]*CacheEntry),
		maxSize: maxSize,
		ttl:     ttl,
		cleanupInterval: time.Minute,
		stopCleanup: make(chan bool),
	}

	// Start cleanup goroutine
	go c.cleanupExpired()

	return c
}

// Set stores a value in the cache
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict oldest entry if cache is full (before adding new entry)
	if len(c.data) >= c.maxSize {
		// Don't evict if we're just updating an existing key
		if _, exists := c.data[key]; !exists {
			c.evictOldest()
		}
	}

	now := time.Now()
	c.data[key] = &CacheEntry{
		Value:     value,
		CreatedAt: now,
		ExpiresAt: now.Add(c.ttl),
	}
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	entry, exists := c.data[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		c.Delete(key)
		return nil, false
	}

	// Update access count
	c.mu.Lock()
	entry.AccessCount++
	c.mu.Unlock()

	return entry.Value, true
}

// Delete removes a value from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

// Clear removes all values from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]*CacheEntry)
}

// Size returns the number of items in the cache
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}

// Stats returns cache statistics
func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	totalAccess := 0
	for _, entry := range c.data {
		totalAccess += entry.AccessCount
	}

	return CacheStats{
		Size:        len(c.data),
		MaxSize:     c.maxSize,
		TotalAccess: totalAccess,
		TTL:         c.ttl.String(),
	}
}

// CacheStats contains cache statistics
type CacheStats struct {
	Size        int    `json:"size"`
	MaxSize     int    `json:"max_size"`
	TotalAccess int    `json:"total_access"`
	TTL         string `json:"ttl"`
}

// evictOldest removes the oldest entry from the cache
func (c *Cache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.data {
		if oldestTime.IsZero() || entry.CreatedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.CreatedAt
		}
	}

	if oldestKey != "" {
		delete(c.data, oldestKey)
	}
}

// cleanupExpired periodically removes expired entries
func (c *Cache) cleanupExpired() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.removeExpired()
		case <-c.stopCleanup:
			return
		}
	}
}

// removeExpired removes all expired entries
func (c *Cache) removeExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.data {
		if now.After(entry.ExpiresAt) {
			delete(c.data, key)
		}
	}
}

// Stop stops the cleanup goroutine
func (c *Cache) Stop() {
	c.stopCleanup <- true
}

// GenerateKey creates a cache key from components
func GenerateKey(components ...string) string {
	hash := md5.Sum([]byte(fmt.Sprint(components)))
	return fmt.Sprintf("%x", hash)
}

// KeyBuilder helps construct cache keys
type KeyBuilder struct {
	components []string
}

// NewKeyBuilder creates a new key builder
func NewKeyBuilder() *KeyBuilder {
	return &KeyBuilder{components: make([]string, 0)}
}

// Add adds a component to the key
func (kb *KeyBuilder) Add(component string) *KeyBuilder {
	kb.components = append(kb.components, component)
	return kb
}

// Build generates the final cache key
func (kb *KeyBuilder) Build() string {
	return GenerateKey(kb.components...)
}
