package viacep

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/gob"
	"fmt"
	"strings"
	"sync"
	"time"
)

const cacheTTL = 3600 * time.Second
const cachePrefix = "viacep:"

type Cache interface {
	// Get retrieves an item from the cache by its key.
	//
	// Parameters:
	//   - ctx: The context for managing cancellation, timeouts, and deadlines.
	//   - key: The key of the cache entry to retrieve.
	//   - dest: A pointer to a variable where the cached value should be stored.
	//           The type of `dest` must be compatible with the cached value type.
	//
	// Returns:
	//   - A boolean indicating whether the key was found in the cache (true if found, false if not).
	//   - If the key is found, the value will be copied into `dest`.
	//
	// Example:
	//   var username string
	//   found := cache.Get(ctx, "username", &username)
	//   if found {
	//       fmt.Println("Username found:", username)
	//   } else {
	//       fmt.Println("Username not found")
	//   }
	Get(ctx context.Context, key string, dest any) bool

	// Set stores an item in the cache with the specified key, value, and optional time-to-live (TTL).
	//
	// Parameters:
	//   - ctx: The context for managing cancellation, timeouts, and deadlines.
	//   - key: The key under which the value will be stored in the cache.
	//   - value: The value to store in the cache, which can be of any type.
	//   - ttl: The time-to-live (TTL) duration for the cache entry. If ttl is zero, the entry will not expire.
	//
	// Returns:
	//   - An error if the cache operation fails, or nil if the operation is successful.
	//
	// Example:
	//   err := cache.Set(ctx, "userID", 1234, 10*time.Minute)
	//   if err != nil {
	//       fmt.Println("Error setting cache:", err)
	//   }
	Set(ctx context.Context, key string, value any, ttl time.Duration) error

	// Delete removes an item from the cache by its key.
	//
	// Parameters:
	//   - ctx: The context for managing cancellation, timeouts, and deadlines.
	//   - key: The key of the cache entry to remove.
	//
	// Returns:
	//   - An error if the cache operation fails, or nil if the operation is successful.
	//
	// Example:
	//   err := cache.Delete(ctx, "username")
	//   if err != nil {
	//       fmt.Println("Error deleting cache:", err)
	//   }
	Delete(ctx context.Context, key string) error
}

type memoryCache struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func newMemoryCache() *memoryCache {
	return &memoryCache{
		data: make(map[string][]byte),
	}
}

func (c *memoryCache) Get(_ context.Context, key string, dest any) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	serialized, exists := c.data[key]
	if !exists {
		return false
	}

	buffer := bytes.NewBuffer([]byte(serialized))
	decoder := gob.NewDecoder(buffer)
	if err := decoder.Decode(dest); err != nil {
		return false
	}

	return true
}

func (c *memoryCache) Set(_ context.Context, key string, value any, ttl time.Duration) error {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(value); err != nil {
		return fmt.Errorf("failed to encode value of type %T: %w", value, err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = buffer.Bytes()

	if ttl > 0 {
		go func() {
			time.Sleep(ttl)
			_ = c.Delete(context.TODO(), key)
		}()
	}

	return nil
}

func (c *memoryCache) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
	return nil
}

func cacheKey(value ...string) string {
	return fmt.Sprintf("%s%x", cachePrefix, md5.Sum([]byte(strings.Join(value, ","))))
}
