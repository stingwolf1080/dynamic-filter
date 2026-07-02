package cache

import (
	"sync"
)

// Cache is a simple, thread-safe in-memory cache that persists until restart.
type Cache struct {
	data  map[string]any
	mutex sync.RWMutex
}

// NewCache creates a new Cache instance.
func NewCache() *Cache {
	return &Cache{
		data: make(map[string]any),
	}
}

// Get retrieves a value from the cache by key.
func (c *Cache) Get(key string) (any, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	val, ok := c.data[key]
	return val, ok
}

// Set adds a value to the cache by key.
func (c *Cache) Set(key string, val any) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data[key] = val
}

// Clear empties the entire cache.
func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data = make(map[string]any)
}

// Delete removes a specific key from the cache.
func (c *Cache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.data, key)
}
