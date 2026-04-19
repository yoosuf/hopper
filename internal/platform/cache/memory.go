package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/yoosuf/hopper/internal/platform/logger"
)

// MemoryCache implements the Cache interface using in-memory storage
type MemoryCache struct {
	data   map[string]*cacheEntry
	mu     sync.RWMutex
	logger logger.Logger
}

type cacheEntry struct {
	value      []byte
	expiration time.Time
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(log logger.Logger) *MemoryCache {
	c := &MemoryCache{
		data:   make(map[string]*cacheEntry),
		logger: log,
	}
	
	// Start periodic cleanup goroutine
	go c.cleanupExpired()
	
	return c
}

// Get retrieves a value from the cache
func (c *MemoryCache) Get(ctx context.Context, key string, dest interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		return fmt.Errorf("cache miss")
	}

	// Check if expired
	if time.Now().After(entry.expiration) {
		return fmt.Errorf("cache miss")
	}

	if err := json.Unmarshal(entry.value, dest); err != nil {
		return fmt.Errorf("failed to unmarshal cached value: %w", err)
	}

	return nil
}

// Set stores a value in the cache with a TTL
func (c *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value for cache: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = &cacheEntry{
		value:      data,
		expiration: time.Now().Add(ttl),
	}

	return nil
}

// Delete removes a value from the cache
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
	return nil
}

// Exists checks if a key exists in the cache
func (c *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		return false, nil
	}

	// Check if expired
	if time.Now().After(entry.expiration) {
		return false, nil
	}

	return true, nil
}

// Clear removes all keys from the cache
func (c *MemoryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*cacheEntry)
	return nil
}

// cleanupExpired periodically removes expired entries
func (c *MemoryCache) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.data {
			if now.After(entry.expiration) {
				delete(c.data, key)
			}
		}
		c.mu.Unlock()
	}
}

// Close is a no-op for memory cache
func (c *MemoryCache) Close() error {
	return nil
}
