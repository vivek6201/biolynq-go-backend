package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache is a type-safe generic wrapper around the Redis client.
type Cache[T any] struct {
	rdb *redis.Client
}

// NewCache instantiates a new generic cache for a specific model type T.
func NewCache[T any](rdb *redis.Client) *Cache[T] {
	return &Cache[T]{rdb: rdb}
}

// BuildKey generates a formatted string cache key, e.g. "session:123-abc"
func BuildKey(prefix string, args ...interface{}) string {
	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = fmt.Sprintf("%v", arg)
	}
	return prefix + ":" + strings.Join(parts, ":")
}

// Get retrieves a key from Redis and deserializes it into the strictly typed T structure.
func (c *Cache[T]) Get(ctx context.Context, key string) (*T, error) {
	val, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, err // returns redis.Nil on cache miss
	}

	var dest T
	if err := json.Unmarshal([]byte(val), &dest); err != nil {
		return nil, err
	}
	return &dest, nil
}

// Set marshals a strictly typed T value and stores it in Redis with an expiration time.
func (c *Cache[T]) Set(ctx context.Context, key string, value *T, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, data, expiration).Err()
}

// Delete removes a key from the cache.
func (c *Cache[T]) Delete(ctx context.Context, key string) error {
	return c.rdb.Del(ctx, key).Err()
}

// InvalidateAsync deletes one or more cache keys in a background goroutine.
// It returns immediately so the caller is never blocked by cache eviction.
func (c *Cache[T]) InvalidateAsync(keys ...string) {
	go func() {
		ctx := context.Background()
		for _, key := range keys {
			_ = c.rdb.Del(ctx, key).Err()
		}
	}()
}

// Fetch implements the Cache-Aside pattern strictly typed to T.
// 1. Checks Redis cache. If hit, returns the parsed struct pointer.
// 2. On miss, runs the dbQuery callback to load strictly typed T data from PostgreSQL.
// 3. Caches the loaded data in Redis asynchronously, then returns it.
func (c *Cache[T]) Fetch(
	ctx context.Context,
	key string,
	expiration time.Duration,
	dbQuery func() (*T, error),
) (*T, error) {
	// 1. Try Cache hit
	result, err := c.Get(ctx, key)
	if err == nil {
		return result, nil
	}

	// 2. Cache miss -> Hit Database
	result, err = dbQuery()
	if err != nil {
		return nil, err
	}

	// 3. Cache the loaded result asynchronously (non-blocking)
	go func() {
		c.Set(context.Background(), key, result, expiration)
	}()

	return result, nil
}
