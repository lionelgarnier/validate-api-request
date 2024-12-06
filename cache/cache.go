package cache

import (
	"sync"
	"time"
)

// CacheStats holds cache performance metrics
type CacheStats struct {
	Hits       int64
	Misses     int64
	Size       int
	MaxSize    int
	EvictCount int64
}

// CacheEntry represents a cached item with metadata
type CacheEntry[T any] struct {
	Value      T
	ExpiresAt  time.Time
	LastAccess time.Time
}

// Cache defines the generic caching interface
type Cache[T any] interface {
	Get(key uint64) (T, bool)
	Set(key uint64, value T)
	Remove(key uint64)
	Clear()
	Stats() CacheStats
}

// BaseCache implements common caching functionality
type BaseCache[T any] struct {
	entries map[uint64]*CacheEntry[T]
	maxSize int
	stats   CacheStats
	ttl     time.Duration
	mu      sync.RWMutex
}

func NewBaseCache[T any](maxSize int, ttl time.Duration) *BaseCache[T] {
	return &BaseCache[T]{
		entries: make(map[uint64]*CacheEntry[T]),
		maxSize: maxSize,
		ttl:     ttl,
	}
}

// Common methods implementation
func (c *BaseCache[T]) Get(key uint64) (T, bool) {
	c.mu.RLock()
	entry, exists := c.entries[key]
	if !exists {
		c.mu.RUnlock()
		c.stats.Misses++
		var zero T
		return zero, false
	}

	if time.Now().After(entry.ExpiresAt) {
		c.mu.RUnlock()
		c.mu.Lock()
		if entry, exists = c.entries[key]; exists && time.Now().After(entry.ExpiresAt) {
			delete(c.entries, key)
			c.stats.Size = len(c.entries)
			c.stats.Misses++
		}
		c.mu.Unlock()
		var zero T
		return zero, false
	}

	entry.LastAccess = time.Now()
	value := entry.Value
	c.mu.RUnlock()
	c.stats.Hits++
	return value, true
}

func (c *BaseCache[T]) Set(key uint64, value T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.entries) >= c.maxSize {
		c.evictOldest()
	}

	c.entries[key] = &CacheEntry[T]{
		Value:      value,
		ExpiresAt:  time.Now().Add(c.ttl),
		LastAccess: time.Now(),
	}
	c.stats.Size = len(c.entries)
}

func (c *BaseCache[T]) Remove(key uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
	c.stats.Size = len(c.entries)
}

func (c *BaseCache[T]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[uint64]*CacheEntry[T])
	c.stats.Size = 0
}

func (c *BaseCache[T]) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats
}

func (c *BaseCache[T]) evictOldest() {
	var oldestKey uint64
	var oldestAccess time.Time

	for key, entry := range c.entries {
		if oldestKey == 0 || entry.LastAccess.Before(oldestAccess) {
			oldestKey = key
			oldestAccess = entry.LastAccess
		}
	}

	if oldestKey != 0 {
		delete(c.entries, oldestKey)
		c.stats.EvictCount++
	}
}
