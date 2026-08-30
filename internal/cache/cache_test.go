package cache

import (
	"testing"
	"time"
)

func TestCacheSetGet(t *testing.T) {
	c := NewCache(10, 1*time.Second)
	defer c.Stop()

	c.Set("key1", "value1")

	val, exists := c.Get("key1")
	if !exists {
		t.Errorf("Expected key to exist")
	}

	if val != "value1" {
		t.Errorf("Expected 'value1', got %v", val)
	}
}

func TestCacheExpiration(t *testing.T) {
	c := NewCache(10, 100*time.Millisecond)
	defer c.Stop()

	c.Set("key1", "value1")

	// Value should exist immediately
	_, exists := c.Get("key1")
	if !exists {
		t.Errorf("Expected key to exist immediately")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Value should be expired
	_, exists = c.Get("key1")
	if exists {
		t.Errorf("Expected key to be expired")
	}
}

func TestCacheDelete(t *testing.T) {
	c := NewCache(10, 1*time.Second)
	defer c.Stop()

	c.Set("key1", "value1")
	c.Delete("key1")

	_, exists := c.Get("key1")
	if exists {
		t.Errorf("Expected key to be deleted")
	}
}

func TestCacheClear(t *testing.T) {
	c := NewCache(10, 1*time.Second)
	defer c.Stop()

	c.Set("key1", "value1")
	c.Set("key2", "value2")
	c.Set("key3", "value3")

	if c.Size() != 3 {
		t.Errorf("Expected size 3, got %d", c.Size())
	}

	c.Clear()

	if c.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", c.Size())
	}
}

func TestCacheEviction(t *testing.T) {
	c := NewCache(2, 10*time.Second)
	defer c.Stop()

	c.Set("key1", "value1")
	c.Set("key2", "value2")
	c.Set("key3", "value3") // Should evict key1

	if c.Size() != 2 {
		t.Errorf("Expected size 2, got %d", c.Size())
	}

	_, exists := c.Get("key1")
	if exists {
		t.Errorf("Expected key1 to be evicted")
	}
}

func TestKeyBuilder(t *testing.T) {
	kb := NewKeyBuilder().
		Add("session").
		Add("12345").
		Add("decisions")

	key := kb.Build()

	if key == "" {
		t.Errorf("Expected key to be generated")
	}

	// Same components should generate same key
	kb2 := NewKeyBuilder().
		Add("session").
		Add("12345").
		Add("decisions")

	if key != kb2.Build() {
		t.Errorf("Expected same key for same components")
	}
}

func TestCacheStats(t *testing.T) {
	c := NewCache(10, 1*time.Second)
	defer c.Stop()

	c.Set("key1", "value1")
	c.Get("key1")
	c.Get("key1")

	stats := c.Stats()

	if stats.Size != 1 {
		t.Errorf("Expected size 1, got %d", stats.Size)
	}

	if stats.MaxSize != 10 {
		t.Errorf("Expected maxSize 10, got %d", stats.MaxSize)
	}

	if stats.TotalAccess != 2 {
		t.Errorf("Expected 2 accesses, got %d", stats.TotalAccess)
	}
}
