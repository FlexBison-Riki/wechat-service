package cache

import (
	"context"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/silenceper/wechat/v2/cache"
)

// RedisCache implements cache.Cache interface using Redis
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(addr, password string, db int) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisCache{client: client}
}

// NewMemoryCache creates a memory cache for development
func NewMemoryCache() *MemoryCache {
	mr, _ := miniredis.Run()
	return &MemoryCache{
		client: redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		}),
	}
}

// Get retrieves a value from cache
func (c *RedisCache) Get(key string) (interface{}, error) {
	return c.client.Get(context.Background(), key).Result()
}

// Set stores a value in cache with TTL
func (c *RedisCache) Set(key string, value interface{}, ttl time.Duration) error {
	data, err := serialize(value)
	if err != nil {
		return err
	}
	return c.client.Set(context.Background(), key, data, ttl).Err()
}

// Delete removes a key from cache
func (c *RedisCache) Delete(key string) error {
	return c.client.Del(context.Background(), key).Err()
}

// MemoryCache is a simple in-memory cache implementation
type MemoryCache struct {
	client *redis.Client
}

// Get retrieves a value from memory cache
func (c *MemoryCache) Get(key string) (interface{}, error) {
	return c.client.Get(context.Background(), key).Result()
}

// Set stores a value in memory cache with TTL
func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) error {
	data, err := serialize(value)
	if err != nil {
		return err
	}
	return c.client.Set(context.Background(), key, data, ttl).Err()
}

// Delete removes a key from memory cache
func (c *MemoryCache) Delete(key string) error {
	return c.client.Del(context.Background(), key).Err()
}

// Ensure RedisCache implements cache.Cache interface
var _ cache.Cache = (*RedisCache)(nil)
